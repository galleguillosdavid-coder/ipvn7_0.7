package l1

import (
	"bytes"
	"testing"
)

func TestSubPortMux_RegisterAndRoute(t *testing.T) {
	mux := NewSubPortMux()

	receivedChat := make(chan []byte, 1)
	receivedTelemetry := make(chan []byte, 1)

	err := mux.Register(SubPortChat, func(srcDID string, payload []byte) {
		receivedChat <- payload
	})
	if err != nil {
		t.Fatalf("Register SubPortChat failed: %v", err)
	}

	err = mux.Register(SubPortTelemetry, func(srcDID string, payload []byte) {
		receivedTelemetry <- payload
	})
	if err != nil {
		t.Fatalf("Register SubPortTelemetry failed: %v", err)
	}

	// Register with nil handler
	if err := mux.Register(SubPortDAGStore, nil); err == nil {
		t.Fatal("expected error on nil handler registration")
	}

	// Encapsular payload de chat
	chatPayload := []byte("Hola ipvn7 subports")
	encChat := EncodeSubPortFrame(SubPortChat, chatPayload)
	if len(encChat) != len(chatPayload)+2 {
		t.Fatalf("unexpected frame length: got %d, expected %d", len(encChat), len(chatPayload)+2)
	}

	// Decodificar y despachar
	sp, decPayload, err := DecodeSubPortFrame(encChat)
	if err != nil {
		t.Fatalf("DecodeSubPortFrame failed: %v", err)
	}
	if sp != SubPortChat {
		t.Fatalf("expected SubPortChat (%d), got %d", SubPortChat, sp)
	}

	handled := mux.Dispatch(sp, "did:ipvn7:test-node-b", decPayload)
	if !handled {
		t.Fatal("expected handled=true")
	}

	select {
	case msg := <-receivedChat:
		if !bytes.Equal(msg, chatPayload) {
			t.Fatalf("chat payload mismatch: got %s, expected %s", msg, chatPayload)
		}
	default:
		t.Fatal("chat handler was not invoked")
	}

	// Encapsular y despachar telemetría
	telemPayload := []byte(`{"rssi": -42, "battery": 98}`)
	encTelem := EncodeSubPortFrame(SubPortTelemetry, telemPayload)
	spTelem, decTelem, err := DecodeSubPortFrame(encTelem)
	if err != nil {
		t.Fatalf("Decode telemetry frame failed: %v", err)
	}
	handled = mux.Dispatch(spTelem, "did:ipvn7:test-node-b", decTelem)
	if !handled {
		t.Fatal("expected handled=true for telemetry")
	}

	select {
	case msg := <-receivedTelemetry:
		if !bytes.Equal(msg, telemPayload) {
			t.Fatalf("telem payload mismatch: got %s, expected %s", msg, telemPayload)
		}
	default:
		t.Fatal("telemetry handler was not invoked")
	}

	// Probar frame menor a 2 bytes
	_, _, err = DecodeSubPortFrame([]byte{0x01})
	if err == nil {
		t.Fatal("expected error decoding frame < 2 bytes")
	}

	// Probar sub-port no registrado
	handled = mux.Dispatch(SubPortDynamicMax, "did:ipvn7:test", []byte("unhandled"))
	if handled {
		t.Fatal("expected handled=false for unregistered sub-port")
	}

	// Probar desregistro
	mux.Unregister(SubPortChat)
	handled = mux.Dispatch(SubPortChat, "did:ipvn7:test", chatPayload)
	if handled {
		t.Fatal("expected handled=false after unregistering SubPortChat")
	}
}
