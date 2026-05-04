package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/victx/r2d2-monitor/r2d2"
)

type FocusMode int
const (
	FocusProcs FocusMode = iota
	FocusDisk
	FocusNet
)

type MonitorModel struct {
	Stats           r2d2.SysStats
	Config          r2d2.Config
	SM              *r2d2.StatsManager // single source of truth, created in main
	Width           int
	Height          int
	Ready           bool
	Cursor          int
	SearchMode      bool
	SearchQuery     string
	Sorting         string // "cpu", "mem"
	Inspecting      bool
	ConfirmKill     bool
	CurrentFace     string
	IsBlinking      bool
	DisplayMsg      string
	MsgLockedUntil  time.Time
	NetRecvHistory  []float64
	NetSentHistory  []float64
	Details         string
	SelectedProcess r2d2.ProcessInfo
	
	// Layout preset system
	PresetController *PresetController
	Focus            FocusMode
}

func InitialMonitor(sm *r2d2.StatsManager, cfg r2d2.Config) MonitorModel {
	m := MonitorModel{
		SM:          sm,
		Config:      cfg,
		CurrentFace: "idle",
		Sorting:     "cpu",
		PresetController: NewPresetController(cfg.LayoutPreset),
	}
	m.setReaction("idle", 0)
	return m
}

func (m MonitorModel) Init() tea.Cmd {
	return tea.Batch(r2d2.Tick(), m.blinkCmd(), m.idleMsgCmd())
}

func (m *MonitorModel) setReaction(face string, lockDuration time.Duration) {
	m.CurrentFace = face
	reaction := R2Reactions[face]
	if len(reaction.Dialogue) > 0 {
		m.DisplayMsg = reaction.Dialogue[r2d2.RandomInt(len(reaction.Dialogue))]
	}
	if lockDuration > 0 {
		m.MsgLockedUntil = time.Now().Add(lockDuration)
	}
}

func (m MonitorModel) blinkCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*time.Duration(1500+r2d2.RandomInt(2000)), func(t time.Time) tea.Msg {
		return blinkMsg{}
	})
}

func (m MonitorModel) idleMsgCmd() tea.Cmd {
	return tea.Tick(time.Second*4, func(t time.Time) tea.Msg {
		return rotateIdleMsg{}
	})
}

type blinkMsg struct{}
type rotateIdleMsg struct{}
type endBlinkMsg struct{}

