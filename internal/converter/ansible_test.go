package converter

import (
	"testing"

	"github.com/DelineaXPM/delinea-platform/delinea-netconfig/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestAnsibleConverter_Convert(t *testing.T) {
	tests := []struct {
		name         string
		entries      []types.NetworkEntry
		checkContent func(*testing.T, *AnsibleOutput)
	}{
		{
			name: "converts single outbound entry",
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
			checkContent: func(t *testing.T, output *AnsibleOutput) {
				assert.Len(t, output.DelineaFirewall.Outbound, 1)
				assert.Len(t, output.DelineaFirewall.Inbound, 0)

				rule := output.DelineaFirewall.Outbound[0]
				assert.Equal(t, "test_service_us_ipv4", rule.Name)
				assert.Equal(t, "test_service", rule.Service)
				assert.Equal(t, "us", rule.Region)
				assert.Equal(t, "ipv4", rule.Type)
				assert.Equal(t, []string{"192.168.1.0/24"}, rule.Destinations)
				assert.Nil(t, rule.Sources)
				assert.Equal(t, "tcp", rule.Protocol)
				assert.Equal(t, []int{443}, rule.Ports)
				assert.Equal(t, "Test service", rule.Description)
			},
		},
		{
			name: "converts single inbound entry",
			entries: []types.NetworkEntry{
				{
					Direction:   "inbound",
					Service:     "webhooks",
					Region:      "global",
					Type:        "ipv4",
					Values:      []string{"192.168.1.0/24"},
					Protocol:    "tcp",
					Ports:       []int{443},
					Description: "Webhook IPs",
					Redundancy:  "",
				},
			},
			checkContent: func(t *testing.T, output *AnsibleOutput) {
				assert.Len(t, output.DelineaFirewall.Outbound, 0)
				assert.Len(t, output.DelineaFirewall.Inbound, 1)

				rule := output.DelineaFirewall.Inbound[0]
				assert.Equal(t, "webhooks_global_ipv4", rule.Name)
				assert.Equal(t, []string{"192.168.1.0/24"}, rule.Sources)
				assert.Nil(t, rule.Destinations)
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
			checkContent: func(t *testing.T, output *AnsibleOutput) {
				rule := output.DelineaFirewall.Outbound[0]
				assert.Len(t, rule.Destinations, 3)
				assert.Equal(t, "192.168.1.0/24", rule.Destinations[0])
				assert.Equal(t, "10.0.0.0/8", rule.Destinations[1])
				assert.Equal(t, "172.16.0.0/12", rule.Destinations[2])
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
			checkContent: func(t *testing.T, output *AnsibleOutput) {
				rule := output.DelineaFirewall.Outbound[0]
				assert.Len(t, rule.Ports, 4)
				assert.Equal(t, []int{443, 8443, 53, 123}, rule.Ports)
				assert.Equal(t, "both", rule.Protocol)
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
			checkContent: func(t *testing.T, output *AnsibleOutput) {
				rule := output.DelineaFirewall.Outbound[0]
				assert.Equal(t, "primary", rule.Redundancy)
			},
		},
		{
			name: "converts entry with tags",
			entries: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "tagged_service",
					Region:      "us",
					Type:        "ipv4",
					Values:      []string{"192.168.1.0/24"},
					Protocol:    "tcp",
					Ports:       []int{443},
					Description: "Service with tags",
					Redundancy:  "",
					Tags:        []string{"critical", "production"},
				},
			},
			checkContent: func(t *testing.T, output *AnsibleOutput) {
				rule := output.DelineaFirewall.Outbound[0]
				assert.Len(t, rule.Tags, 2)
				assert.Equal(t, []string{"critical", "production"}, rule.Tags)
			},
		},
		{
			name:    "handles empty entries",
			entries: []types.NetworkEntry{},
			checkContent: func(t *testing.T, output *AnsibleOutput) {
				assert.Empty(t, output.DelineaFirewall.Outbound)
				assert.Empty(t, output.DelineaFirewall.Inbound)
				assert.Equal(t, "Delinea Network Requirements - Ansible Variables", output.Comment)
			},
		},
		{
			name: "separates outbound and inbound rules",
			entries: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "service1",
					Region:      "us",
					Type:        "ipv4",
					Values:      []string{"192.168.1.0/24"},
					Protocol:    "tcp",
					Ports:       []int{443},
					Description: "Outbound service",
				},
				{
					Direction:   "inbound",
					Service:     "service2",
					Region:      "us",
					Type:        "ipv4",
					Values:      []string{"10.0.0.0/8"},
					Protocol:    "tcp",
					Ports:       []int{80},
					Description: "Inbound service",
				},
				{
					Direction:   "outbound",
					Service:     "service3",
					Region:      "eu",
					Type:        "ipv6",
					Values:      []string{"2001:db8::/32"},
					Protocol:    "tcp",
					Ports:       []int{443},
					Description: "Another outbound service",
				},
			},
			checkContent: func(t *testing.T, output *AnsibleOutput) {
				assert.Len(t, output.DelineaFirewall.Outbound, 2)
				assert.Len(t, output.DelineaFirewall.Inbound, 1)

				// Check first outbound rule
				assert.Equal(t, "service1", output.DelineaFirewall.Outbound[0].Service)
				assert.NotNil(t, output.DelineaFirewall.Outbound[0].Destinations)

				// Check inbound rule
				assert.Equal(t, "service2", output.DelineaFirewall.Inbound[0].Service)
				assert.NotNil(t, output.DelineaFirewall.Inbound[0].Sources)

				// Check second outbound rule
				assert.Equal(t, "service3", output.DelineaFirewall.Outbound[1].Service)
			},
		},
		{
			name: "handles hostname type",
			entries: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "api",
					Region:      "global",
					Type:        "hostname",
					Values:      []string{"api.example.com", "app.example.com"},
					Protocol:    "tcp",
					Ports:       []int{443},
					Description: "API endpoints",
				},
			},
			checkContent: func(t *testing.T, output *AnsibleOutput) {
				rule := output.DelineaFirewall.Outbound[0]
				assert.Equal(t, "hostname", rule.Type)
				assert.Equal(t, "api_global_hostname", rule.Name)
				assert.Len(t, rule.Destinations, 2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := &AnsibleConverter{}
			result, err := converter.Convert(tt.entries)
			require.NoError(t, err)
			require.NotNil(t, result)

			// Parse the YAML to verify structure
			var output AnsibleOutput
			err = yaml.Unmarshal(result, &output)
			require.NoError(t, err)

			// Verify structure
			assert.NotNil(t, output.DelineaFirewall)

			if tt.checkContent != nil {
				tt.checkContent(t, &output)
			}
		})
	}
}

