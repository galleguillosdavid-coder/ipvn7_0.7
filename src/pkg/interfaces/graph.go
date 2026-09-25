// Package interfaces define los contratos raíz universales e inmutables del protocolo IPVN7 (v0.7).
package interfaces

import (
	"context"
	"time"
)

// GraphNodeEntity representa un vértice tipado dentro del grafo relacional Kùzu.
type GraphNodeEntity interface {
	// ID retorna el identificador único del nodo en el grafo.
	ID() string

	// Label retorna la etiqueta tipada (ej: "Component", "Interface", "ProcessFlow", "Resource").
	Label() string

	// Name retorna el nombre legible para visualización y consultas Cypher.
	Name() string

	// Layer retorna la capa arquitectónica a la que pertenece (ej: "L0", "L1", "Core", "App").
	Layer() string

	// Properties retorna los atributos semánticos asociados al vértice.
	Properties() map[string]interface{}
}

// GraphEdgeEntity representa una relación dirigida tipada entre dos vértices del grafo Kùzu.
type GraphEdgeEntity interface {
	// ID retorna el identificador único de la arista.
	ID() string

	// SourceID retorna el ID del vértice de origen.
	SourceID() string

	// TargetID retorna el ID del vértice de destino.
	TargetID() string

	// Type retorna la relación semántica (ej: "TRANSFERS_DATA_TO", "IMPLEMENTS", "REQUIRES").
	Type() string

	// Properties retorna atributos cuantitativos o cualitativos de la relación.
	Properties() map[string]interface{}
}

// CypherExecutionResult contiene los resultados de una consulta estructurada openCypher.
type CypherExecutionResult struct {
	Query        string                   `json:"query"`
	Columns      []string                 `json:"columns"`
	Rows         []map[string]interface{} `json:"rows"`
	RowCount     int                      `json:"row_count"`
	MatchedNodes []string                 `json:"matched_nodes"`
	MatchedEdges []string                 `json:"matched_edges"`
	ExecMs       float64                  `json:"exec_ms"`
	Timestamp    time.Time                `json:"timestamp"`
}

// KnowledgeGraphEngine formaliza la ejecución Cypher y el análisis topológico del sistema en Kùzu.
type KnowledgeGraphEngine interface {
	// ExecuteQuery procesa una sentencia Cypher y extrae nodos y aristas coincidentes.
	ExecuteQuery(ctx context.Context, cypher string) (*CypherExecutionResult, error)

	// RegisterNode inserta o actualiza un vértice de conocimiento en la topología activa.
	RegisterNode(node GraphNodeEntity)

	// RegisterEdge inserta o actualiza una relación semántica entre vértices existentes.
	RegisterEdge(edge GraphEdgeEntity)

	// ExportTopology retorna la totalidad de nodos y aristas estructurados con métricas de cobertura.
	ExportTopology() ([]GraphNodeEntity, []GraphEdgeEntity, map[string]interface{})

	// ValidateZeroOrphans verifica formalmente que ningún vértice se encuentre desconectado.
	ValidateZeroOrphans() (orphanCount int, coveragePerc float64)
}
