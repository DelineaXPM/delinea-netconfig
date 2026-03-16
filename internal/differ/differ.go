package differ

import (
	"fmt"
	"sort"
	"strings"

	"github.com/DelineaXPM/delinea-netconfig/pkg/types"
)

// DiffResult holds the three categories of changes between two network requirements sets.
type DiffResult struct {
	Added    []types.NetworkEntry
	Removed  []types.NetworkEntry
	Modified []types.NetworkEntry
}

// Compare compares two slices of NetworkEntry and returns a DiffResult.
func Compare(old, new []types.NetworkEntry) DiffResult {
	oldMap := make(map[string]types.NetworkEntry)
	newMap := make(map[string]types.NetworkEntry)

	for _, entry := range old {
		oldMap[EntryKey(entry)] = entry
	}
	for _, entry := range new {
		newMap[EntryKey(entry)] = entry
	}

	var result DiffResult

	// Find added entries (in new but not in old)
	for key, entry := range newMap {
		if _, exists := oldMap[key]; !exists {
			result.Added = append(result.Added, entry)
		}
	}

	// Find removed entries (in old but not in new)
	for key, entry := range oldMap {
		if _, exists := newMap[key]; !exists {
			result.Removed = append(result.Removed, entry)
		}
	}

	// Find modified entries (in both but with different ports or description)
	for key, newEntry := range newMap {
		if oldEntry, exists := oldMap[key]; exists {
			if !EntriesEqual(oldEntry, newEntry) {
				result.Modified = append(result.Modified, newEntry)
			}
		}
	}

	SortEntries(result.Added)
	SortEntries(result.Removed)
	SortEntries(result.Modified)

	return result
}

// EntryKey generates a unique key for a network entry.
func EntryKey(e types.NetworkEntry) string {
	return fmt.Sprintf("%s|%s|%s|%s|%s|%s",
		e.Direction,
		e.Service,
		e.Region,
		e.Type,
		strings.Join(e.Values, ","),
		e.Protocol)
}

// EntriesEqual checks if two entries are equal by description and ports.
func EntriesEqual(e1, e2 types.NetworkEntry) bool {
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

// SortEntries sorts entries: outbound first, then service, region, type alphabetically.
func SortEntries(entries []types.NetworkEntry) {
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
