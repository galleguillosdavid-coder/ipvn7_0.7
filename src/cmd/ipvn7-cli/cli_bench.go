package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"ipvn7/pkg/core"
)

func handleBenchmarkCommand(args []string) {
	if len(args) < 1 {
		fmt.Println("Uso: ipvn7-cli bench <target_ip:puerto> [duracion_seg] [tam_paquete_bytes]")
		fmt.Println("Ejemplo: ipvn7-cli bench 192.168.1.106:7001 5 1280")
		return
	}

	target := args[0]
	durationSec := 5
	packetSize := 1280

	if len(args) >= 2 {
		if d, err := strconv.Atoi(args[1]); err == nil && d > 0 {
			durationSec = d
		}
	}
	if len(args) >= 3 {
		if s, err := strconv.Atoi(args[2]); err == nil && s >= 32 {
			packetSize = s
		}
	}

	fmt.Println("==================================================================")
	fmt.Println("       ipvn7 EMPIRICAL NETWORK BENCHMARK (RFC 3550 WIRE)          ")
	fmt.Println("==================================================================")
	fmt.Printf(" Destino Objetivo : %s\n", target)
	fmt.Printf(" Tamaño Datagrama : %d bytes (MTU Canónico)\n", packetSize)
	fmt.Printf(" Duración Ensayo  : %d segundos\n", durationSec)
	fmt.Println("------------------------------------------------------------------")
	fmt.Println("[*] Transmitiendo ráfaga de saturación sobre enlace físico real...")

	res, err := core.RunBenchmarkClient(core.BenchmarkConfig{
		Target:     target,
		Duration:   time.Duration(durationSec) * time.Second,
		PacketSize: packetSize,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Falla en la prueba de benchmark: %v\n", err)
		return
	}

	fmt.Println("\n==================================================================")
	fmt.Println("              RESULTADOS CUANTITATIVOS DEL ENLACE                 ")
	fmt.Println("==================================================================")
	fmt.Printf(" Datos Transmitidos   : %.2f MB en %.2f segundos\n", res.MBTotal, res.DurationSec)
	fmt.Printf(" Throughput Sostenido : %.2f MB/s (%.2f Mbps)\n", res.ThroughputMBs, res.ThroughputMbps)
	fmt.Printf(" Tasa de Datagramas   : %.0f paquetes/segundo (PPS)\n", res.PPS)
	fmt.Printf(" Total Datagramas Tx  : %d paquetes\n", res.PacketsSent)
	fmt.Printf(" Jitter RFC 3550      : %.3f ms\n", res.JitterMs)
	fmt.Printf(" Pérdida Observada    : %.2f%% (%d paquetes)\n", res.LossPercentage, res.PacketsLost)
	fmt.Println("==================================================================")
}
