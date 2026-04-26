package ui

// PresetConfigs defines the four predefined layout presets
var PresetConfigs = map[int]*LayoutConfig{
	0: { // FULL - Complete layout with all panels
		ShowR2D2:    true,
		ShowCPU:     true,
		ShowMemory:  true,
		ShowDisk:    true,
		ShowNetwork: true,
		ShowProcess: true,
		LeftPanelRatio:     0.38,
		NetworkHeightRatio: 0.25,
		PanelPriorities: map[PanelType]int{
			PanelProcess: 1, // Highest priority
			PanelCPU:     2,
			PanelMemory:  3,
			PanelR2D2:    4,
			PanelNetwork: 5,
			PanelDisk:    6, // Lowest priority
		},
	},
	1: { // COMPACT - No R2-D2, more space for processes
		ShowR2D2:    false,
		ShowCPU:     true,
		ShowMemory:  true,
		ShowDisk:    true,
		ShowNetwork: true,
		ShowProcess: true,
		LeftPanelRatio:     0.25, // Smaller left panel
		NetworkHeightRatio: 0.25,
		PanelPriorities: map[PanelType]int{
			PanelProcess: 1,
			PanelCPU:     2,
			PanelMemory:  3,
			PanelNetwork: 4,
			PanelDisk:    5,
		},
	},
	2: { // NET-FOCUS - Enlarged network panel, no disk
		ShowR2D2:    true,
		ShowCPU:     true,
		ShowMemory:  true,
		ShowDisk:    false, // Hidden to make room for network
		ShowNetwork: true,
		ShowProcess: true,
		LeftPanelRatio:     0.35,
		NetworkHeightRatio: 0.60, // 60% more space for network
		PanelPriorities: map[PanelType]int{
			PanelNetwork: 1, // Network gets highest priority
			PanelProcess: 2,
			PanelCPU:     3,
			PanelMemory:  4,
			PanelR2D2:    5,
		},
	},
	3: { // CPU-ONLY - Only CPU and processes for performance debugging
		ShowR2D2:    false,
		ShowCPU:     true,
		ShowMemory:  false,
		ShowDisk:    false,
		ShowNetwork: false,
		ShowProcess: true,
		LeftPanelRatio:     0.30, // Small left panel for CPU
		NetworkHeightRatio: 0.0,  // No network panel
		PanelPriorities: map[PanelType]int{
			PanelCPU:     1, // CPU gets highest priority
			PanelProcess: 2,
		},
	},
}

// PresetNames maps preset numbers to their display names
var PresetNames = map[int]string{
	0: "FULL",
	1: "COMPACT",
	2: "NET-FOCUS",
	3: "CPU-ONLY",
}

// GetPresetConfig returns the configuration for a given preset
func GetPresetConfig(preset int) (*LayoutConfig, error) {
	config, exists := PresetConfigs[preset]
	if !exists {
		return nil, NewLayoutError(ErrorInvalidPreset, "Invalid preset number")
	}
	
	// Return a copy to prevent modification
	configCopy := *config
	configCopy.PanelPriorities = make(map[PanelType]int)
	for k, v := range config.PanelPriorities {
		configCopy.PanelPriorities[k] = v
	}
	
	return &configCopy, nil
}

// GetPresetName returns the display name for a preset
func GetPresetName(preset int) string {
	name, exists := PresetNames[preset]
	if !exists {
		return "UNKNOWN"
	}
	return name
}

// ValidatePreset checks if a preset number is valid
func ValidatePreset(preset int) error {
	if preset < 0 || preset > 3 {
		return NewLayoutError(ErrorInvalidPreset, "Preset must be between 0 and 3")
	}
	return nil
}

// IsPresetSupported checks if a preset can be displayed in the given terminal size
func IsPresetSupported(preset int, width, height int) bool {
	// All presets support minimal fallback
	if width < 80 || height < 24 {
		return true // Will use minimal layout
	}
	
	_, err := GetPresetConfig(preset)
	if err != nil {
		return false
	}
	
	// Calculate minimum required space for this preset
	minWidth := 80  // Base minimum
	minHeight := 24 // Base minimum
	
	// Adjust based on preset requirements
	switch preset {
	case 0: // FULL - needs space for all panels
		minWidth = 90
		minHeight = 30
	case 1: // COMPACT - more efficient
		minWidth = 80
		minHeight = 25
	case 2: // NET-FOCUS - needs height for network panel
		minWidth = 85
		minHeight = 28
	case 3: // CPU-ONLY - most efficient
		minWidth = 70
		minHeight = 20
	}
	
	return width >= minWidth && height >= minHeight
}

