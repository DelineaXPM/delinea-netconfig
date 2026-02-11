package converter

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/DelineaXPM/delinea-platform/delinea-netconfig/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPANOSConverter_Convert(t *testing.T) {
	tests := []struct {
		name    string
		entries []types.NetworkEntry
		checks  func(t *testing.T, output string)
		wantErr bool
	}{
		{
			name: "single outbound rule with IP address",
			entries: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "platform",
					Region:      "us",
					Type:        "ipv4",
					Values:      []string{"10.0.0.0/24"},
					Protocol:    "tcp",
					Ports:       []int{443},
					Description: "Platform API",
				},
			},
			checks: func(t *testing.T, output string) {
				assert.Contains(t, output, "<?xml version")
				assert.Contains(t, output, "<config>")
				assert.Contains(t, output, "<address>")
				assert.Contains(t, output, "ip-10-0-0-0-24")
				assert.Contains(t, output, "<ip-netmask>10.0.0.0/24</ip-netmask>")
				assert.Contains(t, output, "<service>")
				assert.Contains(t, output, "svc-platform-tcp-443")
				assert.Contains(t, output, "<port>443</port>")
				assert.Contains(t, output, "<rulebase>")
				assert.Contains(t, output, "rule-1-platform-outbound")
				assert.Contains(t, output, "<from>")
				assert.Contains(t, output, "<member>trust</member>")
				assert.Contains(t, output, "<to>")
				assert.Contains(t, output, "<member>untrust</member>")
			},
		},
		{
			name: "single inbound rule",
			entries: []types.NetworkEntry{
				{
					Direction:   "inbound",
					Service:     "webhooks",
					Region:      "us",
					Type:        "ipv4",
					Values:      []string{"192.168.0.0/24"},
					Protocol:    "tcp",
					Ports:       []int{8443},
					Description: "Webhook callbacks",
				},
			},
			checks: func(t *testing.T, output string) {
				assert.Contains(t, output, "rule-1-webhooks-inbound")
				assert.Contains(t, output, "<from>")
				assert.Contains(t, output, "<member>untrust</member>")
				assert.Contains(t, output, "<to>")
				assert.Contains(t, output, "<member>trust</member>")
				assert.Contains(t, output, "<source>")
				assert.Contains(t, output, "ip-192-168-0-0-24")
			},
		},
		{
			name: "hostname entry creates FQDN object",
			entries: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "api",
					Region:      "us",
					Type:        "hostname",
					Values:      []string{"api.example.com"},
					Protocol:    "tcp",
					Ports:       []int{443},
					Description: "API endpoint",
				},
			},
			checks: func(t *testing.T, output string) {
				assert.Contains(t, output, "fqdn-api-example-com")
				assert.Contains(t, output, "<fqdn>api.example.com</fqdn>")
				assert.NotContains(t, output, "<ip-netmask>")
			},
		},
		{
			name: "multiple ports",
			entries: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "app",
					Region:      "us",
					Type:        "ipv4",
					Values:      []string{"10.0.0.0/24"},
					Protocol:    "tcp",
					Ports:       []int{80, 443, 8080},
					Description: "Multiple ports",
				},
			},
			checks: func(t *testing.T, output string) {
				assert.Contains(t, output, "svc-app-tcp-80,443,8080")
				assert.Contains(t, output, "<port>80,443,8080</port>")
			},
		},
		{
			name: "consecutive ports create range",
			entries: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "app",
					Region:      "us",
					Type:        "ipv4",
					Values:      []string{"10.0.0.0/24"},
					Protocol:    "tcp",
					Ports:       []int{8080, 8081, 8082, 8083},
					Description: "Port range",
				},
			},
			checks: func(t *testing.T, output string) {
				assert.Contains(t, output, "svc-app-tcp-8080-8083")
				assert.Contains(t, output, "<port>8080-8083</port>")
			},
		},
		{
			name: "UDP protocol",
			entries: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "dns",
					Region:      "global",
					Type:        "ipv4",
					Values:      []string{"8.8.8.8/32"},
					Protocol:    "udp",
					Ports:       []int{53},
					Description: "DNS queries",
				},
			},
			checks: func(t *testing.T, output string) {
				assert.Contains(t, output, "svc-dns-udp-53")
				assert.Contains(t, output, "<udp>")
				assert.Contains(t, output, "<port>53</port>")
				assert.NotContains(t, output, "<tcp>")
			},
		},
		{
			name: "both protocol creates TCP and UDP services",
			entries: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "app",
					Region:      "us",
					Type:        "ipv4",
					Values:      []string{"10.0.0.0/24"},
					Protocol:    "both",
					Ports:       []int{443},
					Description: "TCP and UDP",
				},
			},
			checks: func(t *testing.T, output string) {
				assert.Contains(t, output, "svc-app-both-443-tcp")
				assert.Contains(t, output, "svc-app-both-443-udp")
				// Rule should reference both services
				assert.Contains(t, output, "<service>")
			},
		},
		{
			name: "multiple values create multiple rules",
			entries: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "cdn",
					Region:      "global",
					Type:        "ipv4",
					Values:      []string{"1.2.3.0/24", "5.6.7.0/24"},
					Protocol:    "tcp",
					Ports:       []int{443},
					Description: "CDN endpoints",
				},
			},
			checks: func(t *testing.T, output string) {
				assert.Contains(t, output, "ip-1-2-3-0-24")
				assert.Contains(t, output, "ip-5-6-7-0-24")
				assert.Contains(t, output, "rule-1-cdn-outbound")
				assert.Contains(t, output, "rule-2-cdn-outbound")
			},
		},
		{
			name: "no ports uses any service",
			entries: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "vpn",
					Region:      "global",
					Type:        "ipv4",
					Values:      []string{"10.0.0.1/32"},
					Protocol:    "tcp",
					Ports:       []int{},
					Description: "VPN",
				},
			},
			checks: func(t *testing.T, output string) {
				// Should have address but service should be "any"
				assert.Contains(t, output, "ip-10-0-0-1-32")
				// No custom service should be created
				lines := strings.Split(output, "\n")
				serviceCount := 0
				for _, line := range lines {
					if strings.Contains(line, "<entry name=\"svc-") {
						serviceCount++
					}
				}
				assert.Equal(t, 0, serviceCount, "should not create custom service when no ports specified")
			},
		},
		{
			name:    "empty entries",
			entries: []types.NetworkEntry{},
			checks: func(t *testing.T, output string) {
				assert.Contains(t, output, "<?xml version")
				assert.Contains(t, output, "<config>")
				assert.Contains(t, output, "vsys1")
				// Should not have address, service, or rulebase sections
				assert.NotContains(t, output, "<address>")
				assert.NotContains(t, output, "<service>")
				assert.NotContains(t, output, "<rulebase>")
			},
		},
		{
			name: "valid XML structure",
			entries: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "test",
					Region:      "us",
					Type:        "ipv4",
					Values:      []string{"10.0.0.0/24"},
					Protocol:    "tcp",
					Ports:       []int{443},
					Description: "Test",
				},
			},
			checks: func(t *testing.T, output string) {
				// Verify it's valid XML by unmarshaling
				var config PANOSConfig
				err := xml.Unmarshal([]byte(output), &config)
				require.NoError(t, err, "output should be valid XML")
				assert.Equal(t, "localhost.localdomain", config.Devices.Entry.Name)
				assert.Equal(t, "vsys1", config.Devices.Entry.Vsys.Entry.Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &PANOSConverter{}
			got, err := c.Convert(tt.entries)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			output := string(got)

			if tt.checks != nil {
				tt.checks(t, output)
			}
		})
	}
}

