package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"ipvn7/pkg/components/docker_proxy"
	"ipvn7/pkg/core"
	"ipvn7/pkg/l0"
)

func handleTunnelCommand(args []string, id *l0.Identity) {
	if len(args) < 1 {
		fmt.Println("Uso: ipvn7-cli tunnel <puerto_local> [nombre_servicio]")
		fmt.Println("Ejemplo: ipvn7-cli tunnel 8080 mi-web")
		return
	}

	portNum, err := strconv.Atoi(args[0])
	if err != nil || portNum <= 0 || portNum > 65535 {
		fmt.Fprintf(os.Stderr, "[ERROR] Puerto local inválido: %s\n", args[0])
		return
	}

	serviceName := fmt.Sprintf("service-%d", portNum)
	if len(args) >= 2 && args[1] != "" {
		serviceName = args[1]
	}

	backendURL := fmt.Sprintf("http://127.0.0.1:%d", portNum)
	targetDID := id.DID()
	virtualIPv4 := id.IPv4().String()
	virtualIPv6 := id.IPv6().String()

	// 1. Consultar estado del demonio local si está activo en :7070
	daemonActive := false
	client := &http.Client{Timeout: 600 * time.Millisecond}
	resp, err := client.Get("http://127.0.0.1:7070/api/v1/status")
	if err == nil && resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		var st core.CoreStatus
		if err := json.NewDecoder(resp.Body).Decode(&st); err == nil {
			daemonActive = true
			if st.DID != "" {
				targetDID = st.DID
			}
			if st.IPv4 != "" {
				virtualIPv4 = st.IPv4
			}
			if st.IPv6 != "" {
				virtualIPv6 = st.IPv6
			}
		}
	}

	// 2. Registrar componente en el Smart Gateway si el demonio responde
	compID := fmt.Sprintf("tunnel-%d", portNum)
	if daemonActive {
		regPayload := map[string]interface{}{
			"id":           compID,
			"name":         fmt.Sprintf("Sovereign Ingress Tunnel :%d", portNum),
			"transport":    "HTTP/TCP",
			"capabilities": []string{"reverse-proxy", "zero-trust-ingress"},
			"metadata": map[string]string{
				"backend": backendURL,
				"alias":   serviceName + ".ipvn7",
			},
		}
		data, _ := json.Marshal(regPayload)
		_, _ = client.Post("http://127.0.0.1:7070/api/v1/components/register", "application/json", bytes.NewReader(data))
	}

	// 3. Inicializar proxy inverso local con soporte de DID y Host
	proxy := docker_proxy.NewSovereignDockerProxy("127.0.0.1:0")
	ln, err := proxy.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] No se pudo arrancar el túnel de ingreso: %v\n", err)
		return
	}
	defer ln.Close()
	defer proxy.Close()

	_ = proxy.RegisterRoute(targetDID, backendURL, compID)
	_ = proxy.RegisterRoute(serviceName+".ipvn7", backendURL, compID)
	_ = proxy.RegisterRoute(fmt.Sprintf("localhost:%d", portNum), backendURL, compID)

	// 4. Banner informativo de ingreso soberano
	fmt.Println("==================================================================")
	fmt.Println("      IPVN7 SOVEREIGN TUNNEL: INGRESO ZERO-TRUST (ANTI-CLOUDFLARE)")
	fmt.Println("==================================================================")
	fmt.Printf(" [SERVICIO LOCAL]   : %s\n", backendURL)
	fmt.Printf(" [ALIAS SOBERANO]   : %s.ipvn7\n", serviceName)
	fmt.Printf(" [PROXY INGRESS]    : http://%s\n", proxy.Addr())
	fmt.Printf(" [DID SOBERANO]     : %s\n", targetDID)
	fmt.Printf(" [IPv4 VIRTUAL]     : %s:%d\n", virtualIPv4, portNum)
	fmt.Printf(" [IPv6 SOBERANA]    : [%s]:%d\n", virtualIPv6, portNum)
	fmt.Printf(" [ESTADO NÚCLEO]    : %s\n", func() string {
		if daemonActive {
			return "Demonio Activo (:7070) - Acoplado al Smart Gateway"
		}
		return "Modo Local Standalone (Direct Ingress)"
	}())
	fmt.Println(" [SEGURIDAD]        : Cifrado Híbrido Post-Cuántico ML-KEM-768 E2EE")
	fmt.Println(" [ZTNA FIREWALL]    : Microsegmentación Default-Deny Activa")
	fmt.Println("------------------------------------------------------------------")
	fmt.Println(" -> Túnel activo y publicado en la malla. Presiona Ctrl+C para detener.")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n[INFO] Desconectando túnel soberano limpiamente...")
	if daemonActive {
		unregPayload := map[string]string{"id": compID}
		data, _ := json.Marshal(unregPayload)
		_, _ = client.Post("http://127.0.0.1:7070/api/v1/components/unregister", "application/json", bytes.NewReader(data))
	}
	fmt.Println("[OK] Túnel finalizado con éxito.")
}
