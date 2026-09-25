package l1

import (
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/fxamacker/cbor/v2"
)

const (
	DefaultGossipFanout    = 3
	DefaultGossipCacheTTL  = 24 * time.Hour
	MaxGossipSeenRecords   = 10000
)

// GossipBindingEnvelope envuelve un BindingRecord para su difusión epidémica P2P
type GossipBindingEnvelope struct {
	Record    *BindingRecord `cbor:"1,keyasint" json:"record"`
	HopCount  uint8          `cbor:"2,keyasint" json:"hop_count"`
	OriginDID string         `cbor:"3,keyasint" json:"origin_did"`
	Timestamp int64          `cbor:"4,keyasint" json:"timestamp"`
}

// UINGossipManager administra la propagación descentralizada de registros de vinculación
type UINGossipManager struct {
	mu           sync.RWMutex
	identityMgr  *UINIdentityManager
	fanout       int
	seenBindings map[string]int64 // key: hex(KeyID) -> expiration timestamp
	seenOrder    []string
}

// NewUINGossipManager inicializa el motor de gossip para identidades UIN
func NewUINGossipManager(idMgr *UINIdentityManager, fanout int) *UINGossipManager {
	if fanout <= 0 {
		fanout = DefaultGossipFanout
	}
	return &UINGossipManager{
		identityMgr:  idMgr,
		fanout:       fanout,
		seenBindings: make(map[string]int64),
		seenOrder:    make([]string, 0, MaxGossipSeenRecords),
	}
}

// CreateEnvelope empaqueta un BindingRecord firmado para difusión
func (g *UINGossipManager) CreateEnvelope(record *BindingRecord, originDID string) (*GossipBindingEnvelope, []byte, error) {
	if record == nil {
		return nil, nil, errors.New("record no puede ser nulo")
	}

	env := &GossipBindingEnvelope{
		Record:    record,
		HopCount:  0,
		OriginDID: originDID,
		Timestamp: time.Now().Unix(),
	}

	data, err := cbor.Marshal(env)
	if err != nil {
		return nil, nil, fmt.Errorf("error serializando gossip envelope: %w", err)
	}

	// Marcar localmente como ya visto
	keyHex := hex.EncodeToString(record.KeyID[:])
	g.markSeen(keyHex)

	return env, data, nil
}

// ProcessIncomingEnvelope valida criptográficamente un sobre recibido y determina si debe retransmitirse
func (g *UINGossipManager) ProcessIncomingEnvelope(wire []byte, rootPubKey ed25519.PublicKey) (*BindingRecord, bool, error) {
	var env GossipBindingEnvelope
	if err := cbor.Unmarshal(wire, &env); err != nil {
		return nil, false, fmt.Errorf("error deserializando CBOR gossip: %w", err)
	}

	if env.Record == nil {
		return nil, false, errors.New("sobre gossip sin registro de vinculación")
	}

	keyHex := hex.EncodeToString(env.Record.KeyID[:])

	g.mu.Lock()
	now := time.Now().Unix()
	if exp, exists := g.seenBindings[keyHex]; exists && exp > now {
		g.mu.Unlock()
		return env.Record, false, nil // Ya visto, descartar retransmisión
	}
	g.mu.Unlock()

	// Validación criptográfica y temporal del registro
	if env.Record.Version != 1 {
		return nil, false, errors.New("versión de BindingRecord no soportada")
	}

	nowTime := time.Now().UTC()
	if nowTime.Before(env.Record.ValidFrom) || nowTime.After(env.Record.ValidUntil) {
		return nil, false, fmt.Errorf("BindingRecord expirado o fuera de vigencia (%s)", keyHex)
	}

	// Validar firma digital de la raíz soberana
	payload := env.Record.canonicalBytes()
	if !ed25519.Verify(rootPubKey, payload, env.Record.SigRoot) {
		return nil, false, errors.New("firma digital de raíz inválida en BindingRecord")
	}

	// Registrar como visto y almacenar en gestor UIN si corresponde
	g.markSeen(keyHex)

	return env.Record, true, nil
}

// SelectGossipTargets selecciona aleatoriamente un subconjunto de pares para difusión (Fanout)
func (g *UINGossipManager) SelectGossipTargets(peers []*PeerNode, excludeDID string) []*PeerNode {
	if len(peers) == 0 {
		return nil
	}

	eligible := make([]*PeerNode, 0, len(peers))
	for _, p := range peers {
		if p != nil && p.DID != excludeDID {
			eligible = append(eligible, p)
		}
	}

	if len(eligible) <= g.fanout {
		return eligible
	}

	// Mezclar aleatoriamente y tomar fanout
	shuffled := make([]*PeerNode, len(eligible))
	copy(shuffled, eligible)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	return shuffled[:g.fanout]
}

func (g *UINGossipManager) markSeen(keyHex string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.seenBindings[keyHex] = time.Now().Add(DefaultGossipCacheTTL).Unix()
	g.seenOrder = append(g.seenOrder, keyHex)

	if len(g.seenOrder) > MaxGossipSeenRecords {
		oldest := g.seenOrder[0]
		g.seenOrder = g.seenOrder[1:]
		delete(g.seenBindings, oldest)
	}
}
