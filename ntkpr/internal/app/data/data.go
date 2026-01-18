package data

import "github.com/haochend413/ntkpr/internal/models"

// DataMgr should handle the switching logic between threads, branches and notes. It keeps record of all threads, and exposing current threads, branches and notes.
// DataMgr should only expose 1 threadlist, 1 branchlist, and 1 notelist. ContextMgr will Demonstrate based on that. THe optimization and storage should happen at this level.
// Maybe a context manager should have a DataMgr inside? I suppose.
// DataMgr will also be responsible for syncing with database, instead of being handled directly by App.
// It is the one-above layer of db.
// With this package, we can easily manage what local data is stored for our app.
// Cursor needs much more careful handling!
// I came up with this at Houston Airport IAH. What a stroke of genious.
// I will have to re-write the context handling stuff based on this. Awwww
type DataMgr struct {
	threads         []*models.Thread // All the threads
	branches        []*models.Branch // The branch list of the active thread
	notes           []*models.Note   // the note list of the active branch
	activeThreadPtr int              // which thread is active ? // Maybe this should be passed down.
	activeBranchPtr int              //which branch is active ? // Maybe this should be passed down.
	activeNotePtr   int              //This is less useful, just a cursor.
}

func NewDataMgr(threads []*models.Thread) *DataMgr {
	dm := &DataMgr{
		threads:         threads,
		activeThreadPtr: 0,
		activeBranchPtr: 0,
		activeNotePtr:   0,
	}

	// Initialize branches and notes from first thread if available
	if len(threads) > 0 {
		dm.branches = threads[0].Branches
		if len(dm.branches) > 0 {
			dm.notes = dm.branches[0].Notes
		} else {
			dm.notes = []*models.Note{}
		}
	} else {
		dm.branches = []*models.Branch{}
		dm.notes = []*models.Note{}
	}

	return dm
}

func (dm *DataMgr) NewDataMgr() *DataMgr {
	return &DataMgr{
		threads:  []*models.Thread{},
		branches: []*models.Branch{},
		notes:    []*models.Note{},
	}
}

func (dm *DataMgr) GetThreads() []*models.Thread {
	return dm.threads
}

func (dm *DataMgr) GetActiveThread() *models.Thread {
	if len(dm.threads) == 0 || dm.activeThreadPtr < 0 || dm.activeThreadPtr >= len(dm.threads) {
		return nil
	}
	return dm.threads[dm.activeThreadPtr]
}

func (dm *DataMgr) GetActiveBranchList() []*models.Branch {
	return dm.branches
}

func (dm *DataMgr) GetActiveBranch() *models.Branch {
	if len(dm.branches) == 0 || dm.activeBranchPtr < 0 || dm.activeBranchPtr >= len(dm.branches) {
		return nil
	}
	return dm.branches[dm.activeBranchPtr]
}

func (dm *DataMgr) GetActiveNoteList() []*models.Note {
	return dm.notes
}

func (dm *DataMgr) GetActiveNote() *models.Note {
	if len(dm.notes) == 0 || dm.activeNotePtr < 0 || dm.activeNotePtr >= len(dm.notes) {
		return nil
	}
	return dm.notes[dm.activeNotePtr]
}

// GetActiveThreadPtr returns the current thread pointer
func (dm *DataMgr) GetActiveThreadPtr() int {
	return dm.activeThreadPtr
}

// GetActiveBranchPtr returns the current branch pointer
func (dm *DataMgr) GetActiveBranchPtr() int {
	return dm.activeBranchPtr
}

// GetActiveNotePtr returns the current note pointer
func (dm *DataMgr) GetActiveNotePtr() int {
	return dm.activeNotePtr
}

// RefreshData updates datamgr with new thread list.
// This should come with states. Implement later.
func (dm *DataMgr) RefreshData(threads []*models.Thread, tc *int, bc *int, nc *int) {
	dm.threads = threads
	if tc == nil {
		dm.activeThreadPtr = 0
	} else {
		dm.activeThreadPtr = *tc
	}

	// Handle empty threads or out of bounds
	if len(dm.threads) == 0 || dm.activeThreadPtr >= len(dm.threads) {
		dm.activeThreadPtr = 0
		dm.branches = []*models.Branch{}
		dm.activeBranchPtr = 0
		dm.notes = []*models.Note{}
		dm.activeNotePtr = 0
		return
	}

	dm.branches = dm.threads[dm.activeThreadPtr].Branches

	if bc == nil {
		dm.activeBranchPtr = 0
	} else {
		dm.activeBranchPtr = *bc
	}

	// Handle empty branches or out of bounds
	if len(dm.branches) == 0 || dm.activeBranchPtr >= len(dm.branches) {
		dm.activeBranchPtr = 0
		dm.notes = []*models.Note{}
		dm.activeNotePtr = 0
		return
	}

	dm.notes = dm.branches[dm.activeBranchPtr].Notes

	if nc == nil {
		dm.activeNotePtr = 0
	} else {
		dm.activeNotePtr = *nc
	}

	if dm.activeNotePtr >= len(dm.notes) {
		dm.activeNotePtr = 0
	}
}

