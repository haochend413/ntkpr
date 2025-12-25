package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Default STATE PATH for macOS
// expandPath expands ~ to home directory
// expandPath expands ~ and %APPDATA%
func expandPath(path string) (string, error) {
	// Windows env vars
	if strings.Contains(path, "%APPDATA%") {
		path = strings.ReplaceAll(path, "%APPDATA%", os.Getenv("APPDATA"))
	}

	// Unix ~
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if path == "~" {
			return home, nil
		}
		path = filepath.Join(home, path[2:])
	}

	return path, nil
}

// stateFilePath points to the state.json file
func stateFilePathDefault() (string, error) {
	var path string

	switch runtime.GOOS {
	case "darwin":
		path = "~/Library/Application Support/ntkpr/state.json"
	case "linux":
		path = "~/.local/state/ntkpr/state.json"
	case "windows":
		path = "%APPDATA%\\ntkpr\\state.json"
	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	return expandPath(path)
}

// dataFilePathDefault points to the FOLDER that contains all the dbs.
func dataFilePathDefault() (string, error) {
	var path string

	switch runtime.GOOS {
	case "darwin":
		path = "~/Library/Application Support/ntkpr/db"
	case "linux":
		path = "~/.local/share/ntkpr/db"
	case "windows":
		path = "%APPDATA%\\ntkpr\\db"
	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	return expandPath(path)
}

func ConfigPath() (string, error) {
	var path string

	switch runtime.GOOS {
	case "darwin":
		path = "~/Library/Application Support/ntkpr/config.yaml"
	case "linux":
		path = "~/.config/ntkpr/config.yaml"
	case "windows":
		path = "%APPDATA%\\ntkpr\\config.yaml"
	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	return expandPath(path)
}
