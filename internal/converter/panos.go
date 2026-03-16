package converter

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/DelineaXPM/delinea-netconfig/pkg/types"
)

// PANOSConverter converts network entries to PAN-OS XML format
type PANOSConverter struct{}

// PAN-OS XML structures
type PANOSConfig struct {
	XMLName xml.Name     `xml:"config"`
	Devices PANOSDevices `xml:"devices"`
}

type PANOSDevices struct {
	Entry PANOSDeviceEntry `xml:"entry"`
}

type PANOSDeviceEntry struct {
	Name string    `xml:"name,attr"`
	Vsys PANOSVsys `xml:"vsys"`
}

type PANOSVsys struct {
	Entry PANOSVsysEntry `xml:"entry"`
}

type PANOSVsysEntry struct {
	Name     string           `xml:"name,attr"`
	Address  *PANOSAddresses  `xml:"address,omitempty"`
	Service  *PANOSServices   `xml:"service,omitempty"`
	Rulebase *PANOSRulebase   `xml:"rulebase,omitempty"`
}

type PANOSAddresses struct {
	Entries []PANOSAddressEntry `xml:"entry"`
}

type PANOSAddressEntry struct {
	Name        string  `xml:"name,attr"`
	IPNetmask   *string `xml:"ip-netmask,omitempty"`
	FQDN        *string `xml:"fqdn,omitempty"`
	Description *string `xml:"description,omitempty"`
}

type PANOSServices struct {
	Entries []PANOSServiceEntry `xml:"entry"`
}

type PANOSServiceEntry struct {
	Name        string                `xml:"name,attr"`
	Protocol    *PANOSServiceProtocol `xml:"protocol,omitempty"`
	Description *string               `xml:"description,omitempty"`
}

type PANOSServiceProtocol struct {
	TCP *PANOSServicePorts `xml:"tcp,omitempty"`
	UDP *PANOSServicePorts `xml:"udp,omitempty"`
}

type PANOSServicePorts struct {
	Port string `xml:"port"`
}

type PANOSRulebase struct {
	Security PANOSSecurity `xml:"security"`
}

type PANOSSecurity struct {
	Rules PANOSRules `xml:"rules"`
}

type PANOSRules struct {
	Entries []PANOSRuleEntry `xml:"entry"`
}

type PANOSRuleEntry struct {
	Name             string            `xml:"name,attr"`
	From             PANOSZones        `xml:"from"`
	To               PANOSZones        `xml:"to"`
	Source           PANOSMembers      `xml:"source"`
	Destination      PANOSMembers      `xml:"destination"`
	Service          PANOSMembers      `xml:"service"`
	Application      PANOSMembers      `xml:"application"`
	Action           string            `xml:"action"`
	Description      *string           `xml:"description,omitempty"`
}

type PANOSZones struct {
	Members []string `xml:"member"`
}

type PANOSMembers struct {
	Members []string `xml:"member"`
}

