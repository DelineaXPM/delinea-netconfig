package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/DelineaXPM/delinea-platform/delinea-netconfig/pkg/types"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectStatistics(t *testing.T) {
	tests := []struct {
		name       string
		entries    []types.NetworkEntry
		checkStats func(*testing.T, Statistics)
	}{
		{
			name:    "empty entries",
			entries: []types.NetworkEntry{},
			checkStats: func(t *testing.T, stats Statistics) {
				assert.Equal(t, 0, stats.TotalEntries)
				assert.Equal(t, 0, stats.TotalValues)
				assert.Equal(t, 0, stats.UniqueValues)
				assert.Empty(t, stats.ByDirection)
				assert.Empty(t, stats.ByService)
				assert.Empty(t, stats.ByRegion)
				assert.Empty(t, stats.ByProtocol)
				assert.Empty(t, stats.ByType)
			},
		},
		{
			name: "single entry",
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
				},
			},
			checkStats: func(t *testing.T, stats Statistics) {
				assert.Equal(t, 1, stats.TotalEntries)
				assert.Equal(t, 1, stats.TotalValues)
				assert.Equal(t, 1, stats.UniqueValues)
				assert.Equal(t, 1, stats.ByDirection["outbound"])
				assert.Equal(t, 1, stats.ByService["test_service"])
				assert.Equal(t, 1, stats.ByRegion["us"])
				assert.Equal(t, 1, stats.ByProtocol["tcp"])
				assert.Equal(t, 1, stats.ByType["ipv4"])
				assert.Equal(t, 1, stats.PortsUsed[443])
				assert.Equal(t, 1, stats.IPv4Count)
				assert.Equal(t, 0, stats.IPv6Count)
				assert.Equal(t, 0, stats.HostnameCount)
			},
		},
		{
			name: "multiple entries same service",
			entries: []types.NetworkEntry{
				{
					Direction: "outbound",
					Service:   "platform",
					Region:    "us",
					Type:      "ipv4",
					Values:    []string{"192.168.1.0/24"},
					Protocol:  "tcp",
					Ports:     []int{443},
				},
				{
					Direction: "outbound",
					Service:   "platform",
					Region:    "eu",
					Type:      "ipv4",
					Values:    []string{"10.0.0.0/8"},
					Protocol:  "tcp",
					Ports:     []int{443},
				},
			},
			checkStats: func(t *testing.T, stats Statistics) {
				assert.Equal(t, 2, stats.TotalEntries)
				assert.Equal(t, 2, stats.TotalValues)
				assert.Equal(t, 2, stats.UniqueValues) // Different values
				assert.Equal(t, 2, stats.ByService["platform"])
				assert.Equal(t, 2, stats.PortsUsed[443]) // Port used twice
			},
		},
		{
			name: "duplicate values across entries",
			entries: []types.NetworkEntry{
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
					Direction: "outbound",
					Service:   "service2",
					Region:    "eu",
					Type:      "ipv4",
					Values:    []string{"192.168.1.0/24"}, // Same value
					Protocol:  "tcp",
					Ports:     []int{443},
				},
			},
			checkStats: func(t *testing.T, stats Statistics) {
				assert.Equal(t, 2, stats.TotalEntries)
				assert.Equal(t, 2, stats.TotalValues)
				assert.Equal(t, 1, stats.UniqueValues) // Same value, so only 1 unique
			},
		},
		{
			name: "multiple values in single entry",
			entries: []types.NetworkEntry{
				{
					Direction: "outbound",
					Service:   "multi",
					Region:    "global",
					Type:      "ipv4",
					Values:    []string{"192.168.1.0/24", "10.0.0.0/8", "172.16.0.0/12"},
					Protocol:  "tcp",
					Ports:     []int{443},
				},
			},
			checkStats: func(t *testing.T, stats Statistics) {
				assert.Equal(t, 1, stats.TotalEntries)
				assert.Equal(t, 3, stats.TotalValues)
				assert.Equal(t, 3, stats.UniqueValues)
			},
		},
		{
			name: "multiple ports",
			entries: []types.NetworkEntry{
				{
					Direction: "outbound",
					Service:   "multi_port",
					Region:    "us",
					Type:      "ipv4",
					Values:    []string{"192.168.1.0/24"},
					Protocol:  "tcp",
					Ports:     []int{443, 8443, 80, 443}, // 443 appears twice
				},
			},
			checkStats: func(t *testing.T, stats Statistics) {
				assert.Equal(t, 2, stats.PortsUsed[443]) // Used twice
				assert.Equal(t, 1, stats.PortsUsed[8443])
				assert.Equal(t, 1, stats.PortsUsed[80])
			},
		},
		{
			name: "different directions",
			entries: []types.NetworkEntry{
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
					Region:    "us",
					Type:      "ipv4",
					Values:    []string{"10.0.0.0/8"},
					Protocol:  "tcp",
					Ports:     []int{80},
				},
			},
			checkStats: func(t *testing.T, stats Statistics) {
				assert.Equal(t, 1, stats.ByDirection["outbound"])
				assert.Equal(t, 1, stats.ByDirection["inbound"])
			},
		},
		{
			name: "different protocols",
			entries: []types.NetworkEntry{
				{
					Direction: "outbound",
					Service:   "tcp_service",
					Region:    "us",
					Type:      "ipv4",
					Values:    []string{"192.168.1.0/24"},
					Protocol:  "tcp",
					Ports:     []int{443},
				},
				{
					Direction: "outbound",
					Service:   "udp_service",
					Region:    "us",
					Type:      "ipv4",
					Values:    []string{"10.0.0.0/8"},
					Protocol:  "udp",
					Ports:     []int{53},
				},
				{
					Direction: "outbound",
					Service:   "both_service",
					Region:    "us",
					Type:      "ipv4",
					Values:    []string{"172.16.0.0/12"},
					Protocol:  "both",
					Ports:     []int{80},
				},
			},
			checkStats: func(t *testing.T, stats Statistics) {
				assert.Equal(t, 1, stats.ByProtocol["tcp"])
				assert.Equal(t, 1, stats.ByProtocol["udp"])
				assert.Equal(t, 1, stats.ByProtocol["both"])
			},
		},
		{
			name: "different address types",
			entries: []types.NetworkEntry{
				{
					Direction: "outbound",
					Service:   "ipv4_service",
					Region:    "us",
					Type:      "ipv4",
					Values:    []string{"192.168.1.0/24"},
					Protocol:  "tcp",
					Ports:     []int{443},
				},
				{
					Direction: "outbound",
					Service:   "ipv6_service",
					Region:    "us",
					Type:      "ipv6",
					Values:    []string{"2001:db8::/32"},
					Protocol:  "tcp",
					Ports:     []int{443},
				},
				{
					Direction: "outbound",
					Service:   "hostname_service",
					Region:    "us",
					Type:      "hostname",
					Values:    []string{"api.example.com"},
					Protocol:  "tcp",
					Ports:     []int{443},
				},
				{
					Direction: "outbound",
					Service:   "hostname_signed",
					Region:    "us",
					Type:      "hostname_ca_signed",
					Values:    []string{"secure.example.com"},
					Protocol:  "tcp",
					Ports:     []int{443},
				},
			},
			checkStats: func(t *testing.T, stats Statistics) {
				assert.Equal(t, 1, stats.IPv4Count)
				assert.Equal(t, 1, stats.IPv6Count)
				assert.Equal(t, 2, stats.HostnameCount) // hostname + hostname_ca_signed
				assert.Equal(t, 1, stats.ByType["ipv4"])
				assert.Equal(t, 1, stats.ByType["ipv6"])
				assert.Equal(t, 1, stats.ByType["hostname"])
				assert.Equal(t, 1, stats.ByType["hostname_ca_signed"])
			},
		},
		{
			name: "multiple regions",
			entries: []types.NetworkEntry{
				{
					Direction: "outbound",
					Service:   "global_service",
					Region:    "global",
					Type:      "ipv4",
					Values:    []string{"192.168.1.0/24"},
					Protocol:  "tcp",
					Ports:     []int{443},
				},
				{
					Direction: "outbound",
					Service:   "us_service",
					Region:    "us",
					Type:      "ipv4",
					Values:    []string{"10.0.0.0/8"},
					Protocol:  "tcp",
					Ports:     []int{443},
				},
				{
					Direction: "outbound",
					Service:   "eu_service",
					Region:    "eu",
					Type:      "ipv4",
					Values:    []string{"172.16.0.0/12"},
					Protocol:  "tcp",
					Ports:     []int{443},
				},
			},
			checkStats: func(t *testing.T, stats Statistics) {
				assert.Equal(t, 1, stats.ByRegion["global"])
				assert.Equal(t, 1, stats.ByRegion["us"])
				assert.Equal(t, 1, stats.ByRegion["eu"])
			},
		},
		{
			name: "services per direction",
			entries: []types.NetworkEntry{
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
					Direction: "outbound",
					Service:   "service1",
					Region:    "eu",
					Type:      "ipv4",
					Values:    []string{"10.0.0.0/8"},
					Protocol:  "tcp",
					Ports:     []int{443},
				},
				{
					Direction: "inbound",
					Service:   "service2",
					Region:    "global",
					Type:      "ipv4",
					Values:    []string{"172.16.0.0/12"},
					Protocol:  "tcp",
					Ports:     []int{80},
				},
			},
			checkStats: func(t *testing.T, stats Statistics) {
				assert.Equal(t, 2, stats.ServicesPerDir["outbound"]["service1"])
				assert.Equal(t, 1, stats.ServicesPerDir["inbound"]["service2"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := collectStatistics(tt.entries)

			if tt.checkStats != nil {
				tt.checkStats(t, stats)
			}
		})
	}
}

