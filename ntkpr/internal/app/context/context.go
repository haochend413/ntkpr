package context

// Context is a wrapper that helps ordering and filtering the raw lists of notes, and store lists from the map.
// In the future, maybe we can add a config menu that allows for multiple orders.

import (
	"sort"
	"strings"

	"github.com/haochend413/ntkpr/internal/models"
)

// This should be improved. How ? What is it that I want ?
// How do we define recent activity ? Is "recent note" still meaningful?
// Should this happen globally, across all ? or separate into layers ?

// Think.

// Well, we will have a "changelog" window to demonstrate logs.
// Then how should context work ? search is definitely useful. We probably need recent ? Or should search be in a different regime.

// OK. Enable contexts for all thread, branch and notes. At least for now.

// NONONO. forget about it. This is silly. let's only do context for notes for now, with minimum change.
// Wait. But in that case, we still need a wrapper to render lists from maps. This layer is still required, it's just that we remove the unecessary functionalities.

// This can Actually be done as a general type. However, I am expecting something different for branch / thread / notes, thus let's do it this way.

type ContextPtr int

const (
	None    ContextPtr = -1
	Default ContextPtr = 0
	Recent  ContextPtr = 1
	Search  ContextPtr = 2
)

// This is replicative, but can be useful.
type ContextOrder int

