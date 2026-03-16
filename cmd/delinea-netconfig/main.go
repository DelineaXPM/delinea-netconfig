package main

import (
	"os"

	"github.com/DelineaXPM/delinea-netconfig/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
