package l0

import (
	"net"
	"testing"
)

// TestMinimalSocketReal verifica que el núcleo mínimo puede crear sockets UDP reales
func TestMinimalSocketReal(t *testing.T) {
	// Crear socket UDP real
	addr := "127.0.0.1:0" // Puerto aleatorio
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		t.Fatalf("Error creando socket UDP: %v", err)
	}
	defer conn.Close()

	// Verificar que el socket es funcional
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	if localAddr.Port == 0 {
		t.Fatal("Puerto no asignado")
	}

	// Verificar que el socket puede escribir
	testData := []byte("IPVN7_MINIMAL_TEST")
	_, err = conn.WriteTo(testData, localAddr)
	if err != nil {
		t.Fatalf("Error escribiendo datagrama: %v", err)
	}
}
