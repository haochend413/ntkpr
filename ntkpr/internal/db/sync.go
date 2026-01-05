package db

import (
	"errors"
	"log"
	"strings"

	editstack "github.com/haochend413/ntkpr/internal/app/editStack"
	"github.com/haochend413/ntkpr/internal/models"
	"gorm.io/gorm"
)

// This needs further testings and debuggings. I have a feeling that this is not entirely correct.

func (d *DB) SyncData(notesMap map[uint]*models.Note,
	threadsMap map[uint]*models.Thread,
	branchesMap map[uint]*models.Branch,
	editMap map[uint]*editstack.Edit) ([]*models.Note, []*models.Topic, []*models.Thread, []*models.Branch, error) {
	// Categorize edits from the editMap
	noteCreateIDs := make([]uint, 0)
	notePendingIDs := make([]uint, 0)
	noteDeleteIDs := make([]uint, 0)

	threadCreateIDs := make([]uint, 0)
	threadDeleteIDs := make([]uint, 0)

	branchCreateIDs := make([]uint, 0)
	branchPendingIDs := make([]uint, 0)
	branchDeleteIDs := make([]uint, 0)

	for id, edit := range editMap {
		switch edit.EditType {
		// Notes
		case editstack.CreateNote:
			noteCreateIDs = append(noteCreateIDs, id)
		case editstack.UpdateNote:
			notePendingIDs = append(notePendingIDs, id)
		case editstack.DeleteNote:
			noteDeleteIDs = append(noteDeleteIDs, id)

		// Threads
		case editstack.CreateThread:
			threadCreateIDs = append(threadCreateIDs, id)
		case editstack.DeleteThread:
			threadDeleteIDs = append(threadDeleteIDs, id)

		// Branches
		case editstack.CreateBranch:
			branchCreateIDs = append(branchCreateIDs, id)
		case editstack.AddNoteToBranch, editstack.RemoveNoteFromBranch:
			branchPendingIDs = append(branchPendingIDs, id)
		case editstack.DeleteBranch:
			branchDeleteIDs = append(branchDeleteIDs, id)

		case editstack.None:
			// Skip
		}
	}

	noteCreateIDs = uniqueIDs(noteCreateIDs)
	notePendingIDs = uniqueIDs(notePendingIDs)
	noteDeleteIDs = uniqueIDs(noteDeleteIDs)
	threadCreateIDs = uniqueIDs(threadCreateIDs)
	threadDeleteIDs = uniqueIDs(threadDeleteIDs)
	branchCreateIDs = uniqueIDs(branchCreateIDs)
	branchPendingIDs = uniqueIDs(branchPendingIDs)
	branchDeleteIDs = uniqueIDs(branchDeleteIDs)

	// Create in order: Threads -> Notes -> Branches
	// (Notes must exist before branches can reference them)

	for _, thread := range collectThreads(threadsMap, threadCreateIDs) {
		if err := d.persistThread(thread, true); err != nil {
			log.Printf("Error creating thread %d: %v", thread.ID, err)
		}
	}

	for _, note := range collectNotes(notesMap, noteCreateIDs) {
		sanitizeNote(note)
		if err := d.persistNote(note, true); err != nil {
			log.Printf("Error creating note %d: %v", note.ID, err)
		}
	}

	for _, branch := range collectBranches(branchesMap, branchCreateIDs) {
		if err := d.persistBranch(branch, true); err != nil {
			log.Printf("Error creating branch %d: %v", branch.ID, err)
		}
	}

	for _, note := range collectNotes(notesMap, notePendingIDs) {
		sanitizeNote(note)
		if err := d.persistNote(note, false); err != nil {
			log.Printf("Error updating note %d: %v", note.ID, err)
		}
	}

	for _, branch := range collectBranches(branchesMap, branchPendingIDs) {
		if err := d.persistBranch(branch, false); err != nil {
			log.Printf("Error updating branch %d: %v", branch.ID, err)
		}
	}

	if err := d.deleteNotes(noteDeleteIDs); err != nil {
		return nil, nil, nil, nil, err
	}

	if err := d.deleteBranches(branchDeleteIDs); err != nil {
		return nil, nil, nil, nil, err
	}

	if err := d.deleteThreads(threadDeleteIDs); err != nil {
		return nil, nil, nil, nil, err
	}

	return d.loadAll()
}

func (d *DB) persistNote(note *models.Note, isCreate bool) error {
	if note == nil {
		return nil
	}
	topics, err := d.ensureTopics(note.Topics)
	if err != nil {
		return err
	}
	note.Topics = topics

	var result *gorm.DB
	if isCreate {
		result = d.Conn.Create(note)
	} else {
		result = d.Conn.Save(note)
	}
	if result.Error != nil {
		return result.Error
	}

	return d.Conn.Model(note).Association("Topics").Replace(note.Topics)
}