func (m MonitorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Auto-reset from special faces to idle with fresh random message
	if time.Now().After(m.MsgLockedUntil) && m.CurrentFace != "idle" && !m.SearchMode && !m.Inspecting {
		m.setReaction("idle", 0)
	}

	// Always ensure there is a message if idle
	if m.DisplayMsg == "" {
		m.setReaction("idle", 0)
	}
	if m.Stats.CPU > 90 || m.Stats.RAM > 90 {
		if m.CurrentFace != "alarm" { m.setReaction("alarm", time.Second*5) }
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width, m.Height = msg.Width, msg.Height
	case blinkMsg:
		m.IsBlinking = true
		cmds = append(cmds, tea.Tick(time.Millisecond*150, func(t time.Time) tea.Msg { return endBlinkMsg{} }))
	case endBlinkMsg:
		m.IsBlinking = false
		cmds = append(cmds, m.blinkCmd())
	case rotateIdleMsg:
		if m.CurrentFace == "idle" && !m.SearchMode && !m.Inspecting { m.setReaction("idle", 0) }
		cmds = append(cmds, m.idleMsgCmd())
	case r2d2.TickMsg:
		cmds = append(cmds, r2d2.GetStatsCmd(m.SM, m.getVisiblePIDs(), m.Config), r2d2.Tick())
		// SPONTANEITY: Randomly trigger thinking/scanning animations to feel "alive"
		if !m.Inspecting && !m.ConfirmKill && m.CurrentFace == "idle" && r2d2.RandomInt(100) < 15 {
			m.setReaction("thinking", time.Second*2)
		}
		if m.Inspecting {
			cmds = append(cmds, r2d2.ScanProcessCmd(m.SelectedProcess.ID))
		}
	case r2d2.StatsMsg:
		s := r2d2.SysStats(msg)
		m.Stats, m.Ready = s, true
		m.NetRecvHistory = append(m.NetRecvHistory, s.NetRecv)
		if len(m.NetRecvHistory) > 50 { m.NetRecvHistory = m.NetRecvHistory[1:] }
		m.NetSentHistory = append(m.NetSentHistory, s.NetSent)
		if len(m.NetSentHistory) > 50 { m.NetSentHistory = m.NetSentHistory[1:] }
	case r2d2.ScanResultMsg:
		m.Details = string(msg)
	case tea.KeyMsg:
		if m.ConfirmKill {
			switch msg.String() {
			case "y", "Y":
				entries := m.visibleEntries()
				if m.Cursor < len(entries) {
					m.setReaction("alarm", time.Second*4)
					r2d2.KillProcess(entries[m.Cursor].ID)
				}
				m.ConfirmKill = false
			default:
				m.ConfirmKill = false
				m.setReaction("idle", 0)
			}
			return m, nil
		}
		if m.SearchMode {
			if key.Matches(msg, DefaultKeyMap.Quit) { 
				m.SearchMode, m.SearchQuery = false, ""
				m.setReaction("idle", 0)
				return m, nil
			}
			if msg.Type == tea.KeyEnter { 
				m.SearchMode = false
				m.setReaction("success", time.Second*2)
				return m, nil
			}
			if msg.Type == tea.KeyBackspace && len(m.SearchQuery) > 0 { m.SearchQuery = m.SearchQuery[:len(m.SearchQuery)-1] }
			if msg.Type == tea.KeyRunes { m.SearchQuery += msg.String() }
			// Check easter eggs on every keystroke
			if egg, ok := R2EasterEggs[strings.ToLower(m.SearchQuery)]; ok {
				m.DisplayMsg = egg
				m.CurrentFace = "scanning"
				m.MsgLockedUntil = time.Now().Add(time.Second * 4)
			}
			return m, nil
		}
		switch {
		case key.Matches(msg, DefaultKeyMap.Quit):
			if m.Inspecting { m.Inspecting = false ; m.setReaction("idle", 0) ; return m, nil }
			return m, tea.Quit
		case key.Matches(msg, DefaultKeyMap.Up): 
			if m.Focus == FocusProcs {
				if m.Cursor > 0 { m.Cursor-- }
			} else if m.Focus == FocusDisk {
				m.cycleDisk(-1)
			} else if m.Focus == FocusNet {
				m.cycleNet(-1)
			}
		case key.Matches(msg, DefaultKeyMap.Down): 
			if m.Focus == FocusProcs {
				if m.Cursor < len(m.visibleEntries())-1 { m.Cursor++ }
			} else if m.Focus == FocusDisk {
				m.cycleDisk(1)
			} else if m.Focus == FocusNet {
				m.cycleNet(1)
			}
		case key.Matches(msg, DefaultKeyMap.Tab):
			m.Focus = (m.Focus + 1) % 3
			m.setReaction("thinking", time.Second)
		case msg.Type == tea.KeyEnter:
			m.Inspecting = !m.Inspecting
			if m.Inspecting {
				entries := m.visibleEntries()
				if m.Cursor < len(entries) {
					m.SelectedProcess = entries[m.Cursor]
					m.setReaction("scanning", 0)
					m.Details = "SCANNING..."
					cmds = append(cmds, r2d2.ScanProcessCmd(m.SelectedProcess.ID))
				}
			} else { m.setReaction("idle", 0) }
		case key.Matches(msg, DefaultKeyMap.SortCPU): m.Sorting = "cpu" ; m.setReaction("thinking", time.Second*2)
		case key.Matches(msg, DefaultKeyMap.SortMem): m.Sorting = "mem" ; m.setReaction("thinking", time.Second*2)
		case key.Matches(msg, DefaultKeyMap.Theme):
			m.Config.ThemeIdx = (m.Config.ThemeIdx + 1) % len(Themes)
			m.setReaction("success", time.Second)
			r2d2.SaveConfig(m.Config)
		case key.Matches(msg, DefaultKeyMap.Preset):
			// Handle preset cycling
			oldPreset := m.PresetController.GetCurrentPresetNumber()
			err := m.PresetController.CyclePreset()
			if err == nil {
				newPreset := m.PresetController.GetCurrentPresetNumber()
				// Update config and save
				m.Config.LayoutPreset = newPreset
				r2d2.SaveConfig(m.Config)
				
				// Trigger appropriate R2-D2 reaction
				oldConfig, _ := GetPresetConfig(oldPreset)
				newConfig, _ := GetPresetConfig(newPreset)
				
				if oldConfig != nil && newConfig != nil {
					if oldConfig.ShowR2D2 && !newConfig.ShowR2D2 {
						// R2-D2 is being hidden - farewell reaction
						m.setReaction("success", time.Second*2)
					} else if !oldConfig.ShowR2D2 && newConfig.ShowR2D2 {
						// R2-D2 is being shown - greeting reaction
						m.setReaction("success", time.Second*2)
					} else if newConfig.ShowR2D2 {
						// R2-D2 stays visible - positive reaction
						m.setReaction("success", time.Second*2)
					}
				}
			}
		case key.Matches(msg, DefaultKeyMap.Search): 
			m.SearchMode, m.SearchQuery = true, ""
			m.setReaction("thinking", 0)
		case key.Matches(msg, DefaultKeyMap.Kill):
			entries := m.visibleEntries()
			if m.Cursor < len(entries) {
				m.ConfirmKill = true
				m.setReaction("alarm", time.Second*4)
			}
		}
	}
	return m, tea.Batch(cmds...)
}

