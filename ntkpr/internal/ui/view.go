package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/haochend413/lipgloss/v2"

	// "charm.land/lipgloss/v2"

	// "github.com/charmbracelet/lipgloss"
	"github.com/haochend413/ntkpr/internal/ui/styles"
)

// define modular view functions
func (m Model) quitConfirmView() tea.View {
	v := tea.NewView("You sure you wanna quit ntkpr? This will reset recent table since this feature is not yet supported.\nTo quit, type y.\nTo go back, type n.")
	v.AltScreen = true
	return v
}

func (m Model) appView() tea.View {
	if !m.ready {
		v := tea.NewView("Initializing...")
		v.AltScreen = true
		return v
	}

	// Main content height
	mainContentHeight := m.height - 3 // Reserve 3 lines for help + status bar

	// Render left side: always show all three tables. When editing, render the previously
	// focused table with the OnEdit style and the others with Base style.
	var threadsBox, branchesBox, notesBox string
	if m.focus == FocusEdit {
		// Apply styles for rendering (mutating here is fine — rendering reflects current mode)
		switch m.previousFocus {
		case FocusThreads:
			m.threadsTable.SetStyles(styles.FocusedTableStyleOnEdit)
			m.branchesTable.SetStyles(styles.BaseTableStyle)
			m.notesTable.SetStyles(styles.BaseTableStyle)
			threadsBox = styles.BaseStyle.BorderTitle("[1]-Threads (Editing)").Render(m.threadsTable.View())
			branchesBox = styles.BaseStyle.BorderTitle("Branches").Render(m.branchesTable.View())
			notesBox = styles.BaseStyle.BorderTitle("Notes").Render(m.notesTable.View())
		case FocusBranches:
			m.threadsTable.SetStyles(styles.BaseTableStyle)
			m.branchesTable.SetStyles(styles.FocusedTableStyleOnEdit)
			m.notesTable.SetStyles(styles.BaseTableStyle)
			threadsBox = styles.BaseStyle.BorderTitle("Threads").Render(m.threadsTable.View())
			branchesBox = styles.BaseStyle.BorderTitle("[2]-Branches (Editing)").Render(m.branchesTable.View())
			notesBox = styles.BaseStyle.BorderTitle("Notes").Render(m.notesTable.View())
		default:
			m.threadsTable.SetStyles(styles.BaseTableStyle)
			m.branchesTable.SetStyles(styles.BaseTableStyle)
			m.notesTable.SetStyles(styles.FocusedTableStyleOnEdit)
			threadsBox = styles.BaseStyle.BorderTitle("Threads").Render(m.threadsTable.View())
			branchesBox = styles.BaseStyle.BorderTitle("Branches").Render(m.branchesTable.View())
			notesBox = styles.BaseStyle.BorderTitle("[3]-Notes (Editing)").Render(m.notesTable.View())
		}
	} else {
		threadsBox = m.renderThreadsTableBox()
		branchesBox = m.renderBranchesTableBox()
		notesBox = m.renderNotesTableBox()
	}
	leftSide := lipgloss.JoinVertical(lipgloss.Left,
		threadsBox,
		branchesBox,
		notesBox,
	)

	// Render right side (textarea and changelog)
	var editBox string
	if m.focus == FocusEdit {
		// re-render previous table
		switch m.previousFocus {
		case FocusBranches:

		}
		editBox = styles.FocusedStyle.BorderTitle("[4]-Editor").Render(m.textArea.View())
	} else {
		editBox = styles.BaseStyle.BorderTitle("[4]-Editor").Render(m.textArea.View())
	}

	changelogBox := m.renderChangelogTableBox()

	rightSide := lipgloss.JoinVertical(lipgloss.Left,
		// styles.TitleStyle.Render("Editor"),
		editBox,
		// styles.TitleStyle.Render("Changes"),
		changelogBox,
	)

	// Join left and right sides horizontally
	mainContent := lipgloss.NewStyle().
		Height(mainContentHeight).
		Render(lipgloss.JoinHorizontal(lipgloss.Top, leftSide, rightSide))

	help := ""
	if m.focus == FocusEdit {
		// more complete, multi-line help for edit mode
		help = styles.HelpStyle.Render(
			"Arrows: move • Home/End: line start/end • Alt/Option+←/→: word backward/forward • c-k: del-after • c-u: del-before • " +
				"c-h/Backspace: del-back • Del/c-d: del-forward • Alt/Option+c/l/u: Cap/Lower/Upper • c-t: transpose",
		)
	} else {
		// Global/table help derived from tableKeys and globalKeys
		help = styles.HelpStyle.Render(
			"Tab: tables • Enter: select • Esc: back/cancel • e: edit • n: new • r: recent edits • " +
				"k/j: move to upper/lower item • l/h: move to upper/lower table • c-d: delete • c-h: highlight • c-p: private • c-l: changelog • " +
				"c-s: save • c-q: sync • c-c: quit",
		)
	}

	// Render status bar
	statusBarBox := m.statusBar.View()

	// Create base layer with original content
	baseContent := lipgloss.JoinVertical(lipgloss.Top,
		mainContent,
		help,
		statusBarBox.Content,
	)

	// Create compositor for layering
	// Base layer with original content
	baseLayer := lipgloss.NewLayer(baseContent).Z(0)

	var compositor *lipgloss.Compositor
	var output string

	// Only show recent table overlay when explicitly focused on it (via R key)
	if m.focus == FocusRecent {
		// Recent table layer positioned in the middle
		recentTableBox := m.renderRecentTableBox()
		recentLayer := lipgloss.NewLayer(recentTableBox).
			X((m.width - lipgloss.Width(recentTableBox)) / 2).
			Y((m.height - lipgloss.Height(recentTableBox)) / 2).
			Z(1)
		// Create compositor with both layers
		compositor = lipgloss.NewCompositor(baseLayer, recentLayer)
		output = compositor.Render()
	} else {
		// No recent focus, just show base layer
		compositor = lipgloss.NewCompositor(baseLayer)
		output = compositor.Render()
	}

	// Create view with composited output
	v := tea.NewView(output)
	v.AltScreen = true
	return v
}

