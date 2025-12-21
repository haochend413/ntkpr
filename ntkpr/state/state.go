package state

// This package defines the persisted program states.
// This is even higher level than app.

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/haochend413/ntkpr/internal/app/context"
)

// Default STATE PATH for macOS
var STATE_FILE_PATH string = "~/Library/Application Support/ntkpr/state.json"

type State struct {
	LastContext context.ContextPtr `json:"lastContext"` // previous context
	LastCursor  int                `json:"lastCursor"`  // previous current note
}

// use a function to return different instances. Trick.
func DefaultState() *State {
	return &State{
		LastContext: context.Default,
		LastCursor:  0,
	}
}

func LoadState() (*State, error) {
	path, err := expandPath(STATE_FILE_PATH)
	if err != nil {
		return nil, err
	}

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

func SaveState(s *State) error {
	path, err := expandPath(STATE_FILE_PATH)
	if err != nil {
		return err
	}

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

// expandPath expands ~ to home directory
func expandPath(path string) (string, error) {
	if len(path) == 0 || path[0] != '~' {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	if len(path) == 1 {
		return home, nil
	}

	return filepath.Join(home, path[2:]), nil
}
