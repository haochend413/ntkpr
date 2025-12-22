package styles

import (
	"github.com/charmbracelet/lipgloss"
)

// UI styles
var (
	BaseStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240"))

	FocusedStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("123"))

	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("211")).
			Bold(true).
			Padding(0, 1)

	SimpleTopicsStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.HiddenBorder()).
				Padding(0, 0).
				Margin(0, 0)

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Padding(1, 0, 0, 2)

	HighlightFlagStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("190"))
	PrivateflagStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("013"))
)
