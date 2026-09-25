package core

import (
	"testing"

	"ipvn7/pkg/interfaces"
)

func TestDeterministicNodeFSM_Lifecycle(t *testing.T) {
	fsm := NewDeterministicNodeFSM(interfaces.NodeStateDisconnected)

	if fsm.CurrentState() != interfaces.NodeStateDisconnected {
		t.Fatalf("Esperado estado inicial Disconnected, obtenido: %s", fsm.CurrentState())
	}

	var transitions []string
	fsm.RegisterHook(func(from, to interfaces.NodeState, event interfaces.FSMEvent, payload interface{}) {
		transitions = append(transitions, from.String()+"->"+to.String())
	})

	// 1. Recibe señal de red -> Sincronizando
	s, err := fsm.TriggerEvent(interfaces.EventNetworkSignal, nil)
	if err != nil || s != interfaces.NodeStateSyncing {
		t.Fatalf("Fallo en transición a Syncing: err=%v, state=%s", err, s)
	}

	// 2. Descarga completada -> En Validación
	s, err = fsm.TriggerEvent(interfaces.EventSyncCompleted, nil)
	if err != nil || s != interfaces.NodeStateValidating {
		t.Fatalf("Fallo en transición a Validating: err=%v, state=%s", err, s)
	}

	// 3. Firma DID validada -> Activo
	s, err = fsm.TriggerEvent(interfaces.EventDIDVerified, "did:ipvn7:test")
	if err != nil || s != interfaces.NodeStateActive {
		t.Fatalf("Fallo en transición a Active: err=%v, state=%s", err, s)
	}

	// 4. Detección de amenaza -> Confinamiento en Cuarentena
	s, err = fsm.TriggerEvent(interfaces.EventSecurityThreat, "replay_attack")
	if err != nil || s != interfaces.NodeStateQuarantined {
		t.Fatalf("Fallo en transición a Quarantined: err=%v, state=%s", err, s)
	}

	// 5. Intento de transición inválida desde Cuarentena
	_, err = fsm.TriggerEvent(interfaces.EventNetworkSignal, nil)
	if err == nil {
		t.Fatalf("Se esperaba error al intentar señal de red en cuarentena")
	}

	// 6. Reset formal -> Regreso a Desconectado
	s, err = fsm.TriggerEvent(interfaces.EventReset, nil)
	if err != nil || s != interfaces.NodeStateDisconnected {
		t.Fatalf("Fallo en reseteo a Disconnected: err=%v, state=%s", err, s)
	}

	if len(transitions) != 5 {
		t.Fatalf("Esperadas 5 transiciones registradas, obtenidas: %d", len(transitions))
	}
}

func TestDeterministicNodeFSM_InvalidTransitions(t *testing.T) {
	fsm := NewDeterministicNodeFSM(interfaces.NodeStateDisconnected)

	// No puede saltar directo de Disconnected a Active sin validar
	_, err := fsm.TriggerEvent(interfaces.EventDIDVerified, nil)
	if err == nil {
		t.Fatalf("Se esperaba rechazo en transición directa Disconnected -> Active")
	}

	if fsm.CurrentState() != interfaces.NodeStateDisconnected {
		t.Fatalf("El estado no debió mutar tras transición rechazada")
	}
}
