// Package interfaces define los contratos raíz universales e inmutables del protocolo IPVN7 (v0.7).
package interfaces

import (
	"context"
	"net"
)

// PacketContext encapsula el ciclo de vida y los metadatos de un paquete a través
// de la línea de producción (Pipes & Filters).
type PacketContext struct {
	// RawData contiene el payload de bytes crudos recibidos del socket.
	RawData []byte

	// Buffer referencia el búfer de memoria Zero-Copy asignado desde el pool.
	Buffer PacketBuffer

	// RemoteAddr contiene el socket físico de origen del datagrama (IP:Puerto).
	RemoteAddr net.Addr

	// Packet almacena la representación deserializada en formato canónico CBOR RFC 8949.
	Packet CanonicalPacket

	// SourceDID contiene el Identificador Descentralizado soberano del remitente validado.
	SourceDID string

	// Decision registra el dictamen de seguridad ZTNA y control de acceso.
	Decision SecurityDecision

	// Dropped indica si el paquete fue descartado en alguna estación de la línea.
	Dropped bool

	// DropReason especifica la causa técnica del descarte (para auditoría y telemetría).
	DropReason string

	// Handled indica si el paquete fue consumido completamente por una estación terminal.
	Handled bool
}

// Reset reinicializa todos los campos de PacketContext para su reutilización Zero-Copy.
func (p *PacketContext) Reset() {
	p.RawData = nil
	p.Buffer = nil
	p.RemoteAddr = nil
	p.Packet = nil
	p.SourceDID = ""
	p.Decision = 0
	p.Dropped = false
	p.DropReason = ""
	p.Handled = false
}


// PipelineStage representa una estación individual en la línea de montaje de procesamiento.
// Principio de Diseño: Responsabilidad Única (SRP) y no mutación de estado externo sin contrato.
type PipelineStage interface {
	// Name retorna el identificador único de la estación (ej: "decode_cbor", "ztna_filter").
	Name() string

	// Process ejecuta el procesamiento determinista del paquete en la estación.
	// Si retorna error o marca pCtx.Dropped = true, la línea detiene el avance inmediatamente.
	Process(ctx context.Context, pCtx *PacketContext) error
}

// DeadLetterHandler define el contrato de captura y auditoría para paquetes descartados o corruptos.
type DeadLetterHandler interface {
	// HandleDeadLetter registra y libera los recursos de un paquete que no pudo completar la línea.
	HandleDeadLetter(ctx context.Context, pCtx *PacketContext, stageName string, err error)
}

// DataPipeline abstrae la línea de montaje secuencial y unidireccional de datagramas.
type DataPipeline interface {
	// AddStage añade una estación al final de la línea de montaje.
	AddStage(stage PipelineStage) DataPipeline

	// SetDeadLetterHandler asigna el manejador para descartes y fallos en la línea.
	SetDeadLetterHandler(handler DeadLetterHandler)

	// Execute introduce un paquete en la estación inicial y lo procesa secuencialmente.
	Execute(ctx context.Context, pCtx *PacketContext) error
}
