package parser

import (
	"fmt"
	"strconv"
	"strings"
)

// ParsePorts parses a port specification string into a list of unique ports.
// Supports single ports (80), ranges (1-1024), and comma-separated lists.
func ParsePorts(spec string) ([]uint16, error) {
	seen := make(map[uint16]struct{})
	var result []uint16

	for _, token := range strings.Split(spec, ",") {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}

		ports, err := parsePortToken(token)
		if err != nil {
			return nil, err
		}

		result = appendUniquePorts(result, ports, seen)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no valid ports specified")
	}

	return result, nil
}

func parsePortToken(token string) ([]uint16, error) {
	if strings.Contains(token, "-") {
		start, end, err := parsePortRange(token)
		if err != nil {
			return nil, err
		}
		return buildPortRange(start, end), nil
	}

	port, err := parseSinglePort(token)
	if err != nil {
		return nil, err
	}
	return []uint16{port}, nil
}

func parseSinglePort(value string) (uint16, error) {
	num, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || num < 1 || num > 65535 {
		return 0, fmt.Errorf("invalid port: %s", value)
	}
	return uint16(num), nil
}

func parsePortRange(token string) (int, int, error) {
	parts := strings.Split(token, "-")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid port range: %s", token)
	}

	start, err := parseRangeBoundary(parts[0], "start")
	if err != nil {
		return 0, 0, err
	}

	end, err := parseRangeBoundary(parts[1], "end")
	if err != nil {
		return 0, 0, err
	}

	if start > end {
		return 0, 0, fmt.Errorf("invalid port range: start > end in %s", token)
	}

	return start, end, nil
}

func parseRangeBoundary(value string, position string) (int, error) {
	num, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || num < 1 || num > 65535 {
		return 0, fmt.Errorf("invalid %s port in range: %s", position, value)
	}
	return num, nil
}

func buildPortRange(start, end int) []uint16 {
	ports := make([]uint16, 0, end-start+1)
	for p := start; p <= end && p <= 65535; p++ {
		ports = append(ports, uint16(p))
	}
	return ports
}

func appendUniquePorts(dest []uint16, ports []uint16, seen map[uint16]struct{}) []uint16 {
	for _, port := range ports {
		if _, exists := seen[port]; exists {
			continue
		}
		dest = append(dest, port)
		seen[port] = struct{}{}
	}
	return dest
}
