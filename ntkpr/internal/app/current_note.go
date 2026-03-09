package app

import (
	"log"
	"time"

	editstack "github.com/haochend413/ntkpr/internal/app/editStack"
	"github.com/haochend413/ntkpr/internal/models"
)

// current_note.go provides a controlled interface for accessing and modifying
// the currently selected note, with proper edit tracking and synchronization.

// =============================================================================
// Helper: Get current note with safety checks
// =============================================================================

func (a *App) getCurrentNote() *models.Note {
	if a.dataMgr == nil {
		log.Fatal("Critical error: dataMgr is nil - app not properly initialized")
	}
	return a.dataMgr.GetActiveNote()
}

// =============================================================================
// Getters - Read-only access to current note properties
// =============================================================================

// GetCurrentNoteContent returns the content of the current note, or empty string if none selected
func (a *App) GetCurrentNoteContent() string {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	note := a.getCurrentNote()
	if note == nil {
		return ""
	}
	return note.Content
}

// GetCurrentNoteID returns the ID of the current note, or 0 if none selected
func (a *App) GetCurrentNoteID() uint {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	note := a.getCurrentNote()
	if note == nil {
		return 0
	}
	return note.ID
}

// GetCurrentNoteTopics returns a copy of the current note's topics to prevent external mutation
// (Topics removed) GetCurrentNoteTopics no longer available.

// GetCurrentNoteHighlight returns whether the current note is highlighted
func (a *App) GetCurrentNoteHighlight() bool {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	note := a.getCurrentNote()
	if note == nil {
		return false
	}
	return note.Highlight
}

// GetCurrentNotePrivate returns whether the current note is private
func (a *App) GetCurrentNotePrivate() bool {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	note := a.getCurrentNote()
	if note == nil {
		return false
	}
	return note.Private
}

// GetCurrentNoteFrequency returns the edit count of the current note
func (a *App) GetCurrentNoteFrequency() int {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	note := a.getCurrentNote()
	if note == nil {
		return 0
	}
	return note.Frequency
}

// GetCurrentNoteLastEdit returns the last edit timestamp of the current note.
func (a *App) GetCurrentNoteLastEdit() time.Time {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	note := a.getCurrentNote()
	if note == nil {
		return time.Time{}
	}
	return note.LastEdit
}

// GetCurrentNoteUpdatedAt returns when the current note was last modified
func (a *App) GetCurrentNoteUpdatedAt() time.Time {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	note := a.getCurrentNote()
	if note == nil {
		return time.Time{}
	}
	return note.UpdatedAt
}

// GetCurrentNoteCreatedAt returns when the current note was created
func (a *App) GetCurrentNoteCreatedAt() time.Time {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	note := a.getCurrentNote()
	if note == nil {
		return time.Time{}
	}
	return note.CreatedAt
}

// HasCurrentNote checks if a note is currently selected
func (a *App) HasCurrentNote() bool {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	return a.getCurrentNote() != nil
}

// =============================================================================
// Setters - Controlled modification with edit tracking
// =============================================================================

// SetCurrentNoteContent updates the current note's content with edit tracking
func (a *App) SetCurrentNoteContent(content string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	note := a.getCurrentNote()
	if note == nil {
		return
	}

	// No-op if content hasn't changed
	if note.Content == content {
		return
	}

	note.Content = content
	note.Frequency++
	note.LastEdit = time.Now()
	a.Synced = false

	edit := &editstack.Edit{ID: note.ID, EditType: editstack.UpdateNote}
	if err := a.editMgr.AddEdit(edit); err != nil {
		log.Printf("Error tracking note update: %v", err)
	}
}

// SetCurrentNoteLastEdit updates the LastEdit timestamp of the current note to the current time.
// Ensures the timestamp is not set to a past time.
func (a *App) SetCurrentNoteLastEdit() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	note := a.getCurrentNote()
	if note == nil {
		return
	}

	// dont set it backwards
	if time.Now().Before(note.LastEdit) {
		return
	}
	note.LastEdit = time.Now()
}

// ToggleCurrentNoteHighlight toggles the highlight status of the current note
func (a *App) ToggleCurrentNoteHighlight() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	note := a.getCurrentNote()
	if note == nil {
		return
	}

	note.Highlight = !note.Highlight
	note.UpdatedAt = time.Now()
	a.Synced = false

	edit := &editstack.Edit{ID: note.ID, EditType: editstack.UpdateNote}
	if err := a.editMgr.AddEdit(edit); err != nil {
		log.Printf("Error tracking note update: %v", err)
	}
}

// ToggleCurrentNotePrivate toggles the private status of the current note
func (a *App) ToggleCurrentNotePrivate() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	note := a.getCurrentNote()
	if note == nil {
		return
	}

	note.Private = !note.Private
	note.UpdatedAt = time.Now()
	a.Synced = false

	edit := &editstack.Edit{ID: note.ID, EditType: editstack.UpdateNote}
	if err := a.editMgr.AddEdit(edit); err != nil {
		log.Printf("Error tracking note update: %v", err)
	}
}

// =============================================================================
// Topic Management
// =============================================================================
// Topic subsystem removed: topic add/remove APIs have been removed.

// DeleteCurrentNote removes the current note from the current branch and tracks the deletion
func (a *App) DeleteCurrentNote() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	note := a.getCurrentNote()
	if note == nil {
		return
	}

	noteID := note.ID
	branch := a.getCurrentBranch()

	edit, exists := a.editMgr.GetEdit(editstack.EntityNote, noteID)
	if exists && edit.EditType == editstack.CreateNote {
		a.editMgr.RemoveEdit(editstack.EntityNote, noteID)
	} else if noteID != 0 {
		deleteEdit := &editstack.Edit{ID: noteID, EditType: editstack.DeleteNote}
		if err := a.editMgr.AddEdit(deleteEdit); err != nil {
			log.Printf("Error tracking note deletion: %v", err)
			return
		}
	}

	// Mark branch as updated to sync the association change
	// Only if branch is not pending (not being created)
	if branch != nil && branch.ID != 0 {
		branchEdit, branchExists := a.editMgr.GetEdit(editstack.EntityBranch, branch.ID)
		if !branchExists || branchEdit.EditType != editstack.CreateBranch {
			// Branch is not pending creation, safe to mark for update
			updateEdit := &editstack.Edit{ID: branch.ID, EditType: editstack.UpdateBranch}
			a.editMgr.AddEdit(updateEdit) // Ignore error - branch might already be marked
		}
	}

	// Find the index of the note in the active note list
	notes := a.dataMgr.GetActiveNoteList()
	for i, n := range notes {
		if n.ID == noteID {
			a.dataMgr.RemoveNote(i)
			break
		}
	}

	a.Synced = false
}
