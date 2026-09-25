// Package core implementa el Colector Distribuido de Rendimiento y Errores de Malla.
// Permite que los nodos actúen como colectores soberanos y almacenen diagnósticos en el DFS.
// Cumple con la regla de atomicidad modular (<=400 líneas) de docs/INGENIERIA_LEAN.md.
package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	"ipvn7/pkg/dfs"
	"ipvn7/pkg/l0"
)

var (
	ErrInvalidReport     = errors.New("collector: reporte de diagnóstico inválido")
	ErrDuplicateReport   = errors.New("collector: reporte ya registrado previamente")
	ErrSignatureMismatch = errors.New("collector: firma criptográfica de reporte no verificada")
)

// DiagnosticReport estructura inmutable de reporte forense y rendimiento
type DiagnosticReport struct {
	ID          string             `json:"id"`           // Hash SHA-256 del contenido
	ReporterDID string             `json:"reporter_did"` // DID del nodo emisor
	Timestamp   int64              `json:"timestamp"`    // Epoch en segundos
	AnomalyType string             `json:"anomaly_type"` // ej. LATENCY_SPIKE, NAT_FAILURE, DROPS_BURST
	Severity    string             `json:"severity"`     // INFO, WARNING, CRITICAL
	Details     string             `json:"details"`      // Explicación técnica sanitizada
	Version     string             `json:"version"`      // Versión de IPVN7 ejecutada
	OS          string             `json:"os"`           // runtime.GOOS
	Arch        string             `json:"arch"`         // runtime.GOARCH
	Metrics     map[string]float64 `json:"metrics"`      // Snapshot numérico (rtt_ms, drops, pdr, etc)
	Signature   []byte             `json:"signature"`    // Firma Ed25519 del emisor
}

// DistributedCollector coordina la ingesta, desduplicación y persistencia de diagnósticos en DFS
type DistributedCollector struct {
	mu            sync.RWMutex
	store         *dfs.ChunkStore
	identity      *l0.Identity
	reports       []*DiagnosticReport
	dedupMap      map[string]bool
	maxMemoryLogs int
}

// NewDistributedCollector inicializa el motor colector soberano
func NewDistributedCollector(store *dfs.ChunkStore, identity *l0.Identity) *DistributedCollector {
	return &DistributedCollector{
		store:         store,
		identity:      identity,
		reports:       make([]*DiagnosticReport, 0, 100),
		dedupMap:      make(map[string]bool),
		maxMemoryLogs: 200,
	}
}

// IngestReport valida, desduplica y persiste un reporte entrante en el DFS local
func (dc *DistributedCollector) IngestReport(report *DiagnosticReport) (string, error) {
	if report == nil || report.ReporterDID == "" || report.AnomalyType == "" {
		return "", ErrInvalidReport
	}

	// 1. Calcular hash unívoco del reporte
	reportBytes, err := json.Marshal(struct {
		ReporterDID string             `json:"reporter_did"`
		Timestamp   int64              `json:"timestamp"`
		AnomalyType string             `json:"anomaly_type"`
		Severity    string             `json:"severity"`
		Details     string             `json:"details"`
		Version     string             `json:"version"`
		OS          string             `json:"os"`
		Arch        string             `json:"arch"`
		Metrics     map[string]float64 `json:"metrics"`
	}{
		ReporterDID: report.ReporterDID,
		Timestamp:   report.Timestamp,
		AnomalyType: report.AnomalyType,
		Severity:    report.Severity,
		Details:     report.Details,
		Version:     report.Version,
		OS:          report.OS,
		Arch:        report.Arch,
		Metrics:     report.Metrics,
	})
	if err != nil {
		return "", fmt.Errorf("error serializando reporte para hash: %w", err)
	}

	h := sha256.Sum256(reportBytes)
	reportHash := hex.EncodeToString(h[:])
	report.ID = reportHash

	dc.mu.Lock()
	defer dc.mu.Unlock()

	// 2. Control de deduplicación atómico
	if dc.dedupMap[reportHash] {
		return reportHash, ErrDuplicateReport
	}

	// 3. Persistir el reporte completo en el Almacén Distribuido DFS si está disponible
	if dc.store != nil {
		fullPayload, _ := json.Marshal(report)
		_, _, _ = dc.store.StoreFile(fullPayload, fmt.Sprintf("diag_%s.json", reportHash[:12]), dc.identity)
	}

	// 4. Indexar en cola circular de memoria
	if len(dc.reports) >= dc.maxMemoryLogs {
		oldest := dc.reports[0]
		delete(dc.dedupMap, oldest.ID)
		dc.reports = dc.reports[1:]
	}

	dc.reports = append(dc.reports, report)
	dc.dedupMap[reportHash] = true

	return reportHash, nil
}

// CreateLocalReport empaqueta y firma un reporte forense generado por el nodo actual
func (dc *DistributedCollector) CreateLocalReport(anomalyType, severity, details string, metrics map[string]float64) (*DiagnosticReport, error) {
	if dc.identity == nil {
		return nil, errors.New("collector: identidad local no disponible para firmar reporte")
	}

	report := &DiagnosticReport{
		ReporterDID: dc.identity.DID(),
		Timestamp:   time.Now().Unix(),
		AnomalyType: anomalyType,
		Severity:    severity,
		Details:     details,
		Version:     "0.7.0",
		OS:          runtime.GOOS,
		Arch:        runtime.GOARCH,
		Metrics:     metrics,
	}

	// Generar firma criptográfica Ed25519
	signBytes, _ := json.Marshal(struct {
		ReporterDID string             `json:"reporter_did"`
		Timestamp   int64              `json:"timestamp"`
		AnomalyType string             `json:"anomaly_type"`
		Severity    string             `json:"severity"`
		Details     string             `json:"details"`
		Version     string             `json:"version"`
		OS          string             `json:"os"`
		Arch        string             `json:"arch"`
		Metrics     map[string]float64 `json:"metrics"`
	}{
		ReporterDID: report.ReporterDID,
		Timestamp:   report.Timestamp,
		AnomalyType: report.AnomalyType,
		Severity:    report.Severity,
		Details:     report.Details,
		Version:     report.Version,
		OS:          report.OS,
		Arch:        report.Arch,
		Metrics:     report.Metrics,
	})

	report.Signature = dc.identity.Sign(signBytes)
	h := sha256.Sum256(signBytes)
	report.ID = hex.EncodeToString(h[:])

	// Auto-ingestión en el colector local
	_, _ = dc.IngestReport(report)

	return report, nil
}

// GetRecentReports retorna los reportes más recientes registrados en memoria
func (dc *DistributedCollector) GetRecentReports(limit int) []*DiagnosticReport {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	total := len(dc.reports)
	if total == 0 {
		return []*DiagnosticReport{}
	}

	if limit <= 0 || limit > total {
		limit = total
	}

	start := total - limit
	out := make([]*DiagnosticReport, limit)
	copy(out, dc.reports[start:])

	// Invertir para mostrar más recientes primero
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}

	return out
}

// TotalReports retorna la cantidad acumulada de reportes únicos
func (dc *DistributedCollector) TotalReports() int {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	return len(dc.reports)
}
