// Package dfs implementa el Sistema de Almacenamiento de Archivos Distribuido
// basado en direccionamiento por contenido (CAS), bloques SHA-256 y firmas Ed25519.
// Cumple con la regla de atomicidad modular (<=400 líneas) de docs/INGENIERIA_LEAN.md.
package dfs

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"ipvn7/pkg/l0"
)

const (
	// DefaultChunkSize tamaño canónico de fragmento (64 KB)
	DefaultChunkSize = 64 * 1024
	// ManifestVersion versión actual del esquema de manifiesto
	ManifestVersion = "1.0.0"
)

var (
	ErrChunkNotFound   = errors.New("dfs: bloque no encontrado en el almacén")
	ErrCorruptChunk    = errors.New("dfs: integridad de bloque comprometida (hash no coincide)")
	ErrInvalidManifest = errors.New("dfs: manifiesto inválido o corrupto")
	ErrSignatureFailed = errors.New("dfs: firma criptográfica del manifiesto inválida")
)

// FileManifest describe un archivo indexado por contenido en el DFS
type FileManifest struct {
	Version     string   `json:"version"`
	FileName    string   `json:"file_name"`
	TotalSize   int64    `json:"total_size"`
	ChunkSize   int      `json:"chunk_size"`
	ChunkHashes []string `json:"chunk_hashes"`
	AuthorDID   string   `json:"author_did"`
	Timestamp   int64    `json:"timestamp"`
	Signature   []byte   `json:"signature,omitempty"`
}

// ChunkStore administra la persistencia local de bloques indexados por SHA-256
type ChunkStore struct {
	mu       sync.RWMutex
	baseDir  string
	blockDir string
	metaDir  string
}

// NewChunkStore inicializa el almacén de bloques local
func NewChunkStore(baseDir string) (*ChunkStore, error) {
	blockDir := filepath.Join(baseDir, "blocks")
	metaDir := filepath.Join(baseDir, "manifests")

	if err := os.MkdirAll(blockDir, 0755); err != nil {
		return nil, fmt.Errorf("error creando directorio de bloques: %w", err)
	}
	if err := os.MkdirAll(metaDir, 0755); err != nil {
		return nil, fmt.Errorf("error creando directorio de manifiestos: %w", err)
	}

	return &ChunkStore{
		baseDir:  baseDir,
		blockDir: blockDir,
		metaDir:  metaDir,
	}, nil
}

// PutChunk almacena un bloque de datos si su hash no existe aún
func (cs *ChunkStore) PutChunk(data []byte) (string, error) {
	h := sha256.Sum256(data)
	hashHex := hex.EncodeToString(h[:])

	cs.mu.Lock()
	defer cs.mu.Unlock()

	targetPath := filepath.Join(cs.blockDir, hashHex)
	if _, err := os.Stat(targetPath); err == nil {
		return hashHex, nil // Ya almacenado (idempotencia y deduplicación)
	}

	tmpPath := targetPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return "", fmt.Errorf("error escribiendo bloque temporal: %w", err)
	}
	if err := os.Rename(tmpPath, targetPath); err != nil {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("error consolidando bloque: %w", err)
	}

	return hashHex, nil
}

// GetChunk recupera un bloque verificando estrictamente su integridad SHA-256
func (cs *ChunkStore) GetChunk(hashHex string) ([]byte, error) {
	cs.mu.RLock()
	targetPath := filepath.Join(cs.blockDir, hashHex)
	cs.mu.RUnlock()

	data, err := os.ReadFile(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrChunkNotFound
		}
		return nil, err
	}

	h := sha256.Sum256(data)
	if hex.EncodeToString(h[:]) != hashHex {
		return nil, ErrCorruptChunk
	}

	return data, nil
}

// HasChunk comprueba la presencia de un bloque en el almacén local
func (cs *ChunkStore) HasChunk(hashHex string) bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	_, err := os.Stat(filepath.Join(cs.blockDir, hashHex))
	return err == nil
}

