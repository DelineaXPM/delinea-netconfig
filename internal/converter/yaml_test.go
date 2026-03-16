package converter

import (
	"testing"

	"github.com/DelineaXPM/delinea-netconfig/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/goccy/go-yaml"
)

func TestYAMLConverter_Convert(t *testing.T) {
	tests := []struct {
		name         string
		entries      []types.NetworkEntry
		checkContent func(*testing.T, map[string]interface{})
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
			checkContent: func(t *testing.T, data map[string]interface{}) {
				reqs := data["delinea_network_requirements"].(map[string]interface{})
				assert.Contains(t, reqs, "outbound")

				outbound := reqs["outbound"].(map[string]interface{})
				assert.Contains(t, outbound, "test_service")

				service := outbound["test_service"].(map[string]interface{})
				assert.Contains(t, service, "us")

				region := service["us"].([]interface{})
				assert.Len(t, region, 1)

				entry := region[0].(map[string]interface{})
				assert.Equal(t, "ipv4", entry["type"])
				assert.Equal(t, "tcp", entry["protocol"])
				assert.Equal(t, "Test service", entry["description"])
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
					Values:      []string{"192.168.1.0/24", "10.0.0.0/8"},
					Protocol:    "tcp",
					Ports:       []int{443},
					Description: "Multi-value service",
					Redundancy:  "",
				},
			},
			checkContent: func(t *testing.T, data map[string]interface{}) {
				reqs := data["delinea_network_requirements"].(map[string]interface{})
				outbound := reqs["outbound"].(map[string]interface{})
				service := outbound["multi_value"].(map[string]interface{})
				region := service["global"].([]interface{})
				entry := region[0].(map[string]interface{})

				values := entry["values"].([]interface{})
				assert.Len(t, values, 2)
				assert.Equal(t, "192.168.1.0/24", values[0])
				assert.Equal(t, "10.0.0.0/8", values[1])
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
					Ports:       []int{443, 8443, 53},
					Description: "Multi-port service",
					Redundancy:  "",
				},
			},
			checkContent: func(t *testing.T, data map[string]interface{}) {
				reqs := data["delinea_network_requirements"].(map[string]interface{})
				outbound := reqs["outbound"].(map[string]interface{})
				service := outbound["multi_port"].(map[string]interface{})
				region := service["us"].([]interface{})
				entry := region[0].(map[string]interface{})

				ports := entry["ports"].([]interface{})
				assert.Len(t, ports, 3)
				assert.EqualValues(t, 443, ports[0])
				assert.EqualValues(t, 8443, ports[1])
				assert.EqualValues(t, 53, ports[2])
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
			checkContent: func(t *testing.T, data map[string]interface{}) {
				reqs := data["delinea_network_requirements"].(map[string]interface{})
				outbound := reqs["outbound"].(map[string]interface{})
				service := outbound["redundant_service"].(map[string]interface{})
				region := service["us"].([]interface{})
				entry := region[0].(map[string]interface{})

				assert.Equal(t, "primary", entry["redundancy"])
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
			checkContent: func(t *testing.T, data map[string]interface{}) {
				reqs := data["delinea_network_requirements"].(map[string]interface{})
				outbound := reqs["outbound"].(map[string]interface{})
				service := outbound["tagged_service"].(map[string]interface{})
				region := service["us"].([]interface{})
				entry := region[0].(map[string]interface{})

				tags := entry["tags"].([]interface{})
				assert.Len(t, tags, 2)
				assert.Equal(t, "critical", tags[0])
				assert.Equal(t, "production", tags[1])
			},
		},
		{
			name:    "handles empty entries",
			entries: []types.NetworkEntry{},
			checkContent: func(t *testing.T, data map[string]interface{}) {
				reqs := data["delinea_network_requirements"].(map[string]interface{})
				assert.Empty(t, reqs)
			},
		},
		{
			name: "organizes multiple directions",
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
			},
			checkContent: func(t *testing.T, data map[string]interface{}) {
				reqs := data["delinea_network_requirements"].(map[string]interface{})
				assert.Contains(t, reqs, "outbound")
				assert.Contains(t, reqs, "inbound")

				outbound := reqs["outbound"].(map[string]interface{})
				assert.Contains(t, outbound, "service1")

				inbound := reqs["inbound"].(map[string]interface{})
				assert.Contains(t, inbound, "service2")
			},
		},
		{
			name: "organizes multiple regions",
			entries: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "service1",
					Region:      "us",
					Type:        "ipv4",
					Values:      []string{"192.168.1.0/24"},
					Protocol:    "tcp",
					Ports:       []int{443},
					Description: "US region",
				},
				{
					Direction:   "outbound",
					Service:     "service1",
					Region:      "eu",
					Type:        "ipv4",
					Values:      []string{"10.0.0.0/8"},
					Protocol:    "tcp",
					Ports:       []int{443},
					Description: "EU region",
				},
			},
			checkContent: func(t *testing.T, data map[string]interface{}) {
				reqs := data["delinea_network_requirements"].(map[string]interface{})
				outbound := reqs["outbound"].(map[string]interface{})
				service := outbound["service1"].(map[string]interface{})

				assert.Contains(t, service, "us")
				assert.Contains(t, service, "eu")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := &YAMLConverter{}
			result, err := converter.Convert(tt.entries)
			require.NoError(t, err)
			require.NotNil(t, result)

			// Parse the YAML to verify structure
			var data map[string]interface{}
			err = yaml.Unmarshal(result, &data)
			require.NoError(t, err)

			// Verify structure
			assert.Contains(t, data, "delinea_network_requirements")

			if tt.checkContent != nil {
				tt.checkContent(t, data)
			}
		})
	}
}

func TestYAMLConverter_Name(t *testing.T) {
	converter := &YAMLConverter{}
	assert.Equal(t, "YAML", converter.Name())
}

func TestYAMLConverter_FileExtension(t *testing.T) {
	converter := &YAMLConverter{}
	assert.Equal(t, "yaml", converter.FileExtension())
}

func TestYAMLConverter_Integration(t *testing.T) {
	// This test verifies the YAML structure is valid and parseable
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
			Tags:        []string{"critical"},
		},
	}

	converter := &YAMLConverter{}
	result, err := converter.Convert(entries)
	require.NoError(t, err)

	// Parse it back
	var data map[string]interface{}
	err = yaml.Unmarshal(result, &data)
	require.NoError(t, err)

	// Verify structure
	reqs := data["delinea_network_requirements"].(map[string]interface{})

	// Check outbound
	outbound := reqs["outbound"].(map[string]interface{})
	platform := outbound["platform"].(map[string]interface{})
	usRegion := platform["us"].([]interface{})
	assert.Len(t, usRegion, 1)

	// Check inbound
	inbound := reqs["inbound"].(map[string]interface{})
	webhooks := inbound["webhooks"].(map[string]interface{})
	globalRegion := webhooks["global"].([]interface{})
	assert.Len(t, globalRegion, 1)

	webhookEntry := globalRegion[0].(map[string]interface{})
	assert.Equal(t, "primary", webhookEntry["redundancy"])

	tags := webhookEntry["tags"].([]interface{})
	assert.Len(t, tags, 1)
	assert.Equal(t, "critical", tags[0])
}
