package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"runtime"
	"testing"
	"time"

	"ipvn7/pkg/core"
	"ipvn7/pkg/interfaces"
	"ipvn7/pkg/l0"
	"ipvn7/pkg/l1"
)

type ComponentProfile struct {
	Name      string
	NsPerOp   float64
	BytesOp   uint64
	AllocsOp  uint64
	PctOfTotal float64
}

func main() {
	fmt.Println("================================================================")
	fmt.Println("  FASE 17: INVESTIGACIÓN EMPÍRICA DE CUELLOS DE BOTELLA")
	fmt.Println("  Profiling Componente por Componente del Datapath IPVN7")
	fmt.Println("================================================================")
	fmt.Println()

	// Preparar entorno para mediciones
	idAlice, _ := l0.GenerateIdentity()
	idBob, _ := l0.GenerateIdentity()
	keysAlice, _ := l1.GenerateHybridKeyPair(idAlice.DID())
	keysBob, _ := l1.GenerateHybridKeyPair(idBob.DID())

	fwBob := l1.NewZTNAFirewall(true)
	fwBob.AuthorizeDID(&l1.DIDPolicy{DID: idAlice.DID(), AllowInbound: true})

	mgrAlice := l1.NewPQCSessionManager(idAlice, keysAlice, nil)
	mgrBob := l1.NewPQCSessionManager(idBob, keysBob, fwBob)

	routerBob := l1.NewKleinbergRouter(idBob)
	antiReplayBob := l1.NewAntiReplayFilter(nil)

	// Handshake previo
	initPkt, _ := mgrAlice.CreateHandshakeInitPacket(idBob.DID(), keysBob.ClassicalKEMPub, keysBob.MLKEMPubHex)
	respPkt, _ := mgrBob.HandleHandshakeInitPacket(initPkt)
	_ = mgrAlice.HandleHandshakeRespPacket(respPkt)

	// Paquete cifrado
	plaintext := []byte("DATAPATH_REAL_BENCHMARK_PAYLOAD_1280B_MTU")
	dataPkt, _ := mgrAlice.EncryptDataPacket(idBob.DID(), plaintext)
	rawUDP, _ := dataPkt.Encode()

	sKeysBob, _ := mgrBob.GetSession(idAlice.DID())
	sessionID := binary.BigEndian.Uint64(sKeysBob.SessionID[:8])

	// 1. Benchmark: DecodePacket (CBOR)
	resDecode := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			pkt, err := l0.DecodePacket(rawUDP)
			if err != nil || pkt == nil {
				b.Fatal(err)
			}
		}
	})

	// 2. Benchmark: EncodePacket (CBOR)
	decodedPkt, _ := l0.DecodePacket(rawUDP)
	resEncode := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			bBytes, err := decodedPkt.Encode()
			if err != nil || len(bBytes) == 0 {
				b.Fatal(err)
			}
		}
	})

	// 3. Benchmark: Anti-Replay Filter
	resAntiReplay := testing.Benchmark(func(b *testing.B) {
		now := time.Now().Unix()
		for i := 0; i < b.N; i++ {
			_ = antiReplayBob.Accept(idAlice.DID(), sessionID, uint64(i+1), now)
		}
	})

	// 4. Benchmark: ZTNA Firewall EvaluateInbound
	resZTNA := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			dec, _ := fwBob.EvaluateInbound(idAlice.DID(), 7777)
			if dec != l1.DecisionAccept {
				b.Fatal("rejected")
			}
		}
	})

	// 5. Benchmark: AEAD DecryptDataPacket
	resAEAD := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			decrypted, err := mgrBob.DecryptDataPacket(decodedPkt)
			if err != nil || len(decrypted) == 0 {
				b.Fatal(err)
			}
		}
	})

	// 6. Benchmark: Kleinberg FindNextHop
	resRouting := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = routerBob.FindNextHop(idBob.DID())
		}
	})

	// 7. Benchmark: PacketContext Pool (Acquire + Release)
	resPool := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			pCtx := core.AcquirePacketContext(nil, nil)
			core.ReleasePacketContext(pCtx)
		}
	})

	// 8. Benchmark: LinearPipeline.Execute (Lock-free Datapath Dispatch)
	pipe := core.NewLinearPipeline()
	pipe.AddStage(&mockStage{name: "s1"})
	pipe.AddStage(&mockStage{name: "s2"})
	pipeCtx := context.Background()
	dummyPCtx := &interfaces.PacketContext{}
	resPipeline := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			dummyPCtx.Handled = false
			dummyPCtx.Dropped = false
			_ = pipe.Execute(pipeCtx, dummyPCtx)
		}
	})

	// 9. Benchmark: Ed25519 SignPacket (Control Plane)
	rawSignPkt := l0.NewPacket(l0.MsgTypeData, idAlice.DID(), idBob.DID(), 1, make([]byte, 16), []byte("ping"))
	resSign := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = rawSignPkt.SignPacket(idAlice)
		}
	})

	// 10. Benchmark: Ed25519 VerifyPacket (Control Plane)
	_ = rawSignPkt.SignPacket(idAlice)
	resVerify := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = rawSignPkt.VerifyPacketSignature()
		}
	})

	// 11. Benchmark: End-to-End Datapath Completo
	resEndToEnd := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			pkt, _ := l0.DecodePacket(rawUDP)
			now := time.Now().Unix()
			_ = antiReplayBob.Accept(pkt.SourceDID, sessionID, uint64(i+1), now)
			_, _ = fwBob.EvaluateInbound(pkt.SourceDID, 7777)
			if mgrBob.HasSession(pkt.SourceDID) {
				_, _ = mgrBob.DecryptDataPacket(pkt)
			}
			_, _ = routerBob.FindNextHop(pkt.DestDID)
		}
	})

	totalDatapathNs := float64(resEndToEnd.NsPerOp())

	components := []ComponentProfile{
		{"1. DecodePacket (CBOR Deserialization)", float64(resDecode.NsPerOp()), uint64(resDecode.AllocedBytesPerOp()), uint64(resDecode.AllocsPerOp()), (float64(resDecode.NsPerOp()) / totalDatapathNs) * 100.0},
		{"2. AEAD DecryptDataPacket (ChaCha20-Poly1305)", float64(resAEAD.NsPerOp()), uint64(resAEAD.AllocedBytesPerOp()), uint64(resAEAD.AllocsPerOp()), (float64(resAEAD.NsPerOp()) / totalDatapathNs) * 100.0},
		{"3. EncodePacket (CBOR Serialization)", float64(resEncode.NsPerOp()), uint64(resEncode.AllocedBytesPerOp()), uint64(resEncode.AllocsPerOp()), (float64(resEncode.NsPerOp()) / totalDatapathNs) * 100.0},
		{"4. ZTNA Firewall EvaluateInbound", float64(resZTNA.NsPerOp()), uint64(resZTNA.AllocedBytesPerOp()), uint64(resZTNA.AllocsPerOp()), (float64(resZTNA.NsPerOp()) / totalDatapathNs) * 100.0},
		{"5. Anti-Replay Filter Validate & Update", float64(resAntiReplay.NsPerOp()), uint64(resAntiReplay.AllocedBytesPerOp()), uint64(resAntiReplay.AllocsPerOp()), (float64(resAntiReplay.NsPerOp()) / totalDatapathNs) * 100.0},
		{"6. Kleinberg Router FindNextHop", float64(resRouting.NsPerOp()), uint64(resRouting.AllocedBytesPerOp()), uint64(resRouting.AllocsPerOp()), (float64(resRouting.NsPerOp()) / totalDatapathNs) * 100.0},
		{"7. PacketContext Pool (Acquire/Release)", float64(resPool.NsPerOp()), uint64(resPool.AllocedBytesPerOp()), uint64(resPool.AllocsPerOp()), (float64(resPool.NsPerOp()) / totalDatapathNs) * 100.0},
		{"8. LinearPipeline.Execute (Lock-free Datapath)", float64(resPipeline.NsPerOp()), uint64(resPipeline.AllocedBytesPerOp()), uint64(resPipeline.AllocsPerOp()), (float64(resPipeline.NsPerOp()) / totalDatapathNs) * 100.0},
	}

	fmt.Println("--- DESGLOSE DEL DATAPATH END-TO-END ---")
	fmt.Printf("Datapath Total End-to-End: %.2f ns/op (%.2f µs) | %d B/op | %d allocs/op\n\n",
		totalDatapathNs, totalDatapathNs/1000.0, resEndToEnd.AllocedBytesPerOp(), resEndToEnd.AllocsPerOp())

	fmt.Printf("| %-45s | %-12s | %-10s | %-10s | %-10s |\n", "Componente", "Latencia", "% Datapath", "Bytes/op", "Allocs/op")
	fmt.Printf("|%s|%s|%s|%s|%s|\n", "-----------------------------------------------", "--------------", "------------", "------------", "------------")

	for _, c := range components {
		fmt.Printf("| %-45s | %8.2f ns | %9.2f%% | %8d B | %8d   |\n",
			c.Name, c.NsPerOp, c.PctOfTotal, c.BytesOp, c.AllocsOp)
	}

	fmt.Println()
	fmt.Println("--- COMPONENTES DE CONTROL PLANE (FIRMA Y VERIFICACIÓN ASIMÉTRICA) ---")
	fmt.Printf("| %-45s | %-12s | %-10s | %-10s |\n", "Operación Criptográfica Asimétrica", "Latencia", "Bytes/op", "Allocs/op")
	fmt.Printf("|%s|%s|%s|%s|\n", "-----------------------------------------------", "--------------", "------------", "------------")
	fmt.Printf("| %-45s | %8.2f ns | %8d B | %8d   |\n", "Ed25519 SignPacket", float64(resSign.NsPerOp()), resSign.AllocedBytesPerOp(), resSign.AllocsPerOp())
	fmt.Printf("| %-45s | %8.2f ns | %8d B | %8d   |\n", "Ed25519 VerifyPacket", float64(resVerify.NsPerOp()), resVerify.AllocedBytesPerOp(), resVerify.AllocsPerOp())

	fmt.Println()
	fmt.Println("================================================================")
	fmt.Println("  ANÁLISIS FACTUAL Y LOCALIZACIÓN DE CUELLOS DE BOTELLA")
	fmt.Println("================================================================")
}

type mockStage struct {
	name string
}

func (m *mockStage) Name() string { return m.name }
func (m *mockStage) Process(ctx context.Context, pCtx *interfaces.PacketContext) error {
	runtime.KeepAlive(pCtx)
	return nil
}