func TestCollectStatistics_Integration(t *testing.T) {
	// Complex scenario with multiple types of entries
	entries := []types.NetworkEntry{
		{
			Direction: "outbound",
			Service:   "platform_api",
			Region:    "us",
			Type:      "hostname",
			Values:    []string{"api.example.com", "app.example.com"},
			Protocol:  "tcp",
			Ports:     []int{443, 8443},
		},
		{
			Direction: "outbound",
			Service:   "platform_api",
			Region:    "eu",
			Type:      "ipv4",
			Values:    []string{"192.168.1.0/24", "10.0.0.0/8"},
			Protocol:  "tcp",
			Ports:     []int{443},
		},
		{
			Direction: "inbound",
			Service:   "webhooks",
			Region:    "global",
			Type:      "ipv4",
			Values:    []string{"192.168.1.0/24"}, // Duplicate value
			Protocol:  "tcp",
			Ports:     []int{443},
		},
		{
			Direction: "outbound",
			Service:   "telemetry",
			Region:    "us",
			Type:      "ipv6",
			Values:    []string{"2001:db8::/32"},
			Protocol:  "udp",
			Ports:     []int{53, 123},
		},
	}

	stats := collectStatistics(entries)

	// Overall counts
	assert.Equal(t, 4, stats.TotalEntries)
	assert.Equal(t, 6, stats.TotalValues) // 2 + 2 + 1 + 1
	assert.Equal(t, 5, stats.UniqueValues) // One duplicate (192.168.1.0/24)

	// By direction
	assert.Equal(t, 3, stats.ByDirection["outbound"])
	assert.Equal(t, 1, stats.ByDirection["inbound"])

	// By service
	assert.Equal(t, 2, stats.ByService["platform_api"])
	assert.Equal(t, 1, stats.ByService["webhooks"])
	assert.Equal(t, 1, stats.ByService["telemetry"])

	// By region
	assert.Equal(t, 2, stats.ByRegion["us"])
	assert.Equal(t, 1, stats.ByRegion["eu"])
	assert.Equal(t, 1, stats.ByRegion["global"])

	// By protocol
	assert.Equal(t, 3, stats.ByProtocol["tcp"])
	assert.Equal(t, 1, stats.ByProtocol["udp"])

	// Address types
	assert.Equal(t, 1, stats.HostnameCount)
	assert.Equal(t, 2, stats.IPv4Count)
	assert.Equal(t, 1, stats.IPv6Count)

	// Ports
	assert.Equal(t, 3, stats.PortsUsed[443]) // Used 3 times
	assert.Equal(t, 1, stats.PortsUsed[8443])
	assert.Equal(t, 1, stats.PortsUsed[53])
	assert.Equal(t, 1, stats.PortsUsed[123])

	// Services per direction
	assert.Equal(t, 2, stats.ServicesPerDir["outbound"]["platform_api"])
	assert.Equal(t, 1, stats.ServicesPerDir["outbound"]["telemetry"])
	assert.Equal(t, 1, stats.ServicesPerDir["inbound"]["webhooks"])
}

