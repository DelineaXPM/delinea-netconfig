package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/DelineaXPM/delinea-platform/delinea-netconfig/pkg/types"
)

// directionTab represents the active direction filter.
type directionTab int

const (
	dirTabAll      directionTab = iota // show all entries
	dirTabOutbound                     // show only outbound
	dirTabInbound                      // show only inbound
)

// BrowserModel is the main scrollable entry list screen.
type BrowserModel struct {
	list              list.Model
	activeTab         directionTab
	allEntries        []types.NetworkEntry
	regionInput       textinput.Model
	regionFilterMode  bool   // true when user is typing a region
	activeRegion      string // current applied region filter (empty = all)
	width             int
	height            int
	readyForDetail    bool
	readyForExport    bool
}

// NewBrowserModel creates a new BrowserModel from a slice of entries.
func NewBrowserModel(entries []types.NetworkEntry, width, height int) BrowserModel {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("39")).
		BorderForeground(lipgloss.Color("39"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("39"))

	l := list.New(NewEntryList(entries), delegate, width, height-5)
	l.SetShowTitle(false)
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.StatusBar = styleStatusBar

	ti := textinput.New()
	ti.Placeholder = "us, eu, global, ap…"
	ti.CharLimit = 32
	ti.Width = 20

	return BrowserModel{
		list:        l,
		activeTab:   dirTabAll,
		allEntries:  entries,
		regionInput: ti,
		width:       width,
		height:      height,
	}
}

func (m BrowserModel) Init() tea.Cmd {
	return nil
}

func (m BrowserModel) Update(msg tea.Msg) (BrowserModel, tea.Cmd) {
	m.readyForDetail = false
	m.readyForExport = false

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Region filter input is active
		if m.regionFilterMode {
			switch msg.String() {
			case "enter":
				m.activeRegion = strings.TrimSpace(m.regionInput.Value())
				m.regionFilterMode = false
				m.regionInput.Blur()
				m.applyFilters()
				return m, nil
			case "esc":
				m.regionFilterMode = false
				m.regionInput.Blur()
				m.regionInput.SetValue(m.activeRegion) // restore previous value
				return m, nil
			}
			var cmd tea.Cmd
			m.regionInput, cmd = m.regionInput.Update(msg)
			return m, cmd
		}

		// Don't handle custom keys when the list's built-in text filter is active
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, browserKeys.Detail):
			if m.list.SelectedItem() != nil {
				m.readyForDetail = true
				return m, nil
			}
		case key.Matches(msg, browserKeys.Export):
			m.readyForExport = true
			return m, nil
		case key.Matches(msg, browserKeys.Tab):
			m.activeTab = (m.activeTab + 1) % 3
			m.applyFilters()
			return m, nil
		case key.Matches(msg, browserKeys.RegionFilter):
			m.regionFilterMode = true
			m.regionInput.SetValue(m.activeRegion)
			return m, m.regionInput.Focus()
		case key.Matches(msg, browserKeys.ClearRegion):
			m.activeRegion = ""
			m.regionInput.SetValue("")
			m.applyFilters()
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-5)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m BrowserModel) View() string {
	// Direction tab bar
	tabs := []string{
		tabLabel("All", m.activeTab == dirTabAll),
		tabLabel("Outbound", m.activeTab == dirTabOutbound),
		tabLabel("Inbound", m.activeTab == dirTabInbound),
	}
	tabBar := ""
	for _, t := range tabs {
		tabBar += t + "  "
	}

	// Region filter line
	regionLine := ""
	if m.regionFilterMode {
		regionLine = styleDetailLabel.Render("Region: ") + m.regionInput.View() +
			styleHelp.Render("  enter confirm  ·  esc cancel") + "\n"
	} else if m.activeRegion != "" {
		regionLine = styleDetailLabel.Render("Region: ") +
			styleDiffAdded.Render(m.activeRegion) +
			styleHelp.Render("  (x to clear)") + "\n"
	}

	// Help bar
	help := helpLine(browserKeys.Up, browserKeys.Down, browserKeys.Detail, browserKeys.Export, browserKeys.Tab, browserKeys.RegionFilter, browserKeys.Quit)

	return fmt.Sprintf("%s\n%s%s\n%s", tabBar, regionLine, m.list.View(), styleStatusBar.Render(help))
}

// SelectedEntry returns the currently highlighted NetworkEntry.
func (m BrowserModel) SelectedEntry() (types.NetworkEntry, bool) {
	item, ok := m.list.SelectedItem().(EntryItem)
	if !ok {
		return types.NetworkEntry{}, false
	}
	return item.Entry, true
}

// FilteredEntries returns the currently visible entries (after all filters).
func (m BrowserModel) FilteredEntries() []types.NetworkEntry {
	items := m.list.Items()
	entries := make([]types.NetworkEntry, 0, len(items))
	for _, item := range items {
		if ei, ok := item.(EntryItem); ok {
			entries = append(entries, ei.Entry)
		}
	}
	return entries
}

// applyFilters rebuilds the list applying both direction and region filters.
func (m *BrowserModel) applyFilters() {
	var filtered []types.NetworkEntry
	for _, e := range m.allEntries {
		// Direction filter
		switch m.activeTab {
		case dirTabOutbound:
			if e.Direction != "outbound" {
				continue
			}
		case dirTabInbound:
			if e.Direction != "inbound" {
				continue
			}
		}

		// Region filter: always include "global"; match specified region case-insensitively
		if m.activeRegion != "" && m.activeRegion != "global" {
			if e.Region != "global" && !strings.EqualFold(e.Region, m.activeRegion) {
				continue
			}
		}

		filtered = append(filtered, e)
	}
	_ = m.list.SetItems(NewEntryList(filtered))
}
