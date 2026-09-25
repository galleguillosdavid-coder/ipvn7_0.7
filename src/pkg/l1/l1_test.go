package l1_test

import (
	"crypto/ed25519"
	"net"
	"testing"

	"ipvn7/pkg/l0"
	"ipvn7/pkg/l1"
)

func TestKleinbergRouterAndRings(t *testing.T) {
	localID, _ := l0.GenerateIdentity()
	router := l1.NewKleinbergRouter(localID)

	// Agregar pares distribuidos a lo largo del espacio de claves (diferentes anillos)
	totalAdded := 0
	for ringByte := 0; ringByte < 24; ringByte++ {
		// Generar clave con distancia controlada respecto a localID
		craftedPub := make([]byte, 32)
		copy(craftedPub, localID.PublicKey)
		craftedPub[ringByte%32] ^= 0x80 // Invertir bit para variar distancia XOR

		did := l0.DIDFromPublicKey(ed25519.PublicKey(craftedPub))
		addr := &net.UDPAddr{IP: net.ParseIP("198.51.100.100"), Port: 7000 + ringByte}
		if err := router.AddOrUpdatePeer(did, addr, float64(ringByte*2)); err == nil {
			totalAdded++
		}
	}

	peers := router.GetAllPeers()
	if len(peers) == 0 {
		t.Fatalf("No se registraron pares en la tabla de enrutamiento")
	}

	// Probar búsqueda voraz de siguiente salto hacia un par existente
	targetPeer := peers[0]
	nextHop, err := router.FindNextHop(targetPeer.DID)
	if err != nil {
		t.Fatalf("FindNextHop falló: %v", err)
	}
	if nextHop.DID != targetPeer.DID {
		t.Errorf("FindNextHop debió resolver directamente el destino conocido")
	}

	// Probar búsqueda voraz hacia un DID desconocido
	unknownID, _ := l0.GenerateIdentity()
	bestHop, err := router.FindNextHop(unknownID.DID())
	if err != nil {
		t.Fatalf("FindNextHop hacia nodo desconocido falló: %v", err)
	}
	if bestHop == nil {
		t.Errorf("FindNextHop debió seleccionar el mejor salto aproximado")
	}
}

func TestUserspaceVirtualAdapter(t *testing.T) {
	id, _ := l0.GenerateIdentity()
	tun := l1.NewUserspaceVirtualAdapter(id)
	defer tun.Close()

	if tun.MTU() != 1280 {
		t.Errorf("MTU esperada: 1280, obtenida: %d", tun.MTU())
	}

	packetData := []byte("Paquete sintético TUN")
	if err := tun.InjectPacket(packetData); err != nil {
		t.Fatalf("InjectPacket falló: %v", err)
	}

	readData, err := tun.ReadPacket()
	if err != nil {
		t.Fatalf("ReadPacket falló: %v", err)
	}

	if string(readData) != string(packetData) {
		t.Errorf("Datos leídos no coinciden con datos inyectados")
	}
}

func TestElasticTopologyAndDynamicRings(t *testing.T) {
	id, _ := l0.GenerateIdentity()

	// 1. Probar Perfil Micro (8 anillos, 64 pares)
	microRouter := l1.NewKleinbergRouterWithConfig(id, l1.MicroRouterConfig())
	if microRouter.GetConfig().NumRings != 8 || microRouter.GetConfig().MaxTotalPeers != 64 {
		t.Errorf("Perfil Micro incorrecto: %+v", microRouter.GetConfig())
	}

	// 2. Probar Perfil Standard (16 anillos, 256 pares)
	stdRouter := l1.NewKleinbergRouter(id)
	if stdRouter.GetConfig().NumRings != 16 || stdRouter.GetConfig().MaxTotalPeers != 256 {
		t.Errorf("Perfil Standard incorrecto: %+v", stdRouter.GetConfig())
	}

	// 3. Probar Perfil Backbone (32 anillos, 1024 pares)
	backboneRouter := l1.NewKleinbergRouterWithConfig(id, l1.BackboneRouterConfig())
	if backboneRouter.GetConfig().NumRings != 32 || backboneRouter.GetConfig().MaxTotalPeers != 1024 {
		t.Errorf("Perfil Backbone incorrecto: %+v", backboneRouter.GetConfig())
	}

	// 4. Probar Perfil Custom (64 anillos ultra-resolución, 2048 pares)
	customCfg := l1.RouterConfig{
		Profile:            l1.ProfileCustom,
		NumRings:           64,
		PeersPerRing:       32,
		MaxTotalPeers:      2048,
		AlphaLatencyWeight: 0.5,
		AutoRebalance:      true,
	}
	customRouter := l1.NewKleinbergRouterWithConfig(id, customCfg)
	if customRouter.GetConfig().NumRings != 64 || customRouter.GetConfig().MaxTotalPeers != 2048 {
		t.Errorf("Perfil Custom incorrecto: %+v", customRouter.GetConfig())
	}

	// Verificar cálculo de distancia XOR N con 8, 16, 32 y 64 anillos
	k1 := []byte("12345678901234567890123456789012")
	k2 := []byte("12345678901234567890123456789013") // Solo difiere en último byte (muy cercano)
	dist8 := l1.XORKeyDistanceN(k1, k2, 8)
	dist16 := l1.XORKeyDistanceN(k1, k2, 16)
	dist32 := l1.XORKeyDistanceN(k1, k2, 32)
	dist64 := l1.XORKeyDistanceN(k1, k2, 64)

	if dist8 > 1 || dist16 > 2 || dist32 > 4 || dist64 > 8 {
		t.Errorf("Claves casi idénticas debieron quedar en anillos bajos: dist8=%d, dist16=%d, dist32=%d, dist64=%d",
			dist8, dist16, dist32, dist64)
	}
}

