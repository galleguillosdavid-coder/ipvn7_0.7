package l0

import (
	"testing"
	"time"
)

func TestDelegationChain_ValidMultiHop(t *testing.T) {
	// Nodo Raiz (Dueño del nodo físico)
	rootId, err := GenerateIdentity()
	if err != nil {
		t.Fatalf("error generando id: %v", err)
	}

	// Agente 1 (Orquestador)
	agent1, _ := GenerateIdentity()

	// Agente 2 (Sub-agente de tarea)
	agent2, _ := GenerateIdentity()

	// 1. Raiz delega a Agente 1 (Escalated: vpn:control)
	link1, err := CreateDelegationLink(rootId, agent1.DID(), "vpn:control", TierEscalated, 10*time.Minute)
	if err != nil {
		t.Fatalf("error creando link 1: %v", err)
	}

	// 2. Agente 1 delega a Agente 2 (Atenuado a Ordinary: network:read)
	link2, err := CreateDelegationLink(agent1, agent2.DID(), "network:read", TierOrdinary, 5*time.Minute)
	if err != nil {
		t.Fatalf("error creando link 2: %v", err)
	}

	chain := AgenticPrincipalChain{
		RootDID: rootId.DID(),
		Links:   []DelegationLink{*link1, *link2},
	}

	tier, err := chain.VerifyChain()
	if err != nil {
		t.Fatalf("esperada cadena valida, obtenido error: %v", err)
	}
	if tier != TierOrdinary {
		t.Errorf("esperado tier atenuado Ordinary, obtenido: %s", tier)
	}
}

func TestDelegationChain_PrivilegeEscalationRejection(t *testing.T) {
	rootId, _ := GenerateIdentity()
	agent1, _ := GenerateIdentity()
	agent2, _ := GenerateIdentity()

	// Raiz delega permiso ordinario
	link1, _ := CreateDelegationLink(rootId, agent1.DID(), "network:read", TierOrdinary, 10*time.Minute)
	// Agente 1 intenta escalar a permiso escalado (ataque de escalacion)
	link2, _ := CreateDelegationLink(agent1, agent2.DID(), "vpn:control", TierEscalated, 5*time.Minute)

	chain := AgenticPrincipalChain{
		RootDID: rootId.DID(),
		Links:   []DelegationLink{*link1, *link2},
	}

	tier, err := chain.VerifyChain()
	if err == nil || tier != TierDeny {
		t.Errorf("esperado rechazo por intento de escalamiento, obtenido tier: %s, err: %v", tier, err)
	}
}

func TestDelegationChain_ExpiredAndTampered(t *testing.T) {
	rootId, _ := GenerateIdentity()
	agent1, _ := GenerateIdentity()

	// Link expirado (-1 segundo)
	linkExp, _ := CreateDelegationLink(rootId, agent1.DID(), "vpn:control", TierEscalated, -1*time.Second)
	chainExp := AgenticPrincipalChain{
		RootDID: rootId.DID(),
		Links:   []DelegationLink{*linkExp},
	}
	tierExp, errExp := chainExp.VerifyChain()
	if errExp == nil || tierExp != TierDeny {
		t.Errorf("esperado rechazo de token expirado")
	}

	// Link con firma alterada
	linkTampered, _ := CreateDelegationLink(rootId, agent1.DID(), "vpn:control", TierEscalated, 10*time.Minute)
	linkTampered.Signature = "0000000000"
	chainTampered := AgenticPrincipalChain{
		RootDID: rootId.DID(),
		Links:   []DelegationLink{*linkTampered},
	}
	tierTamp, errTamp := chainTampered.VerifyChain()
	if errTamp == nil || tierTamp != TierDeny {
		t.Errorf("esperado rechazo de firma alterada")
	}
}

func TestSpendingMandate_AP2Andx402(t *testing.T) {
	payer, _ := GenerateIdentity()
	merchant, _ := GenerateIdentity()

	// 1. Mandato legítimo
	mandate, err := CreateSpendingMandate(payer, merchant.DID(), "inference:gpu_sec", 120, 10*time.Minute)
	if err != nil {
		t.Fatalf("error creando mandato: %v", err)
	}
	if !mandate.Verify() {
		t.Fatalf("esperado mandato valido")
	}

	// 2. Mandato alterado (intento de fraude en unidades)
	mandate.MaxUnits = 99999
	if mandate.Verify() {
		t.Errorf("esperado rechazo de mandato con unidades alteradas")
	}

	// 3. Mandato expirado
	mandateExp, _ := CreateSpendingMandate(payer, merchant.DID(), "inference:gpu_sec", 120, -1*time.Minute)
	if mandateExp.Verify() {
		t.Errorf("esperado rechazo de mandato expirado")
	}
}

