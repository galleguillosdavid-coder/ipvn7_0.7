// Package core implementa los manejadores HTTP REST para el Almacén Distribuido (DFS)
// y el Colector Descentralizado de Diagnósticos y Errores de Malla.
// Cumple con la regla de atomicidad modular (<=400 líneas) de docs/INGENIERIA_LEAN.md.
package core

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// handleTelemetryReport procesa la ingesta distribuida de reportes de diagnóstico y error
func (s *CoreServer) handleTelemetryReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	if s.collector == nil {
		http.Error(w, "Colector distribuido no disponible", http.StatusServiceUnavailable)
		return
	}

	var report DiagnosticReport
	if err := json.NewDecoder(r.Body).Decode(&report); err != nil {
		http.Error(w, "Formato JSON de reporte inválido: "+err.Error(), http.StatusBadRequest)
		return
	}

	reportID, err := s.collector.IngestReport(&report)
	if err != nil && err != ErrDuplicateReport {
		http.Error(w, "Error procesando reporte: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Notificar en tiempo real al bus de eventos y la UI
	if s.gateway != nil {
		s.gateway.PublishEvent(MeshEvent{
			Type:      "DIAGNOSTIC_REPORT_RECEIVED",
			Timestamp: time.Now().UnixNano(),
			Source:    report.ReporterDID,
			Payload: map[string]interface{}{
				"report_id":    reportID,
				"anomaly_type": report.AnomalyType,
				"severity":     report.Severity,
				"version":      report.Version,
				"os":           report.OS,
				"duplicate":    err == ErrDuplicateReport,
			},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ingested",
		"report_id": reportID,
		"duplicate": err == ErrDuplicateReport,
		"timestamp": time.Now().Unix(),
	})
}

// handleTelemetryReports entrega los reportes diagnósticos más recientes para auditoría
func (s *CoreServer) handleTelemetryReports(w http.ResponseWriter, r *http.Request) {
	if s.collector == nil {
		http.Error(w, "Colector distribuido no disponible", http.StatusServiceUnavailable)
		return
	}

	reports := s.collector.GetRecentReports(50)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":        "ok",
		"total_reports": s.collector.TotalReports(),
		"reports":       reports,
	})
}

// handleDFSUpload fragmenta y almacena un archivo en el DFS soberano emitiendo su CID
func (s *CoreServer) handleDFSUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	if s.chunkStore == nil {
		http.Error(w, "Almacén distribuido DFS no inicializado", http.StatusServiceUnavailable)
		return
	}

	var data []byte
	var fileName string

	if strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
		if err := r.ParseMultipartForm(128 << 20); err != nil {
			http.Error(w, "Error procesando formulario multipart: "+err.Error(), http.StatusBadRequest)
			return
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Archivo no adjunto", http.StatusBadRequest)
			return
		}
		defer file.Close()
		fileName = filepath.Base(header.Filename)
		data, err = io.ReadAll(file)
		if err != nil {
			http.Error(w, "Error leyendo archivo: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		var req struct {
			FileName string `json:"file_name"`
			Content  string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Payload JSON inválido", http.StatusBadRequest)
			return
		}
		fileName = req.FileName
		data = []byte(req.Content)
	}

	if fileName == "" {
		fileName = fmt.Sprintf("file_%d.bin", time.Now().Unix())
	}

	manifest, cid, err := s.chunkStore.StoreFile(data, fileName, s.identity)
	if err != nil {
		http.Error(w, "Error almacenando archivo en DFS: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Notificar evento
	if s.gateway != nil {
		s.gateway.PublishEvent(MeshEvent{
			Type:      "DFS_FILE_STORED",
			Timestamp: time.Now().UnixNano(),
			Source:    s.identity.DID(),
			Payload: map[string]interface{}{
				"cid":          cid,
				"file_name":    fileName,
				"total_size":   manifest.TotalSize,
				"chunks_count": len(manifest.ChunkHashes),
			},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       "stored",
		"cid":          cid,
		"file_name":    fileName,
		"total_size":   manifest.TotalSize,
		"chunks_count": len(manifest.ChunkHashes),
		"author_did":   manifest.AuthorDID,
		"timestamp":    manifest.Timestamp,
	})
}

// handleDFSFile descarga y reensambla un archivo completo por su CID
func (s *CoreServer) handleDFSFile(w http.ResponseWriter, r *http.Request) {
	if s.chunkStore == nil {
		http.Error(w, "Almacén distribuido DFS no inicializado", http.StatusServiceUnavailable)
		return
	}

	parts := strings.Split(r.URL.Path, "/api/v1/dfs/file/")
	if len(parts) < 2 || parts[1] == "" {
		http.Error(w, "CID de archivo no especificado", http.StatusBadRequest)
		return
	}
	cid := parts[1]

	manifest, err := s.chunkStore.LoadManifest(cid)
	if err != nil {
		http.Error(w, "Manifiesto no encontrado: "+err.Error(), http.StatusNotFound)
		return
	}

	data, err := s.chunkStore.RetrieveFile(manifest)
	if err != nil {
		http.Error(w, "Error reconstruyendo archivo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", manifest.FileName))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("X-IPVN7-CID", cid)
	_, _ = w.Write(data)
}

// handleDFSManifest devuelve los metadatos y firmas de un archivo indexado en DFS
func (s *CoreServer) handleDFSManifest(w http.ResponseWriter, r *http.Request) {
	if s.chunkStore == nil {
		http.Error(w, "Almacén distribuido DFS no inicializado", http.StatusServiceUnavailable)
		return
	}

	parts := strings.Split(r.URL.Path, "/api/v1/dfs/manifest/")
	if len(parts) < 2 || parts[1] == "" {
		http.Error(w, "CID no especificado", http.StatusBadRequest)
		return
	}
	cid := parts[1]

	manifest, err := s.chunkStore.LoadManifest(cid)
	if err != nil {
		http.Error(w, "Manifiesto no encontrado: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "ok",
		"cid":      cid,
		"manifest": manifest,
	})
}