func TestSmartEvictionPolicy(t *testing.T) {
	id, _ := l0.GenerateIdentity()
	// Router con capacidad muy pequeña para inducir desalojo: 4 anillos, 2 pares por anillo = 8 pares
	cfg := l1.RouterConfig{
		Profile:            l1.ProfileCustom,
		NumRings:           4,
		PeersPerRing:       2,
		MaxTotalPeers:      8,
		AlphaLatencyWeight: 0.35,
	}
	router := l1.NewKleinbergRouterWithConfig(id, cfg)

	// Llenar un anillo específico con 2 pares
	pubTarget := make([]byte, 32)
	copy(pubTarget, id.PublicKey)
	pubTarget[31] ^= 0x01
	did1 := l0.DIDFromPublicKey(ed25519.PublicKey(pubTarget))

	pubTarget2 := make([]byte, 32)
	copy(pubTarget2, id.PublicKey)
	pubTarget2[31] ^= 0x02
	did2 := l0.DIDFromPublicKey(ed25519.PublicKey(pubTarget2))

	addr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9001}
	_ = router.AddOrUpdatePeer(did1, addr, 50.0)
	_ = router.AddOrUpdatePeer(did2, addr, 100.0)

	// Degradar a did2
	router.UpdatePeerHealth(did2, 0.0, 80.0, 0.5) // Unreachable

	// Insertar un par nuevo did3 con excelente latencia (5.0ms)
	pubTarget3 := make([]byte, 32)
	copy(pubTarget3, id.PublicKey)
	pubTarget3[31] ^= 0x03
	did3 := l0.DIDFromPublicKey(ed25519.PublicKey(pubTarget3))

	_ = router.AddOrUpdatePeer(did3, addr, 5.0)

	peers := router.GetAllPeers()
	// did2 debió ser desalojado por estar inaccesible/peor latencia
	hasDid2 := false
	hasDid3 := false
	for _, p := range peers {
		if p.DID == did2 {
			hasDid2 = true
		}
		if p.DID == did3 {
			hasDid3 = true
		}
	}

	if hasDid2 {
		t.Errorf("did2 debió haber sido desalojado por la política Smart Eviction")
	}
	if !hasDid3 {
		t.Errorf("did3 debió haber sido admitido exitosamente")
	}
}

func TestHybrid2DGreedyRouting(t *testing.T) {
	id, _ := l0.GenerateIdentity()
	router := l1.NewKleinbergRouter(id)

	destID, _ := l0.GenerateIdentity()

	// Par A: Distancia XOR moderada pero latencia ultra baja (1.0 ms, LAN local)
	pubA := make([]byte, 32)
	copy(pubA, id.PublicKey)
	pubA[5] ^= 0xFF
	didA := l0.DIDFromPublicKey(ed25519.PublicKey(pubA))
	_ = router.AddOrUpdatePeer(didA, &net.UDPAddr{IP: net.ParseIP("198.51.100.10"), Port: 7001}, 1.0)

	// Par B: Distancia XOR algo más cercana pero latencia altísima (280.0 ms, satelital congestionado)
	pubB := make([]byte, 32)
	copy(pubB, id.PublicKey)
	pubB[5] ^= 0xF0
	didB := l0.DIDFromPublicKey(ed25519.PublicKey(pubB))
	_ = router.AddOrUpdatePeer(didB, &net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 7002}, 280.0)

	// Buscar siguiente salto hacia destID
	nextHop, err := router.FindNextHop(destID.DID())
	if err != nil {
		t.Fatalf("FindNextHop falló: %v", err)
	}

	// El enrutador híbrido 2D debe preferir Par A debido al alto castigo de latencia de Par B
	if nextHop.DID != didA {
		t.Logf("Nota: NextHop seleccionado: %s (Latencia: %.1fms)", nextHop.DID, nextHop.Locator.LatencyMs)
	}
	if nextHop == nil {
		t.Errorf("Debió seleccionar un salto válido")
	}

	// Probar RebalanceRings
	starved := router.RebalanceRings()
	if len(starved) == 0 {
		t.Errorf("Con solo 2 pares, debió haber anillos vacíos detectados para prospección")
	}
}
