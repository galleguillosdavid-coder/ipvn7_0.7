package dfs

import (
	"bytes"
	"crypto/rand"
	"os"
	"path/filepath"
	"testing"

	"ipvn7/pkg/l0"
)

func TestChunkStore_StoreAndRetrieve(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ipvn7_dfs_test_*")
	if err != nil {
		t.Fatalf("Error creando directorio temporal: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cs, err := NewChunkStore(tempDir)
	if err != nil {
		t.Fatalf("Error inicializando ChunkStore: %v", err)
	}

	id, err := l0.GenerateIdentity()
	if err != nil {
		t.Fatalf("Error generando identidad: %v", err)
	}

	// Archivo de 150 KB (3 bloques de 64 KB)
	payloadSize := 150 * 1024
	originalData := make([]byte, payloadSize)
	_, _ = rand.Read(originalData)

	manifest, cid, err := cs.StoreFile(originalData, "test_payload.bin", id)
	if err != nil {
		t.Fatalf("StoreFile falló: %v", err)
	}

	if len(manifest.ChunkHashes) != 3 {
		t.Fatalf("Se esperaban 3 bloques, obtenidos %d", len(manifest.ChunkHashes))
	}
	if cid == "" || len(cid) < 10 {
		t.Fatalf("CID inválido: %s", cid)
	}

	// Verificar firma del autor
	if !manifest.VerifySignature(id.PublicKey) {
		t.Fatalf("La firma del manifiesto no pasó la verificación Ed25519")
	}

	// Recuperar por CID
	loadedManifest, err := cs.LoadManifest(cid)
	if err != nil {
		t.Fatalf("LoadManifest falló: %v", err)
	}
	if loadedManifest.FileName != "test_payload.bin" || loadedManifest.TotalSize != int64(payloadSize) {
		t.Fatalf("Metadatos de manifiesto cargado no coinciden")
	}

	retrievedData, err := cs.RetrieveFile(loadedManifest)
	if err != nil {
		t.Fatalf("RetrieveFile falló: %v", err)
	}

	if !bytes.Equal(originalData, retrievedData) {
		t.Fatalf("Los datos reconstruidos no coinciden con los originales")
	}
}

func TestChunkStore_CorruptChunkDetection(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ipvn7_dfs_corrupt_*")
	if err != nil {
		t.Fatalf("Error creando directorio temporal: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cs, err := NewChunkStore(tempDir)
	if err != nil {
		t.Fatalf("Error inicializando ChunkStore: %v", err)
	}

	data := []byte("bloque_canónico_ipvn7")
	hashHex, err := cs.PutChunk(data)
	if err != nil {
		t.Fatalf("PutChunk falló: %v", err)
	}

	// Corromper manualmente el archivo en disco
	blockPath := filepath.Join(tempDir, "blocks", hashHex)
	_ = os.WriteFile(blockPath, []byte("datos_corrompidos_alterados"), 0644)

	// GetChunk debe fallar con ErrCorruptChunk
	_, err = cs.GetChunk(hashHex)
	if err != ErrCorruptChunk {
		t.Fatalf("Se esperaba ErrCorruptChunk, obtenido: %v", err)
	}
}
