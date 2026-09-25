package l1

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestMemoryArbiter_DoSStress100kPackets somete al sistema a una ráfaga masiva
// de 100,000 datagramas concurrentes simulando un ataque de denegación de servicio (DoS)
// por inundación de memoria, validando que el árbitro protege las cuotas estrictas de RAM,
// no sufre fugas de memoria y recupera su estado base tras el drenado.
func TestMemoryArbiter_DoSStress100kPackets(t *testing.T) {
	// Límite de 10 MB total para el nodo de prueba
	totalBytes := uint64(10 * 1024 * 1024)
	arbiter := NewGlobalMemoryArbiter(totalBytes)

	// Crear planificador WDRR asociado
	scheduler := NewWDRRScheduler(1500, 2048)

	const totalPackets = 100000
	const numWorkers = 10
	packetsPerWorker := totalPackets / numWorkers

	var acceptedPackets uint64
	var rejectedPackets uint64
	var droppedByScheduler uint64

	startMem := runtime.MemStats{}
	runtime.GC()
	runtime.ReadMemStats(&startMem)

	start := time.Now()

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for w := 0; w < numWorkers; w++ {
		go func(workerID int) {
			defer wg.Done()
			payloadSample := []byte("DATAGRAMA_1280B_TRAFICO_CUANTICO_RESISTENTE_A_INUNDACION_DOS")
			pktSize := uint64(len(payloadSample))

			for i := 0; i < packetsPerWorker; i++ {
				// Alternar clases de presupuesto
				class := BudgetQoS
				trafficClass := ClassInteractive
				if i%3 == 0 {
					class = BudgetReplay
					trafficClass = ClassControl
				} else if i%3 == 2 {
					class = BudgetBindings
					trafficClass = ClassBulk
				}

				// 1. Intentar reservar presupuesto en el árbitro
				if arbiter.Reserve(class, pktSize) {
					atomic.AddUint64(&acceptedPackets, 1)

					// 2. Si se reservó memoria, encolar en el planificador
					_, err := scheduler.Enqueue(trafficClass, payloadSample)
					if err != nil {
						atomic.AddUint64(&droppedByScheduler, 1)
					}

					// Simular consumo y liberación rápida del ciclo de procesamiento
					if i%5 == 0 {
						arbiter.Release(class, pktSize)
					}
				} else {
					atomic.AddUint64(&rejectedPackets, 1)
				}
			}
		}(w)
	}

	wg.Wait()
	duration := time.Since(start)

	stats := arbiter.GetStats()

	// Aserciones de seguridad anti-DoS
	if stats.TotalAllocated > totalBytes {
		t.Fatalf("VIOLACIÓN CRÍTICA: la memoria reservada (%d) superó el límite estricto (%d)",
			stats.TotalAllocated, totalBytes)
	}

	for class, quota := range stats.Classes {
		if quota.Allocated > quota.LimitBytes {
			t.Fatalf("VIOLACIÓN CRÍTICA DE CLASE %s: asignado (%d) > límite (%d)",
				class, quota.Allocated, quota.LimitBytes)
		}
	}

	pps := float64(totalPackets) / duration.Seconds()
	t.Logf("Rendimiento bajo ataque DoS: %d paquetes procesados en %v (%.2f paquetes/seg)",
		totalPackets, duration, pps)
	t.Logf("Aceptados: %d, Rechazados preventivamente por RAM: %d, Descartados por Scheduler: %d",
		acceptedPackets, rejectedPackets, droppedByScheduler)

	// Liberar remanentes para comprobar retorno a estado cero
	for class := range stats.Classes {
		arbiter.Release(class, stats.Classes[class].Allocated)
	}

	cleanStats := arbiter.GetStats()
	if cleanStats.TotalAllocated != 0 {
		t.Errorf("fuga detectada: memoria remanente tras drenaje = %d bytes", cleanStats.TotalAllocated)
	}
}

// BenchmarkMemoryArbiter_100kRate evalúa el rendimiento del árbitro en pps puros
func BenchmarkMemoryArbiter_100kRate(b *testing.B) {
	arb := NewGlobalMemoryArbiter(64 * 1024 * 1024)
	pktSize := uint64(1280)
	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if arb.Reserve(BudgetQoS, pktSize) {
				arb.Release(BudgetQoS, pktSize)
			}
		}
	})
}
