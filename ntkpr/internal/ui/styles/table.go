package styles

import (
	// "charm.land/bubbles/table"
	// "charm.land/bubbles/table"
	// "charm.land/bubbles/table"
	// "charm.land/bubbles/table"
	// "charm.land/bubbles/table"
	// "charm.land/bubbles/v2/table"
	// "charm.land/lipgloss/v2"
	"github.com/haochend413/lipgloss/v2"
	// "github.com/haochend413/lipgloss"
	// "github.com/charmbracelet/lipgloss"
	"github.com/haochend413/bubbles/v2/table"
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

var FocusControlTableStyle = table.Styles{
	Header: lipgloss.NewStyle().
		Bold(true).
		Padding(0, 0).
		Foreground(lipgloss.Color("252")),

	Cell: lipgloss.NewStyle().
		Background(lipgloss.Color("20")).
		Padding(0, 0),
	Selected: lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")).
		Background(lipgloss.Color("20")). // purple
		Bold(true),
}

var FocusedTableStyleOnEdit = table.Styles{
	Header: lipgloss.NewStyle().
		Bold(true).
		Padding(0, 0).
		Foreground(lipgloss.Color("252")),

	Cell: lipgloss.NewStyle().
		Padding(0, 0),
	Selected: lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")).
		Background(lipgloss.Color("239")). // grey
		Bold(true),
}
