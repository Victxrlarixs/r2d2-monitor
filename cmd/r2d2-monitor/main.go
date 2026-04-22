// Package main provides the entry point for R2-D2 Monitoring Console.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/victx/r2d2-monitor/r2d2"
	"github.com/victx/r2d2-monitor/r2d2/ui"
)

var (
	isChild bool
)

var rootCmd = &cobra.Command{
	Use:   "r2d2-monitor",
	Short: "R2-D2 Monitor: High-Performance Monitoring Console for Windows",
	Run: func(cmd *cobra.Command, args []string) {
		// Detect if we are running in a terminal. If not, relaunch.
		if !isChild && os.Getenv("R2D2_MONITOR_RUNNING") == "" {
			relaunchInTerminal()
			return
		}

		// Initialize professional components
		r2d2.InitLogger()
		cfg := r2d2.LoadConfig()
		sm := r2d2.NewStatsManager()

		p := tea.NewProgram(ui.InitialMonitor(sm, cfg), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error starting R2-D2 Monitor: %v\n", err)
			time.Sleep(10 * time.Second)
			os.Exit(1)
		}
	},
}

// relaunchInTerminal forces the application to open in a new CMD window.
func relaunchInTerminal() {
	exe, err := os.Executable()
	if err != nil {
		return
	}

	// Launch via 'start' in a new CMD window.
	script := fmt.Sprintf("cmd /C start \"R2-D2 Monitor\" \"%s\" --child", exe)
	cmd := exec.Command("powershell", "-NoProfile", "-WindowStyle", "Hidden", "-Command", script)
	cmd.Env = append(os.Environ(), "R2D2_MONITOR_RUNNING=1")
	
	if err := cmd.Run(); err != nil {
		fmt.Printf("Launch failed: %v\n", err)
		time.Sleep(5 * time.Second)
	}
}

func init() {
	// Disable Cobra's Mousetrap to allow double-clicking in Windows Explorer.
	cobra.MousetrapHelpText = ""

	rootCmd.Flags().BoolVar(&isChild, "child", false, "Internal flag")
	_ = rootCmd.Flags().MarkHidden("child")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
