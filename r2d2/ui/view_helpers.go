package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ProgressBar renders a visual bar representation of a percentage value.
func ProgressBar(percent float64, width int, color lipgloss.Color) string {
	if width <= 0 { return "" }
	f := int((percent / 100.0) * float64(width))
	if f > width { f = width }
	if f < 0 { f = 0 }
	e := width - f

	filled := lipgloss.NewStyle().Foreground(color).Render(strings.Repeat("█", f))
	empty := lipgloss.NewStyle().Foreground(lipgloss.Color("#333333")).Render(strings.Repeat("·", e))

	return filled + empty
}

// BrailleSparkline renders a high-resolution graph using Braille patterns.
func BrailleSparkline(data []float64, width int, color lipgloss.Color) string {
	if len(data) == 0 { return "" }
	if width > len(data) { width = len(data) }
	plot := data[len(data)-width:]
	var maxVal float64 = 0.1
	for _, v := range plot { if v > maxVal { maxVal = v } }

	chars := []string{" ", "⡀", "⣀", "⣄", "⣤", "⣦", "⣶", "⣷", "⣿"}
	var b strings.Builder
	st := lipgloss.NewStyle().Foreground(color)
	for _, v := range plot {
		idx := int((v / maxVal) * float64(len(chars)-1))
		if idx < 0 { idx = 0 }
		if idx >= len(chars) { idx = len(chars) - 1 }
		b.WriteString(st.Render(chars[idx]))
	}
	return b.String()
}

// RenderBox creates a btop-style box with a border and title.
func RenderBox(width, height int, title string, content string, color lipgloss.Color) string {
	innerW := width - 4
	innerH := height - 2
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(color).
		Width(width - 2).
		Height(height - 2).
		Padding(0, 1)

	lines := strings.Split(content, "\n")
	var clamped strings.Builder
	for i := 0; i < innerH; i++ {
		if i < len(lines) {
			l := lines[i]
			if lipgloss.Width(l) > innerW { l = truncate(l, innerW) }
			clamped.WriteString(l + "\n")
		} else { clamped.WriteString("\n") }
	}

	rendered := style.Render(strings.TrimSuffix(clamped.String(), "\n"))
	boxLines := strings.Split(rendered, "\n")
	if len(boxLines) > 0 && title != "" {
		tStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(color).Bold(true).Padding(0, 1)
		tText := tStyle.Render(title)
		top := "╭─" + tText + strings.Repeat("─", (width-2)-lipgloss.Width(tText)-2) + "╮"
		boxLines[0] = top
	}
	return strings.Join(boxLines, "\n")
}

func (m MonitorModel) renderCPUBox(w, h int, theme Theme) string {
	var b strings.Builder
	cyan := lipgloss.Color("#00E5FF")
	model := truncate(m.Stats.CPUModel, w-28)
	b.WriteString(lipgloss.NewStyle().Foreground(cyan).Bold(true).Render("CPU: " + model) + "\n")
	
	b.WriteString(fmt.Sprintf(" OS: %-25s %s\n", truncate(m.Stats.OSName, 25), m.renderBattery(theme)))
	b.WriteString(strings.Repeat("─", w-4) + "\n")
	cols := 4
	if w < 90 { cols = 2 }
	coreW := (w - 8) / cols
	for i := 0; i < len(m.Stats.CPUCores); i += cols {
		for j := 0; j < cols && i+j < len(m.Stats.CPUCores); j++ {
			idx := i + j
			val := m.Stats.CPUCores[idx]
			b.WriteString(fmt.Sprintf("C%-2d:%s%3.0f%% ", idx, ProgressBar(val, coreW-10, theme.CPU), val))
		}
		b.WriteString("\n")
	}
	return RenderBox(w, h, " CPU ", b.String(), theme.CPU)
}

func (m MonitorModel) renderBattery(theme Theme) string {
	if m.Stats.Battery.Percent == 0 { return "" }
	color := lipgloss.Color("#69FF47")
	if m.Stats.Battery.Percent < 20 { color = lipgloss.Color("#FF1744") }
	pBar := ProgressBar(m.Stats.Battery.Percent, 10, color)
	status := "BATTERY"
	if m.Stats.Battery.Status == "Charging" || m.Stats.Battery.Status == "AC Power" { status = "POWERED" }
	return lipgloss.NewStyle().Foreground(color).Render(fmt.Sprintf("%s %s %d%%", status, pBar, int(m.Stats.Battery.Percent)))
}

