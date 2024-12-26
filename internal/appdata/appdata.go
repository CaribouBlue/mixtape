package appdata

import (
	"os"
	"path/filepath"
	"runtime"
)

const AppName = "TopSpot"

func GetAppDataDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	var appDataDir string
	switch runtime := runtime.GOOS; runtime {
	case "windows":
		appDataDir = filepath.Join(os.Getenv("APPDATA"), AppName)
	case "darwin":
		appDataDir = filepath.Join(homeDir, "Library", "Application Support", AppName)
	default: // Unix-like systems
		appDataDir = filepath.Join(homeDir, ".config", AppName)
	}

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(appDataDir, os.ModePerm); err != nil {
		return "", err
	}

	return appDataDir, nil
}
