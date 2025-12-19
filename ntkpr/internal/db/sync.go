package db

import (
	"errors"
	"log"
	"strings"

	"github.com/haochend413/ntkpr/internal/models"
	"gorm.io/gorm"
)

func (d *DB) SyncData(notesMap map[uint]*models.Note, pendingNoteIDs []uint, deletedNoteIDs []uint, createNoteIDs []uint) ([]*models.Note, []*models.Topic, error) {
	createIDs := uniqueIDs(createNoteIDs)
	pendingIDs := uniqueIDs(pendingNoteIDs)
	deletedIDs := uniqueIDs(deletedNoteIDs)

	createLookup := make(map[uint]struct{}, len(createIDs))
	for _, id := range createIDs {
		createLookup[id] = struct{}{}
	}
	pendingIDs = filterIDs(pendingIDs, createLookup)

	for _, note := range collectNotes(notesMap, createIDs) {
		sanitizeNote(note)
		if err := d.persistNote(note, true); err != nil {
			log.Printf("Error creating note %d: %v", note.ID, err)
		}
	}

	for _, note := range collectNotes(notesMap, pendingIDs) {
		sanitizeNote(note)
		if err := d.persistNote(note, false); err != nil {
			log.Printf("Error updating note %d: %v", note.ID, err)
		}
	}

	if err := d.deleteNotes(deletedIDs); err != nil {
		return nil, nil, err
	}

	return d.loadAllNotesAndTopics()
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

func (d *DB) loadAllNotesAndTopics() ([]*models.Note, []*models.Topic, error) {
	var dbNotes []*models.Note
	if err := d.Conn.Preload("Topics").Order("created_at ASC").Find(&dbNotes).Error; err != nil {
		return nil, nil, err
	}

	var dbTopics []*models.Topic
	if err := d.Conn.Preload("Notes").Order("topic ASC").Find(&dbTopics).Error; err != nil {
		return nil, nil, err
	}

	return dbNotes, dbTopics, nil
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

func filterIDs(ids []uint, skip map[uint]struct{}) []uint {
	if len(ids) == 0 {
		return nil
	}
	filtered := make([]uint, 0, len(ids))
	for _, id := range ids {
		if _, exists := skip[id]; exists {
			continue
		}
		filtered = append(filtered, id)
	}
	return filtered
}
