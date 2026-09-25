// Package l1 implementa el mapeo automático de puertos mediante el protocolo
// UPnP IGD (Internet Gateway Device) / SSDP para permitir conectividad P2P
// directa en nodos residenciales sin requerir configuración manual de router.
package l1

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// UPnPMapper gestiona el mapeo de puertos IGD mediante SSDP y SOAP
type UPnPMapper struct {
	timeout time.Duration
}

// NewUPnPMapper inicializa un mapeador UPnP con un tiempo límite de espera
func NewUPnPMapper(timeout time.Duration) *UPnPMapper {
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	return &UPnPMapper{timeout: timeout}
}

// DiscoverAndForward intenta mapear el puerto UDP indicado en el router local
func (u *UPnPMapper) DiscoverAndForward(port int, desc string) error {
	ctx, cancel := context.WithTimeout(context.Background(), u.timeout)
	defer cancel()
	return u.DiscoverAndForwardWithContext(ctx, port, desc)
}

// DiscoverAndForwardWithContext ejecuta el descubrimiento SSDP y mapeo SOAP con soporte de contexto
func (u *UPnPMapper) DiscoverAndForwardWithContext(ctx context.Context, port int, desc string) error {
	if port <= 0 || port > 65535 {
		return fmt.Errorf("upnp: puerto inválido: %d", port)
	}
	if desc == "" {
		desc = "ipvn7-p2p"
	}

	multicastAddr, err := net.ResolveUDPAddr("udp4", "239.255.255.250:1900")
	if err != nil {
		return fmt.Errorf("upnp: resolver multicast: %w", err)
	}

	conn, err := net.ListenUDP("udp4", nil)
	if err != nil {
		return fmt.Errorf("upnp: abrir socket ssdp: %w", err)
	}
	defer conn.Close()

	msearch := "M-SEARCH * HTTP/1.1\r\n" +
		"HOST: 239.255.255.250:1900\r\n" +
		"ST: urn:schemas-upnp-org:service:WANIPConnection:1\r\n" +
		"MAN: \"ssdp:discover\"\r\n" +
		"MX: 2\r\n\r\n"

	_ = conn.SetDeadline(time.Now().Add(u.timeout))
	if _, err := conn.WriteTo([]byte(msearch), multicastAddr); err != nil {
		return fmt.Errorf("upnp: envio ssdp: %w", err)
	}

	buf := make([]byte, 2048)
	doneCh := make(chan string, 1)

	go func() {
		for {
			n, _, err := conn.ReadFrom(buf)
			if err != nil {
				return
			}
			resp := string(buf[:n])
			for _, line := range strings.Split(resp, "\r\n") {
				if strings.HasPrefix(strings.ToUpper(line), "LOCATION:") {
					loc := strings.TrimSpace(line[len("LOCATION:"):])
					select {
					case doneCh <- loc:
					default:
					}
					return
				}
			}
		}
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("upnp: timeout o cancelacion buscando router IGD")
	case location := <-doneCh:
		return u.addPortMapping(ctx, location, port, desc)
	}
}

func (u *UPnPMapper) addPortMapping(ctx context.Context, location string, port int, desc string) error {
	if !strings.HasPrefix(location, "http://") {
		return fmt.Errorf("upnp: ubicacion descriptor invalida: %s", location)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", location, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("upnp: descarga descriptor: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("upnp: lectura xml: %w", err)
	}

	bodyStr := string(body)
	controlURLTag := "<controlURL>"
	controlURLEndTag := "</controlURL>"
	idx := strings.Index(bodyStr, controlURLTag)
	if idx == -1 {
		return fmt.Errorf("upnp: controlURL no encontrado en descriptor XML")
	}
	endIdx := strings.Index(bodyStr[idx:], controlURLEndTag)
	if endIdx == -1 {
		return fmt.Errorf("upnp: etiqueta controlURL malformada")
	}

	controlPath := bodyStr[idx+len(controlURLTag) : idx+endIdx]
	var controlURL string
	if strings.HasPrefix(controlPath, "http://") {
		controlURL = controlPath
	} else {
		parsed := strings.Split(location, "/")
		if len(parsed) >= 3 {
			baseURL := fmt.Sprintf("%s//%s", parsed[0], parsed[2])
			if !strings.HasPrefix(controlPath, "/") {
				controlPath = "/" + controlPath
			}
			controlURL = baseURL + controlPath
		}
	}

	if controlURL == "" {
		return fmt.Errorf("upnp: no se pudo derivar controlURL valida")
	}

	localIP, err := getOutboundIP()
	if err != nil {
		return fmt.Errorf("upnp: obtener ip lan: %w", err)
	}

	soapBody := fmt.Sprintf(`<?xml version="1.0"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
<s:Body>
<u:AddPortMapping xmlns:u="urn:schemas-upnp-org:service:WANIPConnection:1">
<NewRemoteHost></NewRemoteHost>
<NewExternalPort>%d</NewExternalPort>
<NewProtocol>UDP</NewProtocol>
<NewInternalPort>%d</NewInternalPort>
<NewInternalClient>%s</NewInternalClient>
<NewEnabled>1</NewEnabled>
<NewPortMappingDescription>%s</NewPortMappingDescription>
<NewLeaseDuration>3600</NewLeaseDuration>
</u:AddPortMapping>
</s:Body>
</s:Envelope>`, port, port, localIP, desc)

	soapReq, err := http.NewRequestWithContext(ctx, "POST", controlURL, bytes.NewBufferString(soapBody))
	if err != nil {
		return err
	}
	soapReq.Header.Set("Content-Type", "text/xml; charset=\"utf-8\"")
	soapReq.Header.Set("SOAPAction", "\"urn:schemas-upnp-org:service:WANIPConnection:1#AddPortMapping\"")

	soapResp, err := http.DefaultClient.Do(soapReq)
	if err != nil {
		return fmt.Errorf("upnp: peticion soap fallida: %w", err)
	}
	defer soapResp.Body.Close()

	if soapResp.StatusCode != http.StatusOK {
		return fmt.Errorf("upnp: respuesta soap codigo %d (%s)", soapResp.StatusCode, soapResp.Status)
	}

	return nil
}

func getOutboundIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String(), nil
}
