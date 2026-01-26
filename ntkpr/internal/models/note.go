package models

import "gorm.io/gorm"

// Note represents a note entity
type Note struct {
	gorm.Model
	Content   string
	Highlight bool      `gorm:"default:false"`
	Private   bool      `gorm:"default:false"`
	Frequency int       `gorm:"not null;default:0"`
	Branches  []*Branch `gorm:"many2many:branch_notes;constraint:OnDelete:CASCADE;"`
	ThreadID  uint      // Foreign key - note belongs to a single thread
}
