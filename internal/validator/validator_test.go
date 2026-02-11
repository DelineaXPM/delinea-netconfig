package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateIPv4(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectErr bool
	}{
		// Valid cases
		{
			name:      "valid IPv4 address",
			input:     "192.168.1.1",
			expectErr: false,
		},
		{
			name:      "valid IPv4 CIDR /24",
			input:     "192.168.1.0/24",
			expectErr: false,
		},
		{
			name:      "valid IPv4 CIDR /32",
			input:     "192.168.1.1/32",
			expectErr: false,
		},
		{
			name:      "valid IPv4 CIDR /8",
			input:     "10.0.0.0/8",
			expectErr: false,
		},
		{
			name:      "valid public IP",
			input:     "8.8.8.8",
			expectErr: false,
		},
		{
			name:      "valid private IP 10.x",
			input:     "10.0.0.1",
			expectErr: false,
		},
		{
			name:      "valid private IP 172.16.x",
			input:     "172.16.0.1",
			expectErr: false,
		},

		// Invalid cases
		{
			name:      "invalid IP format",
			input:     "256.1.1.1",
			expectErr: true,
		},
		{
			name:      "incomplete IP",
			input:     "192.168.1",
			expectErr: true,
		},
		{
			name:      "too many octets",
			input:     "192.168.1.1.1",
			expectErr: true,
		},
		{
			name:      "invalid CIDR prefix",
			input:     "192.168.1.0/33",
			expectErr: true,
		},
		{
			name:      "invalid CIDR format",
			input:     "192.168.1.0/",
			expectErr: true,
		},
		{
			name:      "IPv6 address (not IPv4)",
			input:     "2001:db8::1",
			expectErr: true,
		},
		{
			name:      "empty string",
			input:     "",
			expectErr: true,
		},
		{
			name:      "hostname instead of IP",
			input:     "example.com",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateIPv4(tt.input)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateIPv6(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectErr bool
	}{
		// Valid cases
		{
			name:      "valid IPv6 address",
			input:     "2001:db8::1",
			expectErr: false,
		},
		{
			name:      "valid IPv6 CIDR",
			input:     "2001:db8::/32",
			expectErr: false,
		},
		{
			name:      "valid IPv6 full address",
			input:     "2001:0db8:0000:0000:0000:0000:0000:0001",
			expectErr: false,
		},
		{
			name:      "valid IPv6 loopback",
			input:     "::1",
			expectErr: false,
		},
		{
			name:      "valid IPv6 all zeros",
			input:     "::",
			expectErr: false,
		},
		{
			name:      "valid IPv6 CIDR /64",
			input:     "2001:db8::/64",
			expectErr: false,
		},

		// Invalid cases
		{
			name:      "IPv4 address (not IPv6)",
			input:     "192.168.1.1",
			expectErr: true,
		},
		{
			name:      "invalid IPv6 format",
			input:     "2001:db8:::1",
			expectErr: true,
		},
		{
			name:      "invalid CIDR prefix",
			input:     "2001:db8::/129",
			expectErr: true,
		},
		{
			name:      "empty string",
			input:     "",
			expectErr: true,
		},
		{
			name:      "hostname instead of IP",
			input:     "example.com",
			expectErr: true,
		},
		{
			name:      "IPv6 with too many segments",
			input:     "2001:db8::1::2",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateIPv6(tt.input)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateHostname(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectErr bool
	}{
		// Valid cases
		{
			name:      "valid hostname",
			input:     "example.com",
			expectErr: false,
		},
		{
			name:      "valid subdomain",
			input:     "api.example.com",
			expectErr: false,
		},
		{
			name:      "valid deep subdomain",
			input:     "api.v1.example.com",
			expectErr: false,
		},
		{
			name:      "hostname with hyphen",
			input:     "my-api.example.com",
			expectErr: false,
		},
		{
			name:      "placeholder with <tenant>",
			input:     "<tenant>.example.com",
			expectErr: false,
		},
		{
			name:      "wildcard hostname",
			input:     "*.example.com",
			expectErr: false,
		},
		{
			name:      "single label hostname",
			input:     "localhost",
			expectErr: false,
		},
		{
			name:      "hostname with numbers",
			input:     "api123.example.com",
			expectErr: false,
		},

		// Invalid cases
		{
			name:      "empty hostname",
			input:     "",
			expectErr: true,
		},
		{
			name:      "hostname with spaces",
			input:     "example .com",
			expectErr: true,
		},
		{
			name:      "hostname with tab",
			input:     "example\t.com",
			expectErr: true,
		},
		{
			name:      "hostname with newline",
			input:     "example\n.com",
			expectErr: true,
		},
		{
			name:      "hostname too long (>255 chars)",
			input:     string(make([]byte, 256)), // 256 'a's
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For the long hostname test, fill it with 'a's
			input := tt.input
			if len(input) == 256 {
				for i := range []byte(input) {
					input = string(append([]byte(input[:i]), 'a'))
				}
			}

			err := validateHostname(input)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name          string
		jsonData      string
		expectErr     bool
		checkResult   func(*testing.T, *ValidationResult)
		checkWarnings func(*testing.T, []string)
	}{
		{
			name: "valid simple JSON",
			jsonData: `{
				"version": "1.0",
				"outbound": {
					"test_service": {
						"description": "Test service",
						"tcp_ports": [443],
						"regions": {
							"us": {
								"ipv4": ["192.168.1.0/24"]
							}
						}
					}
				},
				"inbound": {},
				"region_codes": {
					"us": "United States"
				}
			}`,
			expectErr: false,
			checkResult: func(t *testing.T, result *ValidationResult) {
				assert.Equal(t, "1.0", result.Version)
				assert.Equal(t, 1, result.IPv4Count)
				assert.Equal(t, 0, result.IPv6Count)
				assert.Equal(t, 0, result.HostnameCount)
				assert.Equal(t, 1, result.OutboundServices)
				assert.Equal(t, 0, result.InboundServices)
				assert.Equal(t, 1, result.TotalServices)
				assert.Equal(t, 1, result.RegionCount)
				assert.Empty(t, result.Warnings)
			},
		},
		{
			name: "multiple address types",
			jsonData: `{
				"version": "1.0",
				"outbound": {
					"service1": {
						"description": "Multi-type service",
						"tcp_ports": [443],
						"regions": {
							"us": {
								"ipv4": ["192.168.1.0/24", "10.0.0.0/8"],
								"ipv6": ["2001:db8::/32"],
								"hostnames": ["api.example.com", "app.example.com"]
							}
						}
					}
				},
				"inbound": {},
				"region_codes": {
					"us": "United States"
				}
			}`,
			expectErr: false,
			checkResult: func(t *testing.T, result *ValidationResult) {
				assert.Equal(t, 2, result.IPv4Count)
				assert.Equal(t, 1, result.IPv6Count)
				assert.Equal(t, 2, result.HostnameCount)
				assert.Empty(t, result.Warnings)
			},
		},
		{
			name: "invalid IPv4 generates warning",
			jsonData: `{
				"version": "1.0",
				"outbound": {
					"test": {
						"description": "Test",
						"tcp_ports": [443],
						"regions": {
							"us": {
								"ipv4": ["256.1.1.1"]
							}
						}
					}
				},
				"inbound": {},
				"region_codes": {"us": "US"}
			}`,
			expectErr: false,
			checkResult: func(t *testing.T, result *ValidationResult) {
				assert.Equal(t, 0, result.IPv4Count) // Invalid, so not counted
				assert.Len(t, result.Warnings, 1)
			},
			checkWarnings: func(t *testing.T, warnings []string) {
				assert.Contains(t, warnings[0], "Invalid IPv4")
				assert.Contains(t, warnings[0], "256.1.1.1")
			},
		},
		{
			name: "invalid port generates warning",
			jsonData: `{
				"version": "1.0",
				"outbound": {
					"test": {
						"description": "Test",
						"tcp_ports": [99999],
						"regions": {
							"us": {
								"ipv4": ["192.168.1.0/24"]
							}
						}
					}
				},
				"inbound": {},
				"region_codes": {"us": "US"}
			}`,
			expectErr: false,
			checkResult: func(t *testing.T, result *ValidationResult) {
				assert.Len(t, result.Warnings, 1)
			},
			checkWarnings: func(t *testing.T, warnings []string) {
				assert.Contains(t, warnings[0], "Invalid port")
				assert.Contains(t, warnings[0], "99999")
			},
		},
		{
			name:      "invalid JSON",
			jsonData:  `{invalid json`,
			expectErr: true,
		},
		{
			name:      "empty JSON",
			jsonData:  ``,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Validate([]byte(tt.jsonData))

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)

				if tt.checkResult != nil {
					tt.checkResult(t, result)
				}

				if tt.checkWarnings != nil {
					tt.checkWarnings(t, result.Warnings)
				}
			}
		})
	}
}
