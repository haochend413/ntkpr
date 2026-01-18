/* This is actuallly an important part. This sets up the overall writing mechanism of the notes. Important! */

/*
So how do we interact with the notes ?
We load from database into context;
When we do stuff, like create / delete / restore / update， we keep the actual content within the note structure that is combined with its id;
Its id, then, is managed by this component.
The current content of each note is in default the newest version compared to database.

Functions here should be triggered when we enter keystrokes, handled by bubbletea update function.
*/

/*
Ok. We might need a much more complicated system for thread + branches stuff.
Especially with account of re-arranging notes, not just changing its content.
Maybe we can abstract them in a way.

Operations we can have :

Thread:
Creation / Deletion
Add branch into that thread
Remove branch from that thread

Branch:
Creation / Deletion
Add note into that branch
Remove note from that branch

Note:
Create / Delete / Update

There might be other stuff but I don't know...

Ok here is a new problem :

What about the notes after we remove them from one of the branches ?
Also what happens to the notes after we remove their branch from the thread ?
Maybe we need a "main" branch ? I'm not sure.

Do we support the operation of "SetBranch" ? Append a note to a different branch ? Or init a new branch ?

Let's not support that for now. Keep things simple. Right now, delete branch is equal to delete all its notes.
Same for thread, deleting a thread is deleting all its branches.

There are different types :

What changed ? That's that.


And how do we sync ? What is the sync schedule ?

So we keep what is changed, and it will only be truely helpful if we selectively update. But let's also ignore it for now.
We keep what's changed.

What do we need to specify what has changed: an ID (the ID of a thread / branch / Note), and a type, indicating the type of action.


Actually, note and branch creation / deletion should all be subject to its superior. Thus there is no such thing as "add note to branch"
One note should only belong to a single thread, but there can be multiple branches.
This is the reason why we only have "add note to branch", since it is not equivalent to creating a new note.
*/
/*

Functions here should be triggered when we enter keystrokes, handled by bubbletea update function.
*/

package editstack

import (
	"fmt"
)

type EditType = int

// All possible edit types
const (
	None         EditType = -1
	CreateNote   EditType = 0
	UpdateNote   EditType = 1
	DeleteNote   EditType = 2
	CreateThread EditType = 3
	UpdateThread EditType = 4
	DeleteThread EditType = 6
	CreateBranch EditType = 7 // since branch do not persist across threads, we do not have to introduce add / remove branch
	UpdateBranch EditType = 8
	DeleteBranch EditType = 10
)

// EntityType constants for EditKey
const (
	EntityNote   = "note"
	EntityBranch = "branch"
	EntityThread = "thread"
)

// EditKey is a composite key for EditMap to avoid ID collisions between entity types
type EditKey struct {
	EntityType string // "note", "branch", "thread"
	ID         uint
}

// This is only note-wise, not string - wise
// Also there must be a good mechanis around all this.
type Edit struct {
	ID         uint // We need the same index generating mechanism as in database.
	EditType   EditType
	Additional *uint // optional argument, can be the id of branch / note that are added / removed. Purely for demonstration usage.
}

type EditMgr struct {
	EditStack []*Edit           // Time Order, keep this only for recent case. Actually not needed for functionality.
	EditMap   map[EditKey]*Edit // Keyed by (EntityType, ID) to avoid collisions between notes, branches, and threads
}

// getEntityType returns the entity type string for a given EditType
func getEntityType(tp EditType) string {
	switch tp {
	case CreateNote, UpdateNote, DeleteNote:
		return EntityNote
	case CreateThread, UpdateThread, DeleteThread:
		return EntityThread
	case CreateBranch, UpdateBranch, DeleteBranch:
		return EntityBranch
	default:
		return ""
	}
}

// NewEditMgr creates a new edit manager
func NewEditMgr() *EditMgr {
	return &EditMgr{
		EditStack: make([]*Edit, 0),
		EditMap:   make(map[EditKey]*Edit),
	}
}

