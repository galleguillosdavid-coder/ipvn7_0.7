package core

import (
	"fmt"
	"sync"

	"ipvn7/pkg/interfaces"
)

// DeterministicNodeFSM implementa el autómata determinista de estados finitos para nodos y pares.
type DeterministicNodeFSM struct {
	mu           sync.RWMutex
	currentState interfaces.NodeState
	hooks        []interfaces.StateTransitionHook
}

// NewDeterministicNodeFSM crea una nueva FSM inicializada en estado Disconnected.
func NewDeterministicNodeFSM(initialState interfaces.NodeState) *DeterministicNodeFSM {
	return &DeterministicNodeFSM{
		currentState: initialState,
		hooks:        make([]interfaces.StateTransitionHook, 0),
	}
}

// CurrentState retorna el estado actual del autómata de forma thread-safe.
func (fsm *DeterministicNodeFSM) CurrentState() interfaces.NodeState {
	fsm.mu.RLock()
	defer fsm.mu.RUnlock()
	return fsm.currentState
}

// CanTransitionTo verifica si la transición hacia next es permitida desde el estado actual.
func (fsm *DeterministicNodeFSM) CanTransitionTo(next interfaces.NodeState) bool {
	fsm.mu.RLock()
	defer fsm.mu.RUnlock()
	return fsm.isTransitionAllowed(fsm.currentState, next)
}

func (fsm *DeterministicNodeFSM) isTransitionAllowed(from, to interfaces.NodeState) bool {
	if from == to {
		return true
	}
	switch from {
	case interfaces.NodeStateDisconnected:
		return to == interfaces.NodeStateSyncing
	case interfaces.NodeStateSyncing:
		return to == interfaces.NodeStateValidating ||
			to == interfaces.NodeStateDisconnected ||
			to == interfaces.NodeStateQuarantined
	case interfaces.NodeStateValidating:
		return to == interfaces.NodeStateActive ||
			to == interfaces.NodeStateDisconnected ||
			to == interfaces.NodeStateQuarantined
	case interfaces.NodeStateActive:
		return to == interfaces.NodeStateSyncing ||
			to == interfaces.NodeStateDisconnected ||
			to == interfaces.NodeStateQuarantined
	case interfaces.NodeStateQuarantined:
		return to == interfaces.NodeStateDisconnected
	default:
		return false
	}
}

// TriggerEvent evalúa el evento recibido y ejecuta la transición correspondiente si es válida.
func (fsm *DeterministicNodeFSM) TriggerEvent(event interfaces.FSMEvent, payload interface{}) (interfaces.NodeState, error) {
	fsm.mu.Lock()
	defer fsm.mu.Unlock()

	nextState, err := fsm.resolveNextState(fsm.currentState, event)
	if err != nil {
		return fsm.currentState, err
	}

	prevState := fsm.currentState
	fsm.currentState = nextState

	// Ejecución de hooks de transición
	for _, hook := range fsm.hooks {
		hook(prevState, nextState, event, payload)
	}

	return nextState, nil
}

func (fsm *DeterministicNodeFSM) resolveNextState(current interfaces.NodeState, event interfaces.FSMEvent) (interfaces.NodeState, error) {
	if event == interfaces.EventReset {
		return interfaces.NodeStateDisconnected, nil
	}
	if event == interfaces.EventSecurityThreat {
		return interfaces.NodeStateQuarantined, nil
	}

	switch current {
	case interfaces.NodeStateDisconnected:
		if event == interfaces.EventNetworkSignal {
			return interfaces.NodeStateSyncing, nil
		}
	case interfaces.NodeStateSyncing:
		if event == interfaces.EventSyncCompleted {
			return interfaces.NodeStateValidating, nil
		}
		if event == interfaces.EventHeartbeatLost {
			return interfaces.NodeStateDisconnected, nil
		}
	case interfaces.NodeStateValidating:
		if event == interfaces.EventDIDVerified {
			return interfaces.NodeStateActive, nil
		}
		if event == interfaces.EventHeartbeatLost {
			return interfaces.NodeStateDisconnected, nil
		}
	case interfaces.NodeStateActive:
		if event == interfaces.EventHeartbeatLost {
			return interfaces.NodeStateDisconnected, nil
		}
		if event == interfaces.EventNetworkSignal {
			return interfaces.NodeStateSyncing, nil
		}
	case interfaces.NodeStateQuarantined:
		// El estado confinado no transiciona salvo con EventReset explícito
		return current, fmt.Errorf("nodo en cuarentena: evento %s denegado", event)
	}

	return current, fmt.Errorf("transición inválida: evento '%s' no aplicable en estado '%s'", event, current)
}

// RegisterHook registra un observador que se invoca inmediatamente tras cada transición válida.
func (fsm *DeterministicNodeFSM) RegisterHook(hook interfaces.StateTransitionHook) {
	fsm.mu.Lock()
	defer fsm.mu.Unlock()
	fsm.hooks = append(fsm.hooks, hook)
}
