package core

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

// ARPEntry representa un registro de dirección IP y MAC física en el sistema operativo
type ARPEntry struct {
	IP  string `json:"ip"`
	MAC string `json:"mac"`
}

// PrewarmARPTable sondea concurrentemente la subred IPv4 local calculada desde las interfaces del sistema,
// forzando al kernel del sistema operativo a resolver las direcciones MAC vía ARP de forma 100% autónoma.
func PrewarmARPTable(ctx context.Context) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return
	}

	var wg sync.WaitGroup
	for _, a := range addrs {
		ipNet, ok := a.(*net.IPNet)
		if !ok || ipNet.IP.IsLoopback() {
			continue
		}
		ip4 := ipNet.IP.To4()
		if ip4 == nil {
			continue
		}

		mask := ipNet.Mask
		if len(mask) == 4 && mask[0] == 255 && mask[1] == 255 && mask[2] == 255 {
			base := ip4.Mask(mask)
			for i := 1; i < 255; i++ {
				targetIP := net.IPv4(base[0], base[1], base[2], byte(i)).String()
				if targetIP == ip4.String() {
					continue
				}
				wg.Add(1)
				go func(ip string) {
					defer wg.Done()
					d := net.Dialer{Timeout: 120 * time.Millisecond}
					conn, err := d.DialContext(ctx, "udp4", fmt.Sprintf("%s:9", ip))
					if err == nil {
						_, _ = conn.Write([]byte{0})
						_ = conn.Close()
					}
				}(targetIP)
			}
		}
	}
	wg.Wait()
}

// ReadLocalARPTable lee las entradas dinámicas activas de la tabla ARP real del host
func ReadLocalARPTable() []ARPEntry {
	var entries []ARPEntry
	out, err := exec.Command("arp", "-a").Output()
	if err != nil {
		return entries
	}

	scanner := bufio.NewScanner(bytes.NewReader(out))
	ipRegex := regexp.MustCompile(`(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\s+([0-9a-fA-F[:-]{11,17})`)

	for scanner.Scan() {
		line := scanner.Text()
		matches := ipRegex.FindStringSubmatch(line)
		if len(matches) == 3 {
			ip := matches[1]
			mac := strings.ToLower(strings.ReplaceAll(matches[2], "-", ":"))
			if strings.HasPrefix(ip, "224.") || strings.HasPrefix(ip, "239.") || strings.HasSuffix(ip, ".255") || mac == "ff:ff:ff:ff:ff:ff" {
				continue
			}
			entries = append(entries, ARPEntry{IP: ip, MAC: mac})
		}
	}
	return entries
}

// GetLocalBroadcasts calcula dinámicamente las direcciones de broadcast de todas las interfaces IPv4 activas
func GetLocalBroadcasts() []string {
	broadcasts := []string{"255.255.255.255"}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return broadcasts
	}

	for _, a := range addrs {
		if ipNet, ok := a.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ip4 := ipNet.IP.To4(); ip4 != nil {
				mask := ipNet.Mask
				if len(mask) == 4 {
					bcast := make(net.IP, 4)
					for i := 0; i < 4; i++ {
						bcast[i] = ip4[i] | ^mask[i]
					}
					bcastStr := bcast.String()
					if bcastStr != "255.255.255.255" {
						broadcasts = append(broadcasts, bcastStr)
					}
				}
			}
		}
	}
	return broadcasts
}

var ipExtractRegex = regexp.MustCompile(`\b(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\b`)

// ExtractTargetIP extrae una dirección IPv4 válida presente en el texto o URL
func ExtractTargetIP(input string) string {
	matches := ipExtractRegex.FindStringSubmatch(input)
	if len(matches) > 1 {
		return matches[1]
	}
	return strings.TrimSpace(input)
}
