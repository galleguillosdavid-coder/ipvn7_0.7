// Package l1 implementa el enrutamiento cebolla Sphinx de 3 saltos
// con tramas de longitud fija estricta de 1280 bytes (Deterministic MTU Wire),
// desprendimiento iterativo de capas (peeling) y protección anti-repetición O(1).
// Conforme a genesis.md (L178) y 2.md (Sección 2).
package l1

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"golang.org/x/crypto/chacha20"
)

const (
	// SphinxPacketSize define el tamaño inmutable del wire determinista (1280 bytes)
	SphinxPacketSize = 1280

	// Topología de 3 saltos
	SphinxHopsCount = 3

	// Estructura de cabeceras sincronizadas por capas
	SphinxHopKEMSize        = 160 // 32 bytes X25519 efímero + 128 bytes ML-KEM ciphertext
	SphinxKeyHeaderSize     = SphinxHopsCount * SphinxHopKEMSize // 480 bytes
	SphinxHopDescriptorSize = 128 // Tamaño fijo de descriptor por cada salto
	SphinxRoutingTotalSize  = SphinxHopsCount * SphinxHopDescriptorSize // 384 bytes
	SphinxPayloadSize       = SphinxPacketSize - SphinxKeyHeaderSize - SphinxRoutingTotalSize // 416 bytes

	// Acciones de peeling
	SphinxActionForward = "FORWARD"
	SphinxActionDeliver = "DELIVER"
	SphinxActionDrop    = "DROP"
)

// SphinxHopNode define la identidad y claves públicas de un nodo en el circuito
type SphinxHopNode struct {
	DID          string             `json:"did"`
	Address      string             `json:"address"`
	HybridKeys   *HybridKeyPair     `json:"-"`
	X25519PubHex string             `json:"x25519_pub_hex"`
	MLKEMPubHex  string             `json:"ml_kem_pub_hex"`
}

// SphinxCircuit contiene los 3 nodos seleccionados: Guard, Middle, Exit
type SphinxCircuit struct {
	CircuitID string          `json:"circuit_id"`
	Guard     *SphinxHopNode  `json:"guard"`
	Middle    *SphinxHopNode  `json:"middle"`
	Exit      *SphinxHopNode  `json:"exit"`
	CreatedAt time.Time       `json:"created_at"`
}


// SphinxRouter gestiona la creación de circuitos, envío y peeling con protección anti-replay
type SphinxRouter struct {
	mu           sync.RWMutex
	LocalKeys    *HybridKeyPair
	replayCache  map[string]int64 // tagHash -> expiración epoch unix
	replayOrder  []string
	maxCacheSize int
}

// NewSphinxRouter inicializa el enrutador cebolla
func NewSphinxRouter(localKeys *HybridKeyPair) *SphinxRouter {
	return &SphinxRouter{
		LocalKeys:    localKeys,
		replayCache:  make(map[string]int64),
		replayOrder:  make([]string, 0, 10000),
		maxCacheSize: 10000,
	}
}