func (d *DB) ensureTopics(topics []*models.Topic) ([]*models.Topic, error) {
	normalized := make([]*models.Topic, 0, len(topics))
	seen := make(map[string]struct{}, len(topics))
	for _, topic := range topics {
		if topic == nil {
			continue
		}
		name := strings.ToLower(strings.TrimSpace(topic.Topic))
		if name == "" {
			continue
		}
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}

		existing := &models.Topic{}
		if err := d.Conn.Where("topic = ?", name).First(existing).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				topic.Topic = name
				if err := d.Conn.Create(topic).Error; err != nil {
					return nil, err
				}
				normalized = append(normalized, topic)
				continue
			}
			return nil, err
		}
		normalized = append(normalized, existing)
	}
	return normalized, nil
}

func (d *DB) deleteNotes(ids []uint) error {
	for _, id := range ids {
		if err := d.Conn.Delete(&models.Note{}, id).Error; err != nil {
			return err
		}
	}
	return nil
}

func sanitizeNote(note *models.Note) {
	if note == nil {
		return
	}
	note.Content = strings.TrimSpace(note.Content)
}

func collectNotes(notesMap map[uint]*models.Note, ids []uint) []*models.Note {
	if len(ids) == 0 || notesMap == nil {
		return nil
	}
	notes := make([]*models.Note, 0, len(ids))
	for _, id := range ids {
		if note, exists := notesMap[id]; exists && note != nil {
			notes = append(notes, note)
		}
	}
	return notes
}

func (d *DB) persistThread(thread *models.Thread, isCreate bool) error {
	if thread == nil {
		return nil
	}

	var result *gorm.DB
	if isCreate {
		result = d.Conn.Create(thread)
	} else {
		result = d.Conn.Save(thread)
	}
	if result.Error != nil {
		return result.Error
	}

	// Handle one-to-many: Thread -> Branches
	return d.Conn.Model(thread).Association("Branches").Replace(thread.Branches)
}

func collectThreads(threadsMap map[uint]*models.Thread, ids []uint) []*models.Thread {
	if len(ids) == 0 || threadsMap == nil {
		return nil
	}
	threads := make([]*models.Thread, 0, len(ids))
	for _, id := range ids {
		if t, exists := threadsMap[id]; exists && t != nil {
			threads = append(threads, t)
		}
	}
	return threads
}

func (d *DB) deleteThreads(ids []uint) error {
	for _, id := range ids {
		// Cascading delete will handle branches due to foreign key
		// This will automatically delete the related branches.
		if err := d.Conn.Delete(&models.Thread{}, id).Error; err != nil {
			return err
		}
	}
	return nil
}

func (d *DB) persistBranch(branch *models.Branch, isCreate bool) error {
	if branch == nil {
		return nil
	}

	var result *gorm.DB
	if isCreate {
		result = d.Conn.Create(branch)
	} else {
		result = d.Conn.Save(branch)
	}
	if result.Error != nil {
		return result.Error
	}

	// Handle many-to-many: Branch <-> Notes
	return d.Conn.Model(branch).Association("Notes").Replace(branch.Notes)
}

func (d *DB) deleteBranches(ids []uint) error {
	for _, id := range ids {
		if err := d.Conn.Delete(&models.Branch{}, id).Error; err != nil {
			return err
		}
	}
	return nil
}

func collectBranches(branchesMap map[uint]*models.Branch, ids []uint) []*models.Branch {
	if len(ids) == 0 || branchesMap == nil {
		return nil
	}
	branches := make([]*models.Branch, 0, len(ids))
	for _, id := range ids {
		if branch, exists := branchesMap[id]; exists && branch != nil {
			branches = append(branches, branch)
		}
	}
	return branches
}

func (d *DB) loadAll() ([]*models.Note, []*models.Topic, []*models.Thread, []*models.Branch, error) {
	var dbNotes []*models.Note
	if err := d.Conn.Preload("Topics").Preload("Branches").Order("created_at ASC").Find(&dbNotes).Error; err != nil {
		return nil, nil, nil, nil, err
	}

	var dbTopics []*models.Topic
	if err := d.Conn.Preload("Notes").Order("topic ASC").Find(&dbTopics).Error; err != nil {
		return nil, nil, nil, nil, err
	}

	var dbThreads []*models.Thread
	if err := d.Conn.Preload("Branches").Order("created_at ASC").Find(&dbThreads).Error; err != nil {
		return nil, nil, nil, nil, err
	}

	var dbBranches []*models.Branch
	if err := d.Conn.Preload("Notes").Order("created_at ASC").Find(&dbBranches).Error; err != nil {
		return nil, nil, nil, nil, err
	}

	return dbNotes, dbTopics, dbThreads, dbBranches, nil
}

func uniqueIDs(ids []uint) []uint {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[uint]struct{}, len(ids))
	result := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}
