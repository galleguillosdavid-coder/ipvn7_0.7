package core

import (
	"bytes"
	"testing"
	"time"
)

func TestI7Primitives_ContractualSuite(t *testing.T) {
	// 1. Identity
	id := I7Identity{DID: "did:ipvn7:alice", PublicKey: []byte("pubkey-32b-ed25519-valid-sample-")}
	if id.DID == "" || len(id.PublicKey) == 0 {
		t.Fatalf("Identity inválida: %+v", id)
	}

	// 2. Object
	obj := I7Object{Type: "text/plain", Payload: []byte("hola soberano")}
	if obj.Type != "text/plain" || len(obj.Payload) == 0 {
		t.Fatalf("Object inválido: %+v", obj)
	}

	// 3. Container & 7. MTU
	// Caso válido dentro del invariante (1280B)
	validPayload := make([]byte, 1000)
	tag := []byte("16-byte-tag-poly")
	container, err := NewContainer(1, validPayload, tag)
	if err != nil {
		t.Fatalf("NewContainer falló con tamaño válido: %v", err)
	}
	if container.Version != I7Version || container.ChannelID != 1 {
		t.Fatalf("Datos de contenedor inesperados: %+v", container)
	}

	// Caso de desbordamiento de MTU (>1280B)
	oversizedPayload := make([]byte, 1250) // 1250 + 64 > 1280
	_, err = NewContainer(1, oversizedPayload, tag)
	if err == nil {
		t.Fatalf("NewContainer debió rechazar payload que desborda MTU!")
	}
	if !ValidateMTU(1280) || ValidateMTU(1281) {
		t.Fatalf("ValidateMTU falló aserción de límite 1280")
	}

	// 4. Session
	session := I7Session{
		SessionID: "sess-1",
		LocalDID:  "did:ipvn7:alice",
		RemoteDID: "did:ipvn7:bob",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	if session.IsExpired() {
		t.Fatalf("Session no debería estar expirada")
	}

	expiredSession := I7Session{
		ExpiresAt: time.Now().Add(-1 * time.Second),
	}
	if !expiredSession.IsExpired() {
		t.Fatalf("Expired session debió reportar IsExpired=true")
	}

	// 5. Channel
	ch := I7Channel{ID: 10, Priority: 1, Active: true}
	if !ch.Active || ch.ID != 10 {
		t.Fatalf("Channel inválido: %+v", ch)
	}

	// 6. Path
	path := I7Path{SourceDID: "did:ipvn7:a", TargetDID: "did:ipvn7:b", HopCount: 0, MaxHops: 3}
	for i := 0; i < 3; i++ {
		if err := path.IncrementHop(); err != nil {
			t.Fatalf("IncrementHop falló en salto %d: %v", i, err)
		}
	}
	if err := path.IncrementHop(); err == nil {
		t.Fatalf("IncrementHop debió fallar por exceder MaxHops")
	}

	// 8. Integrity
	tagA := []byte("tag-poly1305-16b")
	tagB := []byte("tag-poly1305-16b")
	tagCorrupt := []byte("tag-poly1305-cor")
	if !VerifyIntegrity(tagA, tagB) {
		t.Fatalf("VerifyIntegrity rechazó tags idénticos")
	}
	if VerifyIntegrity(tagA, tagCorrupt) {
		t.Fatalf("VerifyIntegrity aceptó tag corrupto")
	}

	// 9. Routing (XOR Metric)
	hashA := []byte{0x00, 0xFF, 0x55}
	hashB := []byte{0xFF, 0x00, 0xAA}
	expectedMetric := []byte{0xFF, 0xFF, 0xFF}
	metric := XORMetric(hashA, hashB)
	if !bytes.Equal(metric, expectedMetric) {
		t.Fatalf("XORMetric falló: obtenido %x, esperado %x", metric, expectedMetric)
	}

	// 10. Capability
	cap := I7Capability{
		IssuerDID: "did:ipvn7:admin",
		TargetDID: "did:ipvn7:alice",
		Resource:  "network/egress",
		Actions:   []string{"read", "write"},
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	if err := cap.Validate("write"); err != nil {
		t.Fatalf("Capability debió validar acción permitida: %v", err)
	}
	if err := cap.Validate("delete"); err == nil {
		t.Fatalf("Capability no debió autorizar acción 'delete'")
	}
}
