package ui

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/victx/r2d2-monitor/r2d2"
)

type statsMsg r2d2.SysStats
type blinkMsg struct{}
type moodMsg string
type typeMsg struct{}
type scanMsg string
type killResultMsg struct {
	pid  string
	name string
	err  error
}

// MonitorModel represents the state and logic for the system telemetry dashboard.
type MonitorModel struct {
	Manager        *r2d2.StatsManager
	Config         r2d2.Config
	Stats          r2d2.SysStats
	TargetMsg      string
	DisplayMsg     string
	IsBlinking     bool
	Cursor         int
	Width          int
	Height         int
	CurrentFace    string
	Ready          bool
	Searching       bool
	Search          string
	Inspecting      bool
	SelectedProcess r2d2.ProcessInfo
	Details         string
	MsgLockedUntil  time.Time
}

// InitialMonitor creates a new instance of the monitor with default settings.
func InitialMonitor(sm *r2d2.StatsManager, cfg r2d2.Config) MonitorModel {
	return MonitorModel{
		Manager:     sm,
		Config:      cfg,
		TargetMsg:   "*Bleep boop* (R2-D2 online.)",
		CurrentFace: "idle",
	}
}

// Init initializes the monitor's background processes.
func (m MonitorModel) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		m.fetchStats(),
		m.scheduleBlink(),
		m.scheduleMood(),
		m.scheduleType(),
	)
}

func (m MonitorModel) fetchStats() tea.Cmd {
	return func() tea.Msg { return statsMsg(m.Manager.GetStats()) }
}

func (m MonitorModel) scheduleBlink() tea.Cmd {
	return tea.Tick(time.Duration(1500+rand.Intn(3000))*time.Millisecond, func(t time.Time) tea.Msg {
		return blinkMsg{}
	})
}

func (m MonitorModel) scheduleMood() tea.Cmd {
	return tea.Tick(time.Duration(4+rand.Intn(5))*time.Second, func(t time.Time) tea.Msg {
		return moodMsg("")
	})
}

func (m MonitorModel) scheduleType() tea.Cmd {
	return tea.Tick(40*time.Millisecond, func(t time.Time) tea.Msg {
		return typeMsg{}
	})
}

