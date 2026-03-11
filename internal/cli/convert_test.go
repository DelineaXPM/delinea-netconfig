package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DelineaXPM/delinea-platform/delinea-netconfig/pkg/types"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubstituteTenant(t *testing.T) {
	tests := []struct {
		name           string
		entries        []types.NetworkEntry
		tenant         string
		expectedValues []string
		expectedDesc   string
	}{
		{
			name: "substitutes tenant in hostname values",
			entries: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "secret_server",
					Region:      "global",
					Type:        "hostname",
					Values:      []string{"<tenant>.secretservercloud.com", "api.<tenant>.secretservercloud.com"},
					Description: "Secret Server for <tenant>",
				},
			},
			tenant:         "mycompany",
			expectedValues: []string{"mycompany.secretservercloud.com", "api.mycompany.secretservercloud.com"},
			expectedDesc:   "Secret Server for <tenant>",
		},
		{
			name: "handles multiple placeholders in single value",
			entries: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "test",
					Region:      "global",
					Type:        "hostname",
					Values:      []string{"<tenant>-<tenant>.example.com"},
					Description: "<tenant> service for <tenant>",
				},
			},
			tenant:         "acme",
			expectedValues: []string{"acme-acme.example.com"},
			expectedDesc:   "<tenant> service for <tenant>",
		},
		{
			name: "preserves entries without placeholders",
			entries: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "static",
					Region:      "global",
					Type:        "ipv4",
					Values:      []string{"192.168.1.0/24"},
					Description: "Static IP range",
				},
			},
			tenant:         "mycompany",
			expectedValues: []string{"192.168.1.0/24"},
			expectedDesc:   "Static IP range",
		},
		{
			name: "handles empty values array",
			entries: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "empty",
					Region:      "global",
					Type:        "hostname",
					Values:      []string{},
					Description: "Empty values for <tenant>",
				},
			},
			tenant:         "test",
			expectedValues: []string{},
			expectedDesc:   "Empty values for <tenant>",
		},
		{
			name: "handles nil values array",
			entries: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "nil",
					Region:      "global",
					Type:        "hostname",
					Values:      nil,
					Description: "Nil values for <tenant>",
				},
			},
			tenant:         "test",
			expectedValues: nil,
			expectedDesc:   "Nil values for <tenant>",
		},
		{
			name: "case sensitive replacement",
			entries: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "test",
					Region:      "global",
					Type:        "hostname",
					Values:      []string{"<TENANT>.example.com", "<tenant>.example.com"},
					Description: "<TENANT> and <tenant>",
				},
			},
			tenant:         "mycompany",
			expectedValues: []string{"<TENANT>.example.com", "mycompany.example.com"},
			expectedDesc:   "<TENANT> and <tenant>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := substituteTenant(tt.entries, tt.tenant)

			assert.Equal(t, len(tt.entries), len(result), "should preserve entry count")

			if len(result) > 0 {
				assert.Equal(t, tt.expectedValues, result[0].Values, "values should be substituted correctly")
				assert.Equal(t, tt.expectedDesc, result[0].Description, "description should be substituted correctly")

				// Verify other fields are preserved
				assert.Equal(t, tt.entries[0].Direction, result[0].Direction, "direction should be preserved")
				assert.Equal(t, tt.entries[0].Service, result[0].Service, "service should be preserved")
				assert.Equal(t, tt.entries[0].Region, result[0].Region, "region should be preserved")
				assert.Equal(t, tt.entries[0].Type, result[0].Type, "type should be preserved")
			}
		})
	}
}

