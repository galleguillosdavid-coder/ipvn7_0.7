package core

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"ipvn7/pkg/l1"
)

// Constantes de modo VPN
const (
	VPNModeMeshOnly     = "mesh_only"     // Rojo: Solo Malla P2P (Overlay Mesh)
	VPNModeConnecting   = "connecting"    // Amarillo: Estableciendo túnel/proxy
	VPNModeSOCKS5Active = "socks5_active" // Verde: Internet enrutado por SOCKS5 (127.0.0.1:10807)
	VPNModeTUNActive    = "tun_active"    // Verde: Internet enrutado por TUN Nativo
)

// VPNStatusResponse modela la respuesta del estado del enrutador VPN
type VPNStatusResponse struct {
	Mode             string `json:"mode"`              // mesh_only | connecting | socks5_active | tun_active
	ColorState       string `json:"color_state"`       // red | yellow | green
	Title            string `json:"title"`             // Título amigable del estado
	Description      string `json:"description"`       // Descripción técnica del modo
	SOCKS5Addr       string `json:"socks5_addr"`       // Dirección y puerto del proxy local
	SOCKS5Running    bool   `json:"socks5_running"`    // Si el proxy está escuchando
	BytesTx          uint64 `json:"bytes_tx"`          // Bytes transmitidos por el proxy
	BytesRx          uint64 `json:"bytes_rx"`          // Bytes recibidos por el proxy
	LastDestination  string `json:"last_destination"`  // Último destino alcanzado
	ConnectedSince   int64  `json:"connected_since"`   // Timestamp Unix de conexión
	TunNativeActive  bool   `json:"tun_native_active"` // Si existe adaptador TUN de kernel
}

// VPNStateHolder gestiona la concurrencia del estado VPN
type VPNStateHolder struct {
	mu             sync.RWMutex
	mode           string
	socks5Gateway  *l1.SOCKS5Gateway
	connectedSince time.Time
}

var (
	globalVPNState     *VPNStateHolder
	globalVPNStateOnce sync.Once
)

func getVPNState() *VPNStateHolder {
	globalVPNStateOnce.Do(func() {
		globalVPNState = &VPNStateHolder{
			mode:          VPNModeMeshOnly,
			socks5Gateway: l1.NewSOCKS5Gateway("127.0.0.1:10807", nil),
		}
	})
	return globalVPNState
}

// handleVPNStatus responde con el estado actual de la VPN (Rojo/Amarillo/Verde)
func (s *CoreServer) handleVPNStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	state := getVPNState()
	state.mu.RLock()
	defer state.mu.RUnlock()

	gw := state.socks5Gateway
	var bytesTx, bytesRx uint64
	var lastDest string
	var isRunning bool
	if gw != nil {
		bytesTx = gw.BytesTx
		bytesRx = gw.BytesRx
		lastDest = gw.LastDestination
		isRunning = gw.IsRunning
	}

	colorState := "red"
	title := "Malla P2P Activa (Overlay Mesh)"
	desc := "El nodo está conectado a la malla soberana. El tráfico de Internet NO está canalizado por ipvn7."

	switch state.mode {
	case VPNModeConnecting:
		colorState = "yellow"
		title = "Conectando Túnel Cuántico..."
		desc = "Estableciendo enlace y verificando gateway SOCKS5 de protección..."
	case VPNModeSOCKS5Active:
		colorState = "green"
		title = "Internet Protegido Activo (SOCKS5)"
		desc = "Tráfico canalizado de forma transparente por ipvn7 en 127.0.0.1:10807."
	case VPNModeTUNActive:
		colorState = "green"
		title = "Internet Protegido Activo (TUN Nativo)"
		desc = "Adaptador virtual de kernel activo con enrutamiento global."
	}

	var connSince int64
	if !state.connectedSince.IsZero() {
		connSince = state.connectedSince.Unix()
	}

	resp := VPNStatusResponse{
		Mode:            state.mode,
		ColorState:      colorState,
		Title:           title,
		Description:     desc,
		SOCKS5Addr:      "127.0.0.1:10807",
		SOCKS5Running:   isRunning,
		BytesTx:         bytesTx,
		BytesRx:         bytesRx,
		LastDestination: lastDest,
		ConnectedSince:  connSince,
		TunNativeActive: false,
	}

	_ = json.NewEncoder(w).Encode(resp)
}

// handleVPNToggle conmuta entre modo Malla P2P (Rojo) y modo Internet Protegido (Verde)
func (s *CoreServer) handleVPNToggle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	state := getVPNState()
	state.mu.Lock()
	defer state.mu.Unlock()

	if state.mode == VPNModeMeshOnly {
		// Iniciar pasarela SOCKS5 si no está activa
		if !state.socks5Gateway.IsRunning {
			if err := state.socks5Gateway.Start(); err != nil {
				// Si no puede arrancar el socket, intentar mantenerlo
			}
		}
		state.mode = VPNModeSOCKS5Active
		state.connectedSince = time.Now()
	} else {
		// Volver a modo malla pura
		state.mode = VPNModeMeshOnly
		state.connectedSince = time.Time{}
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"mode":    state.mode,
	})
}

// ActivateVPNMode activa inmediatamente la protección VPN sin requerir petición HTTP externa
func ActivateVPNMode() error {
	state := getVPNState()
	state.mu.Lock()
	defer state.mu.Unlock()

	if state.mode == VPNModeMeshOnly {
		if !state.socks5Gateway.IsRunning {
			if err := state.socks5Gateway.Start(); err != nil {
				return err
			}
		}
		state.mode = VPNModeSOCKS5Active
		state.connectedSince = time.Now()
	}
	return nil
}

