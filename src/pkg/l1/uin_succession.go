// Package l1 implementa el Testamento Criptográfico y la Migración de Identidad
// bajo compromiso hostil de acuerdo con ip7uin_MVP_Spec_v1.1.docx (Sección 4.3).
// Permite que el dueño legítimo de una entidad recupere y migre su identidad y Trust Score
// mediante una clave de sucesión offline (S) que nunca estuvo en línea.
package l1

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidSuccessionProof = errors.New("uin: prueba de sucesion invalida")
	ErrSuccessionHashMismatch = errors.New("uin: hash de clave de sucesion no coincide con el registro previo")
	ErrSuccessionExpired      = errors.New("uin: timestamp de sucesion fuera de ventana admisible")
)

// SuccessionProof representa la prueba de migración soberana firmada por la clave offline S
type SuccessionProof struct {
	OldRootID        [32]byte  `json:"old_root_id"`
	NewRootID        [32]byte  `json:"new_root_id"`
	Timestamp        time.Time `json:"timestamp"`
	SuccessionPubKey []byte    `json:"succession_pub_key"`
	Signature        []byte    `json:"signature"`
}

// CanonicalBytes serializa de forma determinista la prueba para firma y verificación
func (sp *SuccessionProof) CanonicalBytes() []byte {
	buf := make([]byte, 32+32+8+len("SUCCESSION")+len(sp.SuccessionPubKey))
	offset := 0

	copy(buf[offset:offset+32], sp.OldRootID[:])
	offset += 32

	copy(buf[offset:offset+32], sp.NewRootID[:])
	offset += 32

	binary.BigEndian.PutUint64(buf[offset:offset+8], uint64(sp.Timestamp.Unix()))
	offset += 8

	copy(buf[offset:offset+len("SUCCESSION")], []byte("SUCCESSION"))
	offset += len("SUCCESSION")

	copy(buf[offset:], sp.SuccessionPubKey)
	return buf
}

// GenerateSuccessionProof crea y firma una prueba de migración utilizando la clave offline S
func GenerateSuccessionProof(oldRootID, newRootID [32]byte, successionPrivKey ed25519.PrivateKey) (*SuccessionProof, error) {
	if len(successionPrivKey) != ed25519.PrivateKeySize {
		return nil, errors.New("uin: clave privada de sucesion invalida")
	}

	pubKey := successionPrivKey.Public().(ed25519.PublicKey)
	proof := &SuccessionProof{
		OldRootID:        oldRootID,
		NewRootID:        newRootID,
		Timestamp:        time.Now().UTC(),
		SuccessionPubKey: pubKey,
	}

	msg := proof.CanonicalBytes()
	proof.Signature = ed25519.Sign(successionPrivKey, msg)

	return proof, nil
}

// VerifySuccessionMigration valida que la prueba de sucesión coincida con el hash registrado y la firma sea auténtica
func VerifySuccessionMigration(expectedSuccessionKeyHash [32]byte, proof *SuccessionProof) (bool, error) {
	if proof == nil || len(proof.SuccessionPubKey) != ed25519.PublicKeySize || len(proof.Signature) != ed25519.SignatureSize {
		return false, ErrInvalidSuccessionProof
	}

	// 1. Validar que sha256(SuccessionPubKey) coincide con el hash comprometido en el BindingRecord inicial
	computedHash := sha256.Sum256(proof.SuccessionPubKey)
	if computedHash != expectedSuccessionKeyHash {
		return false, fmt.Errorf("%w: esperado %s, obtenido %s", ErrSuccessionHashMismatch,
			hex.EncodeToString(expectedSuccessionKeyHash[:8]), hex.EncodeToString(computedHash[:8]))
	}

	// 2. Ventana de frescura temporal (máximo 48 horas hacia atrás o 5 minutos a futuro)
	now := time.Now().UTC()
	if proof.Timestamp.After(now.Add(5*time.Minute)) || proof.Timestamp.Before(now.Add(-48*time.Hour)) {
		return false, ErrSuccessionExpired
	}

	// 3. Verificación criptográfica estricta de la firma con la clave pública de sucesión
	msg := proof.CanonicalBytes()
	if !ed25519.Verify(proof.SuccessionPubKey, msg, proof.Signature) {
		return false, ErrInvalidSuccessionProof
	}

	return true, nil
}