const (
	CreateAt ContextOrder = 0 // default , time order
	UpdateAt ContextOrder = 1 // recent, most recently updated
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

func (cm *NoteContextMgr) RefreshDefaultContext(notes []*models.Note) {
	cm.Contexts[Default].Notes = notes
}

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

// GetNoteCount returns the number of notes in the current context
func (cm *NoteContextMgr) GetNoteCount() int {
	return len(cm.GetCurrentNotes())
}

// AddNoteToDefault adds a note to the default context in chronological order
func (cm *NoteContextMgr) AddNoteToDefault(note *models.Note) {
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

// only for data storage purposes. Do not use in coding.
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

/* Branches*/

type BranchContext struct {
	Name     ContextPtr
	Branches []*models.Branch
	Order    ContextOrder
	Cursor   uint
}

type BranchContextMgr struct {
	previousContext ContextPtr
	currentContext  ContextPtr
	Contexts        []*BranchContext
}

// NewContextMgr creates a new context manager with all contexts initialized
// We might not need recent for branch, and we might put search into a different category. But now, let's keep things real.
func NewBranchContextMgr() *BranchContextMgr {
	return &BranchContextMgr{
		previousContext: None,
		currentContext:  Default,
		Contexts: []*BranchContext{
			{Name: Default, Branches: make([]*models.Branch, 0), Cursor: 0, Order: CreateAt}, // Default context
			{Name: Recent, Branches: make([]*models.Branch, 0), Cursor: 0, Order: UpdateAt},  // Recent context
			{Name: Search, Branches: make([]*models.Branch, 0), Cursor: 0, Order: CreateAt},  // Search context
		},
	}
}

func (cm *BranchContextMgr) SwitchContext(c ContextPtr) {
	// make sure they are different
	if c != cm.currentContext {
		cm.previousContext = cm.currentContext
		cm.currentContext = c
	}
}

func (cm *BranchContextMgr) GetCurrentBranches() []*models.Branch {
	return cm.Contexts[cm.currentContext].Branches
}

func (cm *BranchContextMgr) GetCurrentContext() ContextPtr {
	return cm.currentContext
}

func (cm *BranchContextMgr) GetPreviousContext() ContextPtr {
	return cm.previousContext
}

func (cm *BranchContextMgr) RefreshDefaultContext(bs []*models.Branch) {
	cm.Contexts[Default].Branches = bs
}

func (cm *BranchContextMgr) RefreshRecentContext() {
	bs := cm.Contexts[Default].Branches
	// Create a copy to avoid modifying the original slice
	bscp := make([]*models.Branch, len(bs))
	copy(bscp, bs)
	//sort by UpdatedAt (most recent first)
	sort.Slice(bscp, func(i, j int) bool {
		return bscp[i].UpdatedAt.After(bscp[j].UpdatedAt)
	})
	// this might need a bit tweeking
	recentCount := 20
	if len(bscp) < recentCount {
		recentCount = len(bscp)
	}
	cm.Contexts[Recent].Branches = bscp[:recentCount]
}

func (cm *BranchContextMgr) RefreshSearchContext(q string) {
	// first, get the current note list
	// I am not sure whether this is correct.
	c := cm.currentContext
	if cm.currentContext == Search {
		c = cm.previousContext
	}
	bs := cm.Contexts[c].Branches

	//loop through and search
	if q == "" {
		cm.Contexts[Search].Branches = bs
		return
	}

	query := strings.ToLower(q)
	filteredBranches := make([]*models.Branch, 0)
	for _, b := range bs {
		if strings.Contains(strings.ToLower(b.Name), query) {
			filteredBranches = append(filteredBranches, b)
			continue
		}
	}
	// Sort by CreatedAt to maintain chronological order
	sort.Slice(filteredBranches, func(i, j int) bool {
		return filteredBranches[i].CreatedAt.Before(filteredBranches[j].CreatedAt)
	})
	cm.Contexts[Search].Branches = filteredBranches
}

// GetCurrentCursor returns the cursor position in the current context
func (cm *BranchContextMgr) GetCurrentCursor() uint {
	return cm.Contexts[cm.currentContext].Cursor
}

// SetCurrentCursor sets the cursor position in the current context
func (cm *BranchContextMgr) SetCurrentCursor(cursor uint) {
	cm.Contexts[cm.currentContext].Cursor = cursor
}

func (cm *BranchContextMgr) GetCurrentNote() *models.Branch {
	bs := cm.GetCurrentBranches()
	cursor := cm.GetCurrentCursor()
	if len(bs) == 0 || int(cursor) >= len(bs) {
		return nil
	}
	return bs[cursor]
}

// GetNoteCount returns the number of notes in the current context
func (cm *BranchContextMgr) GetNoteCount() int {
	return len(cm.GetCurrentBranches())
}

// This overall class might needs restructuring.
func (cm *BranchContextMgr) AddBranchToDefault(b *models.Branch) {
	bs := cm.Contexts[Default].Branches
	// Find the correct position to insert the note to maintain chronological order
	insertPos := len(bs)
	for i, n := range bs {
		if n.CreatedAt.After(b.CreatedAt) {
			insertPos = i
			break
		}
	}
	// Insert at the correct position
	if insertPos == len(bs) {
		cm.Contexts[Default].Branches = append(bs, b)
	} else {
		cm.Contexts[Default].Branches = append(bs[:insertPos], append([]*models.Branch{b}, bs[insertPos:]...)...)
	}
}

// RemoveNoteFromDefault removes a note from default context by ID
func (cm *BranchContextMgr) RemoveBranchFromDefault(ID uint) {
	bs := cm.Contexts[Default].Branches
	for i, b := range bs {
		if b.ID == ID {
			cm.Contexts[Default].Branches = append(bs[:i], bs[i+1:]...)
			break
		}
	}
}

// SortCurrentContext sorts the current context notes by CreatedAt (chronological order)
func (cm *BranchContextMgr) SortCurrentContext() {
	bs := cm.Contexts[cm.currentContext].Branches
	order := cm.Contexts[cm.currentContext].Order

	// well this is in place ! careful!
	switch order {
	case CreateAt:
		sort.Slice(bs, func(i, j int) bool {
			return bs[i].CreatedAt.Before(bs[j].CreatedAt)
		})
	case UpdateAt:
		sort.Slice(bs, func(i, j int) bool {
			return bs[i].UpdatedAt.After(bs[j].UpdatedAt)
		})
	default:
		sort.Slice(bs, func(i, j int) bool {
			return bs[i].CreatedAt.Before(bs[j].CreatedAt)
		})
	}
}

// only for data storage purposes. Do not use in coding.
func (cm *BranchContextMgr) GetCursors() map[ContextPtr]uint {
	return map[ContextPtr]uint{
		Default: cm.Contexts[Default].Cursor,
		Recent:  cm.Contexts[Recent].Cursor,
		Search:  cm.Contexts[Search].Cursor,
	}
}

func (cm *BranchContextMgr) SetCursors(m map[ContextPtr]uint) {
	cm.Contexts[Default].Cursor = m[Default]
	cm.Contexts[Recent].Cursor = m[Recent]
	cm.Contexts[Search].Cursor = m[Search]
}

/* Threads */

type ThreadContext struct {
	Name    ContextPtr
	Threads []*models.Thread
	Order   ContextOrder
	Cursor  uint
}

type ThreadContextMgr struct {
	previousContext ContextPtr
	currentContext  ContextPtr
	Contexts        []*ThreadContext
}

// NewThreadContextMgr creates a new context manager with all contexts initialized
func NewThreadContextMgr() *ThreadContextMgr {
	return &ThreadContextMgr{
		previousContext: None,
		currentContext:  Default,
		Contexts: []*ThreadContext{
			{Name: Default, Threads: make([]*models.Thread, 0), Cursor: 0, Order: CreateAt}, // Default context
			{Name: Recent, Threads: make([]*models.Thread, 0), Cursor: 0, Order: UpdateAt},  // Recent context
			{Name: Search, Threads: make([]*models.Thread, 0), Cursor: 0, Order: CreateAt},  // Search context
		},
	}
}

func (cm *ThreadContextMgr) SwitchContext(c ContextPtr) {
	if c != cm.currentContext {
		cm.previousContext = cm.currentContext
		cm.currentContext = c
	}
}

func (cm *ThreadContextMgr) GetCurrentThreads() []*models.Thread {
	return cm.Contexts[cm.currentContext].Threads
}

func (cm *ThreadContextMgr) GetCurrentContext() ContextPtr {
	return cm.currentContext
}

func (cm *ThreadContextMgr) GetPreviousContext() ContextPtr {
	return cm.previousContext
}

func (cm *ThreadContextMgr) RefreshDefaultContext(threads []*models.Thread) {
	cm.Contexts[Default].Threads = threads
}

func (cm *ThreadContextMgr) RefreshRecentContext() {
	threads := cm.Contexts[Default].Threads
	threadsCopy := make([]*models.Thread, len(threads))
	copy(threadsCopy, threads)
	sort.Slice(threadsCopy, func(i, j int) bool {
		return threadsCopy[i].UpdatedAt.After(threadsCopy[j].UpdatedAt)
	})
	recentCount := 20
	if len(threadsCopy) < recentCount {
		recentCount = len(threadsCopy)
	}
	cm.Contexts[Recent].Threads = threadsCopy[:recentCount]
}

func (cm *ThreadContextMgr) RefreshSearchContext(q string) {
	c := cm.currentContext
	if cm.currentContext == Search {
		c = cm.previousContext
	}
	threads := cm.Contexts[c].Threads

	if q == "" {
		cm.Contexts[Search].Threads = threads
		return
	}

	query := strings.ToLower(q)
	filteredThreads := make([]*models.Thread, 0)
	for _, thread := range threads {
		if strings.Contains(strings.ToLower(thread.Name), query) {
			filteredThreads = append(filteredThreads, thread)
			continue
		}
	}
	sort.Slice(filteredThreads, func(i, j int) bool {
		return filteredThreads[i].CreatedAt.Before(filteredThreads[j].CreatedAt)
	})
	cm.Contexts[Search].Threads = filteredThreads
}

func (cm *ThreadContextMgr) GetCurrentCursor() uint {
	return cm.Contexts[cm.currentContext].Cursor
}

func (cm *ThreadContextMgr) SetCurrentCursor(cursor uint) {
	cm.Contexts[cm.currentContext].Cursor = cursor
}

func (cm *ThreadContextMgr) GetCurrentThread() *models.Thread {
	threads := cm.GetCurrentThreads()
	cursor := cm.GetCurrentCursor()
	if len(threads) == 0 || int(cursor) >= len(threads) {
		return nil
	}
	return threads[cursor]
}

func (cm *ThreadContextMgr) GetThreadCount() int {
	return len(cm.GetCurrentThreads())
}

func (cm *ThreadContextMgr) AddThreadToDefault(thread *models.Thread) {
	threads := cm.Contexts[Default].Threads
	insertPos := len(threads)
	for i, t := range threads {
		if t.CreatedAt.After(thread.CreatedAt) {
			insertPos = i
			break
		}
	}
	if insertPos == len(threads) {
		cm.Contexts[Default].Threads = append(threads, thread)
	} else {
		cm.Contexts[Default].Threads = append(threads[:insertPos], append([]*models.Thread{thread}, threads[insertPos:]...)...)
	}
}

func (cm *ThreadContextMgr) RemoveThreadFromDefault(threadID uint) {
	threads := cm.Contexts[Default].Threads
	for i, thread := range threads {
		if thread.ID == threadID {
			cm.Contexts[Default].Threads = append(threads[:i], threads[i+1:]...)
			break
		}
	}
}

func (cm *ThreadContextMgr) SortCurrentContext() {
	threads := cm.Contexts[cm.currentContext].Threads
	order := cm.Contexts[cm.currentContext].Order

	switch order {
	case CreateAt:
		sort.Slice(threads, func(i, j int) bool {
			return threads[i].CreatedAt.Before(threads[j].CreatedAt)
		})
	case UpdateAt:
		sort.Slice(threads, func(i, j int) bool {
			return threads[i].UpdatedAt.After(threads[j].UpdatedAt)
		})
	default:
		sort.Slice(threads, func(i, j int) bool {
			return threads[i].CreatedAt.Before(threads[j].CreatedAt)
		})
	}
}

func (cm *ThreadContextMgr) GetCursors() map[ContextPtr]uint {
	return map[ContextPtr]uint{
		Default: cm.Contexts[Default].Cursor,
		Recent:  cm.Contexts[Recent].Cursor,
		Search:  cm.Contexts[Search].Cursor,
	}
}

func (cm *ThreadContextMgr) SetCursors(m map[ContextPtr]uint) {
	cm.Contexts[Default].Cursor = m[Default]
	cm.Contexts[Recent].Cursor = m[Recent]
	cm.Contexts[Search].Cursor = m[Search]
}
