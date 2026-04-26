package ui

import (
	"testing"
)

func TestGetPresetConfig(t *testing.T) {
	tests := []struct {
		name     string
		preset   int
		wantErr  bool
		expected *LayoutConfig
	}{
		{
			name:    "Valid preset 0 (FULL)",
			preset:  0,
			wantErr: false,
			expected: &LayoutConfig{
				ShowR2D2: true, ShowCPU: true, ShowMemory: true,
				ShowDisk: true, ShowNetwork: true, ShowProcess: true,
				LeftPanelRatio: 0.38, NetworkHeightRatio: 0.25,
			},
		},
		{
			name:    "Valid preset 1 (COMPACT)",
			preset:  1,
			wantErr: false,
			expected: &LayoutConfig{
				ShowR2D2: false, ShowCPU: true, ShowMemory: true,
				ShowDisk: true, ShowNetwork: true, ShowProcess: true,
				LeftPanelRatio: 0.25, NetworkHeightRatio: 0.25,
			},
		},
		{
			name:    "Valid preset 2 (NET-FOCUS)",
			preset:  2,
			wantErr: false,
			expected: &LayoutConfig{
				ShowR2D2: true, ShowCPU: true, ShowMemory: true,
				ShowDisk: false, ShowNetwork: true, ShowProcess: true,
				LeftPanelRatio: 0.35, NetworkHeightRatio: 0.60,
			},
		},
		{
			name:    "Valid preset 3 (CPU-ONLY)",
			preset:  3,
			wantErr: false,
			expected: &LayoutConfig{
				ShowR2D2: false, ShowCPU: true, ShowMemory: false,
				ShowDisk: false, ShowNetwork: false, ShowProcess: true,
				LeftPanelRatio: 0.30, NetworkHeightRatio: 0.0,
			},
		},
		{
			name:     "Invalid preset -1",
			preset:   -1,
			wantErr:  true,
			expected: nil,
		},
		{
			name:     "Invalid preset 4",
			preset:   4,
			wantErr:  true,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := GetPresetConfig(tt.preset)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("GetPresetConfig() expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("GetPresetConfig() unexpected error: %v", err)
				return
			}
			
			if config == nil {
				t.Errorf("GetPresetConfig() returned nil config")
				return
			}
			
			// Check basic panel visibility
			if config.ShowR2D2 != tt.expected.ShowR2D2 {
				t.Errorf("ShowR2D2 = %v, want %v", config.ShowR2D2, tt.expected.ShowR2D2)
			}
			if config.ShowCPU != tt.expected.ShowCPU {
				t.Errorf("ShowCPU = %v, want %v", config.ShowCPU, tt.expected.ShowCPU)
			}
			if config.ShowMemory != tt.expected.ShowMemory {
				t.Errorf("ShowMemory = %v, want %v", config.ShowMemory, tt.expected.ShowMemory)
			}
			if config.ShowDisk != tt.expected.ShowDisk {
				t.Errorf("ShowDisk = %v, want %v", config.ShowDisk, tt.expected.ShowDisk)
			}
			if config.ShowNetwork != tt.expected.ShowNetwork {
				t.Errorf("ShowNetwork = %v, want %v", config.ShowNetwork, tt.expected.ShowNetwork)
			}
			if config.ShowProcess != tt.expected.ShowProcess {
				t.Errorf("ShowProcess = %v, want %v", config.ShowProcess, tt.expected.ShowProcess)
			}
			
			// Check ratios
			if config.LeftPanelRatio != tt.expected.LeftPanelRatio {
				t.Errorf("LeftPanelRatio = %v, want %v", config.LeftPanelRatio, tt.expected.LeftPanelRatio)
			}
			if config.NetworkHeightRatio != tt.expected.NetworkHeightRatio {
				t.Errorf("NetworkHeightRatio = %v, want %v", config.NetworkHeightRatio, tt.expected.NetworkHeightRatio)
			}
		})
	}
}

func TestGetPresetName(t *testing.T) {
	tests := []struct {
		preset   int
		expected string
	}{
		{0, "FULL"},
		{1, "COMPACT"},
		{2, "NET-FOCUS"},
		{3, "CPU-ONLY"},
		{-1, "UNKNOWN"},
		{4, "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := GetPresetName(tt.preset)
			if result != tt.expected {
				t.Errorf("GetPresetName(%d) = %s, want %s", tt.preset, result, tt.expected)
			}
		})
	}
}