// SwitchActiveThread deals with switching threads. It updates the exposed branch list when we switch threads.
func (dm *DataMgr) SwitchActiveThread(cursor int) {

	if len(dm.threads) == 0 {
		dm.activeThreadPtr = 0
		dm.branches = []*models.Branch{}
		dm.activeBranchPtr = 0
		dm.notes = []*models.Note{}
		dm.activeNotePtr = 0
		return
	}

	if cursor < 0 || cursor >= len(dm.threads) {
		return
	}

	dm.activeThreadPtr = cursor
	dm.branches = dm.threads[cursor].Branches
	dm.activeBranchPtr = 0

	if len(dm.branches) > 0 {
		dm.notes = dm.branches[0].Notes
		dm.activeNotePtr = 0
	} else {
		dm.notes = []*models.Note{}
		dm.activeNotePtr = 0
	}
}

// SwitchActiveBranch switches to a different branch within the current thread and resets the note list.
func (dm *DataMgr) SwitchActiveBranch(cursor int) {
	if len(dm.branches) == 0 {
		dm.activeBranchPtr = 0
		dm.notes = []*models.Note{}
		dm.activeNotePtr = 0
		return
	}

	if cursor < 0 || cursor >= len(dm.branches) {
		return
	}

	dm.activeBranchPtr = cursor
	dm.notes = dm.branches[cursor].Notes
	dm.activeNotePtr = 0
}

// SwitchActiveNote switches to a different note within the current branch.
func (dm *DataMgr) SwitchActiveNote(cursor int) {
	if len(dm.notes) == 0 {
		dm.activeNotePtr = 0
		return
	}
	if cursor < 0 || cursor >= len(dm.notes) {
		return
	}
	dm.activeNotePtr = cursor
}

// AddThread adds a thread to thread list without switching to it.
func (dm *DataMgr) AddThread(t *models.Thread) {
	if t == nil {
		return
	}
	dm.threads = append(dm.threads, t)
}

// RemoveThread removes a thread at the given index and adjusts active pointers.
func (dm *DataMgr) RemoveThread(index int) {
	if index < 0 || index >= len(dm.threads) {
		return
	}

	dm.threads = append(dm.threads[:index], dm.threads[index+1:]...)

	if len(dm.threads) == 0 {
		dm.activeThreadPtr = 0
		dm.branches = []*models.Branch{}
		dm.activeBranchPtr = 0
		dm.notes = []*models.Note{}
		dm.activeNotePtr = 0
		return
	}
	if dm.activeThreadPtr >= len(dm.threads) {
		dm.activeThreadPtr = len(dm.threads) - 1
	}
	dm.SwitchActiveThread(dm.activeThreadPtr)
}

// AddBranch adds a branch to the current thread's branch list without switching to it.
func (dm *DataMgr) AddBranch(b *models.Branch) {
	if b == nil || len(dm.threads) == 0 || dm.activeThreadPtr >= len(dm.threads) {
		return
	}

	thread := dm.threads[dm.activeThreadPtr]
	thread.Branches = append(thread.Branches, b)
	dm.branches = thread.Branches
}

// RemoveBranch removes a branch at the given index from the current thread and adjusts active pointers.
func (dm *DataMgr) RemoveBranch(index int) {
	if len(dm.threads) == 0 || index < 0 || index >= len(dm.branches) || dm.activeThreadPtr >= len(dm.threads) {
		return
	}

	thread := dm.threads[dm.activeThreadPtr]
	thread.Branches = append(thread.Branches[:index], thread.Branches[index+1:]...)
	dm.branches = thread.Branches

	if len(dm.branches) == 0 {
		dm.activeBranchPtr = 0
		dm.notes = []*models.Note{}
		dm.activeNotePtr = 0
		return
	}
	if dm.activeBranchPtr >= len(dm.branches) {
		dm.activeBranchPtr = len(dm.branches) - 1
	}
	dm.SwitchActiveBranch(dm.activeBranchPtr)
}

// AddNote adds a note to the current branch's note list without switching to it.
func (dm *DataMgr) AddNote(n *models.Note) {
	if n == nil || len(dm.branches) == 0 || dm.activeBranchPtr >= len(dm.branches) {
		return
	}

	branch := dm.branches[dm.activeBranchPtr]
	branch.Notes = append(branch.Notes, n)
	dm.notes = branch.Notes
}

// RemoveNote removes a note at the given index from the current branch and adjusts active pointers.
func (dm *DataMgr) RemoveNote(index int) {
	if len(dm.branches) == 0 || index < 0 || index >= len(dm.notes) {
		return
	}

	branch := dm.branches[dm.activeBranchPtr]
	branch.Notes = append(branch.Notes[:index], branch.Notes[index+1:]...)
	dm.notes = branch.Notes

	if len(dm.notes) == 0 {
		dm.activeNotePtr = 0
	} else if dm.activeNotePtr >= len(dm.notes) {
		dm.activeNotePtr = len(dm.notes) - 1
	}
	// Don't call SwitchActiveNote if empty
	if len(dm.notes) > 0 {
		dm.SwitchActiveNote(dm.activeNotePtr)
	}
}
