package styles

import (
	"github.com/charmbracelet/lipgloss"
)

// UI styles
var (
	BaseStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("255")).
			Padding(0, 1)

	FocusedStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("2")).
			Padding(0, 1)

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("62")).
			Padding(1, 0, 0, 2)

	HighlightFlagStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("190"))
	PrivateflagStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("013"))
)
