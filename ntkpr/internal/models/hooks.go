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

// BeforeSave is a GORM hook that ensures LastEdit is set to CreatedAt if it is null.
func (n *Note) BeforeSave(tx *gorm.DB) (err error) {
	if n.LastEdit.IsZero() {
		n.LastEdit = n.CreatedAt
	}
	return nil
}

// BeforeSave is a GORM hook for Branch that ensures LastEdit is set to CreatedAt if it is null.
func (b *Branch) BeforeSave(tx *gorm.DB) (err error) {
	if b.LastEdit.IsZero() {
		b.LastEdit = b.CreatedAt
	}
	return nil
}

// BeforeSave is a GORM hook for Thread that ensures LastEdit is set to CreatedAt if it is null.
func (t *Thread) BeforeSave(tx *gorm.DB) (err error) {
	if t.LastEdit.IsZero() {
		t.LastEdit = t.CreatedAt
	}
	return nil
}
