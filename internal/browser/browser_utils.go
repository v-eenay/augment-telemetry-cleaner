package browser

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"augment-telemetry-cleaner/internal/utils"
)

// cleanSafariBrowser cleans Safari browser data (macOS only)
func (bc *BrowserCleaner) cleanSafariBrowser(profile BrowserProfile, result *BrowserCleanResult) {
	// Clean cookies
	cookiesFile := filepath.Join(profile.ProfilePath, "Cookies", "Cookies.binarycookies")
	if _, err := os.Stat(cookiesFile); err == nil {
		// Safari uses binary cookies format, which is complex to parse
		// For now, we'll skip direct cookie cleaning and recommend manual clearing
		result.Errors = append(result.Errors, "Safari cookie cleaning requires manual intervention")
	}
	
	// Clean local storage
	localStorageDir := filepath.Join(profile.ProfilePath, "LocalStorage")
	if _, err := os.Stat(localStorageDir); err == nil {
		deleted, err := bc.cleanSafariStorage(localStorageDir)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to clean storage: %v", err))
		} else {
			result.StorageDeleted = deleted
		}
	}
	
	// Clean cache
	cacheDir := filepath.Join(profile.ProfilePath, "Cache.db")
	if _, err := os.Stat(cacheDir); err == nil {
		// Safari cache is also in a proprietary format
		result.Errors = append(result.Errors, "Safari cache cleaning requires manual intervention")
	}
}

// cleanSafariStorage cleans Augment-related storage from Safari
func (bc *BrowserCleaner) cleanSafariStorage(storageDir string) (int64, error) {
	var deleted int64
	
	err := filepath.Walk(storageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() && strings.Contains(strings.ToLower(info.Name()), "augment") {
			if err := os.Remove(path); err == nil {
				deleted++
			}
		}
		
		return nil
	})
	
	return deleted, err
}

// containsAugmentData checks if a file contains Augment-related data
func (bc *BrowserCleaner) containsAugmentData(filePath string) bool {
	// This is a simplified implementation
	// In practice, you might want to scan file contents for Augment patterns
	fileName := strings.ToLower(filepath.Base(filePath))
	return strings.Contains(fileName, "augment")
}

// countAugmentData counts Augment-related data in a browser profile
func (bc *BrowserCleaner) countAugmentData(profile BrowserProfile) int64 {
	var count int64
	
	switch profile.Type {
	case Chrome, Edge:
		count += bc.countChromiumData(profile)
	case Firefox:
		count += bc.countFirefoxData(profile)
	case Safari:
		count += bc.countSafariData(profile)
	}
	
	return count
}

// countChromiumData counts Augment data in Chromium browsers
func (bc *BrowserCleaner) countChromiumData(profile BrowserProfile) int64 {
	var count int64
	
	// Count cookies
	cookiesDB := filepath.Join(profile.ProfilePath, "Cookies")
	if _, err := os.Stat(cookiesDB); err == nil {
		if db, err := sql.Open("sqlite3", cookiesDB); err == nil {
			defer db.Close()
			var cookieCount int64
			query := `SELECT COUNT(*) FROM cookies WHERE host_key LIKE '%augment%' OR name LIKE '%augment%'`
			if err := db.QueryRow(query).Scan(&cookieCount); err == nil {
				count += cookieCount
			}
		}
	}
	
	// Count storage files
	storageDir := filepath.Join(profile.ProfilePath, "Local Storage", "leveldb")
	if _, err := os.Stat(storageDir); err == nil {
		filepath.Walk(storageDir, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() && strings.Contains(strings.ToLower(info.Name()), "augment") {
				count++
			}
			return nil
		})
	}
	
	return count
}

// countFirefoxData counts Augment data in Firefox
func (bc *BrowserCleaner) countFirefoxData(profile BrowserProfile) int64 {
	var count int64
	
	// Count cookies
	cookiesDB := filepath.Join(profile.ProfilePath, "cookies.sqlite")
	if _, err := os.Stat(cookiesDB); err == nil {
		if db, err := sql.Open("sqlite3", cookiesDB); err == nil {
			defer db.Close()
			var cookieCount int64
			query := `SELECT COUNT(*) FROM moz_cookies WHERE host LIKE '%augment%' OR name LIKE '%augment%'`
			if err := db.QueryRow(query).Scan(&cookieCount); err == nil {
				count += cookieCount
			}
		}
	}
	
	// Count storage directories
	storageDir := filepath.Join(profile.ProfilePath, "storage", "default")
	if _, err := os.Stat(storageDir); err == nil {
		filepath.Walk(storageDir, func(path string, info os.FileInfo, err error) error {
			if err == nil && info.IsDir() && strings.Contains(strings.ToLower(info.Name()), "augment") {
				count++
			}
			return nil
		})
	}
	
	return count
}

