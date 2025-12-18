package app

import (
	"slices"
	"strings"
	"time"

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
	a.PendingNoteIDs = addUniqueID(a.PendingNoteIDs, a.currentNote.ID)
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
	a.PendingNoteIDs = addUniqueID(a.PendingNoteIDs, a.currentNote.ID)
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
		a.PendingNoteIDs = addUniqueID(a.PendingNoteIDs, a.currentNote.ID)
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
	a.PendingNoteIDs = addUniqueID(a.PendingNoteIDs, a.currentNote.ID)
}

// DeleteCurrentNote deletes the current note from the active list and marks for deletion if needed
func (a *App) DeleteCurrentNote(cursor uint) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.currentNote == nil {
		return
	}

	noteID := a.currentNote.ID
	isInCreateList := slices.Contains(a.CreateNoteIDs, noteID)

	if isInCreateList {
		// Remove from CreateNoteIDs
		for i, id := range a.CreateNoteIDs {
			if id == noteID {
				a.CreateNoteIDs = append(a.CreateNoteIDs[:i], a.CreateNoteIDs[i+1:]...)
				break
			}
		}
		// Remove from NotesMap for newly created notes
		delete(a.NotesMap, noteID)
	} else if noteID != 0 {
		a.DeletedNoteIDs = append(a.DeletedNoteIDs, noteID)
	}

	// Remove from NotesList
	for i, note := range a.NotesList {
		if note.ID == noteID {
			a.NotesList = append(a.NotesList[:i], a.NotesList[i+1:]...)
			break
		}
	}

	// Remove from current list if it's different from NotesList
	if a.CurrentNotesListPtr != &a.NotesList {
		for i, note := range *a.CurrentNotesListPtr {
			if note.ID == noteID {
				*a.CurrentNotesListPtr = append((*a.CurrentNotesListPtr)[:i], (*a.CurrentNotesListPtr)[i+1:]...)
				break
			}
		}
	}
	a.Synced = false

	// adjust cursor 
	if len(*a.CurrentNotesListPtr) == 0 {
		a.currentNote = nil
		a.Synced = false
		return
	}

	if int(cursor) >= len(*a.CurrentNotesListPtr) {
		cursor = uint(len(*a.CurrentNotesListPtr) - 1)
	}
	a.currentNote = (*a.CurrentNotesListPtr)[cursor]
}

func addUniqueID(ids []uint, id uint) []uint {
	if id == 0 {
		return ids
	}
	for _, existing := range ids {
		if existing == id {
			return ids
		}
	}
	return append(ids, id)
}
