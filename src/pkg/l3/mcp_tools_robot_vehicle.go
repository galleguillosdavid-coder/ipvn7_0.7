package l3

import (
	"encoding/json"
	"fmt"
	"time"

	bridge "ipvn7/pkg/components/vehicle_robot_bridge"
)

// ExecuteRobotVehicleTool despacha control y telemetría sobre robots, drones, vehículos y maquinaria
func (s *MCPServer) ExecuteRobotVehicleTool(name string, args map[string]interface{}) (string, error) {
	b := bridge.NewVehicleRobotBridge("mcp_robot_veh_bridge", "MCP Vehicle & Robot Bridge", nil)

	action, _ := args["action"].(string)
	deviceID, _ := args["device_id"].(string)
	if deviceID == "" {
		deviceID = "drone_01"
	}

	params := make(map[string]interface{})
	if pMap, ok := args["params"].(map[string]interface{}); ok {
		params = pMap
	}

	switch action {
	case "telemetry":
		// Si se pide telemetría, retornar el estado del dispositivo
		if tDrone, ok := b.GetDroneTelemetry(deviceID); ok {
			data, _ := json.MarshalIndent(tDrone, "", "  ")
			return string(data), nil
		}
		if tVeh, ok := b.GetVehicleTelemetry(deviceID); ok {
			data, _ := json.MarshalIndent(tVeh, "", "  ")
			return string(data), nil
		}
		res := map[string]interface{}{
			"device_id": deviceID,
			"status":    "active_listener",
			"protocols": []string{"MAVLink_v2", "CAN_Bus_ISO11898", "OBD-II_SAEJ1979", "J1939"},
			"timestamp": time.Now(),
		}
		data, _ := json.MarshalIndent(res, "", "  ")
		return string(data), nil

	default:
		cmd := &bridge.ControlCommand{
			DeviceID:   deviceID,
			Action:     action,
			Parameters: params,
			CallerDID:  "mcp:ai_agent",
			Timestamp:  time.Now().Unix(),
		}
		res := b.ExecuteControlCommand(cmd)
		data, err := json.MarshalIndent(res, "", "  ")
		if err != nil {
			return "", fmt.Errorf("error serializando respuesta de robot/vehiculo: %w", err)
		}
		return string(data), nil
	}
}

func getRobotVehicleTools() []MCPTool {
	return []MCPTool{
		{
			Name:        "ipvn7_robot_vehicle_control",
			Description: "Control soberano de drones, robots autónomos, automóviles y camiones vía MAVLink v2 y CAN Bus/J1939",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"device_id": map[string]interface{}{
						"type":        "string",
						"description": "ID del dispositivo o dron (ej. 'drone_01', 'car_01', 'truck_volvo')",
					},
					"action": map[string]interface{}{
						"type":        "string",
						"description": "Acción a ejecutar: 'telemetry', 'arm', 'disarm', 'takeoff', 'land', 'rtl', 'clear_dtc', 'status'",
						"enum":        []string{"telemetry", "arm", "disarm", "takeoff", "land", "rtl", "clear_dtc", "status"},
					},
					"params": map[string]interface{}{
						"type":        "object",
						"description": "Parámetros adicionales de comando",
					},
				},
				"required": []string{"action"},
			},
		},
	}
}
