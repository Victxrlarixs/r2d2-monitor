package r2d2

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds the user preferences for the R2-D2 Monitor.
type Config struct {
	ThemeIdx     int    `json:"theme_idx"`
	SortBy       string `json:"sort_by"`
	LayoutPreset     int    `json:"layout_preset"` // 0=full 1=compact 2=cpu-focus 3=proc-only
	SelectedDisk     string `json:"selected_disk"`
	SelectedNetInt   string `json:"selected_net_int"`
}

// DefaultConfig returns the standard settings for a fresh install.
func DefaultConfig() Config {
	return Config{
		ThemeIdx:     0,
		SortBy:       "CPU",
		LayoutPreset: 0,
		SelectedDisk: "",
		SelectedNetInt: "",
	}
}

// GetConfigPath returns the path for the configuration file in the same directory as the executable.
func GetConfigPath() string {
	exePath, _ := os.Executable()
	dir := filepath.Dir(exePath)
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
