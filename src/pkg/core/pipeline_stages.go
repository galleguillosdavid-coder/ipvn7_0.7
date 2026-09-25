package core

import (
	"context"
	"fmt"
	"net"
	"time"

	"ipvn7/pkg/interfaces"
	"ipvn7/pkg/l0"
	"ipvn7/pkg/l1"
	"ipvn7/pkg/l2"
)

// TelemetryDeadLetterHandler registra en telemetría y libera memoria Zero-Copy
// para datagramas descartados o corruptos en cualquier estación de la línea.
type TelemetryDeadLetterHandler struct {
	telemetry *l2.TelemetryRingBuffer
}

// NewTelemetryDeadLetterHandler crea el manejador con enlace al ring buffer.
func NewTelemetryDeadLetterHandler(telem *l2.TelemetryRingBuffer) *TelemetryDeadLetterHandler {
	return &TelemetryDeadLetterHandler{telemetry: telem}
}

func (h *TelemetryDeadLetterHandler) HandleDeadLetter(ctx context.Context, pCtx *interfaces.PacketContext, stageName string, err error) {
	if pCtx == nil {
		return
	}
	if h.telemetry != nil {
		pktSize := uint32(len(pCtx.RawData))
		if pktSize == 0 && pCtx.Buffer != nil {
			pktSize = uint32(len(pCtx.Buffer.Data()))
		}
		h.telemetry.RecordEvent(l2.EventDrop, pktSize, 0, 0)
	}
	if pCtx.Buffer != nil {
		pCtx.Buffer.Release()
		pCtx.Buffer = nil
	}
}

// DecodeCBORStage estación 1: deserialización y validación canónica CBOR RFC 8949.
type DecodeCBORStage struct{}

func NewDecodeCBORStage() *DecodeCBORStage {
	return &DecodeCBORStage{}
}

func (s *DecodeCBORStage) Name() string {
	return "decode_cbor"
}

func (s *DecodeCBORStage) Process(ctx context.Context, pCtx *interfaces.PacketContext) error {
	var data []byte
	if pCtx.Buffer != nil {
		data = pCtx.Buffer.Data()
	} else {
		data = pCtx.RawData
	}

	pkt, err := l0.DecodePacket(data)
	if err != nil {
		pCtx.Dropped = true
		pCtx.DropReason = fmt.Sprintf("cbor_decode_error: %v", err)
		return err
	}

	pCtx.Packet = pkt
	pCtx.SourceDID = pkt.SourceDID
	return nil
}

// ZTNAFilterStage estación 2: micro-segmentación y control Default-Deny.
type ZTNAFilterStage struct {
	firewall   *l1.ZTNAFirewall
	listenPort uint16
	gateway    *SmartComponentGateway
}

func NewZTNAFilterStage(fw *l1.ZTNAFirewall, port uint16, gw *SmartComponentGateway) *ZTNAFilterStage {
	return &ZTNAFilterStage{
		firewall:   fw,
		listenPort: port,
		gateway:    gw,
	}
}

func (s *ZTNAFilterStage) Name() string {
	return "ztna_filter"
}

func (s *ZTNAFilterStage) Process(ctx context.Context, pCtx *interfaces.PacketContext) error {
	pkt, ok := pCtx.Packet.(*l0.Packet)
	if !ok || pkt == nil {
		pCtx.Dropped = true
		pCtx.DropReason = "invalid_packet_type"
		return fmt.Errorf("paquete inválido para evaluación ZTNA")
	}

	// Roaming updates se admiten para handshake de descubrimiento mutuo
	if pkt.Type == l0.MsgTypeRoamingUpdate {
		return nil
	}

	decision, _ := s.firewall.EvaluateInbound(pCtx.SourceDID, s.listenPort)
	if decision != l1.DecisionAccept {
		pCtx.Dropped = true
		pCtx.DropReason = "ztna_default_deny"
		if s.gateway != nil {
			s.gateway.PublishEvent(MeshEvent{
				Type:      EventZTNAAlert,
				Timestamp: time.Now().UnixNano(),
				Source:    pCtx.SourceDID,
				Payload:   map[string]interface{}{"action": "drop", "reason": "default_deny"},
			})
		}
		return fmt.Errorf("acceso denegado por ZTNA para DID %s", pCtx.SourceDID)
	}

	return nil
}

// QoSFilterStage estación 3: amortiguación y Token Bucket de ancho de banda.
type QoSFilterStage struct {
	qos *l1.QoSManager
}

func NewQoSFilterStage(qos *l1.QoSManager) *QoSFilterStage {
	return &QoSFilterStage{qos: qos}
}

func (s *QoSFilterStage) Name() string {
	return "qos_filter"
}

func (s *QoSFilterStage) Process(ctx context.Context, pCtx *interfaces.PacketContext) error {
	pktLen := len(pCtx.RawData)
	if pktLen == 0 && pCtx.Buffer != nil {
		pktLen = len(pCtx.Buffer.Data())
	}

	allowed, _ := s.qos.EvaluatePacket(pCtx.SourceDID, l1.ClassControl, pktLen)
	if !allowed {
		pCtx.Dropped = true
		pCtx.DropReason = "qos_rate_limited"
		return fmt.Errorf("tasa de paquetes excedida para DID %s", pCtx.SourceDID)
	}
	return nil
}

// TopologyFSMStage estación 4: autómata de estados, resolución de roaming y topología Kleinberg.
type TopologyFSMStage struct {
	router   *l1.KleinbergRouter
	firewall *l1.ZTNAFirewall
	gateway  *SmartComponentGateway
	fsm      interfaces.DeterministicFSM
}

func NewTopologyFSMStage(r *l1.KleinbergRouter, fw *l1.ZTNAFirewall, gw *SmartComponentGateway, fsm interfaces.DeterministicFSM) *TopologyFSMStage {
	return &TopologyFSMStage{
		router:   r,
		firewall: fw,
		gateway:  gw,
		fsm:      fsm,
	}
}

func (s *TopologyFSMStage) Name() string {
	return "topology_fsm"
}

func (s *TopologyFSMStage) Process(ctx context.Context, pCtx *interfaces.PacketContext) error {
	pkt, ok := pCtx.Packet.(*l0.Packet)
	if !ok || pkt == nil {
		return nil
	}

	if pkt.Type == l0.MsgTypeRoamingUpdate {
		if udpAddr, ok := pCtx.RemoteAddr.(*net.UDPAddr); ok {
			if err := s.router.HandleRoamingUpdate(pkt, udpAddr); err == nil {
				s.firewall.AuthorizeDID(&l1.DIDPolicy{
					DID:           pCtx.SourceDID,
					AllowInbound:  true,
					AllowOutbound: true,
					AllowRelay:    true,
				})
				if s.gateway != nil {
					s.gateway.PublishEvent(MeshEvent{
						Type:      EventPeerJoined,
						Timestamp: time.Now().UnixNano(),
						Source:    pCtx.SourceDID,
						Payload:   map[string]interface{}{"endpoint": udpAddr.String()},
					})
				}
				if s.fsm != nil {
					_, _ = s.fsm.TriggerEvent(interfaces.EventNetworkSignal, udpAddr.String())
					_, _ = s.fsm.TriggerEvent(interfaces.EventDIDVerified, pCtx.SourceDID)
				}
				pCtx.Handled = true
			}
		}
	}

	return nil
}
