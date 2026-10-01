// Package core implementa el Centinela de Versiones y Actualización Atómica con Rollback (Rol M).
// Garantiza que las actualizaciones de binario sean in-place, verificadas criptográficamente
// y con reversión automática en <5 segundos si el nuevo proceso falla.
package core

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Versión canónica del sistema y faro de distribución
var (
	CurrentVersion           = "v0.7.0"
	BuildVersion             = "v0.7.0"
	DefaultUpdateManifestURL = "https://raw.githubusercontent.com/galleguillosdavid-coder/ipvn7_0.7/main/dist/version_manifest.json"
)

// PlatformRelease describe la descarga y firma de un binario para un OS/Arch específico
type PlatformRelease struct {
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
}

// UpdateManifest encapsula el manifiesto público publicado en el faro de actualización
type UpdateManifest struct {
	Version      string                     `json:"version"`
	BuildVersion string                     `json:"build_version"`
	WireVersion  byte                       `json:"wire_version"`
	ReleaseNotes string                     `json:"release_notes"`
	Platforms    map[string]PlatformRelease `json:"platforms"`
	Signature    string                     `json:"signature,omitempty"` // Firma Ed25519 de release
}

// UpdateCheckResult representa el dictamen de verificación de versión en línea
type UpdateCheckResult struct {
	Available      bool   `json:"available"`
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	ReleaseNotes   string `json:"release_notes"`
	Platform       string `json:"platform"`
	DownloadURL    string `json:"download_url,omitempty"`
	ExpectedSHA256 string `json:"expected_sha256,omitempty"`
}

// VersionInfo encapsula los metadatos de versión y compatibilidad
type VersionInfo struct {
	Version      string    `json:"version"`
	BuildVersion string    `json:"build_version"`
	WireVersion  byte      `json:"wire_version"`
	ReleaseDate  time.Time `json:"release_date"`
	Platform     string    `json:"platform"`
	CommitHash   string    `json:"commit_hash,omitempty"`
}

// GetVersionInfo retorna la ficha técnica de la versión activa
func GetVersionInfo() VersionInfo {
	return VersionInfo{
		Version:      CurrentVersion,
		BuildVersion: BuildVersion,
		WireVersion:  0x07,
		ReleaseDate:  time.Now(),
		Platform:     os.Getenv("GOOS") + "/" + os.Getenv("GOARCH"),
	}
}

// VersionManager orquesta la verificación y reemplazo atómico de binarios
type VersionManager struct {
	mu           sync.Mutex
	activeExe    string
	backupExe    string
	pendingNew   string
	updateActive bool
}

// NewVersionManager crea una instancia del gestor de ciclo de vida
func NewVersionManager(currentExePath string) *VersionManager {
	if currentExePath == "" {
		if exe, err := os.Executable(); err == nil {
			currentExePath = exe
		} else {
			currentExePath = filepath.Join("bin", "ipvn7.exe")
		}
	}
	return &VersionManager{
		activeExe: currentExePath,
		backupExe: currentExePath + ".old",
	}
}

// VerifyFileSHA256 comprueba la integridad criptográfica de un binario contra su hash esperado
func VerifyFileSHA256(filePath, expectedHex string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("no se pudo abrir el archivo %s: %w", filePath, err)
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return fmt.Errorf("error calculando hash SHA256: %w", err)
	}

	actualHex := hex.EncodeToString(hasher.Sum(nil))
	if !strings.EqualFold(actualHex, strings.TrimSpace(expectedHex)) {
		return fmt.Errorf("hash SHA256 discordante: esperado %s, obtenido %s", expectedHex, actualHex)
	}
	return nil
}

// ParseSemVer extrae los componentes mayor, menor y parche de versiones tipo "v0.7.0" o "v0.7.9-future"
func ParseSemVer(v string) (major, minor, patch int, err error) {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	if idx := strings.Index(v, "-"); idx != -1 {
		v = v[:idx]
	}
	parts := strings.Split(v, ".")
	if len(parts) < 3 {
		return 0, 0, 0, fmt.Errorf("formato semver invalido: %s", v)
	}
	_, err = fmt.Sscanf(v, "%d.%d.%d", &major, &minor, &patch)
	return
}

// IsHigherVersion valida estrictamente que la versión candidata sea numéricamente superior (anti-rollback)
func IsHigherVersion(current, candidate string) bool {
	cMaj, cMin, cPat, err1 := ParseSemVer(current)
	tMaj, tMin, tPat, err2 := ParseSemVer(candidate)
	if err1 != nil || err2 != nil {
		return false
	}
	if tMaj > cMaj {
		return true
	}
	if tMaj == cMaj && tMin > cMin {
		return true
	}
	if tMaj == cMaj && tMin == cMin && tPat > cPat {
		return true
	}
	return false
}

// VerifyManifestSignature comprueba la firma digital Ed25519 del manifiesto
func VerifyManifestSignature(m *UpdateManifest, pubKey ed25519.PublicKey) (bool, error) {
	if m == nil || m.Signature == "" {
		return false, errors.New("el manifiesto no contiene firma digital")
	}
	sigBytes, err := hex.DecodeString(m.Signature)
	if err != nil || len(sigBytes) != ed25519.SignatureSize {
		return false, fmt.Errorf("firma digital malformada en manifiesto")
	}
	payload := fmt.Sprintf("%s|%s|%d", m.Version, m.BuildVersion, m.WireVersion)
	if !ed25519.Verify(pubKey, []byte(payload), sigBytes) {
		return false, errors.New("firma digital Ed25519 del manifiesto no coincide")
	}
	return true, nil
}

