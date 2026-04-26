package r2d2

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TickMsg is sent periodically to trigger a UI refresh.
type TickMsg time.Time

// StatsMsg contains the latest system telemetry.
type StatsMsg SysStats

// ScanResultMsg contains the output of a deep process scan.
type ScanResultMsg string

// Tick creates a command that sends a TickMsg after a short delay.
func Tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// GetStatsCmd creates a command that fetches system stats using the provided manager.
// The manager is created once in main and passed through the UI model — no global state.
func GetStatsCmd(sm *StatsManager) tea.Cmd {
	return func() tea.Msg {
		return StatsMsg(sm.GetStats())
	}
}

// ScanProcessCmd creates a command that performs a deep scan on a process.
func ScanProcessCmd(pid string) tea.Cmd {
	return func() tea.Msg {
		return ScanResultMsg(GetFullProcessInfo(pid))
	}
}
