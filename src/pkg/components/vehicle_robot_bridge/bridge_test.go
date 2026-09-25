package vehiclerobotbridge

import (
	"encoding/binary"
	"testing"
)

type mockBus struct {
	events []string
}

func (m *mockBus) PublishEvent(eventType string, data []byte) error {
	m.events = append(m.events, eventType)
	return nil
}

func TestMAVLinkV2_EncodeDecode(t *testing.T) {
	// Payload simulado de heartbeat (9 bytes)
	payload := make([]byte, 9)
	binary.LittleEndian.PutUint32(payload[0:4], 4) // GUIDED mode
	payload[6] = 128                              // Armed bit

	frame := EncodeMAVLinkV2Frame(1, 1, 42, MsgIDHeartbeat, payload)
	if len(frame) != 12+9 {
		t.Fatalf("longitud esperada %d, obtenida %d", 12+9, len(frame))
	}

	pkt, err := ParseMAVLinkV2Frame(frame)
	if err != nil {
		t.Fatalf("error parseando MAVLink v2: %v", err)
	}

	if pkt.MsgID != MsgIDHeartbeat {
		t.Fatalf("MsgID esperado %d, obtenido %d", MsgIDHeartbeat, pkt.MsgID)
	}

	var telem DroneTelemetry
	UpdateTelemetryFromMAVLink(pkt, &telem)
	if !telem.Armed {
		t.Errorf("se esperaba que el dron estuviese armado")
	}
	if telem.FlightMode != "GUIDED" {
		t.Errorf("modo de vuelo esperado GUIDED, obtenido %s", telem.FlightMode)
	}
}

func TestCANBus_OBD2_Decode(t *testing.T) {
	bus := &mockBus{}
	bridge := NewVehicleRobotBridge("veh_bridge", "Test Bridge", bus)

	// Simular respuesta OBD-II de RPM: Mode 0x41, PID 0x0C, A=0x1F, B=0x40 -> ((31*256)+64)/4 = 2000 RPM
	frame := &CANFrame{
		ID:       0x7E8,
		Extended: false,
		DLC:      7,
		Data:     [8]byte{0x04, 0x41, OBDPIDEngineRPM, 0x1F, 0x40, 0x00, 0x00, 0x00},
	}

	updated := bridge.IngestCANFrame("car_01", DeviceTypeVehicle, frame)
	if !updated {
		t.Fatalf("se esperaba que la trama CAN OBD-II fuese procesada")
	}

	telem, ok := bridge.GetVehicleTelemetry("car_01")
	if !ok || telem.EngineRPM != 2000 {
		t.Fatalf("RPM esperadas 2000, obtenidas: %v", telem)
	}
}

func TestCANBus_J1939_Truck_Decode(t *testing.T) {
	bus := &mockBus{}
	bridge := NewVehicleRobotBridge("truck_bridge", "Truck Bridge", bus)

	// PGN 65263 (Engine Oil Pressure): CAN ID con bits 8-25 = 65263
	canID := (PGNEngineOil << 8) | 0x80000000 | 0x03 // Prioridad 6, Source 0x03
	frame := &CANFrame{
		ID:       canID,
		Extended: true,
		DLC:      8,
		Data:     [8]byte{0xFF, 0xFF, 0xFF, 0x50, 0xFF, 0xFF, 0xFF, 0xFF}, // Byte 3 = 80 -> 80*4 = 320 kPa
	}

	updated := bridge.IngestCANFrame("truck_volvo", DeviceTypeTruck, frame)
	if !updated {
		t.Fatalf("se esperaba que la trama J1939 fuese procesada")
	}

	telem, ok := bridge.GetVehicleTelemetry("truck_volvo")
	if !ok || telem.OilPressureKPa != 320 {
		t.Fatalf("Presión de aceite esperada 320 kPa, obtenida: %v", telem)
	}
}

func TestBridge_ControlCommands(t *testing.T) {
	bridge := NewVehicleRobotBridge("drone_bridge", "Drone Bridge", nil)

	// Inyectar telemetría inicial de dron
	payload := make([]byte, 9)
	frame := EncodeMAVLinkV2Frame(1, 1, 1, MsgIDHeartbeat, payload)
	_ = bridge.IngestMAVLinkData("quad_01", frame)

	// 1. Arm
	res := bridge.ExecuteControlCommand(&ControlCommand{
		DeviceID: "quad_01",
		Action:   "arm",
	})
	if !res.Success {
		t.Fatalf("fallo comando arm: %s", res.Message)
	}

	// 2. Takeoff
	resTakeoff := bridge.ExecuteControlCommand(&ControlCommand{
		DeviceID: "quad_01",
		Action:   "takeoff",
	})
	if !resTakeoff.Success {
		t.Fatalf("fallo comando takeoff: %s", resTakeoff.Message)
	}

	telem, _ := bridge.GetDroneTelemetry("quad_01")
	if telem.AltitudeMeters != 5.0 {
		t.Errorf("altitud esperada 5.0m tras takeoff, obtenida: %f", telem.AltitudeMeters)
	}
}
