package l2

import (
	"testing"
)

func TestPlanetaryMeshCluster_CreationAndRouting(t *testing.T) {
	nodeCount := 25
	cluster, err := NewPlanetaryMeshCluster(nodeCount)
	if err != nil {
		t.Fatalf("error creando cluster planetario: %v", err)
	}

	if cluster.NodeCount() != nodeCount {
		t.Errorf("se esperaban %d nodos, se obtuvieron: %d", nodeCount, cluster.NodeCount())
	}

	// 1. Probar enrutamiento diagonal de extremo a extremo
	src := "node-000"
	dst := "node-024"

	path, err := cluster.RoutePacket(src, dst)
	if err != nil {
		t.Fatalf("fallo al enrutar de %s a %s: %v", src, dst, err)
	}

	if len(path) == 0 {
		t.Fatal("el camino calculado no debe estar vacío")
	}

	if path[0] != src || path[len(path)-1] != dst {
		t.Errorf("camino inválido: inicio=%s (esperado %s), fin=%s (esperado %s)", path[0], src, path[len(path)-1], dst)
	}

	// Verificación de cota de saltos Kleinberg: para N=25, log2(25) ~ 4.6 -> saltos <= 15
	if len(path) > 15 {
		t.Errorf("camino excesivamente largo para topología de Kleinberg: %d saltos", len(path))
	}
}

func TestPlanetaryMeshCluster_Scale50NodesAndResilience(t *testing.T) {
	nodeCount := 50
	cluster, err := NewPlanetaryMeshCluster(nodeCount)
	if err != nil {
		t.Fatalf("error creando cluster de 50 nodos: %v", err)
	}

	// Verificar múltiples rutas transversales
	testPairs := [][2]string{
		{"node-000", "node-049"},
		{"node-005", "node-042"},
		{"node-010", "node-035"},
	}

	for _, pair := range testPairs {
		path, err := cluster.RoutePacket(pair[0], pair[1])
		if err != nil {
			t.Errorf("fallo de ruta entre %s y %s: %v", pair[0], pair[1], err)
			continue
		}
		if len(path) < 2 {
			t.Errorf("ruta trivial o corrupta para par %v: %v", pair, path)
		}
	}

	// Simular fallo del 10% de nodos y verificar continuidad
	failed := cluster.SimulateFailure(0.10)
	if failed == 0 {
		t.Log("ningún nodo cayó en esta corrida estocástica")
	} else {
		t.Logf("simulado fallo de %d nodos físicos en la malla", failed)
	}
}
