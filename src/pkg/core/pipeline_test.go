package core

import (
	"context"
	"net"
	"testing"
	"time"

	"ipvn7/pkg/interfaces"
	"ipvn7/pkg/l0"
	"ipvn7/pkg/l1"
)

type mockBuffer struct {
	data     []byte
	released bool
}

func (m *mockBuffer) RawSlice() []byte { return m.data }
func (m *mockBuffer) Data() []byte     { return m.data }
func (m *mockBuffer) SetLen(n int)     {}
func (m *mockBuffer) Capacity() int    { return len(m.data) }
func (m *mockBuffer) Retain()          {}
func (m *mockBuffer) Release()         { m.released = true }

type recordingDeadLetter struct {
	called    bool
	stage     string
	reasonErr error
}

func (r *recordingDeadLetter) HandleDeadLetter(ctx context.Context, pCtx *interfaces.PacketContext, stageName string, err error) {
	r.called = true
	r.stage = stageName
	r.reasonErr = err
	if pCtx.Buffer != nil {
		pCtx.Buffer.Release()
		pCtx.Buffer = nil
	}
}

func TestLinearPipeline_SuccessFlow(t *testing.T) {
	// Generar identidad local y de par remoto
	localID, err := l0.GenerateIdentity()
	if err != nil {
		t.Fatalf("Error generando identidad local: %v", err)
	}
	peerID, err := l0.GenerateIdentity()
	if err != nil {
		t.Fatalf("Error generando identidad remota: %v", err)
	}

	pkt := l0.NewPacket(l0.MsgTypeRoamingUpdate, peerID.DID(), localID.DID(), 1, make([]byte, 16), []byte("test"))
	if err := pkt.SignPacket(peerID); err != nil {
		t.Fatalf("Error firmando paquete: %v", err)
	}
	encoded, err := pkt.Encode()
	if err != nil {
		t.Fatalf("Error codificando paquete: %v", err)
	}

	buf := &mockBuffer{data: encoded}
	pCtx := &interfaces.PacketContext{
		Buffer:     buf,
		RemoteAddr: &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9999},
	}

	router := l1.NewKleinbergRouter(localID)
	fw := l1.NewZTNAFirewall(true)
	qos := l1.NewQoSManager()
	fsm := NewDeterministicNodeFSM(interfaces.NodeStateDisconnected)

	pipeline := NewLinearPipeline().
		AddStage(NewDecodeCBORStage()).
		AddStage(NewZTNAFilterStage(fw, 7777, nil)).
		AddStage(NewQoSFilterStage(qos)).
		AddStage(NewTopologyFSMStage(router, fw, nil, fsm))

	deadLetter := &recordingDeadLetter{}
	pipeline.SetDeadLetterHandler(deadLetter)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = pipeline.Execute(ctx, pCtx)
	if err != nil {
		t.Fatalf("Pipeline retornó error inesperado: %v", err)
	}

	if deadLetter.called {
		t.Fatalf("Dead letter no debió ser invocado en flujo exitoso")
	}

	if !pCtx.Handled {
		t.Errorf("El paquete debía marcarse como Handled tras roaming update")
	}
}

func TestLinearPipeline_DeadLetterOnCorruptPacket(t *testing.T) {
	buf := &mockBuffer{data: []byte{0x00, 0x01, 0x02, 0x03}} // CBOR inválido
	pCtx := &interfaces.PacketContext{
		Buffer:     buf,
		RemoteAddr: &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9999},
	}

	pipeline := NewLinearPipeline().
		AddStage(NewDecodeCBORStage())

	deadLetter := &recordingDeadLetter{}
	pipeline.SetDeadLetterHandler(deadLetter)

	ctx := context.Background()
	err := pipeline.Execute(ctx, pCtx)
	if err == nil {
		t.Fatalf("Se esperaba error con paquete corrupto")
	}

	if !deadLetter.called {
		t.Fatalf("DeadLetter debía ser invocado")
	}
	if deadLetter.stage != "decode_cbor" {
		t.Errorf("Esperada falla en decode_cbor, obtenida en: %s", deadLetter.stage)
	}
	if !buf.released {
		t.Errorf("Búfer debía ser liberado por el DeadLetter handler")
	}
}

func TestLinearPipeline_DeadLetterOnZTNADeny(t *testing.T) {
	id, _ := l0.GenerateIdentity()
	// Mensaje ordinario de datos proveniente de par no autorizado
	pkt := l0.NewPacket(l0.MsgTypeData, id.DID(), "did:ipvn7:dest", 1, make([]byte, 16), []byte("secreto"))
	encoded, _ := pkt.Encode()

	buf := &mockBuffer{data: encoded}
	pCtx := &interfaces.PacketContext{
		Buffer:     buf,
		RemoteAddr: &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9999},
	}

	fw := l1.NewZTNAFirewall(true) // Default-Deny estricto sin pares autorizados
	pipeline := NewLinearPipeline().
		AddStage(NewDecodeCBORStage()).
		AddStage(NewZTNAFilterStage(fw, 7777, nil))

	deadLetter := &recordingDeadLetter{}
	pipeline.SetDeadLetterHandler(deadLetter)

	err := pipeline.Execute(context.Background(), pCtx)
	if err == nil {
		t.Fatalf("Se esperaba rechazo por ZTNA Default-Deny")
	}

	if !deadLetter.called {
		t.Fatalf("Dead letter debía capturar el descarte ZTNA")
	}
	if deadLetter.stage != "ztna_filter" {
		t.Errorf("Esperado descarte en ztna_filter, ocurrido en: %s", deadLetter.stage)
	}
}

type noopStage struct{}

func (n *noopStage) Name() string { return "noop" }
func (n *noopStage) Process(ctx context.Context, pCtx *interfaces.PacketContext) error {
	return nil
}

func BenchmarkLinearPipeline_Execute(b *testing.B) {
	pipeline := NewLinearPipeline().
		AddStage(&noopStage{}).
		AddStage(&noopStage{}).
		AddStage(&noopStage{})

	ctx := context.Background()
	pCtx := &interfaces.PacketContext{}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		pCtx.Handled = false
		pCtx.Dropped = false
		_ = pipeline.Execute(ctx, pCtx)
	}
}

func BenchmarkPacketContext_AcquireRelease(b *testing.B) {
	buf := &mockBuffer{data: make([]byte, 1280)}
	addr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 7777}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		pCtx := AcquirePacketContext(buf, addr)
		ReleasePacketContext(pCtx)
	}
}