func TestSortedKeys(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]int
		expected []string
	}{
		{
			name:     "empty map",
			input:    map[string]int{},
			expected: []string{},
		},
		{
			name: "single key",
			input: map[string]int{
				"key1": 1,
			},
			expected: []string{"key1"},
		},
		{
			name: "multiple keys sorted",
			input: map[string]int{
				"zebra": 1,
				"alpha": 2,
				"beta":  3,
			},
			expected: []string{"alpha", "beta", "zebra"},
		},
		{
			name: "numeric strings",
			input: map[string]int{
				"10": 1,
				"2":  2,
				"1":  3,
			},
			expected: []string{"1", "10", "2"}, // Lexicographic sort
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sortedKeys(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDisplayStatistics(t *testing.T) {
	// Helper to capture stdout
	captureOutput := func(f func()) string {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		f()

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		return buf.String()
	}

	tests := []struct {
		name     string
		stats    Statistics
		verbose  bool
		contains []string
	}{
		{
			name: "basic statistics",
			stats: Statistics{
				TotalEntries: 10,
				TotalValues:  25,
				UniqueValues: 20,
				ByDirection: map[string]int{
					"outbound": 8,
					"inbound":  2,
				},
				ByService: map[string]int{
					"platform_api": 5,
					"webhooks":     5,
				},
				ByRegion: map[string]int{
					"us": 7,
					"eu": 3,
				},
				ByProtocol: map[string]int{
					"tcp": 9,
					"udp": 1,
				},
				HostnameCount: 5,
				IPv4Count:     15,
				IPv6Count:     5,
				PortsUsed: map[int]int{
					443: 10,
					80:  5,
				},
			},
			verbose: false,
			contains: []string{
				"Overview:",
				"Total Entries:    10",
				"Total Values:     25",
				"Unique Values:    20",
				"By Direction:",
				"outbound",
				"inbound",
				"By Service:",
				"platform_api",
				"webhooks",
				"By Region:",
				"By Protocol:",
				"By Type:",
				"Hostnames:        5",
				"IPv4 Addresses:   15",
				"IPv6 Addresses:   5",
				"Ports Used:",
			},
		},
		{
			name: "verbose mode shows services per direction",
			stats: Statistics{
				TotalEntries: 2,
				TotalValues:  2,
				UniqueValues: 2,
				ByDirection:  map[string]int{"outbound": 2},
				ByService:    map[string]int{"test": 2},
				ByRegion:     map[string]int{"us": 2},
				ByProtocol:   map[string]int{"tcp": 2},
				ServicesPerDir: map[string]map[string]int{
					"outbound": {
						"service1": 1,
						"service2": 1,
					},
				},
			},
			verbose: true,
			contains: []string{
				"Services per Direction:",
				"outbound:",
				"service1",
				"service2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set verbose mode
			oldVerbose := verbose
			verbose = tt.verbose
			defer func() { verbose = oldVerbose }()

			output := captureOutput(func() {
				displayStatistics(tt.stats)
			})

			for _, expected := range tt.contains {
				assert.Contains(t, output, expected)
			}
		})
	}
}

func TestDisplayMapSorted(t *testing.T) {
	// Helper to capture stdout
	captureOutput := func(f func()) string {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		f()

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		return buf.String()
	}

	tests := []struct {
		name     string
		m        map[string]int
		indent   string
		contains []string
	}{
		{
			name: "displays sorted map",
			m: map[string]int{
				"zebra": 5,
				"apple": 10,
				"beta":  3,
			},
			indent: "  ",
			contains: []string{
				"apple",
				"beta",
				"zebra",
			},
		},
		{
			name: "handles empty map",
			m:    map[string]int{},
			indent: "",
			contains: []string{},
		},
		{
			name: "displays with custom indent",
			m: map[string]int{
				"test": 1,
			},
			indent: "    ",
			contains: []string{
				"    test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				displayMapSorted(tt.m, tt.indent)
			})

			for _, expected := range tt.contains {
				assert.Contains(t, output, expected)
			}
		})
	}
}

func TestDisplayPortsSorted(t *testing.T) {
	// Helper to capture stdout
	captureOutput := func(f func()) string {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		f()

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		return buf.String()
	}

	tests := []struct {
		name     string
		ports    map[int]int
		indent   string
		verbose  bool
		contains []string
	}{
		{
			name: "displays ports sorted by count descending",
			ports: map[int]int{
				443:  10,
				80:   5,
				8443: 15,
			},
			indent: "  ",
			contains: []string{
				"8443",
				"443",
				"80",
				"used 15 times",
				"used 10 times",
				"used 5 times",
			},
		},
		{
			name: "limits to top 10 ports in non-verbose mode",
			ports: map[int]int{
				1: 15, 2: 14, 3: 13, 4: 12, 5: 11,
				6: 10, 7: 9, 8: 8, 9: 7, 10: 6,
				11: 5, 12: 4,
			},
			indent:  "  ",
			verbose: false,
			contains: []string{
				"... (2 more ports, use -v to see all)",
			},
		},
		{
			name: "shows all ports in verbose mode",
			ports: map[int]int{
				1: 15, 2: 14, 3: 13, 4: 12, 5: 11,
				6: 10, 7: 9, 8: 8, 9: 7, 10: 6,
				11: 5, 12: 4,
			},
			indent:  "  ",
			verbose: true,
			contains: []string{
				"11",
				"12",
			},
		},
		{
			name:     "handles empty ports map",
			ports:    map[int]int{},
			indent:   "",
			verbose:  false,
			contains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set verbose mode
			oldVerbose := verbose
			verbose = tt.verbose
			defer func() { verbose = oldVerbose }()

			output := captureOutput(func() {
				displayPortsSorted(tt.ports, tt.indent)
			})

			for _, expected := range tt.contains {
				assert.Contains(t, output, expected)
			}
		})
	}
}

