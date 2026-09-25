// Package l1 implementa el motor de enmascaramiento y camuflaje de tráfico TLS 1.3
// conforme a RFC 8446. Permite que el tráfico cuántico de ipvn7 se transporte
// de forma indistinguible de una navegación web HTTPS legítima en el puerto 443,
// burlando firewalls empresariales de inspección profunda de paquetes (DPI/Layer-7).
package l1

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	TLSRecordHandshake       byte   = 0x16
	TLSRecordApplicationData byte   = 0x17
	TLSHandshakeClientHello  byte   = 0x01
	TLSHandshakeServerHello  byte   = 0x02
	TLSVersionLegacy         uint16 = 0x0303 // TLS 1.2 legacy version wire format
	TLSVersion13             uint16 = 0x0304 // TLS 1.3 negotiation extension
	TLSMaxRecordLength       int    = 16384  // RFC 8446 max plaintext per record
	TLSRecordHeaderSize      int    = 5      // Type(1) + Version(2) + Length(2)

	ExtServerName        uint16 = 0x0000
	ExtSupportedGroups   uint16 = 0x000a
	ExtKeyShare          uint16 = 0x0033
	ExtSupportedVersions uint16 = 0x002b

	GroupX25519 uint16 = 0x001d
)

// TLSProfileConfig define el perfil con el que se mimetiza la conexión
type TLSProfileConfig struct {
	ServerName string // SNI simulado (ej: "api.cloudflare.com" o dominio corporativo)
}

// DefaultTLSProfileConfig perfil de evasión predeterminado
func DefaultTLSProfileConfig() TLSProfileConfig {
	return TLSProfileConfig{
		ServerName: "gateway.cloudflare.com",
	}
}

// TLSOptionEngine coordina la generación y validación de tramas TLS 1.3
type TLSOptionEngine struct {
	cfg TLSProfileConfig
}

// NewTLSOptionEngine crea un nuevo motor de camuflaje TLS 1.3
func NewTLSOptionEngine(cfg TLSProfileConfig) *TLSOptionEngine {
	if cfg.ServerName == "" {
		cfg.ServerName = "gateway.cloudflare.com"
	}
	return &TLSOptionEngine{cfg: cfg}
}

