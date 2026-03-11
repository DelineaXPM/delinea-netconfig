package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
)

// FilePickerModel wraps the bubbles filepicker for the initial file selection screen.
type FilePickerModel struct {
	filepicker filepicker.Model
	width      int
	height     int
	selected   string // set when a file is chosen
	err        error
}

// NewFilePickerModel creates a FilePickerModel starting in the current directory.
func NewFilePickerModel(width, height int) FilePickerModel {
	fp := filepicker.New()
	fp.AllowedTypes = []string{".json"}
	fp.CurrentDirectory = "."
	fp.Height = height - 4
	fp.ShowHidden = false
	return FilePickerModel{
		filepicker: fp,
		width:      width,
		height:     height,
	}
}

func (m FilePickerModel) Init() tea.Cmd {
	return m.filepicker.Init()
}

func (m FilePickerModel) Update(msg tea.Msg) (FilePickerModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.filepicker.Height = msg.Height - 4
	}

	var cmd tea.Cmd
	m.filepicker, cmd = m.filepicker.Update(msg)

	if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
		m.selected = path
	}
	if didSelect, path := m.filepicker.DidSelectDisabledFile(msg); didSelect {
		m.err = fmt.Errorf("%s is not a .json file", path)
	}

	return m, cmd
}

func (m FilePickerModel) View() string {
	title := styleAppTitle.Render("Select a network-requirements.json file")
	help := styleHelp.Render("j/k navigate  ·  enter select  ·  q quit")

	errMsg := ""
	if m.err != nil {
		errMsg = "\n" + styleError.Render(m.err.Error())
	}

	return fmt.Sprintf("%s\n%s\n%s%s",
		title,
		m.filepicker.View(),
		styleStatusBar.Render(help),
		errMsg,
	)
}
