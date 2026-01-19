package ui

import (
	"fmt"
	"strconv"
	"time"

	"github.com/haochend413/bubbles/statusbar"
	"github.com/haochend413/bubbles/table"
	"github.com/haochend413/bubbles/textarea_vim"
	"github.com/haochend413/bubbles/textinput"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/haochend413/ntkpr/config"
	"github.com/haochend413/ntkpr/internal/app"
	"github.com/haochend413/ntkpr/internal/models"
	"github.com/haochend413/ntkpr/state"
)

// FocusState represents the current UI focus
type FocusState int

const (
	FocusThreads FocusState = iota
	FocusBranches
	FocusNotes
	FocusEdit
	FocusChangelog
)

// Model represents the Bubble Tea model
type Model struct {
	app           *app.App
	Config        *config.Config
	threadsTable  table.Model
	branchesTable table.Model
	notesTable    table.Model
	textArea      textarea_vim.Model
	changeTable   table.Model
	statusBar     statusbar.Model
	previousFocus FocusState
	focus         FocusState
	width         int
	height        int
	ready         bool
}

// NewModel initializes a new UI model
func NewModel(application *app.App, cfg *config.Config, s *state.State) Model {
	// Use default state if nil
	// ignore this for now
	if s == nil {
		s = state.DefaultState()
	}
	if cfg == nil {
		temp := config.LoadOrCreateConfig()
		cfg = &temp
	}

	noteColumns := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "Time", Width: 16},
		{Title: "Content", Width: 40},
		{Title: "Flags", Width: 6},
	}

	noteTable := table.New(
		table.WithColumns(noteColumns),
		table.WithFocused(true),
		table.WithHeight(15),
	)
	branchColumns := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "Time", Width: 16},
		{Title: "Name", Width: 40},
		{Title: "Flags", Width: 6},
	}

	branchTable := table.New(
		table.WithColumns(branchColumns),
		table.WithFocused(true),
		table.WithHeight(15),
	)
	threadColumns := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "Time", Width: 16},
		{Title: "Name", Width: 40},
		{Title: "Flags", Width: 6},
	}

	threadTable := table.New(
		table.WithColumns(threadColumns),
		table.WithFocused(true),
		table.WithHeight(15),
	)

	//set sb
	//Left: Context ; NoteID ; Last Update ; Version (frequency)
	//Right: Action ; Synced ? ; Time
	// Example usage with method chaining
	sb := statusbar.New(
		statusbar.WithHeight(1),
		statusbar.WithWidth(100),
		statusbar.WithLeftLen(4),
		statusbar.WithRightLen(3),
	)

	// Configure all left elements in sequence
	sb.GetLeft(0).SetValue("Context: Default").SetColors("0", "39").SetWidth(25)

	sb.GetLeft(1).SetValue("NoteID: -").SetColors("0", "45").SetWidth(15)
	sb.GetLeft(2).SetValue("Updated: Never").SetColors("0", "37").SetWidth(20)
	sb.GetLeft(3).SetValue("Version: 1.0").SetColors("0", "33").SetWidth(15)
	//set tags for quick and consistent access
	sb.SetTag(sb.GetLeft(0), "filter")
	sb.SetTag(sb.GetLeft(1), "NoteID")
	sb.SetTag(sb.GetLeft(2), "LastUpdated")
	sb.SetTag(sb.GetLeft(3), "Version")

	// Configure all right elements in sequence
	sb.GetRight(0).SetValue("").SetColors("0", "46").SetWidth(12)
	sb.GetRight(1).SetValue("Synced").SetColors("0", "208").SetWidth(15)
	sb.GetRight(2).SetValue(time.Now().Format("15:04:05")).SetColors("0", "226").SetWidth(10)
	sb.SetTag(sb.GetRight(0), "Action")
	sb.SetTag(sb.GetRight(1), "Synced")
	sb.SetTag(sb.GetRight(2), "Time")

	// You can also chain model methods
	sb.SetWidth(100).SetHeight(1)

	textArea := textarea_vim.New()

	// Set colors
	textArea.FocusedStyle.CursorLine = lipgloss.NewStyle().Background(lipgloss.Color("240")).Foreground(lipgloss.Color("15"))
	textArea.BlurredStyle.CursorLine = lipgloss.NewStyle().Background(lipgloss.Color("240")).Foreground(lipgloss.Color("15"))

	// Cursor styling
	// ta.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	// // Placeholder styling
	// ta.FocusedStyle.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	// ta.BlurredStyle.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	// Prompt styling (the ">" symbol)
	// ta.Prompt = "❯ "
	// ta.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	textArea.Placeholder = "Write stuff here..." // This should change when we switch between threads / branches / lists
	textArea.SetWidth(50)
	textArea.SetHeight(10)

	// This needs further improving.
	changeColumns := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "Time", Width: 16},
		{Title: "Content", Width: 40},
		{Title: "Flags", Width: 6},
	}

	changeTable := table.New(
		table.WithColumns(changeColumns),
		table.WithFocused(true),
		table.WithHeight(15),
	)

	m := Model{
		app:           application,
		Config:        cfg,
		threadsTable:  threadTable,
		branchesTable: branchTable,
		notesTable:    noteTable,
		textArea:      textArea,
		statusBar:     sb,
		changeTable:   changeTable,
		focus:         FocusThreads,
	}

	//set states
	// m.DistributeState(s)
	// m.updateTopicsTable()
	m.updateThreadsTable()
	m.updateBranchesTable()
	m.updateNotesTable()
	m.updateStatusBar()

	return m
}

