package cli

import (
	"fmt"
	"strings"

	"github.com/DelineaXPM/delinea-platform/delinea-netconfig/internal/differ"
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

	if !quiet {
		fmt.Printf("Comparing:\n  Old: %s\n  New: %s\n\n", file1, file2)
	}

	data1, err := fetcher.FetchFromFile(file1)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", file1, err)
	}

	networkReqs1, err := parser.Parse(data1)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", file1, err)
	}

	entries1 := parser.Normalize(networkReqs1)

	data2, err := fetcher.FetchFromFile(file2)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", file2, err)
	}

	networkReqs2, err := parser.Parse(data2)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", file2, err)
	}

	entries2 := parser.Normalize(networkReqs2)

	result := differ.Compare(entries1, entries2)

	if diffSummaryOnly {
		displaySummary(result.Added, result.Removed, result.Modified)
	} else {
		displayDiff(result.Added, result.Removed, result.Modified)
	}

	return nil
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

	fmt.Println("Summary:")
	fmt.Printf("  Added:    %d entries\n", len(added))
	fmt.Printf("  Removed:  %d entries\n", len(removed))
	fmt.Printf("  Modified: %d entries\n", len(modified))
	fmt.Printf("  Total changes: %d\n", len(added)+len(removed)+len(modified))

	if !hasChanges {
		fmt.Println("\n✓ No differences found")
	}
}
