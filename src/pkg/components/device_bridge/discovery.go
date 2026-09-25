package devicebridge

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"ipvn7/pkg/core"
)

// DiscoverHomeDevices realiza un descubrimiento coordinado y autónomo de la red doméstica
func DiscoverHomeDevices(ctx context.Context, cfg BridgeConfig) ([]*DiscoveredDevice, error) {
	deviceMap := make(map[string]*DiscoveredDevice)

	// 1. Descubrimiento pasivo SSDP / UPnP multicast
	if cfg.EnableSSDP {
		ssdpDevices := discoverSSDP(ctx, 2*time.Second)
		for _, d := range ssdpDevices {
			deviceMap[d.ID] = d
		}
	}

	// 2. Sondeo activo de servicios en hosts locales detectados
	if cfg.EnablePortScan {
		probeDevices := probeCommonLocalDevices(ctx)
		for _, d := range probeDevices {
			key := fmt.Sprintf("%s:%d", d.IPv4, d.Port)
			if existing, ok := deviceMap[key]; ok {
				// Enriquecer capacidades si ya existía
				existing.Capabilities = append(existing.Capabilities, d.Capabilities...)
			} else {
				deviceMap[key] = d
			}
		}
	}

	result := make([]*DiscoveredDevice, 0, len(deviceMap))
	for _, dev := range deviceMap {
		result = append(result, dev)
	}

	return result, nil
}

type upnpXML struct {
	XMLName xml.Name `xml:"root"`
	Device  struct {
		FriendlyName string `xml:"friendlyName"`
		Manufacturer string `xml:"manufacturer"`
		ModelName    string `xml:"modelName"`
		DeviceType   string `xml:"deviceType"`
	} `xml:"device"`
}

func discoverSSDP(ctx context.Context, timeout time.Duration) []*DiscoveredDevice {
	var devices []*DiscoveredDevice

	raddr, err := net.ResolveUDPAddr("udp4", "239.255.255.250:1900")
	if err != nil {
		return devices
	}

	conn, err := net.ListenUDP("udp4", nil)
	if err != nil {
		return devices
	}
	defer conn.Close()

	msearch := []byte(
		"M-SEARCH * HTTP/1.1\r\n" +
			"HOST: 239.255.255.250:1900\r\n" +
			"MAN: \"ssdp:discover\"\r\n" +
			"MX: 1\r\n" +
			"ST: ssdp:all\r\n\r\n",
	)

	_ = conn.SetDeadline(time.Now().Add(timeout))
	_, _ = conn.WriteTo(msearch, raddr)

	buf := make([]byte, 2048)
	seen := make(map[string]bool)

	for {
		select {
		case <-ctx.Done():
			return devices
		default:
		}

		n, src, err := conn.ReadFromUDP(buf)
		if err != nil {
			break
		}

		ip := src.IP.String()
		if seen[ip] {
			continue
		}

		respStr := string(buf[:n])
		headers := parseHTTPHeaders(respStr)

		location := headers["location"]
		dev := queryDeviceMetadata(ctx, ip, src.Port, location, headers["server"], headers["st"])
		if dev != nil {
			seen[ip] = true
			devices = append(devices, dev)
		}
	}

	return devices
}