// countSafariData counts Augment data in Safari
func (bc *BrowserCleaner) countSafariData(profile BrowserProfile) int64 {
	var count int64
	
	// Count storage files
	storageDir := filepath.Join(profile.ProfilePath, "LocalStorage")
	if _, err := os.Stat(storageDir); err == nil {
		filepath.Walk(storageDir, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() && strings.Contains(strings.ToLower(info.Name()), "augment") {
				count++
			}
			return nil
		})
	}
	
	return count
}

// createProfileBackup creates a backup of the browser profile
func (bc *BrowserCleaner) createProfileBackup(profile BrowserProfile) (string, error) {
	timestamp := time.Now().Unix()
	backupName := fmt.Sprintf("%s-backup-%d", 
		strings.ReplaceAll(strings.ToLower(profile.Name), " ", "-"), 
		timestamp)
	
	// Use the same backup directory as other components
	backupDir := filepath.Join("backups", "browser-data")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}
	
	backupPath := filepath.Join(backupDir, backupName)
	
	// Create a simple backup by copying critical files
	criticalFiles := bc.getCriticalFiles(profile)
	
	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create profile backup directory: %w", err)
	}
	
	for _, file := range criticalFiles {
		if _, err := os.Stat(file); err == nil {
			destFile := filepath.Join(backupPath, filepath.Base(file))
			if err := utils.CopyFile(file, destFile); err != nil {
				// Log error but continue with other files
				continue
			}
		}
	}
	
	return backupPath, nil
}

// getCriticalFiles returns a list of critical files to backup for a browser profile
func (bc *BrowserCleaner) getCriticalFiles(profile BrowserProfile) []string {
	var files []string
	
	switch profile.Type {
	case Chrome, Edge:
		files = []string{
			filepath.Join(profile.ProfilePath, "Cookies"),
			filepath.Join(profile.ProfilePath, "Preferences"),
			filepath.Join(profile.ProfilePath, "Local State"),
		}
	case Firefox:
		files = []string{
			filepath.Join(profile.ProfilePath, "cookies.sqlite"),
			filepath.Join(profile.ProfilePath, "prefs.js"),
			filepath.Join(profile.ProfilePath, "places.sqlite"),
		}
	case Safari:
		files = []string{
			filepath.Join(profile.ProfilePath, "Cookies", "Cookies.binarycookies"),
			filepath.Join(profile.ProfilePath, "Preferences.plist"),
		}
	}
	
	return files
}

// IsBrowserRunning checks if a browser process is currently running
func IsBrowserRunning(browserType BrowserType) (bool, error) {
	var processNames []string

	switch browserType {
	case Chrome:
		processNames = []string{"chrome", "google chrome", "googlechrome"}
	case Edge:
		processNames = []string{"msedge", "microsoft edge"}
	case Firefox:
		processNames = []string{"firefox", "mozilla firefox"}
	case Safari:
		processNames = []string{"safari"}
	}

	switch runtime.GOOS {
	case "windows":
		return checkWindowsProcesses(processNames)
	case "darwin":
		return checkMacProcesses(processNames)
	case "linux":
		return checkLinuxProcesses(processNames)
	default:
		return false, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// checkWindowsProcesses checks if processes are running on Windows
func checkWindowsProcesses(processNames []string) (bool, error) {
	cmd := exec.Command("tasklist", "/fo", "csv", "/nh")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to execute tasklist: %w", err)
	}

	outputStr := strings.ToLower(string(output))
	for _, name := range processNames {
		if strings.Contains(outputStr, strings.ToLower(name)) {
			return true, nil
		}
	}

	return false, nil
}

// checkMacProcesses checks if processes are running on macOS
func checkMacProcesses(processNames []string) (bool, error) {
	cmd := exec.Command("ps", "-A")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to execute ps: %w", err)
	}

	outputStr := strings.ToLower(string(output))
	for _, name := range processNames {
		if strings.Contains(outputStr, strings.ToLower(name)) {
			return true, nil
		}
	}

	return false, nil
}

// checkLinuxProcesses checks if processes are running on Linux
func checkLinuxProcesses(processNames []string) (bool, error) {
	cmd := exec.Command("ps", "-A")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to execute ps: %w", err)
	}

	outputStr := strings.ToLower(string(output))
	for _, name := range processNames {
		if strings.Contains(outputStr, strings.ToLower(name)) {
			return true, nil
		}
	}

	return false, nil
}
