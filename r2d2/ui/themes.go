package ui

import "github.com/charmbracelet/lipgloss"

// Theme defines the color palette for the entire UI.
type Theme struct {
	Name       string
	CPU        lipgloss.Color
	RAM        lipgloss.Color
	DSK        lipgloss.Color
	CharMain   lipgloss.Color
	CharAccent lipgloss.Color
}

// Themes is a collection of curated color palettes.
var Themes = []Theme{
	{
		Name:       "CYBER-IMPERIAL",
		CPU:        lipgloss.Color("#00E5FF"), // Electric Cyan
		RAM:        lipgloss.Color("#7C4DFF"), // Deep Purple
		DSK:        lipgloss.Color("#00B0FF"), // Sky Blue
		CharMain:   lipgloss.Color("#C9D1D9"), // Ghost White
		CharAccent: lipgloss.Color("#00E5FF"),
	},
	{
		Name:       "SYNTHWAVE",
		CPU:        lipgloss.Color("#FF00FF"), // Hot Pink
		RAM:        lipgloss.Color("#00FFFF"), // Cyan
		DSK:        lipgloss.Color("#7000FF"), // Neon Purple
		CharMain:   lipgloss.Color("#FFFFFF"),
		CharAccent: lipgloss.Color("#FF00FF"),
	},
	{
		Name:       "MATRIX-CODE",
		CPU:        lipgloss.Color("#00FF41"), // Matrix Green
		RAM:        lipgloss.Color("#003B00"), // Dark Green
		DSK:        lipgloss.Color("#0D0208"), // Black-ish
		CharMain:   lipgloss.Color("#00FF41"),
		CharAccent: lipgloss.Color("#008F11"),
	},
	{
		Name:       "TATOOINE",
		CPU:        lipgloss.Color("#FFAB40"), // Sunset Orange
		RAM:        lipgloss.Color("#FFD180"), // Light Sand
		DSK:        lipgloss.Color("#8D6E63"), // Dark Sand
		CharMain:   lipgloss.Color("#FFF8E1"),
		CharAccent: lipgloss.Color("#FFAB40"),
	},
	{
		Name:       "DEATH-STAR",
		CPU:        lipgloss.Color("#FF1744"), // Alert Red
		RAM:        lipgloss.Color("#9E9E9E"), // Metal Gray
		DSK:        lipgloss.Color("#424242"), // Dark Metal
		CharMain:   lipgloss.Color("#F5F5F5"),
		CharAccent: lipgloss.Color("#FF1744"),
	},
}
