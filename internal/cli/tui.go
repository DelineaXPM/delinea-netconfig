package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/DelineaXPM/delinea-platform/delinea-netconfig/internal/tui"
)

var (
	tuiFile     string
	tuiURL      string
	tuiDiffMode bool
)

var tuiCmd = &cobra.Command{
	Use:   "tui [--diff file1 file2]",
	Short: "Interactive terminal UI for browsing network requirements",
	Long: `Launch an interactive terminal UI to browse, filter, inspect, and export
network requirements — without needing to remember CLI flags.

Features:
  • Browse all outbound/inbound entries with live filtering (/)
  • Tab toggle between outbound and inbound entries
  • Filter by region with r (global, us, eu, ap …)
  • Inspect any entry in full detail with scrollable viewport
  • Export to any supported format via interactive form
  • Copy IP addresses to clipboard
  • Compare two versions in interactive diff view (--diff)

Examples:
  # Launch browser (file picker if no -f given)
  delinea-netconfig tui

  # Load a specific file directly
  delinea-netconfig tui -f network-requirements.json

  # Load from a remote URL
  delinea-netconfig tui -u https://setup.delinea.app/network-requirements.json

  # Compare two versions interactively
  delinea-netconfig tui --diff old.json new.json`,
	Args: cobra.MaximumNArgs(2),
	RunE: runTUI,
}

func init() {
	tuiCmd.Flags().StringVarP(&tuiFile, "file", "f", "", "Path to network-requirements.json")
	tuiCmd.Flags().StringVarP(&tuiURL, "url", "u", "", "URL of network-requirements.json")
	tuiCmd.Flags().BoolVar(&tuiDiffMode, "diff", false, "Diff mode: compare two files (provide as positional args)")
	rootCmd.AddCommand(tuiCmd)
}

func runTUI(cmd *cobra.Command, args []string) error {
	cfg := tui.Config{FilePath: tuiFile, URL: tuiURL, DiffMode: tuiDiffMode}

	if tuiDiffMode {
		if len(args) != 2 {
			return fmt.Errorf("--diff requires exactly two file arguments")
		}
		cfg.File1, cfg.File2 = args[0], args[1]
	}

	p := tea.NewProgram(tui.New(cfg), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
