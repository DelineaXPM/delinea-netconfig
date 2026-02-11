package cli

import (
	"fmt"
	"sort"

	"github.com/DelineaXPM/delinea-platform/delinea-netconfig/internal/fetcher"
	"github.com/DelineaXPM/delinea-platform/delinea-netconfig/internal/parser"
	"github.com/DelineaXPM/delinea-platform/delinea-netconfig/pkg/types"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <file>",
	Short: "Show statistics about network requirements",
	Long: `Display statistical information about network requirements file.

Shows counts of entries by direction, service, region, protocol, and type.
Useful for understanding the scope and complexity of network requirements.

Examples:
  # Show statistics for a file
  delinea-netconfig info network-requirements.json

  # Show verbose statistics
  delinea-netconfig info -v network-requirements.json

  # Show statistics in quiet mode
  delinea-netconfig info -q network-requirements.json`,
	Args: cobra.ExactArgs(1),
	RunE: runInfo,
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

func runInfo(cmd *cobra.Command, args []string) error {
	file := args[0]

	if !quiet {
		fmt.Printf("Network Requirements Statistics\n")
		fmt.Printf("File: %s\n\n", file)
	}

	// Read and parse file
	data, err := fetcher.FetchFromFile(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	networkReqs, err := parser.Parse(data)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	entries := parser.Normalize(networkReqs)

	// Collect statistics
	stats := collectStatistics(entries)

	// Display statistics
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
