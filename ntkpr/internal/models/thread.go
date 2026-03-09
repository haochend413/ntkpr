package models

import (
	"time"

	"gorm.io/gorm"
)

/*
Thread is the highest level unit of managements.
Each thread contains many notes that could be separated to several branches.
*/

type Thread struct {
	gorm.Model
	Name      string
	Summary   string
	LastEdit  time.Time
	Highlight bool `gorm:"default:false"`
	Private   bool `gorm:"default:false"`
	Branches  []*Branch
	Frequency int `gorm:"not null;default:0"`
}
