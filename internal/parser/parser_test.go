package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectErr bool
		errMsg    string
		version   string
	}{
		{
			name: "valid v1 format",
			input: `{
				"version": "1.0",
				"outbound": {
					"test_service": {
						"description": "Test",
						"tcp_ports": [443],
						"regions": {
							"us": {"ipv4": ["192.168.1.0/24"]}
						}
					}
				},
				"inbound": {},
				"region_codes": {"us": "United States"}
			}`,
			version: "1.0",
		},
		{
			name: "valid v2 format",
			input: `{
				"version": "2.0",
				"updated_at": "2025-01-01",
				"outbound": {
					"items": [
						{
							"id": "test_service",
							"description": "Test",
							"ports": [443],
							"protocol": "tcp",
							"regions": {
								"us": {"ipv4": ["192.168.1.0/24"]}
							}
						}
					]
				},
				"inbound": {"items": []}
			}`,
			version: "2.0",
		},
		{
			name:      "missing version field",
			input:     `{"outbound": {}, "inbound": {}}`,
			expectErr: true,
			errMsg:    "missing required field: version",
		},
		{
			name:      "invalid JSON",
			input:     `{not valid json`,
			expectErr: true,
		},
		{
			name:      "empty input",
			input:     ``,
			expectErr: true,
		},
		{
			name:      "empty JSON object",
			input:     `{}`,
			expectErr: true,
			errMsg:    "missing required field: version",
		},
		{
			name: "version only is valid",
			input: `{
				"version": "1.0",
				"outbound": {},
				"inbound": {}
			}`,
			version: "1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Parse([]byte(tt.input))

			if tt.expectErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.version, result.Version)
			}
		})
	}
}