// BuildClientHello genera un mensaje TLS 1.3 ClientHello (RFC 8446) legítimo
// con SNI, extensiones Supported Versions (TLS 1.3) y Key Share X25519 sintético.
func (e *TLSOptionEngine) BuildClientHello() ([]byte, error) {
	// 1. Cuerpo del Handshake ClientHello
	var chBody []byte

	// Legacy Version: 0x0303
	chBody = binary.BigEndian.AppendUint16(chBody, TLSVersionLegacy)

	// Random: 32 bytes
	random := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, random); err != nil {
		return nil, fmt.Errorf("error generando entropía para TLS random: %w", err)
	}
	chBody = append(chBody, random...)

	// Legacy Session ID: 32 bytes (Middlebox compatibility mode)
	sessionID := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, sessionID); err != nil {
		return nil, fmt.Errorf("error generando session id: %w", err)
	}
	chBody = append(chBody, byte(len(sessionID)))
	chBody = append(chBody, sessionID...)

	// Cipher Suites (TLS_AES_128_GCM_SHA256 [0x13, 0x01], TLS_CHACHA20_POLY1305_SHA256 [0x13, 0x03])
	cipherSuites := []byte{0x13, 0x01, 0x13, 0x03}
	chBody = binary.BigEndian.AppendUint16(chBody, uint16(len(cipherSuites)))
	chBody = append(chBody, cipherSuites...)

	// Legacy Compression Methods: 1 byte longitud (0x01), 1 byte null (0x00)
	chBody = append(chBody, 0x01, 0x00)

	// Extensiones
	var extensions []byte

	// 1. Extension SNI (0x0000)
	sniName := []byte(e.cfg.ServerName)
	var sniExtData []byte
	// ServerNameList length
	sniExtData = binary.BigEndian.AppendUint16(sniExtData, uint16(len(sniName)+3))
	sniExtData = append(sniExtData, 0x00) // NameType: host_name (0)
	sniExtData = binary.BigEndian.AppendUint16(sniExtData, uint16(len(sniName)))
	sniExtData = append(sniExtData, sniName...)

	extensions = binary.BigEndian.AppendUint16(extensions, ExtServerName)
	extensions = binary.BigEndian.AppendUint16(extensions, uint16(len(sniExtData)))
	extensions = append(extensions, sniExtData...)

	// 2. Extension Supported Versions (0x002b) -> TLS 1.3 (0x0304)
	var suppVersionsExtData []byte
	suppVersionsExtData = append(suppVersionsExtData, 0x02) // length of list (2 bytes)
	suppVersionsExtData = binary.BigEndian.AppendUint16(suppVersionsExtData, TLSVersion13)

	extensions = binary.BigEndian.AppendUint16(extensions, ExtSupportedVersions)
	extensions = binary.BigEndian.AppendUint16(extensions, uint16(len(suppVersionsExtData)))
	extensions = append(extensions, suppVersionsExtData...)

	// 3. Extension Supported Groups (0x000a) -> X25519 (0x001d)
	var suppGroupsExtData []byte
	suppGroupsExtData = binary.BigEndian.AppendUint16(suppGroupsExtData, 0x0002) // list length
	suppGroupsExtData = binary.BigEndian.AppendUint16(suppGroupsExtData, GroupX25519)

	extensions = binary.BigEndian.AppendUint16(extensions, ExtSupportedGroups)
	extensions = binary.BigEndian.AppendUint16(extensions, uint16(len(suppGroupsExtData)))
	extensions = append(extensions, suppGroupsExtData...)

	// 4. Extension Key Share (0x0033) con clave efímera sintética
	var keyShareExtData []byte
	fakePub := make([]byte, 32)
	_, _ = io.ReadFull(rand.Reader, fakePub)
	clientShareLen := 2 + 2 + 32 // group(2) + keyLen(2) + key(32)
	keyShareExtData = binary.BigEndian.AppendUint16(keyShareExtData, uint16(clientShareLen))
	keyShareExtData = binary.BigEndian.AppendUint16(keyShareExtData, GroupX25519)
	keyShareExtData = binary.BigEndian.AppendUint16(keyShareExtData, 32)
	keyShareExtData = append(keyShareExtData, fakePub...)

	extensions = binary.BigEndian.AppendUint16(extensions, ExtKeyShare)
	extensions = binary.BigEndian.AppendUint16(extensions, uint16(len(keyShareExtData)))
	extensions = append(extensions, keyShareExtData...)

	// Anexar longitud de extensiones y bloque de extensiones
	chBody = binary.BigEndian.AppendUint16(chBody, uint16(len(extensions)))
	chBody = append(chBody, extensions...)

	// 2. Encapsular en Handshake Message (Type=0x01, Length=uint24)
	handshakeLen := len(chBody)
	handshakeMsg := make([]byte, 4+handshakeLen)
	handshakeMsg[0] = TLSHandshakeClientHello
	handshakeMsg[1] = byte(handshakeLen >> 16)
	handshakeMsg[2] = byte(handshakeLen >> 8)
	handshakeMsg[3] = byte(handshakeLen)
	copy(handshakeMsg[4:], chBody)

	// 3. Encapsular en TLS Record (Type=0x16, Version=0x0301 o 0x0303, Length=uint16)
	recordLen := len(handshakeMsg)
	record := make([]byte, TLSRecordHeaderSize+recordLen)
	record[0] = TLSRecordHandshake
	binary.BigEndian.PutUint16(record[1:3], 0x0301) // TLS 1.0 legacy record version in ClientHello
	binary.BigEndian.PutUint16(record[3:5], uint16(recordLen))
	copy(record[5:], handshakeMsg)

	return record, nil
}

// BuildServerHello genera un mensaje TLS 1.3 ServerHello sintético en respuesta
func (e *TLSOptionEngine) BuildServerHello() ([]byte, error) {
	var shBody []byte

	// Legacy Version: 0x0303
	shBody = binary.BigEndian.AppendUint16(shBody, TLSVersionLegacy)

	// Random: 32 bytes
	random := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, random); err != nil {
		return nil, err
	}
	shBody = append(shBody, random...)

	// Legacy Session ID echo (32 bytes)
	shBody = append(shBody, 32)
	fakeSession := make([]byte, 32)
	_, _ = io.ReadFull(rand.Reader, fakeSession)
	shBody = append(shBody, fakeSession...)

	// Cipher suite seleccionada (TLS_AES_128_GCM_SHA256)
	shBody = append(shBody, 0x13, 0x01)
	// Compression: null
	shBody = append(shBody, 0x00)

	// Extensions: Supported Versions (0x002b -> TLS 1.3) y KeyShare (0x0033)
	var extensions []byte
	// Supported Versions server response
	extensions = binary.BigEndian.AppendUint16(extensions, ExtSupportedVersions)
	extensions = binary.BigEndian.AppendUint16(extensions, 2)
	extensions = binary.BigEndian.AppendUint16(extensions, TLSVersion13)

	// Server Key Share
	var keyShareData []byte
	fakePub := make([]byte, 32)
	_, _ = io.ReadFull(rand.Reader, fakePub)
	keyShareData = binary.BigEndian.AppendUint16(keyShareData, GroupX25519)
	keyShareData = binary.BigEndian.AppendUint16(keyShareData, 32)
	keyShareData = append(keyShareData, fakePub...)

	extensions = binary.BigEndian.AppendUint16(extensions, ExtKeyShare)
	extensions = binary.BigEndian.AppendUint16(extensions, uint16(len(keyShareData)))
	extensions = append(extensions, keyShareData...)

	shBody = binary.BigEndian.AppendUint16(shBody, uint16(len(extensions)))
	shBody = append(shBody, extensions...)

	// Handshake record
	handshakeLen := len(shBody)
	handshakeMsg := make([]byte, 4+handshakeLen)
	handshakeMsg[0] = TLSHandshakeServerHello
	handshakeMsg[1] = byte(handshakeLen >> 16)
	handshakeMsg[2] = byte(handshakeLen >> 8)
	handshakeMsg[3] = byte(handshakeLen)
	copy(handshakeMsg[4:], shBody)

	record := make([]byte, TLSRecordHeaderSize+len(handshakeMsg))
	record[0] = TLSRecordHandshake
	binary.BigEndian.PutUint16(record[1:3], TLSVersionLegacy)
	binary.BigEndian.PutUint16(record[3:5], uint16(len(handshakeMsg)))
	copy(record[5:], handshakeMsg)

	return record, nil
}

