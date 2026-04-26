package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// barColor returns a dynamic color based on usage level: green → amber → red.
func barColor(pct float64, theme Theme) lipgloss.Color {
	if pct >= 90 { return lipgloss.Color("#FF1744") }
	if pct >= 70 { return lipgloss.Color("#FFAB40") }
	return theme.CPU
}

// ProgressBar renders a btop-style bar with dynamic fill color and subtle empty track.
func ProgressBar(percent float64, width int, color lipgloss.Color) string {
	if width <= 0 { return "" }
	f := int((percent / 100.0) * float64(width))
	if f > width { f = width }
	if f < 0 { f = 0 }
	e := width - f
	filled := lipgloss.NewStyle().Foreground(color).Render(strings.Repeat("▓", f))
	empty  := lipgloss.NewStyle().Foreground(lipgloss.Color("#2A2A2A")).Render(strings.Repeat("░", e))
	return filled + empty
}

// BrailleSparkline renders a high-resolution graph using Braille patterns.
func BrailleSparkline(data []float64, width int, color lipgloss.Color) string {
	if len(data) == 0 || width <= 0 { return "" }
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

// RenderBox creates a btop-style rounded box with a colored title badge.
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
		} else {
			clamped.WriteString("\n")
		}
	}

	rendered := style.Render(strings.TrimSuffix(clamped.String(), "\n"))
	boxLines := strings.Split(rendered, "\n")
	if len(boxLines) > 0 && title != "" {
		tStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(color).Bold(true).Padding(0, 1)
		tText  := tStyle.Render(title)
		pad    := (width - 2) - lipgloss.Width(tText) - 2
		if pad < 0 { pad = 0 }
		boxLines[0] = "╭─" + tText + strings.Repeat("─", pad) + "╮"
	}
	return strings.Join(boxLines, "\n")
}

// ── CPU ──────────────────────────────────────────────────────────────────────

func (m MonitorModel) renderCPUBox(w, h int, theme Theme) string {
	var b strings.Builder

	// Header row: model + battery flush right
	modelSt := lipgloss.NewStyle().Foreground(theme.CPU).Bold(true)
	model   := truncate(m.Stats.CPUModel, w-6)
	battery := m.renderBattery(theme)
	modelLine := modelSt.Render(model)
	if battery != "" {
		gap := w - 4 - lipgloss.Width(model) - lipgloss.Width(battery)
		if gap < 1 { gap = 1 }
		modelLine += strings.Repeat(" ", gap) + battery
	}
	b.WriteString(modelLine + "\n")

	// OS line
	osSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#8B949E"))
	b.WriteString(osSt.Render(" " + truncate(m.Stats.OSName, w-6)) + "\n")

	// Total CPU bar
	totalColor := barColor(m.Stats.CPU, theme)
	totalBar   := ProgressBar(m.Stats.CPU, w-20, totalColor)
	totalSt    := lipgloss.NewStyle().Foreground(totalColor).Bold(true)
	b.WriteString(fmt.Sprintf(" ALL %s %s\n",
		totalBar,
		totalSt.Render(fmt.Sprintf("%5.1f%%", m.Stats.CPU)),
	))
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#30363D")).Render(strings.Repeat("─", w-4)) + "\n")

	// Per-core grid
	cols := 4
	if w < 100 { cols = 2 }
	coreBarW := (w-4)/cols - 14
	if coreBarW < 4 { coreBarW = 4 }
	for i := 0; i < len(m.Stats.CPUCores); i += cols {
		for j := 0; j < cols && i+j < len(m.Stats.CPUCores); j++ {
			idx := i + j
			val := m.Stats.CPUCores[idx]
			col := barColor(val, theme)
			bar := ProgressBar(val, coreBarW, col)
			pct := lipgloss.NewStyle().Foreground(col).Render(fmt.Sprintf("%3.0f%%", val))
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#8B949E")).Render(fmt.Sprintf("C%-2d", idx)) +
				":" + bar + pct + "  ")
		}
		b.WriteString("\n")
	}
	return RenderBox(w, h, " CPU ", b.String(), theme.CPU)
}

