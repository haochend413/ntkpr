package app

import (
	"log"
	"sync"
	"time"

	"github.com/haochend413/ntkpr/internal/app/data"
	editstack "github.com/haochend413/ntkpr/internal/app/editStack"
	"github.com/haochend413/ntkpr/internal/db"
	"github.com/haochend413/ntkpr/internal/models"
	"github.com/haochend413/ntkpr/state"
)

// App encapsulates application logic and states
// Inside app we deal with how our local data, stored in contexts, interact with database.
// In my opinion, we can just re-write the whole thing.
type App struct {
	db                 *db.DB
	dataMgr            *data.DataMgr
	editMgr            *editstack.EditMgr
	nextThreadCreateID uint
	nextBranchCreateID uint
	nextNoteCreateID   uint
	Synced             bool
	mutex              sync.Mutex
}

// NewApp creates a new application instance and restore app states
func NewApp(dbConn *db.DB, AppState *state.AppState) *App {

	app := &App{
		db:                 dbConn,
		dataMgr:            &data.DataMgr{},
		editMgr:            editstack.NewEditMgr(),
		nextThreadCreateID: 1,
		nextBranchCreateID: 1,
		nextNoteCreateID:   1,
		Synced:             true,
	}

	app.loadData()
	return app
}

// GetDataMgr returns the data manager
func (a *App) GetDataMgr() *data.DataMgr {
	return a.dataMgr
}

// GetEditMap returns the current edit map
func (a *App) GetEditMap() map[editstack.EditKey]*editstack.Edit {
	return a.editMgr.EditMap
}

// loadData loads threads from the database and initializes data manager
func (a *App) loadData() {
	// fetch from db
	_, threads, err := a.db.SyncData(
		[]*models.Thread{},
		make(map[editstack.EditKey]*editstack.Edit),
	)

	if err != nil {
		log.Panic(err)
	}

	a.dataMgr = data.NewDataMgr(threads)

	// Set the next IDs for creation
	a.nextNoteCreateID = a.db.GetCreateNoteID()
	a.nextBranchCreateID = a.db.GetCreateBranchID()
	a.nextThreadCreateID = a.db.GetCreateThreadID()
}

/*
APIs to call, connecting context and database.
*/

func (a *App) CreateNewThread() {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	thread := &models.Thread{Name: ""}
	thread.CreatedAt = time.Now()
	thread.UpdatedAt = time.Now()
	thread.ID = a.nextThreadCreateID
	a.nextThreadCreateID += 1
	a.Synced = false
	edit := &editstack.Edit{EditType: editstack.CreateThread, ID: thread.ID}
	if err := a.editMgr.AddEdit(edit); err != nil {
		log.Printf("Error adding Create edit: %v", err)
		return
	}
	a.dataMgr.AddThread(thread)
}

func (a *App) CreateNewBranch() {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	thread := a.dataMgr.GetActiveThread()
	if thread == nil {
		log.Printf("Cannot create branch: no active thread")
		return
	}
	branch := &models.Branch{Name: ""}
	branch.CreatedAt = time.Now()
	branch.UpdatedAt = time.Now()
	branch.ID = a.nextBranchCreateID
	branch.ThreadID = thread.ID
	a.nextBranchCreateID += 1
	a.Synced = false
	edit := &editstack.Edit{EditType: editstack.CreateBranch, ID: branch.ID}
	if err := a.editMgr.AddEdit(edit); err != nil {
		log.Printf("Error adding Create edit: %v", err)
		return
	}
	a.dataMgr.AddBranch(branch)
}

