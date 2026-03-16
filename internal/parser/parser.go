package parser

import (
	"encoding/json"
	"fmt"

	"github.com/DelineaXPM/delinea-netconfig/pkg/types"
)

// Parse parses the network requirements JSON data
func Parse(data []byte) (*types.NetworkRequirements, error) {
	var networkReqs types.NetworkRequirements

	if err := json.Unmarshal(data, &networkReqs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Basic validation
	if networkReqs.Version == "" {
		return nil, fmt.Errorf("missing required field: version")
	}

	return &networkReqs, nil
}
