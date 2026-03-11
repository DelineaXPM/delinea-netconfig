package converter

import (
	"fmt"

	"github.com/DelineaXPM/delinea-platform/delinea-netconfig/pkg/types"
	"github.com/goccy/go-yaml"
)

// AnsibleConverter converts network entries to Ansible variables format
type AnsibleConverter struct{}

// AnsibleOutput represents the structure of the Ansible output
type AnsibleOutput struct {
	Comment          string                   `yaml:"_comment"`
	DelineaFirewall  AnsibleFirewallRules     `yaml:"delinea_firewall_rules"`
}

// AnsibleFirewallRules represents firewall rules organized by direction
type AnsibleFirewallRules struct {
	Outbound []AnsibleRule `yaml:"outbound"`
	Inbound  []AnsibleRule `yaml:"inbound"`
}

// AnsibleRule represents a single firewall rule for Ansible
type AnsibleRule struct {
	Name        string   `yaml:"name"`
	Service     string   `yaml:"service"`
	Region      string   `yaml:"region"`
	Type        string   `yaml:"type"`
	Destinations []string `yaml:"destinations,omitempty"`
	Sources     []string `yaml:"sources,omitempty"`
	Protocol    string   `yaml:"protocol"`
	Ports       []int    `yaml:"ports,omitempty"`
	Description string   `yaml:"description"`
	Redundancy  string   `yaml:"redundancy,omitempty"`
	Tags        []string `yaml:"tags,omitempty"`
}

// Convert converts network entries to Ansible variables format
func (c *AnsibleConverter) Convert(entries []types.NetworkEntry) ([]byte, error) {
	output := AnsibleOutput{
		Comment: "Delinea Network Requirements - Ansible Variables",
		DelineaFirewall: AnsibleFirewallRules{
			Outbound: []AnsibleRule{},
			Inbound:  []AnsibleRule{},
		},
	}

	for _, entry := range entries {
		rule := AnsibleRule{
			Name:        fmt.Sprintf("%s_%s_%s", entry.Service, entry.Region, entry.Type),
			Service:     entry.Service,
			Region:      entry.Region,
			Type:        entry.Type,
			Protocol:    entry.Protocol,
			Ports:       entry.Ports,
			Description: entry.Description,
			Tags:        entry.Tags,
		}

		if entry.Redundancy != "" {
			rule.Redundancy = entry.Redundancy
		}

		// For outbound rules, values go to destinations
		// For inbound rules, values go to sources
		if entry.Direction == "outbound" {
			rule.Destinations = entry.Values
			output.DelineaFirewall.Outbound = append(output.DelineaFirewall.Outbound, rule)
		} else {
			rule.Sources = entry.Values
			output.DelineaFirewall.Inbound = append(output.DelineaFirewall.Inbound, rule)
		}
	}

	// Marshal to YAML
	data, err := yaml.Marshal(&output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Ansible YAML: %w", err)
	}

	return data, nil
}

// Name returns the name of the converter
func (c *AnsibleConverter) Name() string {
	return "Ansible"
}

// FileExtension returns the file extension for Ansible files
func (c *AnsibleConverter) FileExtension() string {
	return "yml"
}
