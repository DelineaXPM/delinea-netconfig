package converter

import (
	"encoding/csv"
	"strings"
	"testing"

	"github.com/DelineaXPM/delinea-platform/delinea-netconfig/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCSVConverter_Convert(t *testing.T) {
	tests := []struct {
		name          string
		entries       []types.NetworkEntry
		expectedRows  int // Including header
		checkContent  func(*testing.T, [][]string)
	}{
		{
			name: "converts single entry with single value",
			entries: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "test_service",
					Region:      "us",
					Type:        "ipv4",
					Values:      []string{"192.168.1.0/24"},
					Protocol:    "tcp",
					Ports:       []int{443},
					Description: "Test service",
					Redundancy:  "",
				},
			},
			expectedRows: 2, // header + 1 data row
			checkContent: func(t *testing.T, rows [][]string) {
				assert.Equal(t, "direction", rows[0][0])
				assert.Equal(t, "outbound", rows[1][0])
				assert.Equal(t, "test_service", rows[1][1])
				assert.Equal(t, "us", rows[1][2])
				assert.Equal(t, "ipv4", rows[1][3])
				assert.Equal(t, "192.168.1.0/24", rows[1][4])
				assert.Equal(t, "tcp", rows[1][5])
				assert.Equal(t, "443", rows[1][6])
				assert.Equal(t, "Test service", rows[1][7])
				assert.Equal(t, "", rows[1][8])
			},
		},
		{
			name: "converts entry with multiple values",
			entries: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "multi_value",
					Region:      "global",
					Type:        "ipv4",
					Values:      []string{"192.168.1.0/24", "10.0.0.0/8", "172.16.0.0/12"},
					Protocol:    "tcp",
					Ports:       []int{443},
					Description: "Multi-value service",
					Redundancy:  "",
				},
			},
			expectedRows: 4, // header + 3 data rows (one per value)
			checkContent: func(t *testing.T, rows [][]string) {
				assert.Equal(t, "192.168.1.0/24", rows[1][4])
				assert.Equal(t, "10.0.0.0/8", rows[2][4])
				assert.Equal(t, "172.16.0.0/12", rows[3][4])
				// All rows should have the same service name
				assert.Equal(t, "multi_value", rows[1][1])
				assert.Equal(t, "multi_value", rows[2][1])
				assert.Equal(t, "multi_value", rows[3][1])
			},
		},
		{
			name: "converts entry with multiple ports",
			entries: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "multi_port",
					Region:      "us",
					Type:        "ipv4",
					Values:      []string{"192.168.1.0/24"},
					Protocol:    "both",
					Ports:       []int{443, 8443, 53, 123},
					Description: "Multi-port service",
					Redundancy:  "",
				},
			},
			expectedRows: 2,
			checkContent: func(t *testing.T, rows [][]string) {
				assert.Equal(t, "443,8443,53,123", rows[1][6])
				assert.Equal(t, "both", rows[1][5])
			},
		},
		{
			name: "converts entry with redundancy",
			entries: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "redundant_service",
					Region:      "us",
					Type:        "ipv4",
					Values:      []string{"192.168.1.0/24"},
					Protocol:    "tcp",
					Ports:       []int{443},
					Description: "Primary endpoint",
					Redundancy:  "primary",
				},
			},
			expectedRows: 2,
			checkContent: func(t *testing.T, rows [][]string) {
				assert.Equal(t, "primary", rows[1][8])
			},
		},
		{
			name: "handles empty entries",
			entries: []types.NetworkEntry{},
			expectedRows: 1, // header only
			checkContent: func(t *testing.T, rows [][]string) {
				assert.Equal(t, "direction", rows[0][0])
			},
		},
		{
			name: "handles entry with no ports",
			entries: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "no_ports",
					Region:      "global",
					Type:        "ipv4",
					Values:      []string{"192.168.1.0/24"},
					Protocol:    "",
					Ports:       []int{},
					Description: "No ports specified",
					Redundancy:  "",
				},
			},
			expectedRows: 2,
			checkContent: func(t *testing.T, rows [][]string) {
				assert.Equal(t, "", rows[1][6]) // Empty ports field
			},
		},
		{
			name: "converts multiple entries",
			entries: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "service1",
					Region:      "us",
					Type:        "ipv4",
					Values:      []string{"192.168.1.0/24"},
					Protocol:    "tcp",
					Ports:       []int{443},
					Description: "Service 1",
					Redundancy:  "",
				},
				{
					Direction:   "inbound",
					Service:     "service2",
					Region:      "eu",
					Type:        "ipv6",
					Values:      []string{"2001:db8::/32"},
					Protocol:    "udp",
					Ports:       []int{53},
					Description: "Service 2",
					Redundancy:  "dr",
				},
			},
			expectedRows: 3, // header + 2 data rows
			checkContent: func(t *testing.T, rows [][]string) {
				assert.Equal(t, "outbound", rows[1][0])
				assert.Equal(t, "service1", rows[1][1])
				assert.Equal(t, "inbound", rows[2][0])
				assert.Equal(t, "service2", rows[2][1])
			},
		},
		{
			name: "handles special characters in description",
			entries: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "test",
					Region:      "global",
					Type:        "hostname",
					Values:      []string{"api.example.com"},
					Protocol:    "tcp",
					Ports:       []int{443},
					Description: "Test with \"quotes\" and, commas",
					Redundancy:  "",
				},
			},
			expectedRows: 2,
			checkContent: func(t *testing.T, rows [][]string) {
				// CSV library should handle escaping
				assert.Contains(t, rows[1][7], "quotes")
				assert.Contains(t, rows[1][7], "commas")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := &CSVConverter{}
			result, err := converter.Convert(tt.entries)
			require.NoError(t, err)
			require.NotNil(t, result)

			// Parse the CSV to verify structure
			reader := csv.NewReader(strings.NewReader(string(result)))
			rows, err := reader.ReadAll()
			require.NoError(t, err)

			assert.Equal(t, tt.expectedRows, len(rows))

			if tt.checkContent != nil {
				tt.checkContent(t, rows)
			}

			// Verify header is always present and correct
			if len(rows) > 0 {
				expectedHeader := []string{"direction", "service", "region", "type", "value", "protocol", "ports", "description", "redundancy"}
				assert.Equal(t, expectedHeader, rows[0])
			}
		})
	}
}

