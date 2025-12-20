package app

import (
	"log"
	"strings"
	"time"

	editstack "github.com/haochend413/ntkpr/internal/app/editStack"
	"github.com/haochend413/ntkpr/internal/models"
)

// CurrentNoteContent returns the content of the current note
func (a *App) CurrentNoteContent() string {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	if a.currentNote == nil {
		return ""
	}
	return a.currentNote.Content
}

// CurrentNoteTopics returns the topics of the current note
func (a *App) CurrentNoteTopics() []*models.Topic {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	if a.currentNote == nil {
		return nil
	}
	return a.currentNote.Topics
}

// CurrentNoteID returns the ID of the current note or -1
func (a *App) CurrentNoteID() int {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	if a.currentNote == nil {
		return -1
	}
	return int(a.currentNote.ID)
}

func (a *App) CurrentNoteLastUpdate() time.Time {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	if a.currentNote == nil {
		return time.Time{}
	}
	return a.currentNote.UpdatedAt
}

func (a *App) CurrentNoteFrequency() int {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	if a.currentNote == nil {
		return 0
	}
	return a.currentNote.Frequency
}

// HasCurrentNote checks if a note is currently selected
func (a *App) HasCurrentNote() bool {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	return a.currentNote != nil
}

// SaveCurrentNote updates the current note content and marks it pending
func (a *App) SaveCurrentNote(content string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	if a.currentNote == nil {
		return
	}
	if a.currentNote.Content == content {
		return
	}
	a.currentNote.Content = content
	a.currentNote.Frequency += 1
	a.currentNote.UpdatedAt = time.Now()
	a.Synced = false
	if err := a.editMgr.AddEdit(editstack.Update, a.currentNote.ID); err != nil {
		log.Printf("Error adding Update edit: %v", err)
	}
}

func (a *App) HighlightCurrentNote() {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	if a.currentNote == nil {
		return
	}
	a.currentNote.Highlight = !a.currentNote.Highlight
	a.currentNote.UpdatedAt = time.Now()
	a.Synced = false
	if err := a.editMgr.AddEdit(editstack.Update, a.currentNote.ID); err != nil {
		log.Printf("Error adding Update edit: %v", err)
	}
}

// AddTopicsToCurrentNote parses a comma-separated list and appends unique topics
func (a *App) AddTopicsToCurrentNote(topicsText string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	if a.currentNote == nil {
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
		topic := &models.Topic{Topic: topicName}
		exists := false
		for _, existing := range a.currentNote.Topics {
			if existing.Topic == topic.Topic {
				exists = true
				break
			}
		}

		if !exists {
			a.currentNote.Topics = append(a.currentNote.Topics, topic)
			changed = true
		}
	}
	if changed {
		a.currentNote.UpdatedAt = time.Now()
		a.Synced = false
		if err := a.editMgr.AddEdit(editstack.Update, a.currentNote.ID); err != nil {
			log.Printf("Error adding Update edit: %v", err)
		}
	}
}

// RemoveTopicFromCurrentNote removes a topic from the current note
func (a *App) RemoveTopicFromCurrentNote(topicToRemove string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	if a.currentNote == nil {
		return
	}

	var newTopics []*models.Topic
	for _, topic := range a.currentNote.Topics {
		if topic.Topic != topicToRemove {
			newTopics = append(newTopics, topic)
		}
	}
	if len(newTopics) == len(a.currentNote.Topics) {
		return
	}
	a.currentNote.Topics = newTopics
	a.currentNote.UpdatedAt = time.Now()
	a.Synced = false
	if err := a.editMgr.AddEdit(editstack.Update, a.currentNote.ID); err != nil {
		log.Printf("Error adding Update edit: %v", err)
	}
}

// DeleteCurrentNote deletes the current note from the active list and marks for deletion if needed
func (a *App) DeleteCurrentNote(cursor uint) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.currentNote == nil {
		return
	}

	noteID := a.currentNote.ID

	// Check if this note is in the Create list (newly created, not yet synced)
	if edit, exists := a.editMgr.EditMap[noteID]; exists && edit.EditType == editstack.Create {
		// Note was created but not synced - just remove it entirely
		a.editMgr.RemoveEdit(noteID)
		delete(a.NotesMap, noteID)
	} else if noteID != 0 {
		// Note exists in DB - mark for deletion
		if err := a.editMgr.AddEdit(editstack.Delete, noteID); err != nil {
			log.Printf("Error adding Delete edit: %v", err)
			return
		}
	}

	// Remove from default context (and it will be reflected in other contexts)
	a.contextMgr.RemoveNoteFromDefault(noteID)
	a.Synced = false

	// adjust cursor
	notes := a.contextMgr.GetCurrentNotes()
	if len(notes) == 0 {
		a.currentNote = nil
		return
	}

	if int(cursor) >= len(notes) {
		cursor = uint(len(notes) - 1)
	}
	a.currentNote = notes[cursor]
	a.contextMgr.SetCurrentCursor(cursor)
}