func queryDeviceMetadata(ctx context.Context, ip string, port int, location, serverHeader, stHeader string) *DiscoveredDevice {
	client := &http.Client{Timeout: 400 * time.Millisecond}

	// Extraer puerto de servicio real de LOCATION (o descartar puerto efímero UDP)
	realPort := port
	if location != "" {
		if u, err := url.Parse(location); err == nil && u.Port() != "" {
			if p, err := strconv.Atoi(u.Port()); err == nil && p > 0 {
				realPort = p
			}
		}
	}
	if realPort > 32767 || realPort == 0 {
		realPort = 8008
	}

	// Si hay ubicación XML, consultar el descriptor oficial del dispositivo
	if location != "" {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, location, nil)
		if resp, err := client.Do(req); err == nil {
			defer resp.Body.Close()
			data, _ := io.ReadAll(resp.Body)
			var doc upnpXML
			if xml.Unmarshal(data, &doc) == nil && doc.Device.FriendlyName != "" {
				name := doc.Device.FriendlyName
				if doc.Device.Manufacturer != "" {
					name = fmt.Sprintf("%s (%s)", doc.Device.FriendlyName, doc.Device.Manufacturer)
				}
				cat := CategoryIoT
				caps := []string{"ssdp:alive", fmt.Sprintf("model:%s", doc.Device.ModelName)}

				lowerType := strings.ToLower(doc.Device.DeviceType)
				if strings.Contains(lowerType, "dial") || strings.Contains(lowerType, "mediarenderer") || strings.Contains(lowerType, "tv") {
					cat = CategorySmartTV
					caps = append(caps, "tv:cast", "tv:display")
				} else if strings.Contains(lowerType, "printer") {
					cat = CategoryPrinter
					caps = append(caps, "printer:spool")
				}
				return buildDeviceModel(ip, realPort, name, cat, "UPnP/SSDP", caps)
			}
		}
	}

	// Clasificación heurística según cabeceras
	lowerServer := strings.ToLower(serverHeader)
	lowerST := strings.ToLower(stHeader)
	if strings.Contains(lowerST, "dial") || strings.Contains(lowerServer, "dial") || strings.Contains(lowerST, "mediarenderer") {
		return buildDeviceModel(ip, realPort, fmt.Sprintf("Smart TV / Cast (%s)", ip), CategorySmartTV, "DIAL", []string{"tv:cast"})
	}
	if strings.Contains(lowerST, "printer") || strings.Contains(lowerServer, "printer") {
		return buildDeviceModel(ip, realPort, fmt.Sprintf("Impresora de Red (%s)", ip), CategoryPrinter, "IPP", []string{"ipp:print"})
	}
	return nil
}

// probeCommonLocalDevices sondea activamente los hosts de la tabla ARP del sistema operativo
func probeCommonLocalDevices(ctx context.Context) []*DiscoveredDevice {
	var devices []*DiscoveredDevice

	// Pre-calentar dinámicamente la tabla ARP del kernel del sistema operativo
	core.PrewarmARPTable(ctx)

	arpEntries := ReadLocalARPTable()
	if len(arpEntries) == 0 {
		return devices
	}

	client := &http.Client{Timeout: 300 * time.Millisecond}

	for _, entry := range arpEntries {
		select {
		case <-ctx.Done():
			return devices
		default:
		}

		ip := entry.IP
		mac := entry.MAC

		// 1. Comprobar si es una impresora IPP (Puerto 631)
		if isPortOpen(ctx, ip, 631) {
			printerName := fmt.Sprintf("Impresora IPP (%s)", ip)
			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://%s:631/", ip), nil)
			if resp, err := client.Do(req); err == nil {
				_ = resp.Body.Close()
				serverHdr := resp.Header.Get("Server")
				if serverHdr != "" {
					printerName = cleanServerHeader(serverHdr)
				}
			}
			dev := buildDeviceModel(ip, 631, printerName, CategoryPrinter, "IPP", []string{"ipp:print", "cups:spool", fmt.Sprintf("mac:%s", mac)})
			devices = append(devices, dev)
			continue
		}

		// 2. Comprobar si es una Smart TV / Receptor Cast (Puerto 8008 o 8009)
		if isPortOpen(ctx, ip, 8008) || isPortOpen(ctx, ip, 8009) {
			tvName := fmt.Sprintf("Smart TV (%s)", ip)
			caps := []string{"tv:cast", fmt.Sprintf("mac:%s", mac)}
			port := 8008
			if isPortOpen(ctx, ip, 8009) {
				caps = append(caps, "google:cast")
				port = 8009
			}

			// Intentar leer nombre oficial desde el XML de DIAL
			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://%s:8008/ssdp/device-desc.xml", ip), nil)
			if resp, err := client.Do(req); err == nil {
				defer resp.Body.Close()
				data, _ := io.ReadAll(resp.Body)
				var doc upnpXML
				if xml.Unmarshal(data, &doc) == nil && doc.Device.FriendlyName != "" {
					if doc.Device.Manufacturer != "" {
						tvName = fmt.Sprintf("%s (%s)", doc.Device.FriendlyName, doc.Device.Manufacturer)
					} else {
						tvName = doc.Device.FriendlyName
					}
				}
			}

			dev := buildDeviceModel(ip, port, tvName, CategorySmartTV, "DIAL/Cast", caps)
			devices = append(devices, dev)
			continue
		}

		// 3. Comprobar impresora RAW (Puerto 9100)
		if isPortOpen(ctx, ip, 9100) {
			dev := buildDeviceModel(ip, 9100, fmt.Sprintf("Impresora RAW JetDirect (%s)", ip), CategoryPrinter, "RAW_9100", []string{"raw:print", fmt.Sprintf("mac:%s", mac)})
			devices = append(devices, dev)
			continue
		}

		// 4. Dispositivo Web general o Router (Puerto 80)
		if isPortOpen(ctx, ip, 80) {
			devName := fmt.Sprintf("Dispositivo Web (%s)", ip)
			cat := CategoryIoT
			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://%s:80/", ip), nil)
			if resp, err := client.Do(req); err == nil {
				_ = resp.Body.Close()
				hdr := strings.ToLower(resp.Header.Get("Server"))
				if strings.Contains(hdr, "router") || strings.Contains(hdr, "gateway") || strings.HasSuffix(ip, ".1") {
					cat = CategoryGateway
					devName = fmt.Sprintf("Router Gateway (%s)", ip)
				}
			}
			dev := buildDeviceModel(ip, 80, devName, cat, "HTTP", []string{"http:web", fmt.Sprintf("mac:%s", mac)})
			devices = append(devices, dev)
		}
	}

	return devices
}

