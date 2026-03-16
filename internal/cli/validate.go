package cli

import (
	"fmt"

	"github.com/DelineaXPM/delinea-netconfig/internal/fetcher"
	"github.com/DelineaXPM/delinea-netconfig/internal/validator"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate network requirements JSON structure",
	Long: `Validate the structure and content of a network requirements JSON file.

This command checks:
  - Valid JSON syntax
  - Required fields are present
  - Schema version is supported
  - IP addresses and CIDR ranges are valid
  - Port numbers are valid

Examples:
  # Validate local file
  delinea-netconfig validate -f network-requirements.json

  # Validate from URL
  delinea-netconfig validate -u https://example.com/network-requirements.json`,
	RunE: runValidate,
}

func init() {
	validateCmd.Flags().StringVarP(&inputFile, "file", "f", "", "path to network-requirements.json file")
	validateCmd.Flags().StringVarP(&inputURL, "url", "u", "", "URL to fetch network-requirements.json")

	validateCmd.MarkFlagsMutuallyExclusive("file", "url")
}

func runValidate(cmd *cobra.Command, args []string) error {
	logVerbose("Starting validation")

	// Step 1: Fetch the JSON data
	var data []byte
	var err error

	if inputFile != "" {
		logVerbose("Reading from file: %s", inputFile)
		data, err = fetcher.FetchFromFile(inputFile)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
	} else {
		url := inputURL
		if url == "" {
			url = defaultNetworkReqsURL
			logInfo("No file or URL specified, using default: %s", url)
		}
		logVerbose("Fetching from URL: %s", url)
		data, err = fetcher.FetchFromURL(url)
		if err != nil {
			return fmt.Errorf("failed to fetch from URL: %w", err)
		}
	}

	// Step 2: Validate the JSON
	result, err := validator.Validate(data)
	if err != nil {
		logError("Validation failed: %v", err)
		return err
	}

	// Step 3: Print validation results
	fmt.Println("✓ Valid JSON structure")
	fmt.Printf("✓ Schema version: %s\n", result.Version)
	fmt.Println("✓ All required fields present")
	fmt.Printf("✓ %d IPv4 ranges validated\n", result.IPv4Count)
	fmt.Printf("✓ %d IPv6 ranges validated\n", result.IPv6Count)
	fmt.Printf("✓ %d hostnames validated\n", result.HostnameCount)
	fmt.Printf("✓ %d services validated (%d outbound, %d inbound)\n",
		result.TotalServices, result.OutboundServices, result.InboundServices)
	fmt.Printf("✓ %d regions found\n", result.RegionCount)

	if len(result.Warnings) > 0 {
		fmt.Println("\nWarnings:")
		for _, warning := range result.Warnings {
			fmt.Printf("  ⚠ %s\n", warning)
		}
	}

	logInfo("Validation completed successfully")
	return nil
}
