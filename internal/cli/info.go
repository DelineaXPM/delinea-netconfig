package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/DelineaXPM/delinea-platform/delinea-netconfig/internal/fetcher"
	"github.com/DelineaXPM/delinea-platform/delinea-netconfig/internal/parser"
	"github.com/DelineaXPM/delinea-platform/delinea-netconfig/pkg/types"
	"github.com/spf13/cobra"
)

const (
	networkReqsPath      = "/network-requirements"
	networkChangelogPath = "/network-requirements-changelog"
)

// defaultBaseURL is the default base URL for fetching network requirements.
// It can be overridden in tests.
var defaultBaseURL = "https://setup.delinea.app"

var (
	infoUpdates bool
	infoLatest  bool
	infoRaw     bool
	infoTenant  string
)

var infoCmd = &cobra.Command{
	Use:   "info [file]",
	Short: "Show statistics about network requirements",
	Long: `Display statistical information about network requirements.

Shows counts of entries by direction, service, region, protocol, and type.
Useful for understanding the scope and complexity of network requirements.

Use --updates to view the network requirements changelog from Delinea.
Use --latest to check the latest published version of network requirements.
Use --raw to print the raw JSON (pretty-printed), equivalent to curl -s URL | jq .

Examples:
  # Show statistics for a file
  delinea-netconfig info network-requirements.json

  # Show verbose statistics
  delinea-netconfig info -v network-requirements.json

  # View the network requirements changelog
  delinea-netconfig info --updates

  # View changelog from a specific tenant
  delinea-netconfig info --updates --tenant mycompany

  # Check the latest published version
  delinea-netconfig info --latest

  # Check latest version from a specific tenant
  delinea-netconfig info --latest --tenant mycompany

  # Print raw JSON (pretty-printed) from default URL
  delinea-netconfig info --raw

  # Print raw JSON from a local file
  delinea-netconfig info --raw network-requirements.json`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInfo,
}

func init() {
	rootCmd.AddCommand(infoCmd)
	infoCmd.Flags().BoolVar(&infoUpdates, "updates", false, "show the network requirements changelog from Delinea")
	infoCmd.Flags().BoolVar(&infoLatest, "latest", false, "show the latest published version of network requirements")
	infoCmd.Flags().BoolVar(&infoRaw, "raw", false, "print the raw JSON pretty-printed (like curl | jq)")
	infoCmd.Flags().StringVar(&infoTenant, "tenant", "", "tenant name for URL construction (default: setup)")
}

func runInfo(cmd *cobra.Command, args []string) error {
	// Handle --updates flag
	if infoUpdates {
		return runInfoUpdates()
	}

	// Handle --latest flag
	if infoLatest {
		return runInfoLatest()
	}

	// Handle --raw flag
	if infoRaw {
		if len(args) > 0 {
			return runInfoRawFromFile(args[0])
		}
		url := buildBaseURL() + networkReqsPath
		logInfo("No file specified, using default: %s", url)
		return runInfoRawFromURL(url)
	}

	// Default: show statistics — file arg or fall back to default URL
	if len(args) == 0 {
		url := buildBaseURL() + networkReqsPath
		logInfo("No file specified, using default: %s", url)
		return runInfoStatsFromURL(url)
	}

	return runInfoStats(args[0])
}

// buildBaseURL constructs the base URL based on the tenant flag
func buildBaseURL() string {
	if infoTenant == "" {
		return defaultBaseURL
	}
	return fmt.Sprintf("https://%s.delinea.app", infoTenant)
}

