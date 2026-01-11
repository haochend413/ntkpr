package state

// This package defines the persisted program states.
// This is even higher level than app.

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/haochend413/ntkpr/internal/app/context"
)

type UIState struct {
	LastThreadContext context.ContextPtr         `json:"lastThreadContext"` // previous context
	LastBranchContext context.ContextPtr         `json:"lastBranchContext"`
	LastNoteContext   context.ContextPtr         `json:"lastNoteContext"`
	YOffsets_Thread   map[context.ContextPtr]int `json:"yOffsets_thread"` // viewport scroll offsets per context
	YOffsets_Branch   map[context.ContextPtr]int `json:"yOffsets_branch"`
	YOffsets_Note     map[context.ContextPtr]int `json:"yOffsets_note"`
}

type AppState struct {
	ThreadCursors map[context.ContextPtr]uint `json:"thread_cursors"` // cursor positions per context
	BranchCursors map[context.ContextPtr]uint `json:"branch_cursors"`
	NoteCursors   map[context.ContextPtr]uint `json:"note_cursors"`
}

type State struct {
	UI  UIState  `json:"ui"`
	App AppState `json:"app"`
}

// use a function to return different instances. Trick.
func DefaultState() *State {
	return &State{
		UI: UIState{
			LastThreadContext: context.Default,
			LastBranchContext: context.Default,
			LastNoteContext:   context.Default,
			YOffsets_Thread: map[context.ContextPtr]int{
				context.Default: 0,
				context.Recent:  0,
				context.Search:  0,
			},
			YOffsets_Branch: map[context.ContextPtr]int{
				context.Default: 0,
				context.Recent:  0,
				context.Search:  0,
			},
			YOffsets_Note: map[context.ContextPtr]int{
				context.Default: 0,
				context.Recent:  0,
				context.Search:  0,
			},
		},
		App: AppState{
			ThreadCursors: map[context.ContextPtr]uint{
				context.Default: 0,
				context.Recent:  0,
				context.Search:  0,
			},
			BranchCursors: map[context.ContextPtr]uint{
				context.Default: 0,
				context.Recent:  0,
				context.Search:  0,
			},
			NoteCursors: map[context.ContextPtr]uint{
				context.Default: 0,
				context.Recent:  0,
				context.Search:  0,
			},
		},
	}
}

func LoadState(path string) (*State, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return DefaultState(), nil
	}
	if err != nil {
		return nil, err
	}

	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}

	return &s, nil
}

func SaveState(path string, s *State) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	tmp := path + ".tmp"

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}

	return os.Rename(tmp, path)
}
