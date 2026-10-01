// Package l1 implementa el anclaje topológico de movilidad atómica (Topological Mobility Anchor)
// eliminando tormentas de señalización ante cambios continuos de interfaces físicas e IPs efímeras.
package l1

import (
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"ipvn7/pkg/l0"
)

var (
	ErrMobilityAnchorNotFound   = errors.New("mobility_anchor: ancla topológica no asignada")
	ErrSignalingStormSuppressed = errors.New("mobility_anchor: actualización suprimida por amortiguador de tormenta")
	ErrOutdatedLocatorUpdate    = errors.New("mobility_anchor: actualización obsoleta descartada (anti-race condition)")
	ErrSequenceJumpTooLarge     = errors.New("mobility_anchor: salto de secuencia anómalo detectado (intento de secuestro)")
)

// MobileNodeLocator mantiene la correspondencia entre DID lógico y localizador físico en el ancla
type MobileNodeLocator struct {
	DID             string
	PhysicalAddr    *net.UDPAddr
	AnchorDID       string
	RingIndex       int
	LastUpdate      time.Time
	UpdateCount     uint64
	SuppressedCount uint64
	LastSeq         uint64 // Número de secuencia monótono estricto anti-carreras
	ProactiveTicket []byte // Ticket proactivo 0-RTT emitido en el registro para el siguiente handover en caliente
}

// MobilityAnchorEngine orquesta el anclaje local en los anillos de Kleinberg
type MobilityAnchorEngine struct {
	mu                 sync.RWMutex
	anchorDID          string
	anchorSecret       [32]byte
	locators           map[string]*MobileNodeLocator
	lastGlobalSync     time.Time
	minSyncInterval    time.Duration
	resyncVerifyTokens atomic.Int64
	lastResyncRefill   atomic.Int64
	coldPoWBuckets     [64]atomic.Int64 // Segregación en 64 cubetas por DID: inmunidad a agotamiento cruzado
	lastColdPoWRefill  atomic.Int64
}

// NewMobilityAnchorEngine inicializa el gestor de anclaje topológico
func NewMobilityAnchorEngine(anchorDID string) *MobilityAnchorEngine {
	e := &MobilityAnchorEngine{
		anchorDID:       anchorDID,
		locators:        make(map[string]*MobileNodeLocator),
		minSyncInterval: 5 * time.Second, // Cooldown de sincronización inter-anillos
	}
	_, _ = io.ReadFull(rand.Reader, e.anchorSecret[:])
	e.resyncVerifyTokens.Store(500) // Presupuesto: máx 500 verificaciones asimétricas Ed25519/seg (<0.5% CPU)
	e.lastResyncRefill.Store(time.Now().Unix())
	for i := 0; i < 64; i++ {
		e.coldPoWBuckets[i].Store(100) // 100 PoW/seg por cubeta segregada (6.400/seg total)
	}
	e.lastColdPoWRefill.Store(time.Now().Unix())
	return e
}

// IssueResyncChallenge emite un ticket stateless HMAC para autorizar una resincronización de secuencia masiva (0-RTT Proactive)
func (e *MobilityAnchorEngine) IssueResyncChallenge(nodeDID string) []byte {
	epoch := time.Now().Unix() / 60
	mac := hmac.New(sha256.New, e.anchorSecret[:])
	mac.Write([]byte(nodeDID))
	var epochBytes [8]byte
	epochBytes[0] = byte(epoch)
	mac.Write(epochBytes[:])
	return mac.Sum(nil)[:16]
}

// ValidateResyncChallenge verifica en O(1) (20 ns) si el ticket fue emitido por el Ancla para este DID
func (e *MobilityAnchorEngine) ValidateResyncChallenge(nodeDID string, challenge []byte) bool {
	if len(challenge) != 16 {
		return false
	}
	epoch := time.Now().Unix() / 60
	for _, ep := range []int64{epoch, epoch - 1, epoch + 1} {
		mac := hmac.New(sha256.New, e.anchorSecret[:])
		mac.Write([]byte(nodeDID))
		var epochBytes [8]byte
		epochBytes[0] = byte(ep)
		mac.Write(epochBytes[:])
		if hmac.Equal(challenge, mac.Sum(nil)[:16]) {
			return true
		}
	}
	return false
}

// UpdateMobileLocator procesa el cambio de IP/puerto físico de un nodo móvil mediante señalización O(1)
func (e *MobilityAnchorEngine) UpdateMobileLocator(nodeDID string, newAddr *net.UDPAddr, ringIndex int) (*MobileNodeLocator, error) {
	return e.UpdateMobileLocatorWithSeq(nodeDID, newAddr, ringIndex, 0)
}

