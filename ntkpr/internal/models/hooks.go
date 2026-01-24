package models

import (
	"time"

	"gorm.io/gorm"
)

func (n *Note) AfterUpdate(tx *gorm.DB) error {
	// Touch thread
	if err := tx.Model(&Thread{}).
		Where("id = ?", n.ThreadID).
		Update("updated_at", time.Now()).
		Error; err != nil {
		return err
	}

	// Touch branches (SQLite-safe)
	return tx.
		Table("branches").
		Where("id IN (?)",
			tx.Table("branch_notes").
				Select("branch_id").
				Where("note_id = ?", n.ID),
		).
		Update("updated_at", time.Now()).
		Error
}

func (b *Branch) AfterUpdate(tx *gorm.DB) error {
	return tx.Model(&Thread{}).
		Where("id = ?", b.ThreadID).
		Update("updated_at", time.Now()).
		Error
}
