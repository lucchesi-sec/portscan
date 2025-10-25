// Package parser provides port specification parsing functionality.
//
// This package handles conversion of user-friendly port specifications into
// lists of port numbers suitable for network scanning. It supports multiple
// input formats:
//
//   - Single ports: "80", "443", "8080"
//   - Port ranges: "1-1024", "8000-9000"
//   - Comma-separated lists: "22,80,443"
//   - Mixed formats: "80,443,8000-9000"
//
// Example usage:
//
//	ports, err := parser.ParsePorts("22,80,443,8000-9000")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Scanning %d ports\n", len(ports))
//
// The parser validates all port numbers (1-65535) and automatically removes
// duplicates. Invalid specifications return descriptive errors.
//
// Port Range Expansion:
//
// Port ranges are expanded inclusively. For example, "80-83" produces
// [80, 81, 82, 83]. Large ranges are supported and efficiently processed.
//
// Validation:
//
// All port numbers must be in the valid range 1-65535. Ports outside this
// range or malformed specifications (e.g., "abc", "80-", "-443") will
// return descriptive errors indicating the problem.
package parser