// runInfoRawFromFile reads a local file and prints it as pretty-printed JSON.
func runInfoRawFromFile(file string) error {
	data, err := fetcher.FetchFromFile(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	return printPrettyJSON(data)
}

// runInfoRawFromURL fetches a URL and prints it as pretty-printed JSON.
func runInfoRawFromURL(url string) error {
	data, err := fetcher.FetchFromURL(url)
	if err != nil {
		return fmt.Errorf("failed to fetch from URL: %w", err)
	}
	return printPrettyJSON(data)
}

// printPrettyJSON pretty-prints raw JSON bytes to stdout.
func printPrettyJSON(data []byte) error {
	var buf bytes.Buffer
	if err := json.Indent(&buf, data, "", "  "); err != nil {
		return fmt.Errorf("failed to format JSON: %w", err)
	}
	fmt.Println(buf.String())
	return nil
}

// runInfoUpdates fetches and displays the network requirements changelog
func runInfoUpdates() error {
	url := buildBaseURL() + networkChangelogPath

	logInfo("Fetching network requirements changelog from %s", url)

	data, err := fetcher.FetchFromURL(url)
	if err != nil {
		return fmt.Errorf("failed to fetch changelog: %w", err)
	}

	fmt.Print(string(data))
	return nil
}

// runInfoLatest fetches the latest network requirements and shows version info
func runInfoLatest() error {
	url := buildBaseURL() + networkReqsPath

	logInfo("Fetching latest network requirements from %s", url)

	data, err := fetcher.FetchFromURL(url)
	if err != nil {
		return fmt.Errorf("failed to fetch network requirements: %w", err)
	}

	networkReqs, err := parser.Parse(data)
	if err != nil {
		return fmt.Errorf("failed to parse network requirements: %w", err)
	}

	entries := parser.Normalize(networkReqs)

	fmt.Println("Latest Network Requirements")
	fmt.Printf("  Version:     %s\n", networkReqs.Version)
	if networkReqs.UpdatedAt != "" {
		fmt.Printf("  Updated:     %s\n", networkReqs.UpdatedAt)
	}
	if networkReqs.Description != "" {
		fmt.Printf("  Description: %s\n", networkReqs.Description)
	}
	fmt.Printf("  Source:      %s\n", url)
	fmt.Printf("  Entries:     %d\n", len(entries))

	// Count unique services
	services := make(map[string]bool)
	regions := make(map[string]bool)
	for _, e := range entries {
		services[e.Service] = true
		regions[e.Region] = true
	}
	fmt.Printf("  Services:    %d\n", len(services))
	fmt.Printf("  Regions:     %d\n", len(regions))

	// List region codes if available
	if len(networkReqs.RegionCodes) > 0 {
		fmt.Println("\nRegion Codes:")
		codes := sortedKeys(networkReqs.RegionCodes)
		for _, code := range codes {
			fmt.Printf("  %-6s %s\n", code+":", networkReqs.RegionCodes[code])
		}
	}

	return nil
}

// runInfoStats shows statistics for a local file.
func runInfoStats(file string) error {
	data, err := fetcher.FetchFromFile(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	return runInfoStatsFromData(file, data)
}

// runInfoStatsFromURL fetches from a URL and shows statistics.
func runInfoStatsFromURL(url string) error {
	data, err := fetcher.FetchFromURL(url)
	if err != nil {
		return fmt.Errorf("failed to fetch from URL: %w", err)
	}
	return runInfoStatsFromData(url, data)
}

// runInfoStatsFromData parses and displays statistics for already-fetched data.
func runInfoStatsFromData(source string, data []byte) error {
	if !quiet {
		fmt.Printf("Network Requirements Statistics\n")
		fmt.Printf("Source: %s\n\n", source)
	}

	networkReqs, err := parser.Parse(data)
	if err != nil {
		return fmt.Errorf("failed to parse: %w", err)
	}

	if !quiet {
		if networkReqs.Version != "" {
			fmt.Printf("Version: %s\n", networkReqs.Version)
		}
		if networkReqs.UpdatedAt != "" {
			fmt.Printf("Updated: %s\n", networkReqs.UpdatedAt)
		}
		if strings.TrimSpace(networkReqs.Version+networkReqs.UpdatedAt) != "" {
			fmt.Println()
		}
	}

	entries := parser.Normalize(networkReqs)
	stats := collectStatistics(entries)
	displayStatistics(stats)

	return nil
}

// Statistics holds collected statistics
type Statistics struct {
	TotalEntries     int
	ByDirection      map[string]int
	ByService        map[string]int
	ByRegion         map[string]int
	ByProtocol       map[string]int
	ByType           map[string]int
	TotalValues      int
	UniqueValues     int
	PortsUsed        map[int]int
	ServicesPerDir   map[string]map[string]int // direction -> service -> count
	HostnameCount    int
	IPv4Count        int
	IPv6Count        int
}

// collectStatistics collects statistics from network entries
func collectStatistics(entries []types.NetworkEntry) Statistics {
	stats := Statistics{
		ByDirection:    make(map[string]int),
		ByService:      make(map[string]int),
		ByRegion:       make(map[string]int),
		ByProtocol:     make(map[string]int),
		ByType:         make(map[string]int),
		PortsUsed:      make(map[int]int),
		ServicesPerDir: make(map[string]map[string]int),
	}

	uniqueValues := make(map[string]bool)

	for _, entry := range entries {
		stats.TotalEntries++

		// Count by direction
		stats.ByDirection[entry.Direction]++

		// Count by service
		stats.ByService[entry.Service]++

		// Count by region
		stats.ByRegion[entry.Region]++

		// Count by protocol
		stats.ByProtocol[entry.Protocol]++

		// Count by type
		stats.ByType[entry.Type]++

		// Count values
		for _, value := range entry.Values {
			stats.TotalValues++
			uniqueValues[value] = true
		}

		// Count ports
		for _, port := range entry.Ports {
			stats.PortsUsed[port]++
		}

		// Count services per direction
		if stats.ServicesPerDir[entry.Direction] == nil {
			stats.ServicesPerDir[entry.Direction] = make(map[string]int)
		}
		stats.ServicesPerDir[entry.Direction][entry.Service]++

		// Count by address type
		switch entry.Type {
		case "hostname", "hostname_self_signed", "hostname_ca_signed":
			stats.HostnameCount++
		case "ipv4":
			stats.IPv4Count++
		case "ipv6":
			stats.IPv6Count++
		}
	}

	stats.UniqueValues = len(uniqueValues)

	return stats
}

// displayStatistics displays collected statistics
func displayStatistics(stats Statistics) {
	// Overview
	fmt.Println("Overview:")
	fmt.Printf("  Total Entries:    %d\n", stats.TotalEntries)
	fmt.Printf("  Total Values:     %d\n", stats.TotalValues)
	fmt.Printf("  Unique Values:    %d\n", stats.UniqueValues)
	fmt.Println()

	// By Direction
	fmt.Println("By Direction:")
	displayMapSorted(stats.ByDirection, "  ")
	fmt.Println()

	// By Service
	fmt.Println("By Service:")
	displayMapSorted(stats.ByService, "  ")
	fmt.Println()

	// By Region
	fmt.Println("By Region:")
	displayMapSorted(stats.ByRegion, "  ")
	fmt.Println()

	// By Protocol
	fmt.Println("By Protocol:")
	displayMapSorted(stats.ByProtocol, "  ")
	fmt.Println()

	// By Type
	fmt.Println("By Type:")
	fmt.Printf("  Hostnames:        %d\n", stats.HostnameCount)
	fmt.Printf("  IPv4 Addresses:   %d\n", stats.IPv4Count)
	fmt.Printf("  IPv6 Addresses:   %d\n", stats.IPv6Count)
	fmt.Println()

	// Services per Direction (verbose mode only)
	if verbose {
		fmt.Println("Services per Direction:")
		for _, direction := range sortedKeys(stats.ServicesPerDir) {
			fmt.Printf("  %s:\n", direction)
			displayMapSorted(stats.ServicesPerDir[direction], "    ")
		}
		fmt.Println()
	}

	// Most Common Ports
	if len(stats.PortsUsed) > 0 {
		fmt.Println("Ports Used:")
		displayPortsSorted(stats.PortsUsed, "  ")
		fmt.Println()
	}
}

// displayMapSorted displays a map sorted by key
func displayMapSorted(m map[string]int, indent string) {
	keys := sortedKeys(m)
	for _, key := range keys {
		fmt.Printf("%s%-18s %d\n", indent, key+":", m[key])
	}
}

// displayPortsSorted displays ports sorted by count (descending)
func displayPortsSorted(ports map[int]int, indent string) {
	type portCount struct {
		port  int
		count int
	}

	var portCounts []portCount
	for port, count := range ports {
		portCounts = append(portCounts, portCount{port, count})
	}

	// Sort by count (descending), then by port number
	sort.Slice(portCounts, func(i, j int) bool {
		if portCounts[i].count != portCounts[j].count {
			return portCounts[i].count > portCounts[j].count
		}
		return portCounts[i].port < portCounts[j].port
	})

	// Display top 10 or all if fewer
	limit := len(portCounts)
	if limit > 10 {
		limit = 10
	}

	for i := 0; i < limit; i++ {
		pc := portCounts[i]
		fmt.Printf("%s%-6d (used %d times)\n", indent, pc.port, pc.count)
	}

	if len(portCounts) > 10 && !verbose {
		fmt.Printf("%s... (%d more ports, use -v to see all)\n", indent, len(portCounts)-10)
	} else if len(portCounts) > 10 && verbose {
		for i := 10; i < len(portCounts); i++ {
			pc := portCounts[i]
			fmt.Printf("%s%-6d (used %d times)\n", indent, pc.port, pc.count)
		}
	}
}

// sortedKeys returns sorted keys from a map
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
