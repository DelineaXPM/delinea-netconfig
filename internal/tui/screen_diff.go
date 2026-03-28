package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/DelineaXPM/delinea-netconfig/internal/differ"
	"github.com/DelineaXPM/delinea-netconfig/pkg/types"
)

// diffTab represents which subset of diff results is shown.
type diffTab int

const (
	diffTabAll      diffTab = iota
	diffTabAdded
	diffTabRemoved
	diffTabModified
)

var diffTabLabels = []string{"All", "Added", "Removed", "Modified"}

// DiffModel shows a tabbed diff view between two network requirements files.
type DiffModel struct {
	result       differ.DiffResult
	activeTab    diffTab
	viewport     viewport.Model
	file1Label   string
	file2Label   string
	callerScreen screen
	readyForBack bool
	width        int
	height       int
}

// NewDiffModel creates a DiffModel from two sets of entries.
// caller is the screen to return to when pressing Esc; use -1 for top-level (quit on Esc).
func NewDiffModel(old, new []types.NetworkEntry, file1Label, file2Label string, width, height int) DiffModel {
	result := differ.Compare(old, new)
	vp := viewport.New(width, height-6)
	m := DiffModel{
		result:       result,
		activeTab:    diffTabAll,
		viewport:     vp,
		file1Label:   file1Label,
		file2Label:   file2Label,
		callerScreen: -1,
		width:        width,
		height:       height,
	}
	vp.SetContent(renderDiffContent(result, diffTabAll))
	m.viewport = vp
	return m
}

func (m DiffModel) Init() tea.Cmd {
	return nil
}

func (m DiffModel) Update(msg tea.Msg) (DiffModel, tea.Cmd) {
	m.readyForBack = false
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, diffKeys.Back):
			if m.callerScreen >= 0 {
				m.readyForBack = true
				return m, nil
			}
		case key.Matches(msg, diffKeys.NextTab):
			m.activeTab = (m.activeTab + 1) % 4
			m.viewport.SetContent(renderDiffContent(m.result, m.activeTab))
			m.viewport.GotoTop()
			return m, nil
		case key.Matches(msg, diffKeys.PrevTab):
			m.activeTab = (m.activeTab + 3) % 4 // wrap backwards
			m.viewport.SetContent(renderDiffContent(m.result, m.activeTab))
			m.viewport.GotoTop()
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 6
		m.viewport.SetContent(renderDiffContent(m.result, m.activeTab))
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m DiffModel) View() string {
	// Header
	header := styleAppTitle.Render("Diff: ") +
		styleStatusBar.Render(m.file1Label+" → "+m.file2Label)

	// Tab bar
	var tabParts []string
	for i, label := range diffTabLabels {
		count := m.tabCount(diffTab(i))
		fullLabel := fmt.Sprintf("%s (%d)", label, count)
		tabParts = append(tabParts, tabLabel(fullLabel, diffTab(i) == m.activeTab))
		if i < len(diffTabLabels)-1 {
			tabParts = append(tabParts, "  ")
		}
	}
	tabBar := lipgloss.JoinHorizontal(lipgloss.Bottom, tabParts...)

	// Summary line
	summary := styleStatusBar.Render(fmt.Sprintf(
		"%s  %s  %s",
		styleDiffAdded.Render(fmt.Sprintf("+%d added", len(m.result.Added))),
		styleDiffRemoved.Render(fmt.Sprintf("-%d removed", len(m.result.Removed))),
		styleDiffModified.Render(fmt.Sprintf("~%d modified", len(m.result.Modified))),
	))

	var help string
	if m.callerScreen >= 0 {
		help = helpLine(diffKeys.NextTab, diffKeys.Up, diffKeys.Down, diffKeys.Back, diffKeys.Quit)
	} else {
		help = helpLine(diffKeys.NextTab, diffKeys.Up, diffKeys.Down, diffKeys.Quit)
	}

	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s",
		header,
		tabBar,
		m.viewport.View(),
		summary,
		styleStatusBar.Render(help),
	)
}

func (m DiffModel) tabCount(tab diffTab) int {
	switch tab {
	case diffTabAll:
		return len(m.result.Added) + len(m.result.Removed) + len(m.result.Modified)
	case diffTabAdded:
		return len(m.result.Added)
	case diffTabRemoved:
		return len(m.result.Removed)
	case diffTabModified:
		return len(m.result.Modified)
	}
	return 0
}

// renderDiffContent builds the styled string for the viewport for the given tab.
func renderDiffContent(result differ.DiffResult, tab diffTab) string {
	var sb strings.Builder

	renderEntries := func(prefix string, style func(string) string, entries []types.NetworkEntry) {
		for _, e := range entries {
			line := fmt.Sprintf("%s [%s] %s/%s  %s  %s",
				prefix,
				e.Direction,
				e.Service,
				e.Region,
				e.Type,
				FormatPorts(e),
			)
			sb.WriteString(style(line) + "\n")
			if len(e.Values) > 0 {
				sb.WriteString(styleStatusBar.Render("    "+FormatValues(e.Values, 3)) + "\n")
			}
		}
	}

	added := func(s string) string { return styleDiffAdded.Render(s) }
	removed := func(s string) string { return styleDiffRemoved.Render(s) }
	modified := func(s string) string { return styleDiffModified.Render(s) }

	switch tab {
	case diffTabAll:
		if len(result.Added) > 0 {
			sb.WriteString(styleDiffAdded.Render(fmt.Sprintf("Added (%d):", len(result.Added))) + "\n")
			renderEntries("+", added, result.Added)
			sb.WriteString("\n")
		}
		if len(result.Removed) > 0 {
			sb.WriteString(styleDiffRemoved.Render(fmt.Sprintf("Removed (%d):", len(result.Removed))) + "\n")
			renderEntries("-", removed, result.Removed)
			sb.WriteString("\n")
		}
		if len(result.Modified) > 0 {
			sb.WriteString(styleDiffModified.Render(fmt.Sprintf("Modified (%d):", len(result.Modified))) + "\n")
			renderEntries("~", modified, result.Modified)
			sb.WriteString("\n")
		}
		if len(result.Added)+len(result.Removed)+len(result.Modified) == 0 {
			sb.WriteString(styleSuccess.Render("✓ No differences found") + "\n")
		}
	case diffTabAdded:
		if len(result.Added) == 0 {
			sb.WriteString(styleStatusBar.Render("No added entries") + "\n")
		} else {
			renderEntries("+", added, result.Added)
		}
	case diffTabRemoved:
		if len(result.Removed) == 0 {
			sb.WriteString(styleStatusBar.Render("No removed entries") + "\n")
		} else {
			renderEntries("-", removed, result.Removed)
		}
	case diffTabModified:
		if len(result.Modified) == 0 {
			sb.WriteString(styleStatusBar.Render("No modified entries") + "\n")
		} else {
			renderEntries("~", modified, result.Modified)
		}
	}

	return sb.String()
}
