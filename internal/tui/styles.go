package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Color palette
	colorPrimary   = lipgloss.Color("39")  // bright blue
	colorSuccess   = lipgloss.Color("76")  // green
	colorDanger    = lipgloss.Color("196") // red
	colorWarning   = lipgloss.Color("220") // yellow
	colorMuted     = lipgloss.Color("241") // grey
	colorOutbound  = lipgloss.Color("39")  // blue
	colorInbound   = lipgloss.Color("213") // magenta
	colorHighlight = lipgloss.Color("15")  // white

	// App-level styles
	styleAppTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorHighlight)

	styleHeaderBar = lipgloss.NewStyle().
			Foreground(colorMuted).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(colorMuted)

	styleStatusBar = lipgloss.NewStyle().
			Foreground(colorMuted)

	styleHelp = lipgloss.NewStyle().
			Foreground(colorMuted).
			Italic(true)

	// Tab styles
	styleTabActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorHighlight).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderTop(false).
			BorderLeft(false).
			BorderRight(false).
			BorderForeground(colorPrimary).
			Padding(0, 1)

	styleTabInactive = lipgloss.NewStyle().
				Foreground(colorMuted).
				BorderStyle(lipgloss.HiddenBorder()).
				BorderBottom(true).
				BorderTop(false).
				BorderLeft(false).
				BorderRight(false).
				Padding(0, 1)

	// Direction badges
	styleOutbound = lipgloss.NewStyle().
			Foreground(colorOutbound).
			Bold(true)

	styleInbound = lipgloss.NewStyle().
			Foreground(colorInbound).
			Bold(true)

	// Diff entry styles
	styleDiffAdded = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)

	styleDiffRemoved = lipgloss.NewStyle().
				Foreground(colorDanger).
				Bold(true)

	styleDiffModified = lipgloss.NewStyle().
				Foreground(colorWarning).
				Bold(true)

	// Detail panel styles
	styleDetailLabel = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorPrimary).
				Width(14)

	styleDetailValue = lipgloss.NewStyle().
				Foreground(colorHighlight)

	styleDetailSection = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorMuted).
				MarginTop(1)

	styleDetailIP = lipgloss.NewStyle().
			Foreground(colorSuccess)

	// Status message styles
	styleSuccess = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)

	styleError = lipgloss.NewStyle().
			Foreground(colorDanger).
			Bold(true)
)

// tabLabel renders a tab with active/inactive styling.
func tabLabel(label string, active bool) string {
	if active {
		return styleTabActive.Render(label)
	}
	return styleTabInactive.Render(label)
}

// directionStyle returns the correct style for a direction string.
func directionStyle(direction string) lipgloss.Style {
	if direction == "outbound" {
		return styleOutbound
	}
	return styleInbound
}
