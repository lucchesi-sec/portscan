package parser

import (
	"fmt"
	"strconv"
	"strings"
)

func ParsePorts(spec string) ([]uint16, error) {
	var ports []uint16
	seen := make(map[uint16]bool)

	parts := strings.Split(spec, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid port range: %s", part)
			}

			start, err := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
			if err != nil || start < 1 || start > 65535 {
				return nil, fmt.Errorf("invalid start port in range: %s", rangeParts[0])
			}

			end, err := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
			if err != nil || end < 1 || end > 65535 {
				return nil, fmt.Errorf("invalid end port in range: %s", rangeParts[1])
			}

			if start > end {
				return nil, fmt.Errorf("invalid port range: start > end in %s", part)
			}

			// Ensure we don't overflow when converting to uint16
			for p := start; p <= end && p <= 65535; p++ {
				port := uint16(p)
				if !seen[port] {
					ports = append(ports, port)
					seen[port] = true
				}
			}
		} else {
			port, err := strconv.Atoi(part)
			if err != nil || port < 1 || port > 65535 {
				return nil, fmt.Errorf("invalid port: %s", part)
			}
			if !seen[uint16(port)] {
				ports = append(ports, uint16(port))
				seen[uint16(port)] = true
			}
		}
	}

	if len(ports) == 0 {
		return nil, fmt.Errorf("no valid ports specified")
	}

	return ports, nil
}
