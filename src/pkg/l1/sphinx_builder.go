package l1

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"

	"golang.org/x/crypto/chacha20"

	"ipvn7/pkg/l0"
)

// BuildDynamicCircuit genera un circuito Sphinx de 3 saltos seleccionando pares activos de la topología Kleinberg.
// Si el nodo opera en solitario, deriva saltos soberanos anclados a la identidad local y sus anillos.
func (sr *SphinxRouter) BuildDynamicCircuit(router *KleinbergRouter, localIdentity *l0.Identity) (*SphinxCircuit, error) {
	var peers []*PeerNode
	if router != nil {
		peers = router.GetAllPeers()
	}

	hops := make([]*SphinxHopNode, 3)
	roles := []string{"guard", "middle", "exit"}

	for i := 0; i < 3; i++ {
		var did, addr string
		if i < len(peers) && peers[i] != nil {
			did = peers[i].DID
			if peers[i].Locator.PhysicalAddr != nil {
				addr = peers[i].Locator.PhysicalAddr.String()
			} else {
				addr = fmt.Sprintf("127.0.0.1:%d", 7770+i)
			}
		} else {
			prefix := "did:ipvn7:node"
			if localIdentity != nil {
				prefix = localIdentity.DID()
			}
			did = fmt.Sprintf("%s:ring-%d", prefix, (i+1)*4)
			addr = fmt.Sprintf("127.0.0.1:%d", 7780+i)
		}

		hKeys, err := GenerateHybridKeyPair(did)
		if err != nil {
			return nil, fmt.Errorf("fallo generando claves híbridas para salto %s: %w", roles[i], err)
		}

		hops[i] = &SphinxHopNode{
			DID:          did,
			Address:      addr,
			HybridKeys:   hKeys,
			X25519PubHex: hKeys.X25519PubHex,
			MLKEMPubHex:  hKeys.MLKEMPubHex,
		}
	}

	return &SphinxCircuit{
		CircuitID: fmt.Sprintf("circ-%d", time.Now().UnixNano()%100000),
		Guard:     hops[0],
		Middle:    hops[1],
		Exit:      hops[2],
		CreatedAt: time.Now().UTC(),
	}, nil
}

// BuildPacket construye un paquete cebolla determinista de 1280 bytes a través de los 3 saltos
func (sr *SphinxRouter) BuildPacket(circuit *SphinxCircuit, finalPayload []byte, destService string) (*SphinxPacket, error) {
	if circuit == nil || circuit.Guard == nil || circuit.Middle == nil || circuit.Exit == nil {
		return nil, errors.New("circuito incompleto: se requieren 3 saltos (Guard, Middle, Exit)")
	}

	hops := []*SphinxHopNode{circuit.Guard, circuit.Middle, circuit.Exit}

	// 1. Derivar claves de sesión simétricas para cada salto usando KEM híbrido
	hopKeys := make([][]byte, SphinxHopsCount)
	var keyHeader [SphinxKeyHeaderSize]byte

	for i, hop := range hops {
		sharedKey, kemCipher, err := EncapsulateCompact(hop.HybridKeys.ClassicalKEMPub, hop.HybridKeys.MLKEMPubHex)
		if err != nil {
			return nil, fmt.Errorf("error KEM en salto %d (%s): %w", i, hop.DID, err)
		}
		hopKeys[i] = sharedKey

		// Copiar KEM efímero en la posición correspondiente del encabezado de claves
		slotStart := i * SphinxHopKEMSize
		copy(keyHeader[slotStart:slotStart+32], kemCipher.EphemeralX25519)
		copy(keyHeader[slotStart+32:slotStart+160], kemCipher.PQCCiphertext)
	}

	// 2. Preparar el bloque de carga útil (exactamente SphinxPayloadSize bytes)
	var payloadBlock [SphinxPayloadSize]byte
	if len(finalPayload) > SphinxPayloadSize-32 {
		return nil, fmt.Errorf("payload excede el límite del frame (%d > %d bytes)", len(finalPayload), SphinxPayloadSize-32)
	}

	// Codificar metadatos de destino en el inicio del payload interno
	var innerBuf bytes.Buffer
	binary.Write(&innerBuf, binary.BigEndian, uint16(len(destService)))
	innerBuf.WriteString(destService)
	binary.Write(&innerBuf, binary.BigEndian, uint32(len(finalPayload)))
	innerBuf.Write(finalPayload)

	copy(payloadBlock[:], innerBuf.Bytes())

	// Rellenar el resto del payload con bytes pseudoaleatorios deterministas (zero-leak)
	if _, err := io.ReadFull(rand.Reader, payloadBlock[innerBuf.Len():]); err != nil {
		return nil, err
	}

	// 3. Encriptar el payload en capas inversas: Exit (hop 2) -> Middle (hop 1) -> Guard (hop 0)
	for i := SphinxHopsCount - 1; i >= 0; i-- {
		streamCipher, err := chacha20.NewUnauthenticatedCipher(hopKeys[i], make([]byte, 12))
		if err != nil {
			return nil, err
		}
		streamCipher.XORKeyStream(payloadBlock[:], payloadBlock[:])
	}

	// 4. Construir descriptores de enrutamiento para cada salto
	var routingHeader [SphinxRoutingTotalSize]byte

	// Salto 2 (Exit): Próximo destino es LOCAL
	descExit := make([]byte, SphinxHopDescriptorSize)
	copy(descExit[0:32], []byte("EXIT_NODE_LOCAL_DELIVERY"))
	copy(descExit[32:48], []byte(circuit.CircuitID))
	descExit[48] = 0x01 // Flag EXIT

	// Salto 1 (Middle): Próximo destino es Exit
	descMiddle := make([]byte, SphinxHopDescriptorSize)
	copy(descMiddle[0:32], []byte(circuit.Exit.DID))
	copy(descMiddle[32:48], []byte(circuit.CircuitID))
	descMiddle[48] = 0x00 // Flag RELAY

	// Salto 0 (Guard): Próximo destino es Middle
	descGuard := make([]byte, SphinxHopDescriptorSize)
	copy(descGuard[0:32], []byte(circuit.Middle.DID))
	copy(descGuard[32:48], []byte(circuit.CircuitID))
	descGuard[48] = 0x00 // Flag RELAY

	descriptors := [][]byte{descGuard, descMiddle, descExit}

	// Cifrar cada descriptor con la clave simétrica del salto respectivo
	for i := 0; i < SphinxHopsCount; i++ {
		// Calcular MAC sobre descriptor
		mac := hmac.New(sha256.New, hopKeys[i])
		mac.Write(descriptors[i][:49])
		macSum := mac.Sum(nil)
		copy(descriptors[i][49:65], macSum[:16])

		// Cifrar con ChaCha20
		stream, err := chacha20.NewUnauthenticatedCipher(hopKeys[i], make([]byte, 12))
		if err != nil {
			return nil, err
		}
		stream.XORKeyStream(descriptors[i], descriptors[i])
		copy(routingHeader[i*SphinxHopDescriptorSize:(i+1)*SphinxHopDescriptorSize], descriptors[i])
	}

	packet := &SphinxPacket{
		KeyHeader:     keyHeader,
		RoutingHeader: routingHeader,
		Payload:       payloadBlock,
	}

	return packet, nil
}
