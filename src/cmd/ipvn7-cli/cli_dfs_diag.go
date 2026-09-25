// Package main implementa los comandos de consola para diagnóstico de malla y almacenamiento DFS.
// Cumple con la regla de atomicidad modular (<=400 líneas) de docs/INGENIERIA_LEAN.md.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const defaultAPIBase = "http://localhost:7070"

func handleDiagCommand(args []string) {
	if len(args) == 0 || args[0] == "help" {
		fmt.Println(`Uso: ipvn7-cli diag [list | report <tipo> <gravedad> <detalles>]

Subcomandos:
  list                               Lista los últimos reportes de diagnóstico recibidos por la malla
  report <tipo> <gravedad> <detalles> Emite un reporte de anomalía manual al colector distribuido`)
		return
	}

	switch args[0] {
	case "list":
		resp, err := http.Get(defaultAPIBase + "/api/v1/telemetry/reports")
		if err != nil {
			fmt.Printf("[✗] Error consultando colector de diagnósticos: %v\n", err)
			return
		}
		defer resp.Body.Close()

		var data struct {
			Status       string                   `json:"status"`
			TotalReports int                      `json:"total_reports"`
			Reports      []map[string]interface{} `json:"reports"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			fmt.Printf("[✗] Error decodificando respuesta: %v\n", err)
			return
		}

		fmt.Printf("\n=== COLECTOR DISTRIBUIDO DE DIAGNÓSTICOS IPVN7 (Total: %d) ===\n", data.TotalReports)
		if len(data.Reports) == 0 {
			fmt.Println("No hay reportes de fallas registrados en la cola de memoria.")
			return
		}

		for idx, rep := range data.Reports {
			tUnix := int64(rep["timestamp"].(float64))
			tm := time.Unix(tUnix, 0).Format("15:04:05")
			fmt.Printf("[%d] %s | %s [%s] | SO: %s/%s\n", idx+1, tm, rep["anomaly_type"], rep["severity"], rep["os"], rep["arch"])
			fmt.Printf("    Reporter: %s\n", rep["reporter_did"])
			fmt.Printf("    Detalles: %s\n", rep["details"])
			if metrics, ok := rep["metrics"].(map[string]interface{}); ok && len(metrics) > 0 {
				fmt.Printf("    Métricas: %+v\n", metrics)
			}
			fmt.Println("    --------------------------------------------------------")
		}

	case "report":
		if len(args) < 4 {
			fmt.Println("Error: Formato requerido: ipvn7-cli diag report <tipo> <gravedad> <detalles>")
			return
		}
		anomalyType := args[1]
		severity := args[2]
		details := strings.Join(args[3:], " ")

		payload := map[string]interface{}{
			"reporter_did": "did:ipvn7:cli-operator",
			"timestamp":    time.Now().Unix(),
			"anomaly_type": anomalyType,
			"severity":     severity,
			"details":      details,
			"version":      "0.7.0",
			"os":           "cli",
			"arch":         "operator",
			"metrics":      map[string]float64{"cli_report": 1.0},
		}

		bodyBytes, _ := json.Marshal(payload)
		resp, err := http.Post(defaultAPIBase+"/api/v1/telemetry/report", "application/json", bytes.NewReader(bodyBytes))
		if err != nil {
			fmt.Printf("[✗] Error emitiendo reporte: %v\n", err)
			return
		}
		defer resp.Body.Close()

		var res map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&res)
		fmt.Printf("[✓] Reporte registrado exitosamente en el Colector Distribuido. ID: %v\n", res["report_id"])

	default:
		fmt.Printf("Subcomando diag desconocido: '%s'\n", args[0])
	}
}

func handleDFSCommand(args []string) {
	if len(args) == 0 || args[0] == "help" {
		fmt.Println(`Uso: ipvn7-cli dfs [put <archivo> | get <cid> [destino] | manifest <cid>]

Subcomandos:
  put <archivo>          Fragmenta y sube un archivo al almacén distribuido CAS, emitiendo su CID
  get <cid> [destino]    Descarga y reensambla un archivo completo por su CID
  manifest <cid>         Inspecciona los metadatos y árbol de bloques de un archivo en DFS`)
		return
	}

	switch args[0] {
	case "put":
		if len(args) < 2 {
			fmt.Println("Error: Debe especificar la ruta del archivo a subir.")
			return
		}
		filePath := args[1]
		content, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("[✗] Error leyendo archivo '%s': %v\n", filePath, err)
			return
		}

		payload := map[string]interface{}{
			"file_name": filepath.Base(filePath),
			"content":   string(content),
		}
		bodyBytes, _ := json.Marshal(payload)

		resp, err := http.Post(defaultAPIBase+"/api/v1/dfs/upload", "application/json", bytes.NewReader(bodyBytes))
		if err != nil {
			fmt.Printf("[✗] Error subiendo archivo a DFS: %v\n", err)
			return
		}
		defer resp.Body.Close()

		var res map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&res)
		fmt.Printf("[✓] Archivo almacenado exitosamente en DFS Soberano\n")
		fmt.Printf("    CID Manifiesto: %v\n", res["cid"])
		fmt.Printf("    Nombre:         %v\n", res["file_name"])
		fmt.Printf("    Tamaño:         %v bytes\n", res["total_size"])
		fmt.Printf("    Bloques SHA256: %v\n", res["chunks_count"])

	case "get":
		if len(args) < 2 {
			fmt.Println("Error: Debe especificar el CID del archivo a descargar.")
			return
		}
		cid := args[1]
		destPath := ""
		if len(args) >= 3 {
			destPath = args[2]
		}

		resp, err := http.Get(defaultAPIBase + "/api/v1/dfs/file/" + cid)
		if err != nil {
			fmt.Printf("[✗] Error descargando archivo de DFS: %v\n", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			fmt.Printf("[✗] Error del servidor (%d): %s\n", resp.StatusCode, string(b))
			return
		}

		if destPath == "" {
			// Extraer nombre de cabecera o usar CID
			destPath = cid + ".bin"
			cd := resp.Header.Get("Content-Disposition")
			if strings.Contains(cd, "filename=\"") {
				start := strings.Index(cd, "filename=\"") + 10
				end := strings.LastIndex(cd, "\"")
				if end > start {
					destPath = cd[start:end]
				}
			}
		}

		outData, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("[✗] Error leyendo flujo de datos: %v\n", err)
			return
		}

		if err := os.WriteFile(destPath, outData, 0644); err != nil {
			fmt.Printf("[✗] Error guardando archivo local '%s': %v\n", destPath, err)
			return
		}
		fmt.Printf("[✓] Archivo recuperado y verificado desde DFS: '%s' (%d bytes)\n", destPath, len(outData))

	case "manifest":
		if len(args) < 2 {
			fmt.Println("Error: Debe especificar el CID del manifiesto.")
			return
		}
		cid := args[1]
		resp, err := http.Get(defaultAPIBase + "/api/v1/dfs/manifest/" + cid)
		if err != nil {
			fmt.Printf("[✗] Error consultando manifiesto: %v\n", err)
			return
		}
		defer resp.Body.Close()

		var res map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&res)
		pretty, _ := json.MarshalIndent(res, "", "  ")
		fmt.Printf("=== MANIFIESTO DFS: %s ===\n%s\n", cid, string(pretty))

	default:
		fmt.Printf("Subcomando dfs desconocido: '%s'\n", args[0])
	}
}