func (m MonitorModel) View() string {
	if m.Width == 0 || m.Height == 0 { return "Initializing..." }
	if m.Width < 80 || m.Height < 20 {
		return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center,
			lipgloss.NewStyle().Foreground(lipgloss.Color("#FF1744")).Bold(true).
				Render(fmt.Sprintf("TERMINAL TOO SMALL (%dx%d)\nMINIMUM REQUIRED: 80x20", m.Width, m.Height)))
	}
	
	theme := Themes[m.Config.ThemeIdx]
	W, H := m.Width, m.Height
	
	if !m.Ready {
		art := strings.Join(R2Reactions["idle"].Art, "\n")
		return lipgloss.Place(W, H, lipgloss.Center, lipgloss.Center, lipgloss.NewStyle().Foreground(theme.CPU).Render(art))
	}

	// Get layout dimensions from preset controller
	dims := m.PresetController.CalculateDimensions(W, H)
	layout := m.PresetController.GetCurrentLayout()
	
	var topRow string
	var leftCol string
	var rightCol string
	
	// Build top row based on layout configuration
	if layout.ShowR2D2 && layout.ShowCPU {
		r2Box := m.renderR2Box(dims.R2D2Box.Width, dims.R2D2Box.Height, theme)
		cpuBox := m.renderCPUBox(dims.CPUBox.Width, dims.CPUBox.Height, theme)
		topRow = lipgloss.JoinHorizontal(lipgloss.Top, r2Box, cpuBox)
	} else if layout.ShowCPU {
		cpuBox := m.renderCPUBox(dims.CPUBox.Width, dims.CPUBox.Height, theme)
		topRow = cpuBox
	}
	
	// Build left column panels
	var leftPanels []string
	if layout.ShowMemory && dims.MemoryBox.Width > 0 {
		leftPanels = append(leftPanels, m.renderMemBox(dims.MemoryBox.Width, dims.MemoryBox.Height, theme))
	}
	if layout.ShowDisk && dims.DiskBox.Width > 0 {
		leftPanels = append(leftPanels, m.renderDiskBox(dims.DiskBox.Width, dims.DiskBox.Height, theme))
	}
	if layout.ShowNetwork && dims.NetworkBox.Width > 0 {
		leftPanels = append(leftPanels, m.renderNetBox(dims.NetworkBox.Width, dims.NetworkBox.Height, theme))
	}
	
	if len(leftPanels) > 0 {
		leftCol = lipgloss.JoinVertical(lipgloss.Left, leftPanels...)
	}
	
	// Build process panel
	if layout.ShowProcess && dims.ProcessBox.Width > 0 {
		rightCol = m.renderProcBox(dims.ProcessBox.Width, dims.ProcessBox.Height, theme)
	}
	
	// Combine main content
	var mainContent string
	if leftCol != "" && rightCol != "" {
		mainContent = lipgloss.JoinHorizontal(lipgloss.Top, leftCol, rightCol)
	} else if rightCol != "" {
		mainContent = rightCol
	} else if leftCol != "" {
		mainContent = leftCol
	}
	
	// Build footer with preset indicator
	uptime := fmt.Sprintf(" UP: %s ", m.Stats.Uptime)
	presetName := fmt.Sprintf(" PRESET: %s ", m.PresetController.GetCurrentPresetName())
	keys := " [↑↓] NAV  [ENTER] SCAN  [F1/F2] SORT  [F3] THEME  [P] PRESET  [/] SEARCH  [F9] KILL  [Q] QUIT "
	footerSt := lipgloss.NewStyle().Background(theme.CPU).Foreground(lipgloss.Color("#000000")).Bold(true)
	footer := footerSt.Render(fmt.Sprintf(" %-15s %-15s %-15s %s", fmt.Sprintf("%d PROCS", len(m.visibleEntries())), uptime, presetName, keys))
	footer = lipgloss.NewStyle().Width(W).Background(lipgloss.Color("#161B22")).Render(footer)

	// Combine everything
	var view string
	if topRow != "" && mainContent != "" {
		view = lipgloss.JoinVertical(lipgloss.Left, topRow, mainContent, footer)
	} else if mainContent != "" {
		view = lipgloss.JoinVertical(lipgloss.Left, mainContent, footer)
	} else {
		view = footer
	}
	
	return lipgloss.NewStyle().MarginTop(1).Render(view)
}

