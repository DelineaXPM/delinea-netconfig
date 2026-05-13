// Package connchk performs network connectivity diagnostics against the targets
// defined in a normalized list of network entries (DNS resolution, TCP reachability,
// and TLS handshake). It is used by the `check` CLI subcommand to validate that a
// machine can actually reach the URLs Delinea publishes in network-requirements.json.
package connchk

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/DelineaXPM/delinea-netconfig/pkg/types"
)

// DefaultTimeout is the per-probe network timeout (DNS, TCP, TLS).
const DefaultTimeout = 5 * time.Second

// DefaultConcurrency caps the number of probes running in parallel.
const DefaultConcurrency = 10

// tlsPorts is the set of ports for which a TLS handshake is attempted in
// addition to the plain TCP dial. Other ports are assumed to be plaintext and
// only get the TCP reachability check.
var tlsPorts = map[int]bool{
	443:   true, // HTTPS / WSS
	5671:  true, // AMQPS
	8883:  true, // MQTTS
	636:   true, // LDAPS
	993:   true, // IMAPS
	995:   true, // POP3S
	465:   true, // SMTPS
	5986:  true, // WinRM HTTPS
	8443:  true, // alternate HTTPS
}

// ProbeStatus is the outcome of a single check step.
type ProbeStatus string

const (
	StatusOK      ProbeStatus = "ok"
	StatusFail    ProbeStatus = "fail"
	StatusSkipped ProbeStatus = "skipped"
)

// CheckOptions controls how the connectivity sweep is executed.
type CheckOptions struct {
	// Timeout applied to each DNS, TCP, and TLS operation. Zero means DefaultTimeout.
	Timeout time.Duration

	// Concurrency caps the number of targets probed in parallel. Zero means
	// DefaultConcurrency.
	Concurrency int

	// InsecureSkipVerify disables certificate validation during the TLS
	// handshake. Useful when probing self-signed endpoints or when the user
	// only wants to confirm that the handshake completes.
	InsecureSkipVerify bool
}

// ProbeResult captures the DNS + TCP + TLS outcome for a single target.
//
// A "target" is one hostname or IP literal. Each target gets at most one DNS
// probe (IP literals get StatusSkipped) and one TCPProbe per port. Each TCP
// probe may also carry a TLS sub-probe when the port is in tlsPorts and the
// target is a hostname.
type ProbeResult struct {
	Service    string
	Region     string
	Direction  string
	Target     string // hostname or IP literal that was probed
	IsHostname bool

	DNS  DNSProbe
	TCP  []TCPProbe
}

// DNSProbe is the result of resolving a hostname to one or more IPs.
type DNSProbe struct {
	Status    ProbeStatus
	Addresses []string
	Err       string
	Duration  time.Duration
}

// TCPProbe is the result of a single TCP dial against (target, port). When the
// port is TLS-typical and the target is a hostname, TLS holds the handshake
// outcome too.
type TCPProbe struct {
	Port     int
	Status   ProbeStatus
	Err      string
	Duration time.Duration
	TLS      *TLSProbe // nil when no TLS handshake was attempted
}

// TLSProbe is the result of a TLS handshake. CertSubject and CertIssuer come
// from the leaf certificate when the handshake succeeded.
type TLSProbe struct {
	Status      ProbeStatus
	Err         string
	Duration    time.Duration
	CertSubject string
	CertIssuer  string
}

// Summary aggregates pass/fail counts across all probes for a quick verdict line.
type Summary struct {
	Targets    int
	DNSOK      int
	DNSFail    int
	TCPOK      int
	TCPFail    int
	TLSOK      int
	TLSFail    int
	TLSSkipped int
}

// Run performs DNS, TCP, and TLS probes against every dialable value in the
// supplied entries and returns one ProbeResult per (service, region, target)
// tuple. Results are sorted deterministically (service, region, target).
//
// CIDR ranges are skipped: they describe an allow-list of source IPs, not a
// single dialable endpoint. Values that still contain a <tenant> placeholder
// are also skipped — the caller is responsible for substitution.
func Run(ctx context.Context, entries []types.NetworkEntry, opts CheckOptions) ([]ProbeResult, Summary) {
	if opts.Timeout <= 0 {
		opts.Timeout = DefaultTimeout
	}
	if opts.Concurrency <= 0 {
		opts.Concurrency = DefaultConcurrency
	}

	jobs := buildJobs(entries)

	results := make([]ProbeResult, len(jobs))
	sem := make(chan struct{}, opts.Concurrency)
	var wg sync.WaitGroup

	for i := range jobs {
		i := i
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			results[i] = probeOne(ctx, jobs[i], opts)
		}()
	}
	wg.Wait()

	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Service != results[j].Service {
			return results[i].Service < results[j].Service
		}
		if results[i].Region != results[j].Region {
			return results[i].Region < results[j].Region
		}
		return results[i].Target < results[j].Target
	})

	return results, summarize(results)
}

// job describes one (target, ports) probe to perform. Ports are deduplicated
// across all entries that share the same (service, region, target) so a host
// listed under both ipv4 and hostnames isn't dialed twice.
type job struct {
	service   string
	region    string
	direction string
	target    string
	isHost    bool
	ports     []int
}