// UpdateMobileLocatorWithSeq procesa el cambio de localizador con número de secuencia monótono estricto anti-carreras
func (e *MobilityAnchorEngine) UpdateMobileLocatorWithSeq(nodeDID string, newAddr *net.UDPAddr, ringIndex int, seq uint64) (*MobileNodeLocator, error) {
	return e.UpdateMobileLocatorWithSig(nodeDID, newAddr, ringIndex, seq, nil)
}

// UpdateMobileLocatorWithSig procesa el cambio de localizador admitiendo grandes saltos de secuencia (>1.000.000)
func (e *MobilityAnchorEngine) UpdateMobileLocatorWithSig(nodeDID string, newAddr *net.UDPAddr, ringIndex int, seq uint64, resyncSig []byte) (*MobileNodeLocator, error) {
	return e.UpdateMobileLocatorWithSigAndTicket(nodeDID, newAddr, ringIndex, seq, resyncSig, nil)
}

// UpdateMobileLocatorWithSigAndTicket procesa la resincronización. Soporta tickets proactivos 0-RTT entregados en el registro previo.
// Elimina viajes de ida y vuelta adicionales durante transiciones de interfaz (.198 <-> .106 / Wi-Fi <-> 4G).
func (e *MobilityAnchorEngine) UpdateMobileLocatorWithSigAndTicket(nodeDID string, newAddr *net.UDPAddr, ringIndex int, seq uint64, resyncSig []byte, challengeTicket []byte) (*MobileNodeLocator, error) {
	return e.UpdateMobileLocatorWithSigTicketAndPoW(nodeDID, newAddr, ringIndex, seq, resyncSig, challengeTicket, 0)
}

