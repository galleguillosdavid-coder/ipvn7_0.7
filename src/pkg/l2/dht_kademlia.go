// Package l2 implementa el enrutamiento y descubrimiento Kademlia DHT adaptativo
// para la localización descentralizada de nodos en la malla ipvn7 v0.7.
package l2

import (
	"bytes"
	"crypto/sha256"
	"math/bits"
	"sort"
	"sync"
	"time"
)

const (
	// KBucketSize capacidad de cada cubo k según estándar Kademlia
	KBucketSize = 8
	// IDByteLen longitud del espacio de claves (256 bits = 32 bytes)
	IDByteLen = 32
)

// NodeID clave de identificación de 256 bits
type NodeID [IDByteLen]byte

// Contact representa un par conocido en la tabla de rutas
type Contact struct {
	ID       NodeID    `json:"id"`
	DID      string    `json:"did"`
	Endpoint string    `json:"endpoint"`
	LastSeen time.Time `json:"last_seen"`
}

// NewNodeIDFromDID deriva deterministamente el NodeID a partir del DID soberano
func NewNodeIDFromDID(did string) NodeID {
	return sha256.Sum256([]byte(did))
}

// XORDistance calcula la métrica de distancia XOR entre dos identificadores
func XORDistance(a, b NodeID) [IDByteLen]byte {
	var dist [IDByteLen]byte
	for i := 0; i < IDByteLen; i++ {
		dist[i] = a[i] ^ b[i]
	}
	return dist
}

// CommonPrefixLen cuenta los ceros iniciales en la distancia XOR
func CommonPrefixLen(a, b NodeID) int {
	dist := XORDistance(a, b)
	for i, val := range dist {
		if val != 0 {
			return i*8 + bits.LeadingZeros8(val)
		}
	}
	return IDByteLen * 8
}

// CompareDistance compara dos distancias devuelto <0 si d1 < d2, 0 si d1 == d2, >0 si d1 > d2
func CompareDistance(d1, d2 [IDByteLen]byte) int {
	return bytes.Compare(d1[:], d2[:])
}

// RoutingTable gestiona los 256 k-buckets para el espacio de claves de 256 bits
type RoutingTable struct {
	mu      sync.RWMutex
	selfID  NodeID
	buckets [IDByteLen * 8][]Contact
}

// NewRoutingTable inicializa una nueva tabla de rutas Kademlia
func NewRoutingTable(selfID NodeID) *RoutingTable {
	rt := &RoutingTable{
		selfID: selfID,
	}
	for i := 0; i < len(rt.buckets); i++ {
		rt.buckets[i] = make([]Contact, 0, KBucketSize)
	}
	return rt
}

// Update inserta o actualiza un contacto en el k-bucket correspondiente
func (rt *RoutingTable) Update(c Contact) {
	if c.ID == rt.selfID {
		return // Prohibido el auto-emparejamiento (Regla 2)
	}

	cpl := CommonPrefixLen(rt.selfID, c.ID)
	if cpl >= len(rt.buckets) {
		cpl = len(rt.buckets) - 1
	}

	rt.mu.Lock()
	defer rt.mu.Unlock()

	bucket := rt.buckets[cpl]
	// Buscar si el contacto ya existe en el bucket
	for i, item := range bucket {
		if item.ID == c.ID {
			// Mover a la cola (más recientemente visto)
			bucket = append(append(bucket[:i], bucket[i+1:]...), c)
			rt.buckets[cpl] = bucket
			return
		}
	}

	// Si el bucket no está lleno, añadir el nuevo contacto
	if len(bucket) < KBucketSize {
		rt.buckets[cpl] = append(bucket, c)
	}
	// Si está lleno, la política Kademlia conserva los más antiguos y confiables
}

// FindClosest localiza los 'count' contactos más cercanos al objetivo según métrica XOR
func (rt *RoutingTable) FindClosest(target NodeID, count int) []Contact {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	var all []Contact
	for _, bucket := range rt.buckets {
		all = append(all, bucket...)
	}

	type scoredContact struct {
		contact  Contact
		distance [IDByteLen]byte
	}

	scored := make([]scoredContact, len(all))
	for i, c := range all {
		scored[i] = scoredContact{
			contact:  c,
			distance: XORDistance(target, c.ID),
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		return CompareDistance(scored[i].distance, scored[j].distance) < 0
	})

	if len(scored) > count {
		scored = scored[:count]
	}

	res := make([]Contact, len(scored))
	for i, sc := range scored {
		res[i] = sc.contact
	}
	return res
}

// TotalContacts cuenta la cantidad total de pares activos en la tabla
func (rt *RoutingTable) TotalContacts() int {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	total := 0
	for _, bucket := range rt.buckets {
		total += len(bucket)
	}
	return total
}
