// Package interfaces define los contratos raíz universales e inmutables del protocolo IPVN7 (v0.7).
package interfaces

import (
	"net"
	"time"
)

// MeshPeerDescriptor encapsula la información de contacto y telemetría de un par en la malla.
type MeshPeerDescriptor interface {
	// DID retorna el identificador soberano del par.
	DID() string

	// PhysicalAddr retorna la dirección física UDP resuelta (IP:Puerto).
	PhysicalAddr() net.Addr

	// Latency retorna la latencia RTT medida en nanosegundos hacia el par.
	Latency() time.Duration

	// RingIndex retorna el nivel de anillo Kleinberg al que fue asignado según distancia métrica.
	RingIndex() int

	// LastSeen retorna el instante temporal del último paquete válido recibido.
	LastSeen() time.Time

	// IsAlive indica si el par se considera activo dentro de la ventana de expiración.
	IsAlive(ttl time.Duration) bool
}

// KleinbergMetricCalculator calcula distancias euclidianas o hiperbólicas en el espacio de claves de 256 bits.
// Permite garantizar enrutamiento greedily en O(log^2 N) saltos sin inundación de paquetes.
type KleinbergMetricCalculator interface {
	// Distance calcula la distancia métrica escalar normalizada entre dos DIDs soberanos.
	Distance(didA, didB string) float64

	// BestNextHop selecciona, del conjunto de pares disponibles, aquel que minimiza la distancia al destino.
	BestNextHop(targetDID string, peers []MeshPeerDescriptor) MeshPeerDescriptor
}

// RoutingEngine abstrae la tabla de enrutamiento distribuida y los anillos de descubrimiento.
// Regla Inquebrantable de Topología: No permite auto-emparejamiento (Self-Peering) ni nodos fantasma.
type RoutingEngine interface {
	// LookupNextHop resuelve el salto óptimo para alcanzar un nodo destino en la malla.
	LookupNextHop(targetDID string) (MeshPeerDescriptor, error)

	// AddPeer registra o actualiza un par externo verificado. Rechaza si el DID es el nodo local.
	AddPeer(peer MeshPeerDescriptor) error

	// RemovePeer remueve un par de la tabla de enrutamiento por desconexión o desconfianza.
	RemovePeer(did string)

	// ListPeers retorna la lista inmutable de pares activos en la topología limpia.
	ListPeers() []MeshPeerDescriptor

	// TotalPeers retorna la cantidad de pares externos enlazados actualmente.
	TotalPeers() int
}

// ARPTableResolver administra el mapeo bidireccional entre identidades soberanas y direcciones de red.
// Asocia: DID <-> UIN <-> IPv6 criptográfico <-> IPv4 virtual <-> Socket UDP Físico.
type ARPTableResolver interface {
	// RegisterBinding almacena o renueva una correspondencia entre DID y dirección física UDP.
	RegisterBinding(did string, physicalAddr string, ipv4 string, ipv6 string)

	// ResolvePhysical busca el socket UDP físico asociado a un DID o IP virtual.
	ResolvePhysical(identifier string) (string, bool)

	// ResolveDID busca el DID soberano asociado a una dirección IP física o virtual.
	ResolveDID(ipOrAddr string) (string, bool)

	// CleanExpired elimina entradas cuya vigencia temporal haya expirado.
	CleanExpired(ttl time.Duration) int
}
