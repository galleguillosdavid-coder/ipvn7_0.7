package vehiclerobotbridge

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"
)

const (
	MAVLinkV2MagicByte byte = 0xFD

	// Message IDs estándar MAVLink v2
	MsgIDHeartbeat         uint32 = 0
	MsgIDAttitude          uint32 = 30
	MsgIDGlobalPositionInt uint32 = 33
	MsgIDCommandLong       uint32 = 76
	MsgIDBatteryStatus     uint32 = 147

	// Comandos de navegación MAV_CMD
	CmdNavTakeoff       uint16 = 22
	CmdNavLand          uint16 = 21
	CmdNavReturnToLaunch uint16 = 20
	CmdComponentArmDisarm uint16 = 400
)

// MAVLinkV2Packet estructura canónica de trama MAVLink v2
type MAVLinkV2Packet struct {
	MagicByte uint8
	PayloadLen uint8
	IncompatFlags uint8
	CompatFlags   uint8
	Seq           uint8
	SysID         uint8
	CompID        uint8
	MsgID         uint32 // 24 bits
	Payload       []byte
	Checksum      uint16
}

// ParseMAVLinkV2Frame decodifica un paquete MAVLink v2 desde un slice de bytes
func ParseMAVLinkV2Frame(data []byte) (*MAVLinkV2Packet, error) {
	if len(data) < 12 {
		return nil, errors.New("mavlink: trama demasiado corta (<12 bytes)")
	}
	if data[0] != MAVLinkV2MagicByte {
		return nil, fmt.Errorf("mavlink: byte mágico inválido 0x%02X (esperado 0xFD)", data[0])
	}

	payloadLen := data[1]
	expectedTotal := int(payloadLen) + 12
	if len(data) < expectedTotal {
		return nil, fmt.Errorf("mavlink: longitud insuficiente (esperados %d, recibidos %d)", expectedTotal, len(data))
	}

	msgID := uint32(data[7]) | (uint32(data[8]) << 8) | (uint32(data[9]) << 16)
	payload := make([]byte, payloadLen)
	copy(payload, data[10:10+payloadLen])

	chk := binary.LittleEndian.Uint16(data[10+payloadLen : 12+payloadLen])

	return &MAVLinkV2Packet{
		MagicByte:     data[0],
		PayloadLen:    payloadLen,
		IncompatFlags: data[2],
		CompatFlags:   data[3],
		Seq:           data[4],
		SysID:         data[5],
		CompID:        data[6],
		MsgID:         msgID,
		Payload:       payload,
		Checksum:      chk,
	}, nil
}

// EncodeMAVLinkV2Frame empaqueta un mensaje en formato binario MAVLink v2
func EncodeMAVLinkV2Frame(sysID, compID, seq uint8, msgID uint32, payload []byte) []byte {
	pLen := uint8(len(payload))
	buf := make([]byte, 12+int(pLen))
	buf[0] = MAVLinkV2MagicByte
	buf[1] = pLen
	buf[2] = 0 // Incompat flags
	buf[3] = 0 // Compat flags
	buf[4] = seq
	buf[5] = sysID
	buf[6] = compID
	buf[7] = uint8(msgID & 0xFF)
	buf[8] = uint8((msgID >> 8) & 0xFF)
	buf[9] = uint8((msgID >> 16) & 0xFF)
	copy(buf[10:], payload)

	// Cálculo de CRC16 X.25 / MCRF4XX canónico
	crc := CalculateMavlinkCRC(buf[1 : 10+int(pLen)])
	binary.LittleEndian.PutUint16(buf[10+int(pLen):], crc)
	return buf
}

// CalculateMavlinkCRC calcula el CRC-16-CCITT de MAVLink
func CalculateMavlinkCRC(data []byte) uint16 {
	var crc uint16 = 0xFFFF
	for _, b := range data {
		tmp := uint16(b) ^ (crc & 0xFF)
		tmp ^= (tmp << 4) & 0xFF
		crc = (crc >> 8) ^ (tmp << 8) ^ (tmp << 3) ^ (tmp >> 4)
	}
	return crc
}

// UpdateTelemetryFromMAVLink extrae telemetría desde paquetes decodificados
func UpdateTelemetryFromMAVLink(pkt *MAVLinkV2Packet, telem *DroneTelemetry) {
	if telem == nil || pkt == nil {
		return
	}
	telem.LastHeartbeat = time.Now()

	switch pkt.MsgID {
	case MsgIDHeartbeat:
		if len(pkt.Payload) >= 9 {
			// BaseMode bit 7 es Armed
			baseMode := pkt.Payload[6]
			telem.Armed = (baseMode & 128) != 0
			// CustomMode determina modo de vuelo (APM/PX4)
			customMode := binary.LittleEndian.Uint32(pkt.Payload[0:4])
			switch customMode {
			case 0:
				telem.FlightMode = "MANUAL"
			case 3:
				telem.FlightMode = "AUTO"
			case 4:
				telem.FlightMode = "GUIDED"
			case 5:
				telem.FlightMode = "LOITER"
			case 6:
				telem.FlightMode = "RTL"
			default:
				telem.FlightMode = fmt.Sprintf("MODE_%d", customMode)
			}
		}
	case MsgIDGlobalPositionInt:
		if len(pkt.Payload) >= 28 {
			lat := int32(binary.LittleEndian.Uint32(pkt.Payload[4:8]))
			lon := int32(binary.LittleEndian.Uint32(pkt.Payload[8:12]))
			alt := int32(binary.LittleEndian.Uint32(pkt.Payload[12:16]))
			hdg := uint16(binary.LittleEndian.Uint16(pkt.Payload[26:28]))

			telem.Latitude = float64(lat) / 1e7
			telem.Longitude = float64(lon) / 1e7
			telem.AltitudeMeters = float64(alt) / 1000.0
			telem.HeadingDeg = float64(hdg) / 100.0
		}
	case MsgIDBatteryStatus:
		if len(pkt.Payload) >= 36 {
			rem := int8(pkt.Payload[35])
			if rem >= 0 && rem <= 100 {
				telem.BatteryPercent = int(rem)
			}
			voltMv := binary.LittleEndian.Uint16(pkt.Payload[14:16])
			telem.BatteryVoltage = float64(voltMv) / 1000.0
		}
	}
}
