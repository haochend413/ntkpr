package db

import "github.com/haochend413/ntkpr/internal/models"

func (d *DB) GetFirstNoteID() uint {
	var id uint
	err := d.Conn.Model(&models.Note{}).Select("id").Where("deleted_at IS NULL").Order("id ASC").Limit(1).Scan(&id).Error
	if err != nil {
		return 0
	}
	return id
}

func (d *DB) GetCreateNoteID() uint {
	// Query the database for the maximum ID, including deleted notes
	var maxID uint
	if err := d.Conn.Table("notes").Select("MAX(id)").Row().Scan(&maxID); err != nil {
		maxID = 0
	}
	return maxID + 1
}
