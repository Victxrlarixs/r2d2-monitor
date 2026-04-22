package r2d2

import (
	"bytes"
	"fmt"
	"os/exec"
	"time"

	"github.com/shirou/gopsutil/v3/process"
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

// GetFullProcessInfo retrieves extended metadata for a specific PID.
func GetFullProcessInfo(pidStr string) string {
	var pid int32
	fmt.Sscanf(pidStr, "%d", &pid)
	p, err := process.NewProcess(pid)
	if err != nil { return "Process not found" }

	name, _ := p.Name()
	status, _ := p.Status()
	createTime, _ := p.CreateTime()
	parent, _ := p.Parent()
	parentPID := int32(0)
	if parent != nil { parentPID = parent.Pid }
	
	cmdLine, _ := p.Cmdline()
	io, _ := p.IOCounters()
	mem, _ := p.MemoryInfo()

	elapsed := time.Since(time.Unix(createTime/1000, 0)).Truncate(time.Second)

	res := fmt.Sprintf("STATUS: %s\n", status)
	res += fmt.Sprintf("ELAPSED: %s\n", elapsed)
	if io != nil {
		res += fmt.Sprintf("IO/R: %s\n", formatBytes(io.ReadBytes))
		res += fmt.Sprintf("IO/W: %s\n", formatBytes(io.WriteBytes))
	} else {
		res += "IO/R: 0 B\nIO/W: 0 B\n"
	}
	res += fmt.Sprintf("PARENT: %d\n", parentPID)
	if mem != nil {
		res += fmt.Sprintf("MEM_VAL: %s\n", formatBytes(mem.RSS))
	}
	res += fmt.Sprintf("CMD: %s\n", cmdLine)
	res += fmt.Sprintf("NAME: %s\n", name)

	return res
}

func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit { return fmt.Sprintf("%d B", b) }
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
