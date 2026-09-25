// Package interfaces define los contratos raíz universales e inmutables del protocolo IPVN7 (v0.7).
// Todo subsistema actual y futuro debe implementar estos contratos base desacoplados,
// facilitando la reproducibilidad algorítmica, auditoría formal y optimización por IA.
package interfaces

import (
	"context"
	"net"
	"time"
)

// PacketBuffer representa un segmento de memoria gestionado mediante técnicas Zero-Copy.
// Invariante de Rendimiento: La memoria subyacente proviene de un pool preasignado y
// debe liberarse obligatoriamente tras la transmisión o deserialización para prevenir GC stalls.
type PacketBuffer interface {
	// RawSlice retorna el fragmento de memoria contiguo para lectura o escritura directa en socket.
	RawSlice() []byte

	// Data retorna la porción útil de carga (payload) excluyendo cabeceras descartadas.
	Data() []byte

	// SetLen ajusta la longitud de datos válidos contenidos en el búfer.
	SetLen(n int)

	// Capacity retorna la capacidad máxima del búfer preasignado.
	Capacity() int

	// Retain incrementa atómicamente el contador de referencias Zero-Copy.
	Retain()

	// Release devuelve el búfer a su nivel correspondiente en el Pool de Memoria (Zero-Copy).
	Release()
}

// BufferPoolProvider abstrae la asignación de memoria a velocidad de línea para datagramas de red.
// Debe implementar segregación por tramos de tamaño (ej: Tiny 64B, MTU 1500B, Jumbo 64KB).
type BufferPoolProvider interface {
	// Acquire toma un búfer con capacidad mínima suficiente para alojar 'minCapacity' bytes.
	Acquire(minCapacity int) PacketBuffer
}

// PacketTransport abstrae la capa física o virtual de transporte de datagramas sobre el medio.
// Soporta implementaciones UDP directas (sockets reales), túneles TUN en kernel o mock de laboratorio.
type PacketTransport interface {
	// LocalAddr retorna el endpoint local enlazado por el socket físico.
	LocalAddr() net.Addr

	// SendPacket transmite un búfer de datos hacia un endpoint destino sin generar copias intermedias.
	// Retorna la cantidad de bytes escritos físicamente en el socket y error en caso de fallo físico.
	SendPacket(ctx context.Context, dest net.Addr, buf PacketBuffer) (int, error)

	// ReceivePacket extrae el siguiente datagrama disponible del socket físico bloqueando hasta timeout.
	// La implementación adquiere internamente un búfer del pool que el consumidor debe liberar.
	ReceivePacket(ctx context.Context) (PacketBuffer, net.Addr, error)

	// SetReadTimeout establece la ventana temporal máxima de espera antes de emitir error de deadline.
	SetReadTimeout(d time.Duration) error

	// Close cierra el descriptor de socket subyacente y libera recursos del sistema operativo.
	Close() error
}

// TunVirtualDevice abstrae el adaptador de red en espacio de usuario o kernel del sistema operativo.
type TunVirtualDevice interface {
	// Name retorna el identificador del dispositivo virtual (ej: 'ipvn7-tun0' o 'wintun').
	Name() string

	// IsUserspace indica si el adaptador opera en modo emulado de memoria o si se acopló al kernel.
	IsUserspace() bool

	// WriteToTun inyecta un paquete IP deserializado directamente a la pila de red del sistema anfitrión.
	WriteToTun(packet []byte) (int, error)

	// ReadFromTun lee un paquete IP emitido por el sistema anfitrión para ser encapsulado en IPVN7.
	ReadFromTun(buf []byte) (int, error)

	// Close desmonta la interfaz virtual del sistema operativo.
	Close() error
}
