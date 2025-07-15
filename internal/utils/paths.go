package utils

import (
	"os"
	"path/filepath"
	"runtime"
)

// GetHomeDir returns the user's home directory across different platforms
func GetHomeDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return homeDir, nil
}

// GetAppDataDir returns the application data directory across different platforms
// Windows: %APPDATA% (typically C:\Users\<username>\AppData\Roaming)
// macOS: ~/Library/Application Support
// Linux: ~/.local/share
func GetAppDataDir() (string, error) {
	homeDir, err := GetHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData != "" {
			return appData, nil
		}
		return filepath.Join(homeDir, "AppData", "Roaming"), nil
	case "darwin":
		return filepath.Join(homeDir, "Library", "Application Support"), nil
	default: // Linux and other Unix-like systems
		return filepath.Join(homeDir, ".local", "share"), nil
	}
}

// GetStoragePath returns the storage.json path across different platforms
// Windows: %APPDATA%/Code/User/globalStorage/storage.json
// macOS: ~/Library/Application Support/Code/User/globalStorage/storage.json
// Linux: ~/.config/Code/User/globalStorage/storage.json
func GetStoragePath() (string, error) {
	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			homeDir, err := GetHomeDir()
			if err != nil {
				return "", err
			}
			appData = filepath.Join(homeDir, "AppData", "Roaming")
		}
		return filepath.Join(appData, "Code", "User", "globalStorage", "storage.json"), nil
	case "darwin":
		homeDir, err := GetHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, "Library", "Application Support", "Code", "User", "globalStorage", "storage.json"), nil
	default: // Linux and other Unix-like systems
		homeDir, err := GetHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, ".config", "Code", "User", "globalStorage", "storage.json"), nil
	}
}

// GetDBPath returns the state.vscdb path across different platforms
// Windows: %APPDATA%/Code/User/globalStorage/state.vscdb
// macOS: ~/Library/Application Support/Code/User/globalStorage/state.vscdb
// Linux: ~/.config/Code/User/globalStorage/state.vscdb
func GetDBPath() (string, error) {
	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			homeDir, err := GetHomeDir()
			if err != nil {
				return "", err
			}
			appData = filepath.Join(homeDir, "AppData", "Roaming")
		}
		return filepath.Join(appData, "Code", "User", "globalStorage", "state.vscdb"), nil
	case "darwin":
		homeDir, err := GetHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, "Library", "Application Support", "Code", "User", "globalStorage", "state.vscdb"), nil
	default: // Linux and other Unix-like systems
		homeDir, err := GetHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, ".config", "Code", "User", "globalStorage", "state.vscdb"), nil
	}
}

// GetMachineIDPath returns the machine ID file path across different platforms
// Windows: %APPDATA%/Code/User/machineid
// macOS: ~/Library/Application Support/Code/machineid
// Linux: ~/.config/Code/User/machineid
func GetMachineIDPath() (string, error) {
	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			homeDir, err := GetHomeDir()
			if err != nil {
				return "", err
			}
			appData = filepath.Join(homeDir, "AppData", "Roaming")
		}
		return filepath.Join(appData, "Code", "User", "machineid"), nil
	case "darwin":
		homeDir, err := GetHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, "Library", "Application Support", "Code", "machineid"), nil
	default: // Linux and other Unix-like systems
		homeDir, err := GetHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, ".config", "Code", "User", "machineid"), nil
	}
}

// GetWorkspaceStoragePath returns the workspaceStorage path across different platforms
// Windows: %APPDATA%/Code/User/workspaceStorage
// macOS: ~/Library/Application Support/Code/User/workspaceStorage
// Linux: ~/.config/Code/User/workspaceStorage
func GetWorkspaceStoragePath() (string, error) {
	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			homeDir, err := GetHomeDir()
			if err != nil {
				return "", err
			}
			appData = filepath.Join(homeDir, "AppData", "Roaming")
		}
		return filepath.Join(appData, "Code", "User", "workspaceStorage"), nil
	case "darwin":
		homeDir, err := GetHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, "Library", "Application Support", "Code", "User", "workspaceStorage"), nil
	default: // Linux and other Unix-like systems
		homeDir, err := GetHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, ".config", "Code", "User", "workspaceStorage"), nil
	}
}
