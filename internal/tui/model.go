package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/DelineaXPM/delinea-netconfig/internal/fetcher"
	"github.com/DelineaXPM/delinea-netconfig/internal/parser"
	"github.com/DelineaXPM/delinea-netconfig/pkg/types"
)

// screen represents which UI screen is currently active.
type screen int

const (
	screenFilePicker screen = iota
	screenBrowser
	screenDetail
	screenExport
	screenDiff
	screenDiffPicker
)

// Config holds the startup configuration passed from the CLI command.
type Config struct {
	FilePath string // -f flag: load directly into browser
	URL      string // -u flag: fetch from URL directly into browser
	DiffMode bool   // --diff flag: show diff view
	File1    string // first file for diff mode
	File2    string // second file for diff mode
}

// AppModel is the top-level Bubble Tea model managing the active screen.
type AppModel struct {
	config        Config
	currentScreen screen

	// Sub-models (only the active one is updated/rendered)
	filePicker FilePickerModel
	browser    BrowserModel
	detail     DetailModel
	export     ExportModel
	diff       DiffModel
	diffPicker FilePickerModel

	// Shared state
	entries     []types.NetworkEntry
	networkReqs *types.NetworkRequirements
	sourceLabel string

	// Layout
	width  int
	height int

	// Error state
	err error
}

// New creates the AppModel from a Config.
// All sub-model initialization happens here so that Init() only returns a Cmd.
func New(cfg Config) AppModel {
	m := AppModel{
		config: cfg,
		width:  80,
		height: 24,
	}

	if cfg.DiffMode {
		old, _, err1 := loadFile(cfg.File1)
		if err1 != nil {
			m.err = err1
			return m
		}
		newEntries, _, err2 := loadFile(cfg.File2)
		if err2 != nil {
			m.err = err2
			return m
		}
		m.diff = NewDiffModel(old, newEntries, cfg.File1, cfg.File2, m.width, m.height)
		m.currentScreen = screenDiff
		return m
	}

	if cfg.URL != "" {
		entries, reqs, err := loadURL(cfg.URL)
		if err != nil {
			m.err = err
			return m
		}
		m.entries = entries
		m.networkReqs = reqs
		m.sourceLabel = cfg.URL
		m.browser = NewBrowserModel(entries, m.width, m.height)
		m.currentScreen = screenBrowser
		return m
	}

	if cfg.FilePath != "" {
		entries, reqs, err := loadFile(cfg.FilePath)
		if err != nil {
			m.err = err
			return m
		}
		m.entries = entries
		m.networkReqs = reqs
		m.sourceLabel = cfg.FilePath
		m.browser = NewBrowserModel(entries, m.width, m.height)
		m.currentScreen = screenBrowser
		return m
	}

	// No file specified: show file picker
	m.filePicker = NewFilePickerModel(m.width, m.height)
	m.currentScreen = screenFilePicker
	return m
}

// Init returns the initial command for the active screen.
func (m AppModel) Init() tea.Cmd {
	if m.err != nil {
		return tea.Quit
	}
	switch m.currentScreen {
	case screenFilePicker:
		return m.filePicker.Init()
	case screenBrowser:
		return m.browser.Init()
	case screenDiff:
		return m.diff.Init()
	}
	return nil
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Global handlers
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Propagate size to active sub-model
		switch m.currentScreen {
		case screenFilePicker:
			m.filePicker, _ = m.filePicker.Update(msg)
		case screenBrowser:
			m.browser, _ = m.browser.Update(msg)
		case screenDetail:
			m.detail, _ = m.detail.Update(msg)
		case screenDiff:
			m.diff, _ = m.diff.Update(msg)
		case screenDiffPicker:
			m.diffPicker, _ = m.diffPicker.Update(msg)
		}
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == "q" && m.currentScreen != screenExport && m.currentScreen != screenDiffPicker {
			return m, tea.Quit
		}
	}

	// Route to active sub-model
	switch m.currentScreen {
	case screenFilePicker:
		return m.updateFilePicker(msg)
	case screenBrowser:
		return m.updateBrowser(msg)
	case screenDetail:
		return m.updateDetail(msg)
	case screenExport:
		return m.updateExport(msg)
	case screenDiff:
		return m.updateDiff(msg)
	case screenDiffPicker:
		return m.updateDiffPicker(msg)
	}

	return m, nil
}

func (m AppModel) View() string {
	if m.err != nil {
		return styleError.Render("Error: "+m.err.Error()) + "\n"
	}

	header := ""
	if m.currentScreen == screenBrowser || m.currentScreen == screenDetail {
		header = m.headerBar() + "\n"
	}

	switch m.currentScreen {
	case screenFilePicker:
		return m.filePicker.View()
	case screenBrowser:
		return header + m.browser.View()
	case screenDetail:
		return header + m.detail.View()
	case screenExport:
		return m.export.View()
	case screenDiff:
		return m.diff.View()
	case screenDiffPicker:
		return m.diffPicker.View()
	}

	return ""
}

