package cli

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/DelineaXPM/delinea-netconfig/internal/connchk"
	"github.com/DelineaXPM/delinea-netconfig/internal/fetcher"
	"github.com/DelineaXPM/delinea-netconfig/internal/parser"
	"github.com/DelineaXPM/delinea-netconfig/pkg/types"
	"github.com/spf13/cobra"
)

var (
	checkInputFile     string
	checkInputURL      string
	checkTenantName    string
	checkRegionFilter  string
	checkServiceFilter string
	checkTimeout       time.Duration
	checkConcurrency   int
	checkInsecure      bool
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Test DNS, TCP, and TLS connectivity to network requirement endpoints",
	Long: `Probe the endpoints declared in network-requirements.json from the
machine running this command. For each unique hostname/IP the check performs:

  - DNS resolution (hostnames only)
  - TCP connect against every published port
  - TLS handshake on TLS-typical ports (443, 5671, 8883, 636, 993, 995, 465,
    5986, 8443), with SNI when the target is a hostname

CIDR ranges are skipped (they describe allowed source IPs, not dialable
endpoints). Values still containing the <tenant> placeholder are skipped unless
--tenant is provided.

Examples:
  # Check every endpoint from the default URL
  delinea-netconfig check

  # Check only the EU region against a local file
  delinea-netconfig check -f network-requirements.json --region eu

  # Check a single service across all its regions
  delinea-netconfig check --service platform_engine_messaging

  # Substitute tenant placeholder before probing
  delinea-netconfig check --tenant mycompany --region us

  # Tighten the timeout and increase parallelism on a fast network
  delinea-netconfig check --timeout 3s --concurrency 20`,
	RunE: runCheck,
}

func init() {
	checkCmd.Flags().StringVarP(&checkInputFile, "file", "f", "", "path to network-requirements.json file")
	checkCmd.Flags().StringVarP(&checkInputURL, "url", "u", "", "URL to fetch network-requirements.json")
	checkCmd.Flags().StringVarP(&checkTenantName, "tenant", "t", "", "substitute <tenant> placeholder with this value")
	checkCmd.Flags().StringVarP(&checkRegionFilter, "region", "r", "", "limit checks to global + a single region (e.g. eu, au, us)")
	checkCmd.Flags().StringVarP(&checkServiceFilter, "service", "s", "", "limit checks to a single service id (e.g. platform_engine_messaging)")
	checkCmd.Flags().DurationVar(&checkTimeout, "timeout", connchk.DefaultTimeout, "per-probe timeout (DNS, TCP, TLS)")
	checkCmd.Flags().IntVar(&checkConcurrency, "concurrency", connchk.DefaultConcurrency, "number of targets probed in parallel")
	checkCmd.Flags().BoolVar(&checkInsecure, "insecure", false, "skip TLS certificate validation (report handshake reachability only)")

	checkCmd.MarkFlagsMutuallyExclusive("file", "url")

	rootCmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) error {
	logVerbose("Starting connectivity check")

	data, source, err := loadCheckInput()
	if err != nil {
		return err
	}
	logVerbose("Loaded %d bytes from %s", len(data), source)

	networkReqs, err := parser.Parse(data)
	if err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}
	logInfo("Parsed network requirements version %s (updated: %s)", networkReqs.Version, networkReqs.UpdatedAt)

	entries := parser.Normalize(networkReqs)

	if checkTenantName != "" {
		entries = substituteTenant(entries, checkTenantName)
	}

	if checkServiceFilter != "" {
		entries = filterByService(entries, checkServiceFilter)
		if len(entries) == 0 {
			return fmt.Errorf("no entries match service id %q — run 'delinea-netconfig info' to list available services", checkServiceFilter)
		}
	}

	if checkRegionFilter != "" {
		region := strings.ToLower(strings.TrimSpace(checkRegionFilter))
		if region != "global" {
			before := len(entries)
			entries = filterByRegion(entries, region)
			logInfo("Filtered to global + %s: %d of %d entries", region, len(entries), before)
		}
	}

	opts := connchk.CheckOptions{
		Timeout:            checkTimeout,
		Concurrency:        checkConcurrency,
		InsecureSkipVerify: checkInsecure,
	}

	fmt.Printf("=== Connectivity Check ===\n")
	fmt.Printf("Source:     %s\n", source)
	fmt.Printf("Targets:    deriving from %d entries...\n", len(entries))
	fmt.Printf("Timeout:    %s   Concurrency: %d   Insecure TLS: %v\n\n",
		opts.Timeout, opts.Concurrency, opts.InsecureSkipVerify)

	ctx := context.Background()
	results, summary := connchk.Run(ctx, entries, opts)

	if len(results) == 0 {
		fmt.Println("No dialable targets found (only CIDR ranges or unsubstituted <tenant> values).")
		fmt.Println("Tip: pass --tenant <name> to substitute placeholders before probing.")
		return nil
	}

	renderResults(results)
	renderSummary(summary)

	if connchk.HasFailures(summary) {
		os.Exit(1)
	}
	return nil
}