// GetAvailablePresets returns all available presets
func GetAvailablePresets() []PresetInfo {
	presets := make([]PresetInfo, 0, len(PresetConfigs))
	for i := 0; i < 4; i++ {
		presets = append(presets, PresetInfo{
			ID:   i,
			Name: GetPresetName(i),
		})
	}
	return presets
}

// PresetController manages layout preset operations
type PresetController struct {
	currentPreset   int
	layoutConfig    *LayoutConfig
	calculator      *LayoutCalculator
	transitionState TransitionState
}

// TransitionState tracks preset change animations
type TransitionState struct {
	IsTransitioning bool
	FromPreset      int
	ToPreset        int
	Progress        float64
}

// NewPresetController creates a new preset controller
func NewPresetController(initialPreset int) *PresetController {
	controller := &PresetController{
		currentPreset: initialPreset,
		calculator:    NewLayoutCalculator(),
	}
	
	// Load initial configuration
	config, err := GetPresetConfig(initialPreset)
	if err != nil {
		// Fallback to preset 0 if invalid
		config, _ = GetPresetConfig(0)
		controller.currentPreset = 0
	}
	controller.layoutConfig = config
	
	return controller
}

// GetCurrentLayout returns the current layout configuration
func (pc *PresetController) GetCurrentLayout() *LayoutConfig {
	return pc.layoutConfig
}

// SetPreset changes to a new preset
func (pc *PresetController) SetPreset(preset int) error {
	if err := ValidatePreset(preset); err != nil {
		return err
	}
	
	if preset == pc.currentPreset {
		return nil // No change needed
	}
	
	// Start transition
	pc.transitionState = TransitionState{
		IsTransitioning: true,
		FromPreset:      pc.currentPreset,
		ToPreset:        preset,
		Progress:        0.0,
	}
	
	// Load new configuration
	config, err := GetPresetConfig(preset)
	if err != nil {
		pc.transitionState.IsTransitioning = false
		return err
	}
	
	pc.currentPreset = preset
	pc.layoutConfig = config
	
	// Complete transition immediately (no animation for now)
	pc.transitionState.IsTransitioning = false
	pc.transitionState.Progress = 1.0
	
	return nil
}

// CyclePreset advances to the next preset (0->1->2->3->0)
func (pc *PresetController) CyclePreset() error {
	nextPreset := (pc.currentPreset + 1) % 4
	return pc.SetPreset(nextPreset)
}

// CalculateDimensions computes panel dimensions for current layout
func (pc *PresetController) CalculateDimensions(width, height int) *PanelDimensions {
	return pc.calculator.Calculate(pc.layoutConfig, width, height)
}

// GetPresetName returns the name of the current preset
func (pc *PresetController) GetPresetName(preset int) string {
	return GetPresetName(preset)
}

// GetCurrentPresetName returns the name of the current preset
func (pc *PresetController) GetCurrentPresetName() string {
	return GetPresetName(pc.currentPreset)
}

// GetCurrentPresetNumber returns the current preset number
func (pc *PresetController) GetCurrentPresetNumber() int {
	return pc.currentPreset
}

// GetAvailablePresets returns all available presets
func (pc *PresetController) GetAvailablePresets() []PresetInfo {
	return GetAvailablePresets()
}

// ValidatePreset checks if a preset is valid
func (pc *PresetController) ValidatePreset(preset int) error {
	return ValidatePreset(preset)
}

// IsPresetSupported checks if a preset is supported for the given terminal size
func (pc *PresetController) IsPresetSupported(preset int, width, height int) bool {
	return IsPresetSupported(preset, width, height)
}

// IsTransitioning returns true if a preset transition is in progress
func (pc *PresetController) IsTransitioning() bool {
	return pc.transitionState.IsTransitioning
}

// GetTransitionState returns the current transition state
func (pc *PresetController) GetTransitionState() TransitionState {
	return pc.transitionState
}