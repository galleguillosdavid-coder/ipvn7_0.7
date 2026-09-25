package mqttbridge

import (
	"bytes"
	"net"
	"testing"
	"time"
)

func TestMQTTBridge_PhysicalTCPLoopback(t *testing.T) {
	receivedTopic := make(chan string, 1)
	receivedPayload := make(chan []byte, 1)

	bridge := NewMQTTBridge("127.0.0.1:0", func(topic string, payload []byte) {
		receivedTopic <- topic
		receivedPayload <- payload
	})

	if err := bridge.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer bridge.Stop()

	addr := bridge.Addr().String()

	// Conectar cliente TCP físico real
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		t.Fatalf("Dial TCP to MQTT bridge failed: %v", err)
	}
	defer conn.Close()

	// Enviar paquete CONNECT: [0x10, len=12, proto_len=4, 'M','Q','T','T', lvl=4, flags=2, keepalive=60, client_len=0]
	connectPacket := []byte{
		0x10, 0x0c, 0x00, 0x04, 'M', 'Q', 'T', 'T', 0x04, 0x02, 0x00, 0x3c, 0x00, 0x00,
	}
	if _, err := conn.Write(connectPacket); err != nil {
		t.Fatalf("Write CONNECT failed: %v", err)
	}

	// Leer CONNACK esperado: [0x20, 0x02, 0x00, 0x00]
	resp := make([]byte, 4)
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	if _, err := conn.Read(resp); err != nil {
		t.Fatalf("Read CONNACK failed: %v", err)
	}
	if !bytes.Equal(resp, []byte{0x20, 0x02, 0x00, 0x00}) {
		t.Fatalf("unexpected CONNACK response: %x", resp)
	}

	// Enviar PINGREQ: [0xC0, 0x00]
	if _, err := conn.Write([]byte{0xC0, 0x00}); err != nil {
		t.Fatalf("Write PINGREQ failed: %v", err)
	}
	pingResp := make([]byte, 2)
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	if _, err := conn.Read(pingResp); err != nil {
		t.Fatalf("Read PINGRESP failed: %v", err)
	}
	if !bytes.Equal(pingResp, []byte{0xD0, 0x00}) {
		t.Fatalf("unexpected PINGRESP response: %x", pingResp)
	}

	// Enviar PUBLISH: Topic "ipvn7/test", Payload "sensor_val_42"
	topicStr := "ipvn7/test"
	payloadStr := "sensor_val_42"
	pubPacket := []byte{0x30, byte(2 + len(topicStr) + len(payloadStr))}
	pubPacket = append(pubPacket, 0x00, byte(len(topicStr)))
	pubPacket = append(pubPacket, []byte(topicStr)...)
	pubPacket = append(pubPacket, []byte(payloadStr)...)

	if _, err := conn.Write(pubPacket); err != nil {
		t.Fatalf("Write PUBLISH failed: %v", err)
	}

	select {
	case top := <-receivedTopic:
		if top != topicStr {
			t.Fatalf("topic mismatch: got %s, expected %s", top, topicStr)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for MQTT publish topic")
	}

	select {
	case pay := <-receivedPayload:
		if string(pay) != payloadStr {
			t.Fatalf("payload mismatch: got %s, expected %s", string(pay), payloadStr)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for MQTT publish payload")
	}
}
