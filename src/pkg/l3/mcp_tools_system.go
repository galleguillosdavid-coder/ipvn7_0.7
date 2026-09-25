package l3

import (
	"encoding/json"
	"fmt"

	"ipvn7/pkg/components/os_runner"
)

// ExecuteSystemControlTool despacha acciones sobre el hardware y sistema operativo local
func (s *MCPServer) ExecuteSystemControlTool(name string, args map[string]interface{}) (string, error) {
	exec := osrunner.NewPlatformExecutor()
	runner := osrunner.NewOSRunnerComponent("mcp_os_runner", "MCP Hardware Runner", nil, exec)

	action, _ := args["action"].(string)
	if action == "" {
		action = name
	}

	params := make(map[string]string)
	if pMap, ok := args["params"].(map[string]interface{}); ok {
		for k, v := range pMap {
			params[k] = fmt.Sprintf("%v", v)
		}
	}
	if drive, ok := args["drive"].(string); ok {
		params["drive"] = drive
	}

	req := &osrunner.ActionRequest{
		Action:    action,
		Params:    params,
		CallerDID: "mcp:ai_agent",
	}

	resp := runner.ExecuteAction(req)
	data, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error serializando respuesta de sistema: %w", err)
	}

	return string(data), nil
}

func getSystemTools() []MCPTool {
	return []MCPTool{
		{
			Name:        "ipvn7_system_control",
			Description: "Control profundo de hardware y SO: energía (sleep/reboot/shutdown/lock/monitor_off), procesos (kill_heavy/list), batería, térmica y expulsión de discos",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"action": map[string]interface{}{
						"type":        "string",
						"description": "Acción a ejecutar: 'power:sleep', 'power:reboot', 'power:shutdown', 'power:lock', 'power:monitor_off', 'process:kill_heavy', 'process:list', 'hardware:battery', 'hardware:thermal', 'system:clean_recycle_bin', 'storage:safe_eject'",
						"enum": []string{
							"power:sleep", "power:reboot", "power:shutdown", "power:lock",
							"power:monitor_off", "process:kill_heavy", "process:list",
							"hardware:battery", "hardware:thermal", "system:clean_recycle_bin",
							"storage:safe_eject",
						},
					},
					"drive": map[string]interface{}{
						"type":        "string",
						"description": "Letra de unidad para expulsión (ej. 'E:' o 'F:')",
					},
				},
				"required": []string{"action"},
			},
		},
	}
}
