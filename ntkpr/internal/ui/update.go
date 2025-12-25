package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/haochend413/bubbles/key"
	"github.com/haochend413/bubbles/table"
	"github.com/haochend413/ntkpr/internal/app/context"
	"github.com/haochend413/ntkpr/state"
)

// Global keys that work in any mode
type globalKeyMap struct {
	QuitApp       key.Binding
	SwitchContext key.Binding
}

var globalKeys = globalKeyMap{
	QuitApp:       key.NewBinding(key.WithKeys("ctrl+c")),
	SwitchContext: key.NewBinding(key.WithKeys("tab")),
}

// Table focus keys
type tableKeyMap struct {
	GoToTextArea         key.Binding
	CreateNewNote        key.Binding
	SyncWithDB           key.Binding
	Retract              key.Binding
	DeleteNote           key.Binding
	SwitchCtxSearch      key.Binding
	SwitchCtxRecent      key.Binding
	SwitchCtxDefault     key.Binding
	HighlightCurrentNote key.Binding
	PrivatizeCurrentNote key.Binding
}

var tableKeys = tableKeyMap{
	GoToTextArea:         key.NewBinding(key.WithKeys("enter")),
	CreateNewNote:        key.NewBinding(key.WithKeys("ctrl+n", "n")),
	SyncWithDB:           key.NewBinding(key.WithKeys("ctrl+q")),
	Retract:              key.NewBinding(key.WithKeys("ctrl+z")),
	DeleteNote:           key.NewBinding(key.WithKeys("ctrl+d")),
	HighlightCurrentNote: key.NewBinding((key.WithKeys("ctrl+h"))),
	PrivatizeCurrentNote: key.NewBinding(key.WithKeys("ctrl+p")),
	SwitchCtxSearch:      key.NewBinding(key.WithKeys("S")),
	SwitchCtxRecent:      key.NewBinding(key.WithKeys("R")),
	SwitchCtxDefault:     key.NewBinding(key.WithKeys("A")),
}

// Search focus keys
type searchKeyMap struct {
	Enter key.Binding
}

var searchKeys = searchKeyMap{
	Enter: key.NewBinding(key.WithKeys("enter")),
}

// Edit focus keys
type editKeyMap struct {
	SaveCurrentNote key.Binding
}

var editKeys = editKeyMap{
	SaveCurrentNote: key.NewBinding(key.WithKeys("ctrl+s")),
}

// Topics focus keys
type topicsKeyMap struct {
	AddTopic    key.Binding
	DeleteTopic key.Binding
}

var topicsKeys = topicsKeyMap{
	AddTopic:    key.NewBinding(key.WithKeys("enter")),
	DeleteTopic: key.NewBinding(key.WithKeys("ctrl+d")),
}

