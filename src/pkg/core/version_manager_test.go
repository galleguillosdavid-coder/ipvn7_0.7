package core

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestVersionManager_CompareSemVer(t *testing.T) {
	cases := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"0.7.0", "0.7.0", 0},
		{"v0.7.0", "0.7.0", 0},
		{"0.7.1", "0.7.0", 1},
		{"0.7.0", "0.7.1", -1},
		{"0.8.0", "0.7.99", 1},
		{"1.0.0", "0.9.9", 1},
		{"0.7.0-rc1", "0.7.0", -1},
	}

	for _, c := range cases {
		res := CompareSemVer(c.v1, c.v2)
		if res != c.expected {
			t.Errorf("CompareSemVer(%s, %s) = %d, esperado %d", c.v1, c.v2, res, c.expected)
		}
	}
}

func TestVersionManager_AtomicUpdateAndRollbackCycle(t *testing.T) {
	// 1. Generar autoridad de actualización soberana
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Fallo generando par Ed25519: %v", err)
	}

	tempDir := t.TempDir()
	execPath := filepath.Join(tempDir, "ipvn7.exe")

	// Crear binario "actual" v0.7.0
	currentPayload := []byte("IPVN7_BINARY_V0.7.0_MOCK")
	if err := os.WriteFile(execPath, currentPayload, 0755); err != nil {
		t.Fatalf("Fallo escribiendo binario simulado: %v", err)
	}

	engine := NewAtomicUpdateEngine("0.7.0", pub)

	// Crear binario "nuevo" v0.7.1
	newPayload := []byte("IPVN7_BINARY_V0.7.1_MOCK_NEW_FEATURES")
	h := sha256.Sum256(newPayload)
	newHashHex := hex.EncodeToString(h[:])

	// 2. Crear manifiesto firmado válido
	manifest := &VersionManifest{
		Version:          "0.7.1",
		BinarySHA256:     newHashHex,
		ReleaseTimestamp: time.Now().Unix(),
		MinCompat:        "0.7.0",
	}
	sig := ed25519.Sign(priv, manifest.CanonicalPayload())
	manifest.Signature = hex.EncodeToString(sig)

	// 3. Test: Rechazo por firma alterada
	badManifest := *manifest
	badManifest.Signature = hex.EncodeToString(make([]byte, 64))
	if err := engine.ApplyAtomicUpdate(execPath, newPayload, &badManifest); err != ErrInvalidSignature {
		t.Fatalf("Debio fallar con ErrInvalidSignature, obtenido: %v", err)
	}

	// 4. Test: Rechazo por hash SHA-256 no coincidente
	corruptedPayload := []byte("CORRUPTED_DOWNLOAD")
	if err := engine.ApplyAtomicUpdate(execPath, corruptedPayload, manifest); err != ErrBinaryHashMismatch {
		t.Fatalf("Debio fallar con ErrBinaryHashMismatch, obtenido: %v", err)
	}

	// 5. Test: Rechazo por intento de Downgrade
	downgradeManifest := *manifest
	downgradeManifest.Version = "0.6.9"
	downgradeSig := ed25519.Sign(priv, downgradeManifest.CanonicalPayload())
	downgradeManifest.Signature = hex.EncodeToString(downgradeSig)
	if err := engine.ApplyAtomicUpdate(execPath, newPayload, &downgradeManifest); err != ErrDowngradeForbidden {
		t.Fatalf("Debio fallar con ErrDowngradeForbidden, obtenido: %v", err)
	}

	// 6. Aplicar Actualización Atómica Exitosa
	if err := engine.ApplyAtomicUpdate(execPath, newPayload, manifest); err != nil {
		t.Fatalf("Actualizacion atomica fallo: %v", err)
	}

	if engine.CurrentVersion() != "0.7.1" {
		t.Fatalf("Version no actualizada en memoria: %s", engine.CurrentVersion())
	}

	// Verificar contenido en disco del ejecutable principal
	updatedContent, err := os.ReadFile(execPath)
	if err != nil || string(updatedContent) != string(newPayload) {
		t.Fatalf("Contenido en disco no coincide con la nueva version")
	}

	// Verificar existencia del respaldo .old
	oldPath := execPath + ".old"
	oldContent, err := os.ReadFile(oldPath)
	if err != nil || string(oldContent) != string(currentPayload) {
		t.Fatalf("Respaldo .old no contiene la version anterior")
	}

	// 7. Simular fallo de salud en runtime y ejecutar Rollback atómico
	if err := engine.Rollback(execPath); err != nil {
		t.Fatalf("Rollback atomico fallo: %v", err)
	}

	// Verificar que el binario activo volvio a ser el original
	restoredContent, err := os.ReadFile(execPath)
	if err != nil || string(restoredContent) != string(currentPayload) {
		t.Fatalf("Rollback no restauro el binario original")
	}

	// Verificar que el binario fallido quedo preservado como .bad para analisis forense
	badPath := execPath + ".bad"
	if _, err := os.Stat(badPath); os.IsNotExist(err) {
		t.Fatalf("El binario defectuoso debio ser preservado como .bad")
	}
}
