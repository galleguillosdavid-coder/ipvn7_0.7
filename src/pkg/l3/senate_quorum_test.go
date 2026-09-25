package l3

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"testing"
	"time"
)

func TestSenateQuorumEngine(t *testing.T) {
	engine := NewSenateQuorumEngine()

	// 1. Create two keypairs for agents
	pub1, priv1, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key 1: %v", err)
	}
	pub2, priv2, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key 2: %v", err)
	}

	agent1 := "did:ipvn7:agent:001"
	agent2 := "did:ipvn7:agent:002"

	engine.RegisterVoter(agent1, pub1, 50.0)
	engine.RegisterVoter(agent2, pub2, 50.0)

	proposal := &SenateProposal{
		ID:        "PROP-SEC-001",
		Category:  CategorySecurityPatch,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Votes:     make(map[string]*AgentVote),
	}

	// 2. Sign and cast vote for agent1 (SUPPORT)
	msg1 := []byte(fmt.Sprintf("%s:%s:%s:%.4f", proposal.ID, agent1, StanceSupport, 50.0))
	sig1 := ed25519.Sign(priv1, msg1)
	vote1 := &AgentVote{
		ProposalID:   proposal.ID,
		AgentDID:     agent1,
		Stance:       StanceSupport,
		Weight:       50.0,
		SignatureHex: hex.EncodeToString(sig1),
		Timestamp:    time.Now(),
	}

	if err := engine.VerifyVoteSignature(vote1); err != nil {
		t.Fatalf("Vote 1 signature verification failed: %v", err)
	}
	proposal.Votes[agent1] = vote1

	// Quorum check with only 1 voter on SecurityPatch (requires 2 voters)
	eval1 := engine.EvaluateQuorum(proposal)
	if eval1.QuorumReached {
		t.Fatalf("Quorum should not have been reached with only 1 voter")
	}

	// 3. Sign and cast vote for agent2 (SUPPORT)
	msg2 := []byte(fmt.Sprintf("%s:%s:%s:%.4f", proposal.ID, agent2, StanceSupport, 50.0))
	sig2 := ed25519.Sign(priv2, msg2)
	vote2 := &AgentVote{
		ProposalID:   proposal.ID,
		AgentDID:     agent2,
		Stance:       StanceSupport,
		Weight:       50.0,
		SignatureHex: hex.EncodeToString(sig2),
		Timestamp:    time.Now(),
	}

	if err := engine.VerifyVoteSignature(vote2); err != nil {
		t.Fatalf("Vote 2 signature verification failed: %v", err)
	}
	proposal.Votes[agent2] = vote2

	// Quorum check with 2 voters
	eval2 := engine.EvaluateQuorum(proposal)
	if !eval2.QuorumReached || !eval2.Passed {
		t.Fatalf("Expected proposal to pass with 100%% support, got: %+v", eval2)
	}
	if eval2.Status != StatusApproved {
		t.Fatalf("Expected status StatusApproved, got %v", eval2.Status)
	}

	// 4. Test Human Veto
	proposal.Votes[agent2].HumanVeto = true
	proposal.Votes[agent2].Justification = "Manual override by human operator"
	evalVeto := engine.EvaluateQuorum(proposal)
	if evalVeto.Status != StatusVetoed {
		t.Fatalf("Expected status StatusVetoed, got %v", evalVeto.Status)
	}
}
