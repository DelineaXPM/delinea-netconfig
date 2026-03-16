package converter

import (
	"fmt"

	"github.com/DelineaXPM/delinea-netconfig/pkg/types"
)

// Converter is the interface that all format converters must implement
type Converter interface {
	Convert(entries []types.NetworkEntry) ([]byte, error)
	Name() string
	FileExtension() string
}

var converters = map[string]Converter{
	"csv":       &CSVConverter{},
	"yaml":      &YAMLConverter{},
	"terraform": &TerraformConverter{},
	"tf":        &TerraformConverter{}, // Alias
	"ansible":   &AnsibleConverter{},
	"aws-sg":    &AWSSecurityGroupConverter{},
	"cisco":     &CiscoACLConverter{},
	"panos":     &PANOSConverter{},
}

// GetConverter returns a converter for the specified format
func GetConverter(format string) (Converter, error) {
	converter, ok := converters[format]
	if !ok {
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
	return converter, nil
}

// ListConverters returns a list of all available converter names
func ListConverters() []string {
	names := make([]string, 0, len(converters))
	for name := range converters {
		names = append(names, name)
	}
	return names
}

// Ensure converters implement the types.Converter interface
var (
	_ types.Converter = (*CSVConverter)(nil)
	_ types.Converter = (*YAMLConverter)(nil)
	_ types.Converter = (*TerraformConverter)(nil)
	_ types.Converter = (*AnsibleConverter)(nil)
	_ types.Converter = (*AWSSecurityGroupConverter)(nil)
	_ types.Converter = (*CiscoACLConverter)(nil)
	_ types.Converter = (*PANOSConverter)(nil)
)
