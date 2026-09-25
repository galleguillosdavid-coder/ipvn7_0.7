package l0

import (
	"sync"
	"testing"
)

func TestBufferHealthMonitor_TrackingAndAlerts(t *testing.T) {
	budget := uint64(10)
	monitor := NewBufferHealthMonitor(budget)

	snap := monitor.Snapshot()
	if snap.BuffersAllocated != 0 || snap.BuffersInUse != 0 || snap.LeakAlert {
		t.Fatal("initial snapshot should have 0 alloc/in-use and no leak alert")
	}

	// 1. Asignar 5 buffers
	for i := 0; i < 5; i++ {
		monitor.TrackAlloc()
	}

	snap = monitor.Snapshot()
	if snap.BuffersAllocated != 5 || snap.BuffersInUse != 5 || snap.HighWaterMark != 5 {
		t.Fatalf("expected inUse=5, hwm=5, got: inUse=%d, hwm=%d", snap.BuffersInUse, snap.HighWaterMark)
	}

	// 2. Reciclar 2 buffers
	monitor.TrackRecycle()
	monitor.TrackRecycle()

	snap = monitor.Snapshot()
	if snap.BuffersRecycled != 2 || snap.BuffersInUse != 3 || snap.HighWaterMark != 5 {
		t.Fatalf("expected inUse=3, hwm=5, got: inUse=%d, hwm=%d", snap.BuffersInUse, snap.HighWaterMark)
	}

	// 3. Superar el presupuesto para disparar la alarma de fuga
	for i := 0; i < 10; i++ {
		monitor.TrackAlloc()
	}

	snap = monitor.Snapshot()
	if !snap.LeakAlert {
		t.Fatalf("expected LeakAlert=true when inUse (%d) > budget (%d)", snap.BuffersInUse, budget)
	}

	// 4. Concurrencia masiva: 100 goroutines asignando y reciclando
	var wg sync.WaitGroup
	for g := 0; g < 100; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for k := 0; k < 100; k++ {
				monitor.TrackAlloc()
				monitor.TrackRecycle()
			}
		}()
	}
	wg.Wait()

	// 5. Reset
	monitor.Reset()
	snap = monitor.Snapshot()
	if snap.BuffersAllocated != 0 || snap.BuffersRecycled != 0 || snap.HighWaterMark != 0 {
		t.Fatal("snapshot after reset should be zeroed")
	}
}
