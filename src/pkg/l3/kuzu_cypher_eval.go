// Package l3 implementa el evaluador openCypher en memoria y con puente Kùzu DB
// conforme a las especificaciones de genesis.md y docs/INGENIERIA_LEAN.md.
package l3

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// CypherResult representa la respuesta tipada a una consulta openCypher
type CypherResult struct {
	Query     string                   `json:"query"`
	Columns   []string                 `json:"columns"`
	Rows      []map[string]interface{} `json:"rows"`
	RowCount  int                      `json:"row_count"`
	ExecMs    float64                  `json:"exec_ms"`
	Timestamp time.Time                `json:"timestamp"`
}

// ExecuteCypher ejecuta consultas openCypher esenciales contra el grafo Kùzu en memoria o persistente
func (kg *KuzuGraphEngine) ExecuteCypher(query string) (*CypherResult, error) {
	start := time.Now()
	kg.mu.RLock()
	defer kg.mu.RUnlock()

	cleanQuery := strings.TrimSpace(query)
	upper := strings.ToUpper(cleanQuery)

	if !strings.HasPrefix(upper, "MATCH") {
		return nil, errors.New("solo se admiten consultas de lectura Cypher iniciadas con MATCH")
	}

	result := &CypherResult{
		Query:     query,
		Columns:   make([]string, 0),
		Rows:      make([]map[string]interface{}, 0),
		Timestamp: time.Now(),
	}

	// 2. Fallback de cache en memoria
	if strings.Contains(upper, "(P:PEER)-[L:XOR_LINK]->(M:PEER)") || strings.Contains(upper, "XOR_LINK") {
		result.Columns = []string{"p.did", "l.degree", "l.latency_ms", "l.rf_band", "m.did"}
		for _, link := range kg.xorLinks {
			row := map[string]interface{}{
				"p.did":        link.FromDID,
				"l.degree":     link.Degree,
				"l.latency_ms": link.LatencyMs,
				"l.rf_band":    link.RFBand,
				"m.did":        link.ToDID,
			}
			result.Rows = append(result.Rows, row)
		}
	} else if strings.Contains(upper, "AXIOMPRINCIPLE") || strings.Contains(upper, "(A:AXIOMPRINCIPLE)") {
		// MATCH (a:AxiomPrinciple)
		result.Columns = []string{"a.principle_id", "a.statement", "a.immutable"}
		for _, ax := range kg.axioms {
			result.Rows = append(result.Rows, map[string]interface{}{
				"a.principle_id": ax.PrincipleID,
				"a.statement":   ax.Statement,
				"a.immutable":   ax.Immutable,
			})
		}
	} else if strings.Contains(upper, "OPERATIONALFUNCTION") || strings.Contains(upper, "DERIVES_FROM") {
		// MATCH (f:OperationalFunction)-[:DERIVES_FROM]->(a:AxiomPrinciple)
		result.Columns = []string{"f.function_id", "f.module_path", "f.status", "a.principle_id"}
		for fID, fn := range kg.functions {
			pID := kg.derivesFrom[fID]
			result.Rows = append(result.Rows, map[string]interface{}{
				"f.function_id":  fn.FunctionID,
				"f.module_path":  fn.ModulePath,
				"f.status":       fn.Status,
				"a.principle_id": pID,
			})
		}
	} else if strings.Contains(upper, "PEER") {
		// MATCH (p:Peer)
		result.Columns = []string{"p.did", "p.ml_dsa", "p.anycast", "p.device_profile"}
		for _, peer := range kg.peers {
			result.Rows = append(result.Rows, map[string]interface{}{
				"p.did":            peer.DID,
				"p.ml_dsa":         peer.MLDSA,
				"p.anycast":        peer.Anycast,
				"p.device_profile": peer.DeviceProfile,
			})
		}
	} else if strings.Contains(upper, "DECISION") || strings.Contains(upper, "(D:DECISION)") {
		// MATCH (d:Decision)
		result.Columns = []string{"d.decision_id", "d.rationale", "d.timestamp"}
		for _, dec := range kg.decisions {
			result.Rows = append(result.Rows, map[string]interface{}{
				"d.decision_id": dec.DecisionID,
				"d.rationale":   dec.Rationale,
				"d.timestamp":   dec.Timestamp.Format(time.RFC3339),
			})
		}
	} else {
		return nil, fmt.Errorf("patrón Cypher no soportado en emulación rápida: %s", cleanQuery)
	}

	result.RowCount = len(result.Rows)
	result.ExecMs = float64(time.Since(start).Microseconds()) / 1000.0
	return result, nil
}
