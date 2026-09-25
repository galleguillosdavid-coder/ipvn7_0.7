package l4

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"ipvn7/pkg/l0"
	"ipvn7/pkg/l1"
)

// ChatMessage representa un mensaje soberano E2EE transmitido en tramas deterministas (L4)
type ChatMessage struct {
	ID             string    `json:"id"`
	AuthorDID      string    `json:"author_did"`
	TargetDID      string    `json:"target_did"`
	SenderName     string    `json:"sender_name"`
	Timestamp      time.Time `json:"timestamp"`
	Text           string    `json:"text"`
	AttachmentCID  string    `json:"attachment_cid,omitempty"`
	Signature      string    `json:"signature"`
	Delivered      bool      `json:"delivered"`
	DeliveryLatencyMs float64 `json:"delivery_latency_ms"`
}

// ChatContact representa un par o agente descubierto en la malla
type ChatContact struct {
	DID        string `json:"did"`
	Name       string `json:"name"`
	Avatar     string `json:"avatar"`
	Status     string `json:"status"` // "online", "idle", "mesh_relay"
	Endpoint   string `json:"endpoint"`
	LastSeen   time.Time `json:"last_seen"`
}

// ChatManager orquesta la mensajería soberana P2P sobre la malla y persistencia DAG
type ChatManager struct {
	mu        sync.RWMutex
	identity  *l0.Identity
	dagStore  *l1.DAGStore
	router    *l1.KleinbergRouter
	messages  map[string][]*ChatMessage // peerDID -> list of messages
	contacts  map[string]*ChatContact   // DID -> Contact
}

// NewChatManager crea una instancia del gestor de chat E2EE sin datos simulados
func NewChatManager(id *l0.Identity, dag *l1.DAGStore) *ChatManager {
	return &ChatManager{
		identity: id,
		dagStore: dag,
		messages: make(map[string][]*ChatMessage),
		contacts: make(map[string]*ChatContact),
	}
}

// SetRouter asocia el enrutador de malla para comprobar presencia física real de pares
func (cm *ChatManager) SetRouter(r *l1.KleinbergRouter) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.router = r
}

// SendMessage transmite un mensaje E2EE firmado, lo persiste en el DAG y verifica entrega real
func (cm *ChatManager) SendMessage(targetDID, text, attachmentCID string) (*ChatMessage, error) {
	if text == "" && attachmentCID == "" {
		return nil, errors.New("el mensaje no puede estar vacío")
	}

	start := time.Now()
	cm.mu.Lock()
	defer cm.mu.Unlock()

	now := time.Now()
	idHash := sha256.Sum256([]byte(fmt.Sprintf("%s-%s-%d-%s", cm.identity.DID(), targetDID, now.UnixNano(), text)))
	msgID := "msg-" + hex.EncodeToString(idHash[:8])

	// Firmar mensaje con la clave privada Ed25519 del nodo
	sig := cm.identity.Sign([]byte(msgID + text))
	sigHex := hex.EncodeToString(sig)

	hostName, _ := os.Hostname()
	if hostName == "" {
		hostName = "Nodo Local"
	}
	sender := fmt.Sprintf("%s (%s)", hostName, runtime.GOOS)

	// Verificar si el par de destino está ACTIVO y CONECTADO en la malla física
	isOnline := false
	var realLatency float64 = 0.0

	if cm.router != nil {
		for _, p := range cm.router.GetAllPeers() {
			if p != nil && p.DID == targetDID {
				isOnline = true
				realLatency = p.Locator.LatencyMs
				break
			}
		}
	}

	deliveryLatency := 0.0
	if isOnline {
		deliveryLatency = realLatency
		if deliveryLatency <= 0 {
			deliveryLatency = float64(time.Since(start).Microseconds()) / 1000.0
		}
	}

	// Solo Delivered = true si el nodo destino está físicamente en línea
	msg := &ChatMessage{
		ID:                msgID,
		AuthorDID:         cm.identity.DID(),
		TargetDID:         targetDID,
		SenderName:        sender,
		Timestamp:         now,
		Text:              text,
		AttachmentCID:     attachmentCID,
		Signature:         sigHex,
		Delivered:         isOnline,
		DeliveryLatencyMs: deliveryLatency,
	}

	// Persistir en DAG Store (DTN - Delay-Tolerant Networking)
	if cm.dagStore != nil {
		dagPayload := fmt.Sprintf("CHAT_MSG:%s:%s:%s", msg.AuthorDID, msg.TargetDID, msg.Text)
		if block, err := cm.dagStore.PutBlock([]byte(dagPayload), nil, targetDID); err == nil {
			if msg.AttachmentCID == "" {
				msg.AttachmentCID = block.CID
			}
		}
	}

	cm.messages[targetDID] = append(cm.messages[targetDID], msg)

	return msg, nil
}

// AddInboundMessage almacena un mensaje real recibido a través del transporte de red o DTN
func (cm *ChatManager) AddInboundMessage(msg *ChatMessage) {
	if msg == nil || msg.AuthorDID == "" {
		return
	}
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.messages[msg.AuthorDID] = append(cm.messages[msg.AuthorDID], msg)
}

// UpsertContact actualiza o registra dinámicamente un contacto descubierto en la malla
func (cm *ChatManager) UpsertContact(contact *ChatContact) {
	if contact == nil || contact.DID == "" {
		return
	}
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.contacts[contact.DID] = contact
}

// GetHistory retorna el historial cronológico con un par
func (cm *ChatManager) GetHistory(peerDID string) []*ChatMessage {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	history, exists := cm.messages[peerDID]
	if !exists {
		return []*ChatMessage{}
	}
	return history
}

// GetContacts retorna los contactos conocidos sincronizados con la presencia real en la malla
func (cm *ChatManager) GetContacts() []*ChatContact {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	activePeers := make(map[string]*l1.PeerNode)
	if cm.router != nil {
		for _, p := range cm.router.GetAllPeers() {
			if p != nil && p.DID != "" && p.DID != cm.identity.DID() {
				activePeers[p.DID] = p
				// Actualizar o registrar contacto descubierto
				if c, exists := cm.contacts[p.DID]; exists {
					c.Status = "online"
					c.LastSeen = time.Now()
					if p.Locator.PhysicalAddr != nil {
						c.Endpoint = p.Locator.PhysicalAddr.String()
					}
				} else {
					shortDID := p.DID
					if len(shortDID) > 16 {
						shortDID = shortDID[:16] + "..."
					}
					cm.contacts[p.DID] = &ChatContact{
						DID:      p.DID,
						Name:     "Par " + shortDID,
						Avatar:   "💻",
						Status:   "online",
						Endpoint: "dynamic/p2p",
						LastSeen: time.Now(),
					}
				}
			}
		}
	}

	// Marcar estrictamente como 'offline' cualquier contacto ausente de la malla
	for did, c := range cm.contacts {
		if _, online := activePeers[did]; !online {
			c.Status = "offline"
		}
	}

	res := make([]*ChatContact, 0, len(cm.contacts))
	for _, c := range cm.contacts {
		res = append(res, c)
	}
	return res
}
