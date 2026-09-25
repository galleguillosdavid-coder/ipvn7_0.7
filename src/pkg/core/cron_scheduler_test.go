package core

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestCronScheduler_ParseInterval(t *testing.T) {
	d1, err := ParseCronInterval("@every 500ms")
	if err != nil || d1 != 500*time.Millisecond {
		t.Fatalf("error parseando @every: %v", err)
	}

	d2, err := ParseCronInterval("*/5 * * * *")
	if err != nil || d2 != 5*time.Minute {
		t.Fatalf("error parseando */5: %v", err)
	}

	d3, err := ParseCronInterval("@daily")
	if err != nil || d3 != 24*time.Hour {
		t.Fatalf("error parseando @daily: %v", err)
	}
}

func TestCronScheduler_ExecutionLoop(t *testing.T) {
	var count int64
	handler := func(jobID, intent string) error {
		atomic.AddInt64(&count, 1)
		return nil
	}

	sched := NewSovereignCronScheduler(handler)
	job, err := sched.AddJob("test_pulse", "@every 100ms", "telemetry:drone", "did:ipvn7:local")
	if err != nil {
		t.Fatalf("error añadiendo tarea cron: %v", err)
	}
	if job.ID != "test_pulse" {
		t.Fatalf("ID esperado test_pulse")
	}

	sched.Start()
	// Esperar que se ejecute al menos 2 veces
	time.Sleep(350 * time.Millisecond)
	sched.Stop()

	execs := atomic.LoadInt64(&count)
	if execs < 2 {
		t.Fatalf("se esperaban al menos 2 ejecuciones, obtenidas: %d", execs)
	}

	jobs := sched.ListJobs()
	if len(jobs) != 1 || jobs[0].RunCount < 2 {
		t.Fatalf("estado de tarea no actualizado: %+v", jobs)
	}
}
