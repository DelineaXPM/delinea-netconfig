package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/atotto/clipboard"
	"github.com/DelineaXPM/delinea-platform/delinea-netconfig/pkg/types"
)

// DetailModel shows the full detail of a single NetworkEntry.
type DetailModel struct {
	entry          types.NetworkEntry
	viewport       viewport.Model
	width          int
	height         int
	readyForBack   bool
	readyForExport bool
	statusMsg      string
}

// NewDetailModel creates a DetailModel for the given entry.
func NewDetailModel(entry types.NetworkEntry, width, height int) DetailModel {
	vp := viewport.New(width, height-5)
	vp.SetContent(renderDetailContent(entry))
	return DetailModel{
		entry:    entry,
		viewport: vp,
		width:    width,
		height:   height,
	}
}

func (m DetailModel) Init() tea.Cmd {
	return nil
}

func (m DetailModel) Update(msg tea.Msg) (DetailModel, tea.Cmd) {
	m.readyForBack = false
	m.readyForExport = false

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, detailKeys.Back):
			m.readyForBack = true
			return m, nil
		case key.Matches(msg, detailKeys.Export):
			m.readyForExport = true
			return m, nil
		case key.Matches(msg, detailKeys.CopyIPs):
			ips := strings.Join(m.entry.Values, "\n")
			if err := clipboard.WriteAll(ips); err != nil {
				m.statusMsg = styleError.Render("Failed to copy: " + err.Error())
			} else {
				m.statusMsg = styleSuccess.Render("✓ Copied " + fmt.Sprintf("%d", len(m.entry.Values)) + " IPs to clipboard")
			}
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 5
		m.viewport.SetContent(renderDetailContent(m.entry))
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m DetailModel) View() string {
	title := styleAppTitle.Render("Entry Detail")
	help := helpLine(detailKeys.Back, detailKeys.Export, detailKeys.CopyIPs)
	status := ""
	if m.statusMsg != "" {
		status = "\n" + m.statusMsg
	}
	scrollPct := fmt.Sprintf("  %.0f%%", m.viewport.ScrollPercent()*100)
	return fmt.Sprintf("%s%s\n%s\n%s%s",
		title,
		styleStatusBar.Render(scrollPct),
		m.viewport.View(),
		styleStatusBar.Render(help),
		status,
	)
}

// renderDetailContent builds the styled string content for the viewport.
func renderDetailContent(entry types.NetworkEntry) string {
	var sb strings.Builder

	row := func(label, value string) {
		sb.WriteString(styleDetailLabel.Render(label))
		sb.WriteString(styleDetailValue.Render(value))
		sb.WriteString("\n")
	}

	row("Service:", entry.Service)
	row("Direction:", directionStyle(entry.Direction).Render(entry.Direction))
	row("Region:", entry.Region)
	row("Protocol:", entry.Protocol)
	row("Ports:", FormatPorts(entry))
	row("Type:", entry.Type)
	if entry.Redundancy != "" {
		row("Redundancy:", entry.Redundancy)
	}
	if len(entry.Tags) > 0 {
		row("Tags:", strings.Join(entry.Tags, ", "))
	}

	sb.WriteString("\n")
	sb.WriteString(styleDetailSection.Render("Values:"))
	sb.WriteString("\n")
	for _, v := range entry.Values {
		sb.WriteString("  " + styleDetailIP.Render("• "+v) + "\n")
	}

	if entry.Description != "" {
		sb.WriteString("\n")
		sb.WriteString(styleDetailSection.Render("Description:"))
		sb.WriteString("\n")
		sb.WriteString("  " + styleDetailValue.Render(entry.Description) + "\n")
	}

	return sb.String()
}
