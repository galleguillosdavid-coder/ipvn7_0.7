package l1_test

import (
	"bytes"
	"testing"

	"ipvn7/pkg/l1"
)

func TestQUICTransportEngine(t *testing.T) {
	engine := l1.NewQUICTransportEngine()

	// 1. Prueba de serialización y deserialización de StreamFrame
	frame := &l1.StreamFrame{
		StreamID: 42,
		Offset:   0,
		Fin:      false,
		Data:     []byte("hola mundo ipvn7 quic"),
	}

	encoded, err := engine.EncodeStreamFrame(frame)
	if err != nil {
		t.Fatalf("error codificando StreamFrame: %v", err)
	}

	decodedStream, decodedACK, err := engine.DecodeFrame(encoded)
	if err != nil {
		t.Fatalf("error decodificando StreamFrame: %v", err)
	}
	if decodedACK != nil {
		t.Fatal("esperaba StreamFrame, pero obtuve ACKFrame")
	}
	if decodedStream.StreamID != frame.StreamID || decodedStream.Offset != frame.Offset {
		t.Fatalf("datos de cabecera no coinciden: %+v", decodedStream)
	}
	if !bytes.Equal(decodedStream.Data, frame.Data) {
		t.Fatalf("payload no coincide: %s vs %s", string(decodedStream.Data), string(frame.Data))
	}

	// 2. Prueba de serialización y deserialización de ACKFrame
	ack := &l1.ACKFrame{
		StreamID:     42,
		LargestAcked: 100,
		AckBitmask:   0xFFFFFFFFFFFFFFFF,
	}

	encodedACK := engine.EncodeACKFrame(ack)
	decodedStream2, decodedACK2, err := engine.DecodeFrame(encodedACK)
	if err != nil {
		t.Fatalf("error decodificando ACKFrame: %v", err)
	}
	if decodedStream2 != nil {
		t.Fatal("esperaba ACKFrame, pero obtuve StreamFrame")
	}
	if decodedACK2.StreamID != ack.StreamID || decodedACK2.LargestAcked != ack.LargestAcked || decodedACK2.AckBitmask != ack.AckBitmask {
		t.Fatalf("campos de ACK no coinciden: %+v", decodedACK2)
	}

	// 3. Prueba de reordenamiento y reensamblaje con fragmentos desordenados
	chunk1 := &l1.StreamFrame{StreamID: 100, Offset: 0, Data: []byte("Parte1-")}
	chunk2 := &l1.StreamFrame{StreamID: 100, Offset: 7, Data: []byte("Parte2-")}
	chunk3 := &l1.StreamFrame{StreamID: 100, Offset: 14, Data: []byte("Parte3")}

	// Llegada desordenada: chunk 2, luego chunk 3, luego chunk 1
	out2 := engine.ProcessIncomingStreamChunk(chunk2)
	if len(out2) != 0 {
		t.Fatalf("esperaba 0 bytes mientras falta el chunk 1, pero obtuve: %s", string(out2))
	}

	out3 := engine.ProcessIncomingStreamChunk(chunk3)
	if len(out3) != 0 {
		t.Fatalf("esperaba 0 bytes mientras falta el chunk 1, pero obtuve: %s", string(out3))
	}

	out1 := engine.ProcessIncomingStreamChunk(chunk1)
	expectedFull := "Parte1-Parte2-Parte3"
	if string(out1) != expectedFull {
		t.Fatalf("reensamblaje falló. Esperaba '%s', obtuve '%s'", expectedFull, string(out1))
	}

	// 4. Validación de payload excesivo
	oversized := &l1.StreamFrame{
		StreamID: 1,
		Offset:   0,
		Data:     make([]byte, l1.MaxPayloadPerStream+1),
	}
	_, err = engine.EncodeStreamFrame(oversized)
	if err == nil {
		t.Fatal("esperaba error por payload excesivo, pero fue nil")
	}
}