func TestCSVConverter_Name(t *testing.T) {
	converter := &CSVConverter{}
	assert.Equal(t, "CSV", converter.Name())
}

func TestCSVConverter_FileExtension(t *testing.T) {
	converter := &CSVConverter{}
	assert.Equal(t, "csv", converter.FileExtension())
}

func TestFormatPorts(t *testing.T) {
	tests := []struct {
		name     string
		ports    []int
		expected string
	}{
		{
			name:     "single port",
			ports:    []int{443},
			expected: "443",
		},
		{
			name:     "multiple ports",
			ports:    []int{443, 8443, 80},
			expected: "443,8443,80",
		},
		{
			name:     "empty ports",
			ports:    []int{},
			expected: "",
		},
		{
			name:     "nil ports",
			ports:    nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatPorts(tt.ports)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCSVConverter_Integration(t *testing.T) {
	// This test verifies the CSV can be parsed back correctly
	entries := []types.NetworkEntry{
		{
			Direction:   "outbound",
			Service:     "platform",
			Region:      "us",
			Type:        "hostname",
			Values:      []string{"api.example.com", "app.example.com"},
			Protocol:    "tcp",
			Ports:       []int{443, 8443},
			Description: "Platform API endpoints",
			Redundancy:  "",
		},
		{
			Direction:   "inbound",
			Service:     "webhooks",
			Region:      "global",
			Type:        "ipv4",
			Values:      []string{"192.168.1.0/24"},
			Protocol:    "tcp",
			Ports:       []int{443},
			Description: "Webhook IPs",
			Redundancy:  "primary",
		},
	}

	converter := &CSVConverter{}
	result, err := converter.Convert(entries)
	require.NoError(t, err)

	// Parse it back
	reader := csv.NewReader(strings.NewReader(string(result)))
	rows, err := reader.ReadAll()
	require.NoError(t, err)

	// Should have header + 3 data rows (2 values from first entry + 1 from second)
	assert.Equal(t, 4, len(rows))

	// Verify first data row
	assert.Equal(t, "outbound", rows[1][0])
	assert.Equal(t, "platform", rows[1][1])
	assert.Equal(t, "api.example.com", rows[1][4])
	assert.Equal(t, "443,8443", rows[1][6])

	// Verify second data row (second value from first entry)
	assert.Equal(t, "app.example.com", rows[2][4])

	// Verify third data row (second entry)
	assert.Equal(t, "inbound", rows[3][0])
	assert.Equal(t, "webhooks", rows[3][1])
	assert.Equal(t, "primary", rows[3][8])
}
