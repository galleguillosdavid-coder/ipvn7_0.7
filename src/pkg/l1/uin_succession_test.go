package l1

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"testing"
	"time"
)

func TestSuccessionMigration_FullFlow(t *testing.T) {
	// 1. Inicializar identidad original
	mgr, err := NewUINIdentityManager(IdentityModeHybrid)
	if err != nil {
		t.Fatalf("NewUINIdentityManager fallo: %v", err)
	}

	oldRoot := mgr.RootID()

	// 2. Generar par de sucesión offline (S)
	succPub, succPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey para sucesion fallo: %v", err)
	}
	succHash := sha256.Sum256(succPub)

	// Registrar compromiso criptográfico de sucesión
	mgr.SetSuccessionHash(succHash)

	if mgr.SuccessionHash() != succHash {
		t.Fatal("SuccessionHash getter no coincide")
	}

	passport := mgr.GetPassport()
	if passport.SuccessionHashHex == "" {
		t.Fatal("El pasaporte debe reflejar el hash de sucesion")
	}

	// 3. Simular generación de nueva identidad legítima
	var newRoot [32]byte
	_, _ = rand.Read(newRoot[:])

	// 4. Crear prueba de sucesión válida con la clave offline
	proof, err := GenerateSuccessionProof(oldRoot, newRoot, succPriv)
	if err != nil {
		t.Fatalf("GenerateSuccessionProof fallo: %v", err)
	}

	// 5. Aplicar migración de emergencia
	if err := mgr.ApplySuccessionMigration(proof); err != nil {
		t.Fatalf("ApplySuccessionMigration fallo con prueba valida: %v", err)
	}

	// 6. Verificar que la entidad ahora opera bajo newRoot
	if mgr.RootID() != newRoot {
		t.Fatalf("El RootID tras la sucesion debio ser %x, pero fue %x", newRoot, mgr.RootID())
	}
}

func TestSuccessionMigration_RejectsTamperedProof(t *testing.T) {
	mgr, _ := NewUINIdentityManager(IdentityModeHybrid)
	oldRoot := mgr.RootID()

	_, succPriv, _ := ed25519.GenerateKey(rand.Reader)
	otherPub, _, _ := ed25519.GenerateKey(rand.Reader)
	mgr.SetSuccessionHash(sha256.Sum256(otherPub)) // Hash registrado de OTRA clave

	var newRoot [32]byte
	_, _ = rand.Read(newRoot[:])

	proof, _ := GenerateSuccessionProof(oldRoot, newRoot, succPriv)

	// Debe fallar porque la clave usada no coincide con el hash registrado
	err := mgr.ApplySuccessionMigration(proof)
	if err == nil {
		t.Fatal("Se esperaba error al intentar migrar con clave de sucesion que no coincide con el hash")
	}
}

func TestSuccessionMigration_RejectsExpiredProof(t *testing.T) {
	succPub, succPriv, _ := ed25519.GenerateKey(rand.Reader)
	succHash := sha256.Sum256(succPub)

	var oldRoot, newRoot [32]byte
	_, _ = rand.Read(oldRoot[:])
	_, _ = rand.Read(newRoot[:])

	proof, _ := GenerateSuccessionProof(oldRoot, newRoot, succPriv)
	proof.Timestamp = time.Now().UTC().Add(-72 * time.Hour) // 72 horas atrás (supera ventana de 48h)
	// Regenerar firma para el timestamp adulterado
	proof.Signature = ed25519.Sign(succPriv, proof.CanonicalBytes())

	valid, err := VerifySuccessionMigration(succHash, proof)
	if valid || err == nil {
		t.Fatal("Se esperaba rechazo de prueba de sucesion expirada (>48h)")
	}
}
