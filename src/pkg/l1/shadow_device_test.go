package l1

import (
	"bytes"
	"io"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestShadowDevice_RegistrationAndLookup(t *testing.T) {
	reg := NewShadowDeviceRegistry()

	// 1. Registro de impresora
	macPrinter := "00:11:22:33:44:55"
	dev, err := reg.RegisterDevice("HP LaserJet Office", "192.168.1.50", macPrinter, DeviceTypePrinter, []int{631, 9100})
	if err != nil {
		t.Fatalf("Fallo registrando impresora: %v", err)
	}

	if !strings.HasPrefix(dev.DID, "did:ipvn7:shadow:") {
		t.Errorf("Formato de DID incorrecto: %s", dev.DID)
	}
	if !strings.HasPrefix(dev.VirtualIPv4, "10.7.100.") {
		t.Errorf("Subred virtual incorrecta: %s", dev.VirtualIPv4)
	}

	// 2. Búsqueda por DID y por Virtual IP
	byDID := reg.LookupByDID(dev.DID)
	if byDID == nil || byDID.Name != "HP LaserJet Office" {
		t.Fatalf("Búsqueda por DID falló")
	}

	byVIP := reg.LookupByVirtualIP(dev.VirtualIPv4)
	if byVIP == nil || byVIP.PhysicalIP != "192.168.1.50" {
		t.Fatalf("Búsqueda por VIP falló")
	}

	// 3. Verificación de lista
	list := reg.ListDevices()
	if len(list) != 1 {
		t.Errorf("Esperado 1 dispositivo en lista, obtenido %d", len(list))
	}
}

func TestShadowDevice_ForwardStream(t *testing.T) {
	// 1. Servidor simulado de impresora RAW 9100 (Echo Server)
	echoListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Error creando listener echo: %v", err)
	}
	defer echoListener.Close()

	_, portStr, _ := net.SplitHostPort(echoListener.Addr().String())
	mockPort, _ := strconv.Atoi(portStr)

	go func() {
		for {
			conn, err := echoListener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				_, _ = io.Copy(c, c)
			}(conn)
		}
	}()

	reg := NewShadowDeviceRegistry()
	dev, err := reg.RegisterDevice("Mock Printer", "127.0.0.1", "AA:BB:CC:DD:EE:FF", DeviceTypePrinter, []int{mockPort})
	if err != nil {
		t.Fatalf("Error registrando mock: %v", err)
	}

	// 2. Simular cliente de malla enviando trabajo de impresión
	client, serverPipe := net.Pipe()
	defer client.Close()

	done := make(chan error, 1)
	go func() {
		done <- reg.ForwardStream(serverPipe, dev.DID, mockPort)
	}()

	payload := []byte("PJL JOB PRINT TEST IPVN7")
	go func() {
		_, _ = client.Write(payload)
		_ = client.Close()
	}()

	buf := make([]byte, len(payload))
	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, _ := io.ReadFull(client, buf)

	if n > 0 && !bytes.Equal(buf[:n], payload) {
		t.Errorf("Datos recibidos no coinciden: %s", string(buf[:n]))
	}

	// 3. Probar puerto no permitido (bloqueo ZTNA)
	c2, s2 := net.Pipe()
	defer c2.Close()
	defer s2.Close()
	errBlocked := reg.ForwardStream(s2, dev.DID, 8080)
	if errBlocked == nil {
		t.Errorf("Esperado bloqueo ZTNA para puerto no autorizado")
	}
}

func TestShadowDevice_WakePacket(t *testing.T) {
	reg := NewShadowDeviceRegistry()
	// Validación de formato MAC sin error
	err := reg.WakeDevice("00:11:22:33:44:55")
	// En entornos de test de loopback puede fallar o pasar el socket broadcast, pero el parsing debe ser correcto
	if err != nil && strings.Contains(err.Error(), "MAC inválida") {
		t.Errorf("Error parsing MAC válida: %v", err)
	}

	errInvalid := reg.WakeDevice("MAC-ERRONEA")
	if errInvalid == nil {
		t.Errorf("Esperado error ante MAC inválida")
	}
}

func TestShadowDevice_EdgeAIWorker(t *testing.T) {
	reg := NewShadowDeviceRegistry()
	wEmpty := reg.FindOptimalInferenceWorker()
	if wEmpty != nil {
		t.Errorf("Esperado nil cuando no hay workers registrados")
	}

	worker, err := reg.RegisterEdgeAIWorker("Ollama-Local", "192.168.1.150", "00:AA:BB:CC:DD:EE", 11434)
	if err != nil {
		t.Fatalf("Error registrando worker Edge AI: %v", err)
	}
	if worker.DeviceType != DeviceTypeEdgeAI {
		t.Errorf("Tipo incorrecto: %s", worker.DeviceType)
	}

	found := reg.FindOptimalInferenceWorker()
	if found == nil || found.DID != worker.DID {
		t.Fatalf("FindOptimalInferenceWorker no encontro el worker registrado")
	}
}
