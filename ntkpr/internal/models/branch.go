package models

import "gorm.io/gorm"

/*
Each branch contains its own notes, arranged in time order.
Notes should be able to co-exist in several branches.
Branches are managed by Threads.
*/
type Branch struct {
	gorm.Model      // This contains ID.
	ThreadID   uint // Foreign key for Thread.
	Name       string
	Summary    string
	Highlight  bool    `gorm:"default:false"`
	Private    bool    `gorm:"default:false"`
	Notes      []*Note `gorm:"many2many:branch_notes;constraint:OnDelete:CASCADE;"` // Maybe we can improve it ? Let's first keep it this way.
	Frequency  int     `gorm:"not null;default:0"`
}
