package validator

import (
	"fmt"
	"net"
	"strings"

	"github.com/DelineaXPM/delinea-platform/delinea-netconfig/internal/parser"
)

// ValidationResult contains the results of validating a network requirements JSON
type ValidationResult struct {
	Version          string
	IPv4Count        int
	IPv6Count        int
	HostnameCount    int
	TotalServices    int
	OutboundServices int
	InboundServices  int
	RegionCount      int
	Warnings         []string
}

// Validate validates the network requirements JSON and returns statistics
func Validate(data []byte) (*ValidationResult, error) {
	// Parse the JSON
	networkReqs, err := parser.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Initialize result
	result := &ValidationResult{
		Version:          networkReqs.Version,
		OutboundServices: len(networkReqs.Outbound),
		InboundServices:  len(networkReqs.Inbound),
		TotalServices:    len(networkReqs.Outbound) + len(networkReqs.Inbound),
		RegionCount:      len(networkReqs.RegionCodes),
		Warnings:         []string{},
	}

	// Normalize to entries for detailed validation
	entries := parser.Normalize(networkReqs)

	// Validate each entry
	for _, entry := range entries {
		for _, value := range entry.Values {
			switch entry.Type {
			case "ipv4":
				if err := validateIPv4(value); err != nil {
					result.Warnings = append(result.Warnings, fmt.Sprintf("Invalid IPv4 %s in service %s: %v", value, entry.Service, err))
				} else {
					result.IPv4Count++
				}
			case "ipv6":
				if err := validateIPv6(value); err != nil {
					result.Warnings = append(result.Warnings, fmt.Sprintf("Invalid IPv6 %s in service %s: %v", value, entry.Service, err))
				} else {
					result.IPv6Count++
				}
			case "hostname", "hostname_self_signed", "hostname_ca_signed":
				if err := validateHostname(value); err != nil {
					result.Warnings = append(result.Warnings, fmt.Sprintf("Invalid hostname %s in service %s: %v", value, entry.Service, err))
				} else {
					result.HostnameCount++
				}
			}
		}

		// Validate ports
		for _, port := range entry.Ports {
			if port < 1 || port > 65535 {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Invalid port %d in service %s", port, entry.Service))
			}
		}
	}

	return result, nil
}

// validateIPv4 validates an IPv4 address or CIDR range
func validateIPv4(ipstr string) error {
	// Check if it's a CIDR
	if strings.Contains(ipstr, "/") {
		_, _, err := net.ParseCIDR(ipstr)
		if err != nil {
			return fmt.Errorf("invalid CIDR: %w", err)
		}
		return nil
	}

	// Check if it's a single IP
	ip := net.ParseIP(ipstr)
	if ip == nil {
		return fmt.Errorf("invalid IP address")
	}

	// Ensure it's IPv4
	if ip.To4() == nil {
		return fmt.Errorf("not an IPv4 address")
	}

	return nil
}

// validateIPv6 validates an IPv6 address or CIDR range
func validateIPv6(ipstr string) error {
	// Check if it's a CIDR
	if strings.Contains(ipstr, "/") {
		_, _, err := net.ParseCIDR(ipstr)
		if err != nil {
			return fmt.Errorf("invalid CIDR: %w", err)
		}
		return nil
	}

	// Check if it's a single IP
	ip := net.ParseIP(ipstr)
	if ip == nil {
		return fmt.Errorf("invalid IP address")
	}

	// Ensure it's IPv6
	if ip.To4() != nil {
		return fmt.Errorf("not an IPv6 address")
	}

	return nil
}

// validateHostname validates a hostname (basic validation)
func validateHostname(hostname string) error {
	if hostname == "" {
		return fmt.Errorf("empty hostname")
	}

	// Allow placeholders like <tenant>
	if strings.Contains(hostname, "<") && strings.Contains(hostname, ">") {
		return nil // Placeholder is valid
	}

	// Allow wildcards
	if strings.HasPrefix(hostname, "*.") {
		return nil
	}

	// Basic hostname validation
	if len(hostname) > 255 {
		return fmt.Errorf("hostname too long")
	}

	// Check for invalid characters (very basic check)
	if strings.ContainsAny(hostname, " \t\n\r") {
		return fmt.Errorf("hostname contains whitespace")
	}

	return nil
}
