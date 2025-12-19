package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/haochend413/bubbles/table"
	"github.com/haochend413/ntkpr/internal/app/context"
)

// Update handles UI events and updates the model
// On startup settings ?

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.statusBar.SetWidth(m.width)

		tableWidth := m.width/2 - 4
		editWidth := m.width/2 - 4
		columns := []table.Column{
			{Title: "ID", Width: 4},
			{Title: "Time", Width: 16},
			{Title: "Content", Width: max(10, tableWidth-70)},
			{Title: "Topics", Width: 20},
		}
		m.table.SetColumns(columns)
		if m.focus == FocusSearch {
			m.table.SetHeight(m.height - 10)
		} else {
			m.table.SetHeight(m.height - 8)
		}
		m.textarea.SetWidth(max(20, editWidth))
		m.textarea.SetHeight(max(5, m.height/2-6))
		m.searchInput.Width = tableWidth - 25
		m.topicInput.Width = max(20, editWidth)
		topicColumns := []table.Column{
			{Title: "Topic", Width: max(20, editWidth-4)},
		}
		m.topicsTable.SetColumns(topicColumns)
		m.topicsTable.SetWidth(max(20, editWidth))
		m.topicsTable.SetHeight(4)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.app.SaveCurrentNote(m.textarea.Value())
			m.app.SyncWithDatabase()
			return m, tea.Quit
		case "tab":
			switch m.focus {
			case FocusSearch:
				m.focus = FocusTable
				m.searchInput.SetValue("")
				m.table.Focus()
				m.searchInput.Blur()
				m.topicInput.Blur()
				m.topicsTable.Blur()
				// m.table.SetHeight(m.height - 8)
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

		case "enter":
			switch m.focus {
			case FocusSearch:
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
				
			case FocusTable:
				m.app.SelectCurrentNote(m.table.Cursor())
				m.textarea.SetValue(m.app.CurrentNoteContent())
				m.focus = FocusEdit
				m.table.Blur()
				m.textarea.Focus()
				m.topicInput.Blur()
				m.topicsTable.Blur()
				m.updateTopicsTable()
				return m, nil
			case FocusTopics:
				m.app.AddTopicsToCurrentNote(m.topicInput.Value())
				m.topicInput.SetValue("")
				m.updateTopicsTable()
			}
			m.updateStatusBar()

		case "ctrl+s":
			if m.focus == FocusEdit {
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
		case "ctrl+n", "n":
			if m.focus == FocusTable {
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
			}

		case "ctrl+q":
			m.app.SaveCurrentNote(m.textarea.Value())
			m.app.SyncWithDatabase()
			m.app.UpdateRecentNotes()
			m.updateTable(m.NoteSelector)
			m.updateTopicsTable()
			m.updateFullTopicTable()

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
		case "ctrl+z":
			if m.focus == FocusTable {
				restoredNoteID := m.app.GetLastDeletedNoteID()
				m.app.UndoDelete()
				m.app.UpdateCurrentList(m.NoteSelector)
				m.updateTable(m.NoteSelector)

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
			}

		case "ctrl+d":
			switch m.focus {
			case FocusTable:
				oldCursor := m.table.Cursor()
				m.app.DeleteCurrentNote(uint(oldCursor))
				m.updateTable(m.NoteSelector)

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
				m.updateFullTopicTable()
				m.updateStatusBar()
			case FocusTopics:
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
		case "s":
			if m.focus == FocusTable {
				m.NoteSelector = context.Search
				m.app.UpdateCurrentList(m.NoteSelector)
				m.focus = FocusSearch
				m.searchInput.Focus()
				m.table.Blur()
				return m, nil
			}
		case "R":
			if m.focus == FocusTable {
				m.NoteSelector = context.Recent
				m.app.UpdateCurrentList(m.NoteSelector)
				m.updateTable(context.Recent)
				if len(m.app.GetCurrentNotes()) > 0 {
					m.table.SetCursor(0)
					m.app.SelectCurrentNote(0)
					m.textarea.SetValue(m.app.CurrentNoteContent())
					m.updateTopicsTable()
				}
				m.updateStatusBar()
			}
		case "A":
			if m.focus == FocusTable {
				m.NoteSelector = context.Default
				m.app.UpdateCurrentList(m.NoteSelector)
				m.updateTable(context.Default)
				if len(m.app.GetCurrentNotes()) > 0 {
					m.table.SetCursor(0)
					m.app.SelectCurrentNote(0)
					m.textarea.SetValue(m.app.CurrentNoteContent())
					m.updateTopicsTable()
				}
				m.updateStatusBar()
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
