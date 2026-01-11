package app

import (
	"log"
	"sync"
	"time"

	"github.com/haochend413/ntkpr/internal/app/context"
	editstack "github.com/haochend413/ntkpr/internal/app/editStack"
	"github.com/haochend413/ntkpr/internal/db"
	"github.com/haochend413/ntkpr/internal/models"
	"github.com/haochend413/ntkpr/state"
)

// App encapsulates application logic and states
type App struct {
	db                 *db.DB
	ThreadMap          map[uint]*models.Thread // I hope this is enough, but I am not sure.
	Topics             map[uint]*models.Topic
	noteContextMgr     *context.NoteContextMgr
	branchContextMgr   *context.BranchContextMgr
	threadContextMgr   *context.ThreadContextMgr
	editMgr            *editstack.EditMgr
	currentThread      *models.Thread
	currentBranch      *models.Branch
	currentNote        *models.Note
	nextThreadCreateID uint
	nextBranchCreateID uint
	nextNoteCreateID   uint
	Synced             bool
	mutex              sync.Mutex
}

// NewApp creates a new application instance and restore app states
func NewApp(dbConn *db.DB, AppState *state.AppState) *App {
	threadCursors := AppState.ThreadCursors
	branchCursors := AppState.BranchCursors
	noteCursors := AppState.NoteCursors

	noteContextMgr := context.NewNoteContextMgr()
	branchContextMgr := context.NewBranchContextMgr()
	threadContextMgr := context.NewThreadContextMgr()

	noteContextMgr.SetCursors(noteCursors)
	branchContextMgr.SetCursors(branchCursors)
	threadContextMgr.SetCursors(threadCursors)

	app := &App{
		db:                 dbConn,
		ThreadMap:          make(map[uint]*models.Thread),
		Topics:             make(map[uint]*models.Topic),
		noteContextMgr:     context.NewNoteContextMgr(),
		branchContextMgr:   context.NewBranchContextMgr(),
		threadContextMgr:   context.NewThreadContextMgr(),
		editMgr:            editstack.NewEditMgr(),
		nextThreadCreateID: 1,
		nextBranchCreateID: 1,
		nextNoteCreateID:   1,
		Synced:             true,
	}

	app.loadData() // This sets some other fields after loading data
	return app
}

// This function also load all topics
/*
Is this really necessary ? Let's keep everything within threads, shall we?
In that case we might need more reads and writes.
During re-structuring, clear API structure is more and more valuable.
Maybe we should start with that.

Yes. We probably also need to change the sync mechanism for it.
*/

// loadNotes loads threads from the database into the app level, and set related app states.
// threads come with branches preloaded with their notes, used for search and rendering.

func (a *App) loadData() {
	//fetch data from database;
	// This needs further fix!!!
	_, topics, threads, _, err := a.db.SyncData(make(map[uint]*models.Note), make(map[uint]*models.Thread), make(map[uint]*models.Branch), make(map[uint]*editstack.Edit))
	if err != nil {
		log.Panic(err)
	}
	a.mutex.Lock()
	defer a.mutex.Unlock()
	// fill in the variables
	// a.NotesMap = make(map[uint]*models.Note, len(notes))
	a.ThreadMap = make(map[uint]*models.Thread, len(threads))
	// a.BranchMap = make(map[uint]*models.Branch, len(notes))
	a.Topics = make(map[uint]*models.Topic, len(topics))

	// Ok, what we have is a ThreadMap. We need to fetch by its inner cursor: which Branch ? this is branch cursor.
	// Wait, what do we need to do here ?
	// Thread -> which thread ? cursor for default
	a.threadContextMgr.RefreshDefaultContext(threads)
	a.threadContextMgr.RefreshRecentContext()
	// Now, fetch the corrcet branch and note data, from the previous cursors. Notice that we currently do not need cursor right now.
	// Yeah...well, we might need a helper for it. And update the wrappings.
	// For a context, we should have cursors available ? This need some proper design, but maybe not now.
	branches := threads[a.threadContextMgr.Contexts[context.Default].Cursor].Branches
	a.branchContextMgr.RefreshDefaultContext(branches)
	notes := branches[a.branchContextMgr.Contexts[context.Default].Cursor].Notes
	a.branchContextMgr.RefreshRecentContext()
	a.noteContextMgr.RefreshDefaultContext(notes)
	a.noteContextMgr.RefreshRecentContext()

	// for _, note := range notes {
	// 	a.NotesMap[note.ID] = note
	// }
	for _, topic := range topics {
		a.Topics[topic.ID] = topic
	}
	for _, thread := range threads {
		a.ThreadMap[thread.ID] = thread
	}
	// for _, branch := range branches {
	// 	a.BranchMap[branch.ID] = branch
	// }
	// Set the next IDs for creation
	a.nextNoteCreateID = a.db.GetCreateNoteID()
	a.nextBranchCreateID = a.db.GetCreateBranchID()
	a.nextThreadCreateID = a.db.GetCreateThreadID()

	// Set current context and pointers for app
	a.SelectCurrentThread()
}

