package l1

import (
	"bytes"
	"testing"
)

func TestTLSOptionEngine_ClientHelloGeneration(t *testing.T) {
	cfg := TLSProfileConfig{
		ServerName: "corporate.tunnel.cloudflare.com",
	}
	engine := NewTLSOptionEngine(cfg)

	chRecord, err := engine.BuildClientHello()
	if err != nil {
		t.Fatalf("error construyendo ClientHello: %v", err)
	}

	if len(chRecord) < 50 {
		t.Fatalf("registro ClientHello sospechosamente corto: %d bytes", len(chRecord))
	}

	// Verificar byte mágico de cabecera TLS Handshake
	if chRecord[0] != TLSRecordHandshake {
		t.Errorf("byte de cabecera esperado 0x16, obtenido 0x%02x", chRecord[0])
	}

	// Validar con el parser
	sni, hasTLS13, err := ValidateClientHello(chRecord)
	if err != nil {
		t.Fatalf("error validando ClientHello sintético: %v", err)
	}

	if sni != cfg.ServerName {
		t.Errorf("SNI esperado '%s', obtenido '%s'", cfg.ServerName, sni)
	}

	if !hasTLS13 {
		t.Errorf("la extensión Supported Versions debe incluir TLS 1.3 (0x0304)")
	}
}

func TestTLSOptionEngine_ServerHelloGeneration(t *testing.T) {
	engine := NewTLSOptionEngine(DefaultTLSProfileConfig())

	shRecord, err := engine.BuildServerHello()
	if err != nil {
		t.Fatalf("error construyendo ServerHello: %v", err)
	}

	if chType := shRecord[0]; chType != TLSRecordHandshake {
		t.Fatalf("tipo de registro esperado Handshake 0x16, obtenido 0x%02x", chType)
	}

	if handshakeType := shRecord[5]; handshakeType != TLSHandshakeServerHello {
		t.Fatalf("tipo de handshake esperado ServerHello 0x02, obtenido 0x%02x", handshakeType)
	}
}

func TestTLSOptionEngine_ApplicationDataWrapUnwrap(t *testing.T) {
	engine := NewTLSOptionEngine(DefaultTLSProfileConfig())

	rawPayload := []byte("ESTE_ES_UN_PAQUETE_CUANTICO_IPVN7_CON_SPHINX_ENCRIPTADO_1280_BYTES")
	wrapped, err := engine.WrapApplicationData(rawPayload)
	if err != nil {
		t.Fatalf("error envolviendo ApplicationData: %v", err)
	}

	if wrapped[0] != TLSRecordApplicationData {
		t.Errorf("tipo esperado 0x17 (ApplicationData), obtenido 0x%02x", wrapped[0])
	}

	unwrapped, err := engine.UnwrapApplicationData(wrapped)
	if err != nil {
		t.Fatalf("error desenvolviendo ApplicationData: %v", err)
	}

	if !bytes.Equal(unwrapped, rawPayload) {
		t.Errorf("payload desenvuelto no coincide con el original")
	}
}

func BenchmarkTLSOptionEngine_WrapUnwrap(b *testing.B) {
	engine := NewTLSOptionEngine(DefaultTLSProfileConfig())
	payload := make([]byte, 1280) // MTU típica de paquete de malla
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		wrapped, err := engine.WrapApplicationData(payload)
		if err != nil {
			b.Fatal(err)
		}
		_, err = engine.UnwrapApplicationData(wrapped)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTLSOptionEngine_BuildClientHello(b *testing.B) {
	engine := NewTLSOptionEngine(DefaultTLSProfileConfig())
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := engine.BuildClientHello()
		if err != nil {
			b.Fatal(err)
		}
	}
}
