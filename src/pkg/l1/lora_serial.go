// Package l1 implementa el puente físico serial UART para módulos LoRa (SX1276 / Ebyte E32),
// rescatado de Ipv7-8 (transport/lora_serial.go) y adaptado para ipvn7 v0.7.
package l1

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
)

var (
	// SyncWordLoRa identifica el inicio de una trama de radiofrecuencia LoRa
	SyncWordLoRa = []byte{0x3C, 0x5E} // '<', '^'
)

// LoRaFrameHandler define la función receptora de paquetes LoRa decodificados
type LoRaFrameHandler func(payload []byte)

// LoRaSerialBridge gestiona la lectura y escritura sobre el canal serial UART
type LoRaSerialBridge struct {
	mu        sync.RWMutex
	portName  string
	baudRate  int
	rwc       io.ReadWriteCloser
	onFrame   LoRaFrameHandler
	running   bool
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewLoRaSerialBridge inicializa la estructura del puente serial LoRa
func NewLoRaSerialBridge(portName string, baudRate int, rwc io.ReadWriteCloser, onFrame LoRaFrameHandler) *LoRaSerialBridge {
	ctx, cancel := context.WithCancel(context.Background())
	return &LoRaSerialBridge{
		portName: portName,
		baudRate: baudRate,
		rwc:      rwc,
		onFrame:  onFrame,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start activa el bucle continuo de escucha serial
func (b *LoRaSerialBridge) Start() error {
	b.mu.Lock()
	if b.running {
		b.mu.Unlock()
		return nil
	}
	if b.rwc == nil {
		b.mu.Unlock()
		return errors.New("lora_serial: puerto de comunicacion no asignado")
	}
	b.running = true
	b.mu.Unlock()

	go b.readLoop()
	return nil
}

// Stop finaliza el bucle y cierra el puerto físico
func (b *LoRaSerialBridge) Stop() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.running {
		return nil
	}
	b.running = false
	b.cancel()

	if b.rwc != nil {
		return b.rwc.Close()
	}
	return nil
}

// SendFrame empaqueta el payload con SyncWord, longitud y checksum XOR y lo transmite
func (b *LoRaSerialBridge) SendFrame(payload []byte) error {
	b.mu.RLock()
	rwc := b.rwc
	running := b.running
	b.mu.RUnlock()

	if !running || rwc == nil {
		return errors.New("lora_serial: puente no activo")
	}

	frame := EncodeLoRaFrame(payload)
	_, err := rwc.Write(frame)
	if err != nil {
		return fmt.Errorf("lora_serial: fallo al escribir trama: %w", err)
	}
	return nil
}

// EncodeLoRaFrame genera la trama: [SyncWord:2 | Len:2 | Payload:N | ChecksumXOR:1]
func EncodeLoRaFrame(payload []byte) []byte {
	n := len(payload)
	frame := make([]byte, 2+2+n+1)
	copy(frame[0:2], SyncWordLoRa)
	frame[2] = byte(n >> 8)
	frame[3] = byte(n & 0xFF)
	copy(frame[4:4+n], payload)

	// Calcular XOR checksum sobre la longitud y el payload
	var chk byte
	for _, b := range frame[2 : 4+n] {
		chk ^= b
	}
	frame[4+n] = chk
	return frame
}

// DecodeLoRaFrame decodifica y valida una trama LoRa
func DecodeLoRaFrame(data []byte) ([]byte, error) {
	if len(data) < 5 {
		return nil, errors.New("lora_serial: trama demasiado corta")
	}
	if data[0] != SyncWordLoRa[0] || data[1] != SyncWordLoRa[1] {
		return nil, errors.New("lora_serial: SyncWord invalido")
	}

	n := int(data[2])<<8 | int(data[3])
	if len(data) < 4+n+1 {
		return nil, errors.New("lora_serial: longitud insuficiente de payload")
	}

	var chk byte
	for _, b := range data[2 : 4+n] {
		chk ^= b
	}
	if chk != data[4+n] {
		return nil, errors.New("lora_serial: checksum XOR invalido (ruido RF)")
	}

	payload := make([]byte, n)
	copy(payload, data[4:4+n])
	return payload, nil
}

func (b *LoRaSerialBridge) readLoop() {
	buf := make([]byte, 1024)
	for {
		n, err := b.rwc.Read(buf)
		if err != nil {
			select {
			case <-b.ctx.Done():
				return
			default:
				return
			}
		}

		if n >= 5 {
			payload, err := DecodeLoRaFrame(buf[:n])
			if err == nil {
				b.mu.RLock()
				cb := b.onFrame
				b.mu.RUnlock()
				if cb != nil {
					cb(payload)
				}
			}
		}
	}
}