func (m MonitorModel) renderBattery(theme Theme) string {
	if m.Stats.Battery.Percent == 0 { return "" }
	color := lipgloss.Color("#69FF47")
	if m.Stats.Battery.Percent < 40 { color = lipgloss.Color("#FFAB40") }
	if m.Stats.Battery.Percent < 20 { color = lipgloss.Color("#FF1744") }
	pBar   := ProgressBar(m.Stats.Battery.Percent, 8, color)
	status := "BAT"
	if m.Stats.Battery.Status == "Charging" || m.Stats.Battery.Status == "AC Power" { status = "PWR" }
	return lipgloss.NewStyle().Foreground(color).Bold(true).Render(
		fmt.Sprintf("%s %s %d%%", status, pBar, int(m.Stats.Battery.Percent)),
	)
}

// ── MEMORY ───────────────────────────────────────────────────────────────────

func (m MonitorModel) renderMemBox(w, h int, theme Theme) string {
	var b strings.Builder
	barW := w - 20
	if barW < 4 { barW = 4 }

	ramColor  := barColor(m.Stats.RAM, theme)
	swapColor := barColor(m.Stats.Swap, theme)
	if swapColor == theme.CPU { swapColor = theme.RAM }

	// RAM row
	pctSt := lipgloss.NewStyle().Foreground(ramColor).Bold(true)
	b.WriteString(fmt.Sprintf(" RAM  %s %s\n",
		ProgressBar(m.Stats.RAM, barW, ramColor),
		pctSt.Render(fmt.Sprintf("%5.1f%%", m.Stats.RAM)),
	))
	detailSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#8B949E"))
	b.WriteString(detailSt.Render(fmt.Sprintf("      %.1fG / %.1fG  avail %.1fG  cache %.1fG\n",
		m.Stats.RAMUsed, m.Stats.RAMTotal, m.Stats.RAMAvailable, m.Stats.RAMCached)))

	// divider
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#30363D")).Render(strings.Repeat("─", w-4)) + "\n")

	// SWAP row
	swapPctSt := lipgloss.NewStyle().Foreground(swapColor).Bold(true)
	b.WriteString(fmt.Sprintf(" SWAP %s %s\n",
		ProgressBar(m.Stats.Swap, barW, swapColor),
		swapPctSt.Render(fmt.Sprintf("%5.1f%%", m.Stats.Swap)),
	))
	b.WriteString(detailSt.Render(fmt.Sprintf("      %.1fG / %.1fG\n", m.Stats.SwapUsed, m.Stats.SwapTotal)))

	return RenderBox(w, h, " MEMORY ", b.String(), theme.RAM)
}

// ── DISKS ────────────────────────────────────────────────────────────────────

func (m MonitorModel) renderDiskBox(w, h int, theme Theme) string {
	var b strings.Builder
	barW := w - 20
	if barW < 4 { barW = 4 }

	dskColor := barColor(m.Stats.Disk, theme)
	if dskColor == theme.CPU { dskColor = theme.DSK }
	pctSt    := lipgloss.NewStyle().Foreground(dskColor).Bold(true)
	detailSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#8B949E"))
	ioSt     := lipgloss.NewStyle().Foreground(theme.DSK).Bold(true)

	// C: usage bar
	b.WriteString(fmt.Sprintf(" C:   %s %s\n",
		ProgressBar(m.Stats.Disk, barW, dskColor),
		pctSt.Render(fmt.Sprintf("%5.1f%%", m.Stats.Disk)),
	))
	b.WriteString(detailSt.Render(fmt.Sprintf("      %.1fG used / %.1fG total\n", m.Stats.DiskUsed, m.Stats.DiskTotal)))

	// Disk IO inline
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#30363D")).Render(strings.Repeat("─", w-4)) + "\n")
	rdLabel := ioSt.Render(" RD ")
	wrLabel := ioSt.Render(" WR ")
	b.WriteString(fmt.Sprintf("%s%s   %s%s\n",
		rdLabel, detailSt.Render(fmt.Sprintf("%7.1f KB/s", m.Stats.DiskRead)),
		wrLabel, detailSt.Render(fmt.Sprintf("%7.1f KB/s", m.Stats.DiskWrite)),
	))
	return RenderBox(w, h, " DISKS ", b.String(), theme.DSK)
}