func buildJobs(entries []types.NetworkEntry) []job {
	type key struct{ svc, region, target string }
	seen := make(map[key]*job)
	order := make([]key, 0)

	for _, entry := range entries {
		for _, value := range entry.Values {
			target := strings.TrimSpace(value)
			if target == "" || strings.Contains(target, "/") || strings.Contains(target, "<tenant>") {
				continue
			}

			k := key{entry.Service, entry.Region, target}
			existing, ok := seen[k]
			if !ok {
				existing = &job{
					service:   entry.Service,
					region:    entry.Region,
					direction: entry.Direction,
					target:    target,
					isHost:    isHostname(target),
				}
				seen[k] = existing
				order = append(order, k)
			}
			for _, port := range entry.Ports {
				if port > 0 && port <= 65535 {
					existing.ports = mergePort(existing.ports, port)
				}
			}
		}
	}

	jobs := make([]job, 0, len(order))
	for _, k := range order {
		j := seen[k]
		sort.Ints(j.ports)
		jobs = append(jobs, *j)
	}
	return jobs
}

func mergePort(ports []int, p int) []int {
	for _, existing := range ports {
		if existing == p {
			return ports
		}
	}
	return append(ports, p)
}

func probeOne(ctx context.Context, j job, opts CheckOptions) ProbeResult {
	result := ProbeResult{
		Service:    j.service,
		Region:     j.region,
		Direction:  j.direction,
		Target:     j.target,
		IsHostname: j.isHost,
	}

	result.DNS = probeDNS(ctx, j.target, j.isHost, opts.Timeout)

	for _, port := range j.ports {
		result.TCP = append(result.TCP, probePort(ctx, j.target, port, j.isHost, opts))
	}

	return result
}

func probeDNS(ctx context.Context, target string, isHost bool, timeout time.Duration) DNSProbe {
	if !isHost {
		return DNSProbe{Status: StatusSkipped, Addresses: []string{target}}
	}

	dialCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()
	resolver := net.DefaultResolver
	addrs, err := resolver.LookupHost(dialCtx, target)
	elapsed := time.Since(start)
	if err != nil {
		return DNSProbe{Status: StatusFail, Err: err.Error(), Duration: elapsed}
	}
	return DNSProbe{Status: StatusOK, Addresses: addrs, Duration: elapsed}
}

func probePort(ctx context.Context, target string, port int, isHost bool, opts CheckOptions) TCPProbe {
	address := net.JoinHostPort(target, strconv.Itoa(port))

	dialCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	dialer := &net.Dialer{Timeout: opts.Timeout}

	start := time.Now()
	conn, err := dialer.DialContext(dialCtx, "tcp", address)
	elapsed := time.Since(start)
	if err != nil {
		return TCPProbe{Port: port, Status: StatusFail, Err: err.Error(), Duration: elapsed}
	}
	defer conn.Close()

	probe := TCPProbe{Port: port, Status: StatusOK, Duration: elapsed}

	if tlsPorts[port] && isHost {
		probe.TLS = handshakeTLS(conn, target, opts)
	} else if tlsPorts[port] && !isHost {
		probe.TLS = &TLSProbe{Status: StatusSkipped, Err: "TLS handshake skipped for IP target (no SNI)"}
	}
	return probe
}

func handshakeTLS(conn net.Conn, host string, opts CheckOptions) *TLSProbe {
	deadline := time.Now().Add(opts.Timeout)
	_ = conn.SetDeadline(deadline)

	cfg := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: opts.InsecureSkipVerify, // exposed via --insecure flag
	}

	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	tlsConn := tls.Client(conn, cfg)
	start := time.Now()
	err := tlsConn.HandshakeContext(ctx)
	elapsed := time.Since(start)
	if err != nil {
		return &TLSProbe{Status: StatusFail, Err: classifyTLSErr(err), Duration: elapsed}
	}

	probe := &TLSProbe{Status: StatusOK, Duration: elapsed}
	if certs := tlsConn.ConnectionState().PeerCertificates; len(certs) > 0 {
		probe.CertSubject = certs[0].Subject.CommonName
		probe.CertIssuer = certs[0].Issuer.CommonName
	}
	_ = tlsConn.Close()
	return probe
}

func classifyTLSErr(err error) string {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "certificate signed by unknown authority"),
		strings.Contains(msg, "x509: certificate"):
		return fmt.Sprintf("certificate validation failed: %s (possible SSL inspection / proxy)", msg)
	case errors.Is(err, context.DeadlineExceeded), strings.Contains(msg, "i/o timeout"):
		return "TLS handshake timed out"
	default:
		return msg
	}
}

func isHostname(s string) bool {
	return net.ParseIP(s) == nil
}

func summarize(results []ProbeResult) Summary {
	s := Summary{Targets: len(results)}
	for _, r := range results {
		switch r.DNS.Status {
		case StatusOK:
			s.DNSOK++
		case StatusFail:
			s.DNSFail++
		}
		for _, t := range r.TCP {
			if t.Status == StatusOK {
				s.TCPOK++
			} else if t.Status == StatusFail {
				s.TCPFail++
			}
			if t.TLS != nil {
				switch t.TLS.Status {
				case StatusOK:
					s.TLSOK++
				case StatusFail:
					s.TLSFail++
				case StatusSkipped:
					s.TLSSkipped++
				}
			}
		}
	}
	return s
}

// HasFailures reports whether any probe failed. Useful for setting a non-zero
// exit code from the CLI.
func HasFailures(s Summary) bool {
	return s.DNSFail > 0 || s.TCPFail > 0 || s.TLSFail > 0
}
