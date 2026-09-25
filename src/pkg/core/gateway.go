package core

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"ipvn7/pkg/l0"
	"ipvn7/pkg/l1"
	"ipvn7/pkg/l2"
)

var (
	ErrComponentNotFound = errors.New("component not found")
	ErrDuplicateComponent = errors.New("component already registered")
	ErrBusClosed          = errors.New("component bus is closed")
)

// SmartComponentGateway gestiona el acoplamiento y comunicación de componentes externos
type SmartComponentGateway struct {
	mu          sync.RWMutex
	identity    *l0.Identity
	router      *l1.KleinbergRouter
	telemetry   *l2.TelemetryRingBuffer
	firewall    *l1.ZTNAFirewall
	healing     *l1.LinkHealingEngine
	components  map[string]*ComponentRegistration
	subscribers map[string]chan MeshEvent
	inboundMsg  chan *DatagramEnvelope
	ctx         context.Context
	cancel      context.CancelFunc
	closed      bool
}

// NewSmartComponentGateway inicializa el gateway del núcleo
func NewSmartComponentGateway(id *l0.Identity, router *l1.KleinbergRouter, telemetry *l2.TelemetryRingBuffer, firewall *l1.ZTNAFirewall) *SmartComponentGateway {
	ctx, cancel := context.WithCancel(context.Background())
	gw := &SmartComponentGateway{
		identity:    id,
		router:      router,
		telemetry:   telemetry,
		firewall:    firewall,
		components:  make(map[string]*ComponentRegistration),
		subscribers: make(map[string]chan MeshEvent),
		inboundMsg:  make(chan *DatagramEnvelope, 1024),
		ctx:         ctx,
		cancel:      cancel,
	}
	go gw.eventLoop()
	return gw
}

func (gw *SmartComponentGateway) SetHealingEngine(h *l1.LinkHealingEngine) {
	gw.mu.Lock()
	defer gw.mu.Unlock()
	gw.healing = h
}

func (gw *SmartComponentGateway) Register(id, name, version, transport string, capabilities []string) error {
	return gw.RegisterComponent(&ComponentRegistration{ID: id, Name: name, Version: version, Transport: transport, Capabilities: capabilities})
}

func (gw *SmartComponentGateway) PublishSimpleEvent(source, eventType string, payload map[string]interface{}) {
	gw.PublishEvent(MeshEvent{Type: EventType(eventType), Timestamp: time.Now().UnixNano(), Source: source, Payload: payload})
}

// RegisterComponent registra un nuevo componente en el bus inteligente
func (gw *SmartComponentGateway) RegisterComponent(comp *ComponentRegistration) error {
	gw.mu.Lock()
	defer gw.mu.Unlock()

	if gw.closed {
		return ErrBusClosed
	}

	if _, exists := gw.components[comp.ID]; exists {
		return ErrDuplicateComponent
	}

	comp.RegisteredAt = time.Now()
	comp.LastSeenAt = comp.RegisteredAt
	comp.State = StateActive
	comp.Installed = true
	comp.Enabled = true
	gw.components[comp.ID] = comp

	// Emitir evento de componente acoplado
	gw.broadcastEvent(MeshEvent{
		Type:      EventComponentBound,
		Timestamp: time.Now().UnixNano(),
		Source:    comp.ID,
		Payload: map[string]interface{}{
			"component_id": comp.ID,
			"name":         comp.Name,
			"capabilities": comp.Capabilities,
			"transport":    comp.Transport,
			"installed":    comp.Installed,
			"enabled":      comp.Enabled,
		},
	})

	return nil
}

// SetComponentInstalled activa o desactiva la instalación de un plugin
func (gw *SmartComponentGateway) SetComponentInstalled(compID string, installed bool) error {
	gw.mu.Lock()
	defer gw.mu.Unlock()

	comp, exists := gw.components[compID]
	if !exists {
		return ErrComponentNotFound
	}
	comp.Installed = installed
	if !installed {
		comp.Enabled = false
		comp.State = StateSuspended
	} else if comp.Enabled {
		comp.State = StateActive
	}
	return nil
}

// SetComponentEnabled enciende o apaga la ejecución de un plugin
func (gw *SmartComponentGateway) SetComponentEnabled(compID string, enabled bool) error {
	gw.mu.Lock()
	defer gw.mu.Unlock()

	comp, exists := gw.components[compID]
	if !exists {
		return ErrComponentNotFound
	}
	if !comp.Installed && enabled {
		comp.Installed = true
	}
	comp.Enabled = enabled
	if enabled {
		comp.State = StateActive
	} else {
		comp.State = StateSuspended
	}
	return nil
}

// UnregisterComponent desacopla un componente
func (gw *SmartComponentGateway) UnregisterComponent(compID string) error {
	gw.mu.Lock()
	defer gw.mu.Unlock()

	comp, exists := gw.components[compID]
	if !exists {
		return ErrComponentNotFound
	}

	delete(gw.components, compID)

	gw.broadcastEvent(MeshEvent{
		Type:      EventComponentUnbound,
		Timestamp: time.Now().UnixNano(),
		Source:    compID,
		Payload: map[string]interface{}{
			"component_id": compID,
			"name":         comp.Name,
		},
	})

	return nil
}

// Heartbeat renueva la presencia viva de un componente
func (gw *SmartComponentGateway) Heartbeat(compID string) error {
	gw.mu.Lock()
	defer gw.mu.Unlock()

	comp, exists := gw.components[compID]
	if !exists {
		return ErrComponentNotFound
	}

	comp.LastSeenAt = time.Now()
	comp.State = StateActive
	return nil
}

