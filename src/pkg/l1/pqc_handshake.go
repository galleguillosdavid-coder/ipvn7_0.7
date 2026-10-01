package l1

import (
	"crypto/ecdh"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"ipvn7/pkg/l0"
)

// Handshake1RTTInitiation transporta la encapsulación KEM híbrida en el primer datagrama (0.5 RTT).
type Handshake1RTTInitiation struct {
	SenderDID    string               `json:"sender_did"`
	RecipientDID string               `json:"recipient_did"`
	Ciphertext   *HybridKEMCiphertext `json:"ciphertext"`
	Timestamp    int64                `json:"timestamp"`
	Cookie       []byte               `json:"cookie,omitempty"`
}

// Handshake1RTTResponse confirma la sesión y autentica la respuesta en 1 RTT completo.
type Handshake1RTTResponse struct {
	SessionID [16]byte `json:"session_id"`
	AuthTag   [16]byte `json:"auth_tag"`
	Timestamp int64    `json:"timestamp"`
}

// Initiate1RTT construye el mensaje inicial del handshake post-cuántico híbrido.
// Deriva inmediatamente las claves provisionales de sesión sin esperar RTT adicionales.
func Initiate1RTT(
	initiatorKeys *HybridKeyPair,
	targetX25519Pub *ecdh.PublicKey,
	targetMLKEMPubHex string,
	targetDID string,
) (*Handshake1RTTInitiation, *l0.SessionKeys, error) {
	if initiatorKeys == nil || targetX25519Pub == nil {
		return nil, nil, errors.New("parámetros de identidad incompletos para 1-RTT")
	}

	// Encapsulación híbrida compacta (X25519 + ML-KEM-768)
	sharedKey, cipher, err := EncapsulateCompact(targetX25519Pub, targetMLKEMPubHex)
	if err != nil {
		return nil, nil, fmt.Errorf("error en encapsulación KEM híbrida: %w", err)
	}

	now := time.Now().UnixNano()
	sessionKeys := deriveHandshakeKeys(sharedKey, cipher.Salt, initiatorKeys.DID, targetDID, true)

	initMsg := &Handshake1RTTInitiation{
		SenderDID:    initiatorKeys.DID,
		RecipientDID: targetDID,
		Ciphertext:   cipher,
		Timestamp:    now,
	}

	return initMsg, sessionKeys, nil
}

// Respond1RTT procesa el mensaje de inicio en el receptor, decapsula el secreto post-cuántico
// y genera la respuesta de confirmación en 1 RTT exacto.
func Respond1RTT(
	responderKeys *HybridKeyPair,
	initMsg *Handshake1RTTInitiation,
) (*Handshake1RTTResponse, *l0.SessionKeys, error) {
	if responderKeys == nil || initMsg == nil || initMsg.Ciphertext == nil {
		return nil, nil, errors.New("mensaje de inicio 1-RTT inválido o nulo")
	}

	// Ventana de tolerancia temporal anti-replay preliminar (60 segundos)
	now := time.Now().UnixNano()
	age := time.Duration(now - initMsg.Timestamp)
	if age < -5*time.Second || age > 60*time.Second {
		return nil, nil, fmt.Errorf("marca de tiempo del handshake expirada o desfasada (%v)", age)
	}

	// Desencapsulación KEM híbrida (X25519 + ML-KEM-768)
	sharedKey, err := responderKeys.Decapsulate(initMsg.Ciphertext)
	if err != nil {
		return nil, nil, fmt.Errorf("error en desencapsulación KEM: %w", err)
	}

	sessionKeys := deriveHandshakeKeys(sharedKey, initMsg.Ciphertext.Salt, initMsg.SenderDID, responderKeys.DID, false)

	// Generar etiqueta de autenticación mutua
	authTag := computeAuthTag(sessionKeys.TxKey[:], sessionKeys.SessionID[:], now)

	respMsg := &Handshake1RTTResponse{
		SessionID: sessionKeys.SessionID,
		AuthTag:   authTag,
		Timestamp: now,
	}

	return respMsg, sessionKeys, nil
}

// Finalize1RTT valida la respuesta del receptor en el iniciador, completando el handshake 1 RTT.
func Finalize1RTT(sessionKeys *l0.SessionKeys, respMsg *Handshake1RTTResponse) error {
	if sessionKeys == nil || respMsg == nil {
		return errors.New("parámetros de finalización 1-RTT nulos")
	}

	if sessionKeys.SessionID != respMsg.SessionID {
		return errors.New("SessionID de respuesta no coincide con la sesión iniciada")
	}

	expectedTag := computeAuthTag(sessionKeys.RxKey[:], sessionKeys.SessionID[:], respMsg.Timestamp)
	if !hmac.Equal(respMsg.AuthTag[:], expectedTag[:]) {
		return errors.New("etiqueta de autenticación 1-RTT inválida (falla de verificación)")
	}

	return nil
}

func deriveHandshakeKeys(sharedKey, salt []byte, senderDID, recipientDID string, isInitiator bool) *l0.SessionKeys {
	hTx := sha256.New()
	hTx.Write(sharedKey)
	hTx.Write(salt)
	hTx.Write([]byte(senderDID))
	hTx.Write([]byte("->"))
	hTx.Write([]byte(recipientDID))
	hTx.Write([]byte(":TX_KEY"))
	txSum := hTx.Sum(nil)

	hRx := sha256.New()
	hRx.Write(sharedKey)
	hRx.Write(salt)
	hRx.Write([]byte(recipientDID))
	hRx.Write([]byte("->"))
	hRx.Write([]byte(senderDID))
	hRx.Write([]byte(":RX_KEY"))
	rxSum := hRx.Sum(nil)

	hSID := sha256.New()
	hSID.Write(sharedKey)
	hSID.Write(salt)
	hSID.Write([]byte(":SESSION_ID"))
	sidSum := hSID.Sum(nil)

	var txKey, rxKey [32]byte
	var sessionID [16]byte

	copy(sessionID[:], sidSum[:16])

	if isInitiator {
		copy(txKey[:], txSum)
		copy(rxKey[:], rxSum)
	} else {
		copy(txKey[:], rxSum)
		copy(rxKey[:], txSum)
	}

	return &l0.SessionKeys{
		TxKey:     txKey,
		RxKey:     rxKey,
		SessionID: sessionID,
		PFSActive: true,
	}
}

func computeAuthTag(key, sessionID []byte, timestamp int64) [16]byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(sessionID)
	var timeBytes [8]byte
	binary.BigEndian.PutUint64(timeBytes[:], uint64(timestamp))
	mac.Write(timeBytes[:])
	sum := mac.Sum(nil)

	var tag [16]byte
	copy(tag[:], sum[:16])
	return tag
}
