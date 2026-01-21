package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/haochend413/bubbles/key"
	"github.com/haochend413/bubbles/table"
	"github.com/haochend413/ntkpr/sys"
)

// Global keys that work in any mode
type globalKeyMap struct {
	QuitApp           key.Binding
	SwitchFocusWindow key.Binding
	SyncWithDB        key.Binding
	GetHelp           key.Binding
}

var globalKeys = globalKeyMap{
	QuitApp:           key.NewBinding(key.WithKeys("ctrl+c")),
	SwitchFocusWindow: key.NewBinding(key.WithKeys("tab")),
	SyncWithDB:        key.NewBinding(key.WithKeys("ctrl+q")),
	GetHelp:           key.NewBinding(key.WithKeys("H")),
}

// Table focus keys (for threads, branches, notes tables)
type tableKeyMap struct {
	Select        key.Binding // Enter to drill down / select
	Back          key.Binding // Escape to go back up
	CreateNew     key.Binding // Create new item in current table
	Delete        key.Binding // Delete current item
	Highlight     key.Binding // Toggle highlight
	Privatize     key.Binding // Toggle private
	GoToEdit      key.Binding // Go directly to edit mode
	ViewChangelog key.Binding // View changelog
	UpTable       key.Binding // Move to table above (non-circular)
	DownTable     key.Binding // Move to table below (non-circular)
}

var tableKeys = tableKeyMap{
	Select:        key.NewBinding(key.WithKeys("enter")),
	Back:          key.NewBinding(key.WithKeys("esc")),
	CreateNew:     key.NewBinding(key.WithKeys("ctrl+n", "n")),
	Delete:        key.NewBinding(key.WithKeys("ctrl+d")),
	Highlight:     key.NewBinding(key.WithKeys("ctrl+h")),
	Privatize:     key.NewBinding(key.WithKeys("ctrl+p")),
	GoToEdit:      key.NewBinding(key.WithKeys("e", "ctrl+e")),
	ViewChangelog: key.NewBinding(key.WithKeys("ctrl+l")),
	UpTable:       key.NewBinding(key.WithKeys("q")),
	DownTable:     key.NewBinding(key.WithKeys("w")),
}

// Edit focus keys
type editKeyMap struct {
	SaveAndReturn key.Binding
	Cancel        key.Binding
}

var editKeys = editKeyMap{
	SaveAndReturn: key.NewBinding(key.WithKeys("ctrl+s")),
	Cancel:        key.NewBinding(key.WithKeys("ctrl+x")),
}

