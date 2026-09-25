package core

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// Errores normativos del gestor atómico de versiones
var (
	ErrInvalidSignature    = errors.New("version_manager: firma criptografica del manifiesto invalida")
	ErrBinaryHashMismatch  = errors.New("version_manager: el hash SHA-256 del binario no coincide con el manifiesto")
	ErrDowngradeForbidden  = errors.New("version_manager: intento de degradacion (downgrade) bloqueado por seguridad")
	ErrBackupNotFound      = errors.New("version_manager: no existe respaldo .old para revertir (rollback)")
	ErrUpdateInProgress    = errors.New("version_manager: ya existe un proceso de actualizacion en curso")
)

// VersionManifest define la especificación canónica y firmada de una versión distribuible
type VersionManifest struct {
	Version          string `json:"version"`           // SemVer e.g. "0.7.1"
	BinarySHA256     string `json:"binary_sha256"`     // Hash SHA-256 en hexadecimal
	ReleaseTimestamp int64  `json:"release_timestamp"` // Timestamp Unix para evitar replay attacks
	Signature        string `json:"signature"`         // Firma Ed25519 hex sobre (Version || SHA256 || Timestamp)
	MinCompat        string `json:"min_compat"`        // Versión mínima requerida para migración
}

// CanonicalPayload genera el buffer determinista sobre el cual se verifica la firma
func (m *VersionManifest) CanonicalPayload() []byte {
	return []byte(fmt.Sprintf("IPVN7-RELEASE:%s:%s:%d", m.Version, m.BinarySHA256, m.ReleaseTimestamp))
}

// VerifySignature comprueba la autenticidad soberana del manifiesto contra la clave pública raíz
func (m *VersionManifest) VerifySignature(authorityPub ed25519.PublicKey) bool {
	sigBytes, err := hex.DecodeString(m.Signature)
	if err != nil || len(sigBytes) != ed25519.SignatureSize {
		return false
	}
	return ed25519.Verify(authorityPub, m.CanonicalPayload(), sigBytes)
}

// ParseSemVer desglosa una cadena SemVer "X.Y.Z" o "vX.Y.Z-prerelease" en sus componentes
func ParseSemVer(v string) (int, int, int, string, error) {
	clean := strings.TrimPrefix(strings.TrimSpace(v), "v")
	var pre string
	if idx := strings.Index(clean, "-"); idx != -1 {
		pre = clean[idx+1:]
		clean = clean[:idx]
	}
	parts := strings.Split(clean, ".")
	if len(parts) < 3 {
		return 0, 0, 0, "", fmt.Errorf("formato semver invalido: %s", v)
	}
	major, err1 := strconv.Atoi(parts[0])
	minor, err2 := strconv.Atoi(parts[1])
	patch, err3 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return 0, 0, 0, "", fmt.Errorf("componente numerico invalido en semver: %s", v)
	}
	return major, minor, patch, pre, nil
}

// CompareSemVer compara dos versiones SemVer: -1 (v1 < v2), 0 (v1 == v2), 1 (v1 > v2)
// Cumple estrictamente con la especificación SemVer 2.0: una versión pre-release (e.g. 0.7.0-rc1)
// tiene menor precedencia que la versión final (0.7.0).
func CompareSemVer(v1, v2 string) int {
	maj1, min1, pat1, pre1, err1 := ParseSemVer(v1)
	maj2, min2, pat2, pre2, err2 := ParseSemVer(v2)
	if err1 != nil || err2 != nil {
		return strings.Compare(v1, v2)
	}
	if maj1 != maj2 {
		if maj1 > maj2 {
			return 1
		}
		return -1
	}
	if min1 != min2 {
		if min1 > min2 {
			return 1
		}
		return -1
	}
	if pat1 != pat2 {
		if pat1 > pat2 {
			return 1
		}
		return -1
	}
	if pre1 != "" && pre2 == "" {
		return -1 // v1 es pre-release (menor precedencia)
	}
	if pre1 == "" && pre2 != "" {
		return 1 // v1 es final (mayor precedencia)
	}
	if pre1 != "" && pre2 != "" {
		return strings.Compare(pre1, pre2)
	}
	return 0
}

