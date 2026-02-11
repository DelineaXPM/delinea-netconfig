package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Display version, commit, and build date information for delinea-netconfig.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("delinea-netconfig version %s\n", Version)
		if Commit != "unknown" {
			fmt.Printf("Commit: %s\n", Commit)
		}
		if Date != "unknown" {
			fmt.Printf("Built: %s\n", Date)
		}
		fmt.Println("Go version: go1.23+")
		fmt.Println("Platform: Multi-platform (Linux, macOS, Windows, FreeBSD)")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
