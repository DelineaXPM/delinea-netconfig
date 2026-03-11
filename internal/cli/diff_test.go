package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/DelineaXPM/delinea-platform/delinea-netconfig/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		_, _ = io.Copy(&buf, r)
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
		_, _ = io.Copy(&buf, r)
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

func TestRunDiff(t *testing.T) {
	v1JSON := `{
		"version": "1.0",
		"outbound": {
			"service_a": {
				"description": "Service A",
				"tcp_ports": [443],
				"regions": {"us": {"ipv4": ["192.168.1.0/24"]}}
			},
			"service_b": {
				"description": "Service B",
				"tcp_ports": [80],
				"regions": {"global": {"ipv4": ["10.0.0.0/8"]}}
			}
		},
		"inbound": {},
		"region_codes": {"us": "United States"}
	}`

	v2JSON := `{
		"version": "1.0",
		"outbound": {
			"service_a": {
				"description": "Service A",
				"tcp_ports": [443],
				"regions": {"us": {"ipv4": ["192.168.1.0/24"]}}
			},
			"service_c": {
				"description": "Service C (new)",
				"tcp_ports": [8443],
				"regions": {"eu": {"ipv4": ["172.16.0.0/12"]}}
			}
		},
		"inbound": {},
		"region_codes": {"us": "United States"}
	}`

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

	writeFile := func(t *testing.T, dir, name, content string) string {
		t.Helper()
		path := filepath.Join(dir, name)
		require.NoError(t, os.WriteFile(path, []byte(content), 0644))
		return path
	}

	tests := []struct {
		name        string
		file1       string
		file2       string
		summaryOnly bool
		expectErr   bool
		contains    []string
		notContains []string
	}{
		{
			name:     "detects added and removed entries",
			file1:    v1JSON,
			file2:    v2JSON,
			contains: []string{"Added", "Removed", "service_c", "service_b"},
		},
		{
			name:        "summary only flag",
			file1:       v1JSON,
			file2:       v2JSON,
			summaryOnly: true,
			contains:    []string{"Summary:", "Added:", "Removed:"},
			notContains: []string{"service_c", "+ ["},
		},
		{
			name:     "identical files show no differences",
			file1:    v1JSON,
			file2:    v1JSON,
			contains: []string{"No differences found"},
		},
		{
			name:      "missing first file returns error",
			file1:     "",
			file2:     v2JSON,
			expectErr: true,
		},
		{
			name:      "invalid JSON returns error",
			file1:     `{not valid}`,
			file2:     v2JSON,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			var path1, path2 string
			if tt.expectErr && tt.file1 == "" {
				path1 = filepath.Join(tmpDir, "nonexistent.json")
			} else {
				path1 = writeFile(t, tmpDir, "file1.json", tt.file1)
			}
			path2 = writeFile(t, tmpDir, "file2.json", tt.file2)

			oldSummary := diffSummaryOnly
			diffSummaryOnly = tt.summaryOnly
			defer func() { diffSummaryOnly = oldSummary }()

			var output string
			var err error
			output = captureOutput(func() {
				err = runDiff(diffCmd, []string{path1, path2})
			})

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				for _, s := range tt.contains {
					assert.Contains(t, output, s)
				}
				for _, s := range tt.notContains {
					assert.NotContains(t, output, s)
				}
			}
		})
	}
}
