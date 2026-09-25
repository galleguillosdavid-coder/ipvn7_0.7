package l1

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"ipvn7/pkg/l0"
)

var (
	ErrTUNRouterClosed  = errors.New("tun_router: el router de interfaz virtual está cerrado")
	ErrInvalidIPPayload = errors.New("tun_router: trama IP inválida o truncada")
)

// MeshPacketForwarder define la interfaz para inyectar paquetes provenientes de TUN a la malla P2P
type MeshPacketForwarder interface {
	ForwardToMesh(targetDID string, ipPacket []byte) error
}

// TUNRouter coordina el intercambio bidireccional entre la interfaz virtual TUN y la malla IPVN7
type TUNRouter struct {
	mu          sync.RWMutex
	adapter     TunAdapter
	forwarder   MeshPacketForwarder
	ctx         context.Context
	cancel      context.CancelFunc
	active      atomic.Bool
	packetsRx   atomic.Uint64
	packetsTx   atomic.Uint64
	bytesRx     atomic.Uint64
	bytesTx     atomic.Uint64
}

// NewTUNRouter inicializa el enrutador de tráfico de la interfaz de red virtual
func NewTUNRouter(adapter TunAdapter, forwarder MeshPacketForwarder) *TUNRouter {
	ctx, cancel := context.WithCancel(context.Background())
	tr := &TUNRouter{
		adapter:   adapter,
		forwarder: forwarder,
		ctx:       ctx,
		cancel:    cancel,
	}
	tr.active.Store(true)
	go tr.readLoop()
	return tr
}

// readLoop extrae paquetes continuamente desde el adaptador TUN y los remite a la malla
func (tr *TUNRouter) readLoop() {
	for {
		select {
		case <-tr.ctx.Done():
			return
		default:
			if tr.adapter == nil {
				return
			}
			pkt, err := tr.adapter.ReadPacket()
			if err != nil {
				// Pausa ante adaptador cerrado o cola vacía
				time.Sleep(10 * time.Millisecond)
				continue
			}
			if len(pkt) < 20 {
				continue
			}
			tr.handleInboundFromOS(pkt)
		}
	}
}

// handleInboundFromOS parsea el paquete IP capturado del SO y determina el destino DID
func (tr *TUNRouter) handleInboundFromOS(raw []byte) {
	version := raw[0] >> 4
	var destDID string

	if version == 4 && len(raw) >= 20 {
		dstIP := net.IP(raw[16:20])
		destDID = ResolveIPToDID(dstIP)
	} else if version == 6 && len(raw) >= 40 {
		dstIP := net.IP(raw[24:40])
		destDID = ResolveIPToDID(dstIP)
	} else {
		return
	}

	tr.packetsRx.Add(1)
	tr.bytesRx.Add(uint64(len(raw)))

	if tr.forwarder != nil && destDID != "" {
		_ = tr.forwarder.ForwardToMesh(destDID, raw)
	}
}

// InjectFromMesh inyecta una trama IP recibida de la malla hacia el sistema operativo a través de TUN
func (tr *TUNRouter) InjectFromMesh(ipPacket []byte) error {
	if !tr.active.Load() || tr.adapter == nil {
		return ErrTUNRouterClosed
	}
	if len(ipPacket) < 20 {
		return ErrInvalidIPPayload
	}

	err := tr.adapter.WritePacket(ipPacket)
	if err == nil {
		tr.packetsTx.Add(1)
		tr.bytesTx.Add(uint64(len(ipPacket)))
	}
	return err
}

// ResolveIPToDID mapea deterministamente una IP virtual (10.7.X.Y o fd07::/64) al identificador DID
func ResolveIPToDID(ip net.IP) string {
	if ip == nil {
		return ""
	}
	if v4 := ip.To4(); v4 != nil {
		return fmt.Sprintf("did:ipvn7:node-v4-%d-%d-%d-%d", v4[0], v4[1], v4[2], v4[3])
	}
	chk := l0.FastPacketChecksum(ip)
	return fmt.Sprintf("did:ipvn7:node-v6-%x", chk[:8])
}

// Stats retorna las métricas en tiempo real de tránsito de paquetes del adaptador TUN
func (tr *TUNRouter) Stats() (rxPkts, txPkts, rxBytes, txBytes uint64) {
	return tr.packetsRx.Load(), tr.packetsTx.Load(), tr.bytesRx.Load(), tr.bytesTx.Load()
}

// Close apaga el enrutador y desconecta el adaptador
func (tr *TUNRouter) Close() error {
	if tr.active.CompareAndSwap(true, false) {
		tr.cancel()
		if tr.adapter != nil {
			return tr.adapter.Close()
		}
	}
	return nil
}
