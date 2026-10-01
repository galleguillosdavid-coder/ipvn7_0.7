package core

import (
	"bytes"
	"net"
	"testing"
	"time"

	"ipvn7/pkg/l0"
	"ipvn7/pkg/l1"
	"ipvn7/pkg/l2"
)

func TestSmartComponentGateway_Lifecycle(t *testing.T) {
	id, err := l0.GenerateIdentity()
	if err != nil {
		t.Fatalf("error generando identidad: %v", err)
	}

	router := l1.NewKleinbergRouter(id)
	telemetry := l2.NewTelemetryRingBuffer()
	fw := l1.NewZTNAFirewall(false) // permissive para el test

	gw := NewSmartComponentGateway(id, router, telemetry, fw)
	defer gw.Stop()

	// 1. Registro de Componente
	comp := &ComponentRegistration{
		ID:           "test_chat_service",
		Name:         "Sovereign Chat Service",
		Version:      "0.7.0",
		Capabilities: []string{"chat:e2ee", "messaging:p2p"},
		Transport:    "websocket",
		Endpoint:     "ws://127.0.0.1:7070/api/v1/ws",
	}

	if err := gw.RegisterComponent(comp); err != nil {
		t.Fatalf("error registrando componente: %v", err)
	}

	// 2. Comprobar duplicado
	if err := gw.RegisterComponent(comp); err == nil {
		t.Fatalf("se esperaba error de duplicado pero fue nil")
	}

	// 3. Suscripción a eventos
	events := gw.SubscribeEvents("test_subscriber")
	defer gw.UnsubscribeEvents("test_subscriber")

	// 4. Heartbeat
	if err := gw.Heartbeat("test_chat_service"); err != nil {
		t.Fatalf("error enviando heartbeat: %v", err)
	}

	// 5. Lista de componentes
	list := gw.ListComponents()
	if len(list) != 1 {
		t.Fatalf("se esperaba 1 componente registrado, se obtuvieron %d", len(list))
	}
	if list[0].ID != "test_chat_service" {
		t.Errorf("ID inesperado: %s", list[0].ID)
	}

	// 6. Publicación de evento manual
	gw.PublishEvent(MeshEvent{
		Type:      EventPeerJoined,
		Timestamp: time.Now().UnixNano(),
		Source:    id.DID(),
		Payload:   map[string]interface{}{"peer": "test_peer"},
	})

	select {
	case ev := <-events:
		if ev.Type != EventPeerJoined {
			t.Errorf("tipo de evento inesperado: %s", ev.Type)
		}
	case <-time.After(500 * time.Millisecond):
		t.Errorf("timeout esperando evento en el bus")
	}

	// 7. Desregistro
	if err := gw.UnregisterComponent("test_chat_service"); err != nil {
		t.Fatalf("error desregistrando: %v", err)
	}

	if len(gw.ListComponents()) != 0 {
		t.Errorf("se esperaba lista vacía tras desregistro")
	}
}