func TestAnsibleConverter_Name(t *testing.T) {
	converter := &AnsibleConverter{}
	assert.Equal(t, "Ansible", converter.Name())
}

func TestAnsibleConverter_FileExtension(t *testing.T) {
	converter := &AnsibleConverter{}
	assert.Equal(t, "yml", converter.FileExtension())
}

func TestAnsibleConverter_Integration(t *testing.T) {
	// This test verifies the Ansible YAML structure is valid and parseable
	entries := []types.NetworkEntry{
		{
			Direction:   "outbound",
			Service:     "platform_api",
			Region:      "us",
			Type:        "hostname",
			Values:      []string{"api.example.com", "app.example.com"},
			Protocol:    "tcp",
			Ports:       []int{443, 8443},
			Description: "Platform API endpoints",
			Redundancy:  "",
			Tags:        []string{"api", "critical"},
		},
		{
			Direction:   "inbound",
			Service:     "webhooks",
			Region:      "global",
			Type:        "ipv4",
			Values:      []string{"192.168.1.0/24", "10.0.0.0/8"},
			Protocol:    "tcp",
			Ports:       []int{443},
			Description: "Webhook IPs",
			Redundancy:  "primary",
			Tags:        []string{"webhooks"},
		},
	}

	converter := &AnsibleConverter{}
	result, err := converter.Convert(entries)
	require.NoError(t, err)

	// Parse it back
	var output AnsibleOutput
	err = yaml.Unmarshal(result, &output)
	require.NoError(t, err)

	// Verify comment
	assert.Equal(t, "Delinea Network Requirements - Ansible Variables", output.Comment)

	// Verify outbound rules
	assert.Len(t, output.DelineaFirewall.Outbound, 1)
	outboundRule := output.DelineaFirewall.Outbound[0]
	assert.Equal(t, "platform_api_us_hostname", outboundRule.Name)
	assert.Equal(t, "platform_api", outboundRule.Service)
	assert.Len(t, outboundRule.Destinations, 2)
	assert.Nil(t, outboundRule.Sources)
	assert.Len(t, outboundRule.Ports, 2)
	assert.Len(t, outboundRule.Tags, 2)

	// Verify inbound rules
	assert.Len(t, output.DelineaFirewall.Inbound, 1)
	inboundRule := output.DelineaFirewall.Inbound[0]
	assert.Equal(t, "webhooks_global_ipv4", inboundRule.Name)
	assert.Equal(t, "webhooks", inboundRule.Service)
	assert.Len(t, inboundRule.Sources, 2)
	assert.Nil(t, inboundRule.Destinations)
	assert.Equal(t, "primary", inboundRule.Redundancy)
}
