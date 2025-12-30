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

func BasePathDefault() (string, error) {
	var path string

	switch runtime.GOOS {
	case "darwin":
		path = "~/Library/Application Support/ntkpr/"
	case "linux":
		path = "~/.local/state/ntkpr/"
	case "windows":
		path = "%APPDATA%\\ntkpr\\"
	default:
		fmt.Printf("unsupported OS: %s\n", runtime.GOOS)
		expanded, _ := expandPath(path)
		fmt.Printf("BasePathDefault: OS=%s, path=%s, expanded=%s\n", runtime.GOOS, path, expanded)
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	expanded, err := expandPath(path)
	return expanded, err
}

// stateFilePath points to the state.json file
func StateFilePathDefault() string {
	basePath, _ := BasePathDefault()
	return basePath + "/state.json"
}

// dataFilePathDefault points to the FOLDER that contains all the dbs.
func DataFilePathDefault() string {
	basePath, _ := BasePathDefault()
	return basePath + "/db"
}

func ConfigPath() string {
	basePath, _ := BasePathDefault()
	return basePath + "/config.yaml"
}
