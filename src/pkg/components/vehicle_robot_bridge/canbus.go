package vehiclerobotbridge

import (
	"errors"
	"time"
)

// CANFrame representa una trama de red automotriz o de maquinaria CAN Bus
type CANFrame struct {
	ID       uint32  // 11-bit (Standard) o 29-bit (Extended)
	Extended bool    // True para tramas J1939 / CAN 2.0B
	RTR      bool    // Remote Transmission Request
	DLC      uint8   // Data Length Code (0-8 bytes)
	Data     [8]byte // Carga útil de datos
}

// OBD-II Modos y PIDs estándar (ISO 15765-4 / SAE J1979)
const (
	OBDModeShowCurrentData byte = 0x01
	OBDModeShowStoredDTCs  byte = 0x03
	OBDModeClearDTCs       byte = 0x04

	OBDPIDCoolantTemp byte = 0x05 // Temp en °C = A - 40
	OBDPIDEngineRPM   byte = 0x0C // RPM = ((A*256)+B)/4
	OBDPIDSpeedKmh    byte = 0x0D // Velocidad = A km/h
	OBDPIDEngineLoad  byte = 0x04 // Carga = (A*100)/255
	OBDPIDFuelLevel   byte = 0x2F // Nivel combustible = (A*100)/255
)

// J1939 PGNs para camiones y maquinaria pesada
const (
	PGNEEC1       uint32 = 61444 // Electronic Engine Controller 1 (RPM, Torque)
	PGNCCVS       uint32 = 65265 // Cruise Control / Vehicle Speed
	PGNEngineOil  uint32 = 65263 // Presión y nivel de aceite
	PGNCoolant    uint32 = 65262 // Temperatura de motor
	PGNOdometer   uint32 = 65248 // Distancia total recorrida
)

// ParseCANRawBytes convierte un flujo de 16 bytes a una trama CAN estructurada
func ParseCANRawBytes(raw []byte) (*CANFrame, error) {
	if len(raw) < 13 {
		return nil, errors.New("canbus: buffer insuficiente para trama CAN (<13B)")
	}
	id := uint32(raw[0]) | (uint32(raw[1]) << 8) | (uint32(raw[2]) << 16) | (uint32(raw[3]) << 24)
	isExtended := (raw[4] & 0x01) != 0
	dlc := raw[5]
	if dlc > 8 {
		dlc = 8
	}
	var data [8]byte
	copy(data[:], raw[6:6+int(dlc)])

	return &CANFrame{
		ID:       id,
		Extended: isExtended,
		DLC:      dlc,
		Data:     data,
	}, nil
}

// ProcessOBD2Response decodifica una respuesta ECU (ID 0x7E8) y actualiza telemetría
func ProcessOBD2Response(frame *CANFrame, telem *VehicleTelemetry) bool {
	if frame == nil || telem == nil || frame.DLC < 3 {
		return false
	}
	// Respuesta típica OBD-II: [BytesCount, Mode+0x40, PID, A, B, C, D]
	mode := frame.Data[1]
	if mode != 0x41 { // 0x01 + 0x40
		return false
	}
	pid := frame.Data[2]
	telem.LastUpdate = time.Now()

	switch pid {
	case OBDPIDEngineRPM:
		a := int(frame.Data[3])
		b := int(frame.Data[4])
		telem.EngineRPM = ((a * 256) + b) / 4
		return true
	case OBDPIDSpeedKmh:
		telem.SpeedKmh = int(frame.Data[3])
		return true
	case OBDPIDCoolantTemp:
		telem.CoolantTempC = int(frame.Data[3]) - 40
		return true
	case OBDPIDFuelLevel:
		telem.FuelLevelPct = (int(frame.Data[3]) * 100) / 255
		return true
	case OBDPIDEngineLoad:
		telem.EngineLoadPct = (int(frame.Data[3]) * 100) / 255
		return true
	}
	return false
}

// ProcessJ1939Frame decodifica tramas de camiones y maquinaria pesada
func ProcessJ1939Frame(frame *CANFrame, telem *VehicleTelemetry) bool {
	if frame == nil || telem == nil || !frame.Extended {
		return false
	}
	// PGN se deriva de los bits 8-25 del CAN ID de 29 bits
	pgn := (frame.ID >> 8) & 0x3FFFF
	telem.LastUpdate = time.Now()

	switch pgn {
	case PGNEEC1:
		// Bytes 3-4 contienen RPM: factor 0.125 rpm/bit
		rawRPM := uint16(frame.Data[3]) | (uint16(frame.Data[4]) << 8)
		telem.EngineRPM = int(float64(rawRPM) * 0.125)
		return true
	case PGNCCVS:
		// Bytes 1-2 contienen Velocidad en km/h: factor 1/256 km/h por bit
		rawSpeed := uint16(frame.Data[1]) | (uint16(frame.Data[2]) << 8)
		telem.SpeedKmh = int(float64(rawSpeed) / 256.0)
		return true
	case PGNEngineOil:
		// Byte 3 contiene Presión de aceite: factor 4 kPa/bit
		telem.OilPressureKPa = int(frame.Data[3]) * 4
		return true
	case PGNCoolant:
		// Byte 0 contiene Temperatura: -40 deg C offset
		telem.CoolantTempC = int(frame.Data[0]) - 40
		return true
	case PGNOdometer:
		// Bytes 0-3: Distancia total: factor 0.125 km/bit
		rawDist := uint32(frame.Data[0]) | (uint32(frame.Data[1]) << 8) | (uint32(frame.Data[2]) << 16) | (uint32(frame.Data[3]) << 24)
		telem.OdometerKm = int64(float64(rawDist) * 0.125)
		return true
	}
	return false
}
