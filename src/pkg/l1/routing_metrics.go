package l1

import (
	"crypto/ed25519"
	"errors"
)

// CalculateHealthScore calcula el puntaje de mérito de 0 a 100 y el estado FSM
func CalculateHealthScore(pacingCompliance, jitterMs, lossRate float64) (float64, int) {
	jitterPen := (jitterMs / 50.0) * 30.0
	if jitterPen > 30.0 {
		jitterPen = 30.0
	}
	lossPen := lossRate * 40.0
	if lossPen > 40.0 {
		lossPen = 40.0
	}

	score := (pacingCompliance * 30.0) + (30.0 - jitterPen) + (40.0 - lossPen)
	if score < 0 {
		score = 0
	} else if score > 100 {
		score = 100
	}

	state := HealthStateHealthy
	if score < 25.0 || lossRate >= 0.4 {
		state = HealthStateUnreachable
	} else if score < 50.0 || lossRate >= 0.20 || jitterMs > 60.0 {
		state = HealthStateUnstable
	} else if score < 75.0 || lossRate >= 0.05 || jitterMs > 25.0 {
		state = HealthStateDegraded
	}
	return score, state
}

// CalculateHybridCost calcula la función de coste híbrida 2D (distancia XOR logarítmica ponderada con latencia física y salud)
func CalculateHybridCost(peer *PeerNode, pubDest ed25519.PublicKey, numRings int, alpha float64) float64 {
	distRing := XORKeyDistanceN(peer.PublicKey, pubDest, numRings)
	normalizedXOR := float64(distRing) / float64(numRings)

	normLat := peer.Locator.LatencyMs / 250.0
	if normLat > 1.0 {
		normLat = 1.0
	}

	healthBonus := (peer.HealthScore / 100.0) * 0.2
	return ((1.0 - alpha) * normalizedXOR) + (alpha * normLat) - healthBonus
}

// FindBestPeerCandidate evalúa la métrica híbrida sobre los pares registrados para seleccionar el siguiente salto óptimo
func FindBestPeerCandidate(peers map[string]*PeerNode, pubDest ed25519.PublicKey, numRings int, alpha float64) (*PeerNode, error) {
	var bestPeer *PeerNode
	bestCost := 999999.0

	for _, peer := range peers {
		// Descartar pares inaccesibles o en cuarentena
		if peer.HealthState == HealthStateQuarantined || peer.HealthState == HealthStateUnreachable {
			continue
		}

		cost := CalculateHybridCost(peer, pubDest, numRings, alpha)
		if cost < bestCost {
			bestCost = cost
			bestPeer = peer
		}
	}

	// Fallback si no hay candidato por coste: retornar el par más saludable disponible
	if bestPeer == nil {
		for _, peer := range peers {
			if bestPeer == nil || peer.HealthScore > bestPeer.HealthScore {
				bestPeer = peer
			}
		}
	}

	if bestPeer == nil {
		return nil, errors.New("no hay rutas disponibles hacia el destino")
	}

	return bestPeer, nil
}
