package ui

import (
	"github.com/haochend413/ntkpr/internal/app/context"
	"github.com/haochend413/ntkpr/state"
)

// Distribute state in json on startup
func (m *Model) DistributeState(s *state.State) {
	// Initialize YOffsets if nil (for backwards compatibility with old state files)
	if s.YOffsets == nil {
		s.YOffsets = map[context.ContextPtr](int){
			context.Default: 0,
			context.Recent:  0,
			context.Search:  0,
		}
	}
	// Restore YOffsets to the model's yOffsets map
	m.yOffsets = map[context.ContextPtr]int{
		context.Default: s.YOffsets[context.Default],
		context.Recent:  s.YOffsets[context.Recent],
		context.Search:  s.YOffsets[context.Search],
	}
	// First restore saved cursors
	m.app.SetCursors(s.Cursors)
	// Then switch to the saved context (this will use the restored cursor)
	m.app.UpdateCurrentList(s.LastContext, s.Cursors[s.LastContext])
	// Now render the table
	m.updateTable(s.LastContext)
	// Set the table cursor and viewport offset together
	cursor := int(s.Cursors[s.LastContext])
	if cursor >= len(m.app.GetCurrentNotes()) && len(m.app.GetCurrentNotes()) > 0 {
		cursor = len(m.app.GetCurrentNotes()) - 1
	}
	yOffset := s.YOffsets[s.LastContext]
	m.table.SetCursorAndOffset(cursor, yOffset)
	// Select the note at cursor
	if len(m.app.GetCurrentNotes()) > 0 {
		m.app.SelectCurrentNote(cursor)
		m.textarea.SetValue(m.app.CurrentNoteContent())
	}
}

// Collect end state on termination
func (m Model) CollectState() *state.State {
	s := &state.State{}
	s.LastContext = m.CurrentContext
	// Save current cursor position to the current context before collecting
	m.app.UpdateCurrentList(m.CurrentContext, uint(m.table.Cursor()))
	s.Cursors = m.app.GetCursors()
	// Save YOffsets for all contexts, with current context's offset updated
	s.YOffsets = map[context.ContextPtr](int){
		context.Default: m.yOffsets[context.Default],
		context.Recent:  m.yOffsets[context.Recent],
		context.Search:  m.yOffsets[context.Search],
	}
	// Update the current context's YOffset with the actual table value
	s.YOffsets[m.CurrentContext] = m.table.YOffset()
	return s
}

// switchToContext saves the current context's viewport state and switches to a new context
// This preserves both cursor position and viewport YOffset between context switches
func (m *Model) switchToContext(newContext context.ContextPtr) {
	// Save current context's YOffset before switching
	m.yOffsets[m.CurrentContext] = m.table.YOffset()

	// Switch context and get the saved cursor for the new context
	new_cursor := m.app.UpdateCurrentList(newContext, uint(m.table.Cursor()))

	// Render the table for the new context
	m.updateTable(newContext)

	// Restore cursor and YOffset for the new context
	if len(m.app.GetCurrentNotes()) > 0 {
		if int(new_cursor) >= len(m.app.GetCurrentNotes()) {
			new_cursor = uint(len(m.app.GetCurrentNotes()) - 1)
		}
		yOffset := m.yOffsets[newContext]
		m.table.SetCursorAndOffset(int(new_cursor), yOffset)
		m.app.SelectCurrentNote(int(new_cursor))
		m.textarea.SetValue(m.app.CurrentNoteContent())
		m.updateTopicsTable()
	}
	m.updateStatusBar()
}
