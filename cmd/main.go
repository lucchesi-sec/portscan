package main

import (
	"os"

	"github.com/lucchesi-sec/portscan/cmd/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}