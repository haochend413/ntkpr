package db

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/haochend413/ntkpr/internal/models"
)

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

// export the serialized data into desired position
func (d *DB) ExportNoteToJSON(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	tmp := path + ".tmp"
	var notes []models.Note
	err := d.Conn.Find(&notes).Error
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(notes, "", "  ")
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
