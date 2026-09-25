package docker_proxy

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSovereignDockerProxy_LifecycleAndRouting(t *testing.T) {
	// 1. Iniciar un backend HTTP de prueba simulando un contenedor web
	backendExpectedResponse := "Hello from Sovereign Container DID"
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, backendExpectedResponse)
	}))
	defer backendServer.Close()

	// 2. Inicializar el proxy soberano
	proxy := NewSovereignDockerProxy("127.0.0.1:0")
	ln, err := proxy.Start()
	if err != nil {
		t.Fatalf("fallo al arrancar proxy: %v", err)
	}
	defer ln.Close()
	defer proxy.Close()

	// 3. Registrar ruta DID hacia el backend
	testDID := "did:ipvn7:my-container-app"
	if err := proxy.RegisterRoute(testDID, backendServer.URL, "cont-123456"); err != nil {
		t.Fatalf("error registrando ruta: %v", err)
	}

	// 4. Test Health Check
	healthResp, err := http.Get(fmt.Sprintf("http://%s/healthz", proxy.Addr()))
	if err != nil || healthResp.StatusCode != http.StatusOK {
		t.Fatalf("health check fallo: resp=%v, err=%v", healthResp, err)
	}
	healthResp.Body.Close()

	// 5. Test Petición con cabecera X-IPVN7-Target-DID
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/", proxy.Addr()), nil)
	if err != nil {
		t.Fatalf("error creando request: %v", err)
	}
	req.Header.Set("X-IPVN7-Target-DID", testDID)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("error ejecutando petición proxy: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("error leyendo respuesta: %v", err)
	}

	if string(body) != backendExpectedResponse {
		t.Errorf("respuesta inesperada: got %q, want %q", string(body), backendExpectedResponse)
	}

	// 6. Verificar telemetría
	requests, routes := proxy.Stats()
	if requests != 1 || routes != 1 {
		t.Errorf("telemetría incorrecta: requests=%d, routes=%d", requests, routes)
	}

	// 7. Test de DID no registrado -> 502 Bad Gateway
	badReq, _ := http.NewRequest("GET", fmt.Sprintf("http://%s/", proxy.Addr()), nil)
	badReq.Header.Set("X-IPVN7-Target-DID", "did:ipvn7:unknown")
	badResp, err := client.Do(badReq)
	if err != nil {
		t.Fatalf("error en petición desconocida: %v", err)
	}
	badResp.Body.Close()
	if badResp.StatusCode != http.StatusBadGateway {
		t.Errorf("se esperaba 502 Bad Gateway, se obtuvo: %d", badResp.StatusCode)
	}
}
