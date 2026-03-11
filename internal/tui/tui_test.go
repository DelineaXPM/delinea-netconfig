package tui

import (
	"strings"
	"testing"

	"github.com/DelineaXPM/delinea-platform/delinea-netconfig/internal/differ"
	"github.com/DelineaXPM/delinea-platform/delinea-netconfig/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestEntryItem_FilterValue(t *testing.T) {
	item := EntryItem{Entry: types.NetworkEntry{
		Service: "platform_ssc",
		Region:  "us",
		Type:    "ipv4",
		Values:  []string{"192.168.1.0/24"},
	}}
	fv := item.FilterValue()
	assert.Contains(t, fv, "platform_ssc")
	assert.Contains(t, fv, "us")
	assert.Contains(t, fv, "ipv4")
	assert.Contains(t, fv, "192.168.1.0/24")
}

func TestEntryItem_Title(t *testing.T) {
	item := EntryItem{Entry: types.NetworkEntry{
		Service:   "my_service",
		Direction: "outbound",
	}}
	title := item.Title()
	assert.Contains(t, title, "my_service")
	assert.Contains(t, title, "outbound")
}

func TestEntryItem_Description(t *testing.T) {
	item := EntryItem{Entry: types.NetworkEntry{
		Region:   "eu",
		Type:     "ipv4",
		Protocol: "tcp",
		Ports:    []int{443},
		Values:   []string{"10.0.0.0/8"},
	}}
	desc := item.Description()
	assert.Contains(t, desc, "eu")
	assert.Contains(t, desc, "ipv4")
	assert.Contains(t, desc, "tcp:443")
	assert.Contains(t, desc, "10.0.0.0/8")
}

func TestNewEntryList(t *testing.T) {
	entries := []types.NetworkEntry{
		{Service: "svc1", Direction: "outbound"},
		{Service: "svc2", Direction: "inbound"},
	}
	items := NewEntryList(entries)
	assert.Len(t, items, 2)
	assert.Equal(t, "svc1", items[0].(EntryItem).Entry.Service)
	assert.Equal(t, "svc2", items[1].(EntryItem).Entry.Service)
}

func TestFormatPorts(t *testing.T) {
	tests := []struct {
		name     string
		entry    types.NetworkEntry
		expected string
	}{
		{
			name:     "single port tcp",
			entry:    types.NetworkEntry{Protocol: "tcp", Ports: []int{443}},
			expected: "tcp:443",
		},
		{
			name:     "multiple ports",
			entry:    types.NetworkEntry{Protocol: "tcp", Ports: []int{443, 8443, 80}},
			expected: "tcp:443,8443,80",
		},
		{
			name:     "no ports with protocol",
			entry:    types.NetworkEntry{Protocol: "both", Ports: []int{}},
			expected: "both",
		},
		{
			name:     "no ports no protocol",
			entry:    types.NetworkEntry{},
			expected: "-",
		},
		{
			name:     "no protocol defaults to tcp",
			entry:    types.NetworkEntry{Ports: []int{80}},
			expected: "tcp:80",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, FormatPorts(tt.entry))
		})
	}
}

