package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DelineaXPM/delinea-netconfig/internal/converter"
	"github.com/DelineaXPM/delinea-netconfig/internal/fetcher"
	"github.com/DelineaXPM/delinea-netconfig/internal/parser"
	"github.com/DelineaXPM/delinea-netconfig/pkg/types"
	"github.com/spf13/cobra"
)

var (
	// Convert command flags
	inputFile    string
	inputURL     string
	outputFile   string
	outputDir    string
	format       string
	tenantName   string
	regionFilter string
)

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert network requirements to different formats",
	Long: `Convert Delinea network requirements JSON to various firewall
and infrastructure-as-code formats.

Examples:
  # Convert to Terraform
  delinea-netconfig convert -f network-requirements.json --format terraform

  # Convert with tenant substitution
  delinea-netconfig convert -f network-requirements.json --format csv --tenant mycompany

  # Convert for a specific region (includes global + region-specific rules)
  delinea-netconfig convert -f network-requirements.json --format csv --region eu
  delinea-netconfig convert -u https://setup.delinea.app/network-requirements --format yaml --tenant mycompany --region au

  # Convert to multiple formats
  delinea-netconfig convert -f network-requirements.json --format csv,yaml,terraform

  # Fetch from URL and convert
  delinea-netconfig convert -u https://example.com/network-requirements.json --format csv

  # Save to file
  delinea-netconfig convert -f network-requirements.json --format terraform -o output.tf

  # Save multiple formats to directory
  delinea-netconfig convert -f network-requirements.json --format csv,yaml,terraform --output-dir ./configs`,
	RunE: runConvert,
}

func init() {
	convertCmd.Flags().StringVarP(&inputFile, "file", "f", "", "path to network-requirements.json file")
	convertCmd.Flags().StringVarP(&inputURL, "url", "u", "", "URL to fetch network-requirements.json")
	convertCmd.Flags().StringVarP(&outputFile, "output", "o", "", "output file path (default: stdout)")
	convertCmd.Flags().StringVar(&outputDir, "output-dir", "", "output directory for multiple formats")
	convertCmd.Flags().StringVar(&format, "format", "csv", "output format(s): csv, yaml, terraform, ansible, aws-sg, cisco, panos (comma-separated)")
	convertCmd.Flags().StringVarP(&tenantName, "tenant", "t", "", "substitute <tenant> placeholder with this value")
	convertCmd.Flags().StringVarP(&regionFilter, "region", "r", "", "filter to global + region-specific rules (e.g. eu, au, us)")

	convertCmd.MarkFlagsMutuallyExclusive("file", "url")
	convertCmd.MarkFlagsMutuallyExclusive("output", "output-dir")
}

func runConvert(cmd *cobra.Command, args []string) error {
	logVerbose("Starting conversion process")

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

	logVerbose("Successfully fetched %d bytes", len(data))

	// Step 2: Parse the JSON
	logVerbose("Parsing network requirements JSON")
	networkReqs, err := parser.Parse(data)
	if err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	logInfo("Parsed network requirements version %s (updated: %s)", networkReqs.Version, networkReqs.UpdatedAt)

	// Step 3: Normalize to network entries
	logVerbose("Normalizing network requirements to entries")
	entries := parser.Normalize(networkReqs)
	logInfo("Normalized %d network entries", len(entries))

	// Step 3.5: Substitute tenant placeholder if provided
	if tenantName != "" {
		logVerbose("Substituting <tenant> with: %s", tenantName)
		entries = substituteTenant(entries, tenantName)
	}

	// Step 3.6: Filter by region if provided
	if regionFilter != "" {
		region := strings.ToLower(strings.TrimSpace(regionFilter))
		if region == "global" {
			logInfo("Note: global rules are always included; --region filters to global + a specific region (e.g. --region eu)")
		} else {
			before := len(entries)
			entries = filterByRegion(entries, region)
			logInfo("Filtered to global + %s: %d of %d entries", region, len(entries), before)
		}
	}

	// Step 4: Parse formats
	formats := strings.Split(format, ",")
	for i := range formats {
		formats[i] = strings.TrimSpace(formats[i])
	}

	logVerbose("Converting to formats: %v", formats)

	// Step 5: Convert to each format
	for _, fmt := range formats {
		if err := convertToFormat(fmt, entries); err != nil {
			return err
		}
	}

	logInfo("Conversion completed successfully")
	return nil
}

func convertToFormat(formatName string, entries []types.NetworkEntry) error {
	logVerbose("Converting to format: %s", formatName)

	// Get the converter for this format
	conv, err := converter.GetConverter(formatName)
	if err != nil {
		return fmt.Errorf("unsupported format %q: %w", formatName, err)
	}

	// Convert
	output, err := conv.Convert(entries)
	if err != nil {
		return fmt.Errorf("conversion to %s failed: %w", formatName, err)
	}

	// Determine output destination
	var outputPath string
	if outputDir != "" {
		// Multiple formats to directory
		filename := "output." + conv.FileExtension()
		outputPath = filepath.Join(outputDir, filename)

		// Create output directory if it doesn't exist
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		// Write to file
		if err := os.WriteFile(outputPath, output, 0644); err != nil {
			return fmt.Errorf("failed to write to file %s: %w", outputPath, err)
		}
		logInfo("Wrote %s output to: %s", formatName, outputPath)
	} else if outputFile != "" {
		// Single format to specific file
		if len(strings.Split(format, ",")) > 1 {
			return fmt.Errorf("cannot use --output with multiple formats; use --output-dir instead")
		}

		if err := os.WriteFile(outputFile, output, 0644); err != nil {
			return fmt.Errorf("failed to write to file %s: %w", outputFile, err)
		}
		logInfo("Wrote output to: %s", outputFile)
	} else {
		// Output to stdout
		fmt.Print(string(output))
	}

	return nil
}

// filterByRegion returns entries where Region is "global" or matches the given region (case-insensitive).
func filterByRegion(entries []types.NetworkEntry, region string) []types.NetworkEntry {
	var result []types.NetworkEntry
	for _, entry := range entries {
		if entry.Region == "global" || strings.EqualFold(entry.Region, region) {
			result = append(result, entry)
		}
	}
	return result
}

// substituteTenant replaces <tenant> placeholders in network entry values (hostnames/IPs) only
func substituteTenant(entries []types.NetworkEntry, tenant string) []types.NetworkEntry {
	result := make([]types.NetworkEntry, len(entries))

	for i, entry := range entries {
		// Copy the entry
		result[i] = entry

		// Substitute in values only (hostnames/IPs)
		if len(entry.Values) > 0 {
			newValues := make([]string, len(entry.Values))
			for j, value := range entry.Values {
				newValues[j] = strings.ReplaceAll(value, "<tenant>", tenant)
			}
			result[i].Values = newValues
		}
	}

	return result
}
