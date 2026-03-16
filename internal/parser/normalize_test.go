package parser

import (
	"testing"

	"github.com/DelineaXPM/delinea-netconfig/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name          string
		networkReqs   *types.NetworkRequirements
		expectedCount int
		checkFirst    func(*testing.T, types.NetworkEntry)
	}{
		{
			name: "normalizes outbound and inbound services",
			networkReqs: &types.NetworkRequirements{
				Version:   "1.0.0",
				UpdatedAt: "2025-02-10",
				Outbound: map[string]types.Service{
					"test_service": {
						Description: "Test service",
						TCPPorts:    []int{443},
						Regions: map[string]types.RegionData{
							"us": {
								IPv4: []string{"192.168.1.0/24"},
							},
						},
					},
				},
				Inbound: map[string]types.Service{
					"webhooks": {
						Description: "Webhook service",
						TCPPorts:    []int{443},
						Regions: map[string]types.RegionData{
							"global": {
								IPv4: []string{"10.0.0.0/8"},
							},
						},
					},
				},
			},
			expectedCount: 2,
			checkFirst: func(t *testing.T, entry types.NetworkEntry) {
				assert.Equal(t, "outbound", entry.Direction)
				assert.Equal(t, "test_service", entry.Service)
			},
		},
		{
			name: "handles empty network requirements",
			networkReqs: &types.NetworkRequirements{
				Version:  "1.0.0",
				Outbound: map[string]types.Service{},
				Inbound:  map[string]types.Service{},
			},
			expectedCount: 0,
		},
		{
			name: "normalizes multiple regions",
			networkReqs: &types.NetworkRequirements{
				Version: "1.0.0",
				Outbound: map[string]types.Service{
					"global_service": {
						Description: "Global service",
						TCPPorts:    []int{443},
						Regions: map[string]types.RegionData{
							"us": {
								IPv4: []string{"192.168.1.0/24"},
							},
							"eu": {
								IPv4: []string{"192.168.2.0/24"},
							},
							"au": {
								IPv4: []string{"192.168.3.0/24"},
							},
						},
					},
				},
			},
			expectedCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Normalize(tt.networkReqs)
			assert.Equal(t, tt.expectedCount, len(result))

			if tt.checkFirst != nil && len(result) > 0 {
				tt.checkFirst(t, result[0])
			}
		})
	}
}

