package l1

import (
	"bytes"
	"testing"
)

func TestTLSMasquerade_WrapUnwrap(t *testing.T) {
	canonicalPayload := make([]byte, 1280)
	for i := range canonicalPayload {
		canonicalPayload[i] = byte(i % 256)
	}

	record, err := WrapTLSRecord(canonicalPayload)
	if err != nil {
		t.Fatalf("WrapTLSRecord failed: %v", err)
	}

	if len(record) != TLSRecordHeaderLen+1280 {
		t.Fatalf("expected record len %d, got: %d", TLSRecordHeaderLen+1280, len(record))
	}

	if !IsTLSRecord(record) {
		t.Fatal("expected IsTLSRecord to return true")
	}

	if record[0] != TLSContentTypeApplicationData {
		t.Fatalf("expected ContentType 0x17, got: 0x%02x", record[0])
	}
	if record[1] != 0x03 || record[2] != 0x03 {
		t.Fatalf("expected version 0x0303, got: 0x%02x%02x", record[1], record[2])
	}

	unwrapped, err := UnwrapTLSRecord(record)
	if err != nil {
		t.Fatalf("UnwrapTLSRecord failed: %v", err)
	}

	if !bytes.Equal(unwrapped, canonicalPayload) {
		t.Fatal("unwrapped payload mismatch with canonical payload")
	}

	// Probar trama corta
	_, err = UnwrapTLSRecord([]byte{0x17, 0x03})
	if err == nil {
		t.Fatal("expected error on truncated record < 5 bytes")
	}

	// Probar versión inválida
	badVersion := make([]byte, len(record))
	copy(badVersion, record)
	badVersion[1] = 0x02 // TLS 1.1 obsoleto
	_, err = UnwrapTLSRecord(badVersion)
	if err == nil {
		t.Fatal("expected error on invalid TLS version")
	}

	// Probar ContentType inválido
	badType := make([]byte, len(record))
	copy(badType, record)
	badType[0] = 0xFF
	_, err = UnwrapTLSRecord(badType)
	if err == nil {
		t.Fatal("expected error on invalid ContentType")
	}
}
