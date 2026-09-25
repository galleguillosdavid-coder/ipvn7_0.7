package core

import (
	"os"
	"testing"

	"ipvn7/pkg/dfs"
	"ipvn7/pkg/l0"
)

func TestDistributedCollector_IngestAndDeduplicate(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ipvn7_collector_test_*")
	if err != nil {
		t.Fatalf("Error creando directorio temporal: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cs, err := dfs.NewChunkStore(tempDir)
	if err != nil {
		t.Fatalf("Error inicializando ChunkStore: %v", err)
	}

	id, err := l0.GenerateIdentity()
	if err != nil {
		t.Fatalf("Error creando identidad: %v", err)
	}

	collector := NewDistributedCollector(cs, id)

	report, err := collector.CreateLocalReport("LATENCY_SPIKE", "WARNING", "Pico de 120ms observado en enlace UDP", map[string]float64{
		"rtt_ms": 120.5,
		"drops":  2,
	})
	if err != nil {
		t.Fatalf("CreateLocalReport falló: %v", err)
	}

	if report.ID == "" || report.ReporterDID != id.DID() {
		t.Fatalf("Reporte mal formado: %+v", report)
	}

	// Segundo intento con el mismo reporte debe arrojar ErrDuplicateReport
	_, err = collector.IngestReport(report)
	if err != ErrDuplicateReport {
		t.Fatalf("Se esperaba ErrDuplicateReport, obtenido: %v", err)
	}

	// Verificar recuperación en GetRecentReports
	recent := collector.GetRecentReports(10)
	if len(recent) != 1 {
		t.Fatalf("Se esperaba 1 reporte reciente, obtenidos %d", len(recent))
	}
	if recent[0].AnomalyType != "LATENCY_SPIKE" {
		t.Fatalf("Tipo de anomalía esperado LATENCY_SPIKE, obtenido: %s", recent[0].AnomalyType)
	}
}
