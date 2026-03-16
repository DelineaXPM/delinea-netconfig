package differ

import (
	"testing"

	"github.com/DelineaXPM/delinea-netconfig/pkg/types"
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
			result := EntryKey(tt.entry)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCompare(t *testing.T) {
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
				{Direction: "outbound", Service: "service1", Region: "us", Type: "ipv4", Values: []string{"192.168.1.0/24"}, Protocol: "tcp", Ports: []int{443}},
			},
			new: []types.NetworkEntry{
				{Direction: "outbound", Service: "service1", Region: "us", Type: "ipv4", Values: []string{"192.168.1.0/24"}, Protocol: "tcp", Ports: []int{443}},
				{Direction: "inbound", Service: "service2", Region: "eu", Type: "ipv4", Values: []string{"10.0.0.0/8"}, Protocol: "tcp", Ports: []int{80}},
			},
			expectedAdded:    1,
			expectedRemoved:  0,
			expectedModified: 0,
		},
		{
			name: "entry removed",
			old: []types.NetworkEntry{
				{Direction: "outbound", Service: "service1", Region: "us", Type: "ipv4", Values: []string{"192.168.1.0/24"}, Protocol: "tcp", Ports: []int{443}},
				{Direction: "inbound", Service: "service2", Region: "eu", Type: "ipv4", Values: []string{"10.0.0.0/8"}, Protocol: "tcp", Ports: []int{80}},
			},
			new: []types.NetworkEntry{
				{Direction: "outbound", Service: "service1", Region: "us", Type: "ipv4", Values: []string{"192.168.1.0/24"}, Protocol: "tcp", Ports: []int{443}},
			},
			expectedAdded:    0,
			expectedRemoved:  1,
			expectedModified: 0,
		},
		{
			name: "entry modified - ports changed",
			old: []types.NetworkEntry{
				{Direction: "outbound", Service: "test", Region: "us", Type: "ipv4", Values: []string{"192.168.1.0/24"}, Protocol: "tcp", Ports: []int{443}, Description: "Test service"},
			},
			new: []types.NetworkEntry{
				{Direction: "outbound", Service: "test", Region: "us", Type: "ipv4", Values: []string{"192.168.1.0/24"}, Protocol: "tcp", Ports: []int{443, 8443}, Description: "Test service"},
			},
			expectedAdded:    0,
			expectedRemoved:  0,
			expectedModified: 1,
		},
		{
			name: "entry modified - description changed",
			old: []types.NetworkEntry{
				{Direction: "outbound", Service: "test", Region: "us", Type: "ipv4", Values: []string{"192.168.1.0/24"}, Protocol: "tcp", Ports: []int{443}, Description: "Old description"},
			},
			new: []types.NetworkEntry{
				{Direction: "outbound", Service: "test", Region: "us", Type: "ipv4", Values: []string{"192.168.1.0/24"}, Protocol: "tcp", Ports: []int{443}, Description: "New description"},
			},
			expectedAdded:    0,
			expectedRemoved:  0,
			expectedModified: 1,
		},
		{
			name: "multiple changes",
			old: []types.NetworkEntry{
				{Direction: "outbound", Service: "service1", Region: "us", Type: "ipv4", Values: []string{"192.168.1.0/24"}, Protocol: "tcp", Ports: []int{443}},
				{Direction: "inbound", Service: "service2", Region: "eu", Type: "ipv4", Values: []string{"10.0.0.0/8"}, Protocol: "tcp", Ports: []int{80}},
			},
			new: []types.NetworkEntry{
				{Direction: "outbound", Service: "service1", Region: "us", Type: "ipv4", Values: []string{"192.168.1.0/24"}, Protocol: "tcp", Ports: []int{443, 8443}}, // modified
				{Direction: "outbound", Service: "service3", Region: "global", Type: "ipv4", Values: []string{"172.16.0.0/12"}, Protocol: "tcp", Ports: []int{443}},    // added
				// service2 removed
			},
			expectedAdded:    1,
			expectedRemoved:  1,
			expectedModified: 1,
		},
		{
			name:             "empty old, entries added",
			old:              []types.NetworkEntry{},
			new:              []types.NetworkEntry{{Direction: "outbound", Service: "service1", Region: "us", Type: "ipv4", Values: []string{"192.168.1.0/24"}, Protocol: "tcp", Ports: []int{443}}},
			expectedAdded:    1,
			expectedRemoved:  0,
			expectedModified: 0,
		},
		{
			name:             "empty new, entries removed",
			old:              []types.NetworkEntry{{Direction: "outbound", Service: "service1", Region: "us", Type: "ipv4", Values: []string{"192.168.1.0/24"}, Protocol: "tcp", Ports: []int{443}}},
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
			result := Compare(tt.old, tt.new)
			assert.Len(t, result.Added, tt.expectedAdded, "incorrect number of added entries")
			assert.Len(t, result.Removed, tt.expectedRemoved, "incorrect number of removed entries")
			assert.Len(t, result.Modified, tt.expectedModified, "incorrect number of modified entries")
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
			name:     "identical entries",
			e1:       types.NetworkEntry{Direction: "outbound", Service: "test", Description: "Test service", Ports: []int{443}},
			e2:       types.NetworkEntry{Direction: "outbound", Service: "test", Description: "Test service", Ports: []int{443}},
			expected: true,
		},
		{
			name:     "different descriptions",
			e1:       types.NetworkEntry{Description: "Old description", Ports: []int{443}},
			e2:       types.NetworkEntry{Description: "New description", Ports: []int{443}},
			expected: false,
		},
		{
			name:     "different ports",
			e1:       types.NetworkEntry{Description: "Test", Ports: []int{443}},
			e2:       types.NetworkEntry{Description: "Test", Ports: []int{443, 8443}},
			expected: false,
		},
		{
			name:     "different port values",
			e1:       types.NetworkEntry{Description: "Test", Ports: []int{443}},
			e2:       types.NetworkEntry{Description: "Test", Ports: []int{80}},
			expected: false,
		},
		{
			name:     "empty descriptions",
			e1:       types.NetworkEntry{Description: "", Ports: []int{443}},
			e2:       types.NetworkEntry{Description: "", Ports: []int{443}},
			expected: true,
		},
		{
			name:     "empty ports",
			e1:       types.NetworkEntry{Description: "Test", Ports: []int{}},
			e2:       types.NetworkEntry{Description: "Test", Ports: []int{}},
			expected: true,
		},
		{
			name:     "nil vs empty ports",
			e1:       types.NetworkEntry{Description: "Test", Ports: nil},
			e2:       types.NetworkEntry{Description: "Test", Ports: []int{}},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EntriesEqual(tt.e1, tt.e2)
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
			expected: []string{"test", "test", "test"},
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
			SortEntries(entries)
			var actualServices []string
			for _, e := range entries {
				actualServices = append(actualServices, e.Service)
			}
			assert.Equal(t, tt.expected, actualServices)
		})
	}
}
