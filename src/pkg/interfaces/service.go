// Package interfaces define los contratos raíz universales e inmutables del protocolo IPVN7 (v0.7).
package interfaces

import (
	"context"
	"net/http"
)

// ComponentCapability describe una función o recurso expuesto por un componente inteligente.
type ComponentCapability struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Type        string            `json:"type"` // "service", "device", "gateway", "ai-tool"
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Endpoints   []string          `json:"endpoints"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// ServiceMessage encapsula un mensaje de datos intercambiado entre componentes o pares remotos.
type ServiceMessage struct {
	ID          string            `json:"id"`
	SenderDID   string            `json:"sender_did"`
	TargetDID   string            `json:"target_did"`
	ServiceType string            `json:"service_type"` // ej: "chat", "spooler", "iot_control"
	Payload     []byte            `json:"payload"`
	Headers     map[string]string `json:"headers,omitempty"`
}

// SmartComponent define el ciclo de vida y contrato de montaje para extensiones del núcleo.
// Componentes como Chat P2P, Spooler IPP de Impresión o el IoT Device Bridge implementan esta interfaz.
type SmartComponent interface {
	// ComponentID retorna el identificador único del componente (ej: 'spooler-ipp', 'p2p-chat').
	ComponentID() string

	// Capabilities retorna el catálogo de facultades que el componente registra en el Gateway.
	Capabilities() []ComponentCapability

	// HandleServiceMessage procesa un mensaje entrante dirigido a este componente.
	HandleServiceMessage(ctx context.Context, msg *ServiceMessage) ([]byte, error)

	// RegisterHTTPRoutes permite al componente montar endpoints HTTP REST o WebSocket en el servidor.
	RegisterHTTPRoutes(mux *http.ServeMux)

	// Start inicializa los bucles de eventos y recursos en segundo plano del componente.
	Start(ctx context.Context) error

	// Stop detiene de forma limpia los trabajadores y libera recursos.
	Stop() error
}

// ServiceBus administra la suscripción y publicación de eventos internos de la malla y componentes.
type ServiceBus interface {
	// PublishEvent emite un evento a todos los componentes o consumidores suscritos.
	PublishEvent(eventType string, data interface{})

	// Subscribe suscribe un manejador a una categoría de eventos de la malla.
	Subscribe(eventType string, handler func(data interface{})) func()
}
