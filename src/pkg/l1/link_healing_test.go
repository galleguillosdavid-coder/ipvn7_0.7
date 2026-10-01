package l1_test

import (
	"net"
	"testing"

	"ipvn7/pkg/l0"
	"ipvn7/pkg/l1"
)

func TestLinkHealingEngine(t *testing.T) {
	id, _ := l0.GenerateIdentity()
	router := l1.NewKleinbergRouter(id)

	peer1, _ := l0.GenerateIdentity()
	peer2, _ := l0.GenerateIdentity()

	_ = router.AddOrUpdatePeer(peer1.DID(), &net.UDPAddr{IP: net.ParseIP("198.51.100.1"), Port: 7777}, 10.0)
	_ = router.AddOrUpdatePeer(peer2.DID(), &net.UDPAddr{IP: net.ParseIP("198.51.100.2"), Port: 7777}, 15.0)

	engine := l1.NewLinkHealingEngine(router)

	// 1. Enlace saludable
	healed, alt := engine.RecordProbeResult(peer1.DID(), true, 12.0)
	if healed {
		t.Fatalf("un enlace saludable no debió disparar failover")
	}
	if alt != "" {
		t.Fatalf("par alternativo debió ser vacío")
	}

	// 2. Pérdidas consecutivas menores al umbral
	healed, _ = engine.RecordProbeResult(peer1.DID(), false, 0)
	if healed {
		t.Fatalf("1 pérdida no debió disparar failover")
	}
	healed, _ = engine.RecordProbeResult(peer1.DID(), false, 0)
	if healed {
		t.Fatalf("2 pérdidas no debieron disparar failover")
	}

	// 3. 3ra pérdida consecutiva debe disparar failover a peer2
	healed, alt = engine.RecordProbeResult(peer1.DID(), false, 0)
	if !healed {
		t.Fatalf("3 pérdidas consecutivas debieron disparar failover")
	}
	if alt != peer2.DID() {
		t.Fatalf("esperaba failover hacia peer2 (%s), pero obtuve: %s", peer2.DID(), alt)
	}

	// 4. Verificar estadísticas
	stats, found := engine.GetPeerStats(peer1.DID())
	if !found || !stats.Degraded || stats.FailoverCount != 1 {
		t.Fatalf("estadísticas inconsistentes: %+v", stats)
	}

	// 5. Recuperación con histeresis anti-flapping (requiere 2 éxitos consecutivos)
	healed, _ = engine.RecordProbeResult(peer1.DID(), true, 20.0)
	if healed {
		t.Fatalf("recuperación no debió disparar failover")
	}
	// Con 1 éxito aún debe estar degradado (histeresis activa)
	stats, _ = engine.GetPeerStats(peer1.DID())
	if !stats.Degraded {
		t.Fatalf("1 solo éxito no debe levantar la degradación por histeresis anti-flapping")
	}

	// 2do éxito consecutivo: ahora sí debe restaurar el enlace
	healed, _ = engine.RecordProbeResult(peer1.DID(), true, 20.0)
	stats, _ = engine.GetPeerStats(peer1.DID())
	if stats.Degraded {
		t.Fatalf("2 éxitos consecutivos debieron restaurar el enlace")
	}
}

func TestLinkHealingEngine_CatastrophicFastFailover(t *testing.T) {
	id, _ := l0.GenerateIdentity()
	router := l1.NewKleinbergRouter(id)

	peer1, _ := l0.GenerateIdentity()
	peer2, _ := l0.GenerateIdentity()

	_ = router.AddOrUpdatePeer(peer1.DID(), &net.UDPAddr{IP: net.ParseIP("198.51.100.1"), Port: 7777}, 10.0)
	_ = router.AddOrUpdatePeer(peer2.DID(), &net.UDPAddr{IP: net.ParseIP("198.51.100.2"), Port: 7777}, 15.0)

	engine := l1.NewLinkHealingEngine(router)

	// Inyectar un colapso catastrófico de latencia (1800 ms > 1500 ms)
	// Debe conmutar de inmediato (en la primera sonda) sin esperar 3 pérdidas consecutivas
	healed, alt := engine.RecordProbeResult(peer1.DID(), true, 1800.0)
	if !healed {
		t.Fatalf("Degradación catastrófica (>1500ms) debió activar Fast-Failover inmediato")
	}
	if alt != peer2.DID() {
		t.Fatalf("Esperaba failover hacia peer2 (%s), pero obtuve: %s", peer2.DID(), alt)
	}

	stats, found := engine.GetPeerStats(peer1.DID())
	if !found || !stats.Degraded {
		t.Fatalf("El enlace debió quedar marcado como degradado inmediatamente")
	}
}

func TestLinkHealingEngine_HoldDownDampeningAntiFlapLoop(t *testing.T) {
	id, _ := l0.GenerateIdentity()
	router := l1.NewKleinbergRouter(id)

	peer1, _ := l0.GenerateIdentity()
	peer2, _ := l0.GenerateIdentity()

	_ = router.AddOrUpdatePeer(peer1.DID(), &net.UDPAddr{IP: net.ParseIP("198.51.100.1"), Port: 7777}, 10.0)
	_ = router.AddOrUpdatePeer(peer2.DID(), &net.UDPAddr{IP: net.ParseIP("198.51.100.2"), Port: 7777}, 15.0)

	engine := l1.NewLinkHealingEngine(router)

	// 1. Primer failover por micro-corte
	engine.RecordProbeResult(peer1.DID(), false, 0)
	engine.RecordProbeResult(peer1.DID(), false, 0)
	engine.RecordProbeResult(peer1.DID(), false, 0)

	// 2. Recuperación temporal
	engine.RecordProbeResult(peer1.DID(), true, 10.0)
	engine.RecordProbeResult(peer1.DID(), true, 10.0)

	// 3. Segundo failover inmediato (< 10s): comportamiento de FLAP repetido (RF micro-jamming)
	engine.RecordProbeResult(peer1.DID(), false, 0)
	engine.RecordProbeResult(peer1.DID(), false, 0)
	healed, _ := engine.RecordProbeResult(peer1.DID(), false, 0)
	if !healed {
		t.Fatalf("Segundo failover debió conmutar ruta")
	}

	stats, _ := engine.GetPeerStats(peer1.DID())
	if stats.HoldDownUntil.IsZero() {
		t.Fatalf("HoldDownUntil debió configurarse tras detectar flapping repetido")
	}
	if stats.FlapPenaltyCount < 1 {
		t.Fatalf("FlapPenaltyCount debió ser >= 1")
	}

	// 4. Simular que el enlace intenta volver de inmediato tras 2 éxitos:
	// La amortiguación Hold-Down debe mantenerlo degradado e impedir el bucle pendular
	engine.RecordProbeResult(peer1.DID(), true, 10.0)
	engine.RecordProbeResult(peer1.DID(), true, 10.0)

	statsAfterRecover, _ := engine.GetPeerStats(peer1.DID())
	if !statsAfterRecover.Degraded {
		t.Fatalf("El enlace debió mantenerse degradado durante la ventana de enfriamiento Hold-Down para evitar oscilaciones")
	}
}