// View renders the UI
func (m Model) View() tea.View {
	var v tea.View
	switch m.viewMode {
	case ApplicationView:
		v = m.appView()
	case QuitConfirmView:
		v = m.quitConfirmView()
	default:
		v = m.appView()
	}
	return v
}

func (m Model) renderThreadsTableBox() string {
	if m.focus == FocusThreads {
		m.threadsTable.SetStyles(styles.FocusedTableStyle)
		return styles.FocusedStyle.BorderTitle("[1]-Threads").Render(m.threadsTable.View())
	} else {
		m.threadsTable.SetStyles(styles.BaseTableStyle)
		return styles.BaseStyle.BorderTitle("Threads").Render(m.threadsTable.View())
	}
}

func (m Model) renderBranchesTableBox() string {
	if m.focus == FocusBranches {
		m.branchesTable.SetStyles(styles.FocusedTableStyle)
		return styles.FocusedStyle.BorderTitle("[2]-Branches").Render(m.branchesTable.View())
	} else {
		m.branchesTable.SetStyles(styles.BaseTableStyle)
		return styles.BaseStyle.BorderTitle("Branches").Render(m.branchesTable.View())
	}
}

func (m Model) renderNotesTableBox() string {
	if m.focus == FocusNotes {
		m.notesTable.SetStyles(styles.FocusedTableStyle)
		return styles.FocusedStyle.BorderTitle("[3]-Notes").Render(m.notesTable.View())
	} else {
		m.notesTable.SetStyles(styles.BaseTableStyle)
		return styles.BaseStyle.BorderTitle("Notes").Render(m.notesTable.View())
	}
}

func (m Model) renderChangelogTableBox() string {
	if m.focus == FocusChangelog {
		m.changeTable.SetStyles(styles.FocusedTableStyle)
		return styles.FocusedStyle.BorderTitle("Changelog").Render(m.changeTable.View())
	} else {
		m.changeTable.SetStyles(styles.BaseTableStyle)
		return styles.BaseStyle.BorderTitle("[5]-Changelog").Render(m.changeTable.View())
	}
}

func (m Model) renderRecentTableBox() string {
	m.recentTable.SetStyles(styles.FocusedTableStyle)
	return styles.FocusedStyle.
		BorderTitle("Recent Edits").
		Render(m.recentTable.View())
}
