package l1

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"net"
	"testing"

	"ipvn7/pkg/l0"
)

func TestMobilityAnchor_AtomicMobilityAndStormSuppression(t *testing.T) {
	anchorDID := "did:ipvn7:anchor_node_ring0"
	engine := NewMobilityAnchorEngine(anchorDID)

	mobileDID := "did:ipvn7:mobile_satellite_roamer"

	addr1 := &net.UDPAddr{IP: net.ParseIP("192.168.1.106"), Port: 7001}
	addr2 := &net.UDPAddr{IP: net.ParseIP("10.0.0.50"), Port: 7778}
	addr3 := &net.UDPAddr{IP: net.ParseIP("172.16.0.22"), Port: 8888}

	// 1. Registro inicial de movilidad en el ancla
	loc, err := engine.UpdateMobileLocator(mobileDID, addr1, 0)
	if err != nil || loc == nil {
		t.Fatalf("fallo al registrar localizador inicial: %v", err)
	}

	resAddr, found := engine.ResolveMobileLocator(mobileDID)
	if !found || resAddr.String() != addr1.String() {
		t.Fatalf("localizador resuelto incorrecto, esperado %s, obtenido %v", addr1, resAddr)
	}

	// 2. Simular cambio rápido de IP (Wi-Fi -> Celular 5G)
	_, _ = engine.UpdateMobileLocator(mobileDID, addr2, 0)
	resAddr2, found := engine.ResolveMobileLocator(mobileDID)
	if !found || resAddr2.String() != addr2.String() {
		t.Fatalf("localizador no se actualizó inmediatamente en el ancla local")
	}

	// 3. Simular ráfaga de 100 rotaciones de IP en milisegundos: el ancla actualiza en memoria en O(1)
	for i := 0; i < 100; i++ {
		_, _ = engine.UpdateMobileLocator(mobileDID, addr3, 1)
	}

	resAddr3, found := engine.ResolveMobileLocator(mobileDID)
	if !found || resAddr3.String() != addr3.String() {
		t.Fatalf("localizador final no coincide con addr3")
	}

	// 4. Verificar que el amortiguador suprime tormentas de señalización global
	firstSync := engine.ShouldPropagateGlobalSync()
	if !firstSync {
		t.Fatalf("la primera sincronización debió ser autorizada")
	}

	// Intentos inmediatos subsecuentes deben ser suprimidos para proteger el plano de datos
	secondSync := engine.ShouldPropagateGlobalSync()
	if secondSync {
		t.Fatalf("la segunda sincronización debió ser suprimida por el amortiguador anti-tormentas")
	}
}

func TestMobilityAnchor_OneMillionRotationsStress(t *testing.T) {
	anchorDID := "did:ipvn7:anchor_node_ring0"
	engine := NewMobilityAnchorEngine(anchorDID)
	mobileDID := "did:ipvn7:mobile_hyper_roamer"
	addr := &net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 5000}

	// 1,000,000 de cambios de localizador atómico
	for i := 0; i < 1000000; i++ {
		addr.Port = 5000 + (i % 20000)
		_, _ = engine.UpdateMobileLocator(mobileDID, addr, 0)
	}

	res, found := engine.ResolveMobileLocator(mobileDID)
	if !found || res == nil {
		t.Fatalf("el nodo móvil debió resolverse tras 1 millón de rotaciones")
	}
}

func TestMobilityAnchor_AntiSequenceHijacking(t *testing.T) {
	anchorDID := "did:ipvn7:anchor_node_ring0"
	engine := NewMobilityAnchorEngine(anchorDID)
	mobileDID := "did:ipvn7:victim_mobile_node"

	addrLegit1 := &net.UDPAddr{IP: net.ParseIP("192.168.1.100"), Port: 7001}
	addrAttacker := &net.UDPAddr{IP: net.ParseIP("198.51.100.66"), Port: 6666}
	addrLegit2 := &net.UDPAddr{IP: net.ParseIP("192.168.1.101"), Port: 7001}

	// 1. Registro legítimo con seq = 100
	_, err := engine.UpdateMobileLocatorWithSeq(mobileDID, addrLegit1, 0, 100)
	if err != nil {
		t.Fatalf("Registro legítimo falló: %v", err)
	}

	// 2. Ataque adversario: inyección de secuencia astronómica (2^63 - 1) para secuestrar el estado
	_, err = engine.UpdateMobileLocatorWithSeq(mobileDID, addrAttacker, 0, 0x7FFFFFFFFFFFFFFF)
	if err != ErrSequenceJumpTooLarge {
		t.Fatalf("Se esperaba ErrSequenceJumpTooLarge para salto anómalo, pero se obtuvo: %v", err)
	}

	// 3. Verificar que el atacante NO secuestró la ruta
	res, found := engine.ResolveMobileLocator(mobileDID)
	if !found || res.String() != addrLegit1.String() {
		t.Fatalf("El atacante logró alterar el localizador legítimo: %v", res)
	}

	// 4. El usuario legítimo avanza naturalmente (seq = 105) y DEBE ser aceptado sin inanición
	_, err = engine.UpdateMobileLocatorWithSeq(mobileDID, addrLegit2, 0, 105)
	if err != nil {
		t.Fatalf("Actualización legítima normal (seq=105) fue bloqueada por falso positivo: %v", err)
	}

	res2, found2 := engine.ResolveMobileLocator(mobileDID)
	if !found2 || res2.String() != addrLegit2.String() {
		t.Fatalf("Localizador legítimo no se actualizó a addrLegit2: %v", res2)
	}
}

