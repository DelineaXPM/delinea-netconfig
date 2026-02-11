package types

import (
	"encoding/json"
	"fmt"
)

// NetworkRequirements represents the root structure of the network requirements JSON
type NetworkRequirements struct {
	Version     string                 `json:"version"`
	UpdatedAt   string                 `json:"updated_at"`
	Description string                 `json:"description"`
	Outbound    map[string]Service     `json:"-"`  // Custom unmarshal
	Inbound     map[string]Service     `json:"-"`  // Custom unmarshal
	RegionCodes map[string]string      `json:"region_codes"`
}

// UnmarshalJSON custom unmarshaler to handle the description field in outbound/inbound
func (nr *NetworkRequirements) UnmarshalJSON(data []byte) error {
	// First, unmarshal into a temporary structure
	var temp struct {
		Version     string                 `json:"version"`
		UpdatedAt   string                 `json:"updated_at"`
		Description string                 `json:"description"`
		Outbound    map[string]interface{} `json:"outbound"`
		Inbound     map[string]interface{} `json:"inbound"`
		RegionCodes map[string]string      `json:"region_codes"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Copy simple fields
	nr.Version = temp.Version
	nr.UpdatedAt = temp.UpdatedAt
	nr.Description = temp.Description
	nr.RegionCodes = temp.RegionCodes

	// Process outbound, filtering out the top-level description
	nr.Outbound = make(map[string]Service)
	for key, value := range temp.Outbound {
		if key == "description" {
			// Skip the description field
			continue
		}

		// Unmarshal the service
		serviceJSON, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal outbound service %s: %w", key, err)
		}

		var service Service
		if err := json.Unmarshal(serviceJSON, &service); err != nil {
			return fmt.Errorf("failed to unmarshal outbound service %s: %w", key, err)
		}

		nr.Outbound[key] = service
	}

	// Process inbound, filtering out the top-level description
	nr.Inbound = make(map[string]Service)
	for key, value := range temp.Inbound {
		if key == "description" {
			// Skip the description field
			continue
		}

		// Unmarshal the service
		serviceJSON, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal inbound service %s: %w", key, err)
		}

		var service Service
		if err := json.Unmarshal(serviceJSON, &service); err != nil {
			return fmt.Errorf("failed to unmarshal inbound service %s: %w", key, err)
		}

		nr.Inbound[key] = service
	}

	return nil
}

// Service represents a network service with its configuration
type Service struct {
	Description string                 `json:"description"`
	TCPPorts    []int                  `json:"tcp_ports,omitempty"`
	UDPPorts    []int                  `json:"udp_ports,omitempty"`
	Ports       *PortsNested           `json:"ports,omitempty"`
	Regions     map[string]RegionData  `json:"regions,omitempty"`
}

// PortsNested represents nested port configuration (for services like AD Connector)
type PortsNested struct {
	External       *PortSpec `json:"external,omitempty"`
	InternalToADDC *PortSpec `json:"internal_to_ad_dc,omitempty"`
}

// PortSpec defines TCP and UDP ports for a service
type PortSpec struct {
	TCPPorts []int `json:"tcp_ports,omitempty"`
	UDPPorts []int `json:"udp_ports,omitempty"`
}

// RegionData contains regional configuration for a service
type RegionData struct {
	Domain              string   `json:"domain,omitempty"`
	IPv4                []string `json:"ipv4,omitempty"`
	IPv6                []string `json:"ipv6,omitempty"`
	IPv4Primary         []string `json:"ipv4_primary,omitempty"`
	IPv4DR              []string `json:"ipv4_dr,omitempty"`
	Hostnames           []string `json:"hostnames,omitempty"`
	HostnamesSelfSigned []string `json:"hostnames_self_signed,omitempty"`
	HostnamesCASigned   []string `json:"hostnames_ca_signed,omitempty"`
	AWSSES              []string `json:"aws_ses,omitempty"`
}

// NetworkEntry represents a normalized network rule entry
// This is the common format used by all converters
type NetworkEntry struct {
	Direction   string   // "outbound" or "inbound"
	Service     string   // Service name (e.g., "platform_ssc_ips", "webhooks")
	Region      string   // Region code (e.g., "us", "global")
	Type        string   // Entry type (e.g., "ipv4", "ipv6", "hostname")
	Values      []string // IP ranges or hostnames
	Protocol    string   // Protocol (e.g., "tcp", "udp", "both")
	Ports       []int    // Port numbers
	Description string   // Service description
	Redundancy  string   // Redundancy tier ("primary", "dr", or empty)
	Tags        []string // Optional tags for grouping
}

// Converter is the interface that all format converters must implement
type Converter interface {
	// Convert transforms network entries into the target format
	Convert(entries []NetworkEntry) ([]byte, error)

	// Name returns the human-readable name of the converter
	Name() string

	// FileExtension returns the file extension for this format (without the dot)
	FileExtension() string
}
