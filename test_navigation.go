package main

import (
	"fmt"
	"log"

	"github.com/lucchesi-sec/portscan/internal/core"
	"github.com/lucchesi-sec/portscan/internal/ui"
	"github.com/lucchesi-sec/portscan/pkg/config"
)

func main() {
	// Create sample configuration
	cfg := &config.Config{
		UI: config.UIConfig{
			Theme: "default",
		},
	}

	// Create a channel for results
	results := make(chan core.Event, 10)

	// Add some sample data
	go func() {
		defer close(results)
		for i := 1; i <= 10; i++ {
			results <- core.NewResultEvent(core.ResultEvent{
				Host:   "127.0.0.1",
				Port:   uint16(80 + i),
				State:  core.StateOpen,
				Banner: fmt.Sprintf("Test service %d", i),
			})
		}
	}()

	// Create and run UI
	tui := ui.NewScanUI(cfg, 10, results, false)

	fmt.Println("Starting TUI test. Use arrow keys to test navigation.")
	fmt.Println("The down arrow should now move one entry at a time, not skip entries.")
	fmt.Println("Press 'q' to quit when done testing.")

	if err := tui.Run(); err != nil {
		log.Fatal(err)
	}
}
