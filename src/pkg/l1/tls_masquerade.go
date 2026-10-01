// Package l1 implementa el camuflaje criptográfico de tramas ipvn7 bajo el estándar
// TLS 1.3 (RFC 8446) para evasión de inspección profunda de paquetes (DPI) y censura.
package l1

import (
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	// TLSContentTypeApplicationData tipo de registro estándar para datos cifrados
	TLSContentTypeApplicationData byte = 0x17
	// TLSContentTypeHandshake tipo de registro para emulación de saludo inicial
	TLSContentTypeHandshake byte = 0x16

	// TLSVersion12WireFormat versión en cable estándar obligatoria según RFC 8446 §5.1 (0x0303)
	TLSVersion12WireFormat uint16 = 0x0303

	// TLSRecordHeaderLen tamaño de la cabecera del registro TLS en bytes
	TLSRecordHeaderLen = 5

	// MaxTLSRecordPayload tamaño máximo de carga útil permitido por TLS 1.3 (16 KB)
	MaxTLSRecordPayload = 16384
)

var (
	ErrTLSRecordTooShort  = errors.New("tls_masquerade: trama más corta que cabecera TLS de 5 bytes")
	ErrTLSInvalidType     = errors.New("tls_masquerade: ContentType no reconocido")
	ErrTLSInvalidVersion  = errors.New("tls_masquerade: versión TLS no coincide con 0x0303 RFC 8446")
	ErrTLSPayloadTooLarge = errors.New("tls_masquerade: longitud de carga útil excede límite TLS de 16 KB")
	ErrTLSLenMismatch     = errors.New("tls_masquerade: longitud declarada no coincide con bytes recibidos")
)

// WrapTLSRecord encapsula un datagrama ipvn7 dentro de un registro legítimo TLS 1.3
func WrapTLSRecord(payload []byte) ([]byte, error) {
	n := len(payload)
	if n == 0 {
		return nil, errors.New("tls_masquerade: payload no puede ser vacío")
	}
	if n > MaxTLSRecordPayload {
		return nil, ErrTLSPayloadTooLarge
	}

	record := make([]byte, TLSRecordHeaderLen+n)
	// Byte 0: ContentType
	record[0] = TLSContentTypeApplicationData
	// Bytes 1-2: Legacy Record Version (0x0303)
	binary.BigEndian.PutUint16(record[1:3], TLSVersion12WireFormat)
	// Bytes 3-4: Length
	binary.BigEndian.PutUint16(record[3:5], uint16(n))
	// Bytes 5..N: Encrypted payload
	copy(record[TLSRecordHeaderLen:], payload)

	return record, nil
}

// UnwrapTLSRecord valida y desencapsula la carga útil ipvn7 extrayéndola del registro TLS
func UnwrapTLSRecord(data []byte) ([]byte, error) {
	if len(data) < TLSRecordHeaderLen {
		return nil, ErrTLSRecordTooShort
	}

	contentType := data[0]
	if contentType != TLSContentTypeApplicationData && contentType != TLSContentTypeHandshake {
		return nil, fmt.Errorf("%w: 0x%02x", ErrTLSInvalidType, contentType)
	}

	version := binary.BigEndian.Uint16(data[1:3])
	if version != TLSVersion12WireFormat {
		return nil, fmt.Errorf("%w: 0x%04x", ErrTLSInvalidVersion, version)
	}

	length := int(binary.BigEndian.Uint16(data[3:5]))
	if length > MaxTLSRecordPayload {
		return nil, ErrTLSPayloadTooLarge
	}

	if len(data) < TLSRecordHeaderLen+length {
		return nil, fmt.Errorf("%w (esperados %d, recibidos %d)", ErrTLSLenMismatch, TLSRecordHeaderLen+length, len(data))
	}

	payload := make([]byte, length)
	copy(payload, data[TLSRecordHeaderLen:TLSRecordHeaderLen+length])
	return payload, nil
}

// IsTLSRecord comprueba de manera ultrarrápida si una trama recibida tiene cabecera de registro TLS
func IsTLSRecord(data []byte) bool {
	if len(data) < TLSRecordHeaderLen {
		return false
	}
	return (data[0] == TLSContentTypeApplicationData || data[0] == TLSContentTypeHandshake) &&
		binary.BigEndian.Uint16(data[1:3]) == TLSVersion12WireFormat
}