// CheckOnlineUpdate consulta el faro de versiones y evalúa si existe una versión superior
func (vm *VersionManager) CheckOnlineUpdate(manifestURL string) (*UpdateCheckResult, error) {
	if manifestURL == "" {
		manifestURL = DefaultUpdateManifestURL
	} else if !strings.HasPrefix(manifestURL, "https://raw.githubusercontent.com/galleguillosdavid-coder/ipvn7_0.7/") &&
		!strings.HasPrefix(manifestURL, "http://127.0.0.1:") &&
		!strings.HasPrefix(manifestURL, "http://localhost:") {
		return nil, errors.New("url de manifiesto no autorizada: solo se permiten fuentes oficiales de confianza")
	}
	client := &http.Client{Timeout: 6 * time.Second}
	resp, err := client.Get(manifestURL)
	if err != nil {
		return nil, fmt.Errorf("falla al consultar faro de versión %s: %w", manifestURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("faro respondió con código HTTP %d", resp.StatusCode)
	}

	var m UpdateManifest
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, fmt.Errorf("error al decodificar manifiesto de actualización: %w", err)
	}

	platformKey := runtime.GOOS + "/" + runtime.GOARCH
	rel, hasPlatform := m.Platforms[platformKey]
	isNewer := IsHigherVersion(CurrentVersion, m.Version)

	res := &UpdateCheckResult{
		Available:      isNewer && hasPlatform,
		CurrentVersion: CurrentVersion,
		LatestVersion:  m.Version,
		ReleaseNotes:   m.ReleaseNotes,
		Platform:       platformKey,
	}
	if hasPlatform {
		if !strings.HasPrefix(rel.URL, "https://") &&
			!strings.HasPrefix(rel.URL, "http://127.0.0.1:") &&
			!strings.HasPrefix(rel.URL, "http://localhost:") {
			return nil, errors.New("url de descarga no autorizada: debe ser HTTPS oficial o localhost de pruebas")
		}
		res.DownloadURL = rel.URL
		res.ExpectedSHA256 = rel.SHA256
	}
	return res, nil
}

// DownloadAndApplyUpdate descarga el binario verificado y ejecuta el reemplazo atómico in-place
func (vm *VersionManager) DownloadAndApplyUpdate(check *UpdateCheckResult, tempDir string) error {
	if check == nil || !check.Available || check.DownloadURL == "" {
		return errors.New("no hay actualización disponible para aplicar")
	}
	if tempDir == "" {
		tempDir = os.TempDir()
	}
	tmpFile := filepath.Join(tempDir, fmt.Sprintf("ipvn7_update_%d.tmp", time.Now().UnixNano()))
	defer os.Remove(tmpFile)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(check.DownloadURL)
	if err != nil {
		return fmt.Errorf("error al descargar binario: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("servidor de descargas respondió con código HTTP %d", resp.StatusCode)
	}

	out, err := os.OpenFile(tmpFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("no se pudo crear archivo temporal %s: %w", tmpFile, err)
	}
	if _, err := io.Copy(out, resp.Body); err != nil {
		out.Close()
		return fmt.Errorf("error escribiendo binario descargado: %w", err)
	}
	out.Close()

	// Reemplazo atómico con verificación de integridad SHA-256 forzosa (Rol M)
	return vm.PrepareAtomicUpdate(tmpFile, check.ExpectedSHA256)
}

// PrepareAtomicUpdate prepara y ejecuta el reemplazo seguro del binario
func (vm *VersionManager) PrepareAtomicUpdate(newBinaryPath string, expectedSHA256 string) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	// 1. Verificación criptográfica obligatoria (Rol M: prohibido instalar sin hash válido)
	if strings.TrimSpace(expectedSHA256) == "" {
		return errors.New("verificación criptográfica fallida: hash SHA256 esperado es obligatorio y no puede estar vacío")
	}
	if err := VerifyFileSHA256(newBinaryPath, expectedSHA256); err != nil {
		return fmt.Errorf("verificación criptográfica fallida: %w", err)
	}

	// 2. Limpieza de backups previos
	_ = os.Remove(vm.backupExe)

	// 3. Respaldo atómico del binario actual
	if _, err := os.Stat(vm.activeExe); err == nil {
		if err := os.Rename(vm.activeExe, vm.backupExe); err != nil {
			return fmt.Errorf("falla al crear backup de la versión actual (%s): %w", vm.backupExe, err)
		}
	}

	// 4. Mover o copiar el nuevo binario al destino activo
	if err := copyOrMoveFile(newBinaryPath, vm.activeExe); err != nil {
		// Rollback inmediato ante fallo de escritura
		_ = os.Rename(vm.backupExe, vm.activeExe)
		return fmt.Errorf("falla al desplegar nuevo binario: %w", err)
	}

	vm.updateActive = true
	return nil
}

// ConfirmSuccess consolida la actualización eliminando el binario anterior
func (vm *VersionManager) ConfirmSuccess() error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if !vm.updateActive {
		return nil
	}

	vm.updateActive = false
	if err := os.Remove(vm.backupExe); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("aviso: no se pudo eliminar backup %s: %w", vm.backupExe, err)
	}
	return nil
}

// ExecuteRollback revierte inmediatamente el binario al estado anterior
func (vm *VersionManager) ExecuteRollback() error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if _, err := os.Stat(vm.backupExe); err != nil {
		return errors.New("no existe archivo de backup (.old) para ejecutar rollback")
	}

	_ = os.Remove(vm.activeExe)
	if err := os.Rename(vm.backupExe, vm.activeExe); err != nil {
		return fmt.Errorf("error crítico al restaurar backup: %w", err)
	}

	vm.updateActive = false
	return nil
}

func copyOrMoveFile(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
}
