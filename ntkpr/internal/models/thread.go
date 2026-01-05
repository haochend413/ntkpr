package models

import "gorm.io/gorm"

/*
Thread is the highest level unit of managements.
Each thread contains many notes that could be separated to several branches.
*/

type Thread struct {
	gorm.Model
	Name        string
	Highlight   bool `gorm:"default:false"`
	Private     bool `gorm:"default:false"`
	BranchCount int  `gorm:"default:0"`
	NoteCount   int  `gorm:"default:0"`
	Branches    []*Branch
}

/*
Each branch contains its own notes, arranged in time order.
Notes should be able to co-exist in several branches.
Branches are managed by Threads.
*/
type Branch struct {
	gorm.Model      // This contains ID.
	ThreadID   uint // Foreign key for Thread.
	Name       string
	Highlight  bool    `gorm:"default:false"`
	Private    bool    `gorm:"default:false"`
	NoteCount  int     `gorm:"default:0"`
	Notes      []*Note `gorm:"many2many:branch_notes;constraint:OnDelete:CASCADE;"`
}
