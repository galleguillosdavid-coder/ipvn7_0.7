// Package coapproxy implementa un servidor/proxy CoAP físico ultraligero
// conforme a RFC 7252 sobre UDP, rescatado de la versión 0.1 para microcontroladores
// IoT (ESP32, STM32, Zephyr) y adaptado a ipvn7 v0.7.
package coapproxy

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"sync"
)

// CoAPType define el tipo de mensaje según RFC 7252
type CoAPType uint8

const (
	TypeConfirmable    CoAPType = 0
	TypeNonConfirmable CoAPType = 1
	TypeAcknowledgement CoAPType = 2
	TypeReset          CoAPType = 3
)

// CoAPCode códigos de método y respuesta estándar
type CoAPCode uint8

const (
	CodeEmpty   CoAPCode = 0x00
	CodeGET     CoAPCode = 0x01
	CodePOST    CoAPCode = 0x02
	CodePUT     CoAPCode = 0x03
	CodeDELETE  CoAPCode = 0x04
	CodeContent CoAPCode = 0x45 // 2.05 Content
	CodeCreated CoAPCode = 0x41 // 2.01 Created
	CodeNotFound CoAPCode = 0x84 // 4.04 Not Found
)

// CoAPMessage representa un datagrama decodificado RFC 7252
type CoAPMessage struct {
	Type      CoAPType
	Code      CoAPCode
	MessageID uint16
	Token     []byte
	Payload   []byte
}

// CoAPHandler procesa un mensaje entrante y retorna la respuesta CoAP
type CoAPHandler func(msg CoAPMessage, remoteAddr net.Addr) (*CoAPMessage, error)

// CoAPServer orquesta el listener físico UDP
type CoAPServer struct {
	mu         sync.RWMutex
	listenAddr string
	conn       *net.UDPConn
	handler    CoAPHandler
	running    bool
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewCoAPServer inicializa el servidor CoAP RFC 7252
func NewCoAPServer(listenAddr string, handler CoAPHandler) *CoAPServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &CoAPServer{
		listenAddr: listenAddr,
		handler:    handler,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start levanta el socket UDP físico y comienza el bucle de recepción
func (s *CoAPServer) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}

	uAddr, err := net.ResolveUDPAddr("udp", s.listenAddr)
	if err != nil {
		s.mu.Unlock()
		return fmt.Errorf("coap: fallo al resolver direccion: %w", err)
	}

	conn, err := net.ListenUDP("udp", uAddr)
	if err != nil {
		s.mu.Unlock()
		return fmt.Errorf("coap: fallo al escuchar en %s: %w", s.listenAddr, err)
	}

	s.conn = conn
	s.running = true
	s.mu.Unlock()

	go s.listenLoop()
	return nil
}

// LocalAddr retorna la dirección local asignada
func (s *CoAPServer) LocalAddr() net.Addr {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.conn != nil {
		return s.conn.LocalAddr()
	}
	return nil
}

// Stop cierra ordenadamente el socket UDP
func (s *CoAPServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}
	s.running = false
	s.cancel()

	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}

func (s *CoAPServer) listenLoop() {
	buf := make([]byte, 2048)
	for {
		n, rAddr, err := s.conn.ReadFrom(buf)
		if err != nil {
			select {
			case <-s.ctx.Done():
				return
			default:
				return
			}
		}

		if n < 4 {
			continue
		}

		msg, err := ParseCoAP(buf[:n])
		if err != nil {
			continue
		}

		s.mu.RLock()
		h := s.handler
		s.mu.RUnlock()

		if h != nil {
			resp, err := h(msg, rAddr)
			if err == nil && resp != nil {
				respBytes := EncodeCoAP(*resp)
				_, _ = s.conn.WriteTo(respBytes, rAddr)
			}
		}
	}
}

// ParseCoAP decodifica una trama binaria RFC 7252
func ParseCoAP(data []byte) (CoAPMessage, error) {
	if len(data) < 4 {
		return CoAPMessage{}, errors.New("coap: trama menor a 4 bytes de cabecera fija")
	}

	// Byte 0: [Ver:2 | Type:2 | TKL:4]
	tkl := data[0] & 0x0F
	mType := CoAPType((data[0] >> 4) & 0x03)
	code := CoAPCode(data[1])
	msgID := binary.BigEndian.Uint16(data[2:4])

	idx := 4
	if len(data) < idx+int(tkl) {
		return CoAPMessage{}, errors.New("coap: longitud de token excede la trama")
	}

	token := make([]byte, tkl)
	copy(token, data[idx:idx+int(tkl)])
	idx += int(tkl)

	// Buscar marcador de carga útil 0xFF
	var payload []byte
	for i := idx; i < len(data); i++ {
		if data[i] == 0xFF {
			payload = data[i+1:]
			break
		}
	}

	return CoAPMessage{
		Type:      mType,
		Code:      code,
		MessageID: msgID,
		Token:     token,
		Payload:   payload,
	}, nil
}

// EncodeCoAP serializa un CoAPMessage a formato binario RFC 7252
func EncodeCoAP(msg CoAPMessage) []byte {
	tkl := byte(len(msg.Token) & 0x0F)
	out := make([]byte, 4+int(tkl))

	// Ver=1 (01), Type, TKL
	out[0] = (0x01 << 6) | (byte(msg.Type&0x03) << 4) | tkl
	out[1] = byte(msg.Code)
	binary.BigEndian.PutUint16(out[2:4], msg.MessageID)

	if tkl > 0 {
		copy(out[4:4+int(tkl)], msg.Token)
	}

	if len(msg.Payload) > 0 {
		out = append(out, 0xFF)
		out = append(out, msg.Payload...)
	}

	return out
}
