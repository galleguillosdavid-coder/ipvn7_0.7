package core

import (
	"bytes"
	"net"
	"testing"
	"time"
)

func TestI7UDPAdapter_PhysicalTransmissionLoopback(t *testing.T) {
	// 1. Iniciar adaptadores en dos sockets UDP reales del sistema operativo
	adapterA, err := ListenI7UDP("127.0.0.1:0")
	if err != nil {
		t.Fatalf("ListenI7UDP A falló: %v", err)
	}
	defer adapterA.Close()

	adapterB, err := ListenI7UDP("127.0.0.1:0")
	if err != nil {
		t.Fatalf("ListenI7UDP B falló: %v", err)
	}
	defer adapterB.Close()

	addrA := adapterA.LocalAddr().(*net.UDPAddr)
	addrB := adapterB.LocalAddr().(*net.UDPAddr)

	// 2. Transmisión física A -> B
	originalPayloadA := []byte("CARGA_UTIL_NUCLEO_MINIMO_I7_A_TO_B")
	tagA := []byte("tag-poly1305-16b")
	containerA, err := NewContainer(42, originalPayloadA, tagA)
	if err != nil {
		t.Fatalf("NewContainer A falló: %v", err)
	}

	err = adapterA.SendContainer(containerA, addrB)
	if err != nil {
		t.Fatalf("SendContainer A->B falló: %v", err)
	}

	// 3. Recepción en B desde el socket físico
	receivedB, fromAddrB, err := adapterB.ReceiveContainer(2 * time.Second)
	if err != nil {
		t.Fatalf("ReceiveContainer B falló: %v", err)
	}

	if fromAddrB.Port != addrA.Port {
		t.Fatalf("Remitente no coincide con puerto local de A: %d vs %d", fromAddrB.Port, addrA.Port)
	}
	if receivedB.ChannelID != 42 {
		t.Fatalf("ChannelID corrupto en B: %d", receivedB.ChannelID)
	}
	if !bytes.Equal(receivedB.Payload, originalPayloadA) {
		t.Fatalf("Payload recibido en B difiere del emitido por A: %s", string(receivedB.Payload))
	}
	if !VerifyIntegrity(tagA, receivedB.Tag) {
		t.Fatalf("Fallo en verificación de integridad en B")
	}

	// 4. Retorno físico bidireccional B -> A
	originalPayloadB := []byte("RESPUESTA_NUCLEO_MINIMO_I7_B_TO_A_OK")
	tagB := []byte("tag-poly1305-ack")
	containerB, err := NewContainer(42, originalPayloadB, tagB)
	if err != nil {
		t.Fatalf("NewContainer B falló: %v", err)
	}

	err = adapterB.SendContainer(containerB, addrA)
	if err != nil {
		t.Fatalf("SendContainer B->A falló: %v", err)
	}

	receivedA, _, err := adapterA.ReceiveContainer(2 * time.Second)
	if err != nil {
		t.Fatalf("ReceiveContainer A falló: %v", err)
	}

	if !bytes.Equal(receivedA.Payload, originalPayloadB) {
		t.Fatalf("Payload recibido en A difiere del emitido por B: %s", string(receivedA.Payload))
	}
	if !VerifyIntegrity(tagB, receivedA.Tag) {
		t.Fatalf("Fallo en verificación de integridad en A")
	}
}

func TestI7UDPAdapter_SerializationValidation(t *testing.T) {
	// Validación de rechazo de trama corrupta
	_, err := DeserializeContainer([]byte{0x07, 0x01})
	if err == nil {
		t.Fatalf("DeserializeContainer debió fallar ante trama truncada")
	}

	// Validación de rechazo de versión no compatible
	badVersionContainer := &I7Container{Version: 0x99, ChannelID: 1, Payload: []byte("test")}
	_, err = SerializeContainer(badVersionContainer)
	if err == nil {
		t.Fatalf("SerializeContainer debió rechazar versión incompatible")
	}
}