func TestValidatePreset(t *testing.T) {
	tests := []struct {
		preset  int
		wantErr bool
	}{
		{0, false},
		{1, false},
		{2, false},
		{3, false},
		{-1, true},
		{4, true},
		{100, true},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			err := ValidatePreset(tt.preset)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePreset(%d) error = %v, wantErr %v", tt.preset, err, tt.wantErr)
			}
		})
	}
}

func TestPresetController(t *testing.T) {
	t.Run("NewPresetController with valid preset", func(t *testing.T) {
		controller := NewPresetController(1)
		if controller.GetCurrentPresetNumber() != 1 {
			t.Errorf("Expected preset 1, got %d", controller.GetCurrentPresetNumber())
		}
		if controller.GetCurrentPresetName() != "COMPACT" {
			t.Errorf("Expected COMPACT, got %s", controller.GetCurrentPresetName())
		}
	})

	t.Run("NewPresetController with invalid preset falls back to 0", func(t *testing.T) {
		controller := NewPresetController(99)
		if controller.GetCurrentPresetNumber() != 0 {
			t.Errorf("Expected fallback to preset 0, got %d", controller.GetCurrentPresetNumber())
		}
	})

	t.Run("CyclePreset", func(t *testing.T) {
		controller := NewPresetController(0)
		
		// Test cycling through all presets
		expectedSequence := []int{1, 2, 3, 0}
		for i, expected := range expectedSequence {
			err := controller.CyclePreset()
			if err != nil {
				t.Errorf("CyclePreset() step %d error: %v", i, err)
			}
			if controller.GetCurrentPresetNumber() != expected {
				t.Errorf("Step %d: expected preset %d, got %d", i, expected, controller.GetCurrentPresetNumber())
			}
		}
	})

	t.Run("SetPreset", func(t *testing.T) {
		controller := NewPresetController(0)
		
		err := controller.SetPreset(2)
		if err != nil {
			t.Errorf("SetPreset(2) error: %v", err)
		}
		if controller.GetCurrentPresetNumber() != 2 {
			t.Errorf("Expected preset 2, got %d", controller.GetCurrentPresetNumber())
		}
		
		// Test invalid preset
		err = controller.SetPreset(99)
		if err == nil {
			t.Errorf("SetPreset(99) expected error but got none")
		}
		// Should remain at preset 2
		if controller.GetCurrentPresetNumber() != 2 {
			t.Errorf("After invalid preset, expected to remain at 2, got %d", controller.GetCurrentPresetNumber())
		}
	})
}

func TestLayoutCalculator(t *testing.T) {
	calculator := NewLayoutCalculator()
	
	t.Run("Calculate with normal terminal size", func(t *testing.T) {
		config := &LayoutConfig{
			ShowR2D2: true, ShowCPU: true, ShowMemory: true,
			ShowDisk: true, ShowNetwork: true, ShowProcess: true,
			LeftPanelRatio: 0.38, NetworkHeightRatio: 0.25,
		}
		
		dims := calculator.Calculate(config, 120, 40)
		
		// Basic sanity checks
		if dims.CPUBox.Width <= 0 || dims.CPUBox.Height <= 0 {
			t.Errorf("CPU box has invalid dimensions: %+v", dims.CPUBox)
		}
		if dims.ProcessBox.Width <= 0 || dims.ProcessBox.Height <= 0 {
			t.Errorf("Process box has invalid dimensions: %+v", dims.ProcessBox)
		}
		if dims.Footer.Width != 120 || dims.Footer.Height != 1 {
			t.Errorf("Footer has wrong dimensions: %+v", dims.Footer)
		}
	})
	
	t.Run("Calculate with minimal terminal size", func(t *testing.T) {
		config := &LayoutConfig{
			ShowR2D2: true, ShowCPU: true, ShowMemory: true,
			ShowDisk: true, ShowNetwork: true, ShowProcess: true,
			LeftPanelRatio: 0.38, NetworkHeightRatio: 0.25,
		}
		
		dims := calculator.Calculate(config, 70, 15)
		
		// Should fall back to minimal layout
		if dims.CPUBox.Width <= 0 || dims.CPUBox.Height <= 0 {
			t.Errorf("Minimal CPU box has invalid dimensions: %+v", dims.CPUBox)
		}
		if dims.ProcessBox.Width <= 0 || dims.ProcessBox.Height <= 0 {
			t.Errorf("Minimal Process box has invalid dimensions: %+v", dims.ProcessBox)
		}
	})
}