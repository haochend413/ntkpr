package app

import (
	"log"
	"sync"
	"time"

	"github.com/haochend413/ntkpr/internal/app/context"
	"github.com/haochend413/ntkpr/internal/db"
	"github.com/haochend413/ntkpr/internal/models"
)

// App encapsulates application logic and state
type App struct {
	db *db.DB
	NotesMap   map[uint]*models.Note
	contextMgr *context.ContextMgr
	currentNote    *models.Note
	PendingNoteIDs []uint
	DeletedNoteIDs []uint
	CreateNoteIDs  []uint
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
		Topics:           make(map[uint]*models.Topic),
		nextNoteCreateID: 1, // Default starting ID if database is empty
		PendingNoteIDs:   []uint{},
		DeletedNoteIDs:   []uint{},
	}

	//load everything into the app
	app.loadData()
	return app
}

// loadNotes loads notes from the database
// This function also load all topics
func (a *App) loadData() {
	//fetch data from database;
	notes, topics, err := a.db.SyncData(make(map[uint]*models.Note), nil, nil, nil)
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
	a.CreateNoteIDs = append(a.CreateNoteIDs, note.ID)
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

// SyncWithDatabase syncs in-memory changes to the database
// This only work before we sync everything.

// GetLastDeletedNoteID returns the ID of the last deleted note without removing it
func (a *App) GetLastDeletedNoteID() uint {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if len(a.DeletedNoteIDs) == 0 {
		return 0
	}
	return a.DeletedNoteIDs[len(a.DeletedNoteIDs)-1]
}

func (a *App) UndoDelete() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if len(a.DeletedNoteIDs) == 0 {
		return
	}
	lastDeletedID := a.DeletedNoteIDs[len(a.DeletedNoteIDs)-1]
	deletedNote, exists := a.NotesMap[lastDeletedID]
	if !exists {
		return
	}

	// Remove from deleted list
	a.DeletedNoteIDs = a.DeletedNoteIDs[:len(a.DeletedNoteIDs)-1]

	// Add back to default context
	a.contextMgr.AddNoteToDefault(deletedNote)
	a.Synced = false
}

func (a *App) SyncWithDatabase() {
	a.mutex.Lock()
	// Make a copy of the IDs slices for the database operation
	pendingIDs := append([]uint{}, a.PendingNoteIDs...)
	deletedIDs := append([]uint{}, a.DeletedNoteIDs...)
	createIDs := append([]uint{}, a.CreateNoteIDs...)
	// Make a copy of the notes map
	notesMapCopy := make(map[uint]*models.Note, len(a.NotesMap))
	for id, note := range a.NotesMap {
		notesMapCopy[id] = note
	}
	a.mutex.Unlock()
	// Sync with the database
	updatedNotes, updatedTopics, err := a.db.SyncData(notesMapCopy, pendingIDs, deletedIDs, createIDs)
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

	// Set next ID and clear pending/deleted lists
	a.nextNoteCreateID = maxID + 1
	a.PendingNoteIDs = []uint{}
	a.DeletedNoteIDs = []uint{}
	a.CreateNoteIDs = []uint{}
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