func TestPANOSConverter_sanitizeName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "spaces become hyphens",
			input: "test name with spaces",
			want:  "test-name-with-spaces",
		},
		{
			name:  "dots become hyphens",
			input: "api.example.com",
			want:  "api-example-com",
		},
		{
			name:  "colons become hyphens",
			input: "192.168.1.1:8080",
			want:  "192-168-1-1-8080",
		},
		{
			name:  "slashes become hyphens",
			input: "10.0.0.0/24",
			want:  "10-0-0-0-24",
		},
		{
			name:  "angle brackets removed",
			input: "<tenant>.example.com",
			want:  "tenant-example-com",
		},
		{
			name:  "long names truncated to 63 chars",
			input: "this-is-a-very-long-name-that-exceeds-the-maximum-length-allowed-by-panos-which-is-63-characters",
			want:  "this-is-a-very-long-name-that-exceeds-the-maximum-length-allowe",
		},
		{
			name:  "already valid name unchanged",
			input: "valid-name_123",
			want:  "valid-name_123",
		},
	}

	c := &PANOSConverter{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.sanitizeName(tt.input)
			assert.Equal(t, tt.want, got)
			assert.LessOrEqual(t, len(got), 63, "name should not exceed 63 characters")
		})
	}
}

func TestPANOSConverter_formatPortList(t *testing.T) {
	tests := []struct {
		name  string
		ports []int
		want  string
	}{
		{
			name:  "empty ports",
			ports: []int{},
			want:  "",
		},
		{
			name:  "single port",
			ports: []int{443},
			want:  "443",
		},
		{
			name:  "consecutive ports (range)",
			ports: []int{8080, 8081, 8082, 8083},
			want:  "8080-8083",
		},
		{
			name:  "non-consecutive ports (comma-separated)",
			ports: []int{80, 443, 8080},
			want:  "80,443,8080",
		},
		{
			name:  "two consecutive ports (range)",
			ports: []int{8080, 8081},
			want:  "8080-8081",
		},
		{
			name:  "many non-consecutive ports",
			ports: []int{80, 443, 8000, 8080, 9000},
			want:  "80,443,8000,8080,9000",
		},
	}

	c := &PANOSConverter{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.formatPortList(tt.ports)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPANOSConverter_Name(t *testing.T) {
	c := &PANOSConverter{}
	assert.Equal(t, "PAN-OS XML", c.Name())
}

func TestPANOSConverter_FileExtension(t *testing.T) {
	c := &PANOSConverter{}
	assert.Equal(t, "xml", c.FileExtension())
}

func TestPANOSConverter_XMLStructure(t *testing.T) {
	c := &PANOSConverter{}

	entries := []types.NetworkEntry{
		{
			Direction:   "outbound",
			Service:     "platform",
			Region:      "us",
			Type:        "ipv4",
			Values:      []string{"10.0.0.0/24"},
			Protocol:    "tcp",
			Ports:       []int{443},
			Description: "Platform API",
		},
		{
			Direction:   "inbound",
			Service:     "webhooks",
			Region:      "us",
			Type:        "ipv4",
			Values:      []string{"192.168.0.0/24"},
			Protocol:    "tcp",
			Ports:       []int{8443},
			Description: "Webhooks",
		},
	}

	output, err := c.Convert(entries)
	require.NoError(t, err)

	// Unmarshal and verify structure
	var config PANOSConfig
	err = xml.Unmarshal(output, &config)
	require.NoError(t, err)

	// Verify basic structure
	assert.Equal(t, "localhost.localdomain", config.Devices.Entry.Name)
	assert.Equal(t, "vsys1", config.Devices.Entry.Vsys.Entry.Name)

	// Verify address objects exist
	assert.NotNil(t, config.Devices.Entry.Vsys.Entry.Address)
	assert.Greater(t, len(config.Devices.Entry.Vsys.Entry.Address.Entries), 0)

	// Verify service objects exist
	assert.NotNil(t, config.Devices.Entry.Vsys.Entry.Service)
	assert.Greater(t, len(config.Devices.Entry.Vsys.Entry.Service.Entries), 0)

	// Verify rules exist
	assert.NotNil(t, config.Devices.Entry.Vsys.Entry.Rulebase)
	assert.Greater(t, len(config.Devices.Entry.Vsys.Entry.Rulebase.Security.Rules.Entries), 0)

	// Verify first rule structure
	rule := config.Devices.Entry.Vsys.Entry.Rulebase.Security.Rules.Entries[0]
	assert.NotEmpty(t, rule.Name)
	assert.Equal(t, "allow", rule.Action)
	assert.NotEmpty(t, rule.From.Members)
	assert.NotEmpty(t, rule.To.Members)
	assert.NotEmpty(t, rule.Source.Members)
	assert.NotEmpty(t, rule.Destination.Members)
	assert.NotEmpty(t, rule.Service.Members)
	assert.NotEmpty(t, rule.Application.Members)
}