// WrapApplicationData encapsula un payload de malla IPVN7 dentro de un registro TLS 1.3 ApplicationData
func (e *TLSOptionEngine) WrapApplicationData(payload []byte) ([]byte, error) {
	if len(payload) > TLSMaxRecordLength {
		return nil, fmt.Errorf("tamaño del paquete %d excede límite de registro TLS %d", len(payload), TLSMaxRecordLength)
	}

	frame := make([]byte, TLSRecordHeaderSize+len(payload))
	frame[0] = TLSRecordApplicationData
	binary.BigEndian.PutUint16(frame[1:3], TLSVersionLegacy)
	binary.BigEndian.PutUint16(frame[3:5], uint16(len(payload)))
	copy(frame[TLSRecordHeaderSize:], payload)

	return frame, nil
}

// UnwrapApplicationData valida y extrae el payload de red IPVN7 de un registro TLS 1.3 ApplicationData
func (e *TLSOptionEngine) UnwrapApplicationData(frame []byte) ([]byte, error) {
	if len(frame) < TLSRecordHeaderSize {
		return nil, errors.New("trama TLS truncada")
	}

	if frame[0] != TLSRecordApplicationData {
		return nil, fmt.Errorf("tipo de registro TLS inesperado: 0x%02x (se esperaba ApplicationData 0x17)", frame[0])
	}

	recordLen := int(binary.BigEndian.Uint16(frame[3:5]))
	if len(frame) < TLSRecordHeaderSize+recordLen {
		return nil, errors.New("longitud declarada en cabecera TLS mayor que bytes disponibles")
	}

	return frame[TLSRecordHeaderSize : TLSRecordHeaderSize+recordLen], nil
}

// ValidateClientHello verifica la validez estructural de un ClientHello TLS 1.3
func ValidateClientHello(data []byte) (sni string, hasTLS13 bool, err error) {
	if len(data) < TLSRecordHeaderSize+4 {
		return "", false, errors.New("datos insuficientes para cabecera TLS")
	}
	if data[0] != TLSRecordHandshake {
		return "", false, errors.New("no es un registro TLS Handshake")
	}

	handshakeType := data[5]
	if handshakeType != TLSHandshakeClientHello {
		return "", false, errors.New("no es un ClientHello")
	}

	// Parsing simplificado de extensiones
	// Saltar: Header(5) + Type(1) + Length(3) + LegacyVer(2) + Random(32)
	offset := 5 + 4 + 2 + 32
	if len(data) <= offset {
		return "", false, errors.New("ClientHello truncado")
	}

	sessionIDLen := int(data[offset])
	offset += 1 + sessionIDLen
	if len(data) <= offset+2 {
		return "", false, errors.New("ClientHello truncado en session id")
	}

	cipherSuitesLen := int(binary.BigEndian.Uint16(data[offset : offset+2]))
	offset += 2 + cipherSuitesLen
	if len(data) <= offset+1 {
		return "", false, errors.New("ClientHello truncado en cipher suites")
	}

	compMethodsLen := int(data[offset])
	offset += 1 + compMethodsLen
	if len(data) < offset+2 {
		return "", false, nil // Sin extensiones
	}

	extTotalLen := int(binary.BigEndian.Uint16(data[offset : offset+2]))
	offset += 2
	end := offset + extTotalLen
	if end > len(data) {
		end = len(data)
	}

	for offset+4 <= end {
		extType := binary.BigEndian.Uint16(data[offset : offset+2])
		extLen := int(binary.BigEndian.Uint16(data[offset+2 : offset+4]))
		offset += 4

		if offset+extLen > end {
			break
		}

		switch extType {
		case ExtSupportedVersions:
			if extLen >= 1 {
				vListLen := int(data[offset])
				for i := 1; i+2 <= extLen && i < 1+vListLen; i += 2 {
					ver := binary.BigEndian.Uint16(data[offset+i : offset+i+2])
					if ver == TLSVersion13 {
						hasTLS13 = true
					}
				}
			}
		case ExtServerName:
			if extLen > 5 {
				nameLen := int(binary.BigEndian.Uint16(data[offset+3 : offset+5]))
				if offset+5+nameLen <= end {
					sni = string(data[offset+5 : offset+5+nameLen])
				}
			}
		}

		offset += extLen
	}

	return sni, hasTLS13, nil
}
