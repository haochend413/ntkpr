package styles

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/haochend413/bubbles/table"
)

var BaseTableStyle = table.Styles{
	Header: lipgloss.NewStyle().
		Bold(true).
		Padding(0, 0).
		Foreground(lipgloss.Color("252")),

	Cell: lipgloss.NewStyle().
		Padding(0, 0),
	Selected: lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")), // yellow

}

var FocusedTableStyle = table.Styles{
	Header: lipgloss.NewStyle().
		Bold(true).
		Padding(0, 0).
		Foreground(lipgloss.Color("252")),

	Cell: lipgloss.NewStyle().
		Padding(0, 0),
	Selected: lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")).
		Background(lipgloss.Color("20")). // purple
		Bold(true),
}
