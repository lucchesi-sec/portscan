package commands

import (
	"fmt"

	"github.com/lucchesi-sec/portscan/internal/core"
	"github.com/lucchesi-sec/portscan/pkg/config"
)

// ScannerFactory creates scanner instances based on protocol type.
type ScannerFactory struct {
	config *core.Config
}

// NewScannerFactory creates a new scanner factory with the given configuration.
func NewScannerFactory(cfg *config.Config) *ScannerFactory {
	return &ScannerFactory{
		config: buildScannerConfig(cfg),
	}
}

// CreateScanner creates a scanner instance for the specified protocol.
// Supported protocols: "tcp", "udp"
func (f *ScannerFactory) CreateScanner(protocol string) (core.PortScanner, error) {
	switch protocol {
	case "tcp":
		return core.NewScanner(f.config), nil
	case "udp":
		return core.NewUDPScanner(f.config), nil
	default:
		return nil, fmt.Errorf("unsupported protocol: %s (supported: tcp, udp)", protocol)
	}
}