// Convert converts network entries to PAN-OS XML format
func (c *PANOSConverter) Convert(entries []types.NetworkEntry) ([]byte, error) {
	config := PANOSConfig{
		Devices: PANOSDevices{
			Entry: PANOSDeviceEntry{
				Name: "localhost.localdomain",
				Vsys: PANOSVsys{
					Entry: PANOSVsysEntry{
						Name: "vsys1",
					},
				},
			},
		},
	}

	vsysEntry := &config.Devices.Entry.Vsys.Entry

	// Create address objects
	addresses := []PANOSAddressEntry{}
	services := []PANOSServiceEntry{}
	rules := []PANOSRuleEntry{}

	// Track created objects to avoid duplicates
	addressNames := make(map[string]bool)
	serviceNames := make(map[string]bool)
	ruleIndex := 1

	for _, entry := range entries {
		// Create address objects for IP addresses
		for _, value := range entry.Values {
			if strings.Contains(entry.Type, "hostname") {
				// Create FQDN address object
				addressName := c.sanitizeName(fmt.Sprintf("fqdn-%s", value))
				if !addressNames[addressName] {
					fqdn := value
					desc := entry.Description
					addresses = append(addresses, PANOSAddressEntry{
						Name:        addressName,
						FQDN:        &fqdn,
						Description: &desc,
					})
					addressNames[addressName] = true
				}
			} else if strings.Contains(entry.Type, "ipv4") || strings.Contains(entry.Type, "ipv6") {
				// Create IP address object
				addressName := c.sanitizeName(fmt.Sprintf("ip-%s", strings.ReplaceAll(value, "/", "-")))
				if !addressNames[addressName] {
					ipNetmask := value
					desc := entry.Description
					addresses = append(addresses, PANOSAddressEntry{
						Name:        addressName,
						IPNetmask:   &ipNetmask,
						Description: &desc,
					})
					addressNames[addressName] = true
				}
			}
		}

		// Create service objects for ports
		if len(entry.Ports) > 0 {
			serviceName := c.sanitizeName(fmt.Sprintf("svc-%s-%s-%s", entry.Service, entry.Protocol, c.formatPortList(entry.Ports)))
			if !serviceNames[serviceName] {
				desc := entry.Description
				svcEntry := PANOSServiceEntry{
					Name:        serviceName,
					Description: &desc,
					Protocol:    &PANOSServiceProtocol{},
				}

				portStr := c.formatPortList(entry.Ports)

				switch entry.Protocol {
				case "tcp":
					svcEntry.Protocol.TCP = &PANOSServicePorts{Port: portStr}
				case "udp":
					svcEntry.Protocol.UDP = &PANOSServicePorts{Port: portStr}
				case "both":
					// Create two separate service objects for TCP and UDP
					tcpServiceName := serviceName + "-tcp"
					if !serviceNames[tcpServiceName] {
						services = append(services, PANOSServiceEntry{
							Name:        tcpServiceName,
							Description: &desc,
							Protocol: &PANOSServiceProtocol{
								TCP: &PANOSServicePorts{Port: portStr},
							},
						})
						serviceNames[tcpServiceName] = true
					}
					udpServiceName := serviceName + "-udp"
					if !serviceNames[udpServiceName] {
						services = append(services, PANOSServiceEntry{
							Name:        udpServiceName,
							Description: &desc,
							Protocol: &PANOSServiceProtocol{
								UDP: &PANOSServicePorts{Port: portStr},
							},
						})
						serviceNames[udpServiceName] = true
					}
					continue // Skip adding the "both" service entry
				default:
					// Default to TCP
					svcEntry.Protocol.TCP = &PANOSServicePorts{Port: portStr}
				}

				services = append(services, svcEntry)
				serviceNames[serviceName] = true
			}
		}

		// Create security rules
		for _, value := range entry.Values {
			ruleName := c.sanitizeName(fmt.Sprintf("rule-%d-%s-%s", ruleIndex, entry.Service, entry.Direction))
			ruleIndex++

			rule := PANOSRuleEntry{
				Name:        ruleName,
				Action:      "allow",
				Application: PANOSMembers{Members: []string{"any"}},
			}

			desc := fmt.Sprintf("%s - %s", entry.Service, entry.Description)
			rule.Description = &desc

			// Set zones based on direction
			if entry.Direction == "outbound" {
				rule.From = PANOSZones{Members: []string{"trust"}}
				rule.To = PANOSZones{Members: []string{"untrust"}}
				rule.Source = PANOSMembers{Members: []string{"any"}}

				// Set destination
				if strings.Contains(entry.Type, "hostname") {
					addressName := c.sanitizeName(fmt.Sprintf("fqdn-%s", value))
					rule.Destination = PANOSMembers{Members: []string{addressName}}
				} else {
					addressName := c.sanitizeName(fmt.Sprintf("ip-%s", strings.ReplaceAll(value, "/", "-")))
					rule.Destination = PANOSMembers{Members: []string{addressName}}
				}
			} else {
				// Inbound
				rule.From = PANOSZones{Members: []string{"untrust"}}
				rule.To = PANOSZones{Members: []string{"trust"}}
				rule.Destination = PANOSMembers{Members: []string{"any"}}

				// Set source
				if strings.Contains(entry.Type, "hostname") {
					addressName := c.sanitizeName(fmt.Sprintf("fqdn-%s", value))
					rule.Source = PANOSMembers{Members: []string{addressName}}
				} else {
					addressName := c.sanitizeName(fmt.Sprintf("ip-%s", strings.ReplaceAll(value, "/", "-")))
					rule.Source = PANOSMembers{Members: []string{addressName}}
				}
			}

			// Set service
			if len(entry.Ports) > 0 {
				serviceName := c.sanitizeName(fmt.Sprintf("svc-%s-%s-%s", entry.Service, entry.Protocol, c.formatPortList(entry.Ports)))
				if entry.Protocol == "both" {
					// Reference both TCP and UDP services
					rule.Service = PANOSMembers{Members: []string{serviceName + "-tcp", serviceName + "-udp"}}
				} else {
					rule.Service = PANOSMembers{Members: []string{serviceName}}
				}
			} else {
				rule.Service = PANOSMembers{Members: []string{"any"}}
			}

			rules = append(rules, rule)
		}
	}

	// Add objects to config if any were created
	if len(addresses) > 0 {
		vsysEntry.Address = &PANOSAddresses{Entries: addresses}
	}
	if len(services) > 0 {
		vsysEntry.Service = &PANOSServices{Entries: services}
	}
	if len(rules) > 0 {
		vsysEntry.Rulebase = &PANOSRulebase{
			Security: PANOSSecurity{
				Rules: PANOSRules{Entries: rules},
			},
		}
	}

	// Marshal to XML with indentation
	output, err := xml.MarshalIndent(config, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal PAN-OS XML: %w", err)
	}

	// Add XML declaration
	xmlDeclaration := []byte(xml.Header)
	return append(xmlDeclaration, output...), nil
}

