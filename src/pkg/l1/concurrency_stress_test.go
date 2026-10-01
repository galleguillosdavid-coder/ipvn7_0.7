package l1

import (
	"crypto/rand"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"ipvn7/pkg/l0"
)

// TestConcurrency_SessionManager valida la adición, consulta y concurrencia de sesiones PQC.
func TestConcurrency_SessionManager(t *testing.T) {
	aliceID, err := l0.GenerateIdentity()
	if err != nil {
		t.Fatalf("Failed to create identity: %v", err)
	}
	aliceKP, err := GenerateHybridKeyPair(aliceID.DID())
	if err != nil {
		t.Fatalf("Failed to generate keypair: %v", err)
	}
	fw := NewZTNAFirewall(true)
	sm := NewPQCSessionManager(aliceID, aliceKP, fw)

	var wg sync.WaitGroup
	workers := 16
	iterations := 25

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				peerDID := fmt.Sprintf("did:ipvn7:worker-%d-iter-%d", workerID, i)
				sharedKey := make([]byte, 32)
				_, _ = rand.Read(sharedKey)

				var keyArr [32]byte
				copy(keyArr[:], sharedKey)
				sessionKeys := &l0.SessionKeys{
					RxKey: keyArr,
					TxKey: keyArr,
				}

				sm.SetSession(peerDID, sessionKeys)

				// Consultar concurrentemente
				lookedUp, exists := sm.GetSession(peerDID)
				if !exists || lookedUp == nil {
					t.Errorf("worker %d: session %s not found immediately after set", workerID, peerDID)
				}
				if !sm.HasSession(peerDID) {
					t.Errorf("worker %d: HasSession returned false for %s", workerID, peerDID)
				}
			}
		}(w)
	}
	wg.Wait()
}

// TestConcurrency_AntiReplay valida que el filtro por sesión maneje concurrentemente
// múltiples sesiones, peers distintos con el mismo sequence y secuencias concurrentes.
func TestConcurrency_AntiReplay(t *testing.T) {
	ar := NewAntiReplayFilter(nil)
	var wg sync.WaitGroup
	peers := 8
	sequencesPerPeer := 50

	now := time.Now().Unix()

	for p := 0; p < peers; p++ {
		wg.Add(1)
		go func(peerIdx int) {
			defer wg.Done()
			peerDID := fmt.Sprintf("did:ipvn7:replay-peer-%d", peerIdx)
			sessionID := uint64(peerIdx + 1)

			for seq := uint64(1); seq <= uint64(sequencesPerPeer); seq++ {
				// Primer envío debe ser aceptado
				if !ar.Accept(peerDID, sessionID, seq, now) {
					t.Errorf("peer %s seq %d first attempt was rejected", peerDID, seq)
				}

				// Replay inmediato del mismo paquete debe ser rechazado
				if ar.Accept(peerDID, sessionID, seq, now) {
					t.Errorf("peer %s seq %d replay was erroneously accepted", peerDID, seq)
				}
			}
		}(p)
	}
	wg.Wait()

	// Probar colisión intencional de número de secuencia entre dos peers distintos
	// Peer A con seq 9999 y Peer B con seq 9999 deben ser ambos aceptados independientemente
	pA := "did:ipvn7:peer-alpha"
	pB := "did:ipvn7:peer-beta"
	if !ar.Accept(pA, 100, 9999, now) {
		t.Errorf("pA seq 9999 rejected")
	}
	if !ar.Accept(pB, 100, 9999, now) {
		t.Errorf("pB seq 9999 rejected because of pA")
	}
}

// TestConcurrency_Firewall evalúa la política Zero-Trust bajo concurrencia masiva.
func TestConcurrency_Firewall(t *testing.T) {
	fw := NewZTNAFirewall(true)
	var wg sync.WaitGroup
	workers := 10

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				did := fmt.Sprintf("did:ipvn7:worker-%d-did-%d", workerID, i)
				fw.AllowPeer(did, nil)
				if !fw.IsAllowed(did) {
					t.Errorf("DID %s should be allowed", did)
				}
				fw.RevokePeer(did)
				if fw.IsAllowed(did) {
					t.Errorf("DID %s should be revoked", did)
				}
			}
		}(w)
	}
	wg.Wait()
}

// TestConcurrency_Routing evalúa actualizaciones y consultas concurrentes en el árbol topológico Kleinberg.
func TestConcurrency_Routing(t *testing.T) {
	localID, err := l0.GenerateIdentity()
	if err != nil {
		t.Fatalf("Failed to create identity: %v", err)
	}
	router := NewKleinbergRouter(localID)

	var wg sync.WaitGroup
	workers := 8

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < 30; i++ {
				peerID, err := l0.GenerateIdentity()
				if err != nil {
					return
				}
				peerDID := peerID.DID()
				udpAddr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", 8000+workerID*100+i))

				_ = router.AddOrUpdatePeer(peerDID, udpAddr, 10.5)

				// Consulta concurrente
				_, _ = router.FindNextHop(peerDID)
				_ = router.GetAllPeers()
			}
		}(w)
	}
	wg.Wait()
}

// TestConcurrency_SimultaneousHandshakes prueba dos handshakes simultáneos con el mismo peer
func TestConcurrency_SimultaneousHandshakes(t *testing.T) {
	aliceID, err := l0.GenerateIdentity()
	if err != nil {
		t.Fatalf("alice id: %v", err)
	}
	aliceKP, err := GenerateHybridKeyPair(aliceID.DID())
	if err != nil {
		t.Fatalf("alice kp: %v", err)
	}
	bobID, err := l0.GenerateIdentity()
	if err != nil {
		t.Fatalf("bob id: %v", err)
	}
	bobKP, err := GenerateHybridKeyPair(bobID.DID())
	if err != nil {
		t.Fatalf("bob kp: %v", err)
	}

	fwAlice := NewZTNAFirewall(false)
	fwBob := NewZTNAFirewall(false)
	aliceSM := NewPQCSessionManager(aliceID, aliceKP, fwAlice)
	bobSM := NewPQCSessionManager(bobID, bobKP, fwBob)

	var wg sync.WaitGroup
	attempts := 4

	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			initPkt, err := aliceSM.CreateHandshakeInitPacket(bobID.DID(), bobKP.ClassicalKEMPub, bobKP.MLKEMPubHex)
			if err != nil {
				t.Errorf("CreateHandshakeInitPacket failed: %v", err)
				return
			}
			respPkt, err := bobSM.HandleHandshakeInitPacket(initPkt)
			if err != nil {
				t.Errorf("HandleHandshakeInitPacket failed: %v", err)
				return
			}
			if respPkt == nil {
				t.Errorf("respPkt is nil")
				return
			}
		}()
	}
	wg.Wait()
}
