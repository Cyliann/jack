package tui

import "github.com/charmbracelet/lipgloss"

var (
	fileStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("211"))
	titleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("93"))
	doneStyle    = lipgloss.NewStyle().Margin(1, 2)
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).SetString("✗")
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
	etaStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	sizeStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)
