package l2

import (
	"testing"
	"time"
)

func TestKademlia_XORDistanceAndPrefix(t *testing.T) {
	idA := NewNodeIDFromDID("did:ipvn7:node-a")
	idB := NewNodeIDFromDID("did:ipvn7:node-b")

	dist := XORDistance(idA, idB)
	distSelf := XORDistance(idA, idA)

	// La distancia consigo mismo debe ser idéntica a 0
	var zeroDist [IDByteLen]byte
	if distSelf != zeroDist {
		t.Fatal("self distance must be 0")
	}

	if dist == zeroDist {
		t.Fatal("distance between distinct nodes must be non-zero")
	}

	// Simetría XOR: dist(a, b) == dist(b, a)
	distSym := XORDistance(idB, idA)
	if dist != distSym {
		t.Fatal("XOR distance must be symmetric")
	}

	cplSelf := CommonPrefixLen(idA, idA)
	if cplSelf != IDByteLen*8 {
		t.Fatalf("expected CommonPrefixLen for self to be %d, got: %d", IDByteLen*8, cplSelf)
	}
}

func TestRoutingTable_UpdateAndFindClosest(t *testing.T) {
	selfID := NewNodeIDFromDID("did:ipvn7:self-node")
	rt := NewRoutingTable(selfID)

	// 1. Prohibición de Self-Peering (Regla 2)
	rt.Update(Contact{
		ID:       selfID,
		DID:      "did:ipvn7:self-node",
		Endpoint: "127.0.0.1:7777",
		LastSeen: time.Now(),
	})
	if rt.TotalContacts() != 0 {
		t.Fatal("expected 0 contacts after self-peering attempt")
	}

	// 2. Insertar contactos válidos
	contactB := Contact{
		ID:       NewNodeIDFromDID("did:ipvn7:node-b"),
		DID:      "did:ipvn7:node-b",
		Endpoint: "192.168.1.106:7001",
		LastSeen: time.Now(),
	}
	contactC := Contact{
		ID:       NewNodeIDFromDID("did:ipvn7:node-c"),
		DID:      "did:ipvn7:node-c",
		Endpoint: "192.168.1.200:7001",
		LastSeen: time.Now(),
	}

	rt.Update(contactB)
	rt.Update(contactC)

	if rt.TotalContacts() != 2 {
		t.Fatalf("expected 2 contacts, got: %d", rt.TotalContacts())
	}

	// 3. Buscar los más cercanos a un destino arbitrario
	targetID := NewNodeIDFromDID("did:ipvn7:node-target")
	closest := rt.FindClosest(targetID, 1)
	if len(closest) != 1 {
		t.Fatalf("expected 1 closest contact, got: %d", len(closest))
	}

	// 4. Actualizar contacto existente (debe refrescar LastSeen sin duplicar)
	updatedB := contactB
	updatedB.LastSeen = time.Now().Add(5 * time.Second)
	rt.Update(updatedB)

	if rt.TotalContacts() != 2 {
		t.Fatalf("expected still 2 contacts after updating node B, got: %d", rt.TotalContacts())
	}
}
