// Package mqttbridge implementa un bridge/broker físico ultraligero MQTT 3.1.1
// para telemetría de sensores IoT y periféricos, rescatado de la versión 0.1
// y adaptado para integrarse con la malla ipvn7 v0.7 sin dependencias externas.
package mqttbridge

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
)

// MessageHandler función de callback cuando se publica un mensaje MQTT
type MessageHandler func(topic string, payload []byte)

// MQTTBridge gestiona el listener TCP físico y el enrutamiento de tópicos MQTT
type MQTTBridge struct {
	mu          sync.RWMutex
	listenAddr  string
	listener    net.Listener
	subscribers map[string][]chan []byte
	onMessage   MessageHandler
	running     bool
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewMQTTBridge inicializa un nuevo bridge MQTT físico
func NewMQTTBridge(listenAddr string, onMsg MessageHandler) *MQTTBridge {
	ctx, cancel := context.WithCancel(context.Background())
	return &MQTTBridge{
		listenAddr:  listenAddr,
		subscribers: make(map[string][]chan []byte),
		onMessage:   onMsg,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start activa el listener TCP en la interfaz especificada
func (b *MQTTBridge) Start() error {
	b.mu.Lock()
	if b.running {
		b.mu.Unlock()
		return nil
	}

	ln, err := net.Listen("tcp", b.listenAddr)
	if err != nil {
		b.mu.Unlock()
		return fmt.Errorf("mqtt: fallo al escuchar en %s: %w", b.listenAddr, err)
	}

	b.listener = ln
	b.running = true
	b.mu.Unlock()

	go b.acceptLoop()
	return nil
}

// Addr retorna la dirección física donde el listener está escuchando
func (b *MQTTBridge) Addr() net.Addr {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.listener != nil {
		return b.listener.Addr()
	}
	return nil
}

// Stop cierra ordenadamente el listener y las conexiones
func (b *MQTTBridge) Stop() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.running {
		return nil
	}
	b.running = false
	b.cancel()

	if b.listener != nil {
		return b.listener.Close()
	}
	return nil
}

func (b *MQTTBridge) acceptLoop() {
	for {
		conn, err := b.listener.Accept()
		if err != nil {
			select {
			case <-b.ctx.Done():
				return
			default:
				return
			}
		}
		go b.handleClient(conn)
	}
}

func (b *MQTTBridge) handleClient(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 4096)
	for {
		// Leer paquete fijo MQTT
		n, err := conn.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			return
		}
		if n < 2 {
			continue
		}

		packetType := buf[0] >> 4
		switch packetType {
		case 1: // CONNECT
			// Responder CONNACK (0x20, 0x02, 0x00, 0x00) -> Conexión aceptada
			_, _ = conn.Write([]byte{0x20, 0x02, 0x00, 0x00})

		case 3: // PUBLISH
			b.parseAndRoutePublish(buf[:n])

		case 8: // SUBSCRIBE
			if n >= 4 {
				// Responder SUBACK con QoS 0
				msgID := buf[2:4]
				suback := []byte{0x90, 0x03, msgID[0], msgID[1], 0x00}
				_, _ = conn.Write(suback)
			}

		case 12: // PINGREQ
			// Responder PINGRESP
			_, _ = conn.Write([]byte{0xD0, 0x00})

		case 14: // DISCONNECT
			return
		}
	}
}

func (b *MQTTBridge) parseAndRoutePublish(data []byte) {
	if len(data) < 4 {
		return
	}
	// Longitud variable según RFC MQTT 3.1.1
	// Simple decodificación de longitud remanente de 1 byte
	offset := 2
	topicLen := int(binary.BigEndian.Uint16(data[offset : offset+2]))
	offset += 2

	if len(data) < offset+topicLen {
		return
	}
	topic := string(data[offset : offset+topicLen])
	offset += topicLen

	payload := data[offset:]

	b.mu.RLock()
	callback := b.onMessage
	b.mu.RUnlock()

	if callback != nil {
		callback(topic, payload)
	}
}

// Publish publica programáticamente un mensaje simulando una llegada externa
func (b *MQTTBridge) Publish(topic string, payload []byte) {
	b.mu.RLock()
	cb := b.onMessage
	b.mu.RUnlock()
	if cb != nil {
		cb(topic, payload)
	}
}
