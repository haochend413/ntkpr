package ui

import "github.com/haochend413/ntkpr/state"

// Distribute state in json on startup
func (m *Model) DistributeState(s *state.State) {
	m.CurrentContext = s.LastContext
	m.updateTable(m.CurrentContext)
	m.table.SetCursor(s.LastCursor)
}

// Collect end state on termination
func (m Model) CollectState() *state.State {
	s := &state.State{}
	s.LastContext = m.CurrentContext
	s.LastCursor = m.table.Cursor()
	return s
}