// Update handles UI events and updates the model
// On startup settings ? Yeah this is definitely important.

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.statusBar.SetWidth(m.width)

		borderOverhead := 8 // 4 chars per box (2 borders + 2 padding) * 2 boxes
		availableWidth := m.width - borderOverhead

		// Split: 40% for table (left), 60% for edit (right)
		tableContentWidth := int(float64(availableWidth) * 0.4)
		editContentWidth := availableWidth - tableContentWidth

		// Table actual width includes its content
		tableWidth := tableContentWidth
		editWidth := editContentWidth

		// Distribute column widths to use full available width
		idWidth := max(4, int(float64(tableWidth)*0.08))
		timeWidth := max(8, int(float64(tableWidth)*0.22))
		flagWidth := max(4, int(float64(tableWidth)*0.10))
		contentWidth := max(10, int(float64(tableWidth)*0.48))
		topicsWidth := max(5, int(float64(tableWidth)*0.15))

		columns := []table.Column{
			{Title: "ID", Width: idWidth},
			{Title: "Time", Width: timeWidth},
			{Title: "Content", Width: contentWidth},
			{Title: "Flags", Width: flagWidth},
			{Title: "Topics", Width: topicsWidth},
		}
		m.table.SetColumns(columns)
		m.table.SetWidth(tableWidth)

		// Height calculations
		// Total height budget: m.height
		// Used by: main content + help (3 lines) + status bar (1 line)
		// Help has Padding(1, 0, 0, 2) so it's 2 lines total
		// Status bar is 1 line
		// Total reserved for help + status = 3 lines
		mainContentHeight := m.height - 3

		// Table height depends on whether search is visible
		var tableHeight int
		if m.focus == FocusSearch {
			// Search box takes ~3 lines (border + padding + input)
			// Remaining space for table
			tableHeight = max(5, mainContentHeight-6)
		} else {
			// Full height for table (minus some margin)
			tableHeight = max(5, mainContentHeight-3)
		}
		m.table.SetHeight(tableHeight)

		// Textarea dimensions
		// Textarea should fill most of the right side
		m.textarea.SetWidth(editWidth)
		// Reserve space for topics section below
		textareaHeight := max(5, int(float64(mainContentHeight)*0.65))
		m.textarea.SetHeight(textareaHeight)

		// Search input width matches table width
		m.searchInput.Width = tableWidth

		// Topic input width matches edit width
		m.topicInput.Width = editWidth

		// Topics table columns and dimensions
		topicColumns := []table.Column{
			{Title: "Topic", Width: max(10, editWidth-2)},
		}
		m.topicsTable.SetColumns(topicColumns)
		m.topicsTable.SetWidth(editWidth)
		// Topics table height: use remaining space
		topicsTableHeight := max(3, int(float64(mainContentHeight)*0.12))
		m.topicsTable.SetHeight(topicsTableHeight)

	case tea.KeyMsg:
		// Handle global keys first
		switch {
		case key.Matches(msg, globalKeys.QuitApp):
			s := m.CollectState()
			state.SaveState(m.Config.StateFilePath, s) // on quit, save state
			m.app.SaveCurrentNote(m.textarea.Value())
			m.app.SyncWithDatabase()
			return m, tea.Quit
		case key.Matches(msg, globalKeys.SwitchContext):
			switch m.focus {
			case FocusSearch:
				m.focus = FocusTable
				m.searchInput.SetValue("")
				m.table.Focus()
				m.searchInput.Blur()
				m.topicInput.Blur()
				m.topicsTable.Blur()
			case FocusTable:
				m.focus = FocusEdit
				m.table.Blur()
				m.textarea.Focus()
				m.topicInput.Blur()
				m.topicsTable.Blur()
			case FocusEdit:
				m.app.SaveCurrentNote(m.textarea.Value())
				m.updateTable(context.Default)
				m.focus = FocusTopics
				m.textarea.Blur()
				m.topicInput.Focus()
				m.topicsTable.Focus()
			case FocusTopics:
				m.focus = FocusTable
				m.topicInput.Blur()
				m.topicsTable.Blur()
				m.table.Focus()
			}
			m.updateStatusBar()
		}

		// Handle mode-specific keys
		switch m.focus {
		case FocusSearch:
			switch {
			case key.Matches(msg, searchKeys.Enter):
				m.app.SearchNotes(m.searchInput.Value())
				m.focus = FocusTable
				m.table.Focus()
				m.searchInput.Blur()
				m.updateTable(context.Search)
				// Reset cursor to first result
				if len(m.app.GetCurrentNotes()) > 0 {
					m.table.SetCursor(0)
					m.app.SelectCurrentNote(0)
					m.textarea.SetValue(m.app.CurrentNoteContent())
					m.updateTopicsTable()
				}
				m.updateStatusBar()
			}

		case FocusTable:
			switch {
			case key.Matches(msg, tableKeys.GoToTextArea):
				m.app.SelectCurrentNote(m.table.Cursor())
				m.textarea.SetValue(m.app.CurrentNoteContent())
				m.focus = FocusEdit
				m.table.Blur()
				m.textarea.Focus()
				m.topicInput.Blur()
				m.topicsTable.Blur()
				m.updateTopicsTable()
				return m, nil

			case key.Matches(msg, tableKeys.CreateNewNote):
				m.app.CreateNewNote()
				m.updateTable(context.Default) // Update table to make sure new one is loaded

				// Set cursor
				lastIdx := len(m.app.GetCurrentNotes()) - 1
				if lastIdx >= 0 {
					m.table.SetCursor(lastIdx)
					m.app.SelectCurrentNote(lastIdx)
				}

				m.textarea.SetValue(m.app.CurrentNoteContent())
				m.updateStatusBar()

				// Set focus to edit
				m.focus = FocusEdit
				m.table.Blur()
				m.textarea.Focus()
				m.topicInput.Blur()
				m.topicsTable.Blur()
				return m, nil

			case key.Matches(msg, tableKeys.HighlightCurrentNote):
				m.app.ToggleCurrentNoteHighlight()
				m.updateTable(m.CurrentContext)
				return m, nil

			case key.Matches(msg, tableKeys.PrivatizeCurrentNote):
				m.app.ToggleCurrentNotePrivate()
				m.updateTable(m.CurrentContext)
				return m, nil

			case key.Matches(msg, tableKeys.SyncWithDB):
				m.app.SaveCurrentNote(m.textarea.Value())
				m.app.SyncWithDatabase()
				m.app.UpdateRecentNotes()
				m.updateTable(m.CurrentContext)
				m.updateTopicsTable()

				// Ensure cursor is valid after sync
				if len(m.app.GetCurrentNotes()) > 0 {
					cursor := m.table.Cursor()
					if cursor >= len(m.app.GetCurrentNotes()) {
						cursor = len(m.app.GetCurrentNotes()) - 1
						m.table.SetCursor(cursor)
					}
					m.app.SelectCurrentNote(cursor)
					m.textarea.SetValue(m.app.CurrentNoteContent())
				}
				m.updateStatusBar()
				return m, nil

			case key.Matches(msg, tableKeys.Retract):
				// Get the last deleted note ID before undo
				var restoredNoteID uint
				for i := len(m.app.GetEditStack()) - 1; i >= 0; i-- {
					id := m.app.GetEditStack()[i]
					if edit := m.app.GetEdit(id); edit != nil && edit.EditType == 2 { // Delete type
						restoredNoteID = id
						break
					}
				}

				m.app.UndoDelete()
				m.app.UpdateCurrentList(m.CurrentContext)
				m.updateTable(m.CurrentContext)

				// Find the position of the restored note
				if len(m.app.GetCurrentNotes()) > 0 && restoredNoteID > 0 {
					foundIdx := 0
					for i, note := range m.app.GetCurrentNotes() {
						if note.ID == restoredNoteID {
							foundIdx = i
							break
						}
					}
					m.table.SetCursor(foundIdx)
					m.app.SelectCurrentNote(foundIdx)
					m.textarea.SetValue(m.app.CurrentNoteContent())
					m.updateTopicsTable()
				}
				m.updateStatusBar()

			case key.Matches(msg, tableKeys.DeleteNote):
				oldCursor := m.table.Cursor()
				m.app.DeleteCurrentNote(uint(oldCursor))
				m.updateTable(m.CurrentContext)

				// Keep cursor at same position (shows next item naturally)
				if len(m.app.GetCurrentNotes()) > 0 {
					newCursor := oldCursor
					if newCursor >= len(m.app.GetCurrentNotes()) {
						newCursor = len(m.app.GetCurrentNotes()) - 1
					}
					m.table.SetCursor(newCursor)
					m.app.SelectCurrentNote(newCursor)
					m.textarea.SetValue(m.app.CurrentNoteContent())
					m.updateTopicsTable()
				} else {
					m.app.SelectCurrentNote(0)
					m.textarea.SetValue("")
					m.updateTopicsTable()
				}
				m.updateStatusBar()

			case key.Matches(msg, tableKeys.SwitchCtxSearch):
				m.CurrentContext = context.Search
				m.app.UpdateCurrentList(m.CurrentContext)
				m.focus = FocusSearch
				m.searchInput.Focus()
				m.table.Blur()
				return m, nil

			case key.Matches(msg, tableKeys.SwitchCtxRecent):
				m.CurrentContext = context.Recent
				m.app.UpdateCurrentList(m.CurrentContext)
				m.updateTable(context.Recent)
				if len(m.app.GetCurrentNotes()) > 0 {
					m.table.SetCursor(0)
					m.app.SelectCurrentNote(0)
					m.textarea.SetValue(m.app.CurrentNoteContent())
					m.updateTopicsTable()
				}
				m.updateStatusBar()

			case key.Matches(msg, tableKeys.SwitchCtxDefault):
				m.CurrentContext = context.Default
				m.app.UpdateCurrentList(m.CurrentContext)
				m.updateTable(context.Default)
				if len(m.app.GetCurrentNotes()) > 0 {
					m.table.SetCursor(0)
					m.app.SelectCurrentNote(0)
					m.textarea.SetValue(m.app.CurrentNoteContent())
					m.updateTopicsTable()
				}
				m.updateStatusBar()
			}

		case FocusEdit:
			switch {
			case key.Matches(msg, editKeys.SaveCurrentNote):
				m.app.SaveCurrentNote(m.textarea.Value())
				m.updateTable(context.Default)
				m.focus = FocusTable
				m.table.Focus()
				m.textarea.Blur()
				m.searchInput.Blur()
				m.topicInput.Blur()
				m.topicsTable.Blur()
				m.updateStatusBar()
			}

		case FocusTopics:
			switch {
			case key.Matches(msg, topicsKeys.AddTopic):
				m.app.AddTopicsToCurrentNote(m.topicInput.Value())
				m.topicInput.SetValue("")
				m.updateTopicsTable()
				m.updateStatusBar()

			case key.Matches(msg, topicsKeys.DeleteTopic):
				if m.app.HasCurrentNote() && len(m.app.CurrentNoteTopics()) > 0 {
					// Remove the selected topic from the current note
					topics := m.app.CurrentNoteTopics()
					cursor := m.topicsTable.Cursor()
					if cursor < len(topics) && cursor >= 0 {
						m.app.RemoveTopicFromCurrentNote(topics[cursor].Topic)
						m.updateTopicsTable()
						// Adjust cursor if necessary
						if len(m.app.CurrentNoteTopics()) > 0 && cursor >= len(m.app.CurrentNoteTopics()) {
							m.topicsTable.SetCursor(len(m.app.CurrentNoteTopics()) - 1)
						}
					}
				}
				m.updateTable(context.Default) // Update note table to reflect topic changes
			}
		}

	case table.MoveSelectMsg:
		switch m.focus {
		case FocusTable:
			m.app.SelectCurrentNote(m.table.Cursor())
			m.textarea.SetValue(m.app.CurrentNoteContent())
			m.updateTopicsTable()
		case FocusTopics:
			// No action needed for topic table selection, but ensure cursor is valid
			if len(m.app.CurrentNoteTopics()) > 0 && m.topicsTable.Cursor() >= len(m.app.CurrentNoteTopics()) {
				m.topicsTable.SetCursor(len(m.app.CurrentNoteTopics()) - 1)
			}
		}
		m.updateStatusBar()
	}

	switch m.focus {
	case FocusSearch:
		m.searchInput, cmd = m.searchInput.Update(msg)
		cmds = append(cmds, cmd)
	case FocusTable:
		m.table, cmd = m.table.Update(msg)
		cmds = append(cmds, cmd)
	case FocusEdit:
		m.textarea, cmd = m.textarea.Update(msg)
		cmds = append(cmds, cmd)
	case FocusTopics:
		var cmdTopicInput, cmdTopicsTable tea.Cmd
		m.topicInput, cmdTopicInput = m.topicInput.Update(msg)
		m.topicsTable, cmdTopicsTable = m.topicsTable.Update(msg)
		cmds = append(cmds, cmdTopicInput, cmdTopicsTable)
	}

	return m, tea.Batch(cmds...)
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
