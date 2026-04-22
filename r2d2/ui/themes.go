package ui

import "github.com/charmbracelet/lipgloss"

// Theme defines the color palette for the terminal user interface.
type Theme struct {
	ID, Name   string
	CPU, RAM   lipgloss.Color
	DSK        lipgloss.Color
	CharMain   lipgloss.Color
	CharAccent lipgloss.Color
}

// Themes holds the collection of available R2-D2 inspired visual styles.
var Themes = []Theme{
	{ID: "astromech", Name: "Astromech",
		CPU: "#2979FF", RAM: "#E0E0E0", DSK: "#90A4AE", CharMain: "#90A4AE", CharAccent: "#2979FF"},
	{ID: "logic", Name: "Logic Display",
		CPU: "#F44336", RAM: "#00D2FF", DSK: "#ECEFF1", CharMain: "#90A4AE", CharAccent: "#F44336"},
	{ID: "hologram", Name: "Hologram",
		CPU: "#00E5FF", RAM: "#00B8D4", DSK: "#4DD0E1", CharMain: "#4DD0E1", CharAccent: "#00E5FF"},
	{ID: "sunset", Name: "Binary Sunset",
		CPU: "#FFB300", RAM: "#2979FF", DSK: "#1A237E", CharMain: "#E65100", CharAccent: "#FFB300"},
}