func loadCheckInput() ([]byte, string, error) {
	if checkInputFile != "" {
		logVerbose("Reading from file: %s", checkInputFile)
		data, err := fetcher.FetchFromFile(checkInputFile)
		if err != nil {
			return nil, "", fmt.Errorf("failed to read file: %w", err)
		}
		return data, checkInputFile, nil
	}

	url := checkInputURL
	if url == "" {
		url = defaultNetworkReqsURL
		logInfo("No file or URL specified, using default: %s", url)
	}
	logVerbose("Fetching from URL: %s", url)
	data, err := fetcher.FetchFromURL(url)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch from URL: %w", err)
	}
	return data, url, nil
}

func filterByService(entries []types.NetworkEntry, service string) []types.NetworkEntry {
	target := strings.ToLower(strings.TrimSpace(service))
	out := make([]types.NetworkEntry, 0, len(entries))
	for _, entry := range entries {
		if strings.EqualFold(entry.Service, target) {
			out = append(out, entry)
		}
	}
	return out
}

// renderResults prints one block per (service, region) group with the per-target
// DNS, TCP, and TLS findings. The style mirrors the ✓/✗/⚠ markers used by the
// validate subcommand.
func renderResults(results []connchk.ProbeResult) {
	type groupKey struct{ svc, region, direction string }

	grouped := make(map[groupKey][]connchk.ProbeResult)
	order := make([]groupKey, 0)
	for _, r := range results {
		k := groupKey{r.Service, r.Region, r.Direction}
		if _, ok := grouped[k]; !ok {
			order = append(order, k)
		}
		grouped[k] = append(grouped[k], r)
	}
	sort.SliceStable(order, func(i, j int) bool {
		if order[i].svc != order[j].svc {
			return order[i].svc < order[j].svc
		}
		return order[i].region < order[j].region
	})

	for _, k := range order {
		fmt.Printf("[%s] %s / %s\n", k.direction, k.svc, k.region)
		for _, r := range grouped[k] {
			renderTarget(r)
		}
		fmt.Println()
	}
}

func renderTarget(r connchk.ProbeResult) {
	fmt.Printf("  %s\n", r.Target)

	switch r.DNS.Status {
	case connchk.StatusOK:
		fmt.Printf("    ✓ DNS: %s   (%s)\n", strings.Join(r.DNS.Addresses, ", "), formatDuration(r.DNS.Duration))
	case connchk.StatusFail:
		fmt.Printf("    ✗ DNS: %s\n", r.DNS.Err)
	case connchk.StatusSkipped:
		fmt.Printf("    – DNS: skipped (IP literal)\n")
	}

	for _, tcp := range r.TCP {
		switch tcp.Status {
		case connchk.StatusOK:
			fmt.Printf("    ✓ TCP %d reachable   (%s)\n", tcp.Port, formatDuration(tcp.Duration))
		case connchk.StatusFail:
			fmt.Printf("    ✗ TCP %d BLOCKED: %s\n", tcp.Port, tcp.Err)
		}

		if tcp.TLS == nil {
			continue
		}
		switch tcp.TLS.Status {
		case connchk.StatusOK:
			cert := ""
			if tcp.TLS.CertSubject != "" {
				cert = fmt.Sprintf("   cert: %s (issuer: %s)", tcp.TLS.CertSubject, tcp.TLS.CertIssuer)
			}
			fmt.Printf("    ✓ TLS %d handshake OK   (%s)%s\n", tcp.Port, formatDuration(tcp.TLS.Duration), cert)
		case connchk.StatusFail:
			fmt.Printf("    ✗ TLS %d FAILED: %s\n", tcp.Port, tcp.TLS.Err)
		case connchk.StatusSkipped:
			fmt.Printf("    ⚠ TLS %d: %s\n", tcp.Port, tcp.TLS.Err)
		}
	}
}

func renderSummary(s connchk.Summary) {
	fmt.Println("=== Summary ===")
	fmt.Printf("  Targets probed:   %d\n", s.Targets)
	fmt.Printf("  DNS:              %d ok, %d failed\n", s.DNSOK, s.DNSFail)
	fmt.Printf("  TCP:              %d ok, %d failed\n", s.TCPOK, s.TCPFail)
	fmt.Printf("  TLS handshakes:   %d ok, %d failed, %d skipped\n", s.TLSOK, s.TLSFail, s.TLSSkipped)
	if connchk.HasFailures(s) {
		fmt.Println("\nResult: ✗ one or more probes failed — see the report above for details.")
	} else {
		fmt.Println("\nResult: ✓ all probes succeeded.")
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	return fmt.Sprintf("%dms", d.Milliseconds())
}
