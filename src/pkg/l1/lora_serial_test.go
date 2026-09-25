package l1

import (
	"bytes"
	"net"
	"testing"
	"time"
)

func TestLoRaFraming_EncodeDecode(t *testing.T) {
	payload := []byte("Paquete LoRa ipvn7 v0.7 sobre 915MHz")

	frame := EncodeLoRaFrame(payload)
	if len(frame) != 2+2+len(payload)+1 {
		t.Fatalf("unexpected frame size: got %d, expected %d", len(frame), 2+2+len(payload)+1)
	}

	decoded, err := DecodeLoRaFrame(frame)
	if err != nil {
		t.Fatalf("DecodeLoRaFrame failed: %v", err)
	}

	if !bytes.Equal(decoded, payload) {
		t.Fatalf("payload mismatch: got %s, expected %s", decoded, payload)
	}

	// Probar corrupción de checksum por ruido electromagnético
	corrupted := make([]byte, len(frame))
	copy(corrupted, frame)
	corrupted[len(corrupted)-1] ^= 0xFF // Invertir checksum
	_, err = DecodeLoRaFrame(corrupted)
	if err == nil {
		t.Fatal("expected error on corrupted checksum, got nil")
	}

	// Probar trama con SyncWord erróneo
	corruptedSync := make([]byte, len(frame))
	copy(corruptedSync, frame)
	corruptedSync[0] = 0x00
	_, err = DecodeLoRaFrame(corruptedSync)
	if err == nil {
		t.Fatal("expected error on invalid SyncWord, got nil")
	}
}

func TestLoRaSerialBridge_DuplexCommunication(t *testing.T) {
	// Usar net.Pipe() para crear un canal bidireccional físico en memoria sin sockets TCP
	clientRWC, serverRWC := net.Pipe()
	defer clientRWC.Close()
	defer serverRWC.Close()

	receivedCh := make(chan []byte, 1)

	bridge := NewLoRaSerialBridge("COM3", 115200, serverRWC, func(payload []byte) {
		receivedCh <- payload
	})

	if err := bridge.Start(); err != nil {
		t.Fatalf("Start bridge failed: %v", err)
	}
	defer bridge.Stop()

	testMsg := []byte("ALERTA_FIELD_OPS_DESPLIEGUE")
	frame := EncodeLoRaFrame(testMsg)

	// Simular transmisión desde el transceptor LoRa hacia el bridge
	go func() {
		_, _ = clientRWC.Write(frame)
	}()

	select {
	case rx := <-receivedCh:
		if !bytes.Equal(rx, testMsg) {
			t.Fatalf("received message mismatch: got %s, expected %s", rx, testMsg)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for LoRa frame reception")
	}

	// Probar transmisión en sentido inverso (bridge.SendFrame)
	go func() {
		_ = bridge.SendFrame([]byte("ACK_DESDE_GATEWAY"))
	}()

	rxBuf := make([]byte, 256)
	_ = clientRWC.SetDeadline(time.Now().Add(2 * time.Second))
	n, err := clientRWC.Read(rxBuf)
	if err != nil {
		t.Fatalf("Read from bridge failed: %v", err)
	}

	decodedTx, err := DecodeLoRaFrame(rxBuf[:n])
	if err != nil {
		t.Fatalf("Decode transmitted frame failed: %v", err)
	}
	if string(decodedTx) != "ACK_DESDE_GATEWAY" {
		t.Fatalf("transmitted content mismatch: got %s, expected ACK_DESDE_GATEWAY", string(decodedTx))
	}
}
