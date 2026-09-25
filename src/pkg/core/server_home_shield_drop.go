package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"ipvn7/pkg/l1"
	"ipvn7/pkg/l2"
)

// handleHomeIoTShield gestiona la micro-segmentación ZTNA sobre dispositivos en la LAN
func (s *CoreServer) handleHomeIoTShield(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Enabled bool `json:"enabled"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	entries := ReadLocalARPTable()
	rulesApplied := 0
	for _, entry := range entries {
		shadowDID := fmt.Sprintf("did:ipvn7:shadow:mac:%s", strings.ReplaceAll(entry.MAC, ":", ""))
		policy := &l1.DIDPolicy{
			DID:           shadowDID,
			AllowInbound:  !req.Enabled,
			AllowOutbound: !req.Enabled,
			AllowRelay:    true,
		}
		s.firewall.AuthorizeDID(policy)
		rulesApplied++
	}

	statusMsg := fmt.Sprintf("Blindaje ZTNA Activado: %d dispositivos LAN micro-segmentados", rulesApplied)
	if !req.Enabled {
		statusMsg = fmt.Sprintf("Blindaje ZTNA en modo permisivo para %d dispositivos LAN", rulesApplied)
	}

	s.gateway.PublishEvent(MeshEvent{
		Type:      "HOME_IOT_SHIELD",
		Timestamp: time.Now().UnixNano(),
		Source:    s.identity.DID(),
		Payload: map[string]interface{}{
			"enabled":       req.Enabled,
			"rules_applied": rulesApplied,
			"default_deny":  s.firewall.IsDefaultDeny(),
		},
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":          "updated",
		"enabled":         req.Enabled,
		"rules_applied":   rulesApplied,
		"default_deny":    s.firewall.IsDefaultDeny(),
		"active_policies": len(s.firewall.GetAllPolicies()),
		"message":         statusMsg,
	})
}

// handleHomeDrop gestiona el buzón de recepción de archivos ultrarrápido sin intermediarios
func (s *CoreServer) handleHomeDrop(w http.ResponseWriter, r *http.Request) {
	dropDir := filepath.Join("data", "local_drop")
	_ = os.MkdirAll(dropDir, 0755)

	if r.Method == http.MethodGet {
		targetFile := r.URL.Query().Get("download")
		if targetFile != "" {
			safeName := filepath.Base(targetFile)
			filePath := filepath.Join(dropDir, safeName)
			if _, err := os.Stat(filePath); err == nil {
				w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", safeName))
				http.ServeFile(w, r, filePath)
				return
			}
			http.Error(w, "Archivo no encontrado", http.StatusNotFound)
			return
		}
		// Listar archivos disponibles
		entries, _ := os.ReadDir(dropDir)
		files := make([]map[string]interface{}, 0)
		for _, e := range entries {
			if !e.IsDir() {
				info, _ := e.Info()
				files = append(files, map[string]interface{}{
					"name": e.Name(),
					"size": info.Size(),
					"url":  fmt.Sprintf("/api/v1/home/drop?download=%s", e.Name()),
				})
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(files)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	startTime := time.Now()
	var fileName string
	var totalBytes int64
	hasher := sha256.New()

	if strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
		if err := r.ParseMultipartForm(128 << 20); err != nil {
			http.Error(w, "Error procesando formulario multipart: "+err.Error(), http.StatusBadRequest)
			return
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Archivo no encontrado en el formulario", http.StatusBadRequest)
			return
		}
		defer file.Close()

		fileName = filepath.Base(header.Filename)
		dstPath := filepath.Join(dropDir, fileName)
		dstFile, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			http.Error(w, "Error creando archivo en disco: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer dstFile.Close()

		mw := io.MultiWriter(dstFile, hasher)
		n, err := io.Copy(mw, file)
		if err != nil {
			http.Error(w, "Error escribiendo archivo: "+err.Error(), http.StatusInternalServerError)
			return
		}
		totalBytes = n
	} else {
		var req struct {
			FileName string `json:"file_name"`
			FileSize int64  `json:"file_size"`
			Content  string `json:"content"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		fileName = req.FileName
		if fileName == "" {
			fileName = fmt.Sprintf("drop_%d.bin", time.Now().Unix())
		}
		dstPath := filepath.Join(dropDir, fileName)
		data := []byte(req.Content)
		if len(data) == 0 && req.FileSize > 0 {
			data = make([]byte, req.FileSize)
		}
		_ = os.WriteFile(dstPath, data, 0644)
		hasher.Write(data)
		totalBytes = int64(len(data))
	}

	elapsed := time.Since(startTime)
	sec := elapsed.Seconds()
	if sec < 0.000001 {
		sec = 0.000001
	}
	speedMbps := (float64(totalBytes) * 8.0) / (sec * 1e6)
	checksum := hex.EncodeToString(hasher.Sum(nil))

	s.telemetry.RecordEvent(l2.EventTxPacket, uint32(totalBytes), uint32(elapsed.Microseconds()), 0)
	s.gateway.PublishEvent(MeshEvent{
		Type:      "HOME_DROP_RECEIVED",
		Timestamp: time.Now().UnixNano(),
		Source:    s.identity.DID(),
		Payload: map[string]interface{}{
			"file_name":  fileName,
			"size_bytes": totalBytes,
			"speed_mbps": speedMbps,
			"sha256":     checksum,
			"elapsed_ms": float64(elapsed.Microseconds()) / 1000.0,
		},
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "delivered",
		"file_name":   fileName,
		"size_bytes":  totalBytes,
		"speed_mbps":  speedMbps,
		"sha256":      checksum,
		"elapsed_ms":  float64(elapsed.Microseconds()) / 1000.0,
		"storage_dir": dropDir,
		"download_url": fmt.Sprintf("/api/v1/home/drop?download=%s", fileName),
		"message":     fmt.Sprintf("Archivo '%s' (%d bytes) entregado a %.2f Mbps reales", fileName, totalBytes, speedMbps),
	})
}

// handleHomeDropOpen abre la carpeta del buzón en el explorador de archivos nativo
func (s *CoreServer) handleHomeDropOpen(w http.ResponseWriter, r *http.Request) {
	dropDir := filepath.Join("data", "local_drop")
	_ = os.MkdirAll(dropDir, 0755)
	if runtime.GOOS == "windows" {
		absPath, _ := filepath.Abs(dropDir)
		_ = exec.Command("explorer.exe", absPath).Start()
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "opened",
		"dir":    dropDir,
	})
}
