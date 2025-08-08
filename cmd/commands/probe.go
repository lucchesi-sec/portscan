package commands

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var probeCmd = &cobra.Command{
	Use:   "probe",
	Short: "Manage custom UDP probes",
	Long:  `Manage custom UDP probes for the UDP scanner.`,
}

var addProbeCmd = &cobra.Command{
	Use:   "add PORT HEX_DATA",
	Short: "Add a custom UDP probe for a specific port",
	Long: `Add a custom UDP probe for a specific port.

The HEX_DATA should be provided as a hex string without spaces or prefixes.
For example: portscan probe add 1234 000102030405`,
	Args: cobra.ExactArgs(2),
	RunE: runAddProbe,
}

var statsProbeCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show UDP probe statistics",
	Long:  `Show statistics for UDP probe effectiveness.`,
	RunE:  runProbeStats,
}

func init() {
	rootCmd.AddCommand(probeCmd)
	probeCmd.AddCommand(addProbeCmd)
	probeCmd.AddCommand(statsProbeCmd)
}

func runAddProbe(cmd *cobra.Command, args []string) error {
	portStr := args[0]
	hexData := args[1]

	// Parse port
	port64, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return fmt.Errorf("invalid port: %s", portStr)
	}
	port := uint16(port64)

	// Parse hex data
	data, err := hex.DecodeString(hexData)
	if err != nil {
		return fmt.Errorf("invalid hex data: %s", hexData)
	}

	// For now, just print what would be added
	// In a real implementation, this would need to be stored somewhere persistent
	fmt.Printf("Would add custom probe for port %d with data: %s\n", port, hexData)
	fmt.Printf("Decoded data length: %d bytes\n", len(data))

	// Example of how this would be used with a scanner
	// scanner.AddCustomProbe(port, data)

	return nil
}

func runProbeStats(cmd *cobra.Command, args []string) error {
	// This is a placeholder - in a real implementation, we would need to get
	// statistics from a running or completed scan
	fmt.Println("Probe statistics would be shown here after a scan is completed.")
	fmt.Println("This feature requires integration with the scan results.")
	return nil
}
