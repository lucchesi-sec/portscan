package core

import (
	"fmt"
	"math"
	"strings"

	"github.com/lucchesi-sec/portscan/pkg/services"
)

func (s *UDPScanner) parseUDPResponse(port uint16, data []byte) string {
	var service string
	var confidence float64

	switch port {
	case 53:
		service, confidence = parseDNSResponseWithConfidence(data)
	case 123:
		service, confidence = parseNTPResponseWithConfidence(data)
	case 161:
		service, confidence = parseSNMPResponseWithConfidence(data)
	case 67, 68:
		service, confidence = "DHCP", 0.8
	case 137:
		service, confidence = parseNetBIOSResponseWithConfidence(data)
	case 1194:
		if len(data) > 0 {
			service, confidence = "OpenVPN", 0.7
		} else {
			service, confidence = "OpenVPN (no response)", 0.3
		}
	case 500, 4500:
		service, confidence = "IKE/IPSec", 0.8
	case 51820:
		service, confidence = "WireGuard", 0.7
	case 5353:
		service, confidence = "mDNS/Bonjour", 0.8
	default:
		return describeUnknownUDP(port, data)
	}

	if service != "" {
		return fmt.Sprintf("%s (%.1f%%)", service, confidence*100)
	}
	return ""
}

func describeUnknownUDP(port uint16, data []byte) string {
	banner := sanitizeBanner(data)
	if banner != "" {
		confidence := math.Min(0.5+float64(len(banner))/100.0, 0.9)
		return fmt.Sprintf("%s (%.1f%%)", banner, confidence*100)
	}

	serviceName := services.GetName(port)
	if serviceName != "unknown" {
		return fmt.Sprintf("?%s (30.0%%)", serviceName)
	}
	return ""
}

func sanitizeBanner(data []byte) string {
	banner := string(data)
	banner = strings.Map(func(r rune) rune {
		if r >= 32 && r < 127 {
			return r
		}
		if (r >= 160 && r <= 591) || (r >= 880 && r <= 1023) || (r >= 1024 && r <= 1279) {
			return r
		}
		return -1
	}, banner)

	if len(banner) > 64 {
		banner = banner[:64] + "..."
	}
	return banner
}

func parseDNSResponseWithConfidence(data []byte) (string, float64) {
	if len(data) > 12 {
		if len(data) >= 4 && data[2]&0x80 != 0 {
			return "DNS", 0.95
		}
		return "DNS", 0.8
	}
	return "DNS (no response)", 0.3
}

func parseNTPResponseWithConfidence(data []byte) (string, float64) {
	if len(data) >= 48 {
		liVnMode := data[0]
		version := (liVnMode >> 3) & 0x07
		mode := liVnMode & 0x07

		if version >= 2 && version <= 4 && mode == 4 {
			return "NTP", 0.95
		}
		return "NTP", 0.8
	}
	return "NTP (invalid response)", 0.2
}

func parseSNMPResponseWithConfidence(data []byte) (string, float64) {
	if len(data) > 0 {
		if data[0] == 0x30 {
			if len(data) >= 3 && data[1] < 0x80 {
				return "SNMP", 0.9
			}
			return "SNMP", 0.7
		}
		if data[0] == 0xA0 || data[0] == 0xA1 || data[0] == 0xA2 {
			return "SNMP (error)", 0.8
		}
	}
	return "SNMP (no response)", 0.2
}

func parseNetBIOSResponseWithConfidence(data []byte) (string, float64) {
	if len(data) > 12 {
		if len(data) >= 4 {
			flags := uint16(data[2])<<8 | uint16(data[3])
			if flags&0x8000 != 0 {
				return "NetBIOS Name Service", 0.9
			}
		}
		return "NetBIOS Name Service", 0.7
	}
	return "NetBIOS", 0.3
}
