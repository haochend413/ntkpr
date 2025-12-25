package models

import "gorm.io/gorm"

// Note represents a note entity
type Note struct {
	gorm.Model
	Content   string
	Highlight bool     `gorm:"default:false"`
	Private   bool     `gorm:"default:false"`
	Frequency int      // calculated as the number of times that is edited
	Topics    []*Topic `gorm:"many2many:note_topics;constraint:OnDelete:CASCADE;"`
}