// This function sets up edit stack according to basic handling logic.
func (em *EditMgr) AddEdit(curr *Edit) error {
	em.EditStack = append(em.EditStack, curr)
	id := curr.ID
	tp := curr.EditType
	entityType := getEntityType(tp)
	key := EditKey{EntityType: entityType, ID: id}
	// add to map, be sure of index !
	// check
	if edit, exists := em.EditMap[key]; exists {
		// Key exists, edit contains the value
		prevType := edit.EditType
		switch tp {
		case CreateNote:
			switch prevType {
			case CreateNote:
				return fmt.Errorf("invalid state: attempting to CreateNote %d that is already marked for CreateNote", id)
			case UpdateNote:
				return fmt.Errorf("invalid state: attempting to CreateNote %d that is already marked for UpdateNote", id)
			case DeleteNote:
				return fmt.Errorf("invalid state: attempting to CreateNote %d that is already marked for DeleteNote", id)
			}
		case UpdateNote:
			switch prevType {
			case CreateNote:
				// Keep as CreateNote - new note being edited before sync
				// No change needed, edit.EditType is already CreateNote
			case UpdateNote:
				// Already marked as UpdateNote, no change needed
			case DeleteNote:
				return fmt.Errorf("invalid state: attempting to UpdateNote %d that is marked for DeleteNote", id)
			}
		case DeleteNote:
			switch prevType {
			case CreateNote:
				// Created then deleted without sync, no DB operation needed
				em.EditMap[key].EditType = None
			case UpdateNote:
				// Updated then deleted = need to delete from DB
				em.EditMap[key].EditType = DeleteNote
			case DeleteNote:
				return fmt.Errorf("invalid state: attempting to DeleteNote %d that is already marked for DeleteNote", id)
			}

		// Thread operations
		case CreateThread:
			switch prevType {
			case CreateThread:
				return fmt.Errorf("invalid state: attempting to CreateThread %d that is already marked for CreateThread", id)
			case DeleteThread:
				return fmt.Errorf("invalid state: attempting to CreateThread %d that is already marked for DeleteThread", id)
			}
		case UpdateThread:
			switch prevType {
			case CreateThread:
			case UpdateThread:
			case DeleteThread:
				return fmt.Errorf("invalid state: attempting to UpdateNote %d that is marked for DeleteNote", id)
			}
		case DeleteThread:
			switch prevType {
			case CreateThread:
				// Created then deleted without sync, no DB operation needed
				em.EditMap[key].EditType = None
			case UpdateThread:
				em.EditMap[key].EditType = DeleteThread

			case DeleteThread:
				return fmt.Errorf("invalid state: attempting to DeleteThread %d that is already marked for DeleteThread", id)
			}

		// Branch operations
		case CreateBranch:
			switch prevType {
			case CreateBranch:
				return fmt.Errorf("invalid state: attempting to CreateBranch %d that is already marked for CreateBranch", id)
			case UpdateBranch:
				return fmt.Errorf("invalid state: attempting to CreateBranch %d that is already marked for modification", id)
			case DeleteBranch:
				return fmt.Errorf("invalid state: attempting to CreateBranch %d that is already marked for DeleteBranch", id)
			}
		case UpdateBranch:
			switch prevType {
			case CreateBranch:
				// Keep as CreateBranch - new branch being modified before sync
				// No change needed, edit.EditType is already CreateBranch
			case UpdateBranch:
				// Already marked for modification, update to latest operation

			case DeleteBranch:
				return fmt.Errorf("invalid state: attempting to modify branch %d that is marked for DeleteBranch", id)
			}
		case DeleteBranch:
			switch prevType {
			case CreateBranch:
				// Created then deleted without sync, no DB operation needed。 This also should not happen since the ID is going non stop.
				em.EditMap[key].EditType = None
			case UpdateBranch:
				// Modified then deleted = need to delete from DB
				em.EditMap[key].EditType = DeleteBranch
			case DeleteBranch:
				return fmt.Errorf("invalid state: attempting to DeleteBranch %d that is already marked for DeleteBranch", id)
			}
		}
	} else {
		// Key doesn't exist, this is a new edit
		em.EditMap[key] = &Edit{ID: id, EditType: tp}
	}

	return nil
}

// Clear resets the edit manager
func (em *EditMgr) Clear() {
	em.EditStack = make([]*Edit, 0)
	em.EditMap = make(map[EditKey]*Edit)
}

// RemoveEdit removes an edit from the map (for undo operations)
func (em *EditMgr) RemoveEdit(entityType string, id uint) {
	key := EditKey{EntityType: entityType, ID: id}
	delete(em.EditMap, key)
}

// GetEdit retrieves an edit from the map
func (em *EditMgr) GetEdit(entityType string, id uint) (*Edit, bool) {
	key := EditKey{EntityType: entityType, ID: id}
	edit, exists := em.EditMap[key]
	return edit, exists
}