func TestSubstituteTenantMultipleEntries(t *testing.T) {
	entries := []types.NetworkEntry{
		{
			Direction:   "outbound",
			Service:     "service1",
			Region:      "us",
			Type:        "hostname",
			Values:      []string{"<tenant>.example.com"},
			Protocol:    "tcp",
			Ports:       []int{443},
			Description: "Service 1 for <tenant>",
		},
		{
			Direction:   "outbound",
			Service:     "service2",
			Region:      "eu",
			Type:        "hostname",
			Values:      []string{"api.<tenant>.example.com", "<tenant>-cdn.example.com"},
			Protocol:    "tcp",
			Ports:       []int{443},
			Description: "<tenant> CDN",
		},
		{
			Direction:   "inbound",
			Service:     "webhooks",
			Region:      "global",
			Type:        "ipv4",
			Values:      []string{"192.168.1.0/24"},
			Protocol:    "tcp",
			Ports:       []int{443},
			Description: "Static IP range",
		},
	}

	result := substituteTenant(entries, "testco")

	assert.Equal(t, 3, len(result), "should preserve all entries")

	// First entry
	assert.Equal(t, []string{"testco.example.com"}, result[0].Values)
	assert.Equal(t, "Service 1 for <tenant>", result[0].Description)
	assert.Equal(t, "service1", result[0].Service)
	assert.Equal(t, "us", result[0].Region)

	// Second entry
	assert.Equal(t, []string{"api.testco.example.com", "testco-cdn.example.com"}, result[1].Values)
	assert.Equal(t, "<tenant> CDN", result[1].Description)
	assert.Equal(t, "service2", result[1].Service)
	assert.Equal(t, "eu", result[1].Region)

	// Third entry (no substitution)
	assert.Equal(t, []string{"192.168.1.0/24"}, result[2].Values)
	assert.Equal(t, "Static IP range", result[2].Description)
	assert.Equal(t, "webhooks", result[2].Service)
	assert.Equal(t, "global", result[2].Region)
}

func TestSubstituteTenantDoesNotModifyOriginal(t *testing.T) {
	original := []types.NetworkEntry{
		{
			Direction:   "outbound",
			Service:     "test",
			Region:      "global",
			Type:        "hostname",
			Values:      []string{"<tenant>.example.com"},
			Description: "Test <tenant>",
		},
	}

	// Keep a reference to the original values
	originalValues := original[0].Values[0]
	originalDesc := original[0].Description

	// Call substituteTenant
	result := substituteTenant(original, "mycompany")

	// Verify original is not modified
	assert.Equal(t, originalValues, original[0].Values[0], "original values should not be modified")
	assert.Equal(t, originalDesc, original[0].Description, "original description should not be modified")

	// Verify result has the substitution in values only
	assert.Equal(t, "mycompany.example.com", result[0].Values[0])
	assert.Equal(t, "Test <tenant>", result[0].Description)
}

func TestRunConvert(t *testing.T) {
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

	tests := []struct {
		name        string
		format      string
		jsonContent string
		outputFile  string
		tenant      string
		expectErr   bool
	}{
		{
			name:        "convert to CSV",
			format:      "csv",
			jsonContent: validJSON,
			outputFile:  "output.csv",
			tenant:      "",
			expectErr:   false,
		},
		{
			name:        "convert to YAML",
			format:      "yaml",
			jsonContent: validJSON,
			outputFile:  "output.yaml",
			tenant:      "",
			expectErr:   false,
		},
		{
			name:        "convert to Terraform",
			format:      "terraform",
			jsonContent: validJSON,
			outputFile:  "output.tf",
			tenant:      "",
			expectErr:   false,
		},
		{
			name:        "convert with tenant substitution",
			format:      "csv",
			jsonContent: validJSON,
			outputFile:  "output.csv",
			tenant:      "mycompany",
			expectErr:   false,
		},
		{
			name:        "invalid format",
			format:      "invalid",
			jsonContent: validJSON,
			outputFile:  "output.txt",
			tenant:      "",
			expectErr:   true,
		},
		{
			name:        "file not found",
			format:      "csv",
			jsonContent: "",
			outputFile:  "output.csv",
			tenant:      "",
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tmpDir := t.TempDir()

			// Setup input file
			var inputPath string
			if tt.jsonContent != "" {
				inputPath = filepath.Join(tmpDir, "input.json")
				err := os.WriteFile(inputPath, []byte(tt.jsonContent), 0644)
				require.NoError(t, err)
			} else {
				inputPath = filepath.Join(tmpDir, "nonexistent.json")
			}

			outputPath := filepath.Join(tmpDir, tt.outputFile)

			// Reset flags
			inputFile = inputPath
			inputURL = ""
			format = tt.format
			outputFile = outputPath
			tenantName = tt.tenant

			// Create command
			cmd := &cobra.Command{}

			// Run command
			err := runConvert(cmd, []string{})

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify output file was created
				_, err := os.Stat(outputPath)
				assert.NoError(t, err, "output file should exist")
			}
		})
	}
}

