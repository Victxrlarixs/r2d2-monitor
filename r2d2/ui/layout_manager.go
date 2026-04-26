package ui

import (
	"fmt"
)

// PanelType represents different UI panels
type PanelType int

const (
	PanelR2D2 PanelType = iota
	PanelCPU
	PanelMemory
	PanelDisk
	PanelNetwork
	PanelProcess
)

// Rectangle defines a rectangular area with position and dimensions
type Rectangle struct {
	X, Y, Width, Height int
}

// PanelDimensions holds the calculated dimensions for all panels
type PanelDimensions struct {
	R2D2Box    Rectangle
	CPUBox     Rectangle
	MemoryBox  Rectangle
	DiskBox    Rectangle
	NetworkBox Rectangle
	ProcessBox Rectangle
	Footer     Rectangle
}

// LayoutConfig defines which panels to show and their sizing ratios
type LayoutConfig struct {
	ShowR2D2     bool
	ShowCPU      bool
	ShowMemory   bool
	ShowDisk     bool
	ShowNetwork  bool
	ShowProcess  bool
	
	// Dimension ratios (0.0 to 1.0)
	LeftPanelRatio     float64
	NetworkHeightRatio float64
	
	// Panel priorities for responsive behavior (1 = highest priority)
	PanelPriorities map[PanelType]int
}

// LayoutManager interface for managing layout operations
type LayoutManager interface {
	GetCurrentLayout() *LayoutConfig
	SetPreset(preset int) error
	CalculateDimensions(width, height int) *PanelDimensions
	GetPresetName(preset int) string
	GetAvailablePresets() []PresetInfo
	ValidatePreset(preset int) error
	IsPresetSupported(preset int, width, height int) bool
}

// PresetInfo holds information about a layout preset
type PresetInfo struct {
	ID   int
	Name string
}

// ErrorType represents different types of layout errors
type ErrorType int

const (
	ErrorInvalidPreset ErrorType = iota
	ErrorConfigCorrupted
	ErrorTerminalTooSmall
	ErrorLayoutCalculation
	ErrorThemeApplication
)

// LayoutError represents an error in the layout system
type LayoutError struct {
	Type    ErrorType
	Message string
	Context map[string]interface{}
}

func (e *LayoutError) Error() string {
	return fmt.Sprintf("Layout Error [%d]: %s", e.Type, e.Message)
}

// NewLayoutError creates a new layout error
func NewLayoutError(errType ErrorType, message string) *LayoutError {
	return &LayoutError{
		Type:    errType,
		Message: message,
		Context: make(map[string]interface{}),
	}
}

// Minimum panel dimensions
var MinPanelDimensions = map[PanelType]struct{ Width, Height int }{
	PanelR2D2:    {Width: 20, Height: 12},
	PanelCPU:     {Width: 30, Height: 10},
	PanelMemory:  {Width: 25, Height: 8},
	PanelDisk:    {Width: 25, Height: 7},
	PanelNetwork: {Width: 25, Height: 8},
	PanelProcess: {Width: 40, Height: 10},
}

// LayoutCalculator handles the complex logic of panel positioning and sizing
type LayoutCalculator struct {
	minDimensions map[PanelType]struct{ Width, Height int }
	priorities    map[PanelType]int
}

// NewLayoutCalculator creates a new layout calculator
func NewLayoutCalculator() *LayoutCalculator {
	return &LayoutCalculator{
		minDimensions: MinPanelDimensions,
	}
}

// Calculate computes panel dimensions based on layout config and terminal size
func (lc *LayoutCalculator) Calculate(config *LayoutConfig, width, height int) *PanelDimensions {
	// Ensure minimum terminal size
	if width < 80 || height < 24 {
		return lc.calculateMinimal(width, height)
	}

	dims := &PanelDimensions{}
	
	// Calculate basic layout structure
	topH := 10
	footerH := 1
	availableH := height - topH - footerH
	
	leftW := int(float64(width) * config.LeftPanelRatio)
	rightW := width - leftW
	
	// Ensure minimum widths
	if leftW < 25 {
		leftW = 25
		rightW = width - leftW
	}
	if rightW < 40 {
		rightW = 40
		leftW = width - rightW
	}

	// Top row: R2-D2 and CPU
	if config.ShowR2D2 && config.ShowCPU {
		dims.R2D2Box = Rectangle{X: 0, Y: 0, Width: leftW, Height: topH}
		dims.CPUBox = Rectangle{X: leftW, Y: 0, Width: rightW, Height: topH}
	} else if config.ShowCPU {
		dims.CPUBox = Rectangle{X: 0, Y: 0, Width: width, Height: topH}
	}

	// Left column panels (Memory, Disk, Network)
	currentY := topH
	leftPanels := []struct {
		show   bool
		panel  *Rectangle
		height int
	}{
		{config.ShowMemory, &dims.MemoryBox, 8},
		{config.ShowDisk, &dims.DiskBox, 7},
		{config.ShowNetwork, &dims.NetworkBox, int(float64(availableH) * config.NetworkHeightRatio)},
	}

	// Calculate remaining height for network if it's flexible
	usedH := 0
	for _, p := range leftPanels {
		if p.show && p.panel != &dims.NetworkBox {
			usedH += p.height
		}
	}
	
	if config.ShowNetwork && config.NetworkHeightRatio == 0 {
		leftPanels[2].height = availableH - usedH
		if leftPanels[2].height < 5 {
			leftPanels[2].height = 5
		}
	}

	// Position left panels
	for _, p := range leftPanels {
		if p.show && currentY+p.height <= height-footerH {
			*p.panel = Rectangle{X: 0, Y: currentY, Width: leftW, Height: p.height}
			currentY += p.height
		}
	}

	// Process panel takes the right side
	if config.ShowProcess {
		dims.ProcessBox = Rectangle{X: leftW, Y: topH, Width: rightW, Height: availableH}
	}

	// Footer
	dims.Footer = Rectangle{X: 0, Y: height - footerH, Width: width, Height: footerH}

	return dims
}

// calculateMinimal returns a minimal layout for very small terminals
func (lc *LayoutCalculator) calculateMinimal(width, height int) *PanelDimensions {
	dims := &PanelDimensions{}
	
	// In minimal mode, only show CPU and Process
	topH := min(height/3, 8)
	footerH := 1
	
	dims.CPUBox = Rectangle{X: 0, Y: 0, Width: width, Height: topH}
	dims.ProcessBox = Rectangle{X: 0, Y: topH, Width: width, Height: height - topH - footerH}
	dims.Footer = Rectangle{X: 0, Y: height - footerH, Width: width, Height: footerH}
	
	return dims
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}