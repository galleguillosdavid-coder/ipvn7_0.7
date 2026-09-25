package l1

import (
	"math/rand"
	"net"
	"sync"
	"testing"
	"time"

	"ipvn7/pkg/l0"
)

// TestKleinbergRouter_SaturationAndEviction evalúa el comportamiento bajo saturación masiva de anillos
func TestKleinbergRouter_SaturationAndEviction(t *testing.T) {
	localID, err := l0.GenerateIdentity()
	if err != nil {
		t.Fatalf("error generando id: %v", err)
	}

	cfg := RouterConfig{
		Profile:            ProfileCustom,
		NumRings:           8,
		PeersPerRing:       4,
		MaxTotalPeers:      32,
		AlphaLatencyWeight: 0.3,
		AutoRebalance:      false,
	}

	router := NewKleinbergRouterWithConfig(localID, cfg)

	// 1. Prohibición estricta de auto-emparejamiento
	if err := router.AddOrUpdatePeer(localID.DID(), nil, 1.0); err != nil {
		t.Errorf("error inesperado en auto-emparejamiento ignorado: %v", err)
	}
	if len(router.GetAllPeers()) != 0 {
		t.Fatalf("violación: router aceptó auto-emparejamiento de su propio DID")
	}

	// 2. Generar 100 identidades válidas e insertarlas
	peers := make([]*l0.Identity, 100)
	for i := 0; i < 100; i++ {
		id, err := l0.GenerateIdentity()
		if err != nil {
			t.Fatalf("error generando identidad %d: %v", i, err)
		}
		peers[i] = id
		latency := float64(50 + (i % 200))
		addr := &net.UDPAddr{IP: net.ParseIP("192.168.1.100"), Port: 7000 + i}
		_ = router.AddOrUpdatePeer(id.DID(), addr, latency)
	}

	allPeers := router.GetAllPeers()
	if len(allPeers) > cfg.MaxTotalPeers {
		t.Errorf("capacidad total excedida: %d > %d", len(allPeers), cfg.MaxTotalPeers)
	}

	// 3. Verificar que ningún anillo exceda su cuota acotada
	distribution := router.GetRingDistribution()
	for ringIdx, count := range distribution {
		if count > cfg.PeersPerRing {
			t.Errorf("anillo %d sobrecargado: %d > %d", ringIdx, count, cfg.PeersPerRing)
		}
	}
}

// TestKleinbergRouter_ConcurrentAccessStress somete al enrutador a contienda concurrente
func TestKleinbergRouter_ConcurrentAccessStress(t *testing.T) {
	localID, _ := l0.GenerateIdentity()
	router := NewKleinbergRouter(localID)

	// Pre-poblar con 30 pares
	targetPeers := make([]string, 30)
	for i := 0; i < 30; i++ {
		id, _ := l0.GenerateIdentity()
		targetPeers[i] = id.DID()
		_ = router.AddOrUpdatePeer(id.DID(), nil, float64(10+i))
	}

	var wg sync.WaitGroup
	workers := 40
	iterations := 100

	// Goroutines lectoras (FindNextHop masivo)
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerID)))
			for it := 0; it < iterations; it++ {
				target := targetPeers[rng.Intn(len(targetPeers))]
				nextHop, err := router.FindNextHop(target)
				if err == nil && nextHop == nil {
					t.Errorf("nextHop fue nil sin error devuelto")
				}
			}
		}(w)
	}

	// Goroutines escritoras (actualizaciones y nuevos pares simultáneos)
	for w := 0; w < 10; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for it := 0; it < 20; it++ {
				id, _ := l0.GenerateIdentity()
				_ = router.AddOrUpdatePeer(id.DID(), nil, float64(5+it))
				time.Sleep(100 * time.Microsecond)
			}
		}(w)
	}

	wg.Wait()

	// Confirmar que el enrutador conserva estado coherente tras la contienda
	finalPeers := router.GetAllPeers()
	if len(finalPeers) == 0 {
		t.Fatal("estado corrupto: 0 pares tras contienda")
	}
}
