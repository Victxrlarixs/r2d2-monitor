package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ProgressBar renders a visual bar representation of a percentage value.
func ProgressBar(percent float64, width int, color lipgloss.Color) string {
	if width <= 0 {
		return ""
	}
	f := int((percent / 100.0) * float64(width))
	if f > width {
		f = width
	}
	if f < 0 {
		f = 0
	}
	e := width - f

	filled := lipgloss.NewStyle().Foreground(color).Render(strings.Repeat("█", f))
	empty := lipgloss.NewStyle().Foreground(lipgloss.Color("#333333")).Render(strings.Repeat("░", e))

	return "[" + filled + empty + "]"
}