func TestSmartComponentGateway_SendLoopbackDatagram(t *testing.T) {
	id, _ := l0.GenerateIdentity()
	router := l1.NewKleinbergRouter(id)
	telemetry := l2.NewTelemetryRingBuffer()
	fw := l1.NewZTNAFirewall(false)

	gw := NewSmartComponentGateway(id, router, telemetry, fw)
	defer gw.Stop()

	events := gw.SubscribeEvents("loopback_tester")
	defer gw.UnsubscribeEvents("loopback_tester")

	env := &DatagramEnvelope{
		SourceDID: id.DID(),
		TargetDID: id.DID(),
		Protocol:  "mesh:echo",
		Payload:   []byte("test_loopback_payload"),
	}

	if err := gw.SendDatagram(env, l1.ClassInteractive); err != nil {
		t.Fatalf("SendDatagram loopback falló: %v", err)
	}

	select {
	case ev := <-events:
		if ev.Type != EventPacketTx {
			t.Errorf("tipo de evento inesperado: %s", ev.Type)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout esperando evento PACKET_TX en loopback")
	}

	snap := telemetry.Snapshot()
	if snap.PacketsTx == 0 || snap.PacketsRx == 0 {
		t.Errorf("telemetría loopback no registró paquetes: Tx=%d Rx=%d", snap.PacketsTx, snap.PacketsRx)
	}
}

func TestSmartComponentGateway_LinkHealingFailover(t *testing.T) {
	id, _ := l0.GenerateIdentity()
	router := l1.NewKleinbergRouter(id)
	telemetry := l2.NewTelemetryRingBuffer()
	fw := l1.NewZTNAFirewall(false)

	// Generar identidades válidas con claves criptográficas reales para los pares
	p1, err := l0.GenerateIdentity()
	if err != nil {
		t.Fatalf("error generando id p1: %v", err)
	}
	p2, err := l0.GenerateIdentity()
	if err != nil {
		t.Fatalf("error generando id p2: %v", err)
	}

	peer1DID := p1.DID()
	peer2DID := p2.DID()

	_ = router.AddOrUpdatePeer(peer1DID, nil, 10.0)
	_ = router.AddOrUpdatePeer(peer2DID, nil, 20.0)

	gw := NewSmartComponentGateway(id, router, telemetry, fw)
	defer gw.Stop()

	healing := l1.NewLinkHealingEngine(router)
	gw.SetHealingEngine(healing)

	// Provocar 3 fallos consecutivos en peer1 para marcarlo degradado
	healing.RecordProbeResult(peer1DID, false, 999.0)
	healing.RecordProbeResult(peer1DID, false, 999.0)
	healing.RecordProbeResult(peer1DID, false, 999.0)

	events := gw.SubscribeEvents("failover_monitor")
	defer gw.UnsubscribeEvents("failover_monitor")

	env := &DatagramEnvelope{
		SourceDID: id.DID(),
		TargetDID: peer1DID,
		Protocol:  "mesh:p2p",
		Payload:   []byte("test_failover_data"),
	}

	if err := gw.SendDatagram(env, l1.ClassInteractive); err != nil {
		t.Fatalf("SendDatagram falló: %v", err)
	}

	select {
	case ev := <-events:
		if ev.Type != EventPacketTx {
			t.Errorf("tipo de evento inesperado: %s", ev.Type)
		}
		isFailover, ok := ev.Payload["failover_route"].(bool)
		if !ok || !isFailover {
			t.Errorf("se esperaba failover_route=true en evento, pero fue %v", ev.Payload["failover_route"])
		}
		viaPeer := ev.Payload["via_peer"].(string)
		if viaPeer != peer2DID {
			t.Errorf("se esperaba desvío hacia peer2 (%s), pero fue %s", peer2DID, viaPeer)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout esperando evento PACKET_TX en failover")
	}
}

func TestSmartComponentGateway_PhysicalUDPForwarding(t *testing.T) {
	id, _ := l0.GenerateIdentity()
	router := l1.NewKleinbergRouter(id)
	telemetry := l2.NewTelemetryRingBuffer()
	fw := l1.NewZTNAFirewall(false)

	gw := NewSmartComponentGateway(id, router, telemetry, fw)
	defer gw.Stop()

	// 1. Levantar socket receptor UDP real
	receiver, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatalf("error levantando receptor UDP: %v", err)
	}
	defer receiver.Close()

	recAddr := receiver.LocalAddr().(*net.UDPAddr)

	// 2. Levantar socket emisor UDP real y vincularlo al gateway
	sender, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatalf("error levantando emisor UDP: %v", err)
	}
	defer sender.Close()

	gw.SetUDPConnection(sender)

	// 3. Registrar par remoto con dirección física del receptor
	peerID, _ := l0.GenerateIdentity()
	peerDID := peerID.DID()
	_ = router.AddOrUpdatePeer(peerDID, recAddr, 1.0)

	// 4. Enviar datagrama a través del gateway
	testPayload := []byte("PAYLOAD_FISICO_ZERO_SIMULATION_OK")
	env := &DatagramEnvelope{
		SourceDID: id.DID(),
		TargetDID: peerDID,
		Protocol:  "mesh:p2p",
		Payload:   testPayload,
	}

	if err := gw.SendDatagram(env, l1.ClassInteractive); err != nil {
		t.Fatalf("SendDatagram físico falló: %v", err)
	}

	// 5. El receptor físico DEBE recibir los bytes reales desde el socket
	buf := make([]byte, 256)
	_ = receiver.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	n, _, err := receiver.ReadFrom(buf)
	if err != nil {
		t.Fatalf("receptor UDP físico no recibió el paquete: %v", err)
	}

	if !bytes.Equal(buf[:n], testPayload) {
		t.Fatalf("corrupción física: recibido %q, esperado %q", buf[:n], testPayload)
	}
}


