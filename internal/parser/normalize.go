package parser

import (
	"sort"

	"github.com/DelineaXPM/delinea-platform/delinea-netconfig/pkg/types"
)

// Normalize converts NetworkRequirements into a flat list of NetworkEntry structs
func Normalize(nr *types.NetworkRequirements) []types.NetworkEntry {
	var entries []types.NetworkEntry

	// Process outbound services
	for serviceName, service := range nr.Outbound {
		entries = append(entries, normalizeService("outbound", serviceName, service)...)
	}

	// Process inbound services
	for serviceName, service := range nr.Inbound {
		entries = append(entries, normalizeService("inbound", serviceName, service)...)
	}

	// Sort entries for consistent output
	// Sort by: direction (outbound first) -> service -> region -> type
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].Direction != entries[j].Direction {
			// Outbound before inbound
			return entries[i].Direction > entries[j].Direction
		}
		if entries[i].Service != entries[j].Service {
			return entries[i].Service < entries[j].Service
		}
		if entries[i].Region != entries[j].Region {
			return entries[i].Region < entries[j].Region
		}
		return entries[i].Type < entries[j].Type
	})

	return entries
}

// normalizeService converts a Service into NetworkEntry structs
func normalizeService(direction, serviceName string, service types.Service) []types.NetworkEntry {
	var entries []types.NetworkEntry

	// Get ports for this service
	tcpPorts, udpPorts := getPorts(service)

	// Process each region
	for regionCode, regionData := range service.Regions {
		// Process IPv4 addresses
		if len(regionData.IPv4) > 0 {
			entries = append(entries, types.NetworkEntry{
				Direction:   direction,
				Service:     serviceName,
				Region:      regionCode,
				Type:        "ipv4",
				Values:      regionData.IPv4,
				Protocol:    getProtocol(tcpPorts, udpPorts),
				Ports:       mergePorts(tcpPorts, udpPorts),
				Description: service.Description,
				Redundancy:  "",
			})
		}

		// Process IPv4 primary addresses
		if len(regionData.IPv4Primary) > 0 {
			entries = append(entries, types.NetworkEntry{
				Direction:   direction,
				Service:     serviceName,
				Region:      regionCode,
				Type:        "ipv4",
				Values:      regionData.IPv4Primary,
				Protocol:    getProtocol(tcpPorts, udpPorts),
				Ports:       mergePorts(tcpPorts, udpPorts),
				Description: service.Description,
				Redundancy:  "primary",
			})
		}

		// Process IPv4 DR addresses
		if len(regionData.IPv4DR) > 0 {
			entries = append(entries, types.NetworkEntry{
				Direction:   direction,
				Service:     serviceName,
				Region:      regionCode,
				Type:        "ipv4",
				Values:      regionData.IPv4DR,
				Protocol:    getProtocol(tcpPorts, udpPorts),
				Ports:       mergePorts(tcpPorts, udpPorts),
				Description: service.Description,
				Redundancy:  "dr",
			})
		}

		// Process IPv6 addresses
		if len(regionData.IPv6) > 0 {
			entries = append(entries, types.NetworkEntry{
				Direction:   direction,
				Service:     serviceName,
				Region:      regionCode,
				Type:        "ipv6",
				Values:      regionData.IPv6,
				Protocol:    getProtocol(tcpPorts, udpPorts),
				Ports:       mergePorts(tcpPorts, udpPorts),
				Description: service.Description,
				Redundancy:  "",
			})
		}

		// Process hostnames
		if len(regionData.Hostnames) > 0 {
			entries = append(entries, types.NetworkEntry{
				Direction:   direction,
				Service:     serviceName,
				Region:      regionCode,
				Type:        "hostname",
				Values:      regionData.Hostnames,
				Protocol:    getProtocol(tcpPorts, udpPorts),
				Ports:       mergePorts(tcpPorts, udpPorts),
				Description: service.Description,
				Redundancy:  "",
			})
		}

		// Process self-signed hostnames
		if len(regionData.HostnamesSelfSigned) > 0 {
			entries = append(entries, types.NetworkEntry{
				Direction:   direction,
				Service:     serviceName,
				Region:      regionCode,
				Type:        "hostname_self_signed",
				Values:      regionData.HostnamesSelfSigned,
				Protocol:    getProtocol(tcpPorts, udpPorts),
				Ports:       mergePorts(tcpPorts, udpPorts),
				Description: service.Description,
				Redundancy:  "",
				Tags:        []string{"self-signed"},
			})
		}

		// Process CA-signed hostnames
		if len(regionData.HostnamesCASigned) > 0 {
			entries = append(entries, types.NetworkEntry{
				Direction:   direction,
				Service:     serviceName,
				Region:      regionCode,
				Type:        "hostname_ca_signed",
				Values:      regionData.HostnamesCASigned,
				Protocol:    getProtocol(tcpPorts, udpPorts),
				Ports:       mergePorts(tcpPorts, udpPorts),
				Description: service.Description,
				Redundancy:  "",
				Tags:        []string{"ca-signed"},
			})
		}

		// Process AWS SES
		if len(regionData.AWSSES) > 0 {
			entries = append(entries, types.NetworkEntry{
				Direction:   direction,
				Service:     serviceName,
				Region:      regionCode,
				Type:        "ipv4",
				Values:      regionData.AWSSES,
				Protocol:    getProtocol(tcpPorts, udpPorts),
				Ports:       mergePorts(tcpPorts, udpPorts),
				Description: service.Description,
				Redundancy:  "",
				Tags:        []string{"aws-ses"},
			})
		}
	}

	return entries
}

// getPorts extracts TCP and UDP ports from a service.
// Supports both old format (tcp_ports/udp_ports/nested) and new format (flat ports + protocol).
func getPorts(service types.Service) (tcpPorts, udpPorts []int) {
	// New format: flat ports with protocol field
	if len(service.FlatPorts) > 0 {
		switch service.Protocol {
		case "udp":
			return nil, service.FlatPorts
		case "tcp", "":
			return service.FlatPorts, nil
		default:
			// Unknown protocol, treat as TCP
			return service.FlatPorts, nil
		}
	}

	// Old format: direct tcp_ports/udp_ports
	tcpPorts = service.TCPPorts
	udpPorts = service.UDPPorts

	// Old format: nested ports (for services like AD Connector)
	if service.NestedPorts != nil {
		if service.NestedPorts.External != nil {
			if len(service.NestedPorts.External.TCPPorts) > 0 {
				tcpPorts = append(tcpPorts, service.NestedPorts.External.TCPPorts...)
			}
			if len(service.NestedPorts.External.UDPPorts) > 0 {
				udpPorts = append(udpPorts, service.NestedPorts.External.UDPPorts...)
			}
		}
	}

	return tcpPorts, udpPorts
}

// getProtocol determines the protocol string based on available ports
func getProtocol(tcpPorts, udpPorts []int) string {
	hasTCP := len(tcpPorts) > 0
	hasUDP := len(udpPorts) > 0

	if hasTCP && hasUDP {
		return "both"
	} else if hasTCP {
		return "tcp"
	} else if hasUDP {
		return "udp"
	}
	return ""
}

// mergePorts merges TCP and UDP ports into a single slice
func mergePorts(tcpPorts, udpPorts []int) []int {
	// For now, just combine them
	// In the future, we might want to keep them separate or add metadata
	ports := make([]int, 0, len(tcpPorts)+len(udpPorts))
	ports = append(ports, tcpPorts...)
	ports = append(ports, udpPorts...)
	return ports
}