// Init initializes the Bubble Tea model
func (m Model) Init() tea.Cmd {
	// Blink ?
	cmds := []tea.Cmd{}
	cmds = append(cmds, textinput.Blink)
	return tea.Batch(cmds...)
}

// NOTE: This only renders the table. Context switching must be done separately via app.UpdateCurrentList()
func (m *Model) updateThreadsTable() {
	var threads []*models.Thread

	threads = m.app.GetThreadList()

	rows := make([]table.Row, len(threads))
	for i, thread := range threads {

		name := thread.Name
		if len(name) > 38 {
			name = name[:35] + "..."
		}
		idStr := fmt.Sprintf("%d", thread.ID)
		timeStr := thread.CreatedAt.Format("06-01-02 15:04")
		if thread.ID == 0 { // Pending note
			idStr = "P" // Indicate pending
			timeStr = time.Now().Format("06-01-02 15:04")
		}

		flagStrRaw := ""
		if thread.Highlight {
			flagStrRaw += "H"
		}
		if thread.Private {
			flagStrRaw += "P"
		}

		// This needs further tuning.
		rows[i] = table.Row{
			idStr,
			timeStr,
			name,
			flagStrRaw,
		}
	}
	m.threadsTable.SetRows(rows)
}

// NOTE: This only renders the table. Context switching must be done separately via app.UpdateCurrentList()
func (m *Model) updateBranchesTable() {
	var branches []*models.Branch

	branches = m.app.GetActiveBranchList()

	rows := make([]table.Row, len(branches))
	for i, branch := range branches {

		name := branch.Name
		if len(name) > 38 {
			name = name[:35] + "..."
		}
		idStr := fmt.Sprintf("%d", branch.ID)
		timeStr := branch.CreatedAt.Format("06-01-02 15:04")
		if branch.ID == 0 { // Pending note
			idStr = "P" // Indicate pending
			timeStr = time.Now().Format("06-01-02 15:04")
		}

		flagStrRaw := ""
		if branch.Highlight {
			flagStrRaw += "H"
		}
		if branch.Private {
			flagStrRaw += "P"
		}

		// This needs further tuning.
		rows[i] = table.Row{
			idStr,
			timeStr,
			name,
			flagStrRaw,
		}
	}
	m.branchesTable.SetRows(rows)
}

// NOTE: This only renders the table. Context switching must be done separately via app.UpdateCurrentList()
func (m *Model) updateNotesTable() {
	var selectedNotes []*models.Note
	selectedNotes = m.app.GetActiveNoteList()

	rows := make([]table.Row, len(selectedNotes))
	for i, note := range selectedNotes {

		content := note.Content
		if len(content) > 38 {
			content = content[:35] + "..."
		}
		idStr := fmt.Sprintf("%d", note.ID)
		timeStr := note.CreatedAt.Format("06-01-02 15:04")
		if note.ID == 0 { // Pending note
			idStr = "P" // Indicate pending
			timeStr = time.Now().Format("06-01-02 15:04")
		}

		flagStrRaw := ""
		if note.Highlight {
			flagStrRaw += "H"
		}
		if note.Private {
			flagStrRaw += "P"
		}

		rows[i] = table.Row{
			idStr,
			timeStr,
			content,
			flagStrRaw,
		}
	}
	m.notesTable.SetRows(rows)
}

func (m *Model) updateChangelogTable() {
	editMap := m.app.GetEditMap()
	rows := make([]table.Row, 0, len(editMap))

	for key, edit := range editMap {
		if edit.EditType == -1 { // Skip None edits
			continue
		}

		editTypeName := ""
		switch edit.EditType {
		case 0:
			editTypeName = "Create"
		case 1:
			editTypeName = "Update"
		case 2:
			editTypeName = "Delete"
		case 3:
			editTypeName = "Create"
		case 4:
			editTypeName = "Update"
		case 6:
			editTypeName = "Delete"
		case 7:
			editTypeName = "Create"
		case 8:
			editTypeName = "Update"
		case 10:
			editTypeName = "Delete"
		}

		entityType := key.EntityType
		idStr := fmt.Sprintf("%d", key.ID)
		timeStr := time.Now().Format("06-01-02 15:04")
		description := fmt.Sprintf("%s %s", editTypeName, entityType)

		rows = append(rows, table.Row{
			entityType,
			idStr,
			timeStr,
			description,
		})
	}

	m.changeTable.SetRows(rows)
}

func (m *Model) printSync(sync bool) string {
	if sync {
		return "Synced"
	} else {
		return "UnSynced"
	}
}

func (m *Model) updateStatusBar() {
	// Show current table focus
	focusName := "Threads"
	switch m.focus {
	case FocusThreads:
		focusName = "Threads"
	case FocusBranches:
		focusName = "Branches"
	case FocusNotes:
		focusName = "Notes"
	case FocusEdit:
		focusName = "Edit"
	case FocusChangelog:
		focusName = "Changelog"
	}

	m.statusBar.GetTag("filter").SetValue(focusName)
	m.statusBar.GetTag("NoteID").SetValue(strconv.Itoa(int(m.app.GetCurrentNoteID())))
	m.statusBar.GetTag("LastUpdated").SetValue(m.app.GetCurrentNoteUpdatedAt().Format("01-02 15:04"))
	m.statusBar.GetTag("Version").SetValue(strconv.Itoa(m.app.GetCurrentNoteFrequency()))
	m.statusBar.GetTag("Synced").SetValue(m.printSync(m.app.Synced))
	m.statusBar.GetTag("Time").SetValue(time.Now().Format("15:04"))
}

// Here we need handling of change Table which requires extra design.
// Let's ignore it for now.
