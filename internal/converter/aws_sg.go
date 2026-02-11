package converter

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/DelineaXPM/delinea-platform/delinea-netconfig/pkg/types"
)

// AWSSecurityGroupConverter converts network entries to AWS Security Group JSON format
type AWSSecurityGroupConverter struct{}

// AWSSecurityGroupOutput represents the AWS Security Group structure
type AWSSecurityGroupOutput struct {
	Description string               `json:"Description"`
	GroupName   string               `json:"GroupName"`
	VpcId       string               `json:"VpcId,omitempty"`
	IpPermissions []AWSIpPermission  `json:"IpPermissions"`
	IpPermissionsEgress []AWSIpPermission `json:"IpPermissionsEgress"`
	Tags        []AWSTag             `json:"Tags,omitempty"`
}

// AWSIpPermission represents an AWS Security Group rule
type AWSIpPermission struct {
	IpProtocol string            `json:"IpProtocol"`
	FromPort   int               `json:"FromPort,omitempty"`
	ToPort     int               `json:"ToPort,omitempty"`
	IpRanges   []AWSIpRange      `json:"IpRanges,omitempty"`
	Ipv6Ranges []AWSIpv6Range    `json:"Ipv6Ranges,omitempty"`
}

// AWSIpRange represents an IPv4 CIDR range with description
type AWSIpRange struct {
	CidrIp      string `json:"CidrIp"`
	Description string `json:"Description,omitempty"`
}

// AWSIpv6Range represents an IPv6 CIDR range with description
type AWSIpv6Range struct {
	CidrIpv6    string `json:"CidrIpv6"`
	Description string `json:"Description,omitempty"`
}

// AWSTag represents an AWS resource tag
type AWSTag struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

// Convert converts network entries to AWS Security Group JSON format
func (c *AWSSecurityGroupConverter) Convert(entries []types.NetworkEntry) ([]byte, error) {
	output := AWSSecurityGroupOutput{
		Description: "Delinea Platform Network Requirements",
		GroupName:   "delinea-platform-sg",
		IpPermissions: []AWSIpPermission{},
		IpPermissionsEgress: []AWSIpPermission{},
		Tags: []AWSTag{
			{Key: "Name", Value: "delinea-platform-sg"},
			{Key: "ManagedBy", Value: "delinea-netconfig"},
		},
	}

	// Group entries by direction, protocol, and ports for consolidation
	inboundRules := c.groupRules(entries, "inbound")
	outboundRules := c.groupRules(entries, "outbound")

	// Convert to AWS format
	output.IpPermissions = c.convertToAWSPermissions(inboundRules)
	output.IpPermissionsEgress = c.convertToAWSPermissions(outboundRules)

	// Marshal to JSON
	data, err := json.MarshalIndent(&output, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal AWS Security Group JSON: %w", err)
	}

	return data, nil
}

// groupRules groups entries by protocol and ports for consolidation
func (c *AWSSecurityGroupConverter) groupRules(entries []types.NetworkEntry, direction string) map[string][]types.NetworkEntry {
	groups := make(map[string][]types.NetworkEntry)

	for _, entry := range entries {
		if entry.Direction != direction {
			continue
		}

		// Skip hostname entries (AWS SG only supports IP addresses)
		if strings.Contains(entry.Type, "hostname") {
			continue
		}

		// Create a key based on protocol and ports
		key := fmt.Sprintf("%s:%v", entry.Protocol, entry.Ports)
		groups[key] = append(groups[key], entry)
	}

	return groups
}

// convertToAWSPermissions converts grouped rules to AWS IP permissions
func (c *AWSSecurityGroupConverter) convertToAWSPermissions(groups map[string][]types.NetworkEntry) []AWSIpPermission {
	permissions := []AWSIpPermission{}

	for _, entries := range groups {
		if len(entries) == 0 {
			continue
		}

		// Use the first entry to get protocol and ports
		first := entries[0]

		// Determine AWS protocol string
		protocol := c.getAWSProtocol(first.Protocol)

		// Get port range
		fromPort, toPort := c.getPortRange(first.Ports)

		permission := AWSIpPermission{
			IpProtocol: protocol,
			FromPort:   fromPort,
			ToPort:     toPort,
			IpRanges:   []AWSIpRange{},
			Ipv6Ranges: []AWSIpv6Range{},
		}

		// Add all IP ranges from all entries in this group
		for _, entry := range entries {
			description := fmt.Sprintf("%s - %s (%s)", entry.Service, entry.Description, entry.Region)
			if entry.Redundancy != "" {
				description = fmt.Sprintf("%s [%s]", description, entry.Redundancy)
			}

			for _, value := range entry.Values {
				if entry.Type == "ipv4" {
					permission.IpRanges = append(permission.IpRanges, AWSIpRange{
						CidrIp:      value,
						Description: description,
					})
				} else if entry.Type == "ipv6" {
					permission.Ipv6Ranges = append(permission.Ipv6Ranges, AWSIpv6Range{
						CidrIpv6:    value,
						Description: description,
					})
				}
			}
		}

		// Only add if we have IP ranges
		if len(permission.IpRanges) > 0 || len(permission.Ipv6Ranges) > 0 {
			permissions = append(permissions, permission)
		}
	}

	return permissions
}

// getAWSProtocol converts protocol string to AWS format
func (c *AWSSecurityGroupConverter) getAWSProtocol(protocol string) string {
	switch protocol {
	case "tcp":
		return "tcp"
	case "udp":
		return "udp"
	case "both":
		return "-1" // All protocols
	default:
		return "-1"
	}
}

// getPortRange returns the from and to port range
func (c *AWSSecurityGroupConverter) getPortRange(ports []int) (int, int) {
	if len(ports) == 0 {
		return 0, 0
	}

	if len(ports) == 1 {
		return ports[0], ports[0]
	}

	// Find min and max
	min, max := ports[0], ports[0]
	for _, port := range ports[1:] {
		if port < min {
			min = port
		}
		if port > max {
			max = port
		}
	}

	return min, max
}

// Name returns the name of the converter
func (c *AWSSecurityGroupConverter) Name() string {
	return "AWS Security Group"
}

// FileExtension returns the file extension for AWS SG files
func (c *AWSSecurityGroupConverter) FileExtension() string {
	return "json"
}
