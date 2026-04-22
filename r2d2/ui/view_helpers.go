package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// renderHeader combines the ASCII art and system metrics into a single block.
func (m MonitorModel) renderHeader(W int, theme Theme) string {
	accentSt := lipgloss.NewStyle().Foreground(theme.CPU).Bold(true)
	ramSt := lipgloss.NewStyle().Foreground(theme.RAM).Bold(true)
	dskSt := lipgloss.NewStyle().Foreground(theme.DSK).Bold(true)
	cyanSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#00E5FF")).Bold(true)
	greenSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#69FF47")).Bold(true)
	whiteSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#E8E8E8"))

	const artW = 20
	reaction := R2Reactions[m.CurrentFace]
	artLines := make([]string, len(reaction.Art))
	copy(artLines, reaction.Art)
	if m.IsBlinking {
		for i, l := range artLines {
			artLines[i] = strings.ReplaceAll(l, "(O)", "(·)")
		}
	}

	artColor := theme.CharAccent
	if m.CurrentFace == "alarm" {
		artColor = lipgloss.Color("#FF1744")
	}
	if m.CurrentFace == "success" {
		artColor = lipgloss.Color("#69FF47")
	}
	artSt := lipgloss.NewStyle().Foreground(artColor)
	artH := len(artLines)

	metricsW := W - artW - 4
	barLen := metricsW - 35 // Increased space for bars
	if barLen < 15 {
		barLen = 15
	}

	cpuLine := fmt.Sprintf("  CPU %s %s", ProgressBar(m.Stats.CPU, barLen, theme.CPU),
		accentSt.Render(fmt.Sprintf("%5.1f%%", m.Stats.CPU)))
	ramLine := fmt.Sprintf("  RAM %s %s", ProgressBar(m.Stats.RAM, barLen, theme.RAM),
		ramSt.Render(fmt.Sprintf("%4.1f/%.0fG", m.Stats.RAMUsed, m.Stats.RAMTotal)))
	dskLine := fmt.Sprintf("  DSK %s %s", ProgressBar(m.Stats.Disk, barLen, theme.DSK),
		dskSt.Render(fmt.Sprintf("%5.1f%%", m.Stats.Disk)))

	metricLines := []string{
		"",
		cpuLine,
		"",
		ramLine,
		"",
		dskLine,
		"",
		cyanSt.Render(fmt.Sprintf("  UP  %s", m.Stats.Uptime)),
		greenSt.Render(fmt.Sprintf("  NET ↑%.1f ↓%.1f KB/s", m.Stats.NetSent, m.Stats.NetRecv)),
		whiteSt.Bold(true).Render("  " + strings.ToUpper(theme.Name)),
	}
	for len(metricLines) < artH {
		metricLines = append(metricLines, "")
	}

	var hdr strings.Builder
	for i := 0; i < artH; i++ {
		aCell := lipgloss.NewStyle().Width(artW).Render(artSt.Render(artLines[i]))
		mCell := lipgloss.NewStyle().Width(metricsW).Render(metricLines[i])
		hdr.WriteString(aCell + "  " + mCell)
		if i < artH-1 {
			hdr.WriteString("\n")
		}
	}
	return hdr.String()
}

// renderDialogue renders the R2-D2 interaction box.
func (m MonitorModel) renderDialogue(W int) string {
	dlText := m.DisplayMsg
	isLocked := time.Now().Before(m.MsgLockedUntil)
	isError := strings.HasPrefix(m.TargetMsg, "*ERROR*")

	if m.Searching {
		dlText = "SEARCH › " + m.Search + "█"
	} else if len(m.DisplayMsg) < len(m.TargetMsg) {
		dlText += "█"
	}

	prefix := "R2 › "
	if isLocked {
		remaining := int(time.Until(m.MsgLockedUntil).Seconds()) + 1
		prefix = fmt.Sprintf("LOCK %ds › ", remaining)
	}

	style := lipgloss.NewStyle().
		Background(lipgloss.Color("#0D1117")).Foreground(lipgloss.Color("#C9D1D9")).
		Italic(true).Width(W).Padding(0, 1)

	if isLocked && isError {
		style = style.Background(lipgloss.Color("#4A0000")).Foreground(lipgloss.Color("#FF6B6B")).Bold(true)
	}

	// Add vertical padding to the dialogue for a more balanced feel
	return "\n" + style.Render(prefix+dlText) + "\n"
}