func (m MonitorModel) renderProcBox(w, h int, theme Theme) string {
	headerSt := lipgloss.NewStyle().Foreground(theme.CPU).Bold(true)
	selectedSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(theme.CPU).Bold(true)
	labelSt := lipgloss.NewStyle().Foreground(theme.CPU).Bold(true)
	valSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	nameW := w - 36
	if nameW < 10 { nameW = 10 }
	listH := h - 4
	var infoPanel string

	if m.ConfirmKill {
		entries := m.visibleEntries()
		procName := "Unknown"
		if m.Cursor < len(entries) {
			procName = entries[m.Cursor].Name
		}
		confirmBox := lipgloss.NewStyle().
			Background(lipgloss.Color("#FF1744")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			Padding(0, 1).
			Width(w - 4).
			Render(fmt.Sprintf(" WARNING: Kill process %s? (y/N) ", truncate(procName, w-40)))
		infoPanel = confirmBox + "\n"
		listH -= 2
	} else if m.SearchMode {
		searchBox := lipgloss.NewStyle().
			Background(lipgloss.Color("#161B22")).
			Foreground(theme.CPU).
			Bold(true).
			Padding(0, 1).
			Width(w - 4).
			Render(" SEARCH: " + m.SearchQuery + "_")
		infoPanel = searchBox + "\n"
		listH -= 2
	} else if m.Inspecting {
		infoH := 9
		listH -= infoH
		data := make(map[string]string)
		lines := strings.Split(m.Details, "\n")
		for _, l := range lines {
			parts := strings.SplitN(l, ":", 2)
			if len(parts) == 2 { data[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1]) }
		}
		col1 := lipgloss.JoinVertical(lipgloss.Left, labelSt.Render("Status:   ") + valSt.Render(data["STATUS"]), labelSt.Render("Elapsed:  ") + valSt.Render(data["ELAPSED"]))
		col2 := lipgloss.JoinVertical(lipgloss.Left, labelSt.Render("IO/R: ") + valSt.Render(data["IO/R"]), labelSt.Render("IO/W: ") + valSt.Render(data["IO/W"]))
		col3 := lipgloss.JoinVertical(lipgloss.Left, labelSt.Render("Parent: ") + valSt.Render(data["PARENT"]), labelSt.Render("Memory: ") + valSt.Render(fmt.Sprintf("%s / %s", m.SelectedProcess.MEM, data["MEM_VAL"])))
		topRow := lipgloss.JoinHorizontal(lipgloss.Top, lipgloss.NewStyle().Width(w/3).Render(col1), lipgloss.NewStyle().Width(w/3).Render(col2), lipgloss.NewStyle().Width(w/3).Render(col3))
		cmdRow := labelSt.Render("\nCMD: ") + valSt.Render(truncate(data["CMD"], w-10))
		controls := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF1744")).Render("\n[F9] Terminate  [Esc] Hide Details")
		infoBox := lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).BorderForeground(theme.CPU).Width(w - 4).Padding(0, 1).Render(topRow + cmdRow + controls)
		infoPanel = infoBox + "\n"
	}

	cpuSort, memSort := "", ""
	if m.Sorting == "cpu" { cpuSort = " >" } else { memSort = " >" }
	header := headerSt.Render(fmt.Sprintf("  %-7s %-*s %-8s %-10s", "PID", nameW, "NAME", "CPU%"+cpuSort, "MEM"+memSort))
	
	filtered := m.visibleEntries()
	start := m.Cursor - listH/2
	if start < 0 { start = 0 }
	if start+listH > len(filtered) { start = len(filtered) - listH }
	if start < 0 { start = 0 }

	var rows strings.Builder
	for i := start; i < start+listH; i++ {
		if i >= 0 && i < len(filtered) {
			p := filtered[i]
			lineSt, prefix := lipgloss.NewStyle().Foreground(theme.CharMain), " "
			if m.Cursor == i { lineSt, prefix = selectedSt, ">" }
			row := fmt.Sprintf("%s %-7s %-*s %-8.1f %-10s", prefix, p.ID, nameW, truncate(p.Name, nameW), p.CPU, p.MEM)
			rows.WriteString(lineSt.Render(row) + "\n")
		}
	}
	content := infoPanel + header + "\n" + strings.Repeat("-", w-4) + "\n" + rows.String()
	title := " PROCESSES "
	if m.Focus == FocusProcs {
		title = " > PROCESSES "
	}
	return RenderBox(w, h, title, content, theme.CPU)
}

