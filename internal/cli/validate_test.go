package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunValidate(t *testing.T) {
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
		name        string
		jsonContent string
		expectErr   bool
		contains    []string
	}{
		{
			name: "valid JSON file",
			jsonContent: `{
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
			}`,
			expectErr: false,
			contains: []string{
				"✓ Valid JSON structure",
				"✓ Schema version: 1.0",
				"✓ All required fields present",
				"✓ 1 IPv4 ranges validated",
				"✓ 1 services validated",
			},
		},
		{
			name: "invalid JSON with warnings",
			jsonContent: `{
				"version": "1.0",
				"outbound": {
					"test": {
						"description": "Test",
						"tcp_ports": [99999],
						"regions": {
							"us": {
								"ipv4": ["256.1.1.1"]
							}
						}
					}
				},
				"inbound": {},
				"region_codes": {"us": "US"}
			}`,
			expectErr: false,
			contains: []string{
				"✓ Valid JSON structure",
				"Warnings:",
				"⚠",
			},
		},
		{
			name:        "invalid JSON syntax",
			jsonContent: `{invalid json`,
			expectErr:   true,
			contains:    []string{},
		},
		{
			name:        "empty JSON",
			jsonContent: ``,
			expectErr:   true,
			contains:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.json")
			err := os.WriteFile(tmpFile, []byte(tt.jsonContent), 0644)
			require.NoError(t, err)

			// Reset flags
			inputFile = tmpFile
			inputURL = ""

			// Create command
			cmd := &cobra.Command{}

			// Capture output and run
			var output string
			var runErr error

			if tt.expectErr {
				runErr = runValidate(cmd, []string{})
				assert.Error(t, runErr)
			} else {
				output = captureOutput(func() {
					runErr = runValidate(cmd, []string{})
				})
				assert.NoError(t, runErr)

				// Check expected output
				for _, expected := range tt.contains {
					assert.Contains(t, output, expected)
				}
			}
		})
	}
}

func TestRunValidateFileNotFound(t *testing.T) {
	// Set non-existent file
	inputFile = "/nonexistent/file.json"
	inputURL = ""

	cmd := &cobra.Command{}
	err := runValidate(cmd, []string{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read file")
}