func (m MonitorModel) renderMemBox(w, h int, theme Theme) string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf(" RAM  %s %4.1fG\n", ProgressBar(m.Stats.RAM, w-18, theme.RAM), m.Stats.RAMUsed))
	b.WriteString(lipgloss.NewStyle().Foreground(theme.RAM).Render(fmt.Sprintf("      Used: %.1fG / Total: %.1fG\n", m.Stats.RAMUsed, m.Stats.RAMTotal)))
	b.WriteString(lipgloss.NewStyle().Foreground(theme.RAM).Faint(true).Render(fmt.Sprintf("      Avail: %.1fG / Cached: %.1fG\n\n", m.Stats.RAMAvailable, m.Stats.RAMCached)))
	b.WriteString(fmt.Sprintf(" SWAP %s %4.1fG\n", ProgressBar(m.Stats.Swap, w-18, theme.RAM), m.Stats.SwapUsed))
	b.WriteString(lipgloss.NewStyle().Foreground(theme.RAM).Faint(true).Render(fmt.Sprintf("      Used: %.1fG / Total: %.1fG\n", m.Stats.SwapUsed, m.Stats.SwapTotal)))
	return RenderBox(w, h, " MEMORY ", b.String(), theme.RAM)
}

func (m MonitorModel) renderDiskBox(w, h int, theme Theme) string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf(" C:   %s %4.1f%%\n", ProgressBar(m.Stats.Disk, w-18, theme.DSK), m.Stats.Disk))
	b.WriteString(lipgloss.NewStyle().Foreground(theme.DSK).Render(fmt.Sprintf("      Used: %.1fG / Total: %.1fG\n", m.Stats.DiskUsed, m.Stats.DiskTotal)))
	return RenderBox(w, h, " DISKS ", b.String(), theme.DSK)
}

func (m MonitorModel) renderR2Box(w, h int, theme Theme) string {
	reaction := R2Reactions[m.CurrentFace]
	artColor := theme.CharAccent
	if m.CurrentFace == "alarm" { artColor = lipgloss.Color("#FF1744") }
	artSt := lipgloss.NewStyle().Foreground(artColor)
	var b strings.Builder
	for i, line := range reaction.Art {
		if m.IsBlinking { line = strings.ReplaceAll(line, "(O)", "(·)") }
		dialogue := ""
		if i == 2 { dialogue = lipgloss.NewStyle().Foreground(lipgloss.Color("#C9D1D9")).Italic(true).Render(" R2 › " + m.DisplayMsg) }
		b.WriteString(artSt.Render(fmt.Sprintf("%-20s", line)) + dialogue + "\n")
	}
	return RenderBox(w, h, " ASTROMECH ", b.String(), theme.CharAccent)
}

func (m MonitorModel) renderNetBox(w, h int, theme Theme) string {
	var b strings.Builder
	green, cyan, purple := lipgloss.Color("#69FF47"), lipgloss.Color("#00E5FF"), lipgloss.Color("#D100D1")
	b.WriteString(fmt.Sprintf(" IP: %s | PING: %dms\n\n", m.Stats.LocalIP, m.Stats.NetPing))
	
	b.WriteString(lipgloss.NewStyle().Foreground(green).Bold(true).Render(fmt.Sprintf(" NET UP   %8.1f KB/s ", m.Stats.NetSent)))
	b.WriteString(BrailleSparkline(m.NetSentHistory, w-32, green) + "\n")
	b.WriteString(lipgloss.NewStyle().Foreground(cyan).Bold(true).Render(fmt.Sprintf(" NET DOWN %8.1f KB/s ", m.Stats.NetRecv)))
	b.WriteString(BrailleSparkline(m.NetRecvHistory, w-32, cyan) + "\n")

	b.WriteString(strings.Repeat("─", w-4) + "\n")
	b.WriteString(lipgloss.NewStyle().Foreground(purple).Bold(true).Render(fmt.Sprintf(" DISK RD  %8.1f KB/s\n", m.Stats.DiskRead)))
	b.WriteString(lipgloss.NewStyle().Foreground(purple).Bold(true).Render(fmt.Sprintf(" DISK WR  %8.1f KB/s\n", m.Stats.DiskWrite)))
	b.WriteString(lipgloss.NewStyle().Faint(true).Render(fmt.Sprintf(" Total Sent: %.1fG | Total Recv: %.1fG", m.Stats.TotalNetSent, m.Stats.TotalNetRecv)))

	return RenderBox(w, h, " NET & IO ", b.String(), green)
}
