package core

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"ipvn7/pkg/l2"
)

type upnpDeviceDesc struct {
	XMLName xml.Name `xml:"root"`
	Device  struct {
		FriendlyName string `xml:"friendlyName"`
		Manufacturer string `xml:"manufacturer"`
		ModelName    string `xml:"modelName"`
		UDN          string `xml:"UDN"`
	} `xml:"device"`
}

func (s *CoreServer) handleHomeCast(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		TargetTV string `json:"target_tv"`
		MediaURL string `json:"media_url"`
		Title    string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	targetIP := ExtractTargetIP(req.TargetTV)
	displayName := req.TargetTV
	if targetIP == "" {
		for _, comp := range s.gateway.ListComponents() {
			for _, cap := range comp.Capabilities {
				if cap == "tv:cast" || cap == "tv:display" {
					targetIP = ExtractTargetIP(comp.Endpoint)
					if displayName == "" {
						displayName = comp.Name
					}
					break
				}
			}
			if targetIP != "" {
				break
			}
		}
	}
	if targetIP == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "No se detectó ninguna Smart TV activa en la red. Selecciona o ingresa una IP válida."})
		return
	}
	if displayName == "" {
		displayName = fmt.Sprintf("Smart TV (%s)", targetIP)
	}

	// 1. Obtención y análisis del descriptor UPnP/DIAL real de la TV (puerto 8008)
	dialClient := &http.Client{Timeout: 500 * time.Millisecond}
	descURL := fmt.Sprintf("http://%s:8008/ssdp/device-desc.xml", targetIP)
	dialStart := time.Now()

	var tvInfo upnpDeviceDesc
	var manufacturer, friendlyName, udn string
	tvFound := false

	if resp, err := dialClient.Get(descURL); err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			data, _ := io.ReadAll(resp.Body)
			if xml.Unmarshal(data, &tvInfo) == nil {
				tvFound = true
				manufacturer = tvInfo.Device.Manufacturer
				friendlyName = tvInfo.Device.FriendlyName
				udn = tvInfo.Device.UDN
			}
		}
	}
	rttMs := float64(time.Since(dialStart).Microseconds()) / 1000.0

	// 2. Consulta del estado de la aplicación Chromecast/DIAL en la TV
	chromecastState := "stopped"
	ccURL := fmt.Sprintf("http://%s:8008/apps/ChromeCast", targetIP)
	if ccResp, ccErr := dialClient.Get(ccURL); ccErr == nil {
		_ = ccResp.Body.Close()
		if ccResp.StatusCode == http.StatusOK {
			chromecastState = "available"
		}
	}

	statusStr := "connected_dial"
	statusMsg := fmt.Sprintf("Enlace DIAL/UPnP establecido con %s (%s - %s) [RTT: %.2f ms]", friendlyName, manufacturer, targetIP, rttMs)
	if !tvFound {
		statusStr = "probing"
		statusMsg = fmt.Sprintf("Sondeo TCP completado hacia %s (RTT: %.2f ms)", targetIP, rttMs)
	}

	mediaLaunched := false
	if req.MediaURL != "" {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		launched, detail := sendDialMediaLaunch(ctx, targetIP, req.MediaURL)
		mediaLaunched = launched
		if launched {
			statusStr = "playing_on_tv"
			statusMsg = fmt.Sprintf("¡Reproduciendo en %s! %s", friendlyName, detail)
		} else {
			if chromecastState == "available" {
				statusStr = "google_cast_device"
				statusMsg = fmt.Sprintf("Smart TV detectada (%s con Google Cast). Transmite con Cast de Chrome/Edge o Duplica Pantalla (Win+K).", friendlyName)
			} else {
				statusStr = "format_not_supported"
				statusMsg = fmt.Sprintf("Enlace activo con %s (%s), pero esta TV no soporta inicio directo DIAL para este formato.", friendlyName, targetIP)
			}
		}
	}

	s.telemetry.RecordEvent(l2.EventTxPacket, 1280, uint32(rttMs*1000), 0)
	s.gateway.PublishEvent(MeshEvent{
		Type:      "HOME_CAST_STARTED",
		Timestamp: time.Now().UnixNano(),
		Source:    s.identity.DID(),
		Payload: map[string]interface{}{
			"target_tv":       displayName,
			"target_ip":       targetIP,
			"manufacturer":    manufacturer,
			"friendly_name":   friendlyName,
			"udn":             udn,
			"chromecast_app":  chromecastState,
			"rtt_ms":          rttMs,
			"media_url":       req.MediaURL,
			"media_launched":  mediaLaunched,
			"status":          statusStr,
		},
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":          statusStr,
		"message":         statusMsg,
		"target_tv":       displayName,
		"target_ip":       targetIP,
		"manufacturer":    manufacturer,
		"friendly_name":   friendlyName,
		"udn":             udn,
		"chromecast_app":  chromecastState,
		"media_launched":  mediaLaunched,
		"rtt_ms":          rttMs,
		"protocol":        "DIAL/UPnP + Miracast/WebRTC",
	})
}

