// Package ipvn7sdk proporciona el SDK cliente oficial en Go / WebAssembly
// para conectarse a nodos de la malla cuántica soberana IPVN7.
package ipvn7sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// NodeIdentity información soberana del nodo
type NodeIdentity struct {
	DID  string `json:"did"`
	IPv6 string `json:"ipv6"`
	IPv4 string `json:"ipv4"`
}

// PeerInfo información de un par en la malla
type PeerInfo struct {
	DID        string    `json:"did"`
	Address    string    `json:"address"`
	Ring       int       `json:"ring"`
	RTTMs      float64   `json:"rtt_ms"`
	LastSeenAt time.Time `json:"last_seen_at"`
}

// TelemetrySnapshot métricas del nodo
type TelemetrySnapshot struct {
	UptimeSeconds uint64 `json:"uptime_seconds"`
	RxPackets     uint64 `json:"rx_packets"`
	TxPackets     uint64 `json:"tx_packets"`
	ActivePeers   int    `json:"active_peers"`
}

// Client cliente HTTP/REST para el plano de control de IPVN7
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient crea una nueva instancia del SDK cliente
func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = "http://127.0.0.1:7070"
	}
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetIdentity consulta la identidad soberana del nodo local
func (c *Client) GetIdentity(ctx context.Context) (*NodeIdentity, error) {
	url := fmt.Sprintf("%s/api/v1/identity", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error conectando a nodo IPVN7: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error nodo HTTP %d: %s", resp.StatusCode, string(body))
	}

	var identity NodeIdentity
	if err := json.NewDecoder(resp.Body).Decode(&identity); err != nil {
		return nil, fmt.Errorf("error deserializando respuesta de identidad: %w", err)
	}

	return &identity, nil
}

// ListPeers consulta la tabla de pares activos
func (c *Client) ListPeers(ctx context.Context) ([]PeerInfo, error) {
	url := fmt.Sprintf("%s/api/v1/peers", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error HTTP %d", resp.StatusCode)
	}

	var peers []PeerInfo
	if err := json.NewDecoder(resp.Body).Decode(&peers); err != nil {
		return nil, err
	}

	return peers, nil
}

// SendChatMessage envía un mensaje de texto sobre la malla hacia un DID destino
func (c *Client) SendChatMessage(ctx context.Context, destDID, content string) error {
	if destDID == "" || content == "" {
		return errors.New("destDID y content son obligatorios")
	}

	url := fmt.Sprintf("%s/api/v1/chat/send", c.baseURL)
	payload := map[string]string{
		"recipient_did": destDID,
		"content":       content,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("falla al enviar mensaje HTTP %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetTelemetry consulta las métricas de telemetría del nodo
func (c *Client) GetTelemetry(ctx context.Context) (*TelemetrySnapshot, error) {
	url := fmt.Sprintf("%s/api/v1/telemetry", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error HTTP %d", resp.StatusCode)
	}

	var telemetry TelemetrySnapshot
	if err := json.NewDecoder(resp.Body).Decode(&telemetry); err != nil {
		return nil, err
	}

	return &telemetry, nil
}
