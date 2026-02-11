package converter

import (
	"strings"
	"testing"

	"github.com/DelineaXPM/delinea-platform/delinea-netconfig/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTerraformConverter_Convert(t *testing.T) {
	tests := []struct {
		name         string
		entries      []types.NetworkEntry
		checkContent func(*testing.T, string)
	}{
		{
			name: "converts single IPv4 entry",
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
			checkContent: func(t *testing.T, content string) {
				assert.Contains(t, content, `variable "delinea_outbound_test_service_us_ipv4"`)
				assert.Contains(t, content, `description = "test_service - Test service (us)"`)
				assert.Contains(t, content, `type        = list(string)`)
				assert.Contains(t, content, `"192.168.1.0/24"`)
			},
		},
		{
			name: "converts IPv6 entry",
			entries: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "test_service",
					Region:      "us",
					Type:        "ipv6",
					Values:      []string{"2001:db8::/32"},
					Protocol:    "tcp",
					Ports:       []int{443},
					Description: "Test service",
					Redundancy:  "",
				},
			},
			checkContent: func(t *testing.T, content string) {
				assert.Contains(t, content, `variable "delinea_outbound_test_service_us_ipv6"`)
				assert.Contains(t, content, `"2001:db8::/32"`)
			},
		},
		{
			name: "skips hostname entries",
			entries: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "test_service",
					Region:      "us",
					Type:        "hostname",
					Values:      []string{"api.example.com"},
					Protocol:    "tcp",
					Ports:       []int{443},
					Description: "Test service",
					Redundancy:  "",
				},
			},
			checkContent: func(t *testing.T, content string) {
				// Should only have header comments, no variables
				assert.Contains(t, content, "# Delinea Platform Network Requirements")
				assert.NotContains(t, content, `variable`)
				assert.NotContains(t, content, "api.example.com")
			},
		},
		{
			name: "handles multiple values in single variable",
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
			checkContent: func(t *testing.T, content string) {
				assert.Contains(t, content, `variable "delinea_outbound_multi_value_global_ipv4"`)
				assert.Contains(t, content, `"192.168.1.0/24"`)
				assert.Contains(t, content, `"10.0.0.0/8"`)
				assert.Contains(t, content, `"172.16.0.0/12"`)
			},
		},
		{
			name: "includes redundancy in variable name",
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
			checkContent: func(t *testing.T, content string) {
				assert.Contains(t, content, `variable "delinea_outbound_redundant_service_us_ipv4_primary"`)
			},
		},
		{
			name: "sanitizes service names",
			entries: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "test-service.name",
					Region:      "us",
					Type:        "ipv4",
					Values:      []string{"192.168.1.0/24"},
					Protocol:    "tcp",
					Ports:       []int{443},
					Description: "Test service",
					Redundancy:  "",
				},
			},
			checkContent: func(t *testing.T, content string) {
				// Hyphens and dots should be converted to underscores in variable name
				assert.Contains(t, content, `variable "delinea_outbound_test_service_name_us_ipv4"`)
				// Original service name appears in description (which is correct behavior)
			},
		},
		{
			name: "handles empty entries",
			entries: []types.NetworkEntry{},
			checkContent: func(t *testing.T, content string) {
				// Should only have header
				assert.Contains(t, content, "# Delinea Platform Network Requirements")
				assert.NotContains(t, content, `variable`)
			},
		},
		{
			name: "creates multiple variables for different entries",
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
				},
				{
					Direction:   "inbound",
					Service:     "service2",
					Region:      "eu",
					Type:        "ipv4",
					Values:      []string{"10.0.0.0/8"},
					Protocol:    "tcp",
					Ports:       []int{80},
					Description: "Service 2",
				},
			},
			checkContent: func(t *testing.T, content string) {
				assert.Contains(t, content, `variable "delinea_outbound_service1_us_ipv4"`)
				assert.Contains(t, content, `variable "delinea_inbound_service2_eu_ipv4"`)
			},
		},
		{
			name: "merges multiple entries into same variable",
			entries: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "service1",
					Region:      "us",
					Type:        "ipv4",
					Values:      []string{"192.168.1.0/24"},
					Protocol:    "tcp",
					Ports:       []int{443},
					Description: "Service 1 - first entry",
				},
				{
					Direction:   "outbound",
					Service:     "service1",
					Region:      "us",
					Type:        "ipv4",
					Values:      []string{"10.0.0.0/8"},
					Protocol:    "tcp",
					Ports:       []int{443},
					Description: "Service 1 - second entry",
				},
			},
			checkContent: func(t *testing.T, content string) {
				// Should create only one variable with both values
				varCount := strings.Count(content, `variable "delinea_outbound_service1_us_ipv4"`)
				assert.Equal(t, 1, varCount, "Should only create one variable")

				assert.Contains(t, content, `"192.168.1.0/24"`)
				assert.Contains(t, content, `"10.0.0.0/8"`)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := &TerraformConverter{}
			result, err := converter.Convert(tt.entries)
			require.NoError(t, err)
			require.NotNil(t, result)

			content := string(result)

			// Verify header is always present
			assert.Contains(t, content, "# Delinea Platform Network Requirements")
			assert.Contains(t, content, "# Generated by delinea-netconfig")

			if tt.checkContent != nil {
				tt.checkContent(t, content)
			}
		})
	}
}

