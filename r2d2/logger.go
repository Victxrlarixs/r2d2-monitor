package r2d2

import (
	"log"
	"os"
	"path/filepath"
)

var fileLogger *log.Logger

// InitLogger initializes a file-based logger in a 'logs' directory next to the executable.
func InitLogger() {
	exePath, _ := os.Executable()
	dir := filepath.Dir(exePath)
	logDir := filepath.Join(dir, "logs")
	_ = os.MkdirAll(logDir, 0755)
	
	logPath := filepath.Join(logDir, "r2d2.log")
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