func TestFormatValues(t *testing.T) {
	tests := []struct {
		name     string
		values   []string
		max      int
		expected string
	}{
		{
			name:     "empty values",
			values:   []string{},
			max:      2,
			expected: "-",
		},
		{
			name:     "values within max",
			values:   []string{"192.168.1.0/24", "10.0.0.0/8"},
			max:      2,
			expected: "192.168.1.0/24, 10.0.0.0/8",
		},
		{
			name:     "single value",
			values:   []string{"192.168.1.0/24"},
			max:      2,
			expected: "192.168.1.0/24",
		},
		{
			name:     "truncated with overflow",
			values:   []string{"a", "b", "c", "d", "e"},
			max:      2,
			expected: "a, b, ... (+3 more)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatValues(tt.values, tt.max)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRenderDetailContent(t *testing.T) {
	entry := types.NetworkEntry{
		Service:     "test_service",
		Direction:   "outbound",
		Region:      "us",
		Protocol:    "tcp",
		Ports:       []int{443},
		Type:        "ipv4",
		Values:      []string{"192.168.1.0/24", "10.0.0.0/8"},
		Description: "A test service description",
	}
	content := renderDetailContent(entry)
	stripped := stripANSI(content)

	assert.Contains(t, stripped, "test_service")
	assert.Contains(t, stripped, "outbound")
	assert.Contains(t, stripped, "us")
	assert.Contains(t, stripped, "tcp")
	assert.Contains(t, stripped, "443")
	assert.Contains(t, stripped, "ipv4")
	assert.Contains(t, stripped, "192.168.1.0/24")
	assert.Contains(t, stripped, "10.0.0.0/8")
	assert.Contains(t, stripped, "A test service description")
}

func TestRenderDiffContent(t *testing.T) {
	result := differ.DiffResult{
		Added: []types.NetworkEntry{
			{Service: "new_svc", Direction: "outbound", Region: "us", Values: []string{"1.2.3.0/24"}, Protocol: "tcp", Ports: []int{443}},
		},
		Removed: []types.NetworkEntry{
			{Service: "old_svc", Direction: "outbound", Region: "eu", Values: []string{"5.6.7.0/24"}, Protocol: "tcp", Ports: []int{80}},
		},
		Modified: []types.NetworkEntry{
			{Service: "mod_svc", Direction: "outbound", Region: "global", Values: []string{"9.9.9.9/32"}, Protocol: "tcp", Ports: []int{8443}},
		},
	}

	allContent := stripANSI(renderDiffContent(result, diffTabAll))
	assert.Contains(t, allContent, "new_svc")
	assert.Contains(t, allContent, "old_svc")
	assert.Contains(t, allContent, "mod_svc")
	assert.Contains(t, allContent, "+")
	assert.Contains(t, allContent, "-")
	assert.Contains(t, allContent, "~")

	addedContent := stripANSI(renderDiffContent(result, diffTabAdded))
	assert.Contains(t, addedContent, "new_svc")
	assert.NotContains(t, addedContent, "old_svc")

	removedContent := stripANSI(renderDiffContent(result, diffTabRemoved))
	assert.Contains(t, removedContent, "old_svc")
	assert.NotContains(t, removedContent, "new_svc")
}

func TestApplyFilters_DirectionTab(t *testing.T) {
	entries := []types.NetworkEntry{
		{Service: "svc1", Direction: "outbound", Region: "us"},
		{Service: "svc2", Direction: "inbound", Region: "us"},
		{Service: "svc3", Direction: "outbound", Region: "eu"},
	}

	tests := []struct {
		name      string
		tab       directionTab
		wantCount int
		wantSvcs  []string
	}{
		{"all entries", dirTabAll, 3, []string{"svc1", "svc2", "svc3"}},
		{"outbound only", dirTabOutbound, 2, []string{"svc1", "svc3"}},
		{"inbound only", dirTabInbound, 1, []string{"svc2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewBrowserModel(entries, 80, 24)
			m.activeTab = tt.tab
			m.applyFilters()
			got := m.FilteredEntries()
			assert.Len(t, got, tt.wantCount)
			var svcs []string
			for _, e := range got {
				svcs = append(svcs, e.Service)
			}
			for _, want := range tt.wantSvcs {
				assert.Contains(t, svcs, want)
			}
		})
	}
}

func TestApplyFilters_RegionFilter(t *testing.T) {
	entries := []types.NetworkEntry{
		{Service: "svc_us", Direction: "outbound", Region: "us"},
		{Service: "svc_eu", Direction: "outbound", Region: "eu"},
		{Service: "svc_global", Direction: "outbound", Region: "global"},
		{Service: "svc_ap", Direction: "outbound", Region: "ap"},
	}

	tests := []struct {
		name         string
		activeRegion string
		wantSvcs     []string
		notWantSvcs  []string
	}{
		{
			name:         "no region filter includes all",
			activeRegion: "",
			wantSvcs:     []string{"svc_us", "svc_eu", "svc_global", "svc_ap"},
		},
		{
			name:         "filter by us includes global",
			activeRegion: "us",
			wantSvcs:     []string{"svc_us", "svc_global"},
			notWantSvcs:  []string{"svc_eu", "svc_ap"},
		},
		{
			name:         "filter by eu includes global",
			activeRegion: "eu",
			wantSvcs:     []string{"svc_eu", "svc_global"},
			notWantSvcs:  []string{"svc_us", "svc_ap"},
		},
		{
			name:         "filter by global includes all global entries only",
			activeRegion: "global",
			wantSvcs:     []string{"svc_us", "svc_eu", "svc_global", "svc_ap"},
		},
		{
			name:         "region filter is case-insensitive",
			activeRegion: "US",
			wantSvcs:     []string{"svc_us", "svc_global"},
			notWantSvcs:  []string{"svc_eu", "svc_ap"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewBrowserModel(entries, 80, 24)
			m.activeRegion = tt.activeRegion
			m.applyFilters()
			got := m.FilteredEntries()
			var svcs []string
			for _, e := range got {
				svcs = append(svcs, e.Service)
			}
			for _, want := range tt.wantSvcs {
				assert.Contains(t, svcs, want)
			}
			for _, notWant := range tt.notWantSvcs {
				assert.NotContains(t, svcs, notWant)
			}
		})
	}
}

func TestApplyFilters_Combined(t *testing.T) {
	entries := []types.NetworkEntry{
		{Service: "out_us", Direction: "outbound", Region: "us"},
		{Service: "in_us", Direction: "inbound", Region: "us"},
		{Service: "out_global", Direction: "outbound", Region: "global"},
		{Service: "in_global", Direction: "inbound", Region: "global"},
		{Service: "out_eu", Direction: "outbound", Region: "eu"},
	}

	m := NewBrowserModel(entries, 80, 24)
	m.activeTab = dirTabOutbound
	m.activeRegion = "us"
	m.applyFilters()
	got := m.FilteredEntries()

	var svcs []string
	for _, e := range got {
		svcs = append(svcs, e.Service)
	}
	assert.Contains(t, svcs, "out_us")
	assert.Contains(t, svcs, "out_global")
	assert.NotContains(t, svcs, "in_us")
	assert.NotContains(t, svcs, "in_global")
	assert.NotContains(t, svcs, "out_eu")
}

func TestRenderDiffContent_ModifiedTab(t *testing.T) {
	result := differ.DiffResult{
		Added:    []types.NetworkEntry{{Service: "added_svc", Direction: "outbound", Region: "us", Protocol: "tcp", Ports: []int{443}}},
		Modified: []types.NetworkEntry{{Service: "mod_svc", Direction: "outbound", Region: "global", Protocol: "tcp", Ports: []int{8443}}},
	}

	modContent := stripANSI(renderDiffContent(result, diffTabModified))
	assert.Contains(t, modContent, "mod_svc")
	assert.NotContains(t, modContent, "added_svc")
	assert.Contains(t, modContent, "~")
}

func TestRenderDiffContent_EmptyResult(t *testing.T) {
	result := differ.DiffResult{}
	content := stripANSI(renderDiffContent(result, diffTabAll))
	assert.Contains(t, content, "No differences found")
}

func TestFormatValues_ZeroMax(t *testing.T) {
	// max=0 means show nothing, just overflow count
	result := FormatValues([]string{"a", "b", "c"}, 0)
	assert.Contains(t, result, "+3 more")
}

// stripANSI removes ANSI escape codes from a string for assertion purposes.
func stripANSI(s string) string {
	var result strings.Builder
	inEscape := false
	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if s[i] == 'm' {
				inEscape = false
			}
			continue
		}
		result.WriteByte(s[i])
	}
	return result.String()
}
