package l4

import (
	"net"
	"testing"

	"ipvn7/pkg/l0"
	"ipvn7/pkg/l1"
)

func TestChatManager_SendMessageAndHistory(t *testing.T) {
	id, err := l0.GenerateIdentity()
	if err != nil {
		t.Fatalf("Error generando identidad: %v", err)
	}

	dag := l1.NewDAGStore(id)
	cm := NewChatManager(id, dag)

	peerID, err := l0.GenerateIdentity()
	if err != nil {
		t.Fatalf("Error generando identidad del par: %v", err)
	}
	peerDID := peerID.DID()

	// 1. Cero contactos falsos al iniciar (DEC-001/008)
	contacts := cm.GetContacts()
	if len(contacts) != 0 {
		t.Errorf("Se esperaba lista de contactos vacía sin pares reales, obtenidos %d", len(contacts))
	}

	// 2. Registrar contacto empírico descubierto previamente
	cm.UpsertContact(&ChatContact{
		DID:      peerDID,
		Name:     "notebook.ipv7",
		Avatar:   "💻",
		Status:   "offline",
		Endpoint: "dynamic/p2p",
	})

	contacts = cm.GetContacts()
	if len(contacts) != 1 {
		t.Fatalf("Se esperaba 1 contacto registrado, obtenidos %d", len(contacts))
	}
	if contacts[0].Status != "offline" {
		t.Errorf("El contacto ausente en el router debe reportar status 'offline'")
	}

	// 3. Enviar mensaje a peer offline: debe encolarse en DAG pero NO marcarse como Delivered
	msg, err := cm.SendMessage(peerDID, "Mensaje de prueba a nodo apagado", "")
	if err != nil {
		t.Fatalf("Fallo enviando mensaje: %v", err)
	}

	if msg.Delivered {
		t.Errorf("Un mensaje a un nodo offline NO debe marcarse como Delivered (DEC-001/008)")
	}

	if msg.Signature == "" {
		t.Errorf("El mensaje debe contener firma digital Ed25519")
	}

	// 4. Recuperar historial: exactamente 1 mensaje
	history := cm.GetHistory(peerDID)
	if len(history) != 1 {
		t.Errorf("Se esperaba exactamente 1 mensaje en el historial, obtenidos %d", len(history))
	}

	// 5. Conectar router con peer en línea y verificar entrega real
	router := l1.NewKleinbergRouter(id)
	cm.SetRouter(router)
	udpAddr, _ := net.ResolveUDPAddr("udp4", "192.0.2.1:7777")
	_ = router.AddOrUpdatePeer(peerDID, udpAddr, 4.2)

	msgOnline, err := cm.SendMessage(peerDID, "Mensaje con nodo en línea", "")
	if err != nil {
		t.Fatalf("Error enviando mensaje a par en línea: %v", err)
	}
	if !msgOnline.Delivered {
		t.Errorf("Mensaje a nodo en línea debe reportar Delivered = true")
	}
}