// ListComponents retorna copia de los componentes registrados
func (gw *SmartComponentGateway) ListComponents() []*ComponentRegistration {
	gw.mu.RLock()
	defer gw.mu.RUnlock()

	result := make([]*ComponentRegistration, 0, len(gw.components))
	for _, comp := range gw.components {
		copyComp := *comp
		result = append(result, &copyComp)
	}
	return result
}

// SubscribeEvents crea un canal de suscripción para eventos de la malla
func (gw *SmartComponentGateway) SubscribeEvents(subscriberID string) <-chan MeshEvent {
	gw.mu.Lock()
	defer gw.mu.Unlock()

	ch := make(chan MeshEvent, 256)
	gw.subscribers[subscriberID] = ch
	return ch
}

// UnsubscribeEvents elimina una suscripción
func (gw *SmartComponentGateway) UnsubscribeEvents(subscriberID string) {
	gw.mu.Lock()
	defer gw.mu.Unlock()

	if ch, ok := gw.subscribers[subscriberID]; ok {
		close(ch)
		delete(gw.subscribers, subscriberID)
	}
}

// PublishEvent permite a componentes o capas internas emitir un evento
func (gw *SmartComponentGateway) PublishEvent(ev MeshEvent) {
	gw.mu.Lock()
	defer gw.mu.Unlock()
	gw.broadcastEvent(ev)
}

func (gw *SmartComponentGateway) broadcastEvent(ev MeshEvent) {
	for id, ch := range gw.subscribers {
		select {
		case ch <- ev:
		default:
			// Si el suscriptor está lleno, descartamos sin bloquear el núcleo
			_ = id
		}
	}
}

// SendDatagram procesa un paquete saliente emitido por un componente externo
func (gw *SmartComponentGateway) SendDatagram(env *DatagramEnvelope) error {
	gw.mu.RLock()
	defer gw.mu.RUnlock()

	if gw.closed {
		return ErrBusClosed
	}

	if env.SourceDID == "" {
		env.SourceDID = gw.identity.DID()
	}

	// Manejo asertivo de entrega local (Loopback / Echo / Broadcast local)
	if env.TargetDID == gw.identity.DID() || env.Protocol == "mesh:echo" ||
		env.TargetDID == "did:ipvn7:broadcast" || strings.HasPrefix(env.TargetDID, "did:ipvn7:local:") {
		gw.telemetry.RecordEvent(l2.EventTxPacket, uint32(len(env.Payload)), 50, 0)
		gw.telemetry.RecordEvent(l2.EventRxPacket, uint32(len(env.Payload)), 50, 0)

		gw.broadcastEvent(MeshEvent{
			Type:      EventPacketTx,
			Timestamp: time.Now().UnixNano(),
			Source:    env.SourceDID,
			Payload: map[string]interface{}{
				"target_did": env.TargetDID,
				"bytes":      len(env.Payload),
				"protocol":   env.Protocol,
				"via_peer":   "local_loopback",
				"loopback":   true,
			},
		})
		return nil
	}

	// Validar con cortafuegos ZTNA
	if gw.firewall != nil {
		decision, _ := gw.firewall.EvaluateOutbound(env.TargetDID, 7777)
		if decision != l1.DecisionAccept {
			gw.telemetry.RecordEvent(l2.EventDrop, uint32(len(env.Payload)), 0, 0)
			return fmt.Errorf("bloqueado por cortafuegos ZTNA: destino %s no autorizado", env.TargetDID)
		}
	}

	// Seleccionar ruta sobre los 12 anillos de Kleinberg
	nextHop, err := gw.router.FindNextHop(env.TargetDID)
	if err != nil {
		return fmt.Errorf("no existe ruta en la malla para el destino %s: %w", env.TargetDID, err)
	}

	selectedDID := nextHop.DID
	isFailover := false

	// Evaluar centinela de auto-reparación ante degradación
	if gw.healing != nil {
		if stat, ok := gw.healing.GetPeerStats(selectedDID); ok && stat.Degraded {
			_, altDID := gw.healing.RecordProbeResult(selectedDID, false, stat.LastLatencyMs)
			if altDID != "" {
				selectedDID = altDID
				isFailover = true
			}
		}
	}

	// Registrar en telemetría
	gw.telemetry.RecordEvent(l2.EventTxPacket, uint32(len(env.Payload)), 1100, 0)

	// Emitir evento de paquete transmitido
	gw.broadcastEvent(MeshEvent{
		Type:      EventPacketTx,
		Timestamp: time.Now().UnixNano(),
		Source:    env.SourceDID,
		Payload: map[string]interface{}{
			"target_did":     env.TargetDID,
			"bytes":          len(env.Payload),
			"protocol":       env.Protocol,
			"via_peer":       selectedDID,
			"failover_route": isFailover,
		},
	})

	return nil
}

// eventLoop monitorea la salud de componentes y purga los inactivos
func (gw *SmartComponentGateway) eventLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-gw.ctx.Done():
			return
		case <-ticker.C:
			gw.checkComponentLiveness()
		}
	}
}

func (gw *SmartComponentGateway) checkComponentLiveness() {
	gw.mu.Lock()
	defer gw.mu.Unlock()

	now := time.Now()
	for _, comp := range gw.components {
		// Si un componente no reporta actividad en 30s se marca IDLE
		if now.Sub(comp.LastSeenAt) > 30*time.Second && comp.State == StateActive {
			comp.State = StateIdle
		}
	}
}

// Stop cierra limpiamente el gateway
func (gw *SmartComponentGateway) Stop() {
	gw.mu.Lock()
	defer gw.mu.Unlock()

	if gw.closed {
		return
	}

	gw.closed = true
	gw.cancel()

	for id, ch := range gw.subscribers {
		close(ch)
		delete(gw.subscribers, id)
	}
}
