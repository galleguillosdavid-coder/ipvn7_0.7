package l1

import (
	"crypto/ecdh"
	"crypto/mlkem"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"ipvn7/pkg/l0"
)

var (
	ErrNoActiveSession    = errors.New("pqc: no existe sesión activa para el DID")
	ErrInvalidSignature   = errors.New("pqc: firma digital inválida")
	ErrHandshakeMalformed = errors.New("pqc: datagrama de handshake malformado")
)

// rawKEMPayloadSize define el tamaño canónico del criptograma híbrido (32B X25519 + 1088B ML-KEM-768 = 1120B)
// Conforme a la especificación IETF CFRG X-Wing, garantizando que el datagrama completo con firma Ed25519 <= 1280B
const rawKEMPayloadSize = 32 + mlkem.CiphertextSize768 // 1120 bytes

// PendingHandshake almacena el estado de negociación pendiente en el iniciador
type PendingHandshake struct {
	TargetDID   string
	SessionKeys *l0.SessionKeys
	CreatedAt   int64
}

// PQCSessionManager gestiona el ciclo de vida de sesiones PQC híbridas en el camino de datos
type PQCSessionManager struct {
	mu              sync.RWMutex
	identity        *l0.Identity
	localKeys       *HybridKeyPair
	firewall        *ZTNAFirewall
	sessions        map[string]*l0.SessionKeys
	pendingSessions map[string]*PendingHandshake
	seqCounter      atomic.Uint64
}

// NewPQCSessionManager inicializa el gestor de sesiones PQC
func NewPQCSessionManager(id *l0.Identity, keys *HybridKeyPair, fw *ZTNAFirewall) *PQCSessionManager {
	return &PQCSessionManager{
		identity:        id,
		localKeys:       keys,
		firewall:        fw,
		sessions:        make(map[string]*l0.SessionKeys),
		pendingSessions: make(map[string]*PendingHandshake),
	}
}

// HasSession consulta si existe una sesión simétrica activa con el par
func (m *PQCSessionManager) HasSession(did string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if _, exists := m.sessions[did]; exists {
		return true
	}
	_, exists := m.sessions[l0.CanonicalDID(did)]
	return exists
}

// GetSession obtiene las claves simétricas de la sesión activa
func (m *PQCSessionManager) GetSession(did string) (*l0.SessionKeys, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, exists := m.sessions[did]
	if !exists {
		s, exists = m.sessions[l0.CanonicalDID(did)]
	}
	return s, exists
}

