package connchk

import (
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DelineaXPM/delinea-netconfig/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildJobsSkipsCIDRAndPlaceholders(t *testing.T) {
	entries := []types.NetworkEntry{
		{
			Service: "svc",
			Region:  "us",
			Values:  []string{"203.0.113.0/24", "203.0.113.1", "<tenant>.example.com", "host.example.com"},
			Ports:   []int{443, 5671},
		},
	}

	jobs := buildJobs(entries)

	targets := make([]string, 0, len(jobs))
	for _, j := range jobs {
		targets = append(targets, j.target)
	}
	assert.ElementsMatch(t, []string{"203.0.113.1", "host.example.com"}, targets)
	for _, j := range jobs {
		assert.Equal(t, []int{443, 5671}, j.ports)
	}
}

func TestBuildJobsMergesPortsAcrossEntries(t *testing.T) {
	entries := []types.NetworkEntry{
		{Service: "svc", Region: "us", Values: []string{"host.example.com"}, Ports: []int{443}},
		{Service: "svc", Region: "us", Values: []string{"host.example.com"}, Ports: []int{443, 5671}},
	}

	jobs := buildJobs(entries)
	require.Len(t, jobs, 1)
	assert.Equal(t, []int{443, 5671}, jobs[0].ports)
}

func TestIsHostname(t *testing.T) {
	assert.True(t, isHostname("example.com"))
	assert.False(t, isHostname("203.0.113.1"))
	assert.False(t, isHostname("2001:db8::1"))
}

// TestRunAgainstLocalListener spins up a plain TCP server (no TLS) and asserts
// that the probe reports DNS skipped (IP), TCP ok, and no TLS attempt for a
// non-TLS port.
func TestRunAgainstLocalListener(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	go func() {
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

	entries := []types.NetworkEntry{{
		Service: "loopback",
		Region:  "local",
		Values:  []string{"127.0.0.1"},
		Ports:   []int{port},
	}}

	results, summary := Run(context.Background(), entries, CheckOptions{Timeout: time.Second})
	require.Len(t, results, 1)
	r := results[0]
	assert.Equal(t, StatusSkipped, r.DNS.Status, "DNS should be skipped for IP literal")
	require.Len(t, r.TCP, 1)
	assert.Equal(t, StatusOK, r.TCP[0].Status)
	assert.Nil(t, r.TCP[0].TLS, "non-TLS port should not attempt handshake")
	assert.Equal(t, 1, summary.TCPOK)
	assert.False(t, HasFailures(summary))
}

// TestRunTLSAgainstHTTPSServer confirms that a TLS handshake is attempted on
// the configured TLS ports, and that the SAN-fallback subject is used when CN
// is empty (httptest's self-signed cert has CN="example.com" so this also
// exercises the typical CN path).
func TestRunTLSAgainstHTTPSServer(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	host, portStr, _ := net.SplitHostPort(strings.TrimPrefix(server.URL, "https://"))
	port, _ := strconv.Atoi(portStr)

	// Two entries: one targeting the IP (TLS handshake should be skipped —
	// no SNI for IP literals), one targeting "localhost" so SNI matches the
	// httptest cert and the handshake completes.
	entries := []types.NetworkEntry{
		{Service: "https", Region: "local", Values: []string{host}, Ports: []int{port}},
		{Service: "https", Region: "local", Values: []string{"localhost"}, Ports: []int{port}},
	}

	results, _ := Run(context.Background(), entries, CheckOptions{
		Timeout:            2 * time.Second,
		InsecureSkipVerify: true,
		TLSPorts:           []int{port},
	})
	require.Len(t, results, 2)

	var ip, hostname ProbeResult
	for _, r := range results {
		if r.IsHostname {
			hostname = r
		} else {
			ip = r
		}
	}

	require.Len(t, ip.TCP, 1)
	require.NotNil(t, ip.TCP[0].TLS, "TLS probe should exist but be skipped for IP literal")
	assert.Equal(t, StatusSkipped, ip.TCP[0].TLS.Status)

	require.Len(t, hostname.TCP, 1)
	require.NotNil(t, hostname.TCP[0].TLS)
	assert.Equal(t, StatusOK, hostname.TCP[0].TLS.Status)
	assert.NotEmpty(t, hostname.TCP[0].TLS.CertSubject, "cert subject should be populated from CN or SAN")
}

func TestClassifyTLSErr(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		contains string
	}{
		{
			name:     "unknown CA",
			err:      x509.UnknownAuthorityError{},
			contains: "certificate signed by unknown authority (possible SSL inspection / proxy)",
		},
		{
			name:     "hostname mismatch is not labelled as SSL inspection",
			err:      x509.HostnameError{Host: "wrong.example.com"},
			contains: "certificate hostname mismatch",
		},
		{
			name:     "expired or otherwise invalid cert",
			err:      x509.CertificateInvalidError{Reason: x509.Expired, Detail: "leaf expired"},
			contains: "certificate invalid",
		},
		{
			name:     "context deadline exceeded",
			err:      context.DeadlineExceeded,
			contains: "TLS handshake timed out",
		},
		{
			name:     "i/o timeout net error",
			err:      errors.New("read tcp 10.0.0.1:443: i/o timeout"),
			contains: "TLS handshake timed out",
		},
		{
			name:     "passthrough for unknown errors",
			err:      errors.New("connection reset by peer"),
			contains: "connection reset by peer",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := classifyTLSErr(tc.err)
			assert.Contains(t, got, tc.contains)
		})
	}
}

// TestClassifyTLSErrHostnameMismatchIsNotSSLInspection is a focused regression
// guard for the bug surfaced in code review: the old string-based classifier
// labelled "x509: certificate is valid for X, not Y" as SSL inspection. The
// new errors.As-based path must call it a hostname mismatch.
func TestClassifyTLSErrHostnameMismatchIsNotSSLInspection(t *testing.T) {
	got := classifyTLSErr(x509.HostnameError{Host: "wrong.example.com"})
	assert.NotContains(t, got, "SSL inspection")
	assert.Contains(t, got, "hostname mismatch")
}

func TestCertSubjectFallsBackToSAN(t *testing.T) {
	// CN populated → CN wins
	cn := certSubject(&x509.Certificate{
		Subject:  pkix.Name{CommonName: "api.example.com"},
		DNSNames: []string{"san.example.com"},
	})
	assert.Equal(t, "api.example.com", cn)

	// CN empty → first SAN wins
	san := certSubject(&x509.Certificate{DNSNames: []string{"san.example.com", "alt.example.com"}})
	assert.Equal(t, "san.example.com", san)

	// Neither → empty
	empty := certSubject(&x509.Certificate{})
	assert.Equal(t, "", empty)
}
