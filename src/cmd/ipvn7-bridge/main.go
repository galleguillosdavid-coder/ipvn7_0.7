package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	devicebridge "ipvn7/pkg/components/device_bridge"
)

func main() {
	gatewayURL := flag.String("gateway", "http://127.0.0.1:7070", "URL del Smart Component Gateway del Núcleo ipvn7")
	interval := flag.Duration("interval", 30*time.Second, "Intervalo de escaneo de periféricos LAN")
	flag.Parse()

	fmt.Println("================================================================================")
	fmt.Println("        SOVEREIGN DEVICE BRIDGE — ipvn7 Componente Satélite v0.6.0              ")
	fmt.Println("================================================================================")
	fmt.Printf("[+] Conectando con Smart Component Gateway en %s...\n", *gatewayURL)

	cfg := devicebridge.DefaultBridgeConfig()
	cfg.CoreGatewayURL = *gatewayURL
	cfg.ScanInterval = *interval

	bridge := devicebridge.NewSovereignDeviceBridge(cfg)
	if err := bridge.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "[FATAL] Error iniciando el puente: %v\n", err)
		os.Exit(1)
	}
	defer bridge.Stop()

	fmt.Println("[+] Escuchando anuncios SSDP/UPnP (Smart TVs, Cast) y mDNS/IPP (Impresoras)...")
	fmt.Println("[+] Los dispositivos detectados se registrarán con DIDs sombra en el dashboard.")
	fmt.Println("[+] En ejecución continua. Presione Ctrl+C para detener.")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n[*] Apagado ordenado del Sovereign Device Bridge...")
}
