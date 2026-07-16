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
