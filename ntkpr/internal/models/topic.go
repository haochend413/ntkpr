package models

import "gorm.io/gorm"

type Topic struct {
	gorm.Model
	Topic string
	Notes []*Note `gorm:"many2many:note_topics;constraint:OnDelete:CASCADE;"`
}
