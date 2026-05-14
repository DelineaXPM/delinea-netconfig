package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/DelineaXPM/delinea-netconfig/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// TestRunCheckFlowAgainstLocalListener exercises the full runCheck pipeline —
// load → parse → normalize → filter → probe → render — using a temp JSON
// fixture that points at a real loopback TCP listener. This is the gap the PR
// reviewer flagged: prior tests only covered the probe engine in isolation.
func TestRunCheckFlowAgainstLocalListener(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	// Accept-and-close loop so dialers get a clean TCP handshake.
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			c, err := listener.Accept()
			if err != nil {
				return
			}
			_ = c.Close()
		}
	}()

	_, portStr, _ := net.SplitHostPort(listener.Addr().String())
	port, _ := strconv.Atoi(portStr)

	dir := t.TempDir()
	fixturePath := filepath.Join(dir, "network-requirements.json")
	writeFixture(t, fixturePath, port)

	// Reset and configure the package-level flag vars the way cobra would
	// after parsing the command line.
	prev := snapshotCheckFlags()
	defer restoreCheckFlags(prev)

	checkInputFile = fixturePath
	checkInputURL = ""
	checkTenantName = ""
	checkServiceFilter = "test_service"
	checkRegionFilter = ""
	checkTimeout = 2 * time.Second
	checkConcurrency = 4
	checkInsecure = false

	var stdout bytes.Buffer
	checkCmd.SetOut(&stdout)
	checkCmd.SetErr(&stdout) // collapse for assertion convenience
	defer func() {
		checkCmd.SetOut(nil)
		checkCmd.SetErr(nil)
	}()

	err = runCheck(checkCmd, nil)
	require.NoError(t, err, "expected the loopback probe to succeed")

	got := stdout.String()
	assert.Contains(t, got, "=== Connectivity Check ===")
	assert.Contains(t, got, "127.0.0.1")
	assert.Contains(t, got, fmt.Sprintf("TCP %d reachable", port))
	assert.Contains(t, got, "Result: ✓ all probes succeeded.")
}

func TestRunCheckRejectsInvalidFlags(t *testing.T) {
	prev := snapshotCheckFlags()
	defer restoreCheckFlags(prev)

	t.Run("zero timeout", func(t *testing.T) {
		restoreCheckFlags(prev)
		checkTimeout = 0
		err := runCheck(checkCmd, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "--timeout must be > 0")
	})

	t.Run("zero concurrency", func(t *testing.T) {
		restoreCheckFlags(prev)
		checkConcurrency = 0
		err := runCheck(checkCmd, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "--concurrency must be > 0")
	})

	t.Run("unknown service", func(t *testing.T) {
		restoreCheckFlags(prev)
		dir := t.TempDir()
		fixture := filepath.Join(dir, "nr.json")
		writeFixture(t, fixture, 65432) // port doesn't matter, we never dial
		checkInputFile = fixture
		checkServiceFilter = "does_not_exist"

		err := runCheck(checkCmd, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no entries match service id")
	})
}

func TestErrProbesFailedIsSentinel(t *testing.T) {
	// main.go relies on errors.Is for the exit-code branch. Wrapping must
	// keep the sentinel discoverable.
	wrapped := fmt.Errorf("rendered report: %w", ErrProbesFailed)
	assert.True(t, errors.Is(wrapped, ErrProbesFailed))
}

// writeFixture produces a minimal valid network-requirements.json (new "items"
// schema) with a single outbound entry pointing at 127.0.0.1:<port>.
func writeFixture(t *testing.T, path string, port int) {
	t.Helper()

	doc := map[string]any{
		"version":     "0.0.0-test",
		"updated_at":  "2026-01-01T00:00:00Z",
		"description": "test fixture",
		"region_codes": map[string]string{
			"global": "All regions",
		},
		"outbound": map[string]any{
			"description": "test fixture outbound",
			"items": []map[string]any{
				{
					"id":          "test_service",
					"description": "loopback target",
					"protocol":    "tcp",
					"ports":       []int{port},
					"regions": map[string]any{
						"global": map[string]any{
							"ipv4": []string{"127.0.0.1"},
						},
					},
				},
			},
		},
	}

	data, err := json.Marshal(doc)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0o644))
}

type checkFlagSnapshot struct {
	file, url, tenant, region, service string
	timeout                            time.Duration
	concurrency                        int
	insecure                           bool
}

func snapshotCheckFlags() checkFlagSnapshot {
	return checkFlagSnapshot{
		file:        checkInputFile,
		url:         checkInputURL,
		tenant:      checkTenantName,
		region:      checkRegionFilter,
		service:     checkServiceFilter,
		timeout:     checkTimeout,
		concurrency: checkConcurrency,
		insecure:    checkInsecure,
	}
}

func restoreCheckFlags(s checkFlagSnapshot) {
	checkInputFile = s.file
	checkInputURL = s.url
	checkTenantName = s.tenant
	checkRegionFilter = s.region
	checkServiceFilter = s.service
	checkTimeout = s.timeout
	checkConcurrency = s.concurrency
	checkInsecure = s.insecure
}
