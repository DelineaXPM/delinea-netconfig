package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/DelineaXPM/delinea-platform/delinea-netconfig/pkg/types"
)

// EntryItem wraps a NetworkEntry to satisfy the bubbles list.Item interface.
type EntryItem struct {
	Entry types.NetworkEntry
}

// FilterValue returns the string used by the list's built-in filter.
func (i EntryItem) FilterValue() string {
	return i.Entry.Service + " " + i.Entry.Region + " " + i.Entry.Type + " " + strings.Join(i.Entry.Values, " ")
}

// Title returns the primary display line (service name with direction badge).
func (i EntryItem) Title() string {
	dir := directionStyle(i.Entry.Direction).Render(i.Entry.Direction)
	return fmt.Sprintf("[%s] %s", dir, i.Entry.Service)
}

// Description returns the secondary display line (region, type, protocol:ports).
func (i EntryItem) Description() string {
	ports := FormatPorts(i.Entry)
	values := FormatValues(i.Entry.Values, 2)
	return fmt.Sprintf("%s  |  %s  |  %s  |  %s", i.Entry.Region, i.Entry.Type, ports, values)
}

// NewEntryList converts a slice of NetworkEntry into a slice of list.Item.
func NewEntryList(entries []types.NetworkEntry) []list.Item {
	items := make([]list.Item, len(entries))
	for i, e := range entries {
		items[i] = EntryItem{Entry: e}
	}
	return items
}

// FormatPorts returns a compact port string like "tcp:443" or "tcp:443,8443".
func FormatPorts(entry types.NetworkEntry) string {
	if len(entry.Ports) == 0 {
		if entry.Protocol != "" {
			return entry.Protocol
		}
		return "-"
	}
	portStrs := make([]string, len(entry.Ports))
	for i, p := range entry.Ports {
		portStrs[i] = fmt.Sprintf("%d", p)
	}
	proto := entry.Protocol
	if proto == "" {
		proto = "tcp"
	}
	return proto + ":" + strings.Join(portStrs, ",")
}

// FormatValues returns a compact values string truncated to max entries.
// E.g. "192.168.1.0/24, 10.0.0.0/8, ... (+3 more)"
func FormatValues(values []string, max int) string {
	if len(values) == 0 {
		return "-"
	}
	if len(values) <= max {
		return strings.Join(values, ", ")
	}
	shown := strings.Join(values[:max], ", ")
	return fmt.Sprintf("%s, ... (+%d more)", shown, len(values)-max)
}
