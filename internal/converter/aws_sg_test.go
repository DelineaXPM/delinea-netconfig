package converter

import (
	"encoding/json"
	"testing"

	"github.com/DelineaXPM/delinea-netconfig/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAWSSecurityGroupConverter_Convert(t *testing.T) {
	tests := []struct {
		name         string
		entries      []types.NetworkEntry
		checkContent func(*testing.T, *AWSSecurityGroupOutput)
	}{
		{
			name: "converts single inbound IPv4 entry",
			entries: []types.NetworkEntry{
				{
					Direction:   "inbound",
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
			checkContent: func(t *testing.T, output *AWSSecurityGroupOutput) {
				assert.Equal(t, "Delinea Platform Network Requirements", output.Description)
				assert.Equal(t, "delinea-platform-sg", output.GroupName)

				// Check inbound permissions
				assert.Len(t, output.IpPermissions, 1)
				perm := output.IpPermissions[0]
				assert.Equal(t, "tcp", perm.IpProtocol)
				assert.Equal(t, 443, perm.FromPort)
				assert.Equal(t, 443, perm.ToPort)
				assert.Len(t, perm.IpRanges, 1)
				assert.Equal(t, "192.168.1.0/24", perm.IpRanges[0].CidrIp)
				assert.Contains(t, perm.IpRanges[0].Description, "test_service")

				// No egress permissions
				assert.Empty(t, output.IpPermissionsEgress)
			},
		},
		{
			name: "converts single outbound IPv4 entry",
			entries: []types.NetworkEntry{
				{
					Direction:   "outbound",
					Service:     "api",
					Region:      "us",
					Type:        "ipv4",
					Values:      []string{"10.0.0.0/8"},
					Protocol:    "tcp",
					Ports:       []int{443},
					Description: "API service",
					Redundancy:  "",
				},
			},
			checkContent: func(t *testing.T, output *AWSSecurityGroupOutput) {
				// No inbound permissions
				assert.Empty(t, output.IpPermissions)

				// Check egress permissions
				assert.Len(t, output.IpPermissionsEgress, 1)
				perm := output.IpPermissionsEgress[0]
				assert.Equal(t, "tcp", perm.IpProtocol)
				assert.Equal(t, 443, perm.FromPort)
				assert.Equal(t, 443, perm.ToPort)
				assert.Len(t, perm.IpRanges, 1)
				assert.Equal(t, "10.0.0.0/8", perm.IpRanges[0].CidrIp)
			},
		},
		{
			name: "converts IPv6 entry",
			entries: []types.NetworkEntry{
				{
					Direction:   "inbound",
					Service:     "test",
					Region:      "us",
					Type:        "ipv6",
					Values:      []string{"2001:db8::/32"},
					Protocol:    "tcp",
					Ports:       []int{443},
					Description: "IPv6 test",
				},
			},
			checkContent: func(t *testing.T, output *AWSSecurityGroupOutput) {
				perm := output.IpPermissions[0]
				assert.Empty(t, perm.IpRanges)
				assert.Len(t, perm.Ipv6Ranges, 1)
				assert.Equal(t, "2001:db8::/32", perm.Ipv6Ranges[0].CidrIpv6)
			},
		},
		{
			name: "skips hostname entries",
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
			checkContent: func(t *testing.T, output *AWSSecurityGroupOutput) {
				// Should have no permissions (hostname skipped)
				assert.Empty(t, output.IpPermissions)
				assert.Empty(t, output.IpPermissionsEgress)
			},
		},
		{
			name: "handles multiple values in single permission",
			entries: []types.NetworkEntry{
				{
					Direction:   "inbound",
					Service:     "multi",
					Region:      "global",
					Type:        "ipv4",
					Values:      []string{"192.168.1.0/24", "10.0.0.0/8", "172.16.0.0/12"},
					Protocol:    "tcp",
					Ports:       []int{443},
					Description: "Multi-value",
				},
			},
			checkContent: func(t *testing.T, output *AWSSecurityGroupOutput) {
				perm := output.IpPermissions[0]
				assert.Len(t, perm.IpRanges, 3)
				assert.Equal(t, "192.168.1.0/24", perm.IpRanges[0].CidrIp)
				assert.Equal(t, "10.0.0.0/8", perm.IpRanges[1].CidrIp)
				assert.Equal(t, "172.16.0.0/12", perm.IpRanges[2].CidrIp)
			},
		},
		{
			name: "handles port range",
			entries: []types.NetworkEntry{
				{
					Direction:   "inbound",
					Service:     "multi_port",
					Region:      "us",
					Type:        "ipv4",
					Values:      []string{"192.168.1.0/24"},
					Protocol:    "tcp",
					Ports:       []int{443, 8443, 80, 8080},
					Description: "Multi-port service",
				},
			},
			checkContent: func(t *testing.T, output *AWSSecurityGroupOutput) {
				perm := output.IpPermissions[0]
				// Port range should be min to max
				assert.Equal(t, 80, perm.FromPort)
				assert.Equal(t, 8443, perm.ToPort)
			},
		},
		{
			name: "handles UDP protocol",
			entries: []types.NetworkEntry{
				{
					Direction:   "inbound",
					Service:     "dns",
					Region:      "us",
					Type:        "ipv4",
					Values:      []string{"192.168.1.0/24"},
					Protocol:    "udp",
					Ports:       []int{53},
					Description: "DNS service",
				},
			},
			checkContent: func(t *testing.T, output *AWSSecurityGroupOutput) {
				perm := output.IpPermissions[0]
				assert.Equal(t, "udp", perm.IpProtocol)
			},
		},
		{
			name: "handles 'both' protocol",
			entries: []types.NetworkEntry{
				{
					Direction:   "inbound",
					Service:     "all",
					Region:      "us",
					Type:        "ipv4",
					Values:      []string{"192.168.1.0/24"},
					Protocol:    "both",
					Ports:       []int{443},
					Description: "All protocols",
				},
			},
			checkContent: func(t *testing.T, output *AWSSecurityGroupOutput) {
				perm := output.IpPermissions[0]
				assert.Equal(t, "-1", perm.IpProtocol)
			},
		},
		{
			name: "includes redundancy in description",
			entries: []types.NetworkEntry{
				{
					Direction:   "inbound",
					Service:     "service",
					Region:      "us",
					Type:        "ipv4",
					Values:      []string{"192.168.1.0/24"},
					Protocol:    "tcp",
					Ports:       []int{443},
					Description: "Test",
					Redundancy:  "primary",
				},
			},
			checkContent: func(t *testing.T, output *AWSSecurityGroupOutput) {
				perm := output.IpPermissions[0]
				assert.Contains(t, perm.IpRanges[0].Description, "[primary]")
			},
		},
		{
			name:    "handles empty entries",
			entries: []types.NetworkEntry{},
			checkContent: func(t *testing.T, output *AWSSecurityGroupOutput) {
				assert.Empty(t, output.IpPermissions)
				assert.Empty(t, output.IpPermissionsEgress)
				assert.Len(t, output.Tags, 2) // Default tags still present
			},
		},
		{
			name: "groups entries by protocol and ports",
			entries: []types.NetworkEntry{
				{
					Direction:   "inbound",
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
					Ports:       []int{443},
					Description: "Service 2",
				},
			},
			checkContent: func(t *testing.T, output *AWSSecurityGroupOutput) {
				// Both entries should be grouped into one permission (same protocol and ports)
				assert.Len(t, output.IpPermissions, 1)
				perm := output.IpPermissions[0]
				assert.Len(t, perm.IpRanges, 2)
			},
		},
		{
			name: "separates entries with different protocols",
			entries: []types.NetworkEntry{
				{
					Direction:   "inbound",
					Service:     "service1",
					Region:      "us",
					Type:        "ipv4",
					Values:      []string{"192.168.1.0/24"},
					Protocol:    "tcp",
					Ports:       []int{443},
					Description: "TCP service",
				},
				{
					Direction:   "inbound",
					Service:     "service2",
					Region:      "us",
					Type:        "ipv4",
					Values:      []string{"10.0.0.0/8"},
					Protocol:    "udp",
					Ports:       []int{53},
					Description: "UDP service",
				},
			},
			checkContent: func(t *testing.T, output *AWSSecurityGroupOutput) {
				// Different protocols should create separate permissions
				assert.Len(t, output.IpPermissions, 2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := &AWSSecurityGroupConverter{}
			result, err := converter.Convert(tt.entries)
			require.NoError(t, err)
			require.NotNil(t, result)

			// Parse the JSON to verify structure
			var output AWSSecurityGroupOutput
			err = json.Unmarshal(result, &output)
			require.NoError(t, err)

			// Verify default structure
			assert.NotEmpty(t, output.Description)
			assert.NotEmpty(t, output.GroupName)
			assert.NotNil(t, output.Tags)

			if tt.checkContent != nil {
				tt.checkContent(t, &output)
			}
		})
	}
}

func TestAWSSecurityGroupConverter_Name(t *testing.T) {
	converter := &AWSSecurityGroupConverter{}
	assert.Equal(t, "AWS Security Group", converter.Name())
}

func TestAWSSecurityGroupConverter_FileExtension(t *testing.T) {
	converter := &AWSSecurityGroupConverter{}
	assert.Equal(t, "json", converter.FileExtension())
}

func TestGetAWSProtocol(t *testing.T) {
	converter := &AWSSecurityGroupConverter{}

	tests := []struct {
		name     string
		protocol string
		expected string
	}{
		{
			name:     "tcp protocol",
			protocol: "tcp",
			expected: "tcp",
		},
		{
			name:     "udp protocol",
			protocol: "udp",
			expected: "udp",
		},
		{
			name:     "both protocol",
			protocol: "both",
			expected: "-1",
		},
		{
			name:     "unknown protocol",
			protocol: "unknown",
			expected: "-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.getAWSProtocol(tt.protocol)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetPortRange(t *testing.T) {
	converter := &AWSSecurityGroupConverter{}

	tests := []struct {
		name         string
		ports        []int
		expectedFrom int
		expectedTo   int
	}{
		{
			name:         "single port",
			ports:        []int{443},
			expectedFrom: 443,
			expectedTo:   443,
		},
		{
			name:         "two ports in order",
			ports:        []int{80, 443},
			expectedFrom: 80,
			expectedTo:   443,
		},
		{
			name:         "two ports out of order",
			ports:        []int{443, 80},
			expectedFrom: 80,
			expectedTo:   443,
		},
		{
			name:         "multiple ports",
			ports:        []int{443, 8443, 80, 8080, 53},
			expectedFrom: 53,
			expectedTo:   8443,
		},
		{
			name:         "empty ports",
			ports:        []int{},
			expectedFrom: 0,
			expectedTo:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			from, to := converter.getPortRange(tt.ports)
			assert.Equal(t, tt.expectedFrom, from)
			assert.Equal(t, tt.expectedTo, to)
		})
	}
}

func TestAWSSecurityGroupConverter_Integration(t *testing.T) {
	entries := []types.NetworkEntry{
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
		},
		{
			Direction:   "outbound",
			Service:     "platform_api",
			Region:      "us",
			Type:        "ipv4",
			Values:      []string{"172.16.0.0/12"},
			Protocol:    "tcp",
			Ports:       []int{443, 8443},
			Description: "Platform API",
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
		{
			Direction:   "inbound",
			Service:     "ipv6_service",
			Region:      "eu",
			Type:        "ipv6",
			Values:      []string{"2001:db8::/32"},
			Protocol:    "tcp",
			Ports:       []int{80},
			Description: "IPv6 service",
		},
	}

	converter := &AWSSecurityGroupConverter{}
	result, err := converter.Convert(entries)
	require.NoError(t, err)

	// Parse it back
	var output AWSSecurityGroupOutput
	err = json.Unmarshal(result, &output)
	require.NoError(t, err)

	// Verify structure
	assert.Equal(t, "Delinea Platform Network Requirements", output.Description)
	assert.Equal(t, "delinea-platform-sg", output.GroupName)

	// Check tags
	assert.Len(t, output.Tags, 2)
	assert.Equal(t, "Name", output.Tags[0].Key)

	// Should have 2 inbound permissions (webhook + ipv6)
	assert.Len(t, output.IpPermissions, 2)

	// Find webhook permission
	var webhookPerm *AWSIpPermission
	for i := range output.IpPermissions {
		if len(output.IpPermissions[i].IpRanges) > 0 &&
			output.IpPermissions[i].FromPort == 443 {
			webhookPerm = &output.IpPermissions[i]
			break
		}
	}
	require.NotNil(t, webhookPerm)
	assert.Len(t, webhookPerm.IpRanges, 2)
	assert.Contains(t, webhookPerm.IpRanges[0].Description, "[primary]")

	// Should have 1 egress permission (platform_api, telemetry skipped)
	assert.Len(t, output.IpPermissionsEgress, 1)
	egressPerm := output.IpPermissionsEgress[0]
	assert.Equal(t, "tcp", egressPerm.IpProtocol)
	assert.Equal(t, 443, egressPerm.FromPort)
	assert.Equal(t, 8443, egressPerm.ToPort)
	assert.Len(t, egressPerm.IpRanges, 1)
}
