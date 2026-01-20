package context

// This is not a good design. We have to manually get too many things.
// We need to have it expose what we want it , or what the app wants it to expose.
// Context should not be too much. Just a wrapper for sorting and cursor.
// Also, we can use fuzzy lib for search. Yes it is a good idea.
import (
	"sort"
	"strings"

	"github.com/haochend413/ntkpr/internal/models"
)

type NoteContext struct {
	Name   ContextPtr
	Notes  []*models.Note
	Order  ContextOrder
	Cursor uint
}

type NoteContextMgr struct {
	previousContext ContextPtr
	currentContext  ContextPtr
	Contexts        []*NoteContext
}

// NewContextMgr creates a new context manager with all contexts initialized
func NewNoteContextMgr() *NoteContextMgr {
	return &NoteContextMgr{
		previousContext: None,
		currentContext:  Default,
		Contexts: []*NoteContext{
			{Name: Default, Notes: make([]*models.Note, 0), Cursor: 0, Order: CreateAt}, // Default context
			{Name: Recent, Notes: make([]*models.Note, 0), Cursor: 0, Order: UpdateAt},  // Recent context
			{Name: Search, Notes: make([]*models.Note, 0), Cursor: 0, Order: CreateAt},  // Search context
		},
	}
}

// SwitchContext switches the contextMgr into a new context and update previous context.
func (cm *NoteContextMgr) SwitchContext(c ContextPtr) {
	// make sure they are different
	if c != cm.currentContext {
		cm.previousContext = cm.currentContext
		cm.currentContext = c
	}
}

func (cm *NoteContextMgr) GetCurrentNotes() []*models.Note {
	return cm.Contexts[cm.currentContext].Notes
}

func (cm *NoteContextMgr) GetCurrentContext() ContextPtr {
	return cm.currentContext
}

func (cm *NoteContextMgr) GetPreviousContext() ContextPtr {
	return cm.previousContext
}

// There should be other functions that deals with switching Contexts.

func (cm *NoteContextMgr) RefreshDefaultContext(notes []*models.Note) {
	cm.Contexts[Default].Notes = notes
}

// RefreshRecentContext refreshes recent context based on default context notes.
func (cm *NoteContextMgr) RefreshRecentContext() {
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

func (cm *NoteContextMgr) RefreshSearchContext(q string) {
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
		}
	}
	// Sort by CreatedAt to maintain chronological order
	sort.Slice(filteredNotes, func(i, j int) bool {
		return filteredNotes[i].CreatedAt.Before(filteredNotes[j].CreatedAt)
	})
	cm.Contexts[Search].Notes = filteredNotes
}

// GetCurrentCursor returns the cursor position in the current context
func (cm *NoteContextMgr) GetCurrentCursor() uint {
	return cm.Contexts[cm.currentContext].Cursor
}

// SetCurrentCursor sets the cursor position in the current context
func (cm *NoteContextMgr) SetCurrentCursor(cursor uint) {
	cm.Contexts[cm.currentContext].Cursor = cursor
}

// GetCurrentNote returns the note at the current cursor position, or nil if invalid
func (cm *NoteContextMgr) GetCurrentNote() *models.Note {
	notes := cm.GetCurrentNotes()
	cursor := cm.GetCurrentCursor()
	if len(notes) == 0 || int(cursor) >= len(notes) {
		return nil
	}
	return notes[cursor]
}

// // GetNoteCount returns the number of notes in the current context
// func (cm *NoteContextMgr) GetNoteCount() int {
// 	return len(cm.GetCurrentNotes())
// }

// AddNoteToDefault adds a note to the default context in chronological order
// This removes the need of refreshing again after delete from data.
// ... Wait, then when we load, we probably should ... ? Yes, this is definitely a good idea ...
// Or, look : when we do this, we should probably sync back to data ? Like, what's the flow of local data ?
// This only adds note to the context, if we switch back, it is still there.
// No this needs to be changed ...
// I think a good way around this is , since we only care about default (recent and search are all done in-place, we directly interact with dataMgr. )
// DataMgr should have the "update" functions : do ops to data, and return them back.
// Then, we just use editStack to sync between database.
func (cm *NoteContextMgr) AddNote(note *models.Note) {
	// add note to DataMgr

}

// RemoveNoteFromDefault removes a note from default context by ID
func (cm *NoteContextMgr) RemoveNoteFromDefault(noteID uint) {
	notes := cm.Contexts[Default].Notes
	for i, note := range notes {
		if note.ID == noteID {
			cm.Contexts[Default].Notes = append(notes[:i], notes[i+1:]...)
			break
		}
	}
}

// SortCurrentContext sorts the current context notes by CreatedAt (chronological order)
func (cm *NoteContextMgr) SortCurrentContext() {
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

// only for data storage and load purposes. Do not use in coding.
func (cm *NoteContextMgr) GetCursors() map[ContextPtr]uint {
	return map[ContextPtr]uint{
		Default: cm.Contexts[Default].Cursor,
		Recent:  cm.Contexts[Recent].Cursor,
		Search:  cm.Contexts[Search].Cursor,
	}
}

func (cm *NoteContextMgr) SetCursors(m map[ContextPtr]uint) {
	cm.Contexts[Default].Cursor = m[Default]
	cm.Contexts[Recent].Cursor = m[Recent]
	cm.Contexts[Search].Cursor = m[Search]
}

// UpdateContext switches context, saves current cursor, sorts, and returns new cursor
func (cm *NoteContextMgr) UpdateContext(newContext ContextPtr, currentCursor uint) uint {
	c0 := cm.currentContext
	cm.Contexts[c0].Cursor = currentCursor
	cm.SwitchContext(newContext)
	cm.SortCurrentContext()
	return cm.Contexts[newContext].Cursor
}

// Search performs search and switches to search context
func (cm *NoteContextMgr) Search(query string) {
	cm.RefreshSearchContext(query)
	cm.SwitchContext(Search)
}

// SelectItem sets the cursor and returns the item at that position
func (cm *NoteContextMgr) SelectItem(cursor int) *models.Note {
	notes := cm.GetCurrentNotes()
	if len(notes) == 0 || cursor >= len(notes) || cursor < 0 {
		return nil
	}
	cm.SetCurrentCursor(uint(cursor))
	return notes[cursor]
}
