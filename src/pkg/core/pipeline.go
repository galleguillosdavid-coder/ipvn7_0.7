package core

import (
	"context"
	"fmt"
	"net"
	"sync"

	"ipvn7/pkg/interfaces"
)

var packetContextPool = sync.Pool{
	New: func() interface{} {
		return &interfaces.PacketContext{}
	},
}

// AcquirePacketContext adquiere un contexto de paquete preasignado desde el pool Zero-Copy.
func AcquirePacketContext(buf interfaces.PacketBuffer, remoteAddr net.Addr) *interfaces.PacketContext {
	pCtx := packetContextPool.Get().(*interfaces.PacketContext)
	pCtx.Reset()
	pCtx.Buffer = buf
	if buf != nil {
		pCtx.RawData = buf.Data()
	}
	pCtx.RemoteAddr = remoteAddr
	return pCtx
}

// ReleasePacketContext libera y recicla el contexto devolviéndolo al pool.
func ReleasePacketContext(pCtx *interfaces.PacketContext) {
	if pCtx == nil {
		return
	}
	pCtx.Reset()
	packetContextPool.Put(pCtx)
}


// LinearPipeline implementa una línea de montaje secuencial y determinista (Pipes & Filters).
type LinearPipeline struct {
	mu         sync.RWMutex
	stages     []interfaces.PipelineStage
	deadLetter interfaces.DeadLetterHandler
}

// NewLinearPipeline instancia una nueva línea de montaje sin estaciones.
func NewLinearPipeline() *LinearPipeline {
	return &LinearPipeline{
		stages:     make([]interfaces.PipelineStage, 0),
		deadLetter: &DefaultDeadLetterHandler{},
	}
}

// AddStage añade una estación de procesamiento al final de la línea de producción.
// Utiliza Copy-On-Write para permitir lectura sin asignación en Execute.
func (p *LinearPipeline) AddStage(stage interfaces.PipelineStage) interfaces.DataPipeline {
	p.mu.Lock()
	defer p.mu.Unlock()
	newStages := make([]interfaces.PipelineStage, len(p.stages)+1)
	copy(newStages, p.stages)
	newStages[len(p.stages)] = stage
	p.stages = newStages
	return p
}

// SetDeadLetterHandler configura el receptor de paquetes descartados o anómalos.
func (p *LinearPipeline) SetDeadLetterHandler(handler interfaces.DeadLetterHandler) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if handler != nil {
		p.deadLetter = handler
	}
}

// Execute procesa el paquete a través de cada estación secuencialmente.
// Si una estación dictamina descarte (Dropped=true) o retorna error,
// el avance se detiene de inmediato y se entrega al Dead-Letter handler.
func (p *LinearPipeline) Execute(ctx context.Context, pCtx *interfaces.PacketContext) error {
	p.mu.RLock()
	stages := p.stages
	deadLetter := p.deadLetter
	p.mu.RUnlock()

	for _, stage := range stages {
		select {
		case <-ctx.Done():
			pCtx.Dropped = true
			pCtx.DropReason = "context_canceled"
			if deadLetter != nil {
				deadLetter.HandleDeadLetter(ctx, pCtx, stage.Name(), ctx.Err())
			}
			return ctx.Err()
		default:
		}

		err := stage.Process(ctx, pCtx)
		if err != nil || pCtx.Dropped {
			if !pCtx.Dropped {
				pCtx.Dropped = true
				if pCtx.DropReason == "" {
					pCtx.DropReason = fmt.Sprintf("stage_error: %v", err)
				}
			}
			if deadLetter != nil {
				deadLetter.HandleDeadLetter(ctx, pCtx, stage.Name(), err)
			}
			return err
		}

		// Si el paquete ya fue consumido por una estación terminal
		if pCtx.Handled {
			break
		}
	}

	return nil
}

// DefaultDeadLetterHandler libera búferes Zero-Copy garantizando no fugas de memoria.
type DefaultDeadLetterHandler struct{}

func (d *DefaultDeadLetterHandler) HandleDeadLetter(ctx context.Context, pCtx *interfaces.PacketContext, stageName string, err error) {
	if pCtx != nil && pCtx.Buffer != nil {
		pCtx.Buffer.Release()
		pCtx.Buffer = nil
	}
}
