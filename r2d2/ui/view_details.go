package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderInspectionModal renders a detailed overlay for the selected process.
func (m MonitorModel) renderInspectionModal(W, H int, theme Theme) string {
	accentSt := lipgloss.NewStyle().Foreground(theme.CPU).Bold(true)
	boxSt := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.CPU).
		Padding(1, 2).
		Width(W - 10).
		Height(H - 12)

	title := accentSt.Render(fmt.Sprintf(" SCANNING PID %s [%s] ", m.SelectedProcess.ID, m.SelectedProcess.Name))
	
	content := m.Details
	if content == "" {
		content = "No metadata available."
	}

	// Clean up the PowerShell output for the TUI
	lines := strings.Split(content, "\n")
	var cleaned []string
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			cleaned = append(cleaned, " › "+l)
		}
	}
	
	footer := lipgloss.NewStyle().Foreground(theme.CharMain).Italic(true).Render(" [Esc] Close Scan View ")

	modal := boxSt.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			title,
			"",
			strings.Join(cleaned, "\n"),
			"",
			footer,
		),
	)

	return lipgloss.Place(W, H, lipgloss.Center, lipgloss.Center, modal)
}
