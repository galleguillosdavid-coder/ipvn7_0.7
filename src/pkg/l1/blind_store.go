package l1

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// BlindBeaconStore abstrae el sustrato de almacenamiento de las balizas (Firebase, HTTP, memoria, etc.)
type BlindBeaconStore interface {
	PutBeacon(topic string, data []byte, ttl time.Duration) error
	GetBeacon(topic string) ([]byte, error)
	DeleteBeacon(topic string) error
}

// MultiBlindBeaconStore extiende BlindBeaconStore para obtener todas las balizas vigentes en un tópico
type MultiBlindBeaconStore interface {
	BlindBeaconStore
	GetAllBeacons(topic string) ([][]byte, error)
}

// MemoryBlindBeaconStore implementa almacenamiento en memoria para entornos locales o de prueba
type MemoryBlindBeaconStore struct {
	mu      sync.RWMutex
	beacons map[string]storedBeacon
}

type storedBeacon struct {
	data      []byte
	expiresAt time.Time
}

// NewMemoryBlindBeaconStore inicializa el almacén en memoria
func NewMemoryBlindBeaconStore() *MemoryBlindBeaconStore {
	return &MemoryBlindBeaconStore{
		beacons: make(map[string]storedBeacon),
	}
}

func (m *MemoryBlindBeaconStore) PutBeacon(topic string, data []byte, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.beacons[topic] = storedBeacon{
		data:      data,
		expiresAt: time.Now().Add(ttl),
	}
	return nil
}

func (m *MemoryBlindBeaconStore) GetBeacon(topic string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	b, ok := m.beacons[topic]
	if !ok {
		return nil, errors.New("baliza no encontrada")
	}
	if time.Now().After(b.expiresAt) {
		return nil, errors.New("baliza expirada por TTL")
	}
	return b.data, nil
}

func (m *MemoryBlindBeaconStore) GetAllBeacons(topic string) ([][]byte, error) {
	b, err := m.GetBeacon(topic)
	if err != nil {
		return nil, err
	}
	return [][]byte{b}, nil
}

func (m *MemoryBlindBeaconStore) DeleteBeacon(topic string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.beacons, topic)
	return nil
}

// HTTPBlindBeaconStore implementa señalización efímera WAN usando sustratos públicos Zero-Knowledge
type HTTPBlindBeaconStore struct {
	baseURL    string
	httpClient *http.Client
}

// NewHTTPBlindBeaconStore inicializa el sustrato HTTP para balizas EBRA públicas
func NewHTTPBlindBeaconStore(baseURL string) *HTTPBlindBeaconStore {
	if baseURL == "" {
		baseURL = "https://ntfy.sh"
	}
	return &HTTPBlindBeaconStore{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 4 * time.Second,
		},
	}
}

func (h *HTTPBlindBeaconStore) topicURL(topic string) string {
	// ntfy.sh limita los nombres de tópico a <= 64 caracteres.
	// Usamos prefijo "ip7-ebra-" y los primeros 32 caracteres hexadecimales (128 bits de entropía).
	shortTopic := topic
	if len(shortTopic) > 32 {
		shortTopic = shortTopic[:32]
	}
	return fmt.Sprintf("%s/ip7-ebra-%s", h.baseURL, shortTopic)
}

func (h *HTTPBlindBeaconStore) PutBeacon(topic string, data []byte, ttl time.Duration) error {
	url := h.topicURL(topic)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Title", "EBRA Beacon")
	req.Header.Set("Priority", "urgent")
	req.Header.Set("Tags", "satellite,key")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP error %d publicando baliza EBRA", resp.StatusCode)
	}
	return nil
}

func (h *HTTPBlindBeaconStore) GetBeacon(topic string) ([]byte, error) {
	all, err := h.GetAllBeacons(topic)
	if err != nil || len(all) == 0 {
		return nil, errors.New("baliza no encontrada en sustrato WAN")
	}
	return all[len(all)-1], nil
}

func (h *HTTPBlindBeaconStore) GetAllBeacons(topic string) ([][]byte, error) {
	url := fmt.Sprintf("%s/json?poll=1&since=10m", h.topicURL(topic))
	resp, err := h.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP error %d leyendo balizas", resp.StatusCode)
	}

	var results [][]byte
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var msg struct {
			Event   string `json:"event"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal(line, &msg); err == nil && msg.Event == "message" && msg.Message != "" {
			results = append(results, []byte(msg.Message))
		}
	}
	if len(results) == 0 {
		return nil, errors.New("no hay balizas activas en el tópico")
	}
	return results, nil
}

func (h *HTTPBlindBeaconStore) DeleteBeacon(topic string) error {
	return nil
}

// HybridBlindBeaconStore unifica la memoria local con el sustrato WAN HTTP
type HybridBlindBeaconStore struct {
	mem  *MemoryBlindBeaconStore
	http *HTTPBlindBeaconStore
}

// NewHybridBlindBeaconStore inicializa el almacén híbrido de balizas (Local + WAN)
func NewHybridBlindBeaconStore(baseURL string) *HybridBlindBeaconStore {
	return &HybridBlindBeaconStore{
		mem:  NewMemoryBlindBeaconStore(),
		http: NewHTTPBlindBeaconStore(baseURL),
	}
}

func (s *HybridBlindBeaconStore) PutBeacon(topic string, data []byte, ttl time.Duration) error {
	_ = s.mem.PutBeacon(topic, data, ttl)
	return s.http.PutBeacon(topic, data, ttl)
}

func (s *HybridBlindBeaconStore) GetBeacon(topic string) ([]byte, error) {
	if b, err := s.mem.GetBeacon(topic); err == nil {
		return b, nil
	}
	return s.http.GetBeacon(topic)
}

func (s *HybridBlindBeaconStore) GetAllBeacons(topic string) ([][]byte, error) {
	var list [][]byte
	if b, err := s.mem.GetBeacon(topic); err == nil {
		list = append(list, b)
	}
	if remoteList, err := s.http.GetAllBeacons(topic); err == nil {
		list = append(list, remoteList...)
	}
	if len(list) == 0 {
		return nil, errors.New("no hay balizas disponibles")
	}
	return list, nil
}

func (s *HybridBlindBeaconStore) DeleteBeacon(topic string) error {
	_ = s.mem.DeleteBeacon(topic)
	_ = s.http.DeleteBeacon(topic)
	return nil
}
