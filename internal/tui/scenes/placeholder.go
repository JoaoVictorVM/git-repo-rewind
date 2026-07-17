package scenes

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

type Placeholder struct {
	Name string
}

func (p Placeholder) Title() string { return p.Name }

func (p Placeholder) Render(f Frame) string {
	if f.Width < 1 || f.Height < 1 {
		return ""
	}
	message := fmt.Sprintf("cena %s — em construção", p.Name)
	return lipgloss.NewStyle().
		Foreground(f.Theme.Muted).
		Width(f.Width).
		Height(f.Height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(message)
}