func TestNormalizeService(t *testing.T) {
	tests := []struct {
		name          string
		direction     string
		serviceName   string
		service       types.Service
		expectedCount int
		checkEntries  func(*testing.T, []types.NetworkEntry)
	}{
		{
			name:        "normalizes IPv4 addresses",
			direction:   "outbound",
			serviceName: "test_service",
			service: types.Service{
				Description: "Test service",
				TCPPorts:    []int{443},
				Regions: map[string]types.RegionData{
					"us": {
						IPv4: []string{"192.168.1.0/24", "10.0.0.0/8"},
					},
				},
			},
			expectedCount: 1,
			checkEntries: func(t *testing.T, entries []types.NetworkEntry) {
				assert.Equal(t, "outbound", entries[0].Direction)
				assert.Equal(t, "test_service", entries[0].Service)
				assert.Equal(t, "us", entries[0].Region)
				assert.Equal(t, "ipv4", entries[0].Type)
				assert.Equal(t, []string{"192.168.1.0/24", "10.0.0.0/8"}, entries[0].Values)
				assert.Equal(t, "tcp", entries[0].Protocol)
				assert.Equal(t, []int{443}, entries[0].Ports)
				assert.Equal(t, "", entries[0].Redundancy)
			},
		},
		{
			name:        "normalizes IPv6 addresses",
			direction:   "outbound",
			serviceName: "ipv6_service",
			service: types.Service{
				Description: "IPv6 service",
				TCPPorts:    []int{443},
				Regions: map[string]types.RegionData{
					"global": {
						IPv6: []string{"2001:db8::/32"},
					},
				},
			},
			expectedCount: 1,
			checkEntries: func(t *testing.T, entries []types.NetworkEntry) {
				assert.Equal(t, "ipv6", entries[0].Type)
				assert.Equal(t, []string{"2001:db8::/32"}, entries[0].Values)
			},
		},
		{
			name:        "normalizes hostnames",
			direction:   "outbound",
			serviceName: "hostname_service",
			service: types.Service{
				Description: "Hostname service",
				TCPPorts:    []int{443},
				Regions: map[string]types.RegionData{
					"us": {
						Hostnames: []string{"api.example.com", "app.example.com"},
					},
				},
			},
			expectedCount: 1,
			checkEntries: func(t *testing.T, entries []types.NetworkEntry) {
				assert.Equal(t, "hostname", entries[0].Type)
				assert.Equal(t, []string{"api.example.com", "app.example.com"}, entries[0].Values)
			},
		},
		{
			name:        "normalizes self-signed hostnames",
			direction:   "outbound",
			serviceName: "self_signed_service",
			service: types.Service{
				Description: "Self-signed service",
				TCPPorts:    []int{443},
				Regions: map[string]types.RegionData{
					"us": {
						HostnamesSelfSigned: []string{"local.example.com"},
					},
				},
			},
			expectedCount: 1,
			checkEntries: func(t *testing.T, entries []types.NetworkEntry) {
				assert.Equal(t, "hostname_self_signed", entries[0].Type)
				assert.Equal(t, []string{"local.example.com"}, entries[0].Values)
				assert.Contains(t, entries[0].Tags, "self-signed")
			},
		},
		{
			name:        "normalizes CA-signed hostnames",
			direction:   "outbound",
			serviceName: "ca_signed_service",
			service: types.Service{
				Description: "CA-signed service",
				TCPPorts:    []int{443},
				Regions: map[string]types.RegionData{
					"us": {
						HostnamesCASigned: []string{"secure.example.com"},
					},
				},
			},
			expectedCount: 1,
			checkEntries: func(t *testing.T, entries []types.NetworkEntry) {
				assert.Equal(t, "hostname_ca_signed", entries[0].Type)
				assert.Contains(t, entries[0].Tags, "ca-signed")
			},
		},
		{
			name:        "normalizes primary and DR addresses",
			direction:   "outbound",
			serviceName: "redundant_service",
			service: types.Service{
				Description: "Redundant service",
				TCPPorts:    []int{443},
				Regions: map[string]types.RegionData{
					"us": {
						IPv4Primary: []string{"192.168.1.0/24"},
						IPv4DR:      []string{"192.168.2.0/24"},
					},
				},
			},
			expectedCount: 2,
			checkEntries: func(t *testing.T, entries []types.NetworkEntry) {
				assert.Equal(t, "primary", entries[0].Redundancy)
				assert.Equal(t, "dr", entries[1].Redundancy)
			},
		},
		{
			name:        "handles both TCP and UDP ports",
			direction:   "outbound",
			serviceName: "multi_protocol",
			service: types.Service{
				Description: "Multi-protocol service",
				TCPPorts:    []int{443, 8443},
				UDPPorts:    []int{53, 123},
				Regions: map[string]types.RegionData{
					"global": {
						IPv4: []string{"192.168.1.0/24"},
					},
				},
			},
			expectedCount: 1,
			checkEntries: func(t *testing.T, entries []types.NetworkEntry) {
				assert.Equal(t, "both", entries[0].Protocol)
				assert.Equal(t, []int{443, 8443, 53, 123}, entries[0].Ports)
			},
		},
		{
			name:        "handles all address types in one region",
			direction:   "outbound",
			serviceName: "comprehensive_service",
			service: types.Service{
				Description: "Comprehensive service",
				TCPPorts:    []int{443},
				Regions: map[string]types.RegionData{
					"us": {
						IPv4:                []string{"192.168.1.0/24"},
						IPv6:                []string{"2001:db8::/32"},
						Hostnames:           []string{"api.example.com"},
						HostnamesSelfSigned: []string{"local.example.com"},
						HostnamesCASigned:   []string{"secure.example.com"},
					},
				},
			},
			expectedCount: 5,
			checkEntries: func(t *testing.T, entries []types.NetworkEntry) {
				// Verify all 5 entry types are present
				types := make([]string, 0, 5)
				for _, entry := range entries {
					types = append(types, entry.Type)
				}
				assert.Contains(t, types, "ipv4")
				assert.Contains(t, types, "ipv6")
				assert.Contains(t, types, "hostname")
				assert.Contains(t, types, "hostname_self_signed")
				assert.Contains(t, types, "hostname_ca_signed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeService(tt.direction, tt.serviceName, tt.service)
			assert.Equal(t, tt.expectedCount, len(result))

			if tt.checkEntries != nil {
				tt.checkEntries(t, result)
			}
		})
	}
}

func TestGetProtocol(t *testing.T) {
	tests := []struct {
		name     string
		tcpPorts []int
		udpPorts []int
		expected string
	}{
		{
			name:     "TCP only",
			tcpPorts: []int{443},
			udpPorts: []int{},
			expected: "tcp",
		},
		{
			name:     "UDP only",
			tcpPorts: []int{},
			udpPorts: []int{53},
			expected: "udp",
		},
		{
			name:     "both TCP and UDP",
			tcpPorts: []int{443},
			udpPorts: []int{53},
			expected: "both",
		},
		{
			name:     "no ports",
			tcpPorts: []int{},
			udpPorts: []int{},
			expected: "",
		},
		{
			name:     "nil ports",
			tcpPorts: nil,
			udpPorts: nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getProtocol(tt.tcpPorts, tt.udpPorts)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMergePorts(t *testing.T) {
	tests := []struct {
		name     string
		tcpPorts []int
		udpPorts []int
		expected []int
	}{
		{
			name:     "merge TCP and UDP",
			tcpPorts: []int{443, 8443},
			udpPorts: []int{53, 123},
			expected: []int{443, 8443, 53, 123},
		},
		{
			name:     "TCP only",
			tcpPorts: []int{443},
			udpPorts: []int{},
			expected: []int{443},
		},
		{
			name:     "UDP only",
			tcpPorts: []int{},
			udpPorts: []int{53},
			expected: []int{53},
		},
		{
			name:     "empty ports",
			tcpPorts: []int{},
			udpPorts: []int{},
			expected: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergePorts(tt.tcpPorts, tt.udpPorts)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetPorts(t *testing.T) {
	tests := []struct {
		name            string
		service         types.Service
		expectedTCPPorts []int
		expectedUDPPorts []int
	}{
		{
			name: "direct ports",
			service: types.Service{
				TCPPorts: []int{443, 8443},
				UDPPorts: []int{53},
			},
			expectedTCPPorts: []int{443, 8443},
			expectedUDPPorts: []int{53},
		},
		{
			name: "nested external ports",
			service: types.Service{
				NestedPorts: &types.PortsNested{
					External: &types.PortSpec{
						TCPPorts: []int{443},
						UDPPorts: []int{53},
					},
				},
			},
			expectedTCPPorts: []int{443},
			expectedUDPPorts: []int{53},
		},
		{
			name: "combined direct and nested ports",
			service: types.Service{
				TCPPorts: []int{443},
				UDPPorts: []int{53},
				NestedPorts: &types.PortsNested{
					External: &types.PortSpec{
						TCPPorts: []int{8443},
						UDPPorts: []int{123},
					},
				},
			},
			expectedTCPPorts: []int{443, 8443},
			expectedUDPPorts: []int{53, 123},
		},
		{
			name: "flat ports with tcp protocol (v2 format)",
			service: types.Service{
				FlatPorts: []int{443, 5671},
				Protocol:  "tcp",
			},
			expectedTCPPorts: []int{443, 5671},
			expectedUDPPorts: nil,
		},
		{
			name: "flat ports with udp protocol (v2 format)",
			service: types.Service{
				FlatPorts: []int{1812, 1813},
				Protocol:  "udp",
			},
			expectedTCPPorts: nil,
			expectedUDPPorts: []int{1812, 1813},
		},
		{
			name: "flat ports with empty protocol defaults to tcp",
			service: types.Service{
				FlatPorts: []int{443},
			},
			expectedTCPPorts: []int{443},
			expectedUDPPorts: nil,
		},
		{
			name:            "no ports",
			service:         types.Service{},
			expectedTCPPorts: nil,
			expectedUDPPorts: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tcpPorts, udpPorts := getPorts(tt.service)
			assert.Equal(t, tt.expectedTCPPorts, tcpPorts)
			assert.Equal(t, tt.expectedUDPPorts, udpPorts)
		})
	}
}
