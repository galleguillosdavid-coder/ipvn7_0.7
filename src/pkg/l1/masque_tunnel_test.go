package l1

import (
	"bytes"
	"strings"
	"testing"
)

func TestMASQUE_BuildConnectUDPRequest(t *testing.T) {
	req := BuildConnectUDPRequest("192.168.1.106", 7001, "gateway.ipvn7.org")
	if !strings.Contains(req, "CONNECT-UDP /well-known/masque/udp/192.168.1.106/7001/ HTTP/1.1") {
		t.Fatalf("Petición CONNECT-UDP no contiene URI template canónica: %s", req)
	}
	if !strings.Contains(req, "Capsule-Protocol: ?1") {
		t.Fatalf("Encabezado Capsule-Protocol ausente en petición: %s", req)
	}
}

func TestMASQUE_VarintEncodeDecode(t *testing.T) {
	testCases := []uint64{0, 25, 63, 64, 1500, 16383, 16384, 100000, 1073741823, 1073741824}
	for _, tc := range testCases {
		enc := EncodeVarintRFC9000(tc)
		dec, n, err := DecodeVarintRFC9000(enc)
		if err != nil {
			t.Fatalf("Error decodificando varint %d: %v", tc, err)
		}
		if dec != tc {
			t.Fatalf("Desajuste en varint: esperado %d, obtenido %d", tc, dec)
		}
		if n != len(enc) {
			t.Fatalf("Longitud consumida incorrecta: esperado %d, obtenido %d", len(enc), n)
		}
	}
}

func TestMASQUE_WrapAndUnwrapDatagramCapsule(t *testing.T) {
	payload := make([]byte, 1280)
	for i := range payload {
		payload[i] = byte(i % 256)
	}

	capsule := WrapDatagramCapsule(payload)
	unwrapped, err := UnwrapDatagramCapsule(capsule)
	if err != nil {
		t.Fatalf("Error desencapsulando cápsula MASQUE: %v", err)
	}
	if !bytes.Equal(payload, unwrapped) {
		t.Fatalf("Carga útil desencapsulada no coincide con el datagrama original")
	}

	// Cápsula truncada
	_, err = UnwrapDatagramCapsule([]byte{})
	if err == nil {
		t.Fatalf("UnwrapDatagramCapsule debió fallar con buffer vacío")
	}

	// Context ID inválido (ej. 0x05)
	badCapsule := []byte{0x05, 0x01, 0x02, 0x03}
	_, err = UnwrapDatagramCapsule(badCapsule)
	if err == nil {
		t.Fatalf("UnwrapDatagramCapsule debió fallar con Context ID distinto a 0")
	}
}

func TestMASQUETunnelSession_Metrics(t *testing.T) {
	cfg := MASQUETunnelConfig{
		TargetHost:      "192.168.1.106",
		TargetPort:      7001,
		MASQUEProxyAddr: "192.168.1.198:443",
		UseCapsuleProto: true,
	}
	sess := NewMASQUETunnelSession(cfg)

	frame := []byte("Datagrama de prueba MASQUE RFC 9298")
	enc := sess.SendDatagram(frame)
	rec, err := sess.ReceiveDatagram(enc)
	if err != nil {
		t.Fatalf("ReceiveDatagram falló: %v", err)
	}
	if !bytes.Equal(frame, rec) {
		t.Fatalf("Trama recibida en sesión no coincide")
	}

	sBytes, rBytes, sFrames, rFrames := sess.GetStats()
	if sFrames != 1 || rFrames != 1 || sBytes == 0 || rBytes == 0 {
		t.Fatalf("Telemetría de sesión MASQUE inconsistente: sent=%d, recv=%d", sFrames, rFrames)
	}
}