// Update handles incoming messages and state transitions for the monitor.
func (m MonitorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.Searching {
		if km, ok := msg.(tea.KeyMsg); ok {
			switch km.String() {
			case "enter", "esc":
				m.Searching = false
				if egg, ok := R2EasterEggs[strings.ToLower(m.Search)]; ok {
					m.TargetMsg = egg
					m.DisplayMsg = ""
				}
			case "backspace":
				if len(m.Search) > 0 {
					m.Search = m.Search[:len(m.Search)-1]
				}
			default:
				if len(km.String()) == 1 {
					m.Search += km.String()
				}
			}
			return m, nil
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width, m.Height = msg.Width, msg.Height

	case typeMsg:
		if len(m.DisplayMsg) < len(m.TargetMsg) {
			m.DisplayMsg = m.TargetMsg[:len(m.DisplayMsg)+1]
		}
		return m, m.scheduleType()

	case moodMsg:
		if time.Now().Before(m.MsgLockedUntil) {
			return m, m.scheduleMood()
		}
		moods := []string{"idle", "thinking", "scanning"}
		m.CurrentFace = moods[rand.Intn(len(moods))]
		if m.Ready {
			pool := R2Reactions[m.CurrentFace].Dialogue
			if len(pool) > 0 {
				m.TargetMsg = pool[rand.Intn(len(pool))]
				m.DisplayMsg = ""
			}
		}
		return m, m.scheduleMood()

	case blinkMsg:
		m.IsBlinking = true
		return m, tea.Batch(
			tea.Tick(120*time.Millisecond, func(t time.Time) tea.Msg { return nil }),
			m.scheduleBlink(),
		)

	case statsMsg:
		s := r2d2.SysStats(msg)
		if m.Config.SortBy == "CPU" {
			sort.Slice(s.Processes, func(i, j int) bool {
				return s.Processes[i].CPU > s.Processes[j].CPU
			})
		} else {
			sort.Slice(s.Processes, func(i, j int) bool {
				var a, b float64
				fmt.Sscanf(strings.TrimSuffix(s.Processes[i].MEM, "MB"), "%f", &a)
				fmt.Sscanf(strings.TrimSuffix(s.Processes[j].MEM, "MB"), "%f", &b)
				return a > b
			})
		}
		if s.CPU > 85 && m.CurrentFace != "alarm" {
			m.CurrentFace = "alarm"
			m.TargetMsg = "*SCREEEE!* CPU at critical threshold!"
			m.DisplayMsg = ""
		}
		m.Stats = s
		m.Ready = true
		all := m.visibleEntries()
		if m.Cursor >= len(all) && len(all) > 0 {
			m.Cursor = len(all) - 1
		}
		return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg { return m.fetchStats()() })

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Sequence(tea.ExitAltScreen, tea.Quit)
		case "f1":
			m.Config.SortBy = "CPU"
			r2d2.SaveConfig(m.Config)
			m.TargetMsg = "*Whistle* Sorted by CPU."
			m.DisplayMsg = ""
		case "f2":
			m.Config.SortBy = "MEM"
			r2d2.SaveConfig(m.Config)
			m.TargetMsg = "*Boop* Sorted by RAM."
			m.DisplayMsg = ""
		case "f3":
			m.Config.ThemeIdx = (m.Config.ThemeIdx + 1) % len(Themes)
			r2d2.SaveConfig(m.Config)
			m.TargetMsg = "*Bleep!* Theme: " + Themes[m.Config.ThemeIdx].Name
			m.DisplayMsg = ""
		case "/":
			m.Searching = true
			m.Search = ""
			m.CurrentFace = "scanning"
		case "up":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down":
			visible := m.visibleEntries()
			if m.Cursor < len(visible)-1 {
				m.Cursor++
			}
		case "enter":
			if m.Inspecting {
				m.Inspecting = false
			} else {
				visible := m.visibleEntries()
				if len(visible) > 0 && m.Cursor < len(visible) {
					m.Inspecting = true
					m.SelectedProcess = visible[m.Cursor]
					m.Details = "SCANNING SYSTEM..."
					m.CurrentFace = "scanning"
					m.TargetMsg = "*Bleep* Deep scan: PID " + m.SelectedProcess.ID
					m.DisplayMsg = ""
					r2d2.LogAction("PROCESS_SCAN", "Initiated scan for PID "+m.SelectedProcess.ID)
					return m, func() tea.Msg {
						return scanMsg(r2d2.GetFullProcessInfo(m.SelectedProcess.ID))
					}
				}
			}
		case "esc":
			if m.Inspecting {
				m.Inspecting = false
				m.CurrentFace = "idle"
			}
		case "f9":
			visible := m.visibleEntries()
			if len(visible) > 0 && m.Cursor < len(visible) {
				p := visible[m.Cursor]
				pid := p.ID
				name := p.Name
				m.TargetMsg = "*CHARGING...* Targeting PID " + pid + " — " + name
				m.CurrentFace = "alarm"
				m.DisplayMsg = ""
				return m, tea.Cmd(func() tea.Msg {
					_, err := r2d2.KillProcess(pid)
					if err != nil {
						return killResultMsg{pid: pid, name: name, err: err}
					}
					return killResultMsg{pid: pid, name: name, err: nil}
				})
			}
		}

	case scanMsg:
		m.Details = string(msg)
		m.CurrentFace = "thinking"
		return m, nil

	case killResultMsg:
		if msg.err != nil {
			r2d2.LogError(msg.err, fmt.Sprintf("Failed to kill %s (PID %s)", msg.name, msg.pid))
			m.TargetMsg = "*ERROR* " + msg.name + " › " + msg.err.Error()
			m.CurrentFace = "thinking"
			m.MsgLockedUntil = time.Now().Add(8 * time.Second)
		} else {
			r2d2.LogAction("PROCESS_KILL", fmt.Sprintf("PID %s (%s) terminated", msg.pid, msg.name))
			m.TargetMsg = "*ZAP!* PID " + msg.pid + " (" + msg.name + ") eliminated."
			m.CurrentFace = "success"
			m.MsgLockedUntil = time.Now().Add(3 * time.Second)
		}
		m.DisplayMsg = ""
		return m, tea.Batch(
			m.fetchStats(),
			tea.Tick(9*time.Second, func(t time.Time) tea.Msg { return moodMsg("") }),
		)
	}
	return m, nil
}

