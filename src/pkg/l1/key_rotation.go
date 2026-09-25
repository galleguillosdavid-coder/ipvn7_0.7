// Package l1 implementa el centinela de rotación proactiva de claves de sesión (Forward Secrecy)
// para la protección criptográfica continua en la malla ipvn7 v0.7.
package l1

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// DefaultMaxBytesPerKey umbral de volumen para forzar rotación (1 GB)
	DefaultMaxBytesPerKey uint64 = 1024 * 1024 * 1024
	// DefaultMaxPacketsPerKey umbral de datagramas por clave (1M de paquetes)
	DefaultMaxPacketsPerKey uint64 = 1000000
	// DefaultKeyLifetime tiempo máximo de vigencia antes de re-keying (1 hora)
	DefaultKeyLifetime = 1 * time.Hour
	// KeyTransitionGracePeriod ventana de gracia para datagramas desordenados de época anterior
	KeyTransitionGracePeriod = 30 * time.Second
)

// SessionKeys almacena el par de claves simétricas activa y previa para Forward Secrecy
type SessionKeys struct {
	Epoch       uint64
	ActiveKey   [32]byte
	PreviousKey [32]byte
	CreatedAt   time.Time
	HasPrevious bool
}

// KeyRotationSentinel monitoriza atómicamente el tráfico y orquesta la renovación de claves
type KeyRotationSentinel struct {
	mu           sync.RWMutex
	sessionID    string
	keys         SessionKeys
	bytesCount   uint64
	packetsCount uint64
	maxBytes     uint64
	maxPackets   uint64
	lifetime     time.Duration
}

// NewKeyRotationSentinel inicializa el centinela con una clave efímera inicial de 32 bytes
func NewKeyRotationSentinel(sessionID string) (*KeyRotationSentinel, error) {
	var initialKey [32]byte
	if _, err := io.ReadFull(rand.Reader, initialKey[:]); err != nil {
		return nil, fmt.Errorf("key_rotation: fallo al generar entropia: %w", err)
	}

	return &KeyRotationSentinel{
		sessionID: sessionID,
		keys: SessionKeys{
			Epoch:       1,
			ActiveKey:   initialKey,
			CreatedAt:   time.Now(),
			HasPrevious: false,
		},
		maxBytes:   DefaultMaxBytesPerKey,
		maxPackets: DefaultMaxPacketsPerKey,
		lifetime:   DefaultKeyLifetime,
	}, nil
}

// NeedsRotation evalúa si se ha rebasado el presupuesto de bytes, paquetes o tiempo
func (s *KeyRotationSentinel) NeedsRotation() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bytes := atomic.LoadUint64(&s.bytesCount)
	packets := atomic.LoadUint64(&s.packetsCount)
	elapsed := time.Since(s.keys.CreatedAt)

	return bytes >= s.maxBytes || packets >= s.maxPackets || elapsed >= s.lifetime
}

// TrackTraffic contabiliza atómicamente datagramas y bytes cursados
func (s *KeyRotationSentinel) TrackTraffic(packetBytes int) bool {
	atomic.AddUint64(&s.bytesCount, uint64(packetBytes))
	atomic.AddUint64(&s.packetsCount, 1)
	return s.NeedsRotation()
}

// Rotate ejecuta la derivación criptográfica de la siguiente época y purga claves expiradas
func (s *KeyRotationSentinel) Rotate() (uint64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Derivar nueva clave mediante HKDF/Hash sobre la clave activa + entropía fresca
	var salt [32]byte
	if _, err := io.ReadFull(rand.Reader, salt[:]); err != nil {
		return 0, errors.New("key_rotation: fallo al leer salt fresco")
	}

	h := sha256.New()
	h.Write(s.keys.ActiveKey[:])
	h.Write(salt[:])
	h.Write([]byte(fmt.Sprintf("epoch:%d:%s", s.keys.Epoch+1, s.sessionID)))
	var nextKey [32]byte
	copy(nextKey[:], h.Sum(nil))

	// La clave activa pasa a ser la clave previa para la ventana de gracia
	s.keys.PreviousKey = s.keys.ActiveKey
	s.keys.ActiveKey = nextKey
	s.keys.Epoch++
	s.keys.CreatedAt = time.Now()
	s.keys.HasPrevious = true

	// Reiniciar contadores de la época
	atomic.StoreUint64(&s.bytesCount, 0)
	atomic.StoreUint64(&s.packetsCount, 0)

	// Zero-wipe del buffer salt
	for i := range salt {
		salt[i] = 0
	}

	return s.keys.Epoch, nil
}

// GetKeys retorna copia segura de las claves para descifrado en la época actual o previa
func (s *KeyRotationSentinel) GetKeys() SessionKeys {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := s.keys
	// Si expiró el periodo de gracia de 30s, purgar la clave previa en la copia
	if res.HasPrevious && time.Since(res.CreatedAt) > KeyTransitionGracePeriod {
		res.HasPrevious = false
		for i := range res.PreviousKey {
			res.PreviousKey[i] = 0
		}
	}
	return res
}

// SetLimits ajusta los umbrales para pruebas o configuraciones de alta exigencia
func (s *KeyRotationSentinel) SetLimits(maxBytes, maxPackets uint64, lifetime time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.maxBytes = maxBytes
	s.maxPackets = maxPackets
	s.lifetime = lifetime
}
