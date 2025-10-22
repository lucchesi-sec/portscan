package targets

import (
	"fmt"
	"net"
	"strings"
)

const defaultCIDRHostLimit = 65536

// Options customises target resolution behaviours.
type Options struct {
	// CIDRHostLimit restricts the maximum number of hosts produced by a single CIDR.
	// Defaults to defaultCIDRHostLimit when zero or negative.
	CIDRHostLimit int
}

// Resolve normalises a list of user-provided targets (hosts, IPs, CIDRs) into a
// deduplicated slice of scan-ready host strings.
func Resolve(inputs []string, opts Options) ([]string, error) {
	limit := opts.CIDRHostLimit
	if limit <= 0 {
		limit = defaultCIDRHostLimit
	}

	seen := make(map[string]struct{})
	var resolved []string

	for _, raw := range inputs {
		token := strings.TrimSpace(raw)
		if token == "" {
			continue
		}

		expanded, err := expandToken(token, limit)
		if err != nil {
			return nil, err
		}

		for _, host := range expanded {
			if _, exists := seen[host]; exists {
				continue
			}
			seen[host] = struct{}{}
			resolved = append(resolved, host)
		}
	}

	if len(resolved) == 0 {
		return nil, fmt.Errorf("no valid targets provided")
	}

	return resolved, nil
}

func expandToken(token string, limit int) ([]string, error) {
	if ip := net.ParseIP(token); ip != nil {
		return []string{token}, nil
	}

	if strings.Contains(token, "/") {
		_, network, err := net.ParseCIDR(token)
		if err != nil {
			return nil, fmt.Errorf("invalid CIDR %q: %w", token, err)
		}
		return expandCIDR(network, limit)
	}

	if err := validateHostname(token); err != nil {
		return nil, fmt.Errorf("invalid hostname %q: %w", token, err)
	}

	return []string{token}, nil
}

func expandCIDR(network *net.IPNet, limit int) ([]string, error) {
	hostCount, err := cidrHostCount(network)
	if err != nil {
		return nil, err
	}

	if hostCount > uint64(limit) {
		return nil, fmt.Errorf("CIDR %q expands to %d hosts (limit %d)", network.String(), hostCount, limit)
	}

	current := make(net.IP, len(network.IP))
	copy(current, network.IP.Mask(network.Mask))

	hosts := make([]string, 0, hostCount)
	for i := uint64(0); i < hostCount; i++ {
		if !network.Contains(current) {
			break
		}
		hosts = append(hosts, current.String())
		incrementIP(current)
	}

	return hosts, nil
}

func cidrHostCount(network *net.IPNet) (uint64, error) {
	ones, bits := network.Mask.Size()
	if ones == 0 && bits == 0 {
		return 0, fmt.Errorf("unable to determine size for network %q", network.String())
	}

	power := bits - ones
	if power < 0 {
		return 0, fmt.Errorf("invalid CIDR mask for %q", network.String())
	}

	if power >= 63 {
		// Limit to avoid overflow; the caller enforces a stricter bound anyway.
		return ^uint64(0), nil
	}

	return 1 << uint(power), nil
}

func incrementIP(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] != 0 {
			break
		}
	}
}

func validateHostname(hostname string) error {
	if err := validateHostnameLength(hostname); err != nil {
		return err
	}
	if err := validateHostnameEdges(hostname); err != nil {
		return err
	}
	if strings.Contains(hostname, "..") {
		return fmt.Errorf("hostname cannot contain consecutive '.' characters")
	}

	for _, label := range strings.Split(hostname, ".") {
		if err := validateHostnameLabel(label); err != nil {
			return err
		}
	}
	return nil
}

func validateHostnameLength(hostname string) error {
	if len(hostname) == 0 || len(hostname) > 253 {
		return fmt.Errorf("length must be between 1 and 253 characters")
	}
	return nil
}

func validateHostnameEdges(hostname string) error {
	if hostname[0] == '.' || hostname[0] == '-' ||
		hostname[len(hostname)-1] == '.' || hostname[len(hostname)-1] == '-' {
		return fmt.Errorf("hostname cannot start or end with '.' or '-'")
	}
	return nil
}

func validateHostnameLabel(label string) error {
	if len(label) == 0 || len(label) > 63 {
		return fmt.Errorf("hostname labels must be 1-63 characters each")
	}
	if label[0] == '-' || label[len(label)-1] == '-' {
		return fmt.Errorf("hostname labels cannot start or end with '-'")
	}
	for _, ch := range label {
		if (ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' {
			continue
		}
		return fmt.Errorf("invalid character %q in hostname", ch)
	}
	return nil
}
