package theme

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Name    string
	Added   lipgloss.Color
	Deleted lipgloss.Color
	Accent  lipgloss.Color
	Muted   lipgloss.Color
}

func Default() Theme {
	return Theme{
		Name:    "default",
		Added:   lipgloss.Color("#F25FB0"),
		Deleted: lipgloss.Color("#49C7EC"),
		Accent:  lipgloss.Color("#B58CF2"),
		Muted:   lipgloss.Color("#6C7086"),
	}
}

func Nerv() Theme {
	return Theme{
		Name:    "nerv",
		Added:   lipgloss.Color("#43D675"),
		Deleted: lipgloss.Color("#FF4A3D"),
		Accent:  lipgloss.Color("#FFB000"),
		Muted:   lipgloss.Color("#7A6B3F"),
	}
}

func Presets() []Theme {
	return []Theme{Default(), Nerv()}
}
