package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestVersionManager_GetVersionInfo(t *testing.T) {
	v := GetVersionInfo()
	if v.Version == "" || v.WireVersion == 0 {
		t.Errorf("Metadatos de versión vacíos: %+v", v)
	}
}

func TestVersionManager_VerifyFileSHA256(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_binary.bin")
	content := []byte("IPVN7_BINARY_RELEASE_PAYLOAD_TEST")
	if err := os.WriteFile(testFile, content, 0755); err != nil {
		t.Fatalf("error creando archivo test: %v", err)
	}

	h := sha256.Sum256(content)
	correctHash := hex.EncodeToString(h[:])

	// 1. Hash correcto
	if err := VerifyFileSHA256(testFile, correctHash); err != nil {
		t.Errorf("Esperado hash válido, fallo: %v", err)
	}

	// 2. Hash discordante
	tamperedHash := "0000000000000000000000000000000000000000000000000000000000000000"
	if err := VerifyFileSHA256(testFile, tamperedHash); err == nil {
		t.Errorf("Esperado error ante hash corrupto o manipulado")
	}
}

func TestVersionManager_AtomicUpdateAndRollback(t *testing.T) {
	tmpDir := t.TempDir()
	activeExe := filepath.Join(tmpDir, "ipvn7.exe")
	newExe := filepath.Join(tmpDir, "ipvn7_v2.exe")

	origContent := []byte("VERSION_1_ORIGINAL")
	newContent := []byte("VERSION_2_UPDATED")

	if err := os.WriteFile(activeExe, origContent, 0755); err != nil {
		t.Fatalf("Error escribiendo activeExe: %v", err)
	}
	if err := os.WriteFile(newExe, newContent, 0755); err != nil {
		t.Fatalf("Error escribiendo newExe: %v", err)
	}

	vm := NewVersionManager(activeExe)

	// 1. Ejecutar actualización atómica
	hNew := sha256.Sum256(newContent)
	expectedHash := hex.EncodeToString(hNew[:])
	if err := vm.PrepareAtomicUpdate(newExe, expectedHash); err != nil {
		t.Fatalf("Error en PrepareAtomicUpdate: %v", err)
	}

	// Verificar que el ejecutable activo ahora tiene el contenido nuevo
	readNew, err := os.ReadFile(activeExe)
	if err != nil || string(readNew) != string(newContent) {
		t.Fatalf("El binario activo no se actualizó correctamente: %s", string(readNew))
	}

	// Verificar que existe el backup .old
	if _, err := os.Stat(activeExe + ".old"); err != nil {
		t.Fatalf("No se creó el archivo de respaldo .old")
	}

	// 2. Simular fallo en health check y ejecutar Rollback
	if err := vm.ExecuteRollback(); err != nil {
		t.Fatalf("Error en ExecuteRollback: %v", err)
	}

	// Verificar que el binario activo fue restaurado a la versión 1
	restored, err := os.ReadFile(activeExe)
	if err != nil || string(restored) != string(origContent) {
		t.Fatalf("El rollback falló en restaurar la versión 1: %s", string(restored))
	}

	// 3. Probar camino de éxito (ConfirmSuccess)
	hOrig := sha256.Sum256(origContent)
	origHash := hex.EncodeToString(hOrig[:])
	if err := vm.PrepareAtomicUpdate(activeExe, origHash); err == nil {
		_ = vm.ConfirmSuccess()
	}

	// 4. Probar que hash vacío es rechazado
	if err := vm.PrepareAtomicUpdate(activeExe, ""); err == nil {
		t.Fatalf("PrepareAtomicUpdate debió rechazar hash vacío")
	}
}

func TestVersionManager_SemVerAndAntiRollback(t *testing.T) {
	cases := []struct {
		current   string
		candidate string
		higher    bool
	}{
		{"v0.7.0", "v0.7.1", true},
		{"v0.7.0", "v0.8.0", true},
		{"v0.7.0", "v1.0.0", true},
		{"v0.7.0", "v0.7.9-future", true},
		{"v0.7.0", "v0.7.0", false},        // Misma versión
		{"v0.7.1", "v0.7.0", false},        // Downgrade / Rollback attack
		{"v1.0.0", "v0.9.9", false},        // Downgrade
		{"v0.7.0", "invalid", false},
	}

	for _, tc := range cases {
		res := IsHigherVersion(tc.current, tc.candidate)
		if res != tc.higher {
			t.Errorf("IsHigherVersion(%s, %s) = %v, esperado %v", tc.current, tc.candidate, res, tc.higher)
		}
	}
}

func TestVersionManager_OnlineCheckAndDownload(t *testing.T) {
	binaryContent := []byte("IPVN7_NEW_ONLINE_BINARY_PAYLOAD")
	h := sha256.Sum256(binaryContent)
	expectedHash := hex.EncodeToString(h[:])
	platformKey := runtime.GOOS + "/" + runtime.GOARCH

	// Servidor HTTP local real para emular el CDN / GitHub Releases
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/version_manifest.json":
			manifest := UpdateManifest{
				Version:      "v0.7.9-future",
				BuildVersion: "v0.7.9-future",
				ReleaseNotes: "Actualización de seguridad post-cuántica y rendimiento",
				Platforms: map[string]PlatformRelease{
					platformKey: {
						URL:    fmt.Sprintf("http://%s/binary.bin", r.Host),
						SHA256: expectedHash,
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(manifest)
		case "/binary.bin":
			w.Header().Set("Content-Type", "application/octet-stream")
			_, _ = w.Write(binaryContent)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	tmpDir := t.TempDir()
	activeExe := filepath.Join(tmpDir, "ipvn7.exe")
	if err := os.WriteFile(activeExe, []byte("ORIGINAL_v0.7.0"), 0755); err != nil {
		t.Fatalf("Error creando binario original: %v", err)
	}

	vm := NewVersionManager(activeExe)

	// 1. Consultar actualización en línea
	manifestURL := ts.URL + "/version_manifest.json"
	check, err := vm.CheckOnlineUpdate(manifestURL)
	if err != nil {
		t.Fatalf("Falla en CheckOnlineUpdate: %v", err)
	}
	if !check.Available {
		t.Fatalf("Esperado available=true para versión v0.7.9-future")
	}
	if check.LatestVersion != "v0.7.9-future" {
		t.Fatalf("Versión discordante: %s", check.LatestVersion)
	}

	// 2. Descargar y aplicar actualización en caliente
	if err := vm.DownloadAndApplyUpdate(check, tmpDir); err != nil {
		t.Fatalf("Falla en DownloadAndApplyUpdate: %v", err)
	}

	// 3. Verificar que el binario activo ahora contiene la nueva versión
	updatedBytes, err := os.ReadFile(activeExe)
	if err != nil || string(updatedBytes) != string(binaryContent) {
		t.Fatalf("El binario activo no coincide con el descargado: %s", string(updatedBytes))
	}

	// 4. Confirmar éxito y limpiar backup .old
	if err := vm.ConfirmSuccess(); err != nil {
		t.Fatalf("Falla en ConfirmSuccess: %v", err)
	}
	if _, err := os.Stat(activeExe + ".old"); !os.IsNotExist(err) {
		t.Fatalf("El archivo de backup .old debía eliminarse tras ConfirmSuccess")
	}
}