// SetSession almacena directamente una sesión verificada
func (m *PQCSessionManager) SetSession(did string, keys *l0.SessionKeys) {
	canonicalDID := l0.CanonicalDID(did)
	m.mu.Lock()
	m.sessions[canonicalDID] = keys
	m.mu.Unlock()
	if m.firewall != nil {
		m.firewall.AuthorizeDID(&DIDPolicy{
			DID:           canonicalDID,
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

	// Serialización binaria canónica de 1120 bytes para cumplir estrictamente el MTU determinista de 1280B
	rawKEM := make([]byte, rawKEMPayloadSize)
	copy(rawKEM[0:32], initMsg.Ciphertext.EphemeralX25519)
	copy(rawKEM[32:], initMsg.Ciphertext.FullPQCCiphertext)

	canonicalTarget := l0.CanonicalDID(targetDID)
	m.mu.Lock()
	m.pendingSessions[canonicalTarget] = &PendingHandshake{
		TargetDID:   canonicalTarget,
		SessionKeys: provisionalKeys,
		CreatedAt:   time.Now().Unix(),
	}
	m.mu.Unlock()

	seq := m.seqCounter.Add(1)
	// Emplear clave pública compacta de 64 caracteres hex y timestamp UNIX en segundos para respetar MTU de 1280B
	compactSourceDID := hex.EncodeToString(m.identity.PublicKey)
	pkt := l0.NewPacket(l0.MsgTypeHandshakeInit, compactSourceDID, "", seq, nil, rawKEM)
	pkt.Timestamp = time.Now().Unix()
	if err := pkt.SignPacket(m.identity); err != nil {
		return nil, fmt.Errorf("error firmando HandshakeInit con Ed25519: %w", err)
	}
	return pkt, nil
}

// HandleHandshakeInitPacket procesa la solicitud de inicio, verifica firma Ed25519, decapsula KEM y emite respuesta
func (m *PQCSessionManager) HandleHandshakeInitPacket(pkt *l0.Packet) (*l0.Packet, error) {
	if len(pkt.Payload) != rawKEMPayloadSize {
		return nil, fmt.Errorf("%w: tamaño KEM inválido (%d)", ErrHandshakeMalformed, len(pkt.Payload))
	}

	// Verificar destino si está especificado
	if pkt.DestDID != "" && pkt.DestDID != m.identity.DID() {
		return nil, fmt.Errorf("pqc: DestDID (%s) no coincide con identidad local (%s)", pkt.DestDID, m.identity.DID())
	}

	// 1. Validar correspondencia sintáctica del SourceDID con su clave pública
	if _, err := l0.PublicKeyFromDID(pkt.SourceDID); err != nil {
		return nil, fmt.Errorf("pqc: SourceDID malformado: %w", err)
	}

	// 2. CRÍTICO: Verificar firma Ed25519 obligatoria del iniciador sobre el paquete
	valid, err := pkt.VerifyPacketSignature()
	if err != nil || !valid {
		return nil, fmt.Errorf("%w: firma digital de %s no coincide o no es válida", ErrInvalidSignature, pkt.SourceDID)
	}

	// 3. Verificar ventana temporal de frescura (máximo 60 segundos)
	nowSec := time.Now().Unix()
	tsSec := pkt.Timestamp
	if tsSec > 1e14 {
		tsSec = tsSec / 1e9
	}
	diff := nowSec - tsSec
	if diff > 60 || diff < -10 {
		return nil, fmt.Errorf("pqc: timestamp del handshake expirado o desfasado (diff: %d s)", diff)
	}

	ephPub := make([]byte, 32)
	copy(ephPub, pkt.Payload[0:32])
	fullCipher := make([]byte, mlkem.CiphertextSize768)
	copy(fullCipher, pkt.Payload[32:])

	// Sal determinista derivada de EphemeralX25519 y cabecera de FullPQCCiphertext
	hSalt := sha256.New()
	hSalt.Write(ephPub)
	hSalt.Write(fullCipher[:32])
	salt := hSalt.Sum(nil)[:16]

	canonicalSourceDID := l0.CanonicalDID(pkt.SourceDID)

	initMsg := Handshake1RTTInitiation{
		SenderDID:    canonicalSourceDID,
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

	// Registrar sesión y autorizar DID en firewall SOLO tras verificar firma y KEM
	m.mu.Lock()
	m.sessions[canonicalSourceDID] = sessionKeys
	m.mu.Unlock()

	if m.firewall != nil {
		m.firewall.AuthorizeDID(&DIDPolicy{
			DID:           canonicalSourceDID,
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
	respPkt := l0.NewPacket(l0.MsgTypeHandshakeResp, m.identity.DID(), canonicalSourceDID, seq, nil, respPayload)
	if err := respPkt.SignPacket(m.identity); err != nil {
		return nil, err
	}

	return respPkt, nil
}

// HandleHandshakeRespPacket finaliza el intercambio 1-RTT en el iniciador tras verificar firma e identidad
func (m *PQCSessionManager) HandleHandshakeRespPacket(pkt *l0.Packet) error {
	// 1. Validar que la respuesta esté explícitamente destinada a este nodo
	if pkt.DestDID != m.identity.DID() {
		return fmt.Errorf("pqc: DestDID de respuesta (%s) no coincide con identidad local (%s)", pkt.DestDID, m.identity.DID())
	}

	// 2. Verificar firma Ed25519 del respondedor
	valid, err := pkt.VerifyPacketSignature()
	if err != nil || !valid {
		return ErrInvalidSignature
	}

	canonicalSourceDID := l0.CanonicalDID(pkt.SourceDID)

	// 3. Validar correspondencia con sesión pendiente
	m.mu.Lock()
	pending, exists := m.pendingSessions[canonicalSourceDID]
	if !exists {
		pending, exists = m.pendingSessions[pkt.SourceDID]
	}
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("no existe handshake pendiente para %s", pkt.SourceDID)
	}
	delete(m.pendingSessions, canonicalSourceDID)
	delete(m.pendingSessions, pkt.SourceDID)
	m.mu.Unlock()

	// 4. Validar expiración de la sesión pendiente (timeout de 60s)
	if time.Since(time.Unix(pending.CreatedAt, 0)) > 60*time.Second {
		return errors.New("pqc: sesión pendiente expirada (timeout)")
	}

	var respMsg Handshake1RTTResponse
	if err := json.Unmarshal(pkt.Payload, &respMsg); err != nil {
		return fmt.Errorf("%w: %v", ErrHandshakeMalformed, err)
	}

	if err := Finalize1RTT(pending.SessionKeys, &respMsg); err != nil {
		return fmt.Errorf("error finalizando handshake: %w", err)
	}

	m.mu.Lock()
	m.sessions[canonicalSourceDID] = pending.SessionKeys
	m.mu.Unlock()

	if m.firewall != nil {
		m.firewall.AuthorizeDID(&DIDPolicy{
			DID:           canonicalSourceDID,
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