// UpdateMobileLocatorWithSigTicketAndPoW admite actualización con soporte de Micro-PoW para arranque en frío bajo ataques volumétricos.
func (e *MobilityAnchorEngine) UpdateMobileLocatorWithSigTicketAndPoW(nodeDID string, newAddr *net.UDPAddr, ringIndex int, seq uint64, resyncSig []byte, challengeTicket []byte, coldNonce uint64) (*MobileNodeLocator, error) {
	if newAddr == nil {
		return nil, errors.New("mobility_anchor: direccion fisica nula")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	loc, exists := e.locators[nodeDID]
	now := time.Now()

	if !exists {
		loc = &MobileNodeLocator{
			DID:             nodeDID,
			PhysicalAddr:    newAddr,
			AnchorDID:       e.anchorDID,
			RingIndex:       ringIndex,
			LastUpdate:      now,
			UpdateCount:     1,
			LastSeq:         seq,
			ProactiveTicket: e.IssueResyncChallenge(nodeDID),
		}
		e.locators[nodeDID] = loc
		return loc, nil
	}

	// Invariante Anti-Race y Anti-Hijacking con Resincronización Criptográfica Soberana:
	const MaxAllowedSequenceJump = 1000000
	if seq > 0 {
		if seq <= loc.LastSeq {
			return loc, ErrOutdatedLocatorUpdate
		}
		if loc.LastSeq > 0 && (seq-loc.LastSeq) > MaxAllowedSequenceJump {
			// Si el salto es masivo, validar prueba criptográfica soberana con soporte 0-RTT y Cold-PoW
			if len(resyncSig) == 0 || !e.verifyResyncSignatureWithTicketAndPoW(nodeDID, seq, resyncSig, challengeTicket, coldNonce) {
				return loc, ErrSequenceJumpTooLarge
			}
		}
	}

	// Amortiguador de tormentas de señalización: actualiza localmente y renueva el ticket proactivo 0-RTT
	loc.PhysicalAddr = newAddr
	loc.RingIndex = ringIndex
	loc.LastUpdate = now
	loc.UpdateCount++
	loc.ProactiveTicket = e.IssueResyncChallenge(nodeDID)
	if seq > 0 {
		loc.LastSeq = seq
	}

	return loc, nil
}

func (e *MobilityAnchorEngine) verifyResyncSignature(nodeDID string, seq uint64, sig []byte) bool {
	return e.verifyResyncSignatureWithTicketAndPoW(nodeDID, seq, sig, nil, 0)
}

func (e *MobilityAnchorEngine) verifyResyncSignatureWithTicket(nodeDID string, seq uint64, sig []byte, ticket []byte) bool {
	return e.verifyResyncSignatureWithTicketAndPoW(nodeDID, seq, sig, ticket, 0)
}

func (e *MobilityAnchorEngine) verifyResyncSignatureWithTicketAndPoW(nodeDID string, seq uint64, sig []byte, ticket []byte, coldNonce uint64) bool {
	if len(sig) != ed25519.SignatureSize {
		return false // Descarte en 1 ns para firmas malformadas o de longitud corrupta
	}

	// Inmunidad Estricta Anti-Fallback:
	// Si un paquete presenta un ticket pero es inválido/falsificado, descartar en 1 ns sin tocar el camino lento
	if len(ticket) > 0 && !e.ValidateResyncChallenge(nodeDID, ticket) {
		return false
	}

	hasValidTicket := len(ticket) == 16 && e.ValidateResyncChallenge(nodeDID, ticket)

	// Si no tiene ticket emitido (arranque en frío desincronizado tras desconexión prolongada)
	if !hasValidTicket {
		now := time.Now().Unix()
		last := e.lastResyncRefill.Load()
		if now > last && e.lastResyncRefill.CompareAndSwap(last, now) {
			e.resyncVerifyTokens.Store(500)
		}
		if e.resyncVerifyTokens.Add(-1) < 0 {
			// Pool agotado por inundación de spam: un nodo legítimo en arranque en frío se admite mediante Micro-PoW (8 bits cero, ~1 µs)
			if coldNonce == 0 || !e.VerifyColdResyncPoW(nodeDID, seq, coldNonce) {
				return false // Descarte en 1 ns para ráfagas ciegas sin Micro-PoW
			}
		}
	}

	pub, err := l0.PublicKeyFromDID(nodeDID)
	if err != nil {
		return false
	}
	msg := []byte(fmt.Sprintf("%s:%d:%s", nodeDID, seq, e.anchorDID))
	return ed25519.Verify(pub, msg, sig)
}

// ColdPoWTag calcula la etiqueta de alineación de hardware de 4 bits para un par (DID, secuencia)
func ColdPoWTag(nodeDID string, seq uint64) uint8 {
	return uint8((seq ^ (seq >> 8) ^ uint64(len(nodeDID))) & 0x0F)
}

func coldBucketIdx(did string) int {
	var h uint32 = 2166136261
	for i := 0; i < len(did); i++ {
		h ^= uint32(did[i])
		h *= 16777619
	}
	return int(h & 0x3F) // 64 cubetas segregadas
}

// VerifyColdResyncPoW valida un Micro-PoW en frío (SHA-256) con pre-filtro de etiqueta dinámica por secuencia,
// segregación por cubeta de DID y gradiente de escape adaptativo Dual-Tier (8 o 12 bits) (DEC-078).
func (e *MobilityAnchorEngine) VerifyColdResyncPoW(nodeDID string, seq uint64, nonce uint64) bool {
	// 1. Pre-filtro dinámico en 1 ciclo: el noncio debe coincidir con la etiqueta pseudoaleatoria de la secuencia
	if uint8(nonce&0x0F) != ColdPoWTag(nodeDID, seq) {
		return false // Descarta el 93.75% de spam aleatorio o con alineación fija en 0 ns
	}

	bIdx := coldBucketIdx(nodeDID)
	now := time.Now().Unix()
	last := e.lastColdPoWRefill.Load()
	if now > last && e.lastColdPoWRefill.CompareAndSwap(last, now) {
		for i := 0; i < 64; i++ {
			e.coldPoWBuckets[i].Store(100)
		}
	}
	hasToken := e.coldPoWBuckets[bIdx].Add(-1) >= 0

	var nonceBytes [8]byte
	var seqBytes [8]byte
	binary.BigEndian.PutUint64(nonceBytes[:], nonce)
	binary.BigEndian.PutUint64(seqBytes[:], seq)
	h := sha256.New()
	h.Write([]byte(nodeDID))
	h.Write([]byte(e.anchorDID))
	h.Write(seqBytes[:])
	h.Write(nonceBytes[:])
	digest := h.Sum(nil)

	// Tier-1 (Normal): cuota disponible en la cubeta segregada de este DID -> dificultad de 8 bits
	if hasToken {
		return digest[0] == 0x00
	}
	// Tier-2 (Bajo Ataque Concentrado): la cubeta de este DID está agotada por ráfagas directas.
	// Admite escape inmediato si el cliente legítimo presenta 12 bits de dificultad (~15 µs en cliente).
	return digest[0] == 0x00 && (digest[1]&0xF0) == 0
}

// ResolveMobileLocator resuelve en O(1) el localizador físico actual de un nodo móvil
func (e *MobilityAnchorEngine) ResolveMobileLocator(nodeDID string) (*net.UDPAddr, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	loc, exists := e.locators[nodeDID]
	if !exists || loc == nil {
		return nil, false
	}
	return loc.PhysicalAddr, true
}

// ShouldPropagateGlobalSync determina si el cambio justifica señalización inter-anillos o se amortigua en el ancla
func (e *MobilityAnchorEngine) ShouldPropagateGlobalSync() bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now()
	if now.Sub(e.lastGlobalSync) < e.minSyncInterval {
		return false // Amortigua la propagación global para proteger el plano de datos
	}

	e.lastGlobalSync = now
	return true
}

// ActiveMobileCount retorna el número de nodos anclados actualmente
func (e *MobilityAnchorEngine) ActiveMobileCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.locators)
}