// CreateNewNote creates a new pending note
func (a *App) CreateNewNote() {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	thread := a.dataMgr.GetActiveThread()
	branch := a.dataMgr.GetActiveBranch()
	if thread == nil {
		log.Printf("Cannot create note: no active thread")
		return
	}
	if branch == nil {
		log.Printf("Cannot create note: no active branch")
		return
	}
	note := &models.Note{Content: ""}
	note.CreatedAt = time.Now()
	note.UpdatedAt = time.Now()
	note.ID = a.nextNoteCreateID
	note.ThreadID = thread.ID
	// note.Branches = []*models.Branch{branch}
	a.nextNoteCreateID += 1
	a.Synced = false

	edit := &editstack.Edit{EditType: editstack.CreateNote, ID: note.ID}
	if err := a.editMgr.AddEdit(edit); err != nil {
		log.Printf("Error adding Create edit: %v", err)
		return
	}

	// Mark the branch as updated only if it already exists in the DB
	// If the branch is pending (being created), the CreateBranch will handle the association
	branchEdit, branchExists := a.editMgr.GetEdit(editstack.EntityBranch, branch.ID)
	if !branchExists || branchEdit.EditType != editstack.CreateBranch {
		// Branch either doesn't have pending edits or is not being created
		// Safe to mark for update
		updateEdit := &editstack.Edit{ID: branch.ID, EditType: editstack.UpdateBranch}
		a.editMgr.AddEdit(updateEdit) // Ignore error - branch might already be marked
	}

	a.dataMgr.AddNote(note)
}

func (a *App) GetThreadList() []*models.Thread {
	if a == nil {
		log.Panic("null app")
	}
	return a.dataMgr.GetThreads()
}

func (a *App) GetActiveBranchList() []*models.Branch {
	if a == nil {
		log.Panic("null app")
	}
	return a.dataMgr.GetActiveBranchList()
}

func (a *App) GetActiveNoteList() []*models.Note {
	if a == nil {
		log.Panic("null app")
	}
	return a.dataMgr.GetActiveNoteList()
}

// // GetEditStack returns the edit stack for UI access
// func (a *App) GetEditMgr() *editstack.EditMgr {
// 	return a.editMgr
// }

/*
This can be implemented later.
*/

// // UndoDelete undoes the last delete operation
// func (a *App) UndoDelete() {
// 	a.mutex.Lock()
// 	defer a.mutex.Unlock()

// 	currentBranch := a.branchContextMgr.GetCurrentBranch()
// 	if currentBranch == nil {
// 		log.Printf("No current branch selected")
// 		return
// 	}

// 	// Find the most recent delete from the edit stack
// 	var lastDeleteEdit *editstack.Edit
// 	for i := len(a.editMgr.EditStack) - 1; i >= 0; i-- {
// 		edit := a.editMgr.EditStack[i]
// 		if edit.EditType == editstack.Delete {
// 			lastDeleteEdit = edit
// 			break
// 		}
// 	}

// 	if lastDeleteEdit == nil {
// 		return
// 	}

// 	deletedNote, exists := currentBranch.Notes[lastDeleteEdit.ID]
// 	if !exists {
// 		return
// 	}

// 	// Remove the delete edit
// 	a.editMgr.RemoveEdit(lastDeleteEdit.ID)

// 	// Add back to default context
// 	a.noteContextMgr.AddNoteToDefault(deletedNote)
// 	a.Synced = false
// }

// SyncWithDatabase syncs the current state with the database
func (a *App) SyncWithDatabase() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	// Get threads from data manager and copy edit map
	threads := a.dataMgr.GetThreads()
	editMapCopy := make(map[editstack.EditKey]*editstack.Edit)
	for k, v := range a.editMgr.EditMap {
		editMapCopy[k] = v
	}

	// Sync with the database
	_, updatedThreads, err := a.db.SyncData(threads, editMapCopy)

	if err != nil {
		log.Printf("Error syncing with database: %v", err)
		return
	}

	tc := a.dataMgr.GetActiveThreadPtr()
	bc := a.dataMgr.GetActiveBranchPtr()
	nc := a.dataMgr.GetActiveNotePtr()

	// Refresh data manager with updated threads
	a.dataMgr.RefreshData(updatedThreads, &tc, &bc, &nc)

	// Get next IDs from database (includes soft-deleted records)
	// This ensures we never reuse an ID that exists in the DB
	a.nextNoteCreateID = a.db.GetCreateNoteID()
	a.nextBranchCreateID = a.db.GetCreateBranchID()
	a.nextThreadCreateID = a.db.GetCreateThreadID()
	a.editMgr.Clear()
	a.Synced = true
}