// StoreFile fragmenta un archivo completo, guarda sus bloques y emite un manifiesto firmado
func (cs *ChunkStore) StoreFile(data []byte, filename string, author *l0.Identity) (*FileManifest, string, error) {
	totalLen := len(data)
	numChunks := (totalLen + DefaultChunkSize - 1) / DefaultChunkSize
	if totalLen == 0 {
		numChunks = 1
	}

	chunkHashes := make([]string, 0, numChunks)

	for i := 0; i < totalLen; i += DefaultChunkSize {
		end := i + DefaultChunkSize
		if end > totalLen {
			end = totalLen
		}
		chunkData := data[i:end]
		hashHex, err := cs.PutChunk(chunkData)
		if err != nil {
			return nil, "", fmt.Errorf("error guardando bloque %d: %w", len(chunkHashes), err)
		}
		chunkHashes = append(chunkHashes, hashHex)
	}

	manifest := &FileManifest{
		Version:     ManifestVersion,
		FileName:    filename,
		TotalSize:   int64(totalLen),
		ChunkSize:   DefaultChunkSize,
		ChunkHashes: chunkHashes,
		Timestamp:   time.Now().Unix(),
	}

	if author != nil {
		manifest.AuthorDID = author.DID()
		manifestBytes, _ := json.Marshal(struct {
			Version     string   `json:"version"`
			FileName    string   `json:"file_name"`
			TotalSize   int64    `json:"total_size"`
			ChunkHashes []string `json:"chunk_hashes"`
			AuthorDID   string   `json:"author_did"`
			Timestamp   int64    `json:"timestamp"`
		}{
			Version:     manifest.Version,
			FileName:    manifest.FileName,
			TotalSize:   manifest.TotalSize,
			ChunkHashes: manifest.ChunkHashes,
			AuthorDID:   manifest.AuthorDID,
			Timestamp:   manifest.Timestamp,
		})
		manifest.Signature = author.Sign(manifestBytes)
	}

	cid, err := cs.SaveManifest(manifest)
	if err != nil {
		return nil, "", err
	}

	return manifest, cid, nil
}

// RetrieveFile reensambla un archivo completo a partir de sus bloques descritos en el manifiesto
func (cs *ChunkStore) RetrieveFile(manifest *FileManifest) ([]byte, error) {
	if manifest == nil || len(manifest.ChunkHashes) == 0 {
		return nil, ErrInvalidManifest
	}

	result := make([]byte, 0, manifest.TotalSize)

	for idx, chunkHash := range manifest.ChunkHashes {
		chunkData, err := cs.GetChunk(chunkHash)
		if err != nil {
			return nil, fmt.Errorf("error recuperando bloque %d (%s): %w", idx, chunkHash, err)
		}
		result = append(result, chunkData...)
	}

	if int64(len(result)) != manifest.TotalSize {
		return nil, fmt.Errorf("dfs: tamaño reconstruido (%d) no coincide con manifiesto (%d)", len(result), manifest.TotalSize)
	}

	return result, nil
}

// SaveManifest serializa y persiste un manifiesto identificado por su hash SHA-256 (CID)
func (cs *ChunkStore) SaveManifest(manifest *FileManifest) (string, error) {
	data, err := json.Marshal(manifest)
	if err != nil {
		return "", fmt.Errorf("error serializando manifiesto: %w", err)
	}

	h := sha256.Sum256(data)
	cid := "cid:" + hex.EncodeToString(h[:])

	cs.mu.Lock()
	defer cs.mu.Unlock()

	targetPath := filepath.Join(cs.metaDir, hex.EncodeToString(h[:])+jsonExt)
	if err := os.WriteFile(targetPath, data, 0644); err != nil {
		return "", fmt.Errorf("error guardando manifiesto: %w", err)
	}

	return cid, nil
}

// LoadManifest carga un manifiesto desde el almacenamiento local por su CID
func (cs *ChunkStore) LoadManifest(cid string) (*FileManifest, error) {
	cleanHash := cid
	if len(cleanHash) > 4 && cleanHash[:4] == "cid:" {
		cleanHash = cleanHash[4:]
	}

	cs.mu.RLock()
	targetPath := filepath.Join(cs.metaDir, cleanHash+jsonExt)
	cs.mu.RUnlock()

	data, err := os.ReadFile(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("dfs: manifiesto '%s' no encontrado", cid)
		}
		return nil, err
	}

	var manifest FileManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, ErrInvalidManifest
	}

	return &manifest, nil
}

// VerifySignature verifica la firma criptográfica del manifiesto
func (m *FileManifest) VerifySignature(pubKey []byte) bool {
	if len(m.Signature) == 0 || len(pubKey) == 0 {
		return false
	}
	manifestBytes, _ := json.Marshal(struct {
		Version     string   `json:"version"`
		FileName    string   `json:"file_name"`
		TotalSize   int64    `json:"total_size"`
		ChunkHashes []string `json:"chunk_hashes"`
		AuthorDID   string   `json:"author_did"`
		Timestamp   int64    `json:"timestamp"`
	}{
		Version:     m.Version,
		FileName:    m.FileName,
		TotalSize:   m.TotalSize,
		ChunkHashes: m.ChunkHashes,
		AuthorDID:   m.AuthorDID,
		Timestamp:   m.Timestamp,
	})
	return l0.VerifySignature(pubKey, manifestBytes, m.Signature)
}

const jsonExt = ".json"
