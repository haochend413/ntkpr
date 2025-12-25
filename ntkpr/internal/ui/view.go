package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/haochend413/ntkpr/internal/ui/styles"
)

// View renders the UI
func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// Main content height is set in Update, use it for layout
	mainContentHeight := m.height - 3 // Reserve 3 lines for help + status bar

	// Render left side (table and optionally search)
	var leftSide string
	if m.focus == FocusSearch {
		// Both search and table visible
		searchBox := styles.FocusedStyle.Render(m.searchInput.View())
		tableBox := m.renderTableBox()

		leftSide = lipgloss.JoinVertical(lipgloss.Top,
			searchBox,
			tableBox,
		)
	} else {
		// Only table visible
		tableBox := m.renderTableBox()
		leftSide = tableBox
	}

	// Render right side (textarea, topics table, topic input)
	var editBox string
	if m.focus == FocusEdit {
		editBox = styles.FocusedStyle.Render(m.textarea.View())
	} else {
		editBox = styles.BaseStyle.Render(m.textarea.View())
	}

	var topicsTableBox string
	if m.focus == FocusTopics {
		topicsTableBox = styles.FocusedStyle.Render(m.topicsTable.View())
	} else {
		topicsTableBox = styles.BaseStyle.Render(m.topicsTable.View())
	}

	var topicInputBox string
	if m.focus == FocusTopics {
		topicInputBox = styles.FocusedStyle.Render(m.topicInput.View())
	} else {
		topicInputBox = styles.BaseStyle.Render(m.topicInput.View())
	}

	rightSide := lipgloss.JoinVertical(lipgloss.Left,
		editBox,
		styles.TitleStyle.Render("Topics"),
		topicsTableBox,
		styles.TitleStyle.Render("Add Topics"),
		topicInputBox,
	)

	// Join left and right sides horizontally
	mainContent := lipgloss.NewStyle().
		Height(mainContentHeight).
		Render(lipgloss.JoinHorizontal(lipgloss.Top, leftSide, rightSide))

	help := styles.HelpStyle.Render(
		"Tab: cycle focus • Enter: select/search/add-topic • /: search • Ctrl+N: new note (table only) • Ctrl+S: save • Ctrl+Q: sync DB • Del: delete note/topic • Ctrl+C: quit",
	)

	// Render status bar
	statusBarBox := m.statusBar.View()

	// Join everything vertically
	return lipgloss.JoinVertical(lipgloss.Top,
		mainContent,
		help,
		statusBarBox,
	)
}

func (m Model) renderTableBox() string {
	if m.focus == FocusTable {
		m.table.SetStyles(styles.FocusedTableStyle)
		return styles.FocusedStyle.Render(m.table.View())
	} else {
		m.table.SetStyles(styles.BaseTableStyle)
		return styles.BaseStyle.Render(m.table.View())
	}
}
