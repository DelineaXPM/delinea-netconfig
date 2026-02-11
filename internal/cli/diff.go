package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/DelineaXPM/delinea-platform/delinea-netconfig/internal/fetcher"
	"github.com/DelineaXPM/delinea-platform/delinea-netconfig/internal/parser"
	"github.com/DelineaXPM/delinea-platform/delinea-netconfig/pkg/types"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff <file1> <file2>",
	Short: "Compare two network requirements files",
	Long: `Compare two versions of network requirements files and show differences.

Displays what entries were added, removed, or modified between versions.
Useful for tracking changes in network requirements over time.

Examples:
  # Compare two versions
  delinea-netconfig diff old-requirements.json new-requirements.json

  # Compare with quiet mode (less verbose)
  delinea-netconfig diff -q v1.json v2.json

  # Show only summary statistics
  delinea-netconfig diff --summary v1.json v2.json`,
	Args: cobra.ExactArgs(2),
	RunE: runDiff,
}

var (
	diffSummaryOnly bool
)

func init() {
	diffCmd.Flags().BoolVarP(&diffSummaryOnly, "summary", "s", false, "Show only summary statistics")
	rootCmd.AddCommand(diffCmd)
}

func runDiff(cmd *cobra.Command, args []string) error {
	file1 := args[0]
	file2 := args[1]

	// Parse both files
	if !quiet {
		fmt.Printf("Comparing:\n  Old: %s\n  New: %s\n\n", file1, file2)
	}

	// Read and parse file 1
	data1, err := fetcher.FetchFromFile(file1)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", file1, err)
	}

	networkReqs1, err := parser.Parse(data1)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", file1, err)
	}

	entries1 := parser.Normalize(networkReqs1)

	// Read and parse file 2
	data2, err := fetcher.FetchFromFile(file2)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", file2, err)
	}

	networkReqs2, err := parser.Parse(data2)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", file2, err)
	}

	entries2 := parser.Normalize(networkReqs2)

	// Compare entries
	added, removed, modified := compareEntries(entries1, entries2)

	// Display results
	if diffSummaryOnly {
		displaySummary(added, removed, modified)
	} else {
		displayDiff(added, removed, modified)
	}

	return nil
}

// entryKey generates a unique key for a network entry
func entryKey(e types.NetworkEntry) string {
	return fmt.Sprintf("%s|%s|%s|%s|%s|%s",
		e.Direction,
		e.Service,
		e.Region,
		e.Type,
		strings.Join(e.Values, ","),
		e.Protocol)
}

// compareEntries compares two sets of network entries
func compareEntries(old, new []types.NetworkEntry) (added, removed, modified []types.NetworkEntry) {
	oldMap := make(map[string]types.NetworkEntry)
	newMap := make(map[string]types.NetworkEntry)

	// Build maps for quick lookup
	for _, entry := range old {
		key := entryKey(entry)
		oldMap[key] = entry
	}

	for _, entry := range new {
		key := entryKey(entry)
		newMap[key] = entry
	}

	// Find added entries (in new but not in old)
	for key, entry := range newMap {
		if _, exists := oldMap[key]; !exists {
			added = append(added, entry)
		}
	}

	// Find removed entries (in old but not in new)
	for key, entry := range oldMap {
		if _, exists := newMap[key]; !exists {
			removed = append(removed, entry)
		}
	}

	// Find modified entries (in both but with different ports or description)
	for key, newEntry := range newMap {
		if oldEntry, exists := oldMap[key]; exists {
			if !entriesEqual(oldEntry, newEntry) {
				modified = append(modified, newEntry)
			}
		}
	}

	// Sort results for consistent output
	sortEntries(added)
	sortEntries(removed)
	sortEntries(modified)

	return added, removed, modified
}

// entriesEqual checks if two entries are equal (including ports and description)
func entriesEqual(e1, e2 types.NetworkEntry) bool {
	if e1.Description != e2.Description {
		return false
	}

	if len(e1.Ports) != len(e2.Ports) {
		return false
	}

	for i, port := range e1.Ports {
		if port != e2.Ports[i] {
			return false
		}
	}

	return true
}

// sortEntries sorts entries for consistent output
func sortEntries(entries []types.NetworkEntry) {
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].Direction != entries[j].Direction {
			return entries[i].Direction > entries[j].Direction // outbound first
		}
		if entries[i].Service != entries[j].Service {
			return entries[i].Service < entries[j].Service
		}
		if entries[i].Region != entries[j].Region {
			return entries[i].Region < entries[j].Region
		}
		return entries[i].Type < entries[j].Type
	})
}

// displaySummary displays only summary statistics
func displaySummary(added, removed, modified []types.NetworkEntry) {
	fmt.Println("Summary:")
	fmt.Printf("  Added:    %d entries\n", len(added))
	fmt.Printf("  Removed:  %d entries\n", len(removed))
	fmt.Printf("  Modified: %d entries\n", len(modified))
	fmt.Printf("  Total changes: %d\n", len(added)+len(removed)+len(modified))

	if len(added)+len(removed)+len(modified) == 0 {
		fmt.Println("\n✓ No differences found")
	}
}

// displayDiff displays detailed differences
func displayDiff(added, removed, modified []types.NetworkEntry) {
	hasChanges := false

	// Display added entries
	if len(added) > 0 {
		hasChanges = true
		fmt.Printf("Added (%d entries):\n", len(added))
		for _, entry := range added {
			fmt.Printf("  + [%s] %s/%s: %s (%s:%v)\n",
				entry.Direction,
				entry.Service,
				entry.Region,
				strings.Join(entry.Values, ", "),
				entry.Protocol,
				entry.Ports)
			if entry.Description != "" && !quiet {
				fmt.Printf("    → %s\n", entry.Description)
			}
		}
		fmt.Println()
	}

	// Display removed entries
	if len(removed) > 0 {
		hasChanges = true
		fmt.Printf("Removed (%d entries):\n", len(removed))
		for _, entry := range removed {
			fmt.Printf("  - [%s] %s/%s: %s (%s:%v)\n",
				entry.Direction,
				entry.Service,
				entry.Region,
				strings.Join(entry.Values, ", "),
				entry.Protocol,
				entry.Ports)
			if entry.Description != "" && !quiet {
				fmt.Printf("    → %s\n", entry.Description)
			}
		}
		fmt.Println()
	}

	// Display modified entries
	if len(modified) > 0 {
		hasChanges = true
		fmt.Printf("Modified (%d entries):\n", len(modified))
		for _, entry := range modified {
			fmt.Printf("  ~ [%s] %s/%s: %s (%s:%v)\n",
				entry.Direction,
				entry.Service,
				entry.Region,
				strings.Join(entry.Values, ", "),
				entry.Protocol,
				entry.Ports)
			if entry.Description != "" && !quiet {
				fmt.Printf("    → %s\n", entry.Description)
			}
		}
		fmt.Println()
	}

	// Display summary
	fmt.Println("Summary:")
	fmt.Printf("  Added:    %d entries\n", len(added))
	fmt.Printf("  Removed:  %d entries\n", len(removed))
	fmt.Printf("  Modified: %d entries\n", len(modified))
	fmt.Printf("  Total changes: %d\n", len(added)+len(removed)+len(modified))

	if !hasChanges {
		fmt.Println("\n✓ No differences found")
	}
}
