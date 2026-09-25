package l0

import (
	"bytes"
	"testing"
)

func TestMLKEM768Adapter_EncapsulateDecapsulate(t *testing.T) {
	// 1. Bob genera su par de claves ML-KEM-768
	bobKeyPair, err := GenerateMLKEM768KeyPair()
	if err != nil {
		t.Fatalf("error generando par de claves de Bob: %v", err)
	}

	bobAdapter := NewMLKEM768Adapter(bobKeyPair)
	aliceAdapter := NewMLKEM768Adapter(nil) // Alice solo necesita la clave pública de Bob

	// 2. Alice encapsula un secreto compartido para la clave pública de Bob
	ciphertext, aliceSharedSecret, err := aliceAdapter.Encapsulate(bobKeyPair.PublicKey[:])
	if err != nil {
		t.Fatalf("Alice falló al encapsular: %v", err)
	}

	if len(ciphertext) != MLKEM768CiphertextBytes {
		t.Errorf("longitud de ciphertext inesperada: %d != %d", len(ciphertext), MLKEM768CiphertextBytes)
	}
	if len(aliceSharedSecret) != MLKEM768SharedSecretLen {
		t.Errorf("longitud de secreto compartido inesperada: %d != %d", len(aliceSharedSecret), MLKEM768SharedSecretLen)
	}

	// 3. Bob desencapsula el texto cifrado con su par de claves
	bobSharedSecret, err := bobAdapter.Decapsulate(ciphertext)
	if err != nil {
		t.Fatalf("Bob falló al decapsular: %v", err)
	}

	// 4. Los secretos compartidos DEBEN ser estrictamente idénticos
	if !bytes.Equal(aliceSharedSecret, bobSharedSecret) {
		t.Fatalf("discordancia crítica: secreto de Alice y Bob no coinciden")
	}
}

func TestMLKEM768Adapter_RejectionOnTamperedCiphertext(t *testing.T) {
	bobKeyPair, _ := GenerateMLKEM768KeyPair()
	bobAdapter := NewMLKEM768Adapter(bobKeyPair)
	aliceAdapter := NewMLKEM768Adapter(nil)

	ciphertext, _, _ := aliceAdapter.Encapsulate(bobKeyPair.PublicKey[:])

	// Corromper 1 byte de la etiqueta de autenticación
	corruptCiphertext := make([]byte, len(ciphertext))
	copy(corruptCiphertext, ciphertext)
	corruptCiphertext[40] ^= 0xFF

	_, err := bobAdapter.Decapsulate(corruptCiphertext)
	if err == nil {
		t.Fatal("vulnerabilidad: Bob aceptó un ciphertext corrupto sin error")
	}
	if err != ErrCorruptPQCCiphertext {
		t.Errorf("error inesperado al rechazar: %v", err)
	}
}

func TestDeriveHybridSecret_DeterministicCombination(t *testing.T) {
	classicSecret := []byte("x25519-shared-secret-32-bytes-ok")
	pqcSecret := []byte("mlkem-shared-secret-32-bytes-ok!")

	key1, err := DeriveHybridSecret(classicSecret, pqcSecret)
	if err != nil {
		t.Fatalf("error derivando clave híbrida: %v", err)
	}

	key2, err := DeriveHybridSecret(classicSecret, pqcSecret)
	if err != nil {
		t.Fatalf("error en segunda derivación: %v", err)
	}

	if !bytes.Equal(key1[:], key2[:]) {
		t.Fatal("la derivación híbrida no es determinista")
	}

	// Probar fallo ante secretos vacíos
	if _, err := DeriveHybridSecret(nil, pqcSecret); err == nil {
		t.Fatal("se esperaba error con secreto clásico nulo")
	}
}
