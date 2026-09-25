package l2

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// GraphOperationType define si se añade o retira un elemento del grafo topológico
type GraphOperationType byte

const (
	OpAddNode GraphOperationType = iota + 1
	OpRemoveNode
	OpAddEdge
	OpRemoveEdge
)

// CRDTNode representa un nodo topológico con control de concurrencia LWW (Last-Write-Wins)
type CRDTNode struct {
	DID       string
	Label     string
	Timestamp int64
	Tombstone bool
}

// CRDTEdge representa una arista direccional en el grafo topológico
type CRDTEdge struct {
	ID        string
	SourceDID string
	TargetDID string
	Relation  string
	LatencyMs float64
	Timestamp int64
	Tombstone bool
}

// GraphDelta transporta una mutación atómica para ser replicada por la malla
type GraphDelta struct {
	OpType    GraphOperationType
	Node      *CRDTNode
	Edge      *CRDTEdge
	OriginDID string
	Clock     uint64
}

// CRDTTopologyGraph mantiene la réplica local convergente del grafo de red
type CRDTTopologyGraph struct {
	mu           sync.RWMutex
	localDID     string
	lamportClock atomic.Uint64
	nodes        map[string]*CRDTNode
	edges        map[string]*CRDTEdge
	deltaLog     []*GraphDelta
}

// NewCRDTTopologyGraph inicializa el grafo distribuido CRDT
func NewCRDTTopologyGraph(localDID string) *CRDTTopologyGraph {
	return &CRDTTopologyGraph{
		localDID: localDID,
		nodes:    make(map[string]*CRDTNode),
		edges:    make(map[string]*CRDTEdge),
		deltaLog: make([]*GraphDelta, 0, 1024),
	}
}

// AddNode inserta o actualiza un nodo con semántica LWW
func (g *CRDTTopologyGraph) AddNode(did, label string) *GraphDelta {
	g.mu.Lock()
	defer g.mu.Unlock()

	clk := g.lamportClock.Add(1)
	now := time.Now().UnixNano()

	node := &CRDTNode{
		DID:       did,
		Label:     label,
		Timestamp: now,
		Tombstone: false,
	}

	existing, exists := g.nodes[did]
	if !exists || node.Timestamp > existing.Timestamp {
		g.nodes[did] = node
	}

	delta := &GraphDelta{
		OpType:    OpAddNode,
		Node:      node,
		OriginDID: g.localDID,
		Clock:     clk,
	}
	g.deltaLog = append(g.deltaLog, delta)
	return delta
}

// RemoveNode marca un nodo con lápida (tombstone)
func (g *CRDTTopologyGraph) RemoveNode(did string) *GraphDelta {
	g.mu.Lock()
	defer g.mu.Unlock()

	clk := g.lamportClock.Add(1)
	now := time.Now().UnixNano()

	node := &CRDTNode{
		DID:       did,
		Timestamp: now,
		Tombstone: true,
	}

	existing, exists := g.nodes[did]
	if exists && node.Timestamp >= existing.Timestamp {
		existing.Tombstone = true
		existing.Timestamp = now
	}

	delta := &GraphDelta{
		OpType:    OpRemoveNode,
		Node:      node,
		OriginDID: g.localDID,
		Clock:     clk,
	}
	g.deltaLog = append(g.deltaLog, delta)
	return delta
}

// AddEdge inserta una conexión topológica entre dos nodos
func (g *CRDTTopologyGraph) AddEdge(id, srcDID, dstDID, relation string, latencyMs float64) (*GraphDelta, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	clk := g.lamportClock.Add(1)
	now := time.Now().UnixNano()

	edge := &CRDTEdge{
		ID:        id,
		SourceDID: srcDID,
		TargetDID: dstDID,
		Relation:  relation,
		LatencyMs: latencyMs,
		Timestamp: now,
		Tombstone: false,
	}

	existing, exists := g.edges[id]
	if !exists || edge.Timestamp > existing.Timestamp {
		g.edges[id] = edge
	}

	delta := &GraphDelta{
		OpType:    OpAddEdge,
		Edge:      edge,
		OriginDID: g.localDID,
		Clock:     clk,
	}
	g.deltaLog = append(g.deltaLog, delta)
	return delta, nil
}

// MergeDelta aplica un cambio remoto asegurando convergencia determinista (LWW)
func (g *CRDTTopologyGraph) MergeDelta(d *GraphDelta) error {
	if d == nil {
		return errors.New("crdt: delta nulo")
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	// Actualizar reloj lógico Lamport
	currentClk := g.lamportClock.Load()
	if d.Clock > currentClk {
		g.lamportClock.Store(d.Clock + 1)
	} else {
		g.lamportClock.Add(1)
	}

	switch d.OpType {
	case OpAddNode, OpRemoveNode:
		if d.Node == nil {
			return errors.New("crdt: payload de nodo ausente")
		}
		existing, exists := g.nodes[d.Node.DID]
		if !exists || d.Node.Timestamp > existing.Timestamp {
			g.nodes[d.Node.DID] = d.Node
		}
	case OpAddEdge, OpRemoveEdge:
		if d.Edge == nil {
			return errors.New("crdt: payload de arista ausente")
		}
		existing, exists := g.edges[d.Edge.ID]
		if !exists || d.Edge.Timestamp > existing.Timestamp {
			g.edges[d.Edge.ID] = d.Edge
		}
	}

	g.deltaLog = append(g.deltaLog, d)
	return nil
}

// ActiveElements retorna el conteo de nodos y aristas no eliminados
func (g *CRDTTopologyGraph) ActiveElements() (activeNodes, activeEdges int) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	for _, n := range g.nodes {
		if !n.Tombstone {
			activeNodes++
		}
	}
	for _, e := range g.edges {
		if !e.Tombstone {
			activeEdges++
		}
	}
	return activeNodes, activeEdges
}

// GetDeltaLog retorna una copia del historial de operaciones para sincronización inicial
func (g *CRDTTopologyGraph) GetDeltaLog() []*GraphDelta {
	g.mu.RLock()
	defer g.mu.RUnlock()

	copied := make([]*GraphDelta, len(g.deltaLog))
	copy(copied, g.deltaLog)
	return copied
}
