package l1

import (
	"testing"
	"time"
)

func TestEgressCapability_CBORSerialization(t *testing.T) {
	orig := &EgressCapability{
		CanExit:        true,
		BandwidthMbps:  300,
		ActiveSessions: 5,
		CountryCode:    "CL",
		UpdatedAt:      time.Now().Truncate(time.Second),
	}

	encoded, err := EncodeEgressCapability(orig)
	if err != nil {
		t.Fatalf("falla al codificar CBOR: %v", err)
	}

	decoded, err := DecodeEgressCapability(encoded)
	if err != nil {
		t.Fatalf("falla al decodificar CBOR: %v", err)
	}

	if decoded.CanExit != orig.CanExit {
		t.Errorf("CanExit discrepancia: got %v, want %v", decoded.CanExit, orig.CanExit)
	}
	if decoded.BandwidthMbps != orig.BandwidthMbps {
		t.Errorf("BandwidthMbps discrepancia: got %d, want %d", decoded.BandwidthMbps, orig.BandwidthMbps)
	}
	if decoded.ActiveSessions != orig.ActiveSessions {
		t.Errorf("ActiveSessions discrepancia: got %d, want %d", decoded.ActiveSessions, orig.ActiveSessions)
	}
	if decoded.CountryCode != orig.CountryCode {
		t.Errorf("CountryCode discrepancia: got %s, want %s", decoded.CountryCode, orig.CountryCode)
	}
}

func TestCalculateGatewayScore(t *testing.T) {
	// Gateway A: RTT 10ms, Loss 0%, 2 sesiones
	scoreA := CalculateGatewayScore(10.0, 0.0, 2)
	// Gateway B: RTT 50ms, Loss 5%, 20 sesiones
	scoreB := CalculateGatewayScore(50.0, 0.05, 20)

	if scoreA >= scoreB {
		t.Errorf("Gateway A debería tener mejor (menor) score que B: A=%f, B=%f", scoreA, scoreB)
	}
}

func TestEgressRegistry(t *testing.T) {
	reg := NewEgressRegistry()
	didA := "did:ipvn7:nodo_a"
	didB := "did:ipvn7:nodo_b"

	reg.Set(didA, &EgressCapability{CanExit: true, BandwidthMbps: 500})
	reg.Set(didB, &EgressCapability{CanExit: false, BandwidthMbps: 50})

	gwList := reg.ListGateways()
	if len(gwList) != 1 || gwList[0] != didA {
		t.Errorf("ListGateways inesperado: got %v, want [%s]", gwList, didA)
	}

	capA, ok := reg.Get(didA)
	if !ok || !capA.CanExit {
		t.Errorf("Get(didA) falló o CanExit es falso")
	}

	reg.Remove(didA)
	if _, ok := reg.Get(didA); ok {
		t.Errorf("Remove(didA) falló, el par todavía existe")
	}
}