// Update handles UI events and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.statusBar.SetWidth(m.width)

		borderOverhead := 8
		availableWidth := m.width - borderOverhead

		// Split: 40% for tables (left), 60% for edit (right)
		tableContentWidth := int(float64(availableWidth) * 0.4)
		editContentWidth := availableWidth - tableContentWidth

		tableWidth := tableContentWidth
		editWidth := editContentWidth

		// Distribute column widths for tables
		idWidth := max(4, int(float64(tableWidth)*0.08))
		timeWidth := max(8, int(float64(tableWidth)*0.22))
		flagWidth := max(4, int(float64(tableWidth)*0.15))
		CountWidth := max(4, int(float64(tableWidth)*0.07))
		nameWidth := max(10, int(float64(tableWidth)*0.51))
		contentWidth := max(10, int(float64(tableWidth)*0.58))

		// Separate column definitions for threads, branches (Name), and notes (Content)
		threadColumns := []table.Column{
			{Title: "ID", Width: idWidth},
			{Title: "Time", Width: timeWidth},
			{Title: "Name", Width: nameWidth},
			{Title: "#Bs", Width: CountWidth},
			{Title: "Flags", Width: flagWidth},
		}

		branchColumns := []table.Column{
			{Title: "ID", Width: idWidth},
			{Title: "Time", Width: timeWidth},
			{Title: "Name", Width: nameWidth},
			{Title: "#Ns", Width: CountWidth},
			{Title: "Flags", Width: flagWidth},
		}

		noteColumns := []table.Column{
			{Title: "ID", Width: idWidth},
			{Title: "Time", Width: timeWidth},
			{Title: "Content", Width: contentWidth},
			{Title: "Flags", Width: flagWidth},
		}

		// Set columns and width for each table with appropriate column types
		m.threadsTable.SetColumns(threadColumns)
		m.threadsTable.SetWidth(tableWidth)
		m.branchesTable.SetColumns(branchColumns)
		m.branchesTable.SetWidth(tableWidth)
		m.notesTable.SetColumns(noteColumns)
		m.notesTable.SetWidth(tableWidth)

		// Height calculations
		mainContentHeight := m.height - 5 // Reserve for help + status bar

		// Each table gets 1/3 of the left side height
		tableHeight := max(3, (mainContentHeight-3)/10) // -6 for borders/margins
		standard_thread_height := tableHeight + 2 - 3
		standard_branch_height := tableHeight + 2 - 3
		standard_notes_height := tableHeight*8 + 4 - 3
		m.threadsTable.SetHeight(standard_thread_height + 6)
		m.branchesTable.SetHeight(standard_branch_height)
		m.notesTable.SetHeight(standard_notes_height)

		// Textarea takes most of right side
		m.textArea.SetWidth(editWidth)
		textareaHeight := max(5, int(float64(mainContentHeight)*0.7)) - 1
		m.textArea.SetHeight(textareaHeight)

		// Changelog table (below textarea)
		changeColumns := []table.Column{
			{Title: "Type", Width: max(6, int(float64(editWidth)*0.15))},
			{Title: "ID", Width: max(4, int(float64(editWidth)*0.10))},
			{Title: "Time", Width: max(12, int(float64(editWidth)*0.25))},
			{Title: "Description", Width: max(15, int(float64(editWidth)*0.40))},
		}
		m.changeTable.SetColumns(changeColumns)
		m.changeTable.SetWidth(editWidth)
		changeTableHeight := max(5, int(float64(mainContentHeight)*0.3)) - 3
		m.changeTable.SetHeight(changeTableHeight)

	case tea.KeyMsg:
		// Handle global keys first
		switch {
		case key.Matches(msg, globalKeys.QuitApp):
			m.app.SyncWithDatabase()
			return m, tea.Quit

		case key.Matches(msg, globalKeys.SyncWithDB):
			m.app.SyncWithDatabase()
			m.updateThreadsTable()
			m.updateBranchesTable()
			m.updateNotesTable()
			m.updateChangelogTable()
			m.updateStatusBar()
			return m, nil

		case key.Matches(msg, globalKeys.SwitchFocusWindow):
			// Tab cycles through three tables only: Threads -> Branches -> Notes -> Threads
			// Edit and Changelog can only be accessed via specific keys (e/ctrl+e and ctrl+l)
			if m.focus == FocusEdit {
				m.ExitEdit(true)
				return m, nil
			}
			if m.focus == FocusChangelog {
				// If in changelog, tab does nothing
				return m, nil
			}
			// Cycle focus: Threads -> Branches -> Notes -> Threads
			switch m.focus {
			case FocusThreads:
				m.SetFocus(FocusBranches)
			case FocusBranches:
				m.SetFocus(FocusNotes)
			case FocusNotes:
				m.SetFocus(FocusThreads)
			}
			return m, nil
		}

		// Handle mode-specific keys
		switch m.focus {
		case FocusThreads:
			switch {
			case key.Matches(msg, tableKeys.Select):
				cursor := m.threadsTable.Cursor()
				m.app.GetDataMgr().SwitchActiveThread(cursor)
				m.updateBranchesTable()
				m.branchesTable.SetCursor(0)
				m.updateNotesTable()
				m.notesTable.SetCursor(0)
				m.SetFocus(FocusBranches)
				return m, nil

			case key.Matches(msg, tableKeys.CreateNew):
				m.app.CreateNewThread()
				m.updateThreadsTable()
				lastIdx := len(m.app.GetThreadList()) - 1
				if lastIdx >= 0 {
					m.threadsTable.SetCursor(lastIdx)
				}
				cursor := m.threadsTable.Cursor()
				m.app.GetDataMgr().SwitchActiveThread(cursor)
				m.updateBranchesTable()
				m.branchesTable.SetCursor(0)
				m.updateNotesTable()
				m.notesTable.SetCursor(0)
				m.SetFocus(FocusThreads)
				return m, nil

			case key.Matches(msg, tableKeys.Delete):
				m.app.DeleteCurrentThread()
				m.updateThreadsTable()
				m.updateBranchesTable()
				m.updateNotesTable()
				threadRows := m.threadsTable.Rows()
				branchRows := m.branchesTable.Rows()
				noteRows := m.notesTable.Rows()
				prevThreadCursor := m.threadsTable.Cursor()
				prevBranchCursor := m.branchesTable.Cursor()
				prevNoteCursor := m.notesTable.Cursor()
				if len(threadRows) > 0 {
					if prevThreadCursor < len(threadRows) {
						m.threadsTable.SetCursor(prevThreadCursor)
					} else {
						m.threadsTable.SetCursor(len(threadRows) - 1)
					}
				}
				if len(branchRows) > 0 {
					if prevBranchCursor < len(branchRows) {
						m.branchesTable.SetCursor(prevBranchCursor)
					} else {
						m.branchesTable.SetCursor(len(branchRows) - 1)
					}
				}
				if len(noteRows) > 0 {
					if prevNoteCursor < len(noteRows) {
						m.notesTable.SetCursor(prevNoteCursor)
					} else {
						m.notesTable.SetCursor(len(noteRows) - 1)
					}
				}
				m.SetFocus(FocusThreads)
				return m, nil

			case key.Matches(msg, tableKeys.GoToEdit):
				m.EnterEdit(FocusThreads)
				return m, nil

			case key.Matches(msg, tableKeys.DownTable):
				m.SetFocus(FocusBranches)
				return m, nil
			}

		case FocusBranches:
			switch {
			case key.Matches(msg, tableKeys.Select):
				cursor := m.branchesTable.Cursor()
				m.app.GetDataMgr().SwitchActiveBranch(cursor)
				m.updateNotesTable()
				m.notesTable.SetCursor(0)
				m.SetFocus(FocusNotes)
				return m, nil

			case key.Matches(msg, tableKeys.Back):
				m.SetFocus(FocusThreads)
				return m, nil

			case key.Matches(msg, tableKeys.CreateNew):
				m.app.CreateNewBranch()
				m.updateBranchesTable()
				lastIdx := len(m.app.GetActiveBranchList()) - 1
				if lastIdx >= 0 {
					m.branchesTable.SetCursor(lastIdx)
				}
				cursor := m.branchesTable.Cursor()
				m.app.GetDataMgr().SwitchActiveBranch(cursor)
				m.updateNotesTable()
				m.notesTable.SetCursor(0)
				m.SetFocus(FocusBranches)
				return m, nil

			case key.Matches(msg, tableKeys.Delete):
				m.app.DeleteCurrentBranch()
				m.updateBranchesTable()
				m.updateNotesTable()
				threadRows := m.threadsTable.Rows()
				branchRows := m.branchesTable.Rows()
				noteRows := m.notesTable.Rows()
				prevThreadCursor := m.threadsTable.Cursor()
				prevBranchCursor := m.branchesTable.Cursor()
				prevNoteCursor := m.notesTable.Cursor()
				if len(threadRows) > 0 {
					if prevThreadCursor < len(threadRows) {
						m.threadsTable.SetCursor(prevThreadCursor)
					} else {
						m.threadsTable.SetCursor(len(threadRows) - 1)
					}
				}
				if len(branchRows) > 0 {
					if prevBranchCursor < len(branchRows) {
						m.branchesTable.SetCursor(prevBranchCursor)
					} else {
						m.branchesTable.SetCursor(len(branchRows) - 1)
					}
				}
				if len(noteRows) > 0 {
					if prevNoteCursor < len(noteRows) {
						m.notesTable.SetCursor(prevNoteCursor)
					} else {
						m.notesTable.SetCursor(len(noteRows) - 1)
					}
				}
				m.SetFocus(FocusBranches)
				return m, nil

			case key.Matches(msg, tableKeys.GoToEdit):
				m.EnterEdit(FocusBranches)
				return m, nil

			case key.Matches(msg, tableKeys.UpTable):
				m.SetFocus(FocusThreads)
				return m, nil

			case key.Matches(msg, tableKeys.DownTable):
				m.SetFocus(FocusNotes)
				return m, nil
			}

		case FocusNotes:
			switch {
			case key.Matches(msg, tableKeys.Select):
				cursor := m.notesTable.Cursor()
				m.app.GetDataMgr().SwitchActiveNote(cursor)
				m.EnterEdit(FocusNotes)
				m.updateStatusBar()
				return m, nil

			case key.Matches(msg, tableKeys.Back):
				m.SetFocus(FocusBranches)
				return m, nil

			case key.Matches(msg, tableKeys.CreateNew):
				m.app.CreateNewNote()
				m.updateNotesTable()
				lastIdx := len(m.app.GetActiveNoteList()) - 1
				if lastIdx >= 0 {
					m.notesTable.SetCursor(lastIdx)
				}
				cursor := m.notesTable.Cursor()
				m.app.GetDataMgr().SwitchActiveNote(cursor)
				m.SetFocus(FocusNotes)
				m.updateNotesTable()
				return m, nil

			case key.Matches(msg, tableKeys.Delete):
				m.app.DeleteCurrentNote()
				m.updateNotesTable()
				threadRows := m.threadsTable.Rows()
				branchRows := m.branchesTable.Rows()
				noteRows := m.notesTable.Rows()
				prevThreadCursor := m.threadsTable.Cursor()
				prevBranchCursor := m.branchesTable.Cursor()
				prevNoteCursor := m.notesTable.Cursor()
				if len(threadRows) > 0 {
					if prevThreadCursor < len(threadRows) {
						m.threadsTable.SetCursor(prevThreadCursor)
					} else {
						m.threadsTable.SetCursor(len(threadRows) - 1)
					}
				}
				if len(branchRows) > 0 {
					if prevBranchCursor < len(branchRows) {
						m.branchesTable.SetCursor(prevBranchCursor)
					} else {
						m.branchesTable.SetCursor(len(branchRows) - 1)
					}
				}
				if len(noteRows) > 0 {
					if prevNoteCursor < len(noteRows) {
						m.notesTable.SetCursor(prevNoteCursor)
					} else {
						m.notesTable.SetCursor(len(noteRows) - 1)
					}
				}
				m.SetFocus(FocusNotes)
				return m, nil

			case key.Matches(msg, tableKeys.Highlight):
				m.app.ToggleCurrentNoteHighlight()
				m.updateNotesTable()
				return m, nil

			case key.Matches(msg, tableKeys.Privatize):
				m.app.ToggleCurrentNotePrivate()
				m.updateNotesTable()
				return m, nil

			case key.Matches(msg, tableKeys.GoToEdit):
				cursor := m.notesTable.Cursor()
				m.app.GetDataMgr().SwitchActiveNote(cursor)
				m.EnterEdit(FocusNotes)
				return m, nil

			case key.Matches(msg, tableKeys.ViewChangelog):
				m.updateChangelogTable()
				m.SetFocus(FocusChangelog)
				return m, nil

			case key.Matches(msg, tableKeys.UpTable):
				m.SetFocus(FocusBranches)
				return m, nil
			}

		case FocusChangelog:
			switch {
			case key.Matches(msg, tableKeys.Back):
				m.SetFocus(FocusNotes)
				return m, nil
			}

		case FocusEdit:
			switch {
			case key.Matches(msg, editKeys.SaveAndReturn):
				m.ExitEdit(true)
				return m, nil

			case key.Matches(msg, editKeys.Cancel):
				m.ExitEdit(false)
				return m, nil
			}
		}

	case table.MoveSelectMsg:
		// Handle cursor movement in tables
		switch m.focus {
		case FocusThreads:
			cursor := m.threadsTable.Cursor()
			m.app.GetDataMgr().SwitchActiveThread(cursor)
			m.updateBranchesTable()
			// Reset branch cursor to 0 when thread changes
			m.branchesTable.SetCursor(0)
			m.updateNotesTable()
			// Reset note cursor to 0 when branch changes
			m.notesTable.SetCursor(0)
			m.textArea.SetValue(m.app.GetCurrentThreadSummary())
			m.updateStatusBar()
		case FocusBranches:
			cursor := m.branchesTable.Cursor()
			m.app.GetDataMgr().SwitchActiveBranch(cursor)
			m.updateNotesTable()
			// Reset note cursor to 0 when branch changes
			m.notesTable.SetCursor(0)
			m.textArea.SetValue(m.app.GetCurrentBranchSummary())
			m.updateStatusBar()
		case FocusNotes:
			cursor := m.notesTable.Cursor()
			m.app.GetDataMgr().SwitchActiveNote(cursor)
			m.textArea.SetValue(m.app.GetCurrentNoteContent())
			m.updateStatusBar()
		}
	}

	// Update the focused component
	switch m.focus {
	case FocusThreads:
		m.threadsTable, cmd = m.threadsTable.Update(msg)
		cmds = append(cmds, cmd)
	case FocusBranches:
		m.branchesTable, cmd = m.branchesTable.Update(msg)
		cmds = append(cmds, cmd)
	case FocusNotes:
		m.notesTable, cmd = m.notesTable.Update(msg)
		cmds = append(cmds, cmd)
	case FocusEdit:
		m.textArea, cmd = m.textArea.Update(msg)
		cmds = append(cmds, cmd)
	case FocusChangelog:
		m.changeTable, cmd = m.changeTable.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// Helper functions

// SetFocus centralizes focus switching, previousFocus, blur/focus, and textArea population
func (m *Model) SetFocus(focus FocusState) {
	if m.focus == FocusEdit {
		// Should use ExitEdit instead
		return
	}
	// This should add animation for size shifting.
	m.blurAllTables()
	m.focus = focus
	tableHeight := max(3, (m.height-5-3)/10)
	standard_thread_height := tableHeight + 2 - 3
	standard_branch_height := tableHeight + 2 - 3
	standard_notes_height := tableHeight*8 + 4 - 3

	switch focus {
	case FocusThreads:
		m.threadsTable.Focus()
		m.textArea.SetValue(m.app.GetCurrentThreadSummary())
		m.threadsTable.SetHeight(standard_thread_height + 6)
		m.branchesTable.SetHeight(standard_branch_height)
		m.notesTable.SetHeight(standard_notes_height)
	case FocusBranches:
		m.branchesTable.Focus()
		m.textArea.SetValue(m.app.GetCurrentBranchSummary())
		m.threadsTable.SetHeight(standard_thread_height)
		m.branchesTable.SetHeight(standard_branch_height + 6)
		m.notesTable.SetHeight(standard_notes_height)
	case FocusNotes:
		m.notesTable.Focus()
		m.textArea.SetValue(m.app.GetCurrentNoteContent())
		m.threadsTable.SetHeight(standard_thread_height)
		m.branchesTable.SetHeight(standard_branch_height)
		m.notesTable.SetHeight(standard_notes_height + 6)
	case FocusChangelog:
		m.changeTable.Focus()
	}
	m.updateStatusBar()
}

// EnterEdit switches to edit mode from a given focus, sets previousFocus, populates textarea, and focuses textarea
func (m *Model) EnterEdit(from FocusState) {
	//after enter edit, we should always switch to the previous stored input method.
	// fmt.Printf(m.editPrevInputMethodID)
	// id, _ := sys.InputMethodID(m.editPrevIMEType)
	// sys.SwitchInputMethod(id) // bring back to previous method
	m.previousFocus = from
	switch from {
	case FocusThreads:
		m.textArea.SetValue(m.app.GetCurrentThreadSummary())
	case FocusBranches:
		m.textArea.SetValue(m.app.GetCurrentBranchSummary())
	case FocusNotes:
		m.textArea.SetValue(m.app.GetCurrentNoteContent())
	}
	m.focus = FocusEdit
	m.blurAllTables()
	m.textArea.Focus()
}

// ExitEdit leaves edit mode, optionally saving, and returns to previous focus
func (m *Model) ExitEdit(save bool) {
	// after we exit, we should always switch to english input method to prevent keypress blocking by chinese input method.
	// t, _ := sys.GetCurrentInputMethod()

	sys.SwitchInputMethod(sys.ENGLISH_INPUT_METHOD_ID)

	if save {
		switch m.previousFocus {
		case FocusThreads:
			m.app.SetCurrentThreadSummary(m.textArea.Value())
			m.updateThreadsTable()
		case FocusBranches:
			m.app.SetCurrentBranchSummary(m.textArea.Value())
			m.updateBranchesTable()
		case FocusNotes:
			m.app.SetCurrentNoteContent(m.textArea.Value())
			m.updateNotesTable()
		}
	}
	m.focus = m.previousFocus
	m.textArea.Blur()
	m.focusCurrentTable()
	m.updateStatusBar()
}

func (m *Model) blurAllTables() {
	m.threadsTable.Blur()
	m.branchesTable.Blur()
	m.notesTable.Blur()
}

func (m *Model) focusCurrentTable() {
	m.blurAllTables()
	switch m.focus {
	case FocusThreads:
		m.threadsTable.Focus()
	case FocusBranches:
		m.branchesTable.Focus()
	case FocusNotes:
		m.notesTable.Focus()
	case FocusChangelog:
		m.changeTable.Focus()
	}
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
