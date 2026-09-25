// Package core implementa el Programador de Tareas Autónomas y Cron de Malla Soberana
// para ipvn7 v0.7, permitiendo a los nodos ejecutar intenciones y monitoreo en bucle.
package core

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// CronJobHandler función ejecutada cuando se dispara una tarea cron
type CronJobHandler func(jobID string, intent string) error

// CronJob definición de una tarea programada autónoma
type CronJob struct {
	ID          string        `json:"id"`
	Expression  string        `json:"expression"`  // Ej: "*/5 * * * *", "@every 10s", "@every 1m"
	Intent      string        `json:"intent"`      // Intención u orden a ejecutar (ej. "hardware:battery", "telemetry:drone")
	TargetDID   string        `json:"target_did"`  // DID local o remoto
	Interval    time.Duration `json:"interval"`
	LastRun     time.Time     `json:"last_run"`
	NextRun     time.Time     `json:"next_run"`
	RunCount    uint64        `json:"run_count"`
	Enabled     bool          `json:"enabled"`
	LastError   string        `json:"last_error,omitempty"`
}

// SovereignCronScheduler orquesta la ejecución periódica de tareas en el nodo ipvn7
type SovereignCronScheduler struct {
	mu       sync.RWMutex
	jobs     map[string]*CronJob
	handler  CronJobHandler
	running  bool
	stopChan chan struct{}
}

// NewSovereignCronScheduler crea una nueva instancia del programador cron
func NewSovereignCronScheduler(h CronJobHandler) *SovereignCronScheduler {
	return &SovereignCronScheduler{
		jobs:     make(map[string]*CronJob),
		handler:  h,
		running:  false,
		stopChan: make(chan struct{}),
	}
}

// AddJob añade o actualiza una tarea programada
func (s *SovereignCronScheduler) AddJob(id, expr, intent, targetDID string) (*CronJob, error) {
	if id == "" {
		return nil, errors.New("cron: id de tarea no puede ser vacío")
	}
	interval, err := ParseCronInterval(expr)
	if err != nil {
		return nil, fmt.Errorf("cron: expresión inválida '%s': %w", expr, err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	job := &CronJob{
		ID:         id,
		Expression: expr,
		Intent:     intent,
		TargetDID:  targetDID,
		Interval:   interval,
		NextRun:    now.Add(interval),
		Enabled:    true,
	}
	s.jobs[id] = job
	return job, nil
}

// RemoveJob elimina una tarea programada por su ID
func (s *SovereignCronScheduler) RemoveJob(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.jobs[id]; ok {
		delete(s.jobs, id)
		return true
	}
	return false
}

// Start inicia el bucle de despacho cron en segundo plano
func (s *SovereignCronScheduler) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.stopChan = make(chan struct{})
	s.mu.Unlock()

	go s.runLoop()
}

// Stop detiene el bucle de despacho
func (s *SovereignCronScheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	close(s.stopChan)
	s.mu.Unlock()
}

// runLoop bucle ticker para evaluación de tareas activas
func (s *SovereignCronScheduler) runLoop() {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case now := <-ticker.C:
			s.evaluateDueJobs(now)
		}
	}
}

// evaluateDueJobs ejecuta las tareas cuya fecha de expiración ha llegado
func (s *SovereignCronScheduler) evaluateDueJobs(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, job := range s.jobs {
		if !job.Enabled {
			continue
		}
		if now.After(job.NextRun) || now.Equal(job.NextRun) {
			job.LastRun = now
			job.NextRun = now.Add(job.Interval)
			job.RunCount++

			// Ejecutar en goroutine no bloqueante
			go s.dispatchJob(job.ID, job.Intent)
		}
	}
}

// dispatchJob invoca el manejador asignado
func (s *SovereignCronScheduler) dispatchJob(id, intent string) {
	if s.handler == nil {
		return
	}
	err := s.handler(id, intent)

	s.mu.Lock()
	defer s.mu.Unlock()
	if job, ok := s.jobs[id]; ok {
		if err != nil {
			job.LastError = err.Error()
		} else {
			job.LastError = ""
		}
	}
}

// ListJobs retorna una copia de todas las tareas activas
func (s *SovereignCronScheduler) ListJobs() []*CronJob {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := make([]*CronJob, 0, len(s.jobs))
	for _, j := range s.jobs {
		cp := *j
		res = append(res, &cp)
	}
	return res
}

// ParseCronInterval interpreta macros @every o sintaxis cron básica de minutos
func ParseCronInterval(expr string) (time.Duration, error) {
	trimmed := strings.TrimSpace(expr)
	if strings.HasPrefix(trimmed, "@every ") {
		durStr := strings.TrimPrefix(trimmed, "@every ")
		return time.ParseDuration(durStr)
	}

	// Soporte para patrón estándar: "*/N * * * *" (cada N minutos)
	parts := strings.Fields(trimmed)
	if len(parts) == 5 && strings.HasPrefix(parts[0], "*/") {
		numStr := strings.TrimPrefix(parts[0], "*/")
		mins, err := strconv.Atoi(numStr)
		if err == nil && mins > 0 {
			return time.Duration(mins) * time.Minute, nil
		}
	}

	// Por defecto para "0 * * * *" (cada hora)
	if trimmed == "0 * * * *" || trimmed == "@hourly" {
		return time.Hour, nil
	}
	if trimmed == "@daily" {
		return 24 * time.Hour, nil
	}

	return 0, fmt.Errorf("formato no soportado (use '@every 10s', '@every 5m', '*/10 * * * *')")
}
