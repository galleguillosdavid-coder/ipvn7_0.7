package l2

import (
	"testing"
	"time"
)

func TestSplitBrainGuard(t *testing.T) {
	timeout := 200 * time.Millisecond
	threshold := 0.50 // al menos 50%
	guard := NewSplitBrainGuard(timeout, threshold)

	peerA := "did:ipvn7:node:peer-a"
	peerB := "did:ipvn7:node:peer-b"
	peerC := "did:ipvn7:node:peer-c"
	peerD := "did:ipvn7:node:peer-d"

	guard.RegisterPeer(peerA)
	guard.RegisterPeer(peerB)
	guard.RegisterPeer(peerC)
	guard.RegisterPeer(peerD)

	now := time.Now()
	// Record heartbeats for peerA, peerB, peerC (3/4 = 75% > 50%)
	guard.RecordHeartbeatAt(peerA, now)
	guard.RecordHeartbeatAt(peerB, now)
	guard.RecordHeartbeatAt(peerC, now)

	state, msg := guard.AuditLiveness(now)
	if state != StateHealthy {
		t.Fatalf("Expected StateHealthy, got %v: %s", state, msg)
	}
	if !guard.IsSafeToWrite() {
		t.Fatalf("Expected safe to write in healthy state")
	}

	// Advance time past timeout: peerB and peerC go silent, only peerA remains (1/4 = 25% < 50%)
	nowLater := now.Add(300 * time.Millisecond)
	// Only peerA sends heartbeat at nowLater
	guard.RecordHeartbeatAt(peerA, nowLater)

	state2, msg2 := guard.AuditLiveness(nowLater)
	if state2 != StatePartitioned {
		t.Fatalf("Expected StatePartitioned, got %v: %s", state2, msg2)
	}
	if guard.IsSafeToWrite() {
		t.Fatalf("Expected NOT safe to write during partition")
	}

	// Restore heartbeats for peerB and peerC at nowLater
	guard.RecordHeartbeatAt(peerB, nowLater)
	guard.RecordHeartbeatAt(peerC, nowLater)

	state3, _ := guard.AuditLiveness(nowLater)
	if state3 != StateRecovering {
		t.Fatalf("Expected StateRecovering upon restoring quorum, got %v", state3)
	}

	// Advance past recovery window while keeping heartbeats active
	tHealthy := nowLater.Add(6 * time.Second)
	guard.RecordHeartbeatAt(peerA, tHealthy)
	guard.RecordHeartbeatAt(peerB, tHealthy)
	guard.RecordHeartbeatAt(peerC, tHealthy)

	state4, msg4 := guard.AuditLiveness(tHealthy)
	if state4 != StateHealthy {
		t.Fatalf("Expected StateHealthy after recovery window, got %v (%s)", state4, msg4)
	}
	if !guard.IsSafeToWrite() {
		t.Fatalf("Expected safe to write once recovered to healthy")
	}
}
