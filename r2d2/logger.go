package r2d2

import (
	"log"
	"os"
	"path/filepath"
)

var fileLogger *log.Logger

// InitLogger initializes a file-based logger for background auditing.
func InitLogger() {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".r2d2-monitor")
	_ = os.MkdirAll(dir, 0755)
	
	logPath := filepath.Join(dir, "r2d2.log")
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	
	fileLogger = log.New(f, "[R2-D2] ", log.LstdFlags)
	fileLogger.Println("Telemetry system initialized.")
}

// LogInfo records a standard informational message.
func LogInfo(msg string) {
	if fileLogger != nil {
		fileLogger.Printf("INFO: %s\n", msg)
	}
}

// LogError records an error event for later debugging.
func LogError(err error, context string) {
	if fileLogger != nil {
		fileLogger.Printf("ERROR: %s -> %v\n", context, err)
	}
}

// LogAction records a system action like killing a process.
func LogAction(action, details string) {
	if fileLogger != nil {
		fileLogger.Printf("ACTION: %s (%s)\n", action, details)
	}
}