// ── ASTROMECH ────────────────────────────────────────────────────────────────

func (m MonitorModel) renderR2Box(w, h int, theme Theme) string {
	reaction := R2Reactions[m.CurrentFace]
	artColor := theme.CharAccent
	if m.CurrentFace == "alarm" { artColor = lipgloss.Color("#FF1744") }
	artSt     := lipgloss.NewStyle().Foreground(artColor)
	dialSt    := lipgloss.NewStyle().Foreground(lipgloss.Color("#C9D1D9")).Italic(true)
	prefixSt  := lipgloss.NewStyle().Foreground(artColor).Bold(true)

	// Inner usable width for the art (box padding = 2 sides × 1)
	artW := w - 4

	var b strings.Builder
	for _, line := range reaction.Art {
		if m.IsBlinking { line = strings.ReplaceAll(line, "(O)", "(·)") }
		// Pad or truncate art to exactly artW so it never clips
		aw := lipgloss.Width(line)
		if aw < artW {
			line = line + strings.Repeat(" ", artW-aw)
		} else if aw > artW {
			line = truncate(line, artW)
		}
		b.WriteString(artSt.Render(line) + "\n")
	}

	// Dialogue on its own line below the art, full width
	msg := prefixSt.Render("R2›") + " " + dialSt.Render(truncate(m.DisplayMsg, artW-4))
	b.WriteString(msg + "\n")

	return RenderBox(w, h, " ASTROMECH ", b.String(), theme.CharAccent)
}

// ── NET & IO ─────────────────────────────────────────────────────────────────

func (m MonitorModel) renderNetBox(w, h int, theme Theme) string {
	var b strings.Builder
	green  := lipgloss.Color("#69FF47")
	cyan   := lipgloss.Color("#00E5FF")
	gray   := lipgloss.Color("#8B949E")

	// Header: IP + ping
	pingSt := lipgloss.NewStyle().Foreground(cyan).Bold(true)
	pingColor := green
	if m.Stats.NetPing > 100 { pingColor = lipgloss.Color("#FFAB40") }
	if m.Stats.NetPing > 300 { pingColor = lipgloss.Color("#FF1744") }
	pingVal := lipgloss.NewStyle().Foreground(pingColor).Bold(true).Render(fmt.Sprintf("%dms", m.Stats.NetPing))
	b.WriteString(lipgloss.NewStyle().Foreground(gray).Render(" IP: ") +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#C9D1D9")).Render(m.Stats.LocalIP) +
		"  " + pingSt.Render("PING ") + pingVal + "\n")

	// Sparkline rows
	sparkW := w - 30
	if sparkW < 4 { sparkW = 4 }
	upSt   := lipgloss.NewStyle().Foreground(green).Bold(true)
	dnSt   := lipgloss.NewStyle().Foreground(cyan).Bold(true)
	b.WriteString(upSt.Render(fmt.Sprintf(" ↑ %7.1f KB/s ", m.Stats.NetSent)) +
		BrailleSparkline(m.NetSentHistory, sparkW, green) + "\n")
	b.WriteString(dnSt.Render(fmt.Sprintf(" ↓ %7.1f KB/s ", m.Stats.NetRecv)) +
		BrailleSparkline(m.NetRecvHistory, sparkW, cyan) + "\n")

	// Totals footer
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#30363D")).Render(strings.Repeat("─", w-4)) + "\n")
	b.WriteString(lipgloss.NewStyle().Foreground(gray).Render(
		fmt.Sprintf(" Total ↑ %.2fG  ↓ %.2fG\n", m.Stats.TotalNetSent, m.Stats.TotalNetRecv)))

	return RenderBox(w, h, " NET ", b.String(), green)
}
