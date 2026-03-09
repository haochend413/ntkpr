package models

// we will use superlink to jump directly between notes.
// this can also be used to directly jump between notes! great idea.

type Superlink struct {
	ThreadID int
	BranchID int
	NoteID   int
}
