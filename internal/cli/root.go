package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version is the current version of the CLI
	// This can be overridden at build time with ldflags
	Version = "0.3.0-dev"
	// Commit is the git commit hash (injected at build time)
	Commit = "unknown"
	// Date is the build date (injected at build time)
	Date = "unknown"
)

// defaultNetworkReqsURL is the canonical Delinea network requirements URL,
// used as a fallback when no -f or -u flag is provided.
const defaultNetworkReqsURL = "https://setup.delinea.app/network-requirements.json"

var (
	// Global flags
	verbose bool
	quiet   bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "delinea-netconfig",
	Short: "Convert Delinea network requirements to various formats",
	Long: `delinea-netconfig is a CLI tool that converts Delinea's Platform IP/CIDR
network requirements JSON into various firewall and infrastructure-as-code formats.

Supported output formats:
  - CSV
  - YAML
  - Terraform (HCL)
  - Ansible
  - AWS Security Group JSON
  - Cisco ACL
  - PAN-OS XML`,
	Version: Version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose logging")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress non-error output")

	// Enable shell completion
	rootCmd.CompletionOptions.DisableDefaultCmd = false
	rootCmd.CompletionOptions.HiddenDefaultCmd = false

	// Add subcommands
	rootCmd.AddCommand(convertCmd)
	rootCmd.AddCommand(validateCmd)
}

// logInfo prints an info message unless quiet mode is enabled
func logInfo(format string, args ...interface{}) {
	if !quiet {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}

// logVerbose prints a verbose message if verbose mode is enabled
func logVerbose(format string, args ...interface{}) {
	if verbose && !quiet {
		fmt.Fprintf(os.Stderr, "[VERBOSE] "+format+"\n", args...)
	}
}

// logError prints an error message
func logError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
}