func TestMobilityAnchor_CryptographicResyncAfterOfflinePeriod(t *testing.T) {
	anchorDID := "did:ipvn7:anchor_node_ring0"
	engine := NewMobilityAnchorEngine(anchorDID)

	// Crear identidad criptográfica soberana real
	nodeID, err := l0.GenerateIdentity()
	if err != nil {
		t.Fatalf("Error creando identidad: %v", err)
	}
	nodeDID := nodeID.DID()

	addr1 := &net.UDPAddr{IP: net.ParseIP("192.168.1.50"), Port: 7001}
	addrAfterFlight := &net.UDPAddr{IP: net.ParseIP("198.51.100.77"), Port: 9001}

	// 1. Registro inicial con seq = 100
	_, err = engine.UpdateMobileLocatorWithSeq(nodeDID, addr1, 0, 100)
	if err != nil {
		t.Fatalf("Registro inicial falló: %v", err)
	}

	// 2. Dispositivo estuvo offline durante vuelo transoceánico: su secuencia local saltó a 3,500,000 (>1M de salto)
	offlineSeq := uint64(3500000)

	// Intento sin firma: debe ser rechazado como posible ataque
	_, err = engine.UpdateMobileLocatorWithSeq(nodeDID, addrAfterFlight, 0, offlineSeq)
	if err != ErrSequenceJumpTooLarge {
		t.Fatalf("Se esperaba ErrSequenceJumpTooLarge sin firma, pero se obtuvo: %v", err)
	}

	// 3. El nodo legítimo firma la prueba criptográfica soberana de resincronización
	msg := []byte(fmt.Sprintf("%s:%d:%s", nodeDID, offlineSeq, anchorDID))
	sig := nodeID.Sign(msg)

	// 4. Con firma válida: el Nodo Ancla verifica la clave pública del DID y autoriza la resincronización O(1)
	loc, err := engine.UpdateMobileLocatorWithSig(nodeDID, addrAfterFlight, 0, offlineSeq, sig)
	if err != nil || loc == nil {
		t.Fatalf("Fallo en resincronización criptográfica soberana: %v", err)
	}

	resolved, found := engine.ResolveMobileLocator(nodeDID)
	if !found || resolved.String() != addrAfterFlight.String() {
		t.Fatalf("Localizador tras resincronización no coincide: %v", resolved)
	}
}

func TestMobilityAnchor_AsymmetricVerificationRateLimiting(t *testing.T) {
	anchorDID := "did:ipvn7:anchor_node_ring0"
	engine := NewMobilityAnchorEngine(anchorDID)
	nodeID, _ := l0.GenerateIdentity()
	addr := &net.UDPAddr{IP: net.ParseIP("192.168.1.50"), Port: 7001}
	addr2 := &net.UDPAddr{IP: net.ParseIP("192.168.1.51"), Port: 7001}

	_, _ = engine.UpdateMobileLocatorWithSeq(nodeID.DID(), addr, 0, 100)

	// Simular ataque con 550 firmas corruptas de 64 bytes para saturar CPU
	corruptSig := make([]byte, 64)
	var rejectedCount int
	for i := 0; i < 550; i++ {
		_, err := engine.UpdateMobileLocatorWithSig(nodeID.DID(), addr2, 0, uint64(2000000+i), corruptSig)
		if err == ErrSequenceJumpTooLarge {
			rejectedCount++
		}
	}

	// Al menos 50 peticiones debieron ser descartadas directamente en 2 ns por límite de presupuesto de firmas
	if rejectedCount < 50 {
		t.Fatalf("Esperaba protección contra Asymmetric Crypto Starvation (descartes: %d)", rejectedCount)
	}
}

