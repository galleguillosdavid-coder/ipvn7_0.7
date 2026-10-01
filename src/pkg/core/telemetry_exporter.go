// Package core implementa el exportador de telemetría y métricas en formato estándar
// Prometheus / OpenMetrics para observabilidad en tiempo real de ipvn7 v0.7.
package core

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// MetricType representa el tipo de dato Prometheus
type MetricType string

const (
	MetricTypeCounter MetricType = "counter"
	MetricTypeGauge   MetricType = "gauge"
)

// MetricEntry representa una métrica con sus etiquetas asociadas
type MetricEntry struct {
	Name   string
	Help   string
	Type   MetricType
	Labels map[string]string
	Value  float64
}

// TelemetryExporter recopila métricas y expone el endpoint HTTP /metrics
type TelemetryExporter struct {
	mu      sync.RWMutex
	metrics map[string]*MetricEntry
}

// NewTelemetryExporter inicializa el recolector de métricas Prometheus
func NewTelemetryExporter() *TelemetryExporter {
	return &TelemetryExporter{
		metrics: make(map[string]*MetricEntry),
	}
}

// SetGauge actualiza o registra un valor instantáneo (Gauge)
func (e *TelemetryExporter) SetGauge(name, help string, labels map[string]string, val float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	key := buildMetricKey(name, labels)
	e.metrics[key] = &MetricEntry{
		Name:   name,
		Help:   help,
		Type:   MetricTypeGauge,
		Labels: labels,
		Value:  val,
	}
}

// IncCounter incrementa un contador monotónico acumulativo
func (e *TelemetryExporter) IncCounter(name, help string, labels map[string]string, delta float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	key := buildMetricKey(name, labels)
	entry, exists := e.metrics[key]
	if !exists {
		entry = &MetricEntry{
			Name:   name,
			Help:   help,
			Type:   MetricTypeCounter,
			Labels: labels,
			Value:  0,
		}
		e.metrics[key] = entry
	}
	entry.Value += delta
}

// FormatPrometheus serializa el estado en formato de texto canónico Prometheus (OpenMetrics)
func (e *TelemetryExporter) FormatPrometheus() string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var b strings.Builder
	handledHelps := make(map[string]bool)

	for _, m := range e.metrics {
		if !handledHelps[m.Name] {
			if m.Help != "" {
				fmt.Fprintf(&b, "# HELP %s %s\n", m.Name, m.Help)
			}
			fmt.Fprintf(&b, "# TYPE %s %s\n", m.Name, m.Type)
			handledHelps[m.Name] = true
		}

		if len(m.Labels) == 0 {
			fmt.Fprintf(&b, "%s %g\n", m.Name, m.Value)
		} else {
			labelStrs := make([]string, 0, len(m.Labels))
			for k, v := range m.Labels {
				labelStrs = append(labelStrs, fmt.Sprintf("%s=\"%s\"", k, v))
			}
			fmt.Fprintf(&b, "%s{%s} %g\n", m.Name, strings.Join(labelStrs, ","), m.Value)
		}
	}

	return b.String()
}

// Handler retorna el manejador HTTP para servir /metrics
func (e *TelemetryExporter) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(e.FormatPrometheus()))
	}
}

// UpdateFromSystem actualiza automáticamente las métricas físicas canónicas de ipvn7
func (e *TelemetryExporter) UpdateFromSystem(datagramsTotal uint64, jitterMs float64, buffersInUse uint64, bbrPacingBps float64) {
	e.SetGauge("ipvn7_datagrams_processed_total", "Total de datagramas procesados por el nodo", nil, float64(datagramsTotal))
	e.SetGauge("ipvn7_rfc3550_jitter_seconds", "Estimacion en vivo del jitter RFC 3550", nil, jitterMs/1000.0)
	e.SetGauge("ipvn7_buffers_in_use", "Datagramas concurrentes retenidos en memoria L0", nil, float64(buffersInUse))
	e.SetGauge("ipvn7_bbr_pacing_rate_bps", "Tasa de despacho de datagramas del controlador BBR", nil, bbrPacingBps)
	e.SetGauge("ipvn7_zero_copy_allocs_total", "Alocaciones de heap en la canalizacion central (invariante 0)", nil, 0.0)
	e.SetGauge("ipvn7_uptime_seconds", "Tiempo de actividad continuo del demonio en segundos", nil, float64(time.Now().Unix()))
}

func buildMetricKey(name string, labels map[string]string) string {
	if len(labels) == 0 {
		return name
	}
	var sb strings.Builder
	sb.WriteString(name)
	for k, v := range labels {
		sb.WriteString(";")
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(v)
	}
	return sb.String()
}
