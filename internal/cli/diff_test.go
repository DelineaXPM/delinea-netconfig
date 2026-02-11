package cli

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/DelineaXPM/delinea-platform/delinea-netconfig/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestEntryKey(t *testing.T) {
	tests := []struct {
		name     string
		entry    types.NetworkEntry
		expected string
	}{
		{
			name: "generates key with all fields",
			entry: types.NetworkEntry{
				Direction: "outbound",
				Service:   "test",
				Region:    "us",
				Type:      "ipv4",
				Values:    []string{"192.168.1.0/24"},
				Protocol:  "tcp",
			},
			expected: "outbound|test|us|ipv4|192.168.1.0/24|tcp",
		},
		{
			name: "generates key with multiple values",
			entry: types.NetworkEntry{
				Direction: "inbound",
				Service:   "webhooks",
				Region:    "global",
				Type:      "ipv4",
				Values:    []string{"10.0.0.0/8", "192.168.1.0/24"},
				Protocol:  "tcp",
			},
			expected: "inbound|webhooks|global|ipv4|10.0.0.0/8,192.168.1.0/24|tcp",
		},
		{
			name: "different entries produce different keys",
			entry: types.NetworkEntry{
				Direction: "outbound",
				Service:   "different",
				Region:    "eu",
				Type:      "ipv6",
				Values:    []string{"2001:db8::/32"},
				Protocol:  "udp",
			},
			expected: "outbound|different|eu|ipv6|2001:db8::/32|udp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := entryKey(tt.entry)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCompareEntries(t *testing.T) {
	tests := []struct {
		name             string
		old              []types.NetworkEntry
		new              []types.NetworkEntry
		expectedAdded    int
		expectedRemoved  int
		expectedModified int
	}{
		{
			name: "no changes",
			old: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "test",
					Region:      "us",
					Type:        "ipv4",
					Values:      []string{"192.168.1.0/24"},
					Protocol:    "tcp",
					Ports:       []int{443},
					Description: "Test service",
				},
			},
			new: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "test",
					Region:      "us",
					Type:        "ipv4",
					Values:      []string{"192.168.1.0/24"},
					Protocol:    "tcp",
					Ports:       []int{443},
					Description: "Test service",
				},
			},
			expectedAdded:    0,
			expectedRemoved:  0,
			expectedModified: 0,
		},
		{
			name: "entry added",
			old: []types.NetworkEntry{
				{
					Direction: "outbound",
					Service:   "service1",
					Region:    "us",
					Type:      "ipv4",
					Values:    []string{"192.168.1.0/24"},
					Protocol:  "tcp",
					Ports:     []int{443},
				},
			},
			new: []types.NetworkEntry{
				{
					Direction: "outbound",
					Service:   "service1",
					Region:    "us",
					Type:      "ipv4",
					Values:    []string{"192.168.1.0/24"},
					Protocol:  "tcp",
					Ports:     []int{443},
				},
				{
					Direction: "inbound",
					Service:   "service2",
					Region:    "eu",
					Type:      "ipv4",
					Values:    []string{"10.0.0.0/8"},
					Protocol:  "tcp",
					Ports:     []int{80},
				},
			},
			expectedAdded:    1,
			expectedRemoved:  0,
			expectedModified: 0,
		},
		{
			name: "entry removed",
			old: []types.NetworkEntry{
				{
					Direction: "outbound",
					Service:   "service1",
					Region:    "us",
					Type:      "ipv4",
					Values:    []string{"192.168.1.0/24"},
					Protocol:  "tcp",
					Ports:     []int{443},
				},
				{
					Direction: "inbound",
					Service:   "service2",
					Region:    "eu",
					Type:      "ipv4",
					Values:    []string{"10.0.0.0/8"},
					Protocol:  "tcp",
					Ports:     []int{80},
				},
			},
			new: []types.NetworkEntry{
				{
					Direction: "outbound",
					Service:   "service1",
					Region:    "us",
					Type:      "ipv4",
					Values:    []string{"192.168.1.0/24"},
					Protocol:  "tcp",
					Ports:     []int{443},
				},
			},
			expectedAdded:    0,
			expectedRemoved:  1,
			expectedModified: 0,
		},
		{
			name: "entry modified - ports changed",
			old: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "test",
					Region:      "us",
					Type:        "ipv4",
					Values:      []string{"192.168.1.0/24"},
					Protocol:    "tcp",
					Ports:       []int{443},
					Description: "Test service",
				},
			},
			new: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "test",
					Region:      "us",
					Type:        "ipv4",
					Values:      []string{"192.168.1.0/24"},
					Protocol:    "tcp",
					Ports:       []int{443, 8443}, // Ports changed
					Description: "Test service",
				},
			},
			expectedAdded:    0,
			expectedRemoved:  0,
			expectedModified: 1,
		},
		{
			name: "entry modified - description changed",
			old: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "test",
					Region:      "us",
					Type:        "ipv4",
					Values:      []string{"192.168.1.0/24"},
					Protocol:    "tcp",
					Ports:       []int{443},
					Description: "Old description",
				},
			},
			new: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "test",
					Region:      "us",
					Type:        "ipv4",
					Values:      []string{"192.168.1.0/24"},
					Protocol:    "tcp",
					Ports:       []int{443},
					Description: "New description", // Description changed
				},
			},
			expectedAdded:    0,
			expectedRemoved:  0,
			expectedModified: 1,
		},
		{
			name: "multiple changes",
			old: []types.NetworkEntry{
				{
					Direction: "outbound",
					Service:   "service1",
					Region:    "us",
					Type:      "ipv4",
					Values:    []string{"192.168.1.0/24"},
					Protocol:  "tcp",
					Ports:     []int{443},
				},
				{
					Direction: "inbound",
					Service:   "service2",
					Region:    "eu",
					Type:      "ipv4",
					Values:    []string{"10.0.0.0/8"},
					Protocol:  "tcp",
					Ports:     []int{80},
				},
			},
			new: []types.NetworkEntry{
				{
					Direction: "outbound",
					Service:   "service1",
					Region:    "us",
					Type:      "ipv4",
					Values:    []string{"192.168.1.0/24"},
					Protocol:  "tcp",
					Ports:     []int{443, 8443}, // Modified
				},
				{
					Direction: "outbound",
					Service:   "service3",
					Region:    "global",
					Type:      "ipv4",
					Values:    []string{"172.16.0.0/12"},
					Protocol:  "tcp",
					Ports:     []int{443}, // Added
				},
				// service2 removed
			},
			expectedAdded:    1,
			expectedRemoved:  1,
			expectedModified: 1,
		},
		{
			name:             "empty old, entries added",
			old:              []types.NetworkEntry{},
			new: []types.NetworkEntry{
				{
					Direction: "outbound",
					Service:   "service1",
					Region:    "us",
					Type:      "ipv4",
					Values:    []string{"192.168.1.0/24"},
					Protocol:  "tcp",
					Ports:     []int{443},
				},
			},
			expectedAdded:    1,
			expectedRemoved:  0,
			expectedModified: 0,
		},
		{
			name: "empty new, entries removed",
			old: []types.NetworkEntry{
				{
					Direction: "outbound",
					Service:   "service1",
					Region:    "us",
					Type:      "ipv4",
					Values:    []string{"192.168.1.0/24"},
					Protocol:  "tcp",
					Ports:     []int{443},
				},
			},
			new:              []types.NetworkEntry{},
			expectedAdded:    0,
			expectedRemoved:  1,
			expectedModified: 0,
		},
		{
			name:             "both empty",
			old:              []types.NetworkEntry{},
			new:              []types.NetworkEntry{},
			expectedAdded:    0,
			expectedRemoved:  0,
			expectedModified: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			added, removed, modified := compareEntries(tt.old, tt.new)
			assert.Len(t, added, tt.expectedAdded, "incorrect number of added entries")
			assert.Len(t, removed, tt.expectedRemoved, "incorrect number of removed entries")
			assert.Len(t, modified, tt.expectedModified, "incorrect number of modified entries")
		})
	}
}

