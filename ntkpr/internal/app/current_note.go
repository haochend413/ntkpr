package app

import (
	"log"
	"strings"
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
func (a *App) GetCurrentNoteTopics() []*models.Topic {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	note := a.getCurrentNote()
	if note == nil {
		return nil
	}

	// Return a copy to prevent external mutation
	topics := make([]*models.Topic, len(note.Topics))
	copy(topics, note.Topics)
	return topics
}

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
	note.UpdatedAt = time.Now()
	a.Synced = false

	edit := &editstack.Edit{ID: note.ID, EditType: editstack.UpdateNote}
	if err := a.editMgr.AddEdit(edit); err != nil {
		log.Printf("Error tracking note update: %v", err)
	}
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

// AddTopicsToCurrentNote parses a comma-separated list and adds unique topics to the current note
func (a *App) AddTopicsToCurrentNote(topicsText string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	note := a.getCurrentNote()
	if note == nil {
		return
	}

	topicsText = strings.ToLower(strings.TrimSpace(topicsText))
	if topicsText == "" {
		return
	}

	changed := false
	topicNames := strings.Split(topicsText, ",")

	for _, topicName := range topicNames {
		topicName = strings.TrimSpace(topicName)
		if topicName == "" {
			continue
		}

		// Check if topic already exists
		exists := false
		for _, existing := range note.Topics {
			if existing.Topic == topicName {
				exists = true
				break
			}
		}

		if !exists {
			topic := &models.Topic{Topic: topicName}
			note.Topics = append(note.Topics, topic)
			changed = true
		}
	}

	if changed {
		note.UpdatedAt = time.Now()
		a.Synced = false

		edit := &editstack.Edit{ID: note.ID, EditType: editstack.UpdateNote}
		if err := a.editMgr.AddEdit(edit); err != nil {
			log.Printf("Error tracking note update: %v", err)
		}
	}
}

// RemoveTopicFromCurrentNote removes a specific topic from the current note
func (a *App) RemoveTopicFromCurrentNote(topicToRemove string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	note := a.getCurrentNote()
	if note == nil {
		return
	}

	newTopics := make([]*models.Topic, 0, len(note.Topics))
	for _, topic := range note.Topics {
		if topic.Topic != topicToRemove {
			newTopics = append(newTopics, topic)
		}
	}

	// No-op if nothing was removed
	if len(newTopics) == len(note.Topics) {
		return
	}

	note.Topics = newTopics
	note.UpdatedAt = time.Now()
	a.Synced = false

	edit := &editstack.Edit{ID: note.ID, EditType: editstack.UpdateNote}
	if err := a.editMgr.AddEdit(edit); err != nil {
		log.Printf("Error tracking note update: %v", err)
	}
}

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
