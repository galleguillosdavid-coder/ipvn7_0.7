package l1_test

import (
	"testing"

	"ipvn7/pkg/interfaces"
	"ipvn7/pkg/l1"
)

// Certificación de interfaces L1 en tiempo de compilación (Axioma de Contratos Canónicos v0.7)
var (
	_ interfaces.PacketBuffer       = (*l1.PacketBuffer)(nil)
	_ interfaces.BufferPoolProvider = (*l1.BufferPool)(nil)
	_ interfaces.ZTNAFirewall       = (*l1.ZTNAFirewall)(nil)
)

func TestL1InterfaceConformance(t *testing.T) {
	bp := l1.NewBufferPool()
	var bpp interfaces.BufferPoolProvider = bp
	buf := bpp.Acquire(64)
	if buf == nil {
		t.Fatal("BufferPoolProvider.Acquire retornó nil")
	}
	buf.SetLen(32)
	if len(buf.Data()) != 32 {
		t.Fatalf("PacketBuffer.Data() longitud incorrecta: %d", len(buf.Data()))
	}
	buf.Release()

	fw := l1.NewZTNAFirewall(true)
	var ztna interfaces.ZTNAFirewall = fw
	ztna.AllowPeer("did:ipvn7:testpeer", []uint16{7001})
	if !ztna.IsAllowed("did:ipvn7:testpeer") {
		t.Fatal("ZTNAFirewall.IsAllowed retornó false para par autorizado")
	}
}
