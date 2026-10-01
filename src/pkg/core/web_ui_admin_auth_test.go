package core

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"ipvn7/pkg/l0"
)

// TestAdminWithoutAuth valida que una petición administrativa sin cabecera de autenticación
// sea rechazada con 401 Unauthorized cuando existen tokens de administración configurados.
func TestAdminWithoutAuth(t *testing.T) {
	id, _ := l0.GenerateIdentity()
	ui := StartWebUI(0, id, nil, nil)
	if ui == nil {
		t.Fatal("StartWebUI falló")
	}
	defer ui.Stop()

	ui.SetAdminTokens(map[string]string{
		"secret_token_123": "admin",
	})

	req := httptest.NewRequest("POST", "/api/v1/vpn/connect", nil)
	w := httptest.NewRecorder()
	ui.handleConnect(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("Esperado 401 Unauthorized sin auth, obtenido %d", w.Code)
	}
}

// TestAdminAuthenticatedUnauthorized valida que un usuario autenticado con un rol insuficiente
// (ej. 'viewer') sea rechazado con 403 Forbidden al intentar ejecutar una operación administrativa.
func TestAdminAuthenticatedUnauthorized(t *testing.T) {
	id, _ := l0.GenerateIdentity()
	ui := StartWebUI(0, id, nil, nil)
	if ui == nil {
		t.Fatal("StartWebUI falló")
	}
	defer ui.Stop()

	ui.SetAdminTokens(map[string]string{
		"admin_secret": "admin",
		"viewer_token": "viewer",
	})

	req := httptest.NewRequest("POST", "/api/v1/vpn/connect", nil)
	req.Header.Set("Authorization", "Bearer viewer_token")
	w := httptest.NewRecorder()
	ui.handleConnect(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("Esperado 403 Forbidden para rol viewer, obtenido %d", w.Code)
	}
}

// TestAdminAuthorized valida que un administrador con credenciales válidas
// pueda ejecutar la operación de control de red con éxito (200 OK).
func TestAdminAuthorized(t *testing.T) {
	id, _ := l0.GenerateIdentity()
	ui := StartWebUI(0, id, nil, nil)
	if ui == nil {
		t.Fatal("StartWebUI falló")
	}
	defer ui.Stop()

	ui.SetCallbacks(func() error { return nil }, nil, nil)
	ui.SetAdminTokens(map[string]string{
		"admin_secret": "admin",
	})

	req := httptest.NewRequest("POST", "/api/v1/vpn/connect", nil)
	req.Header.Set("Authorization", "Bearer admin_secret")
	w := httptest.NewRecorder()
	ui.handleConnect(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Esperado 200 OK para administrador autorizado, obtenido %d", w.Code)
	}
}

// TestUpdateCheck_SSRF_Rejection valida que la API de actualización rechace URLs arbitrarias externas.
func TestUpdateCheck_SSRF_Rejection(t *testing.T) {
	id, _ := l0.GenerateIdentity()
	ui := StartWebUI(0, id, nil, nil)
	if ui == nil {
		t.Fatal("StartWebUI falló")
	}
	defer ui.Stop()

	req := httptest.NewRequest("GET", "/api/v1/update/check?url=https://attacker.evil.com/manifest.json", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	w := httptest.NewRecorder()
	ui.handleUpdateCheck(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Esperado 400 Bad Request ante SSRF, obtenido %d", w.Code)
	}
}

// TestCORS_NoWildcardOnAdmin valida que las cabeceras CORS del panel de administración no usen comodín '*'.
func TestCORS_NoWildcardOnAdmin(t *testing.T) {
	id, _ := l0.GenerateIdentity()
	ui := StartWebUI(0, id, nil, nil)
	if ui == nil {
		t.Fatal("StartWebUI falló")
	}
	defer ui.Stop()

	req := httptest.NewRequest("POST", "/api/v1/vpn/connect", nil)
	w := httptest.NewRecorder()
	ui.handleConnect(w, req)

	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin == "*" {
		t.Fatalf("CORS crítico: Access-Control-Allow-Origin no debe ser '*' en panel administrativo")
	}
}
