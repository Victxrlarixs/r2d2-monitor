package r2d2

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds the user preferences for the R2-D2 Monitor.
type Config struct {
	ThemeIdx int    `json:"theme_idx"`
	SortBy   string `json:"sort_by"`
}

// DefaultConfig returns the standard settings for a fresh install.
func DefaultConfig() Config {
	return Config{
		ThemeIdx: 0,
		SortBy:   "CPU",
	}
}

// GetConfigPath returns the OS-specific path for the configuration file.
func GetConfigPath() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".r2d2-monitor")
	_ = os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "config.json")
}

// LoadConfig reads the configuration from disk or returns defaults if missing.
func LoadConfig() Config {
	path := GetConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultConfig()
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig()
	}
	return cfg
}

// SaveConfig persists the current settings to disk.
func SaveConfig(cfg Config) error {
	path := GetConfigPath()
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
