package converter

import (
	"fmt"

	"github.com/DelineaXPM/delinea-platform/delinea-netconfig/pkg/types"
	"gopkg.in/yaml.v3"
)

// YAMLConverter converts network entries to YAML format
type YAMLConverter struct{}

// YAMLOutput represents the structure of the YAML output
type YAMLOutput struct {
	DelineaNetworkRequirements map[string]interface{} `yaml:"delinea_network_requirements"`
}

// Convert converts network entries to YAML
func (c *YAMLConverter) Convert(entries []types.NetworkEntry) ([]byte, error) {
	// Organize entries by direction -> service -> region
	output := YAMLOutput{
		DelineaNetworkRequirements: make(map[string]interface{}),
	}

	directionMap := make(map[string]map[string]map[string][]map[string]interface{})

	for _, entry := range entries {
		// Initialize nested maps if needed
		if directionMap[entry.Direction] == nil {
			directionMap[entry.Direction] = make(map[string]map[string][]map[string]interface{})
		}
		if directionMap[entry.Direction][entry.Service] == nil {
			directionMap[entry.Direction][entry.Service] = make(map[string][]map[string]interface{})
		}

		// Create entry map
		entryMap := map[string]interface{}{
			"type":        entry.Type,
			"values":      entry.Values,
			"protocol":    entry.Protocol,
			"ports":       entry.Ports,
			"description": entry.Description,
		}

		if entry.Redundancy != "" {
			entryMap["redundancy"] = entry.Redundancy
		}

		if len(entry.Tags) > 0 {
			entryMap["tags"] = entry.Tags
		}

		directionMap[entry.Direction][entry.Service][entry.Region] = append(
			directionMap[entry.Direction][entry.Service][entry.Region],
			entryMap,
		)
	}

	// Convert to output structure
	for direction, services := range directionMap {
		serviceMap := make(map[string]interface{})
		for service, regions := range services {
			serviceMap[service] = regions
		}
		output.DelineaNetworkRequirements[direction] = serviceMap
	}

	// Marshal to YAML
	data, err := yaml.Marshal(&output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal YAML: %w", err)
	}

	return data, nil
}

// Name returns the name of the converter
func (c *YAMLConverter) Name() string {
	return "YAML"
}

// FileExtension returns the file extension for YAML files
func (c *YAMLConverter) FileExtension() string {
	return "yaml"
}
