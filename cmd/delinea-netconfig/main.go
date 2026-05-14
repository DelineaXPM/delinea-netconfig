package main

import (
	"os"

	"github.com/DelineaXPM/delinea-netconfig/internal/cli"
)

func main() {
	// Cobra renders error messages itself (and per-command SilenceUsage
	// suppresses the help dump). All we need here is the exit code so CI
	// pipelines and shell scripts can detect failure — notably from the
	// `check` subcommand, which returns cli.ErrProbesFailed when any DNS,
	// TCP, or TLS probe fails.
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
