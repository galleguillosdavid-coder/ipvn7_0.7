// Package vehiclerobotbridge implementa el puente de telemetría y control soberano
// para robots, drones, vehículos autónomos (automóviles/camiones vía CAN/OBD-II)
// y maquinaria industrial en ipvn7 v0.7.
package vehiclerobotbridge

import "time"

// Constantes de tipos de dispositivos soportados
const (
	DeviceTypeDrone     = "drone"     // Drones y aeronaves autónomas (MAVLink v2)
	DeviceTypeRover     = "rover"     // Robots terrestres, juguetes y rovers (MAVLink/Serial)
	DeviceTypeVehicle   = "vehicle"   // Automóviles y camionetas (CAN Bus / OBD-II)
	DeviceTypeTruck     = "truck"     // Camiones y transporte de carga (CAN Bus / J1939)
	DeviceTypeMachinery = "machinery" // Maquinaria pesada e industrial (CAN/Modbus/J1939)
)

// DroneTelemetry telemetría canónica de aeronave o robot autónomo
type DroneTelemetry struct {
	Armed          bool      `json:"armed"`
	FlightMode     string    `json:"flight_mode"`     // GUIDED, AUTO, RTL, LOITER, MANUAL
	BatteryPercent int       `json:"battery_percent"` // 0-100
	BatteryVoltage float64   `json:"battery_voltage"` // Voltios
	AltitudeMeters float64   `json:"altitude_meters"` // Altitud relativa
	GroundSpeedMPS float64   `json:"ground_speed_mps"`
	Latitude       float64   `json:"latitude"`
	Longitude      float64   `json:"longitude"`
	Satellites     int       `json:"satellites"`
	HeadingDeg     float64   `json:"heading_deg"` // 0-360
	LastHeartbeat  time.Time `json:"last_heartbeat"`
}

// VehicleTelemetry telemetría canónica de automóvil, camión o máquina
type VehicleTelemetry struct {
	DeviceType     string    `json:"device_type"`
	EngineRPM      int       `json:"engine_rpm"`
	SpeedKmh       int       `json:"speed_kmh"`
	CoolantTempC   int       `json:"coolant_temp_c"`
	FuelLevelPct   int       `json:"fuel_level_pct"`
	EngineLoadPct  int       `json:"engine_load_pct"`
	OdometerKm     int64     `json:"odometer_km"`
	OilPressureKPa int       `json:"oil_pressure_kpa"` // J1939 para camiones/maquinaria
	DTCFaultCodes  []string  `json:"dtc_fault_codes"`  // Códigos de error OBD-II (ej. P0300)
	LastUpdate     time.Time `json:"last_update"`
}

// ControlCommand comando de control para robot, dron o actuador de vehículo
type ControlCommand struct {
	DeviceID   string                 `json:"device_id"`
	Action     string                 `json:"action"` // arm, disarm, takeoff, land, rtl, speed_limit, horn, lock_doors
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	CallerDID  string                 `json:"caller_did"`
	Timestamp  int64                  `json:"timestamp"`
}

// CommandResult resultado estructurado de la ejecución sobre el periférico
type CommandResult struct {
	Success   bool      `json:"success"`
	DeviceID  string    `json:"device_id"`
	Action    string    `json:"action"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}
