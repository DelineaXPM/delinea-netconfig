package tui

import "github.com/charmbracelet/bubbles/key"

// browserKeyMap defines keys for the main browser screen.
type browserKeyMap struct {
	Up           key.Binding
	Down         key.Binding
	Detail       key.Binding
	Export       key.Binding
	Tab          key.Binding
	RegionFilter key.Binding
	ClearRegion  key.Binding
	Quit         key.Binding
}

// detailKeyMap defines keys for the entry detail screen.
type detailKeyMap struct {
	Back    key.Binding
	Up      key.Binding
	Down    key.Binding
	Export  key.Binding
	CopyIPs key.Binding
}

// diffKeyMap defines keys for the diff screen.
type diffKeyMap struct {
	NextTab key.Binding
	PrevTab key.Binding
	Up      key.Binding
	Down    key.Binding
	Quit    key.Binding
}

var browserKeys = browserKeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Detail: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "detail"),
	),
	Export: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "export"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "direction"),
	),
	RegionFilter: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "region"),
	),
	ClearRegion: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "clear region"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

var detailKeys = detailKeyMap{
	Back: key.NewBinding(
		key.WithKeys("esc", "backspace"),
		key.WithHelp("esc", "back"),
	),
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "scroll up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "scroll down"),
	),
	Export: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "export"),
	),
	CopyIPs: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "copy IPs"),
	),
}

var diffKeys = diffKeyMap{
	NextTab: key.NewBinding(
		key.WithKeys("tab", "right", "l"),
		key.WithHelp("tab/→", "next tab"),
	),
	PrevTab: key.NewBinding(
		key.WithKeys("shift+tab", "left", "h"),
		key.WithHelp("shift+tab/←", "prev tab"),
	),
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "scroll up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "scroll down"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c", "esc"),
		key.WithHelp("q", "quit"),
	),
}


// helpLine returns a formatted help string for the given bindings.
func helpLine(bindings ...key.Binding) string {
	var parts []string
	for _, b := range bindings {
		if b.Help().Key != "" {
			parts = append(parts, styleHelp.Render(b.Help().Key+" "+b.Help().Desc))
		}
	}
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += styleHelp.Render("  ·  ")
		}
		result += p
	}
	return result
}