func TestBuildBaseURL(t *testing.T) {
	tests := []struct {
		name     string
		tenant   string
		expected string
	}{
		{
			name:     "default (no tenant)",
			tenant:   "",
			expected: "https://setup.delinea.app",
		},
		{
			name:     "with tenant",
			tenant:   "mycompany",
			expected: "https://mycompany.delinea.app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldTenant := infoTenant
			infoTenant = tt.tenant
			defer func() { infoTenant = oldTenant }()

			result := buildBaseURL()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRunInfoRequiresArgsOrFlags(t *testing.T) {
	cmd := &cobra.Command{}

	// Reset flags
	oldUpdates := infoUpdates
	oldLatest := infoLatest
	infoUpdates = false
	infoLatest = false
	defer func() {
		infoUpdates = oldUpdates
		infoLatest = oldLatest
	}()

	err := runInfo(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file argument required")
}

func TestRunInfo(t *testing.T) {
	validJSON := `{
		"version": "1.0",
		"outbound": {
			"test_service": {
				"description": "Test",
				"tcp_ports": [443],
				"regions": {
					"us": {
						"ipv4": ["192.168.1.0/24"]
					}
				}
			}
		},
		"inbound": {},
		"region_codes": {"us": "United States"}
	}`

	t.Run("valid JSON file", func(t *testing.T) {
		// Create temp file
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "test.json")
		err := os.WriteFile(tmpFile, []byte(validJSON), 0644)
		require.NoError(t, err)

		// Create command
		cmd := &cobra.Command{}

		// Capture output
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Pass file as argument
		err = runInfo(cmd, []string{tmpFile})

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		output := buf.String()

		// Verify
		assert.NoError(t, err)
		assert.Contains(t, output, "Overview:")
		assert.Contains(t, output, "Total Entries:")
		assert.Contains(t, output, "By Direction:")
	})

	t.Run("file not found", func(t *testing.T) {
		cmd := &cobra.Command{}

		// Pass non-existent file as argument
		err := runInfo(cmd, []string{"/nonexistent/file.json"})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read file")
	})
}
