package l1_test

import (
	"encoding/binary"
	"testing"
	"time"

	"ipvn7/pkg/l0"
	"ipvn7/pkg/l1"
)

// BenchmarkDatapathEndToEnd caracteriza de forma realista y honesta el camino de datos completo:
// UDP Buffer -> DecodePacket -> Frame Validation -> Anti-Replay -> ZTNA -> Session Lookup -> AEAD Decrypt -> Routing Lookup.
// Permite documentar con exactitud factual cuántas alocaciones y bytes reales genera el procesamiento completo por paquete.
func BenchmarkDatapathEndToEnd(b *testing.B) {
	idAlice, _ := l0.GenerateIdentity()
	idBob, _ := l0.GenerateIdentity()

	keysAlice, _ := l1.GenerateHybridKeyPair(idAlice.DID())
	keysBob, _ := l1.GenerateHybridKeyPair(idBob.DID())

	fwBob := l1.NewZTNAFirewall(true)
	fwBob.AuthorizeDID(&l1.DIDPolicy{DID: idAlice.DID(), AllowInbound: true, AllowOutbound: true})
	mgrAlice := l1.NewPQCSessionManager(idAlice, keysAlice, nil)
	mgrBob := l1.NewPQCSessionManager(idBob, keysBob, fwBob)

	routerBob := l1.NewKleinbergRouter(idBob)
	antiReplayBob := l1.NewAntiReplayFilter(nil)

	// Establecer sesión PQC previa
	initPkt, err := mgrAlice.CreateHandshakeInitPacket(idBob.DID(), keysBob.ClassicalKEMPub, keysBob.MLKEMPubHex)
	if err != nil {
		b.Fatalf("Error en init: %v", err)
	}
	respPkt, err := mgrBob.HandleHandshakeInitPacket(initPkt)
	if err != nil {
		b.Fatalf("Error en handle init: %v", err)
	}
	if err := mgrAlice.HandleHandshakeRespPacket(respPkt); err != nil {
		b.Fatalf("Error en handle resp: %v", err)
	}

	// Crear paquete de datos cifrado con AEAD ChaCha20-Poly1305
	plaintext := []byte("DATAPATH_REAL_BENCHMARK_PAYLOAD_1280B_MTU")
	dataPkt, err := mgrAlice.EncryptDataPacket(idBob.DID(), plaintext)
	if err != nil {
		b.Fatalf("Error cifrando paquete: %v", err)
	}

	rawUDP, err := dataPkt.Encode()
	if err != nil {
		b.Fatalf("Error codificando trama UDP: %v", err)
	}

	sKeysBob, _ := mgrBob.GetSession(idAlice.DID())
	sessionID := binary.BigEndian.Uint64(sKeysBob.SessionID[:8])

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// 1. Frame decode
		pkt, err := l0.DecodePacket(rawUDP)
		if err != nil {
			b.Fatal(err)
		}

		// 2. Validation
		if pkt.Magic != l0.MagicBytes || pkt.Version != l0.WireVersion {
			b.Fatal("invalid frame")
		}

		// 3. Anti-Replay
		now := time.Now().Unix()
		_ = antiReplayBob.Accept(pkt.SourceDID, sessionID, uint64(i+1), now)

		// 4. ZTNA Default-Deny Evaluation
		dec, _ := fwBob.EvaluateInbound(pkt.SourceDID, 7777)
		if dec != l1.DecisionAccept {
			b.Fatal("ZTNA rejected")
		}

		// 5. Session Lookup & AEAD Decryption
		if mgrBob.HasSession(pkt.SourceDID) {
			decrypted, err := mgrBob.DecryptDataPacket(pkt)
			if err != nil || len(decrypted) == 0 {
				b.Fatal("decryption failed")
			}
		}

		// 6. Routing Decision
		_, _ = routerBob.FindNextHop(pkt.DestDID)
	}
}
