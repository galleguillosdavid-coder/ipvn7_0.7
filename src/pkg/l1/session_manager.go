package l1

import (
	"crypto/ecdh"
	"crypto/mlkem"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"ipvn7/pkg/l0"
)

var (
	ErrNoActiveSession    = errors.New("pqc: no existe sesión activa para el DID")
	ErrInvalidSignature   = errors.New("pqc: firma digital inválida")
	ErrHandshakeMalformed = errors.New("pqc: datagrama de handshake malformado")
)

const rawKEMPayloadSize = 32 + 16 + mlkem.CiphertextSize768 // 1136 bytes

// PQCSessionManager gestiona el ciclo de vida de sesiones PQC híbridas en el camino de datos
type PQCSessionManager struct {
	mu              sync.RWMutex
	identity        *l0.Identity
	localKeys       *HybridKeyPair
	firewall        *ZTNAFirewall
	sessions        map[string]*l0.SessionKeys
	pendingSessions map[string]*l0.SessionKeys
	seqCounter      atomic.Uint64
}

// NewPQCSessionManager inicializa el gestor de sesiones PQC
func NewPQCSessionManager(id *l0.Identity, keys *HybridKeyPair, fw *ZTNAFirewall) *PQCSessionManager {
	return &PQCSessionManager{
		identity:        id,
		localKeys:       keys,
		firewall:        fw,
		sessions:        make(map[string]*l0.SessionKeys),
		pendingSessions: make(map[string]*l0.SessionKeys),
	}
}

// HasSession consulta si existe una sesión simétrica activa con el par
func (m *PQCSessionManager) HasSession(did string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.sessions[did]
	return exists
}

// GetSession obtiene las claves simétricas de la sesión activa
func (m *PQCSessionManager) GetSession(did string) (*l0.SessionKeys, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, exists := m.sessions[did]
	return s, exists
}

// SetSession almacena directamente una sesión verificada
func (m *PQCSessionManager) SetSession(did string, keys *l0.SessionKeys) {
	m.mu.Lock()
	m.sessions[did] = keys
	m.mu.Unlock()
	if m.firewall != nil {
		m.firewall.AuthorizeDID(&DIDPolicy{
			DID:           did,
			AllowInbound:  true,
			AllowOutbound: true,
			AllowRelay:    true,
		})
	}
}

// CreateHandshakeInitPacket inicia un intercambio 1-RTT con encapsulación KEM ML-KEM-768
func (m *PQCSessionManager) CreateHandshakeInitPacket(targetDID string, targetX25519Pub *ecdh.PublicKey, targetMLKEMHex string) (*l0.Packet, error) {
	if m.localKeys == nil {
		return nil, errors.New("claves híbridas locales no inicializadas")
	}

	initMsg, provisionalKeys, err := Initiate1RTT(m.localKeys, targetX25519Pub, targetMLKEMHex, targetDID)
	if err != nil {
		return nil, fmt.Errorf("error iniciando handshake 1-RTT: %w", err)
	}

	// Serialización binaria directa de 1136 bytes para cumplir el MTU de 1280B
	rawKEM := make([]byte, rawKEMPayloadSize)
	copy(rawKEM[0:32], initMsg.Ciphertext.EphemeralX25519)
	copy(rawKEM[32:48], initMsg.Ciphertext.Salt)
	copy(rawKEM[48:], initMsg.Ciphertext.FullPQCCiphertext)

	m.mu.Lock()
	m.pendingSessions[targetDID] = provisionalKeys
	m.mu.Unlock()

	seq := m.seqCounter.Add(1)
	pkt := l0.NewPacket(l0.MsgTypeHandshakeInit, m.identity.DID(), "", seq, nil, rawKEM)
	return pkt, nil
}