// sanitizeName sanitizes names for PAN-OS (alphanumeric, hyphens, underscores)
func (c *PANOSConverter) sanitizeName(name string) string {
	// Replace invalid characters with hyphens
	replacer := strings.NewReplacer(
		" ", "-",
		".", "-",
		":", "-",
		"/", "-",
		"<", "",
		">", "",
	)
	sanitized := replacer.Replace(name)

	// Ensure name doesn't exceed PAN-OS limit (63 characters)
	if len(sanitized) > 63 {
		sanitized = sanitized[:63]
	}

	return sanitized
}

// formatPortList formats a list of ports for PAN-OS
func (c *PANOSConverter) formatPortList(ports []int) string {
	if len(ports) == 0 {
		return ""
	}

	if len(ports) == 1 {
		return fmt.Sprintf("%d", ports[0])
	}

	// Check if ports are consecutive
	isConsecutive := true
	for i := 1; i < len(ports); i++ {
		if ports[i] != ports[i-1]+1 {
			isConsecutive = false
			break
		}
	}

	if isConsecutive && len(ports) > 1 {
		// Use range notation
		return fmt.Sprintf("%d-%d", ports[0], ports[len(ports)-1])
	}

	// Use comma-separated list
	portStrs := make([]string, len(ports))
	for i, port := range ports {
		portStrs[i] = fmt.Sprintf("%d", port)
	}
	return strings.Join(portStrs, ",")
}

// Name returns the name of the converter
func (c *PANOSConverter) Name() string {
	return "PAN-OS XML"
}

// FileExtension returns the file extension for PAN-OS XML files
func (c *PANOSConverter) FileExtension() string {
	return "xml"
}
