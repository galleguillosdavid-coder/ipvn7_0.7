package core

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTelemetryExporter_FormattingAndHTTPHandler(t *testing.T) {
	exp := NewTelemetryExporter()

	// 1. Registrar métricas arbitrarias
	exp.IncCounter("ipvn7_requests_total", "Total de peticiones recibidas", map[string]string{"proto": "udp"}, 10)
	exp.IncCounter("ipvn7_requests_total", "Total de peticiones recibidas", map[string]string{"proto": "udp"}, 5)
	exp.SetGauge("ipvn7_active_peers", "Pares activos conectados", map[string]string{"tier": "pqc_l3"}, 2)

	// 2. Actualizar métricas del sistema
	exp.UpdateFromSystem(50000, 0.05, 12, 12500000)

	// 3. Obtener formato Prometheus y verificar cabeceras y valores
	promText := exp.FormatPrometheus()

	if !strings.Contains(promText, "# HELP ipvn7_requests_total") {
		t.Fatal("expected HELP for ipvn7_requests_total")
	}
	if !strings.Contains(promText, "# TYPE ipvn7_requests_total counter") {
		t.Fatal("expected TYPE counter for ipvn7_requests_total")
	}
	if !strings.Contains(promText, "ipvn7_requests_total{proto=\"udp\"} 15") {
		t.Fatalf("expected counter value 15, got output: %s", promText)
	}
	if !strings.Contains(promText, "ipvn7_active_peers{tier=\"pqc_l3\"} 2") {
		t.Fatalf("expected gauge value 2, got output: %s", promText)
	}
	if !strings.Contains(promText, "ipvn7_zero_copy_allocs_total 0") {
		t.Fatalf("expected zero copy invariant 0 in text, got: %s", promText)
	}

	// 4. Probar HTTP handler
	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()

	handler := exp.Handler()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got: %d", rec.Code)
	}
	contentType := rec.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/plain") {
		t.Fatalf("expected text/plain content type, got: %s", contentType)
	}
	if !strings.Contains(rec.Body.String(), "ipvn7_datagrams_processed_total 50000") {
		t.Fatalf("HTTP body missing metric: %s", rec.Body.String())
	}
}
