package l1_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"

	"ipvn7/pkg/l1"
)

func TestUINGossipManager_DisseminationAndDeduplication(t *testing.T) {
	// 1. Inicializar autoridad raíz UIN
	uinMgr, err := l1.NewUINIdentityManager(l1.IdentityModeRoot)
	if err != nil {
		t.Fatalf("Fallo creando UINIdentityManager: %v", err)
	}

	gossipMgr := l1.NewUINGossipManager(uinMgr, 3)

	// 2. Emitir BindingRecord para un agente IA
	agentPub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Fallo generando clave de agente: %v", err)
	}

	record, err := uinMgr.IssueBinding(agentPub, l1.BindingAlgoEd25519, l1.BindingScopeAIAgent, 24*time.Hour, [16]byte{})
	if err != nil {
		t.Fatalf("IssueBinding falló: %v", err)
	}

	// 3. Crear envelope para difusión
	_, wire, err := gossipMgr.CreateEnvelope(record, uinMgr.EntityDID())
	if err != nil {
		t.Fatalf("CreateEnvelope falló: %v", err)
	}

	// 4. Simular nodo receptor que procesa el sobre por primera vez
	rootPub := uinMgr.RootPublicKey()
	receiverGossip := l1.NewUINGossipManager(nil, 3)

	recReceived, isNew, err := receiverGossip.ProcessIncomingEnvelope(wire, rootPub)
	if err != nil {
		t.Fatalf("ProcessIncomingEnvelope falló en primera entrega: %v", err)
	}
	if !isNew {
		t.Errorf("El sobre debía marcarse como nuevo en primera recepción")
	}
	if recReceived.KeyID != record.KeyID {
		t.Errorf("KeyID no coincide en registro deserializado")
	}

	// 5. Segunda entrega idéntica: deduplicación O(1)
	_, isNewDuplicate, err := receiverGossip.ProcessIncomingEnvelope(wire, rootPub)
	if err != nil {
		t.Fatalf("Error en entrega duplicada: %v", err)
	}
	if isNewDuplicate {
		t.Errorf("El sobre duplicado debía suprimirse (isNew == false)")
	}
}

func TestUINGossipManager_FullCycle(t *testing.T) {
	// Firmar con IssueBinding de un UINIdentityManager real
	uinMgr, err := l1.NewUINIdentityManager(l1.IdentityModeHybrid)
	if err != nil {
		t.Fatalf("Error creando UINIdentityManager: %v", err)
	}

	pub, _, _ := ed25519.GenerateKey(rand.Reader)
	realRec, err := uinMgr.IssueBinding(pub, l1.BindingAlgoEd25519, l1.BindingScopeGeneral, 10*time.Minute, [16]byte{})
	if err != nil {
		t.Fatalf("Error emitiendo binding: %v", err)
	}

	mgr := l1.NewUINGossipManager(uinMgr, 3)
	env, wire, err := mgr.CreateEnvelope(realRec, "did:ipvn7:origin")
	if err != nil {
		t.Fatalf("CreateEnvelope falló: %v", err)
	}

	if env.Record.KeyID != realRec.KeyID {
		t.Errorf("KeyID en sobre no coincide")
	}

	// Validar recepción en un segundo gestor gossip usando el wire generado
	receiverMgr := l1.NewUINGossipManager(nil, 3)
	rec, isNew, err := receiverMgr.ProcessIncomingEnvelope(wire, uinMgr.RootPublicKey())
	if err != nil {
		t.Fatalf("ProcessIncomingEnvelope falló: %v", err)
	}
	if !isNew || rec.KeyID != realRec.KeyID {
		t.Errorf("Recuperación de BindingRecord en receptor incorrecta")
	}

	// Verificar selección de targets de fanout
	peer1 := &l1.PeerNode{DID: "did:ipvn7:peer1"}
	peer2 := &l1.PeerNode{DID: "did:ipvn7:peer2"}
	peer3 := &l1.PeerNode{DID: "did:ipvn7:peer3"}
	peer4 := &l1.PeerNode{DID: "did:ipvn7:peer4"}
	peers := []*l1.PeerNode{peer1, peer2, peer3, peer4}

	targets := receiverMgr.SelectGossipTargets(peers, "did:ipvn7:peer1")
	if len(targets) != 3 {
		t.Errorf("Se esperaban 3 targets por fanout, obtenidos: %d", len(targets))
	}
	for _, trg := range targets {
		if trg.DID == "did:ipvn7:peer1" {
			t.Errorf("Peer excluido no debió ser seleccionado como target")
		}
	}
}

func BenchmarkUINGossipManager_CreateEnvelope(b *testing.B) {
	uinMgr, _ := l1.NewUINIdentityManager(l1.IdentityModeHybrid)
	rec, _ := uinMgr.IssueBinding(make([]byte, 32), l1.BindingAlgoEd25519, l1.BindingScopeAIAgent, 1*time.Hour, [16]byte{})
	gossipMgr := l1.NewUINGossipManager(uinMgr, 3)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, _ = gossipMgr.CreateEnvelope(rec, "did:ipvn7:origin")
	}
}
