package l0

import (
	"bytes"
	"testing"
	"time"
)

func TestPacketBatcher_BatchAccumulationAndFlush(t *testing.T) {
	pool := NewBatchBufferPool()
	dispatchedBatches := make([][][]byte, 0)

	handler := func(batch []*PacketBuffer) error {
		captured := make([][]byte, len(batch))
		for i, buf := range batch {
			data := make([]byte, buf.Length)
			copy(data, buf.Bytes())
			captured[i] = data
		}
		dispatchedBatches = append(dispatchedBatches, captured)
		return nil
	}

	batchSize := 4
	batcher := NewPacketBatcher(batchSize, 10*time.Millisecond, pool, handler)

	// Enviar 3 paquetes (no debe disparar flush automático todavía)
	for i := 0; i < 3; i++ {
		payload := []byte{byte(i + 1), 0xAA, 0xBB}
		if err := batcher.Push(payload); err != nil {
			t.Fatalf("Push %d failed: %v", i, err)
		}
	}

	if len(dispatchedBatches) != 0 {
		t.Fatalf("expected 0 batches dispatched before reaching batchSize, got: %d", len(dispatchedBatches))
	}

	// Enviar el 4to paquete (debe disparar flush automático del lote de 4)
	if err := batcher.Push([]byte{0x04, 0xAA, 0xBB}); err != nil {
		t.Fatalf("Push 4 failed: %v", err)
	}

	if len(dispatchedBatches) != 1 {
		t.Fatalf("expected 1 batch dispatched, got: %d", len(dispatchedBatches))
	}
	if len(dispatchedBatches[0]) != 4 {
		t.Fatalf("expected batch size 4, got: %d", len(dispatchedBatches[0]))
	}

	// Enviar 2 paquetes más y hacer Flush manual
	_ = batcher.Push([]byte("pkt5"))
	_ = batcher.Push([]byte("pkt6"))

	if err := batcher.Flush(); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	if len(dispatchedBatches) != 2 {
		t.Fatalf("expected 2 batches dispatched after manual flush, got: %d", len(dispatchedBatches))
	}
	if len(dispatchedBatches[1]) != 2 {
		t.Fatalf("expected second batch to contain 2 packets, got: %d", len(dispatchedBatches[1]))
	}
	if !bytes.Equal(dispatchedBatches[1][0], []byte("pkt5")) {
		t.Fatalf("payload mismatch in batch: got %s, expected pkt5", dispatchedBatches[1][0])
	}

	// Probar descarte de paquete que exceda MTU canónico 1280B
	oversized := make([]byte, CanonicalPacketSize+1)
	if err := batcher.Push(oversized); err == nil {
		t.Fatal("expected error on oversized packet > 1280B, got nil")
	}

	// Cerrar batcher
	if err := batcher.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if err := batcher.Push([]byte("closed")); err == nil {
		t.Fatal("expected error pushing to closed batcher")
	}
}

func TestBatchBufferPool_Recycle(t *testing.T) {
	pool := NewBatchBufferPool()
	buf1 := pool.Get()
	buf1.Data[0] = 0xFF
	buf1.Length = 100

	pool.Put(buf1)

	buf2 := pool.Get()
	if buf2.Length != 0 {
		t.Fatalf("expected recycled buffer length to be reset to 0, got: %d", buf2.Length)
	}
	pool.Put(buf2)
}
