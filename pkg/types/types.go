package types

import (
	"encoding/json"
	"fmt"
)

// NetworkRequirements represents the root structure of the network requirements JSON
type NetworkRequirements struct {
	Version     string             `json:"version"`
	UpdatedAt   string             `json:"updated_at"`
	Description string             `json:"description"`
	Outbound    map[string]Service `json:"-"` // Custom unmarshal
	Inbound     map[string]Service `json:"-"` // Custom unmarshal
	RegionCodes map[string]string  `json:"region_codes"`
}

// UnmarshalJSON custom unmarshaler to handle both old and new JSON formats.
//
// Old format (v1): outbound/inbound are maps of service_name -> service_object
//
//	{"outbound": {"description": "...", "service_name": {...}}}
//
// New format (v2): outbound/inbound have "description" and "items" array
//
//	{"outbound": {"description": "...", "items": [{"id": "service_name", ...}]}}
func (nr *NetworkRequirements) UnmarshalJSON(data []byte) error {
	var temp struct {
		Version     string            `json:"version"`
		UpdatedAt   string            `json:"updated_at"`
		Description string            `json:"description"`
		Outbound    json.RawMessage   `json:"outbound"`
		Inbound     json.RawMessage   `json:"inbound"`
		RegionCodes map[string]string `json:"region_codes"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	nr.Version = temp.Version
	nr.UpdatedAt = temp.UpdatedAt
	nr.Description = temp.Description
	nr.RegionCodes = temp.RegionCodes

	var err error
	nr.Outbound, err = parseDirection(temp.Outbound, "outbound")
	if err != nil {
		return err
	}

	nr.Inbound, err = parseDirection(temp.Inbound, "inbound")
	if err != nil {
		return err
	}

	return nil
}

// parseDirection handles both old (map) and new (items array) formats
func parseDirection(raw json.RawMessage, direction string) (map[string]Service, error) {
	services := make(map[string]Service)

	if raw == nil || string(raw) == "null" || string(raw) == "{}" {
		return services, nil
	}

	// Try new format first: {"description": "...", "items": [...]}
	var newFormat struct {
		Description string            `json:"description"`
		Items       []json.RawMessage `json:"items"`
		HasItems    bool              `json:"-"`
	}
	// Check if "items" key exists in the JSON
	var rawMap map[string]json.RawMessage
	if err := json.Unmarshal(raw, &rawMap); err == nil {
		if _, ok := rawMap["items"]; ok {
			newFormat.HasItems = true
		}
	}
	if err := json.Unmarshal(raw, &newFormat); err == nil && newFormat.HasItems {
		for _, itemRaw := range newFormat.Items {
			var item serviceV2
			if err := json.Unmarshal(itemRaw, &item); err != nil {
				return nil, fmt.Errorf("failed to unmarshal %s service items: %w", direction, err)
			}

			service := Service{
				ID:          item.ID,
				Description: item.Description,
				Protocol:    item.Protocol,
				FlatPorts:   item.Ports,
				Regions:     item.Regions,
				Required:    item.Required,
			}
			services[item.ID] = service
		}
		return services, nil
	}

	// Fall back to old format: {"description": "...", "service_name": {...}}
	var oldFormat map[string]interface{}
	if err := json.Unmarshal(raw, &oldFormat); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s: %w", direction, err)
	}

	for key, value := range oldFormat {
		if key == "description" {
			continue
		}

		serviceJSON, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal %s service %s: %w", direction, key, err)
		}

		var service Service
		if err := json.Unmarshal(serviceJSON, &service); err != nil {
			return nil, fmt.Errorf("failed to unmarshal %s service %s: %w", direction, key, err)
		}

		services[key] = service
	}

	return services, nil
}

// serviceV2 represents the new JSON format for services (with "id" and flat "ports")
type serviceV2 struct {
	ID          string                `json:"id"`
	Description string                `json:"description"`
	Protocol    string                `json:"protocol,omitempty"`
	Ports       []int                 `json:"ports,omitempty"`
	Regions     map[string]RegionData `json:"regions,omitempty"`
	Required    *bool                 `json:"required,omitempty"`
}

// Service represents a network service with its configuration
type Service struct {
	// Common fields
	Description string                `json:"description"`
	Regions     map[string]RegionData `json:"regions,omitempty"`

	// New format fields (v2)
	ID        string `json:"id,omitempty"`
	Protocol  string `json:"protocol,omitempty"`
	FlatPorts []int  `json:"ports,omitempty"`
	Required  *bool  `json:"required,omitempty"`

	// Old format fields (v1)
	TCPPorts []int        `json:"tcp_ports,omitempty"`
	UDPPorts []int        `json:"udp_ports,omitempty"`
	NestedPorts *PortsNested `json:"nested_ports,omitempty"`
}

// UnmarshalJSON handles the "ports" field which can be either a flat []int (v2)
// or a nested object (v1 AD Connector style)
func (s *Service) UnmarshalJSON(data []byte) error {
	// Use an alias to avoid infinite recursion
	type ServiceAlias Service

	// First try with flat ports (new format or simple old format)
	var flat struct {
		ServiceAlias
		Ports json.RawMessage `json:"ports,omitempty"`
	}

	if err := json.Unmarshal(data, &flat); err != nil {
		return err
	}

	*s = Service(flat.ServiceAlias)

	// Parse "ports" field which could be []int or nested object
	if len(flat.Ports) > 0 {
		// Try as []int first
		var flatPorts []int
		if err := json.Unmarshal(flat.Ports, &flatPorts); err == nil {
			s.FlatPorts = flatPorts
			return nil
		}

		// Try as nested object
		var nested PortsNested
		if err := json.Unmarshal(flat.Ports, &nested); err == nil {
			s.NestedPorts = &nested
			return nil
		}
	}

	return nil
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
