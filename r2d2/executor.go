package r2d2

import (
	"bytes"
	"fmt"
	"os/exec"
)

// Commander defines the behavior for executing system instructions.
type Commander interface {
	Execute(cmdStr string) (string, error)
}

// PSExecutor is a concrete implementation of Commander using PowerShell.
type PSExecutor struct{}

// Execute runs a PowerShell command and returns the combined output or an error.
func (e *PSExecutor) Execute(cmdStr string) (string, error) {
	cmd := exec.Command("powershell", "-NoProfile", "-Command", cmdStr)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		if stderr.String() != "" {
			return stderr.String(), err
		}
		return out.String(), err
	}
	return out.String(), nil
}

// DefaultExecutor is a global instance for standard command execution.
var DefaultExecutor Commander = &PSExecutor{}

// ExecuteCommand provides a package-level entry point for command execution.
func ExecuteCommand(cmdStr string) (string, error) {
	return DefaultExecutor.Execute(cmdStr)
}

// KillProcess forcefully terminates a process and its entire child tree.
// It uses the Windows 'taskkill' utility with /F (force) and /T (tree) flags.
func KillProcess(pid string) (string, error) {
	cmd := exec.Command("taskkill", "/F", "/T", "/PID", pid)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		msg := stderr.String()
		if msg == "" {
			msg = out.String()
		}
		if msg == "" {
			msg = err.Error()
		}
		return msg, fmt.Errorf("taskkill error: %s", msg)
	}
	return out.String(), nil
}

// GetFullProcessInfo retrieves extended metadata for a specific PID using WMI/PowerShell.
func GetFullProcessInfo(pid string) string {
	script := fmt.Sprintf("Get-Process -Id %s | Select-Object Path, StartTime, Company, Description | Format-List", pid)
	out, err := ExecuteCommand(script)
	if err != nil {
		return "Scanning error: " + err.Error()
	}
	return out
}