func (m *MonitorModel) visibleEntries() []r2d2.ProcessInfo {
	all := m.Stats.Processes
	if m.Search == "" {
		return all
	}
	q := strings.ToLower(m.Search)
	var filtered []r2d2.ProcessInfo
	for _, p := range all {
		if strings.Contains(strings.ToLower(p.Name), q) {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

// View renders the terminal user interface for the monitor.
func (m MonitorModel) View() string {
	if m.Width == 0 || m.Height == 0 {
		return "Initializing..."
	}

	theme := Themes[m.Config.ThemeIdx]
	W, H := m.Width, m.Height

	accentSt := lipgloss.NewStyle().Foreground(theme.CPU).Bold(true)
	dimSt := lipgloss.NewStyle().Foreground(theme.CharMain)
	whiteSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#E8E8E8"))
	selectedSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(theme.CPU).Bold(true)

	if !m.Ready {
		art := strings.Join(R2Reactions["idle"].Art, "\n")
		return lipgloss.Place(W, H, lipgloss.Center, lipgloss.Center,
			lipgloss.NewStyle().Foreground(theme.CPU).Render(art))
	}

	// 1. Build Top Section dynamically
	header := m.renderHeader(W, theme)
	dialogue := m.renderDialogue(W)
	sep := accentSt.Render(strings.Repeat("─", W))

	fixedCols := 26
	nameW := W - fixedCols - 2
	if nameW < 14 {
		nameW = 14
	}

	sortSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(theme.CPU).Bold(true)
	cCPU, cMEM := "CPU%    ", "MEM      "
	if m.Config.SortBy == "CPU" {
		cCPU = "CPU%↓   "
	} else {
		cMEM = "MEM↓     "
	}
	
	hCPU := dimSt.Render(cCPU)
	if m.Config.SortBy == "CPU" {
		hCPU = sortSt.Render(cCPU)
	}
	hMEM := dimSt.Render(cMEM)
	if m.Config.SortBy == "MEM" {
		hMEM = sortSt.Render(cMEM)
	}

	tableHeader := dimSt.Render(fmt.Sprintf(" %-7s %-*s ", "PID", nameW, "NAME")) + hCPU + hMEM

	// Create the static top block to measure it
	topBlock := lipgloss.JoinVertical(lipgloss.Left,
		header,
		dialogue,
		sep,
		tableHeader,
	)
	topH := lipgloss.Height(topBlock)

	// 2. Calculate remaining space for the list (leaving 1 line for footer)
	listH := H - topH - 1
	if listH < 1 {
		listH = 1
	}

	filtered := m.visibleEntries()
	start := m.Cursor - listH/2
	if start < 0 {
		start = 0
	}
	if start+listH > len(filtered) {
		start = len(filtered) - listH
	}
	if start < 0 {
		start = 0
	}

	// 3. Render the list to exactly listH
	var rows strings.Builder
	for i := start; i < start+listH; i++ {
		if i >= 0 && i < len(filtered) {
			p := filtered[i]
			cur, rowSt := " ", dimSt
			if m.Cursor == i {
				cur, rowSt = "›", selectedSt
			}

			cpuV, memV := fmt.Sprintf("%-8.1f", p.CPU), fmt.Sprintf("%-10s", p.MEM)
			if m.Cursor != i {
				if m.Config.SortBy == "CPU" {
					cpuV = whiteSt.Render(cpuV)
				} else {
					memV = whiteSt.Render(memV)
				}
			}

			n := truncate(p.Name, nameW)
			line := fmt.Sprintf("%s %-7s %-*s %s%s", cur, p.ID, nameW, n, cpuV, memV)
			rows.WriteString(rowSt.Render(line) + "\n")
		} else {
			rows.WriteString(strings.Repeat(" ", W) + "\n")
		}
	}

	// 4. Assemble final view with anchored footer
	info := fmt.Sprintf(" %d processes ", len(filtered))
	keys := "[↑↓] NAV  [ENTER] SCAN  [F1] CPU  [F2] MEM  [F3] THEME  [/] SEARCH  [F9] KILL  [Q] QUIT"
	footer := lipgloss.NewStyle().
		Background(lipgloss.Color("#161B22")).Foreground(theme.CPU).Bold(true).Width(W).
		Render(info + "│ " + keys)

	finalView := lipgloss.JoinVertical(lipgloss.Left,
		topBlock,
		strings.TrimSuffix(rows.String(), "\n"),
		footer,
	)

	if m.Inspecting {
		return m.renderInspectionModal(W, H, theme)
	}

	return finalView
}

func truncate(s string, l int) string {
	r := []rune(s)
	if len(r) <= l {
		return s
	}
	if l > 2 {
		return string(r[:l-2]) + ".."
	}
	return string(r[:l])
}
