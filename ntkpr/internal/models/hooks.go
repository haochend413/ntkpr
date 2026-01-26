package models

import (
	"time"

	"gorm.io/gorm"
)

func (n *Note) AfterUpdate(tx *gorm.DB) error {
	// Touch thread: bump updated_at and increment frequency by 1
	if err := tx.Model(&Thread{}).
		Where("id = ?", n.ThreadID).
		Updates(map[string]interface{}{
			"updated_at": time.Now(),
		}).Error; err != nil {
		return err
	}

	// Touch branches (SQLite-safe): bump updated_at and increment frequency
	sub := tx.Table("branch_notes").Select("branch_id").Where("note_id = ?", n.ID)
	if err := tx.Model(&Branch{}).
		Where("id IN (?)", sub).
		Updates(map[string]interface{}{
			"updated_at": time.Now(),
		}).Error; err != nil {
		return err
	}

	return nil
}

func (b *Branch) AfterUpdate(tx *gorm.DB) error {
	return tx.Model(&Thread{}).
		Where("id = ?", b.ThreadID).
		Updates(map[string]interface{}{
			"updated_at": time.Now(),
		}).Error
}