// Switch to a different context for notes list, stores cursor for previous context, and return new cursor position.
func (a *App) UpdateCurrentNoteContext(c context.ContextPtr, currentCursor uint) uint {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	c0 := a.noteContextMgr.GetCurrentContext()
	a.noteContextMgr.Contexts[c0].Cursor = currentCursor
	a.noteContextMgr.SwitchContext(c)
	a.noteContextMgr.SortCurrentContext()
	return a.noteContextMgr.Contexts[c].Cursor
}

// Switch to a different context for branches list, stores cursor for previous context, and return new cursor position.
func (a *App) UpdateCurrentBranchContext(c context.ContextPtr, currentCursor uint) uint {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	c0 := a.branchContextMgr.GetCurrentContext()
	a.branchContextMgr.Contexts[c0].Cursor = currentCursor
	a.branchContextMgr.SwitchContext(c)
	a.branchContextMgr.SortCurrentContext()
	return a.branchContextMgr.Contexts[c].Cursor
}

// Switch to a different context for threads list, stores cursor for previous context, and return new cursor position.
func (a *App) UpdateCurrentThreadContext(c context.ContextPtr, currentCursor uint) uint {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	c0 := a.threadContextMgr.GetCurrentContext()
	a.threadContextMgr.Contexts[c0].Cursor = currentCursor
	a.threadContextMgr.SwitchContext(c)
	a.threadContextMgr.SortCurrentContext()
	return a.threadContextMgr.Contexts[c].Cursor
}

// SearchNotes searches the current list and populates search context
func (a *App) SearchNotes(query string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.noteContextMgr.RefreshSearchContext(query)
	a.noteContextMgr.SwitchContext(context.Search)
}

// SearchBranches searches the current branch list and populates search context
func (a *App) SearchBranches(query string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.branchContextMgr.RefreshSearchContext(query)
	a.branchContextMgr.SwitchContext(context.Search)
}

// SearchThreads searches the current thread list and populates search context
func (a *App) SearchThreads(query string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.threadContextMgr.RefreshSearchContext(query)
	a.threadContextMgr.SwitchContext(context.Search)
}

// SelectCurrentNote sets the current note based on table cursor
func (a *App) SelectCurrentNote(cursor int) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	notes := a.noteContextMgr.GetCurrentNotes()
	if len(notes) == 0 || cursor >= len(notes) {
		a.currentNote = nil
		return
	}
	a.currentNote = notes[cursor]
	a.noteContextMgr.SetCurrentCursor(uint(cursor))
}

// SelectCurrentBranch sets the current branch based on table cursor
func (a *App) SelectCurrentBranch(cursor int) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	branches := a.branchContextMgr.GetCurrentBranches()
	if len(branches) == 0 || cursor >= len(branches) {
		a.currentBranch = nil
		return
	}
	a.currentBranch = branches[cursor]
	a.branchContextMgr.SetCurrentCursor(uint(cursor))
}

// SelectCurrentThread sets the current thread based on table cursor
func (a *App) SelectCurrentThread(cursor int) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	threads := a.threadContextMgr.GetCurrentThreads()
	if len(threads) == 0 || cursor >= len(threads) {
		a.currentThread = nil
		return
	}
	a.currentThread = threads[cursor]
	a.threadContextMgr.SetCurrentCursor(uint(cursor))
}

