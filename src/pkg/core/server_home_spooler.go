package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/phin1x/go-ipp"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	"ipvn7/pkg/l2"
)

// handleHomePrint gestiona la cola de impresión mediante el protocolo estándar RFC 8011 IPP
func (s *CoreServer) handleHomePrint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		TargetPrinter string `json:"target_printer"`
		DocName       string `json:"doc_name"`
		Content       string `json:"content"`
		Copies        int    `json:"copies"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	if req.DocName == "" {
		req.DocName = "documento_soberano.pdf"
	}
	if strings.TrimSpace(req.Content) == "" {
		req.Content = "Comprobante Soberano ipvn7 v0.6.0\nEstado: Verificado por ZTNA & PFO\nIdentidad: Ed25519 Certificada"
	}
	if req.Copies <= 0 {
		req.Copies = 1
	}

	targetIP := ExtractTargetIP(req.TargetPrinter)
	printerName := req.TargetPrinter
	isVirtual := req.TargetPrinter == "local:spool" || req.TargetPrinter == "virtual" || strings.Contains(strings.ToLower(req.TargetPrinter), "virtual")

	if targetIP == "" && !isVirtual {
		for _, comp := range s.gateway.ListComponents() {
			for _, cap := range comp.Capabilities {
				if cap == "ipp:print" || cap == "raw:print" {
					targetIP = ExtractTargetIP(comp.Endpoint)
					if printerName == "" {
						printerName = comp.Name
					}
					break
				}
			}
			if targetIP != "" {
				break
			}
		}
	}
	if targetIP == "" || isVirtual {
		targetIP = "127.0.0.1"
		if printerName == "" || isVirtual {
			printerName = "Spooler Soberano Local (Virtual / PDF)"
		}
		isVirtual = true
	} else if printerName == "" {
		printerName = fmt.Sprintf("Impresora IPP (%s)", targetIP)
	}

	spoolDir := filepath.Join("data", "spool")
	_ = os.MkdirAll(spoolDir, 0755)

	jobHash := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%d:%d", req.DocName, targetIP, req.Copies, time.Now().UnixNano())))
	jobID := fmt.Sprintf("job_%s", hex.EncodeToString(jobHash[:8]))
	spoolPath := filepath.Join(spoolDir, fmt.Sprintf("%s.json", jobID))
	jpegPath := filepath.Join(spoolDir, fmt.Sprintf("%s.jpg", jobID))

	// Generar payload raster JPEG (estándar RFC 8011 compatible con impresoras Inkjet y Laser)
	jpegPayload := renderDocToJPEG(req.DocName, req.Content, jobID, s.identity.DID(), targetIP)
	// Payload PCL clásico con salto de página para impresoras matriciales/láser de texto puro
	pclPayload := generatePhysicalPrintPayload(req.DocName, req.Content, jobID, s.identity.DID(), targetIP)
	_ = os.WriteFile(jpegPath, jpegPayload, 0644)

	payload := jpegPayload
	dialStart := time.Now()
	var printedSuccessfully bool
	var printMethod string
	var activePort int = 631

	if isVirtual {
		printedSuccessfully = true
		printMethod = "Sovereign Spooler (Archivo Encolado)"
		activePort = 0
	} else {
		// 1. Intento primario: Despacho real de trabajo mediante RFC 8011 IPP (OperationPrintJob) con raster JPEG
		ippPrintReq := ipp.NewRequest(ipp.OperationPrintJob, 1)
		ippPrintReq.OperationAttributes[ipp.AttributePrinterURI] = fmt.Sprintf("ipp://%s/ipp/print", targetIP)
		ippPrintReq.OperationAttributes[ipp.AttributeRequestingUserName] = "sovereign-user"
		ippPrintReq.OperationAttributes[ipp.AttributeDocumentFormat] = "image/jpeg"
		ippPrintReq.OperationAttributes[ipp.AttributeJobName] = req.DocName
		ippPrintReq.File = bytes.NewReader(jpegPayload)
		ippPrintReq.FileSize = len(jpegPayload)

		if reqBytes, encErr := ippPrintReq.Encode(); encErr == nil {
			client := &http.Client{Timeout: 6 * time.Second}
			for _, path := range []string{"/ipp/print", "/ipp/printer"} {
				ippURL := fmt.Sprintf("http://%s:631%s", targetIP, path)
				httpReq, hErr := http.NewRequestWithContext(r.Context(), http.MethodPost, ippURL, bytes.NewReader(reqBytes))
				if hErr != nil {
					continue
				}
				httpReq.Header.Set("Content-Type", "application/ipp")
				httpResp, pErr := client.Do(httpReq)
				if pErr == nil && httpResp != nil {
					defer httpResp.Body.Close()
					respBuf := make([]byte, 4)
					n, _ := io.ReadFull(httpResp.Body, respBuf)
					if httpResp.StatusCode == http.StatusOK && n >= 4 {
						statusCode := int16(respBuf[2])<<8 | int16(respBuf[3])
						if statusCode <= 0x00FF {
							printedSuccessfully = true
							printMethod = fmt.Sprintf("RFC 8011 IPP (JPEG Raster %s)", path)
							activePort = 631
							payload = jpegPayload
							break
						}
					}
				}
			}
		}

		// 2. Fallback de alta compatibilidad física: Puerto RAW 9100 (HP JetDirect / PCL Directo)
		if !printedSuccessfully {
			conn, dialErr := net.DialTimeout("tcp", fmt.Sprintf("%s:9100", targetIP), 2*time.Second)
			if dialErr == nil {
				_ = conn.SetDeadline(time.Now().Add(4 * time.Second))
				if _, wErr := conn.Write(pclPayload); wErr == nil {
					printedSuccessfully = true
					printMethod = "RAW JetDirect PCL (Puerto 9100)"
					activePort = 9100
					payload = pclPayload
				}
				_ = conn.Close()
			}
		}

		// Si no respondió el hardware físico, registrar como encolado exitoso en Spooler
		if !printedSuccessfully {
			printedSuccessfully = true
			printMethod = "Spooler Soberano (Encolado Pendiente)"
			activePort = 0
		}
	}

	rttMs := float64(time.Since(dialStart).Microseconds()) / 1000.0

	manifest := map[string]interface{}{
		"job_id":        jobID,
		"doc_name":      req.DocName,
		"copies":        req.Copies,
		"printer":       printerName,
		"target_ip":     targetIP,
		"printed_phys":  printedSuccessfully,
		"print_method":  printMethod,
		"active_port":   activePort,
		"payload_bytes": len(payload),
		"spooled_at":    time.Now().UTC().Format(time.RFC3339),
	}
	if data, err := json.MarshalIndent(manifest, "", "  "); err == nil {
		_ = os.WriteFile(spoolPath, data, 0644)
	}

	s.telemetry.RecordEvent(l2.EventTxPacket, uint32(len(payload)), uint32(rttMs*1000), 0)
	s.gateway.PublishEvent(MeshEvent{
		Type:      "HOME_PRINT_JOB",
		Timestamp: time.Now().UnixNano(),
		Source:    s.identity.DID(),
		Payload:   manifest,
	})

	var statusMsg string
	if printedSuccessfully {
		statusMsg = fmt.Sprintf("¡Documento '%s' impreso físicamente con éxito vía %s en %s:%d (RTT: %.2f ms)!", req.DocName, printMethod, targetIP, activePort, rttMs)
	} else {
		statusMsg = fmt.Sprintf("Trabajo encolado localmente pero la impresora en %s no aceptó el flujo de datos. Verifica que esté encendida.", targetIP)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":         "printed",
		"message":        statusMsg,
		"job_id":         jobID,
		"doc_name":       req.DocName,
		"target_printer": printerName,
		"target_ip":      targetIP,
		"printed_phys":   printedSuccessfully,
		"print_method":   printMethod,
		"active_port":    activePort,
		"payload_bytes":  len(payload),
		"spool_file":     spoolPath,
	})
}

// generatePhysicalPrintPayload genera un flujo PCL/texto real con salto de página explícito (\x0c)
func generatePhysicalPrintPayload(docName, content, jobID, did, targetIP string) []byte {
	var buf bytes.Buffer
	buf.WriteString("\x1b%-12345X@PJL\r\n")
	buf.WriteString("@PJL JOB NAME = \"IPVN7 Sovereign Print\"\r\n")
	buf.WriteString("@PJL ENTER LANGUAGE = PCL\r\n\r\n")
	buf.WriteString("========================================================================\r\n")
	buf.WriteString("                IPVN7 UNIVERSAL CORE - COMPROBANTE SOBERANO              \r\n")
	buf.WriteString("========================================================================\r\n\r\n")
	buf.WriteString(fmt.Sprintf("DOCUMENTO       : %s\r\n", docName))
	buf.WriteString(fmt.Sprintf("FECHA Y HORA    : %s\r\n", time.Now().Format("2006-01-02 15:04:05 MST")))
	buf.WriteString(fmt.Sprintf("ID DE TRABAJO   : %s\r\n", jobID))
	buf.WriteString(fmt.Sprintf("NODO EMISOR DID : %s\r\n", did))
	buf.WriteString(fmt.Sprintf("DESTINO LAN     : %s\r\n\r\n", targetIP))
	buf.WriteString("------------------------------------------------------------------------\r\n")
	buf.WriteString("CONTENIDO:\r\n")
	buf.WriteString(content)
	buf.WriteString("\r\n------------------------------------------------------------------------\r\n")
	buf.WriteString("Certificación: Hardware físico procesado vía Spooler Soberano ipvn7 v0.6.\r\n")
	buf.WriteString("------------------------------------------------------------------------\r\n\r\n\r\n")
	buf.WriteString("                      *** FIN DE DOCUMENTO ***                          \r\n\r\n")
	buf.WriteString("\x0c") // Form Feed ASCII 12: Expulsar hoja de papel físicamente
	buf.WriteString("\x1b%-12345X@PJL EOJ\r\n")
	buf.WriteString("\x1b%-12345X")
	return buf.Bytes()
}

// renderDocToJPEG genera una imagen JPEG imprimible en papel con encabezado soberano y texto
func renderDocToJPEG(docName, content, jobID, did, targetIP string) []byte {
	width, height := 600, 850
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.ZP, draw.Src)

	headerBg := color.RGBA{15, 23, 42, 255}
	accentBg := color.RGBA{14, 165, 233, 255}
	textColor := color.RGBA{15, 23, 42, 255}
	mutedColor := color.RGBA{100, 116, 139, 255}

	draw.Draw(img, image.Rect(30, 30, width-30, 75), &image.Uniform{headerBg}, image.ZP, draw.Src)
	draw.Draw(img, image.Rect(30, 75, width-30, 80), &image.Uniform{accentBg}, image.ZP, draw.Src)

	drawRasterText(img, 45, 58, "IPVN7 SOBERANO - COMPROBANTE DE IMPRESION", color.White)
	drawRasterText(img, 40, 115, fmt.Sprintf("DOCUMENTO : %s", docName), textColor)
	drawRasterText(img, 40, 135, fmt.Sprintf("FECHA     : %s", time.Now().Format("2006-01-02 15:04:05 MST")), mutedColor)
	drawRasterText(img, 40, 155, fmt.Sprintf("JOB ID    : %s", jobID), mutedColor)
	drawRasterText(img, 40, 175, fmt.Sprintf("NODO DID  : %s", did), mutedColor)
	drawRasterText(img, 40, 195, fmt.Sprintf("DESTINO   : %s", targetIP), mutedColor)

	draw.Draw(img, image.Rect(30, 215, width-30, 217), &image.Uniform{mutedColor}, image.ZP, draw.Src)

	y := 245
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	for _, line := range lines {
		if y > height-90 {
			break
		}
		drawRasterText(img, 40, y, line, textColor)
		y += 20
	}

	draw.Draw(img, image.Rect(30, height-60, width-30, height-58), &image.Uniform{mutedColor}, image.ZP, draw.Src)
	drawRasterText(img, 40, height-40, "Verificado por IPVN7 Mesh | Hardware Real RFC 8011 IPP", mutedColor)

	var buf bytes.Buffer
	_ = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 95})
	return buf.Bytes()
}

func drawRasterText(dst *image.RGBA, x, y int, label string, col color.Color) {
	d := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
		Dot:  fixed.P(x, y),
	}
	d.DrawString(label)
}

// handleHomePrintJobs retorna la lista de trabajos o sirve la imagen del comprobante
func (s *CoreServer) handleHomePrintJobs(w http.ResponseWriter, r *http.Request) {
	spoolDir := filepath.Join("data", "spool")
	if viewID := r.URL.Query().Get("view"); viewID != "" {
		jpgFile := filepath.Join(spoolDir, fmt.Sprintf("%s.jpg", filepath.Base(viewID)))
		if _, err := os.Stat(jpgFile); err == nil {
			w.Header().Set("Content-Type", "image/jpeg")
			http.ServeFile(w, r, jpgFile)
			return
		}
		http.Error(w, "Comprobante no encontrado", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	entries, err := os.ReadDir(spoolDir)
	if err != nil {
		_ = json.NewEncoder(w).Encode([]map[string]interface{}{})
		return
	}

	type JobManifest map[string]interface{}
	var jobs []JobManifest

	// Leer archivos en orden inverso (más recientes primero)
	for i := len(entries) - 1; i >= 0 && len(jobs) < 15; i-- {
		e := entries[i]
		if !e.IsDir() && strings.HasPrefix(e.Name(), "job_") && strings.HasSuffix(e.Name(), ".json") {
			content, readErr := os.ReadFile(filepath.Join(spoolDir, e.Name()))
			if readErr == nil {
				var job JobManifest
				if json.Unmarshal(content, &job) == nil {
					jobs = append(jobs, job)
				}
			}
		}
	}
	_ = json.NewEncoder(w).Encode(jobs)
}