func TestTerraformConverter_Name(t *testing.T) {
	converter := &TerraformConverter{}
	assert.Equal(t, "Terraform", converter.Name())
}

func TestTerraformConverter_FileExtension(t *testing.T) {
	converter := &TerraformConverter{}
	assert.Equal(t, "tf", converter.FileExtension())
}

func TestSanitizeForTerraform(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "replaces hyphens",
			input:    "test-service",
			expected: "test_service",
		},
		{
			name:     "replaces dots",
			input:    "test.service",
			expected: "test_service",
		},
		{
			name:     "replaces spaces",
			input:    "test service",
			expected: "test_service",
		},
		{
			name:     "converts to lowercase",
			input:    "TestService",
			expected: "testservice",
		},
		{
			name:     "handles multiple special characters",
			input:    "Test-Service.Name With Spaces",
			expected: "test_service_name_with_spaces",
		},
		{
			name:     "handles already valid names",
			input:    "test_service",
			expected: "test_service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeForTerraform(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGroupEntriesForTerraform(t *testing.T) {
	tests := []struct {
		name          string
		entries       []types.NetworkEntry
		expectedVars  int
		checkGroups   func(*testing.T, map[string]terraformVarData)
	}{
		{
			name: "groups single entry",
			entries: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "test",
					Region:      "us",
					Type:        "ipv4",
					Values:      []string{"192.168.1.0/24"},
					Description: "Test",
				},
			},
			expectedVars: 1,
			checkGroups: func(t *testing.T, groups map[string]terraformVarData) {
				varData, exists := groups["delinea_outbound_test_us_ipv4"]
				assert.True(t, exists)
				assert.Len(t, varData.Values, 1)
				assert.Equal(t, "192.168.1.0/24", varData.Values[0])
			},
		},
		{
			name: "merges entries with same key",
			entries: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "test",
					Region:      "us",
					Type:        "ipv4",
					Values:      []string{"192.168.1.0/24"},
					Description: "Test 1",
				},
				{
					Direction:   "outbound",
					Service:     "test",
					Region:      "us",
					Type:        "ipv4",
					Values:      []string{"10.0.0.0/8"},
					Description: "Test 2",
				},
			},
			expectedVars: 1,
			checkGroups: func(t *testing.T, groups map[string]terraformVarData) {
				varData, exists := groups["delinea_outbound_test_us_ipv4"]
				assert.True(t, exists)
				assert.Len(t, varData.Values, 2)
			},
		},
		{
			name: "skips non-IP entries",
			entries: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "test",
					Region:      "us",
					Type:        "hostname",
					Values:      []string{"api.example.com"},
					Description: "Test",
				},
			},
			expectedVars: 0,
			checkGroups: func(t *testing.T, groups map[string]terraformVarData) {
				assert.Empty(t, groups)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groups := groupEntriesForTerraform(tt.entries)
			assert.Len(t, groups, tt.expectedVars)

			if tt.checkGroups != nil {
				tt.checkGroups(t, groups)
			}
		})
	}
}

func TestTerraformConverter_Integration(t *testing.T) {
	entries := []types.NetworkEntry{
		{
			Direction:   "outbound",
			Service:     "platform-api",
			Region:      "us",
			Type:        "ipv4",
			Values:      []string{"192.168.1.0/24", "192.168.2.0/24"},
			Protocol:    "tcp",
			Ports:       []int{443, 8443},
			Description: "Platform API endpoints",
			Redundancy:  "primary",
		},
		{
			Direction:   "inbound",
			Service:     "webhooks",
			Region:      "global",
			Type:        "ipv6",
			Values:      []string{"2001:db8::/32"},
			Protocol:    "tcp",
			Ports:       []int{443},
			Description: "Webhook endpoints",
		},
		{
			Direction:   "outbound",
			Service:     "telemetry",
			Region:      "us",
			Type:        "hostname",
			Values:      []string{"telemetry.example.com"},
			Protocol:    "tcp",
			Ports:       []int{443},
			Description: "Telemetry endpoint",
		},
	}

	converter := &TerraformConverter{}
	result, err := converter.Convert(entries)
	require.NoError(t, err)

	content := string(result)

	// Should have 2 variables (hostname is skipped)
	varCount := strings.Count(content, "variable ")
	assert.Equal(t, 2, varCount)

	// Check platform-api variable (with redundancy)
	assert.Contains(t, content, `variable "delinea_outbound_platform_api_us_ipv4_primary"`)
	assert.Contains(t, content, `"192.168.1.0/24"`)
	assert.Contains(t, content, `"192.168.2.0/24"`)

	// Check webhooks variable
	assert.Contains(t, content, `variable "delinea_inbound_webhooks_global_ipv6"`)
	assert.Contains(t, content, `"2001:db8::/32"`)

	// Should not have telemetry (hostname type)
	assert.NotContains(t, content, "telemetry")
	assert.NotContains(t, content, "telemetry.example.com")

	// Verify proper HCL format
	assert.Contains(t, content, `type        = list(string)`)
	assert.Contains(t, content, `default = [`)
}