// CreateNewNote creates a new pending note
func (a *App) CreateNewNote() {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	note := &models.Note{Content: ""}
	note.CreatedAt = time.Now()
	note.UpdatedAt = time.Now()
	note.ID = a.nextNoteCreateID
	a.nextNoteCreateID += 1
	a.Synced = false
	edit := &editstack.Edit{EditType: editstack.CreateNote, ID: note.ID}
	if err := a.editMgr.AddEdit(edit); err != nil {
		log.Printf("Error adding Create edit: %v", err)
		return
	}
	a.currentBranch.Notes[note.ID] = note   // This is synced with database
	a.noteContextMgr.AddNoteToDefault(note) // This list is handled with context
	a.currentNote = note
}

func (a *App) UpdateRecentNotes() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.contextMgr.RefreshRecentContext()
}

// GetCurrentNotes returns the notes in the current context
func (a *App) GetCurrentNotes() []*models.Note {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	return a.contextMgr.GetCurrentNotes()
}

// GetCurrentContext returns the current context
func (a *App) GetCurrentContext() context.ContextPtr {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	return a.contextMgr.GetCurrentContext()
}

// GetEditStack returns the edit stack for UI access
func (a *App) GetEditStack() []uint {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	return a.editMgr.EditStack
}

// GetEdit returns an edit by ID
func (a *App) GetEdit(id uint) *editstack.Edit {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	return a.editMgr.EditMap[id]
}

// UndoDelete undoes the last delete operation
func (a *App) UndoDelete() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	// Find the most recent delete from the edit stack
	var lastDeletedID uint
	for i := len(a.editMgr.EditStack) - 1; i >= 0; i-- {
		id := a.editMgr.EditStack[i]
		if edit, exists := a.editMgr.EditMap[id]; exists && edit.EditType == editstack.Delete {
			lastDeletedID = id
			break
		}
	}

	if lastDeletedID == 0 {
		return
	}

	deletedNote, exists := a.NotesMap[lastDeletedID]
	if !exists {
		return
	}

	// Remove the delete edit
	a.editMgr.RemoveEdit(lastDeletedID)

	// Add back to default context
	a.contextMgr.AddNoteToDefault(deletedNote)
	a.Synced = false
}

// only for data storage purposes. Do not use in coding.
func (a *App) GetCursors() map[context.ContextPtr]uint {
	return a.contextMgr.GetCursors()
}
func (a *App) SetCursors(m map[context.ContextPtr]uint) {
	a.contextMgr.SetCursors(m)
}

func (a *App) SyncWithDatabase() {
	a.mutex.Lock()
	// Make a copy of the notes map and editMap
	notesMapCopy := make(map[uint]*models.Note, len(a.NotesMap))
	for id, note := range a.NotesMap {
		notesMapCopy[id] = note
	}
	editMapCopy := make(map[uint]*editstack.Edit, len(a.editMgr.EditMap))
	for id, edit := range a.editMgr.EditMap {
		editMapCopy[id] = edit
	}
	a.mutex.Unlock()
	// Sync with the database using the editMap directly
	updatedNotes, updatedTopics, err := a.db.SyncData(notesMapCopy, editMapCopy)
	if err != nil {
		log.Printf("Error syncing with database: %v", err)
		return
	}

	a.mutex.Lock()
	defer a.mutex.Unlock()

	// Update in-memory state with the fresh data from database
	a.contextMgr.RefreshDefaultContext(updatedNotes)
	a.NotesMap = make(map[uint]*models.Note, len(updatedNotes))
	a.Topics = make(map[uint]*models.Topic, len(updatedTopics))

	// Find max ID for next note creation
	maxID := uint(0)
	for _, note := range updatedNotes {
		a.NotesMap[note.ID] = note
		if note.ID > maxID {
			maxID = note.ID
		}
	}

	// Update topics
	for _, topic := range updatedTopics {
		a.Topics[topic.ID] = topic
	}

	// Set next ID and clear edit manager
	a.nextNoteCreateID = maxID + 1
	a.editMgr.Clear()
	a.Synced = true

	// If we had a current note, try to find it in the updated notes
	if a.currentNote != nil {
		currentID := a.currentNote.ID
		if note, exists := a.NotesMap[currentID]; exists {
			a.currentNote = note
		} else {
			// If current note was deleted, select the first note or nil
			a.currentNote = a.contextMgr.GetCurrentNote()
		}
	}
}
