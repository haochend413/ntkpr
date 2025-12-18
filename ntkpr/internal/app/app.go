package app

import (
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/haochend413/ntkpr/internal/db"
	"github.com/haochend413/ntkpr/internal/models"
	"github.com/haochend413/ntkpr/internal/types"
)

// App encapsulates application logic and state
type App struct {
	db *db.DB

	NotesMap            map[uint]*models.Note
	NotesList           []*models.Note
	FilteredNotesList   []*models.Note
	RecentNotes         []*models.Note
	CurrentNotesListPtr *[]*models.Note

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
		db:                dbConn,
		NotesMap:          make(map[uint]*models.Note),
		NotesList:         make([]*models.Note, 0),
		FilteredNotesList: make([]*models.Note, 0),
		RecentNotes:       make([]*models.Note, 0),
		Topics:            make(map[uint]*models.Topic),
		nextNoteCreateID:  1, // Default starting ID if database is empty
		PendingNoteIDs:    []uint{},
		DeletedNoteIDs:    []uint{},
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
	a.FilteredNotesList = notes
	a.NotesList = notes
	a.NotesMap = make(map[uint]*models.Note, len(notes))
	a.Topics = make(map[uint]*models.Topic, len(topics))
	// init recent
	recentNotes, _ := a.db.GetRecentNotes()
	a.RecentNotes = recentNotes
	for _, note := range notes {
		a.NotesMap[note.ID] = note
	}
	for _, topic := range topics {
		a.Topics[topic.ID] = topic
	}

	a.CurrentNotesListPtr = &a.FilteredNotesList
	// Query the database for the maximum ID, including deleted notes
	var maxID uint
	if err := a.db.Conn.Table("notes").Select("MAX(id)").Row().Scan(&maxID); err != nil {
		maxID = 0
	}

	// Set the next ID for note creation
	a.nextNoteCreateID = maxID + 1
}

// SearchNotes searches the current list and populates FilteredNotesList
func (a *App) SearchNotes(query string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	curr := *a.CurrentNotesListPtr
	if query == "" {
		a.FilteredNotesList = curr
		return
	}
	query = strings.ToLower(query)
	a.FilteredNotesList = make([]*models.Note, 0)
	for _, note := range curr {
		if strings.Contains(strings.ToLower(note.Content), query) {
			a.FilteredNotesList = append(a.FilteredNotesList, note)
			continue
		}
		for _, topic := range note.Topics {
			if strings.Contains(strings.ToLower(topic.Topic), query) {
				a.FilteredNotesList = append(a.FilteredNotesList, note)
				break
			}
		}
	}
	sort.Slice(a.FilteredNotesList, func(i, j int) bool {
		return a.FilteredNotesList[i].ID < a.FilteredNotesList[j].ID
	})
}

// SelectCurrentNote sets the current note based on table cursor
// This one should be slow since it turns. Maybe consider using more space. We should do this at the start of the program.
func (a *App) SelectCurrentNote(cursor int) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	//This might be wrong: 1. needs sort;
	if len(*a.CurrentNotesListPtr) == 0 || cursor >= len(*a.CurrentNotesListPtr) {
		a.currentNote = nil
		return
	}
	a.currentNote = (*a.CurrentNotesListPtr)[cursor]
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
	a.NotesList = append(a.NotesList, note)
	if a.CurrentNotesListPtr != &a.NotesList {
		*a.CurrentNotesListPtr = append(*a.CurrentNotesListPtr, note)
	}
	a.currentNote = note
}

func (a *App) UpdateRecentNotes() {
	d, _ := a.db.GetRecentNotes()
	//lock
	a.mutex.Lock()
	defer a.mutex.Unlock()
	// Sort by ID to match display order
	sort.Slice(d, func(i, j int) bool {
		return d[i].ID < d[j].ID
	})
	a.RecentNotes = d
}

func (a *App) UpdateCurrentList(s types.Selector) {
	switch s {
	case types.Default:
		a.CurrentNotesListPtr = &a.NotesList
	case types.Search:
		a.CurrentNotesListPtr = &a.FilteredNotesList
	case types.Recent:
		a.CurrentNotesListPtr = &a.RecentNotes
	}
	sort.Slice(*a.CurrentNotesListPtr, func(i, j int) bool {
		return (*a.CurrentNotesListPtr)[i].ID < (*a.CurrentNotesListPtr)[j].ID
	})
}

// SyncWithDatabase syncs in-memory changes to the database
// This only work before we sync everything.

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

	// Check if note already exists in NotesList (shouldn't happen but safeguard)
	alreadyInNotesList := false
	for _, note := range a.NotesList {
		if note.ID == deletedNote.ID {
			alreadyInNotesList = true
			break
		}
	}

	// Add back to NotesList if not already there
	if !alreadyInNotesList {
		a.NotesList = append(a.NotesList, deletedNote)
	}

	// Add back to current list if it's different from NotesList
	if a.CurrentNotesListPtr != &a.NotesList {
		alreadyInCurrentList := false
		for _, note := range *a.CurrentNotesListPtr {
			if note.ID == deletedNote.ID {
				alreadyInCurrentList = true
				break
			}
		}
		if !alreadyInCurrentList {
			*a.CurrentNotesListPtr = append(*a.CurrentNotesListPtr, deletedNote)
		}
	}

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
	a.NotesList = updatedNotes
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
		} else if len(a.NotesList) > 0 {
			// If current note was deleted, select the first note
			a.currentNote = a.NotesList[0]
		} else {
			a.currentNote = nil
		}
	}
}
