package l2

import (
	"testing"
	"time"
)

func TestCRDTTopologyGraph_AddAndMergeConvergence(t *testing.T) {
	nodeA := NewCRDTTopologyGraph("did:ipvn7:nodeA")
	nodeB := NewCRDTTopologyGraph("did:ipvn7:nodeB")

	// Nodo A genera mutaciones
	delta1 := nodeA.AddNode("did:ipvn7:peer1", "Relay Alpha")
	delta2, _ := nodeA.AddEdge("edge_1_2", "did:ipvn7:peer1", "did:ipvn7:peer2", "P2P_MESH", 12.5)

	// Nodo B genera mutaciones concurrentes
	delta3 := nodeB.AddNode("did:ipvn7:peer2", "Edge Node Beta")

	// Intercambio cruzado de deltas (simulando tránsito de red asíncrono y desordenado)
	if err := nodeB.MergeDelta(delta2); err != nil {
		t.Fatalf("nodeB MergeDelta(delta2) falló: %v", err)
	}
	if err := nodeB.MergeDelta(delta1); err != nil {
		t.Fatalf("nodeB MergeDelta(delta1) falló: %v", err)
	}

	if err := nodeA.MergeDelta(delta3); err != nil {
		t.Fatalf("nodeA MergeDelta(delta3) falló: %v", err)
	}

	// Comprobar que ambos nodos convergen exactamente al mismo conteo de elementos
	nodesA, edgesA := nodeA.ActiveElements()
	nodesB, edgesB := nodeB.ActiveElements()

	if nodesA != 2 || nodesB != 2 {
		t.Errorf("discordancia en nodos activos: A=%d, B=%d (esperado 2)", nodesA, nodesB)
	}
	if edgesA != 1 || edgesB != 1 {
		t.Errorf("discordancia en aristas activas: A=%d, B=%d (esperado 1)", edgesA, edgesB)
	}
}

func TestCRDTTopologyGraph_TombstoneLWW(t *testing.T) {
	graph := NewCRDTTopologyGraph("did:ipvn7:local")

	graph.AddNode("did:ipvn7:temp_peer", "Ephemeral Node")
	nodes, _ := graph.ActiveElements()
	if nodes != 1 {
		t.Fatalf("se esperaba 1 nodo activo, encontrados %d", nodes)
	}

	time.Sleep(1 * time.Millisecond) // Asegurar incremento de timestamp
	graph.RemoveNode("did:ipvn7:temp_peer")

	nodes, _ = graph.ActiveElements()
	if nodes != 0 {
		t.Errorf("se esperaba 0 nodos activos tras tombstone, encontrados %d", nodes)
	}
}