// headerBar renders the top bar showing file info and version.
func (m AppModel) headerBar() string {
	src := styleStatusBar.Render("File: " + m.sourceLabel)
	ver := ""
	if m.networkReqs != nil {
		ver = styleStatusBar.Render("  │  Version: " + m.networkReqs.Version)
		if m.networkReqs.UpdatedAt != "" {
			ver += styleStatusBar.Render("  │  Updated: " + m.networkReqs.UpdatedAt)
		}
	}
	return styleHeaderBar.Width(m.width).Render(styleAppTitle.Render("delinea-netconfig") + "  " + src + ver)
}

// --- Screen update helpers ---

func (m AppModel) updateFilePicker(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.filePicker, cmd = m.filePicker.Update(msg)

	if m.filePicker.selected != "" {
		path := m.filePicker.selected
		entries, reqs, err := loadFile(path)
		if err != nil {
			m.err = err
			return m, tea.Quit
		}
		m.entries = entries
		m.networkReqs = reqs
		m.sourceLabel = path
		m.browser = NewBrowserModel(entries, m.width, m.height)
		m.currentScreen = screenBrowser
		return m, m.browser.Init()
	}

	return m, cmd
}

func (m AppModel) updateBrowser(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.browser, cmd = m.browser.Update(msg)

	if m.browser.readyForDetail {
		if entry, ok := m.browser.SelectedEntry(); ok {
			m.detail = NewDetailModel(entry, m.width, m.height)
			m.currentScreen = screenDetail
			return m, m.detail.Init()
		}
	}

	if m.browser.readyForDiff {
		m.diffPicker = NewFilePickerModel(m.width, m.height)
		m.currentScreen = screenDiffPicker
		return m, m.diffPicker.Init()
	}

	if m.browser.readyForExport {
		m.export = NewExportModel(
			m.entries,
			m.browser.FilteredEntries(),
			nil,
			screenBrowser,
			m.width, m.height,
		)
		m.currentScreen = screenExport
		return m, m.export.Init()
	}

	return m, cmd
}

func (m AppModel) updateDetail(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.detail, cmd = m.detail.Update(msg)

	if m.detail.readyForBack {
		m.currentScreen = screenBrowser
		return m, nil
	}

	if m.detail.readyForExport {
		entry := m.detail.entry
		m.export = NewExportModel(
			m.entries,
			m.browser.FilteredEntries(),
			&entry,
			screenDetail,
			m.width, m.height,
		)
		m.currentScreen = screenExport
		return m, m.export.Init()
	}

	return m, cmd
}

func (m AppModel) updateExport(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Any key press when done returns to caller screen
	if m.export.done {
		if _, ok := msg.(tea.KeyMsg); ok {
			m.currentScreen = m.export.callerScreen
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.export, cmd = m.export.Update(msg)

	// Form aborted (Esc pressed)
	if m.export.done && m.export.statusMsg == "" {
		m.currentScreen = m.export.callerScreen
		return m, nil
	}

	return m, cmd
}

func (m AppModel) updateDiff(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.diff, cmd = m.diff.Update(msg)

	if m.diff.readyForBack && m.diff.callerScreen >= 0 {
		m.currentScreen = m.diff.callerScreen
		return m, nil
	}

	return m, cmd
}

func (m AppModel) updateDiffPicker(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Esc returns to browser
	if kmsg, ok := msg.(tea.KeyMsg); ok {
		if kmsg.String() == "esc" || kmsg.String() == "q" {
			m.currentScreen = screenBrowser
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.diffPicker, cmd = m.diffPicker.Update(msg)

	if m.diffPicker.selected != "" {
		path := m.diffPicker.selected
		newEntries, _, err := loadFile(path)
		if err != nil {
			m.err = err
			return m, tea.Quit
		}
		m.diff = NewDiffModel(m.entries, newEntries, m.sourceLabel, path, m.width, m.height)
		m.diff.callerScreen = screenBrowser
		m.currentScreen = screenDiff
		return m, m.diff.Init()
	}

	return m, cmd
}

// loadFile reads and parses a local file.
func loadFile(path string) ([]types.NetworkEntry, *types.NetworkRequirements, error) {
	data, err := fetcher.FetchFromFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read %s: %w", path, err)
	}
	return parseData(data, path)
}

// loadURL fetches and parses a remote URL.
func loadURL(url string) ([]types.NetworkEntry, *types.NetworkRequirements, error) {
	data, err := fetcher.FetchFromURL(url)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch %s: %w", url, err)
	}
	return parseData(data, url)
}

func parseData(data []byte, source string) ([]types.NetworkEntry, *types.NetworkRequirements, error) {
	reqs, err := parser.Parse(data)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse %s: %w", source, err)
	}
	return parser.Normalize(reqs), reqs, nil
}