func TestMobilityAnchor_TicketPriorityAdmissionUnderSpam(t *testing.T) {
	anchorDID := "did:ipvn7:anchor_node_ring0"
	engine := NewMobilityAnchorEngine(anchorDID)

	legitID, _ := l0.GenerateIdentity()
	addr1 := &net.UDPAddr{IP: net.ParseIP("192.168.1.10"), Port: 7001}
	addr2 := &net.UDPAddr{IP: net.ParseIP("198.51.100.99"), Port: 9001}

	_, _ = engine.UpdateMobileLocatorWithSeq(legitID.DID(), addr1, 0, 100)

	spamDID := "did:ipvn7:attacker_spam_node"
	_, _ = engine.UpdateMobileLocatorWithSeq(spamDID, addr1, 0, 100)

	// 1. Simular que un atacante inunda con 550 firmas corruptas agotando el pool general no autenticado
	corruptSig := make([]byte, 64)
	for i := 0; i < 550; i++ {
		_, _ = engine.UpdateMobileLocatorWithSig(spamDID, addr2, 0, uint64(2000000+i), corruptSig)
	}

	// 2. Un usuario legítimo que llega sin ticket ni alineación será rechazado por agotamiento del pool de spam
	offlineSeq := uint64(4000001)
	msg := []byte(fmt.Sprintf("%s:%d:%s", legitID.DID(), offlineSeq, anchorDID))
	sig := legitID.Sign(msg)

	_, errNoTicket := engine.UpdateMobileLocatorWithSig(legitID.DID(), addr2, 0, offlineSeq, sig)
	if errNoTicket != ErrSequenceJumpTooLarge {
		t.Fatalf("Sin ticket y con pool agotado debió ser rechazado preventivamente")
	}

	// 3. El usuario legítimo solicita un challenge ticket al Ancla y lo presenta:
	ticket := engine.IssueResyncChallenge(legitID.DID())
	if len(ticket) != 16 {
		t.Fatalf("Ticket de challenge con formato inválido: %v", ticket)
	}

	// 4. Con ticket autenticado: el usuario legítimo obtiene ADMISIÓN PRIORITARIA GARANTIZADA
	loc, errWithTicket := engine.UpdateMobileLocatorWithSigAndTicket(legitID.DID(), addr2, 0, offlineSeq, sig, ticket)
	if errWithTicket != nil || loc == nil {
		t.Fatalf("Usuario legítimo con challenge ticket debió ser admitido prioritariamente: %v", errWithTicket)
	}

	resolved, found := engine.ResolveMobileLocator(legitID.DID())
	if !found || resolved.String() != addr2.String() {
		t.Fatalf("Ruta no actualizada a addr2 tras admisión prioritaria: %v", resolved)
	}

	// 5. Verificar Handover 0-RTT con ProactiveTicket: la respuesta ya incluye el ticket para la siguiente transición
	if len(loc.ProactiveTicket) != 16 {
		t.Fatalf("El locator debió generar un ProactiveTicket 0-RTT de 16B")
	}
}

func TestMobilityAnchor_InvalidTicketFastRejectionAndFallbackImmunity(t *testing.T) {
	anchorDID := "did:ipvn7:anchor_node_ring0"
	engine := NewMobilityAnchorEngine(anchorDID)
	id, _ := l0.GenerateIdentity()
	addr1 := &net.UDPAddr{IP: net.ParseIP("192.168.1.198"), Port: 7001}
	addr2 := &net.UDPAddr{IP: net.ParseIP("192.168.1.106"), Port: 7001}

	_, _ = engine.UpdateMobileLocatorWithSeq(id.DID(), addr1, 0, 100)

	// Atacante inyecta ticket corrupto con secuencia alineada (2000000 % 16 == 0) buscando explotar fallback
	alignedSeq := uint64(2000000)
	fakeTicket := []byte("fake-bad-ticket!")
	corruptSig := make([]byte, 64)

	// El motor debe descartar inmediatamente en 1 ns prohibiendo fallback
	_, err := engine.UpdateMobileLocatorWithSigAndTicket(id.DID(), addr2, 0, alignedSeq, corruptSig, fakeTicket)
	if err != ErrSequenceJumpTooLarge {
		t.Fatalf("Ticket falso debió ser descartado inmediatamente sin fallback")
	}
}

