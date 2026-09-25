package devicebridge

import (
	"bytes"
	"net"
	"testing"
	"time"
)

func TestBuildMagicPacket_Format(t *testing.T) {
	mac, err := net.ParseMAC("00:11:22:33:44:55")
	if err != nil {
		t.Fatalf("ParseMAC failed: %v", err)
	}

	packet, err := BuildMagicPacket(mac)
	if err != nil {
		t.Fatalf("BuildMagicPacket failed: %v", err)
	}

	if len(packet) != 102 {
		t.Fatalf("expected Magic Packet size 102, got: %d", len(packet))
	}

	// Verificar 6 bytes 0xFF iniciales
	for i := 0; i < 6; i++ {
		if packet[i] != 0xFF {
			t.Fatalf("expected byte %d to be 0xFF, got: 0x%02x", i, packet[i])
		}
	}

	// Verificar 16 repeticiones de la dirección MAC
	for i := 0; i < 16; i++ {
		segment := packet[6+i*6 : 6+(i+1)*6]
		if !bytes.Equal(segment, mac) {
			t.Fatalf("MAC repetition %d mismatch: got %x, expected %x", i, segment, mac)
		}
	}
}

func TestWoLGateway_RegisterAndPhysicalEmission(t *testing.T) {
	// Levantar un listener UDP local para recibir el Magic Packet real
	ln, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatalf("ListenUDP failed: %v", err)
	}
	defer ln.Close()

	localTarget := ln.LocalAddr().String()
	gw := NewWoLGateway(localTarget)

	// Registrar un dispositivo con MAC conocida
	dev, err := gw.RegisterShadowDevice("aa:bb:cc:dd:ee:ff", "SmartTV_Living", net.ParseIP("192.168.1.50"))
	if err != nil {
		t.Fatalf("RegisterShadowDevice failed: %v", err)
	}

	expectedDID := "did:ipvn7:shadow:aabbccddeeff"
	if dev.ShadowDID != expectedDID {
		t.Fatalf("expected Shadow DID %s, got: %s", expectedDID, dev.ShadowDID)
	}

	// Emisión física del Magic Packet
	if err := gw.Wake(dev.ShadowDID); err != nil {
		t.Fatalf("Wake failed: %v", err)
	}

	// Recibir paquete del socket UDP físico
	buf := make([]byte, 256)
	_ = ln.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, _, err := ln.ReadFrom(buf)
	if err != nil {
		t.Fatalf("ReadFrom UDP listener failed: %v", err)
	}

	if n != 102 {
		t.Fatalf("expected 102 bytes received, got: %d", n)
	}

	// Verificar que el paquete recibido coincide exactamente con el Magic Packet de la MAC
	expectedPacket, _ := BuildMagicPacket(dev.MAC)
	if !bytes.Equal(buf[:n], expectedPacket) {
		t.Fatal("received payload does not match expected Magic Packet")
	}

	// Verificar que LastWake se actualizó
	savedDev, exists := gw.GetDevice(dev.ShadowDID)
	if !exists || savedDev.LastWake.IsZero() {
		t.Fatal("expected LastWake timestamp to be set")
	}
}