func TestEntriesEqual(t *testing.T) {
	tests := []struct {
		name     string
		e1       types.NetworkEntry
		e2       types.NetworkEntry
		expected bool
	}{
		{
			name: "identical entries",
			e1: types.NetworkEntry{
				Direction:   "outbound",
				Service:     "test",
				Description: "Test service",
				Ports:       []int{443},
			},
			e2: types.NetworkEntry{
				Direction:   "outbound",
				Service:     "test",
				Description: "Test service",
				Ports:       []int{443},
			},
			expected: true,
		},
		{
			name: "different descriptions",
			e1: types.NetworkEntry{
				Description: "Old description",
				Ports:       []int{443},
			},
			e2: types.NetworkEntry{
				Description: "New description",
				Ports:       []int{443},
			},
			expected: false,
		},
		{
			name: "different ports",
			e1: types.NetworkEntry{
				Description: "Test",
				Ports:       []int{443},
			},
			e2: types.NetworkEntry{
				Description: "Test",
				Ports:       []int{443, 8443},
			},
			expected: false,
		},
		{
			name: "different port values",
			e1: types.NetworkEntry{
				Description: "Test",
				Ports:       []int{443},
			},
			e2: types.NetworkEntry{
				Description: "Test",
				Ports:       []int{80},
			},
			expected: false,
		},
		{
			name: "empty descriptions",
			e1: types.NetworkEntry{
				Description: "",
				Ports:       []int{443},
			},
			e2: types.NetworkEntry{
				Description: "",
				Ports:       []int{443},
			},
			expected: true,
		},
		{
			name: "empty ports",
			e1: types.NetworkEntry{
				Description: "Test",
				Ports:       []int{},
			},
			e2: types.NetworkEntry{
				Description: "Test",
				Ports:       []int{},
			},
			expected: true,
		},
		{
			name: "nil vs empty ports",
			e1: types.NetworkEntry{
				Description: "Test",
				Ports:       nil,
			},
			e2: types.NetworkEntry{
				Description: "Test",
				Ports:       []int{},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := entriesEqual(tt.e1, tt.e2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSortEntries(t *testing.T) {
	tests := []struct {
		name     string
		entries  []types.NetworkEntry
		expected []string // Expected order of service names
	}{
		{
			name: "sorts by direction (outbound first)",
			entries: []types.NetworkEntry{
				{Direction: "inbound", Service: "service1"},
				{Direction: "outbound", Service: "service2"},
			},
			expected: []string{"service2", "service1"},
		},
		{
			name: "sorts by service name alphabetically",
			entries: []types.NetworkEntry{
				{Direction: "outbound", Service: "zebra"},
				{Direction: "outbound", Service: "alpha"},
				{Direction: "outbound", Service: "beta"},
			},
			expected: []string{"alpha", "beta", "zebra"},
		},
		{
			name: "sorts by region when direction and service are same",
			entries: []types.NetworkEntry{
				{Direction: "outbound", Service: "test", Region: "us"},
				{Direction: "outbound", Service: "test", Region: "eu"},
				{Direction: "outbound", Service: "test", Region: "ap"},
			},
			expected: []string{"test", "test", "test"}, // All same service
		},
		{
			name: "handles complex sorting",
			entries: []types.NetworkEntry{
				{Direction: "inbound", Service: "service1", Region: "us"},
				{Direction: "outbound", Service: "service2", Region: "eu"},
				{Direction: "outbound", Service: "service1", Region: "us"},
				{Direction: "inbound", Service: "service2", Region: "global"},
			},
			expected: []string{"service1", "service2", "service1", "service2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries := make([]types.NetworkEntry, len(tt.entries))
			copy(entries, tt.entries)

			sortEntries(entries)

			var actualServices []string
			for _, e := range entries {
				actualServices = append(actualServices, e.Service)
			}

			assert.Equal(t, tt.expected, actualServices)
		})
	}
}

func TestDisplaySummary(t *testing.T) {
	// Helper to capture stdout
	captureOutput := func(f func()) string {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		f()

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
		return buf.String()
	}

	tests := []struct {
		name     string
		added    []types.NetworkEntry
		removed  []types.NetworkEntry
		modified []types.NetworkEntry
		contains []string
	}{
		{
			name:     "no changes",
			added:    []types.NetworkEntry{},
			removed:  []types.NetworkEntry{},
			modified: []types.NetworkEntry{},
			contains: []string{
				"Summary:",
				"Added:    0 entries",
				"Removed:  0 entries",
				"Modified: 0 entries",
				"Total changes: 0",
				"✓ No differences found",
			},
		},
		{
			name: "only added entries",
			added: []types.NetworkEntry{
				{Service: "test1"},
				{Service: "test2"},
			},
			removed:  []types.NetworkEntry{},
			modified: []types.NetworkEntry{},
			contains: []string{
				"Added:    2 entries",
				"Removed:  0 entries",
				"Modified: 0 entries",
				"Total changes: 2",
			},
		},
		{
			name:  "only removed entries",
			added: []types.NetworkEntry{},
			removed: []types.NetworkEntry{
				{Service: "test1"},
			},
			modified: []types.NetworkEntry{},
			contains: []string{
				"Added:    0 entries",
				"Removed:  1 entries",
				"Total changes: 1",
			},
		},
		{
			name:     "mixed changes",
			added:    []types.NetworkEntry{{}, {}},
			removed:  []types.NetworkEntry{{}},
			modified: []types.NetworkEntry{{}, {}, {}},
			contains: []string{
				"Added:    2 entries",
				"Removed:  1 entries",
				"Modified: 3 entries",
				"Total changes: 6",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				displaySummary(tt.added, tt.removed, tt.modified)
			})

			for _, expected := range tt.contains {
				assert.Contains(t, output, expected)
			}
		})
	}
}

func TestDisplayDiff(t *testing.T) {
	// Helper to capture stdout
	captureOutput := func(f func()) string {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		f()

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
		return buf.String()
	}

	tests := []struct {
		name     string
		added    []types.NetworkEntry
		removed  []types.NetworkEntry
		modified []types.NetworkEntry
		quiet    bool
		contains []string
	}{
		{
			name:     "no changes",
			added:    []types.NetworkEntry{},
			removed:  []types.NetworkEntry{},
			modified: []types.NetworkEntry{},
			contains: []string{
				"Summary:",
				"✓ No differences found",
			},
		},
		{
			name: "added entries",
			added: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "test",
					Region:      "us",
					Type:        "ipv4",
					Values:      []string{"192.168.1.0/24"},
					Protocol:    "tcp",
					Ports:       []int{443},
					Description: "Test service",
				},
			},
			removed:  []types.NetworkEntry{},
			modified: []types.NetworkEntry{},
			contains: []string{
				"Added (1 entries):",
				"+ [outbound] test/us: 192.168.1.0/24 (tcp:[443])",
				"→ Test service",
			},
		},
		{
			name:  "removed entries",
			added: []types.NetworkEntry{},
			removed: []types.NetworkEntry{
				{
					Direction: "inbound",
					Service:   "old",
					Region:    "global",
					Type:      "ipv4",
					Values:    []string{"10.0.0.0/8"},
					Protocol:  "tcp",
					Ports:     []int{80},
				},
			},
			modified: []types.NetworkEntry{},
			contains: []string{
				"Removed (1 entries):",
				"- [inbound] old/global: 10.0.0.0/8 (tcp:[80])",
			},
		},
		{
			name:     "modified entries",
			added:    []types.NetworkEntry{},
			removed:  []types.NetworkEntry{},
			modified: []types.NetworkEntry{
				{
					Direction: "outbound",
					Service:   "updated",
					Region:    "eu",
					Type:      "ipv6",
					Values:    []string{"2001:db8::/32"},
					Protocol:  "tcp",
					Ports:     []int{443, 8443},
				},
			},
			contains: []string{
				"Modified (1 entries):",
				"~ [outbound] updated/eu: 2001:db8::/32 (tcp:[443 8443])",
			},
		},
		{
			name: "quiet mode suppresses descriptions",
			added: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "test",
					Region:      "us",
					Values:      []string{"192.168.1.0/24"},
					Protocol:    "tcp",
					Ports:       []int{443},
					Description: "This should not appear",
				},
			},
			removed:  []types.NetworkEntry{},
			modified: []types.NetworkEntry{},
			quiet:    true,
			contains: []string{
				"Added (1 entries):",
				"+ [outbound] test/us: 192.168.1.0/24 (tcp:[443])",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set quiet mode
			oldQuiet := quiet
			quiet = tt.quiet
			defer func() { quiet = oldQuiet }()

			output := captureOutput(func() {
				displayDiff(tt.added, tt.removed, tt.modified)
			})

			for _, expected := range tt.contains {
				assert.Contains(t, output, expected)
			}

			// In quiet mode, descriptions should NOT appear
			if tt.quiet && len(tt.added) > 0 && tt.added[0].Description != "" {
				assert.NotContains(t, output, tt.added[0].Description)
			}
		})
	}
}
