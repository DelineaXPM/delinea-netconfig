package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/huh"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/DelineaXPM/delinea-netconfig/internal/converter"
	"github.com/DelineaXPM/delinea-netconfig/pkg/types"
)

// exportDoneMsg is sent when the export operation completes.
type exportDoneMsg struct {
	outputPath string
	err        error
}

// ExportModel shows the huh export form and manages the export lifecycle.
type ExportModel struct {
	form            *huh.Form
	spinner         spinner.Model
	exporting       bool
	done            bool
	statusMsg       string
	isError         bool
	callerScreen    screen
	allEntries      []types.NetworkEntry
	filteredEntries []types.NetworkEntry
	serviceEntry    *types.NetworkEntry

	// Form-bound values
	fFormat     string
	fTenant     string
	fRegion     string
	fOutputFile string
	fScope      string
}

// scopeAll is the form value for "all entries".
const (
	scopeAll      = "all"
	scopeFiltered = "filtered"
	scopeService  = "service"
)

// NewExportModel creates the export form.
func NewExportModel(
	all, filtered []types.NetworkEntry,
	service *types.NetworkEntry,
	caller screen,
	width, height int,
) ExportModel {
	m := ExportModel{
		callerScreen:    caller,
		allEntries:      all,
		filteredEntries: filtered,
		serviceEntry:    service,
		fFormat:         "csv",
		fScope:          scopeAll,
	}

	// Default output file
	m.fOutputFile = "network-rules." + m.formatExt()

	// Build scope options
	scopeOptions := []huh.Option[string]{
		huh.NewOption("All entries", scopeAll),
	}
	if len(filtered) > 0 && len(filtered) < len(all) {
		scopeOptions = append(scopeOptions, huh.NewOption(
			fmt.Sprintf("Current filter (%d entries)", len(filtered)), scopeFiltered,
		))
	}
	if service != nil {
		scopeOptions = append(scopeOptions, huh.NewOption(
			fmt.Sprintf("This service only (%s)", service.Service), scopeService,
		))
	}

	// All supported format names
	formatOptions := []huh.Option[string]{
		huh.NewOption("CSV", "csv"),
		huh.NewOption("YAML", "yaml"),
		huh.NewOption("Terraform (HCL)", "terraform"),
		huh.NewOption("Ansible", "ansible"),
		huh.NewOption("AWS Security Group", "aws-sg"),
		huh.NewOption("Cisco ACL", "cisco"),
		huh.NewOption("PAN-OS XML", "panos"),
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Output format").
				Options(formatOptions...).
				Value(&m.fFormat),

			huh.NewSelect[string]().
				Title("Scope").
				Options(scopeOptions...).
				Value(&m.fScope),

			huh.NewInput().
				Title("Tenant name (optional)").
				Description("Replaces <tenant> in hostnames").
				Placeholder("acme-corp").
				Value(&m.fTenant),

			huh.NewInput().
				Title("Region filter (optional)").
				Description("Leave blank for all regions").
				Placeholder("us, eu, global …").
				Value(&m.fRegion),

			huh.NewInput().
				Title("Output file").
				Value(&m.fOutputFile).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("output file cannot be empty")
					}
					return nil
				}),
		),
	).WithWidth(width).WithHeight(height - 4)

	m.form = form

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	m.spinner = sp

	return m
}

// formatExt returns the file extension for the current format selection.
func (m ExportModel) formatExt() string {
	conv, err := converter.GetConverter(m.fFormat)
	if err != nil {
		return "txt"
	}
	return conv.FileExtension()
}

func (m ExportModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m ExportModel) Update(msg tea.Msg) (ExportModel, tea.Cmd) {
	if m.done {
		return m, nil
	}

	// Handle export completion
	if done, ok := msg.(exportDoneMsg); ok {
		m.exporting = false
		m.done = true
		if done.err != nil {
			m.statusMsg = styleError.Render("Export failed: " + done.err.Error())
			m.isError = true
		} else {
			m.statusMsg = styleSuccess.Render("✓ Exported to " + done.outputPath)
		}
		return m, nil
	}

	// Handle spinner tick while exporting
	if m.exporting {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	// Update the form
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	// Check if form is completed
	if m.form.State == huh.StateCompleted {
		m.exporting = true
		return m, tea.Batch(m.spinner.Tick, m.doExport())
	}

	// Check if form is aborted (Esc)
	if m.form.State == huh.StateAborted {
		m.done = true
		m.statusMsg = ""
		return m, nil
	}

	return m, cmd
}

func (m ExportModel) View() string {
	if m.exporting {
		return fmt.Sprintf("\n  %s Exporting to %s…\n",
			m.spinner.View(),
			styleDetailValue.Render(m.fOutputFile),
		)
	}
	if m.done {
		help := styleHelp.Render("press any key to continue")
		return fmt.Sprintf("\n  %s\n\n  %s\n", m.statusMsg, help)
	}
	return m.form.View()
}

// doExport performs the actual export as a tea.Cmd (runs in goroutine).
func (m ExportModel) doExport() tea.Cmd {
	return func() tea.Msg {
		// Pick entries based on scope
		entries := m.allEntries
		switch m.fScope {
		case scopeFiltered:
			entries = m.filteredEntries
		case scopeService:
			if m.serviceEntry != nil {
				entries = []types.NetworkEntry{*m.serviceEntry}
			}
		}

		// Apply region filter
		if r := strings.TrimSpace(m.fRegion); r != "" && r != "global" {
			var filtered []types.NetworkEntry
			for _, e := range entries {
				if e.Region == "global" || strings.EqualFold(e.Region, r) {
					filtered = append(filtered, e)
				}
			}
			entries = filtered
		}

		// Apply tenant substitution
		if t := strings.TrimSpace(m.fTenant); t != "" {
			for i, e := range entries {
				for j, v := range e.Values {
					entries[i].Values[j] = strings.ReplaceAll(v, "<tenant>", t)
				}
			}
		}

		// Get converter
		conv, err := converter.GetConverter(m.fFormat)
		if err != nil {
			return exportDoneMsg{err: err}
		}

		// Convert
		data, err := conv.Convert(entries)
		if err != nil {
			return exportDoneMsg{err: err}
		}

		// Write file
		outPath := strings.TrimSpace(m.fOutputFile)
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return exportDoneMsg{err: err}
		}
		if err := os.WriteFile(outPath, data, 0644); err != nil {
			return exportDoneMsg{err: err}
		}

		return exportDoneMsg{outputPath: outPath}
	}
}