// PeelLayer procesa un paquete cebolla en el nodo actual, pelando una capa de cifrado
func (sr *SphinxRouter) PeelLayer(packet *SphinxPacket) (*PeelResult, error) {
	start := time.Now()

	if packet == nil {
		return nil, errors.New("paquete nulo")
	}

	// 1. Control Anti-Repetición O(1) usando el hash de la primera ranura de clave
	tagHashBytes := sha256.Sum256(packet.KeyHeader[:SphinxHopKEMSize])
	tagHash := hex.EncodeToString(tagHashBytes[:])

	sr.mu.Lock()
	now := time.Now().Unix()
	if exp, exists := sr.replayCache[tagHash]; exists && exp > now {
		sr.mu.Unlock()
		return &PeelResult{
			Action:     SphinxActionDrop,
			TagHash:    tagHash,
			PeelTimeUs: time.Since(start).Microseconds(),
		}, errors.New("ataque de repetición detectado: paquete cebolla ya procesado")
	}
	sr.replayCache[tagHash] = now + 300 // TTL 5 minutos
	sr.replayOrder = append(sr.replayOrder, tagHash)
	if len(sr.replayOrder) > sr.maxCacheSize {
		oldest := sr.replayOrder[0]
		sr.replayOrder = sr.replayOrder[1:]
		delete(sr.replayCache, oldest)
	}
	sr.mu.Unlock()

	// 2. Derivar clave compartida para este salto a partir de la primera ranura de la cabecera efímera
	kemCipher := &HybridKEMCiphertext{
		Algorithm:       HybridKEMAlgorithm,
		EphemeralX25519: packet.KeyHeader[0:32],
		PQCCiphertext:   packet.KeyHeader[32:160],
		Salt:            make([]byte, 16),
	}

	hopKey, err := sr.LocalKeys.Decapsulate(kemCipher)
	if err != nil {
		return nil, fmt.Errorf("error decapsulando KEM en salto: %w", err)
	}

	// 3. Descifrar el descriptor de enrutamiento actual (primeros 128 bytes)
	descBytes := make([]byte, SphinxHopDescriptorSize)
	copy(descBytes, packet.RoutingHeader[:SphinxHopDescriptorSize])

	stream, err := chacha20.NewUnauthenticatedCipher(hopKey, make([]byte, 12))
	if err != nil {
		return nil, err
	}
	stream.XORKeyStream(descBytes, descBytes)

	// Validar MAC del descriptor
	mac := hmac.New(sha256.New, hopKey)
	mac.Write(descBytes[:49])
	expectedMAC := mac.Sum(nil)[:16]
	if !hmac.Equal(descBytes[49:65], expectedMAC) {
		return nil, errors.New("descriptor de enrutamiento corrupto o clave incorrecta (fallo de MAC)")
	}

	nextHopDID := string(bytes.Trim(descBytes[0:32], "\x00"))
	circuitID := string(bytes.Trim(descBytes[32:48], "\x00"))
	isExit := descBytes[48] == 0x01 || nextHopDID == "EXIT_NODE_LOCAL_DELIVERY"

	// 4. Pelar una capa del payload con ChaCha20
	peeledPayload := packet.Payload
	streamPayload, err := chacha20.NewUnauthenticatedCipher(hopKey, make([]byte, 12))
	if err != nil {
		return nil, err
	}
	streamPayload.XORKeyStream(peeledPayload[:], peeledPayload[:])

	// 5. Determinar acción: DELIVER (nodo de salida) o FORWARD (nodo de paso)
	if isExit {
		// Extraer carga útil y servicio destino
		var serviceLen uint16
		r := bytes.NewReader(peeledPayload[:])
		_ = binary.Read(r, binary.BigEndian, &serviceLen)
		serviceBytes := make([]byte, serviceLen)
		_, _ = r.Read(serviceBytes)

		var payloadLen uint32
		_ = binary.Read(r, binary.BigEndian, &payloadLen)
		if payloadLen > uint32(SphinxPayloadSize) {
			payloadLen = uint32(SphinxPayloadSize - int(serviceLen) - 6)
		}
		rawInner := make([]byte, payloadLen)
		_, _ = r.Read(rawInner)

		return &PeelResult{
			Action:     SphinxActionDeliver,
			CircuitID:  circuitID,
			NextHopDID: "LOCAL",
			RawPayload: rawInner,
			IsExit:     true,
			TagHash:    tagHash,
			PeelTimeUs: time.Since(start).Microseconds(),
		}, nil
	}

	// 6. Nodo intermedio: Desplazar KeyHeader y RoutingHeader y rellenar con ruido pseudoaleatorio
	var nextKeyHeader [SphinxKeyHeaderSize]byte
	copy(nextKeyHeader[:], packet.KeyHeader[SphinxHopKEMSize:])
	noiseKey := make([]byte, SphinxHopKEMSize)
	io.ReadFull(rand.Reader, noiseKey)
	copy(nextKeyHeader[SphinxKeyHeaderSize-SphinxHopKEMSize:], noiseKey)

	var nextRoutingHeader [SphinxRoutingTotalSize]byte
	copy(nextRoutingHeader[:], packet.RoutingHeader[SphinxHopDescriptorSize:])
	noiseRouting := make([]byte, SphinxHopDescriptorSize)
	io.ReadFull(rand.Reader, noiseRouting)
	copy(nextRoutingHeader[SphinxRoutingTotalSize-SphinxHopDescriptorSize:], noiseRouting)

	nextPacket := &SphinxPacket{
		KeyHeader:     nextKeyHeader,
		RoutingHeader: nextRoutingHeader,
		Payload:       peeledPayload,
	}

	return &PeelResult{
		Action:     SphinxActionForward,
		CircuitID:  circuitID,
		NextHopDID: nextHopDID,
		NextPacket: nextPacket,
		IsExit:     false,
		TagHash:    tagHash,
		PeelTimeUs: time.Since(start).Microseconds(),
	}, nil
}