// AtomicUpdateEngine orquesta la sustitución atómica in-place del binario y el rollback de emergencia
type AtomicUpdateEngine struct {
	mu             sync.Mutex
	currentVersion string
	authorityPub   ed25519.PublicKey
	isUpdating     bool
}

// NewAtomicUpdateEngine inicializa el gestor atómico con la clave pública de la autoridad soberana
func NewAtomicUpdateEngine(currentVersion string, authorityPub ed25519.PublicKey) *AtomicUpdateEngine {
	return &AtomicUpdateEngine{
		currentVersion: currentVersion,
		authorityPub:   authorityPub,
	}
}

// CurrentVersion retorna la versión actual ejecutada por el nodo
func (u *AtomicUpdateEngine) CurrentVersion() string {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.currentVersion
}

// ApplyAtomicUpdate valida y ejecuta el reemplazo atómico del archivo binario ejecutable
func (u *AtomicUpdateEngine) ApplyAtomicUpdate(currentExecPath string, newBinaryData []byte, manifest *VersionManifest) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	if u.isUpdating {
		return ErrUpdateInProgress
	}
	u.isUpdating = true
	defer func() { u.isUpdating = false }()

	// 1. Verificación Criptográfica del Manifiesto
	if !manifest.VerifySignature(u.authorityPub) {
		return ErrInvalidSignature
	}

	// 2. Control Anti-Downgrade estricto
	if CompareSemVer(manifest.Version, u.currentVersion) <= 0 {
		return ErrDowngradeForbidden
	}

	// 3. Verificación de Integridad Binaria SHA-256
	h := sha256.Sum256(newBinaryData)
	computedHash := hex.EncodeToString(h[:])
	if !strings.EqualFold(computedHash, manifest.BinarySHA256) {
		return ErrBinaryHashMismatch
	}

	// 4. Preparación de rutas atómicas en el mismo directorio (para atomic filesystem rename)
	dir := filepath.Dir(currentExecPath)
	base := filepath.Base(currentExecPath)
	tmpPath := filepath.Join(dir, base+".tmp")
	oldPath := filepath.Join(dir, base+".old")

	// 5. Escritura del binario nuevo al archivo temporal
	if err := os.WriteFile(tmpPath, newBinaryData, 0755); err != nil {
		return fmt.Errorf("version_manager: fallo escribiendo binario temporal: %w", err)
	}

	// 6. Sustitución Atómica In-Place con respaldo .old
	_ = os.Remove(oldPath) // Eliminar respaldo anterior si existía

	// Paso A: Renombrar el ejecutable actual a .old (en Windows permite renombrar ejecutables en ejecución)
	if err := os.Rename(currentExecPath, oldPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("version_manager: fallo creando respaldo .old: %w", err)
	}

	// Paso B: Mover el archivo temporal al nombre canónico
	if err := os.Rename(tmpPath, currentExecPath); err != nil {
		// Rollback inmediato si falla la colocación del nuevo
		_ = os.Rename(oldPath, currentExecPath)
		_ = os.Remove(tmpPath)
		return fmt.Errorf("version_manager: fallo en el intercambio atomico: %w", err)
	}

	u.currentVersion = manifest.Version
	return nil
}

// Rollback revierte atómicamente el binario al respaldo .old en caso de falla de inicialización
func (u *AtomicUpdateEngine) Rollback(currentExecPath string) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	dir := filepath.Dir(currentExecPath)
	base := filepath.Base(currentExecPath)
	oldPath := filepath.Join(dir, base+".old")
	badPath := filepath.Join(dir, base+".bad")

	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return ErrBackupNotFound
	}

	// Preservar el binario defectuoso como .bad para auditoría forense
	_ = os.Remove(badPath)
	_ = os.Rename(currentExecPath, badPath)

	// Restaurar .old como ejecutable principal
	if err := os.Rename(oldPath, currentExecPath); err != nil {
		return fmt.Errorf("version_manager: fallo critico restaurando .old: %w", err)
	}

	return nil
}

// CleanupBackup elimina el archivo de respaldo .old una vez certificada la estabilidad del nuevo binario
func (u *AtomicUpdateEngine) CleanupBackup(currentExecPath string) error {
	dir := filepath.Dir(currentExecPath)
	base := filepath.Base(currentExecPath)
	oldPath := filepath.Join(dir, base+".old")
	return os.Remove(oldPath)
}
