// Package core implementa el endpoint REST para ejecución de intenciones en lenguaje natural.
package core

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"ipvn7/pkg/components/os_runner"
	vehbridge "ipvn7/pkg/components/vehicle_robot_bridge"
	"ipvn7/pkg/l3"
)

// IntentOrchestrator motor de resolución y despacho de intenciones en el núcleo
type IntentOrchestrator struct {
	osRunner  *osrunner.OSRunnerComponent
	vehBridge *vehbridge.VehicleRobotBridge
}

// NewIntentOrchestrator inicializa el orquestador
func NewIntentOrchestrator(osRunner *osrunner.OSRunnerComponent, vehBridge *vehbridge.VehicleRobotBridge) *IntentOrchestrator {
	return &IntentOrchestrator{
		osRunner:  osRunner,
		vehBridge: vehBridge,
	}
}

// Dispatch ejecuta localmente la intención si es aplicable
func (io *IntentOrchestrator) Dispatch(intent l3.ParsedIntent) (*l3.IntentExecutionResponse, error) {
	now := time.Now()
	res := &l3.IntentExecutionResponse{
		Category:   intent.Category,
		Action:     intent.Action,
		TargetNode: intent.TargetNode,
		Timestamp:  now,
	}

	switch intent.Category {
	case l3.CategoryHardwareOS:
		if io.osRunner == nil {
			exec := osrunner.NewPlatformExecutor()
			io.osRunner = osrunner.NewOSRunnerComponent("orchestrator_os_runner", "OS Runner", nil, exec)
		}
		actReq := &osrunner.ActionRequest{
			Action:    intent.Action,
			Params:    intent.Parameters,
			CallerDID: "intent:human_user",
		}
		actResp := io.osRunner.ExecuteAction(actReq)
		res.Success = actResp.Success
		res.HumanMessage = actResp.Message
		res.Details = actResp.Details
		return res, nil

	case l3.CategoryRobotDrone, l3.CategoryVehicle:
		if io.vehBridge == nil {
			io.vehBridge = vehbridge.NewVehicleRobotBridge("orchestrator_veh_bridge", "Veh Bridge", nil)
		}
		cmd := &vehbridge.ControlCommand{
			DeviceID:  intent.TargetNode,
			Action:    intent.Action,
			CallerDID: "intent:human_user",
		}
		cmdRes := io.vehBridge.ExecuteControlCommand(cmd)
		res.Success = cmdRes.Success
		res.HumanMessage = cmdRes.Message
		return res, nil

	default:
		return nil, errors.New("intención no reconocida o categoría sin ejecutor")
	}
}

// IntentAPIRequest estructura de entrada para la API de intenciones
type IntentAPIRequest struct {
	Prompt     string `json:"prompt"`      // Texto en lenguaje natural (ej. "suspende el equipo", "bloquea la pantalla")
	TargetNode string `json:"target_node"` // "local", "notebook", DID
}

// handleIntentProcess procesa solicitudes en lenguaje cotidiano y las ejecuta
func (s *CoreServer) handleIntentProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido (use POST)", http.StatusMethodNotAllowed)
		return
	}

	var req IntentAPIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Prompt == "" {
		http.Error(w, "El campo 'prompt' no puede estar vacío", http.StatusBadRequest)
		return
	}

	// 1. Interpretar semánticamente el texto (0 tokens / offline)
	parsed := l3.ParseNaturalLanguage(req.Prompt)
	if req.TargetNode != "" && req.TargetNode != "local" {
		parsed.TargetNode = req.TargetNode
	}

	// 2. Orquestar la ejecución
	orch := NewIntentOrchestrator(nil, nil)
	resp, err := orch.Dispatch(parsed)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
			"parsed":  parsed,
		})
		return
	}

	// 3. Emitir evento de auditoría en la malla
	s.gateway.PublishEvent(MeshEvent{
		Type:      EventType("intent:executed"),
		Timestamp: resp.Timestamp.UnixNano(),
		Source:    "intent_api",
		Payload: map[string]interface{}{
			"prompt":   req.Prompt,
			"action":   resp.Action,
			"target":   resp.TargetNode,
			"success":  resp.Success,
		},
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
