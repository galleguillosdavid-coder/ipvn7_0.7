// Package interfaces define los contratos raíz universales e inmutables del protocolo IPVN7 (v0.7).
package interfaces

import "fmt"

// NodeState representa el estado formal de un nodo o par en el ciclo de vida de la red.
type NodeState uint8

const (
	// NodeStateDisconnected: El nodo no cuenta con enlace físico ni señal de pares.
	NodeStateDisconnected NodeState = iota

	// NodeStateSyncing: Señal detectada; descargando metadatos y tabla de enrutamiento inicial.
	NodeStateSyncing

	// NodeStateValidating: Verificando firma criptográfica DID y contratos de gobernanza.
	NodeStateValidating

	// NodeStateActive: Nodo plenamente validado y operativo para tráfico de red y servicios.
	NodeStateActive

	// NodeStateQuarantined: Nodo confinado por anomalías de seguridad o agresión ZTNA.
	NodeStateQuarantined
)

func (s NodeState) String() string {
	switch s {
	case NodeStateDisconnected:
		return "DISCONNECTED"
	case NodeStateSyncing:
		return "SYNCING"
	case NodeStateValidating:
		return "VALIDATING"
	case NodeStateActive:
		return "ACTIVE"
	case NodeStateQuarantined:
		return "QUARANTINED"
	default:
		return fmt.Sprintf("UNKNOWN_STATE(%d)", s)
	}
}

// FSMEvent define las señales deterministas que provocan transiciones de estado.
type FSMEvent string

const (
	EventNetworkSignal  FSMEvent = "NETWORK_SIGNAL"
	EventSyncCompleted  FSMEvent = "SYNC_COMPLETED"
	EventDIDVerified    FSMEvent = "DID_VERIFIED"
	EventHeartbeatLost  FSMEvent = "HEARTBEAT_LOST"
	EventSecurityThreat FSMEvent = "SECURITY_THREAT"
	EventReset          FSMEvent = "RESET"
)

// StateTransitionHook es la firma de los callbacks ejecutados tras una transición válida.
type StateTransitionHook func(from, to NodeState, event FSMEvent, payload interface{})

// DeterministicFSM abstrae el autómata finito determinista para nodos y sesiones.
type DeterministicFSM interface {
	// CurrentState retorna el estado activo actual.
	CurrentState() NodeState

	// CanTransitionTo verifica si la transición es admisible según la matriz de estados.
	CanTransitionTo(next NodeState) bool

	// TriggerEvent procesa un evento y ejecuta la transición correspondiente.
	TriggerEvent(event FSMEvent, payload interface{}) (NodeState, error)

	// RegisterHook suscribe un observador a las transiciones de estado.
	RegisterHook(hook StateTransitionHook)
}
