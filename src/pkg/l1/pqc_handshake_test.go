package l1_test

import (
	"bytes"
	"testing"

	"ipvn7/pkg/l0"
	"ipvn7/pkg/l1"
)

func TestHybridHandshake1RTT_CompleteFlow(t *testing.T) {
	// 1. Generar identidades híbridas de Alice y Bob
	aliceKeys, err := l1.GenerateHybridKeyPair("did:ipvn7:alice-test-node")
	if err != nil {
		t.Fatalf("Error generando claves de Alice: %v", err)
	}

	bobKeys, err := l1.GenerateHybridKeyPair("did:ipvn7:bob-test-node")
	if err != nil {
		t.Fatalf("Error generando claves de Bob: %v", err)
	}

	// 2. Alice inicia el apretón de manos (0.5 RTT)
	initMsg, aliceSession, err := l1.Initiate1RTT(
		aliceKeys,
		bobKeys.ClassicalKEMPub,
		bobKeys.MLKEMPubHex,
		bobKeys.DID,
	)
	if err != nil {
		t.Fatalf("Initiate1RTT falló: %v", err)
	}

	if initMsg.SenderDID != aliceKeys.DID {
		t.Errorf("SenderDID no coincide: %s", initMsg.SenderDID)
	}

	// 3. Bob procesa el mensaje y genera respuesta (1.0 RTT)
	respMsg, bobSession, err := l1.Respond1RTT(bobKeys, initMsg)
	if err != nil {
		t.Fatalf("Respond1RTT falló: %v", err)
	}

	// 4. Alice finaliza la sesión validando la respuesta
	if err := l1.Finalize1RTT(aliceSession, respMsg); err != nil {
		t.Fatalf("Finalize1RTT falló: %v", err)
	}

	// 5. Verificar coincidencia simétrica de claves y SessionID
	if aliceSession.SessionID != bobSession.SessionID {
		t.Errorf("SessionID no coincide entre Alice y Bob")
	}

	if aliceSession.TxKey != bobSession.RxKey {
		t.Errorf("Alice TxKey no coincide con Bob RxKey")
	}

	if aliceSession.RxKey != bobSession.TxKey {
		t.Errorf("Alice RxKey no coincide con Bob TxKey")
	}

	// 6. Probar cifrado y descifrado seguro E2EE con las claves acordadas
	nonce, _ := l0.GenerateNonce()
	plaintext := []byte("Datagrama PQC Post-Cuántica en 1-RTT")
	ad := []byte("metadata-segura")

	ciphertext, err := l0.EncryptPayload(aliceSession.TxKey[:], nonce, plaintext, ad)
	if err != nil {
		t.Fatalf("EncryptPayload falló: %v", err)
	}

	decrypted, err := l0.DecryptPayload(bobSession.RxKey[:], nonce, ciphertext, ad)
	if err != nil {
		t.Fatalf("DecryptPayload falló: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Texto descifrado no coincide con el texto original")
	}
}

func BenchmarkHybridHandshake1RTT(b *testing.B) {
	aliceKeys, _ := l1.GenerateHybridKeyPair("did:ipvn7:alice-benchmark")
	bobKeys, _ := l1.GenerateHybridKeyPair("did:ipvn7:bob-benchmark")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		initMsg, aliceSession, err := l1.Initiate1RTT(
			aliceKeys,
			bobKeys.ClassicalKEMPub,
			bobKeys.MLKEMPubHex,
			bobKeys.DID,
		)
		if err != nil {
			b.Fatalf("Initiate error: %v", err)
		}

		respMsg, bobSession, err := l1.Respond1RTT(bobKeys, initMsg)
		if err != nil {
			b.Fatalf("Respond error: %v", err)
		}

		if err := l1.Finalize1RTT(aliceSession, respMsg); err != nil {
			b.Fatalf("Finalize error: %v", err)
		}

		if aliceSession.TxKey != bobSession.RxKey {
			b.Fatalf("Key mismatch")
		}
	}
}