// HandleHandshakeInitPacket procesa la solicitud de inicio, decapsula KEM y emite respuesta
func (m *PQCSessionManager) HandleHandshakeInitPacket(pkt *l0.Packet) (*l0.Packet, error) {
	if len(pkt.Payload) != rawKEMPayloadSize {
		return nil, fmt.Errorf("%w: tamaño KEM inválido (%d)", ErrHandshakeMalformed, len(pkt.Payload))
	}

	ephPub := make([]byte, 32)
	copy(ephPub, pkt.Payload[0:32])
	salt := make([]byte, 16)
	copy(salt, pkt.Payload[32:48])
	fullCipher := make([]byte, mlkem.CiphertextSize768)
	copy(fullCipher, pkt.Payload[48:])

	initMsg := Handshake1RTTInitiation{
		SenderDID:    pkt.SourceDID,
		RecipientDID: m.identity.DID(),
		Ciphertext: &HybridKEMCiphertext{
			Algorithm:         HybridKEMAlgorithm,
			EphemeralX25519:   ephPub,
			Salt:              salt,
			FullPQCCiphertext: fullCipher,
		},
		Timestamp: pkt.Timestamp,
	}

	respMsg, sessionKeys, err := Respond1RTT(m.localKeys, &initMsg)
	if err != nil {
		return nil, fmt.Errorf("error respondiendo handshake 1-RTT: %w", err)
	}

	m.mu.Lock()
	m.sessions[pkt.SourceDID] = sessionKeys
	m.mu.Unlock()

	if m.firewall != nil {
		m.firewall.AuthorizeDID(&DIDPolicy{
			DID:           pkt.SourceDID,
			AllowInbound:  true,
			AllowOutbound: true,
			AllowRelay:    true,
		})
	}

	respPayload, err := json.Marshal(respMsg)
	if err != nil {
		return nil, err
	}

	seq := m.seqCounter.Add(1)
	respPkt := l0.NewPacket(l0.MsgTypeHandshakeResp, m.identity.DID(), pkt.SourceDID, seq, nil, respPayload)
	if err := respPkt.SignPacket(m.identity); err != nil {
		return nil, err
	}

	return respPkt, nil
}

// HandleHandshakeRespPacket finaliza el intercambio 1-RTT en el iniciador
func (m *PQCSessionManager) HandleHandshakeRespPacket(pkt *l0.Packet) error {
	valid, err := pkt.VerifyPacketSignature()
	if err != nil || !valid {
		return ErrInvalidSignature
	}

	var respMsg Handshake1RTTResponse
	if err := json.Unmarshal(pkt.Payload, &respMsg); err != nil {
		return fmt.Errorf("%w: %v", ErrHandshakeMalformed, err)
	}

	m.mu.Lock()
	pending, exists := m.pendingSessions[pkt.SourceDID]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("no existe handshake pendiente para %s", pkt.SourceDID)
	}
	delete(m.pendingSessions, pkt.SourceDID)
	m.mu.Unlock()

	if err := Finalize1RTT(pending, &respMsg); err != nil {
		return fmt.Errorf("error finalizando handshake: %w", err)
	}

	m.mu.Lock()
	m.sessions[pkt.SourceDID] = pending
	m.mu.Unlock()

	if m.firewall != nil {
		m.firewall.AuthorizeDID(&DIDPolicy{
			DID:           pkt.SourceDID,
			AllowInbound:  true,
			AllowOutbound: true,
			AllowRelay:    true,
		})
	}

	return nil
}

// EncryptDataPacket cifra los datos útiles con ChaCha20-Poly1305 bajo la sesión PQC
func (m *PQCSessionManager) EncryptDataPacket(destDID string, plaintext []byte) (*l0.Packet, error) {
	session, exists := m.GetSession(destDID)
	if !exists {
		return nil, ErrNoActiveSession
	}

	nonce, err := l0.GenerateNonce()
	if err != nil {
		return nil, err
	}

	ciphertext, err := l0.EncryptPayload(session.TxKey[:], nonce, plaintext, []byte(destDID))
	if err != nil {
		return nil, fmt.Errorf("error cifrando payload AEAD: %w", err)
	}

	seq := m.seqCounter.Add(1)
	pkt := l0.NewPacket(l0.MsgTypeData, m.identity.DID(), destDID, seq, nonce, ciphertext)
	return pkt, nil
}

// DecryptDataPacket descifra los datos útiles con ChaCha20-Poly1305 bajo la sesión PQC
func (m *PQCSessionManager) DecryptDataPacket(pkt *l0.Packet) ([]byte, error) {
	session, exists := m.GetSession(pkt.SourceDID)
	if !exists {
		return nil, ErrNoActiveSession
	}

	plaintext, err := l0.DecryptPayload(session.RxKey[:], pkt.Nonce, pkt.Payload, []byte(pkt.DestDID))
	if err != nil {
		return nil, fmt.Errorf("error descifrando payload AEAD: %w", err)
	}

	return plaintext, nil
}