func TestFilterByRegion(t *testing.T) {
	entries := []types.NetworkEntry{
		{Service: "svc1", Region: "global", Values: []string{"1.1.1.1"}},
		{Service: "svc2", Region: "eu", Values: []string{"2.2.2.2"}},
		{Service: "svc3", Region: "us", Values: []string{"3.3.3.3"}},
		{Service: "svc4", Region: "au", Values: []string{"4.4.4.4"}},
		{Service: "svc5", Region: "global", Values: []string{"5.5.5.5"}},
	}

	tests := []struct {
		name     string
		region   string
		expected []string // expected service names
	}{
		{
			name:     "eu includes global and eu only",
			region:   "eu",
			expected: []string{"svc1", "svc2", "svc5"},
		},
		{
			name:     "us includes global and us only",
			region:   "us",
			expected: []string{"svc1", "svc3", "svc5"},
		},
		{
			name:     "au includes global and au only",
			region:   "au",
			expected: []string{"svc1", "svc4", "svc5"},
		},
		{
			name:     "case insensitive match",
			region:   "EU",
			expected: []string{"svc1", "svc2", "svc5"},
		},
		{
			name:     "unknown region returns only global",
			region:   "xyz",
			expected: []string{"svc1", "svc5"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterByRegion(entries, tt.region)
			var got []string
			for _, e := range result {
				got = append(got, e.Service)
			}
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestConvertToFormat(t *testing.T) {
	entries := []types.NetworkEntry{
		{
			Direction:   "outbound",
			Service:     "test",
			Region:      "us",
			Type:        "ipv4",
			Values:      []string{"192.168.1.0/24"},
			Protocol:    "tcp",
			Ports:       []int{443},
			Description: "Test",
		},
	}

	tests := []struct {
		name       string
		formatName string
		expectErr  bool
	}{
		{
			name:       "CSV format",
			formatName: "csv",
			expectErr:  false,
		},
		{
			name:       "YAML format",
			formatName: "yaml",
			expectErr:  false,
		},
		{
			name:       "Terraform format",
			formatName: "terraform",
			expectErr:  false,
		},
		{
			name:       "Ansible format",
			formatName: "ansible",
			expectErr:  false,
		},
		{
			name:       "AWS Security Group format",
			formatName: "aws-sg",
			expectErr:  false,
		},
		{
			name:       "Cisco ACL format",
			formatName: "cisco",
			expectErr:  false,
		},
		{
			name:       "PAN-OS format",
			formatName: "panos",
			expectErr:  false,
		},
		{
			name:       "invalid format",
			formatName: "invalid",
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp output file
			tmpDir := t.TempDir()
			outputPath := filepath.Join(tmpDir, "output.txt")

			// Set outputFile flag
			outputFile = outputPath

			// Call convertToFormat (writes to file or stdout)
			err := convertToFormat(tt.formatName, entries)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify output file was created and has content
				stat, err := os.Stat(outputPath)
				if assert.NoError(t, err, "output file should exist") {
					assert.Greater(t, stat.Size(), int64(0), "output file should not be empty")
				}
			}
		})
	}
}
