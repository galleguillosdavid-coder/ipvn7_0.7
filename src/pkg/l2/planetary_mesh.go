package l2

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sync"
)

var (
	ErrNodeNotFound     = errors.New("planetary_mesh: nodo no encontrado en el cluster")
	ErrUnreachableTarget = errors.New("planetary_mesh: destino inalcanzable tras agotar saltos de Kleinberg")
	ErrClusterPartition = errors.New("planetary_mesh: cluster fragmentado o sin conectividad mínima")
)

// Coordinates2D representa la posición euclidiana de un nodo en el plano de Kleinberg
type Coordinates2D struct {
	X float64
	Y float64
}

// Distance calcula la distancia euclidiana canónica entre dos coordenadas
func (c Coordinates2D) Distance(other Coordinates2D) float64 {
	dx := c.X - other.X
	dy := c.Y - other.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// PlanetaryNode representa un nodo individual en la malla distribuida
type PlanetaryNode struct {
	ID         string
	DID        string
	Coords     Coordinates2D
	ShortLinks []string // Enlaces locales (rejilla de proximidad)
	LongLinks  []string // Atajos de largo alcance probabilísticos (Kleinberg $r=2$)
	IsOnline   bool
}

// PlanetaryMeshCluster orquesta una topología de $N$ nodos geodistribuidos
type PlanetaryMeshCluster struct {
	mu         sync.RWMutex
	nodes      map[string]*PlanetaryNode
	gridDim    int
	maxHops    int
}

// NewPlanetaryMeshCluster crea un clúster de malla planetaria con topología de Kleinberg
func NewPlanetaryMeshCluster(nodeCount int) (*PlanetaryMeshCluster, error) {
	if nodeCount < 4 {
		return nil, errors.New("planetary_mesh: se requieren al menos 4 nodos para formar la malla")
	}

	dim := int(math.Ceil(math.Sqrt(float64(nodeCount))))
	cluster := &PlanetaryMeshCluster{
		nodes:   make(map[string]*PlanetaryNode),
		gridDim: dim,
		maxHops: dim * 4,
	}

	r := rand.New(rand.NewSource(1337)) // Semilla determinista para pruebas reproducibles

	// 1. Inicializar nodos en una retícula 2D
	idx := 0
	for x := 0; x < dim && idx < nodeCount; x++ {
		for y := 0; y < dim && idx < nodeCount; y++ {
			nodeID := fmt.Sprintf("node-%03d", idx)
			cluster.nodes[nodeID] = &PlanetaryNode{
				ID:       nodeID,
				DID:      fmt.Sprintf("did:ipvn7:planet:%03d", idx),
				Coords:   Coordinates2D{X: float64(x), Y: float64(y)},
				IsOnline: true,
			}
			idx++
		}
	}

	// 2. Establecer enlaces locales (Short Links: vecinos adyacentes Manhattan)
	for id, node := range cluster.nodes {
		for _, other := range cluster.nodes {
			if id == other.ID {
				continue
			}
			dist := node.Coords.Distance(other.Coords)
			if dist <= 1.05 { // Vecinos inmediatos en la rejilla
				node.ShortLinks = append(node.ShortLinks, other.ID)
			}
		}
	}

	// 3. Establecer enlaces de largo alcance (Kleinberg Long Links con probabilidad $P \propto d^{-2}$)
	for id, node := range cluster.nodes {
		candidates := make([]string, 0)
		weights := make([]float64, 0)
		totalWeight := 0.0

		for otherID, otherNode := range cluster.nodes {
			if id == otherID {
				continue
			}
			dist := node.Coords.Distance(otherNode.Coords)
			if dist > 1.05 {
				prob := 1.0 / (dist * dist)
				candidates = append(candidates, otherID)
				weights = append(weights, prob)
				totalWeight += prob
			}
		}

		// Seleccionar hasta 2 atajos de largo alcance ponderados
		numShortcuts := 2
		if len(candidates) < numShortcuts {
			numShortcuts = len(candidates)
		}

		for s := 0; s < numShortcuts; s++ {
			if totalWeight <= 0 {
				break
			}
			threshold := r.Float64() * totalWeight
			acc := 0.0
			for i, candID := range candidates {
				acc += weights[i]
				if acc >= threshold {
					node.LongLinks = append(node.LongLinks, candID)
					break
				}
			}
		}
	}

	return cluster, nil
}

// RoutePacket realiza el enrutamiento greedy descentralizado de Kleinberg
func (c *PlanetaryMeshCluster) RoutePacket(srcID, dstID string) ([]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	dst, exists := c.nodes[dstID]
	if !exists || !dst.IsOnline {
		return nil, ErrNodeNotFound
	}

	current, exists := c.nodes[srcID]
	if !exists || !current.IsOnline {
		return nil, ErrNodeNotFound
	}

	path := []string{srcID}
	visited := make(map[string]bool)
	visited[srcID] = true

	for len(path) < c.maxHops {
		if current.ID == dstID {
			return path, nil
		}

		bestNext := ""
		bestDist := current.Coords.Distance(dst.Coords)

		allNeighbors := append(append([]string{}, current.ShortLinks...), current.LongLinks...)
		for _, neighborID := range allNeighbors {
			neighbor, ok := c.nodes[neighborID]
			if !ok || !neighbor.IsOnline || visited[neighborID] {
				continue
			}

			d := neighbor.Coords.Distance(dst.Coords)
			if d < bestDist {
				bestDist = d
				bestNext = neighborID
			}
		}

		if bestNext == "" {
			// Fallback local: backtracking mínimo o selección del vecino online no visitado
			for _, neighborID := range allNeighbors {
				neighbor, ok := c.nodes[neighborID]
				if ok && neighbor.IsOnline && !visited[neighborID] {
					bestNext = neighborID
					break
				}
			}
			if bestNext == "" {
				return nil, ErrUnreachableTarget
			}
		}

		visited[bestNext] = true
		path = append(path, bestNext)
		current = c.nodes[bestNext]
	}

	return nil, ErrUnreachableTarget
}

// SimulateFailure desconecta aleatoriamente un porcentaje de nodos probando la resiliencia
func (c *PlanetaryMeshCluster) SimulateFailure(failureRatio float64) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	failedCount := 0
	for _, node := range c.nodes {
		if rand.Float64() < failureRatio {
			node.IsOnline = false
			failedCount++
		}
	}
	return failedCount
}

// NodeCount retorna el número total de nodos en la malla
func (c *PlanetaryMeshCluster) NodeCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.nodes)
}