func isPortOpen(ctx context.Context, ip string, port int) bool {
	d := net.Dialer{Timeout: 150 * time.Millisecond}
	conn, err := d.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", ip, port))
	if err == nil {
		_ = conn.Close()
		return true
	}
	return false
}

func cleanServerHeader(hdr string) string {
	parts := strings.Split(hdr, ";")
	if len(parts) > 1 {
		return strings.TrimSpace(parts[1])
	}
	return strings.TrimSpace(parts[0])
}

type ARPEntry = core.ARPEntry

func ReadLocalARPTable() []ARPEntry {
	return core.ReadLocalARPTable()
}

func buildDeviceModel(ip string, port int, name string, cat DeviceCategory, protocol string, caps []string) *DiscoveredDevice {
	h := sha256.Sum256([]byte(fmt.Sprintf("ipvn7:device:shadow:%s:%s", ip, cat)))
	shadowDID := fmt.Sprintf("did:ipvn7:shadow:%s", hex.EncodeToString(h[:16]))

	petnamePrefix := "device"
	switch cat {
	case CategoryPrinter:
		petnamePrefix = "printer"
	case CategorySmartTV, CategoryCast:
		petnamePrefix = "tv"
	case CategoryGateway:
		petnamePrefix = "gateway"
	}

	petname := fmt.Sprintf("%s.%s.ipv7", petnamePrefix, strings.ReplaceAll(ip, ".", "-"))

	return &DiscoveredDevice{
		ID:           fmt.Sprintf("dev_%s", strings.ReplaceAll(ip, ".", "_")),
		Name:         name,
		Category:     cat,
		IPv4:         ip,
		Port:         port,
		Protocol:     protocol,
		ShadowDID:    shadowDID,
		Petname:      petname,
		LastSeen:     time.Now(),
		Capabilities: caps,
	}
}

func parseHTTPHeaders(raw string) map[string]string {
	headers := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(raw))
	for scanner.Scan() {
		line := scanner.Text()
		idx := strings.Index(line, ":")
		if idx > 0 {
			k := strings.TrimSpace(strings.ToLower(line[:idx]))
			v := strings.TrimSpace(line[idx+1:])
			headers[k] = v
		}
	}
	return headers
}
