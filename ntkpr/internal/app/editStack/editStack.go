package editstack

import (
	"fmt"
)

/* This is actuallly an important part. This sets up the overall writing mechanism of the notes. Important! */

/*
So how do we interact with the notes ?
We load from database into context;
When we do stuff, like create / delete / restore / update， we keep the actual content within the note structure that is combined with its id;
Its id, then, is managed by this component.
The current content of each note is in default the newest version compared to database.

Functions here should be triggered when we enter keystrokes, handled by bubbletea update function.
*/

type EditType = int

const (
	None   EditType = -1
	Create EditType = 0
	Update EditType = 1
	Delete EditType = 2
)

// This is only note-wise, not string - wise
// Also there must be a good mechanis around all this.
type Edit struct {
	NoteID   uint // We need the same index generating mechanism as in database.
	EditType EditType
}

type EditMgr struct {
	EditStack []uint // Time Order, keep this only for recent case. Actually not needed for functionality.
	EditMap   map[uint]*Edit
}

// This function sets up edit stack according to basic handling logic.
func (em *EditMgr) AddEdit(tp EditType, id uint) error {
	em.EditStack = append(em.EditStack, id)

	// add to map, be sure of index !
	// check
	if edit, exists := em.EditMap[id]; exists {
		// Key exists, edit contains the value
		prevType := edit.EditType
		switch tp {
		case Create:
			switch prevType {
			case Create:
				return fmt.Errorf("invalid state: attempting to Create note %d that is already marked for Create", id)
			case Update:
				return fmt.Errorf("invalid state: attempting to Create note %d that is already marked for Update", id)
			case Delete:
				return fmt.Errorf("invalid state: attempting to Create note %d that is already marked for Delete", id)
			}
		case Update:
			switch prevType {
			case Create:
				// Keep as Create - new note being edited before sync
				// No change needed, edit.EditType is already Create
			case Update:
				// Already marked as Update, no change needed
			case Delete:
				return fmt.Errorf("invalid state: attempting to Update note %d that is marked for Delete", id)
			}
		case Delete:
			switch prevType {
			case Create:
				// Created then deleted without sync, no DB operation needed
				em.EditMap[id].EditType = None
			case Update:
				// Updated then deleted = need to delete from DB
				em.EditMap[id].EditType = Delete
			case Delete:
				return fmt.Errorf("invalid state: attempting to Delete note %d that is already marked for Delete", id)
			}
		}
	} else {
		// Key doesn't exist, this is a new edit
		em.EditMap[id] = &Edit{NoteID: id, EditType: tp}
	}

	return nil
}

// NewEditMgr creates a new edit manager
func NewEditMgr() *EditMgr {
	return &EditMgr{
		EditStack: make([]uint, 0),
		EditMap:   make(map[uint]*Edit),
	}
}

// Clear resets the edit manager
func (em *EditMgr) Clear() {
	em.EditStack = make([]uint, 0)
	em.EditMap = make(map[uint]*Edit)
}

// RemoveEdit removes an edit from the map (for undo operations)
func (em *EditMgr) RemoveEdit(id uint) {
	delete(em.EditMap, id)
}
