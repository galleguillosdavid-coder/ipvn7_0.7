package coapproxy

import (
	"bytes"
	"net"
	"testing"
	"time"
)

func TestCoAPServer_PhysicalUDPExchange(t *testing.T) {
	// Manejador que responde ACK con código 2.05 Content al recibir un GET
	handler := func(msg CoAPMessage, remoteAddr net.Addr) (*CoAPMessage, error) {
		if msg.Code == CodeGET {
			return &CoAPMessage{
				Type:      TypeAcknowledgement,
				Code:      CodeContent,
				MessageID: msg.MessageID,
				Token:     msg.Token,
				Payload:   []byte(`{"temp": 24.5, "hum": 60}`),
			}, nil
		}
		return nil, nil
	}

	server := NewCoAPServer("127.0.0.1:0", handler)
	if err := server.Start(); err != nil {
		t.Fatalf("Start CoAPServer failed: %v", err)
	}
	defer server.Stop()

	serverUDPAddr := server.LocalAddr().(*net.UDPAddr)

	// Crear cliente UDP físico real
	clientConn, err := net.DialUDP("udp", nil, serverUDPAddr)
	if err != nil {
		t.Fatalf("DialUDP failed: %v", err)
	}
	defer clientConn.Close()

	// Armar request CoAP Confirmable GET con Token "TK1" y MessageID 0x1234
	reqMsg := CoAPMessage{
		Type:      TypeConfirmable,
		Code:      CodeGET,
		MessageID: 0x1234,
		Token:     []byte("TK1"),
		Payload:   nil,
	}

	reqBytes := EncodeCoAP(reqMsg)
	if _, err := clientConn.Write(reqBytes); err != nil {
		t.Fatalf("Write CoAP request failed: %v", err)
	}

	// Leer respuesta física UDP
	buf := make([]byte, 1024)
	_ = clientConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err := clientConn.Read(buf)
	if err != nil {
		t.Fatalf("Read CoAP response failed: %v", err)
	}

	respMsg, err := ParseCoAP(buf[:n])
	if err != nil {
		t.Fatalf("ParseCoAP response failed: %v", err)
	}

	if respMsg.Type != TypeAcknowledgement {
		t.Fatalf("expected TypeAcknowledgement, got %v", respMsg.Type)
	}
	if respMsg.Code != CodeContent {
		t.Fatalf("expected CodeContent, got %v", respMsg.Code)
	}
	if respMsg.MessageID != 0x1234 {
		t.Fatalf("expected MessageID 0x1234, got 0x%04x", respMsg.MessageID)
	}
	if !bytes.Equal(respMsg.Token, []byte("TK1")) {
		t.Fatalf("token mismatch: got %s, expected TK1", respMsg.Token)
	}
	expectedPayload := `{"temp": 24.5, "hum": 60}`
	if string(respMsg.Payload) != expectedPayload {
		t.Fatalf("payload mismatch: got %s, expected %s", string(respMsg.Payload), expectedPayload)
	}
}

func TestCoAP_CodecErrors(t *testing.T) {
	// Probar trama < 4 bytes
	_, err := ParseCoAP([]byte{0x01, 0x02})
	if err == nil {
		t.Fatal("expected error on < 4 bytes")
	}

	// Probar trama con token truncado
	// data[0]=0x05 (tkl=5), pero solo 5 bytes en total (faltan 4 bytes de token)
	_, err = ParseCoAP([]byte{0x45, 0x01, 0x00, 0x01, 0xAA})
	if err == nil {
		t.Fatal("expected error on truncated token")
	}
}