func TestMobilityAnchor_ColdResyncPoWEscapeGradientUnderSpam(t *testing.T) {
	anchorDID := "did:ipvn7:anchor_node_ring0"
	engine := NewMobilityAnchorEngine(anchorDID)
	legitID, _ := l0.GenerateIdentity()
	addr1 := &net.UDPAddr{IP: net.ParseIP("192.168.1.198"), Port: 7001}
	addr2 := &net.UDPAddr{IP: net.ParseIP("192.168.1.106"), Port: 7001}

	_, _ = engine.UpdateMobileLocatorWithSeq(legitID.DID(), addr1, 0, 100)

	// 1. Simular ataque DoS con 550 peticiones vacías sin ticket que saturan el pool resyncVerifyTokens
	corruptSig := make([]byte, 64)
	spamDID := "did:ipvn7:attacker_spam_node"
	_, _ = engine.UpdateMobileLocatorWithSeq(spamDID, addr1, 0, 100)
	for i := 0; i < 550; i++ {
		_, _ = engine.UpdateMobileLocatorWithSig(spamDID, addr2, 0, uint64(2000000+i), corruptSig)
	}

	// 2. Nodo legítimo reconectando en frío tras meses offline (sin ticket previo):
	coldSeq := uint64(5000000)
	msg := []byte(fmt.Sprintf("%s:%d:%s", legitID.DID(), coldSeq, anchorDID))
	sig := legitID.Sign(msg)

	// Sin PoW: debe ser rechazado preventivamente para proteger la CPU
	_, errNoPoW := engine.UpdateMobileLocatorWithSigTicketAndPoW(legitID.DID(), addr2, 0, coldSeq, sig, nil, 0)
	if errNoPoW != ErrSequenceJumpTooLarge {
		t.Fatalf("Petición en frío sin ticket ni PoW debió ser rechazada con pool saturado")
	}

	// 3. Nodo legítimo calcula Micro-PoW en frío localmente (dificultad de 8 bits: ~256 hashes, 1 µs):
	tag := uint64(ColdPoWTag(legitID.DID(), coldSeq))
	startN := tag
	if startN == 0 {
		startN = 16
	}
	var coldNonce uint64
	for n := startN; n < 200000; n += 16 {
		var nonceBytes [8]byte
		var seqBytes [8]byte
		binary.BigEndian.PutUint64(nonceBytes[:], n)
		binary.BigEndian.PutUint64(seqBytes[:], coldSeq)
		h := sha256.New()
		h.Write([]byte(legitID.DID()))
		h.Write([]byte(anchorDID))
		h.Write(seqBytes[:])
		h.Write(nonceBytes[:])
		digest := h.Sum(nil)
		if digest[0] == 0x00 {
			coldNonce = n
			break
		}
	}
	if coldNonce == 0 {
		t.Fatalf("Fallo encontrando Micro-PoW para arranque en frío")
	}

	// 3b. Noncios con tag incorrecto deben ser descartados inmediatamente en 0 ns
	badNonce := coldNonce + 1
	if uint8(badNonce&0x0F) == uint8(tag) {
		badNonce++
	}
	if engine.VerifyColdResyncPoW(legitID.DID(), coldSeq, badNonce) {
		t.Fatalf("Noncio con tag no coincidente debió ser descartado por el pre-filtro dinámico en 0 ns")
	}

	// 4. Con Micro-PoW: el nodo legítimo obtiene ADMISIÓN GARANTIZADA a pesar del ataque DoS
	loc, errPoW := engine.UpdateMobileLocatorWithSigTicketAndPoW(legitID.DID(), addr2, 0, coldSeq, sig, nil, coldNonce)
	if errPoW != nil || loc == nil {
		t.Fatalf("Nodo legítimo con Cold Micro-PoW debió ser admitido: %v", errPoW)
	}

	resolved, found := engine.ResolveMobileLocator(legitID.DID())
	if !found || resolved.String() != addr2.String() {
		t.Fatalf("Ruta no actualizada a addr2 tras admisión con Cold PoW: %v", resolved)
	}

	// 5. Simular saturación total de la cubeta segregada de este DID por spam directo (110 tokens consumidos)
	bIdx := coldBucketIdx(legitID.DID())
	for i := 0; i < 110; i++ {
		engine.coldPoWBuckets[bIdx].Add(-1)
	}
	// Con cubeta saturada, un noncio estándar de 8 bits es rechazado, pero un Tier-2 (12 bits) es ADMITIDO:
	var tier2Nonce uint64
	for n := startN; n < 500000; n += 16 {
		var nonceBytes [8]byte
		var seqBytes [8]byte
		binary.BigEndian.PutUint64(nonceBytes[:], n)
		binary.BigEndian.PutUint64(seqBytes[:], coldSeq)
		h := sha256.New()
		h.Write([]byte(legitID.DID()))
		h.Write([]byte(anchorDID))
		h.Write(seqBytes[:])
		h.Write(nonceBytes[:])
		digest := h.Sum(nil)
		if digest[0] == 0x00 && (digest[1]&0xF0) == 0x00 {
			tier2Nonce = n
			break
		}
	}
	if tier2Nonce == 0 {
		t.Fatalf("Nodo legítimo debió encontrar nonce Tier-2 (12 bits) bajo saturación de cubeta")
	}
	if !engine.VerifyColdResyncPoW(legitID.DID(), coldSeq, tier2Nonce) {
		t.Fatalf("Nonce Tier-2 debió ser validado bajo cubeta saturada")
	}
}