func (m MonitorModel) visibleEntries() []r2d2.ProcessInfo {
	if m.SearchQuery == "" { return m.Stats.Processes }
	var filtered []r2d2.ProcessInfo
	q := strings.ToLower(m.SearchQuery)
	for _, p := range m.Stats.Processes {
		if strings.Contains(strings.ToLower(p.Name), q) || strings.Contains(p.ID, q) {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

func (m MonitorModel) getVisiblePIDs() []string {
	filtered := m.visibleEntries()
	if len(filtered) == 0 { return nil }
	
	// Estimate list height (roughly) or just take a safe window around cursor
	// Since we don't know the exact listH here without re-calculating everything,
	// we'll just take the 40 processes around the cursor as "visible".
	start := m.Cursor - 20
	if start < 0 { start = 0 }
	end := start + 40
	if end > len(filtered) { end = len(filtered) }
	
	var pids []string
	for i := start; i < end; i++ {
		pids = append(pids, filtered[i].ID)
	}
	return pids
}

func (m *MonitorModel) cycleDisk(delta int) {
	if len(m.Stats.AllDisks) == 0 { return }
	idx := -1
	for i, d := range m.Stats.AllDisks {
		if d == m.Stats.SelectedDisk { idx = i; break }
	}
	if idx == -1 { idx = 0 }
	idx = (idx + delta + len(m.Stats.AllDisks)) % len(m.Stats.AllDisks)
	m.Config.SelectedDisk = m.Stats.AllDisks[idx]
	r2d2.SaveConfig(m.Config)
}

func (m *MonitorModel) cycleNet(delta int) {
	if len(m.Stats.AllNet) == 0 { return }
	idx := -1
	for i, n := range m.Stats.AllNet {
		if n == m.Stats.SelectedNet { idx = i; break }
	}
	if idx == -1 { idx = 0 }
	idx = (idx + delta + len(m.Stats.AllNet)) % len(m.Stats.AllNet)
	m.Config.SelectedNetInt = m.Stats.AllNet[idx]
	r2d2.SaveConfig(m.Config)
}

func truncate(s string, l int) string {
	if lipgloss.Width(s) <= l { return s }
	if l < 3 { return s[:l] }
	return s[:l-3] + "..."
}

type KeyMap struct {
	Up, Down, Quit, Theme, Search, Kill, SortCPU, SortMem, Preset, Tab key.Binding
}

var DefaultKeyMap = KeyMap{
	Up:      key.NewBinding(key.WithKeys("up", "k")),
	Down:    key.NewBinding(key.WithKeys("down", "j")),
	Quit:    key.NewBinding(key.WithKeys("q", "esc", "ctrl+c")),
	Theme:   key.NewBinding(key.WithKeys("f3")),
	Search:  key.NewBinding(key.WithKeys("/")),
	Kill:    key.NewBinding(key.WithKeys("f9")),
	SortCPU: key.NewBinding(key.WithKeys("f1")),
	SortMem: key.NewBinding(key.WithKeys("f2")),
	Preset:  key.NewBinding(key.WithKeys("p")),
	Tab:     key.NewBinding(key.WithKeys("tab")),
}
