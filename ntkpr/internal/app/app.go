package app

import (
	"log"
	"sync"
	"time"

	"github.com/haochend413/ntkpr/internal/app/context"
	editstack "github.com/haochend413/ntkpr/internal/app/editStack"
	"github.com/haochend413/ntkpr/internal/db"
	"github.com/haochend413/ntkpr/internal/models"
)

// App encapsulates application logic and state
type App struct {
	db *db.DB
	NotesMap   map[uint]*models.Note
	contextMgr *context.ContextMgr
	editMgr    *editstack.EditMgr
	currentNote    *models.Note
	nextNoteCreateID uint
	Synced           bool
	Topics           map[uint]*models.Topic
	mutex            sync.Mutex
}

// NewApp creates a new application instance
func NewApp(dbConn *db.DB) *App {
	app := &App{
		db:               dbConn,
		NotesMap:         make(map[uint]*models.Note),
		contextMgr:       context.NewContextMgr(),
		editMgr:          editstack.NewEditMgr(),
		Topics:           make(map[uint]*models.Topic),
		nextNoteCreateID: 1, // Default starting ID if database is empty
		Synced:           true,
	}

	//load everything into the app
	app.loadData()
	return app
}

// loadNotes loads notes from the database
// This function also load all topics
func (a *App) loadData() {
	//fetch data from database;
	notes, topics, err := a.db.SyncData(make(map[uint]*models.Note), make(map[uint]*editstack.Edit))
	if err != nil {
		log.Panic(err)
	}
	a.mutex.Lock()
	defer a.mutex.Unlock()
	// fill in the variables
	a.NotesMap = make(map[uint]*models.Note, len(notes))
	a.Topics = make(map[uint]*models.Topic, len(topics))
	
	// Initialize context manager with all notes
	a.contextMgr.RefreshDefaultContext(notes)
	
	// init recent
	a.contextMgr.RefreshRecentContext()
	
	for _, note := range notes {
		a.NotesMap[note.ID] = note
	}
	for _, topic := range topics {
		a.Topics[topic.ID] = topic
	}

	// Query the database for the maximum ID, including deleted notes
	var maxID uint
	if err := a.db.Conn.Table("notes").Select("MAX(id)").Row().Scan(&maxID); err != nil {
		maxID = 0
	}

	// Set the next ID for note creation
	a.nextNoteCreateID = maxID + 1
}

// SearchNotes searches the current list and populates search context
func (a *App) SearchNotes(query string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	
	a.contextMgr.RefreshSearchContext(query)
	a.contextMgr.SwitchContext(context.Search)
}

// SelectCurrentNote sets the current note based on table cursor
// This one should be slow since it turns. Maybe consider using more space. We should do this at the start of the program.
func (a *App) SelectCurrentNote(cursor int) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	//This might be wrong: 1. needs sort;
	notes := a.contextMgr.GetCurrentNotes()
	if len(notes) == 0 || cursor >= len(notes) {
		a.currentNote = nil
		return
	}
	a.currentNote = notes[cursor]
	a.contextMgr.SetCurrentCursor(uint(cursor))
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
	if err := a.editMgr.AddEdit(editstack.Create, note.ID); err != nil {
		log.Printf("Error adding Create edit: %v", err)
		return
	}
	a.NotesMap[note.ID] = note
	a.contextMgr.AddNoteToDefault(note)
	a.currentNote = note
}

func (a *App) UpdateRecentNotes() {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	
	a.contextMgr.RefreshRecentContext()
}

func (a *App) UpdateCurrentList(c context.ContextPtr) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	
	a.contextMgr.SwitchContext(c)
	a.contextMgr.SortCurrentContext()
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
