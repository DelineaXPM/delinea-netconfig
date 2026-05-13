package cli

import (
	"testing"

	"github.com/DelineaXPM/delinea-netconfig/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestFilterByService(t *testing.T) {
	entries := []types.NetworkEntry{
		{Service: "platform_engine_messaging", Region: "eu"},
		{Service: "platform_engine_messaging", Region: "us"},
		{Service: "web_front_end", Region: "global"},
		{Service: "storage", Region: "global"},
	}

	tests := []struct {
		name     string
		filter   string
		expected int
	}{
		{"exact match", "platform_engine_messaging", 2},
		{"single service", "storage", 1},
		{"case insensitive", "Web_Front_End", 1},
		{"no match", "nonexistent", 0},
		{"whitespace", "  storage  ", 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := filterByService(entries, tc.filter)
			assert.Len(t, got, tc.expected)
		})
	}
}

func TestCheckCmdFlags(t *testing.T) {
	// Sanity check: the command is registered with the expected flags so
	// users can rely on them in scripts.
	expected := []string{"file", "url", "tenant", "region", "service", "timeout", "concurrency", "insecure"}
	for _, name := range expected {
		flag := checkCmd.Flags().Lookup(name)
		assert.NotNilf(t, flag, "expected flag --%s to be registered on check command", name)
	}
}
