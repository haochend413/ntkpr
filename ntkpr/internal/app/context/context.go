package context

import (
	"sort"
	"strings"

	"github.com/haochend413/ntkpr/internal/models"
)

type ContextPtr int

const (
	None    ContextPtr = -1
	Default ContextPtr = 0
	Recent  ContextPtr = 1
	Search  ContextPtr = 2
)

type ContextOrder int

const (
	CreateAt ContextOrder = 0 // default , time order
	UpdateAt ContextOrder = 1 // recent, most recently updated
)

type Context struct {
	Name   ContextPtr
	Notes  []*models.Note
	Order  ContextOrder
	Cursor uint
}

type ContextMgr struct {
	previousContext ContextPtr
	currentContext  ContextPtr
	Contexts        []*Context
}

// NewContextMgr creates a new context manager with all contexts initialized
func NewContextMgr() *ContextMgr {
	return &ContextMgr{
		previousContext: None,
		currentContext:  Default,
		Contexts: []*Context{
			{Name: Default, Notes: make([]*models.Note, 0), Cursor: 0, Order: CreateAt}, // Default context
			{Name: Recent, Notes: make([]*models.Note, 0), Cursor: 0, Order: UpdateAt},  // Recent context
			{Name: Search, Notes: make([]*models.Note, 0), Cursor: 0, Order: CreateAt},  // Search context
		},
	}
}

func (cm *ContextMgr) SwitchContext(c ContextPtr) {
	// make sure they are different
	if c != cm.currentContext {
		cm.previousContext = cm.currentContext
		cm.currentContext = c
	}
}

func (cm *ContextMgr) GetCurrentNotes() []*models.Note {
	return cm.Contexts[cm.currentContext].Notes
}

func (cm *ContextMgr) GetCurrentContext() ContextPtr {
	return cm.currentContext
}

func (cm *ContextMgr) GetPreviousContext() ContextPtr {
	return cm.previousContext
}

func (cm *ContextMgr) RefreshDefaultContext(notes []*models.Note) {
	cm.Contexts[Default].Notes = notes
}

func (cm *ContextMgr) RefreshRecentContext() {
	// we should fetch the newest ~ 20 notes from default context
	notes := cm.Contexts[Default].Notes
	// Create a copy to avoid modifying the original slice
	notesCopy := make([]*models.Note, len(notes))
	copy(notesCopy, notes)
	//sort by UpdatedAt (most recent first)
	sort.Slice(notesCopy, func(i, j int) bool {
		return notesCopy[i].UpdatedAt.After(notesCopy[j].UpdatedAt)
	})
	// fetch top
	recentCount := 20
	if len(notesCopy) < recentCount {
		recentCount = len(notesCopy)
	}
	cm.Contexts[Recent].Notes = notesCopy[:recentCount]
}

func (cm *ContextMgr) RefreshSearchContext(q string) {
	// first, get the current note list
	// I am not sure whether this is correct.
	c := cm.currentContext
	if cm.currentContext == Search {
		c = cm.previousContext
	}
	notes := cm.Contexts[c].Notes

	//loop through and search
	if q == "" {
		cm.Contexts[Search].Notes = notes
		return
	}
	query := strings.ToLower(q)
	filteredNotes := make([]*models.Note, 0)
	for _, note := range notes {
		if strings.Contains(strings.ToLower(note.Content), query) {
			filteredNotes = append(filteredNotes, note)
			continue
		}
		for _, topic := range note.Topics {
			if strings.Contains(strings.ToLower(topic.Topic), query) {
				filteredNotes = append(filteredNotes, note)
				break
			}
		}
	}
	// Sort by CreatedAt to maintain chronological order
	sort.Slice(filteredNotes, func(i, j int) bool {
		return filteredNotes[i].CreatedAt.Before(filteredNotes[j].CreatedAt)
	})
	cm.Contexts[Search].Notes = filteredNotes
}

// GetCurrentCursor returns the cursor position in the current context
func (cm *ContextMgr) GetCurrentCursor() uint {
	return cm.Contexts[cm.currentContext].Cursor
}

// SetCurrentCursor sets the cursor position in the current context
func (cm *ContextMgr) SetCurrentCursor(cursor uint) {
	cm.Contexts[cm.currentContext].Cursor = cursor
}

// GetCurrentNote returns the note at the current cursor position, or nil if invalid
func (cm *ContextMgr) GetCurrentNote() *models.Note {
	notes := cm.GetCurrentNotes()
	cursor := cm.GetCurrentCursor()
	if len(notes) == 0 || int(cursor) >= len(notes) {
		return nil
	}
	return notes[cursor]
}

// GetNoteCount returns the number of notes in the current context
func (cm *ContextMgr) GetNoteCount() int {
	return len(cm.GetCurrentNotes())
}

// AddNoteToDefault adds a note to the default context in chronological order
func (cm *ContextMgr) AddNoteToDefault(note *models.Note) {
	notes := cm.Contexts[Default].Notes
	// Find the correct position to insert the note to maintain chronological order
	insertPos := len(notes)
	for i, n := range notes {
		if n.CreatedAt.After(note.CreatedAt) {
			insertPos = i
			break
		}
	}
	// Insert at the correct position
	if insertPos == len(notes) {
		cm.Contexts[Default].Notes = append(notes, note)
	} else {
		cm.Contexts[Default].Notes = append(notes[:insertPos], append([]*models.Note{note}, notes[insertPos:]...)...)
	}
}

// RemoveNoteFromDefault removes a note from default context by ID
func (cm *ContextMgr) RemoveNoteFromDefault(noteID uint) {
	notes := cm.Contexts[Default].Notes
	for i, note := range notes {
		if note.ID == noteID {
			cm.Contexts[Default].Notes = append(notes[:i], notes[i+1:]...)
			break
		}
	}
}

// SortCurrentContext sorts the current context notes by CreatedAt (chronological order)
func (cm *ContextMgr) SortCurrentContext() {
	notes := cm.Contexts[cm.currentContext].Notes
	order := cm.Contexts[cm.currentContext].Order

	// well this is in place ! careful!
	switch order {
	case CreateAt:
		sort.Slice(notes, func(i, j int) bool {
			return notes[i].CreatedAt.Before(notes[j].CreatedAt)
		})
	case UpdateAt:
		sort.Slice(notes, func(i, j int) bool {
			return notes[i].UpdatedAt.After(notes[j].UpdatedAt)
		})
	default:
		sort.Slice(notes, func(i, j int) bool {
			return notes[i].CreatedAt.Before(notes[j].CreatedAt)
		})
	}
}

// only for data storage purposes. Do not use in coding.
func (cm *ContextMgr) GetCursors() map[ContextPtr]uint {
	return map[ContextPtr]uint{
		Default: cm.Contexts[Default].Cursor,
		Recent:  cm.Contexts[Recent].Cursor,
		Search:  cm.Contexts[Search].Cursor,
	}
}

func (cm *ContextMgr) SetCursors(m map[ContextPtr]uint) {
	cm.Contexts[Default].Cursor = m[Default]
	cm.Contexts[Recent].Cursor = m[Recent]
	cm.Contexts[Search].Cursor = m[Search]
}
