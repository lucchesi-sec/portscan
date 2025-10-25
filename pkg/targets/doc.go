// Package targets provides target resolution and expansion functionality.
//
// This package converts user-provided target specifications into scan-ready
// host addresses. It supports multiple input formats:
//
//   - IP addresses: "192.168.1.1", "10.0.0.50"
//   - Hostnames: "example.com", "api.internal"
//   - CIDR notation: "192.168.1.0/24", "10.0.0.0/16"
//   - Mixed lists: combining any of the above
//
// Example usage:
//
//	opts := targets.Options{CIDRHostLimit: 65536}
//	inputs := []string{"192.168.1.1", "10.0.0.0/24", "example.com"}
//	hosts, err := targets.Resolve(inputs, opts)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Scanning %d hosts\n", len(hosts))
//
// CIDR Expansion:
//
// CIDR notation is automatically expanded to individual IP addresses.
// For example, "192.168.1.0/24" produces 254 usable host addresses
// (network and broadcast addresses are excluded). Large CIDR blocks
// can be limited using Options.CIDRHostLimit to prevent excessive
// memory usage.
//
// Validation:
//
// The Validate function provides comprehensive input validation:
//   - IP address format checking
//   - Hostname resolution verification
//   - CIDR notation validation
//   - Protection against private IP scanning (configurable)
//   - Localhost scanning restrictions (configurable)
//
// Deduplication:
//
// All target resolution automatically removes duplicate hosts, even if
// they're specified multiple times or overlap in CIDR ranges.
package targets