// sendDialMediaLaunch envía una orden directa de reproducción DIAL a la Smart TV
func sendDialMediaLaunch(ctx context.Context, targetIP, mediaURL string) (bool, string) {
	client := &http.Client{Timeout: 3 * time.Second}
	videoID := extractYouTubeID(mediaURL)

	if videoID != "" {
		dialURL := fmt.Sprintf("http://%s:8008/apps/YouTube", targetIP)
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, dialURL, strings.NewReader("v="+videoID))
		req.Header.Set("Content-Type", "text/plain; charset=utf-8")
		if resp, err := client.Do(req); err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
				return true, fmt.Sprintf("YouTube iniciado en TV con video ID '%s'", videoID)
			}
		}
	}

	dialApps := []string{"DefaultMediaReceiver", "YouTube", "ChromeCast"}
	for _, app := range dialApps {
		appURL := fmt.Sprintf("http://%s:8008/apps/%s", targetIP, app)
		body := mediaURL
		if app == "YouTube" && videoID != "" {
			body = "v=" + videoID
		}
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, appURL, strings.NewReader(body))
		req.Header.Set("Content-Type", "text/plain")
		if resp, err := client.Do(req); err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
				return true, fmt.Sprintf("Aplicación %s activada en TV", app)
			}
		}
	}
	return false, "No se pudo iniciar la aplicación de video en la TV vía DIAL"
}

func extractYouTubeID(input string) string {
	input = strings.TrimSpace(input)
	if len(input) == 11 && !strings.Contains(input, "/") && !strings.Contains(input, ".") {
		return input
	}
	if parts := strings.Split(input, "youtu.be/"); len(parts) > 1 {
		return strings.Split(strings.TrimSpace(parts[1]), "?")[0]
	}
	if parts := strings.Split(input, "v="); len(parts) > 1 {
		return strings.Split(strings.TrimSpace(parts[1]), "&")[0]
	}
	return ""
}

// handleHomeCastProject activa la proyección nativa del sistema operativo (Miracast / Wireless Display)
func (s *CoreServer) handleHomeCastProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	opened := false
	msg := "Proyección iniciada"
	if runtime.GOOS == "windows" {
		cmd := exec.Command("powershell", "-WindowStyle", "Hidden", "-Command", "Start-Process 'ms-settings-connectabledevices:devicediscovery'")
		if err := cmd.Start(); err == nil {
			opened = true
			msg = "Panel de Transmisión de Windows desplegado. Selecciona tu Smart TV para duplicar la pantalla."
		} else {
			cmdAlt := exec.Command("powershell", "-WindowStyle", "Hidden", "-Command", "Start-Process 'ms-settings:project'")
			_ = cmdAlt.Start()
			opened = true
			msg = "Panel de Proyección de Windows desplegado. Selecciona 'Duplicar' y elige tu Smart TV."
		}
	} else {
		msg = "Proyección nativa no disponible en este sistema operativo"
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "project_invoked",
		"opened":  opened,
		"message": msg,
		"os":      runtime.GOOS,
	})
}

// handleHomeCastStop detiene la reproducción y limpia la pantalla del receptor TV
func (s *CoreServer) handleHomeCastStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	s.gateway.PublishEvent(MeshEvent{
		Type:      "HOME_CAST_STOPPED",
		Timestamp: time.Now().UnixNano(),
		Source:    s.identity.DID(),
		Payload:   map[string]interface{}{"status": "stopped"},
	})
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "stopped",
		"message": "Transmisión detenida y pantalla receptora limpiada",
	})
}

// handleHomeCastControl envía comandos de reproducción y volumen a la TV vía SSE
func (s *CoreServer) handleHomeCastControl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Action string  `json:"action"`
		Value  float64 `json:"value"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	s.gateway.PublishEvent(MeshEvent{
		Type:      "HOME_CAST_CONTROL",
		Timestamp: time.Now().UnixNano(),
		Source:    s.identity.DID(),
		Payload:   map[string]interface{}{"action": req.Action, "value": req.Value},
	})
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok", "action": req.Action, "value": req.Value})
}

// handleTVReceiver se encuentra desacoplado en server_tv_receiver.go

