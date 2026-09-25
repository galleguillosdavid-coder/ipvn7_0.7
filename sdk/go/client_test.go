package ipvn7sdk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSDKClient_GetIdentity(t *testing.T) {
	mockID := NodeIdentity{
		DID:  "did:ipvn7:testnode1234567890",
		IPv6: "fd00:7::1",
		IPv4: "10.7.0.1",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/identity" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(mockID)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	id, err := client.GetIdentity(context.Background())
	if err != nil {
		t.Fatalf("error obteniendo identidad: %v", err)
	}

	if id.DID != mockID.DID {
		t.Errorf("DID esperado %s, obtenido %s", mockID.DID, id.DID)
	}
}

func TestSDKClient_SendChatMessage(t *testing.T) {
	messageSent := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/chat/send" {
			http.NotFound(w, r)
			return
		}
		messageSent = true
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.SendChatMessage(context.Background(), "did:ipvn7:destination", "Hola Malla Soberana")
	if err != nil {
		t.Fatalf("error enviando mensaje: %v", err)
	}

	if !messageSent {
		t.Errorf("el mensaje no fue recibido por el mock server")
	}
}
