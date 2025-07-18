package browser

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// BrowserCleanResult contains the results of browser cleaning operation
type BrowserCleanResult struct {
	Profile         BrowserProfile `json:"profile"`
	BackupPath      string         `json:"backup_path,omitempty"`
	CookiesDeleted  int64          `json:"cookies_deleted"`
	StorageDeleted  int64          `json:"storage_deleted"`
	CacheDeleted    int64          `json:"cache_deleted"`
	FilesDeleted    []string       `json:"files_deleted"`
	Errors          []string       `json:"errors,omitempty"`
}

// BrowserCleaner handles cleaning of browser data
type BrowserCleaner struct {
	detector *BrowserDetector
}

// NewBrowserCleaner creates a new browser cleaner
func NewBrowserCleaner() (*BrowserCleaner, error) {
	detector, err := NewBrowserDetector()
	if err != nil {
		return nil, fmt.Errorf("failed to create browser detector: %w", err)
	}
	
	return &BrowserCleaner{
		detector: detector,
	}, nil
}

// CleanBrowserData cleans Augment-related data from all detected browsers
func (bc *BrowserCleaner) CleanBrowserData(createBackup bool) ([]BrowserCleanResult, error) {
	profiles, err := bc.detector.DetectBrowsers()
	if err != nil {
		return nil, fmt.Errorf("failed to detect browsers: %w", err)
	}
	
	var results []BrowserCleanResult
	
	for _, profile := range profiles {
		// Check if browser is running
		isRunning, err := bc.detector.IsProcessRunning(profile.Type)
		if err != nil {
			result := BrowserCleanResult{
				Profile: profile,
				Errors:  []string{fmt.Sprintf("Failed to check if browser is running: %v", err)},
			}
			results = append(results, result)
			continue
		}
		
		if isRunning {
			result := BrowserCleanResult{
				Profile: profile,
				Errors:  []string{fmt.Sprintf("%s is currently running. Please close it before cleaning.", profile.Type.String())},
			}
			results = append(results, result)
			continue
		}
		
		// Clean the profile
		result := bc.cleanProfile(profile, createBackup)
		results = append(results, result)
	}
	
	return results, nil
}

// GetBrowserDataCount returns the count of Augment-related data in browsers (for dry-run)
func (bc *BrowserCleaner) GetBrowserDataCount() (map[string]int64, error) {
	profiles, err := bc.detector.DetectBrowsers()
	if err != nil {
		return nil, fmt.Errorf("failed to detect browsers: %w", err)
	}
	
	counts := make(map[string]int64)
	
	for _, profile := range profiles {
		count := bc.countAugmentData(profile)
		if count > 0 {
			counts[profile.Name] = count
		}
	}
	
	return counts, nil
}

// cleanProfile cleans a specific browser profile
func (bc *BrowserCleaner) cleanProfile(profile BrowserProfile, createBackup bool) BrowserCleanResult {
	result := BrowserCleanResult{
		Profile: profile,
	}
	
	// Create backup if requested
	if createBackup {
		backupPath, err := bc.createProfileBackup(profile)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to create backup: %v", err))
			return result
		}
		result.BackupPath = backupPath
	}
	
	// Clean based on browser type
	switch profile.Type {
	case Chrome, Edge:
		bc.cleanChromiumBrowser(profile, &result)
	case Firefox:
		bc.cleanFirefoxBrowser(profile, &result)
	case Safari:
		bc.cleanSafariBrowser(profile, &result)
	}
	
	return result
}

// cleanChromiumBrowser cleans Chrome/Edge browsers (Chromium-based)
func (bc *BrowserCleaner) cleanChromiumBrowser(profile BrowserProfile, result *BrowserCleanResult) {
	// Clean cookies database
	cookiesDB := filepath.Join(profile.ProfilePath, "Cookies")
	if _, err := os.Stat(cookiesDB); err == nil {
		deleted, err := bc.cleanChromiumCookies(cookiesDB)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to clean cookies: %v", err))
		} else {
			result.CookiesDeleted = deleted
		}
	}
	
	// Clean local storage
	localStorageDir := filepath.Join(profile.ProfilePath, "Local Storage", "leveldb")
	if _, err := os.Stat(localStorageDir); err == nil {
		deleted, err := bc.cleanChromiumLocalStorage(localStorageDir)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to clean local storage: %v", err))
		} else {
			result.StorageDeleted = deleted
		}
	}
	
	// Clean session storage
	sessionStorageDir := filepath.Join(profile.ProfilePath, "Session Storage")
	if _, err := os.Stat(sessionStorageDir); err == nil {
		deleted, err := bc.cleanChromiumSessionStorage(sessionStorageDir)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to clean session storage: %v", err))
		} else {
			result.StorageDeleted += deleted
		}
	}
	
	// Clean cache
	cacheDir := filepath.Join(profile.ProfilePath, "Cache")
	if _, err := os.Stat(cacheDir); err == nil {
		deleted, err := bc.cleanChromiumCache(cacheDir)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to clean cache: %v", err))
		} else {
			result.CacheDeleted = deleted
		}
	}
}

// cleanChromiumCookies cleans Augment-related cookies from Chromium browsers
func (bc *BrowserCleaner) cleanChromiumCookies(cookiesDBPath string) (int64, error) {
	db, err := sql.Open("sqlite3", cookiesDBPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open cookies database: %w", err)
	}
	defer db.Close()

	// Enhanced patterns for Augment-related domains and cookie names
	augmentPatterns := []string{
		"%augment%",
		"%augmentcode%",
		"%augment-code%",
		"%vscode-augment%",
		"%augment.code%",
		"%augment_telemetry%",
		"%augment_session%",
		"%augment_user%",
	}

	var totalDeleted int64

	// Delete cookies with Augment-related domains or names
	for _, pattern := range augmentPatterns {
		query := `DELETE FROM cookies WHERE host_key LIKE ? OR name LIKE ? OR value LIKE ?`
		result, err := db.Exec(query, pattern, pattern, pattern)
		if err != nil {
			return totalDeleted, fmt.Errorf("failed to delete cookies with pattern %s: %w", pattern, err)
		}

		deleted, err := result.RowsAffected()
		if err != nil {
			return totalDeleted, fmt.Errorf("failed to get affected rows for pattern %s: %w", pattern, err)
		}

		totalDeleted += deleted
	}

	return totalDeleted, nil
}

// cleanChromiumLocalStorage cleans Augment-related local storage
func (bc *BrowserCleaner) cleanChromiumLocalStorage(storageDir string) (int64, error) {
	var deleted int64

	// Enhanced patterns for Augment-related storage files
	augmentPatterns := []string{
		"augment",
		"augmentcode",
		"augment-code",
		"vscode-augment",
		"augment.code",
		"augment_telemetry",
		"augment_session",
		"augment_user",
	}

	// LevelDB files containing Augment data
	err := filepath.Walk(storageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			fileName := strings.ToLower(info.Name())

			// Check if file contains any Augment-related patterns
			for _, pattern := range augmentPatterns {
				if strings.Contains(fileName, pattern) {
					if err := os.Remove(path); err == nil {
						deleted++
					}
					break
				}
			}
		}

		return nil
	})

	return deleted, err
}

// cleanChromiumSessionStorage cleans Augment-related session storage
func (bc *BrowserCleaner) cleanChromiumSessionStorage(storageDir string) (int64, error) {
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

// cleanChromiumCache cleans Augment-related cache files
func (bc *BrowserCleaner) cleanChromiumCache(cacheDir string) (int64, error) {
	var deleted int64
	
	// This is a simplified implementation
	// In practice, cache cleaning is complex and may require
	// parsing cache index files
	
	err := filepath.Walk(cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() {
			// Check if file contains Augment-related data
			// This is a simplified check
			if bc.containsAugmentData(path) {
				if err := os.Remove(path); err == nil {
					deleted++
				}
			}
		}
		
		return nil
	})
	
	return deleted, err
}

// cleanFirefoxBrowser cleans Firefox browser data
func (bc *BrowserCleaner) cleanFirefoxBrowser(profile BrowserProfile, result *BrowserCleanResult) {
	// Clean cookies database
	cookiesDB := filepath.Join(profile.ProfilePath, "cookies.sqlite")
	if _, err := os.Stat(cookiesDB); err == nil {
		deleted, err := bc.cleanFirefoxCookies(cookiesDB)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to clean cookies: %v", err))
		} else {
			result.CookiesDeleted = deleted
		}
	}
	
	// Clean local storage
	storageDir := filepath.Join(profile.ProfilePath, "storage", "default")
	if _, err := os.Stat(storageDir); err == nil {
		deleted, err := bc.cleanFirefoxStorage(storageDir)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to clean storage: %v", err))
		} else {
			result.StorageDeleted = deleted
		}
	}
	
	// Clean cache
	cacheDir := filepath.Join(profile.ProfilePath, "cache2")
	if _, err := os.Stat(cacheDir); err == nil {
		deleted, err := bc.cleanFirefoxCache(cacheDir)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to clean cache: %v", err))
		} else {
			result.CacheDeleted = deleted
		}
	}
}

// cleanFirefoxCookies cleans Augment-related cookies from Firefox
func (bc *BrowserCleaner) cleanFirefoxCookies(cookiesDBPath string) (int64, error) {
	db, err := sql.Open("sqlite3", cookiesDBPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open cookies database: %w", err)
	}
	defer db.Close()
	
	// Delete cookies with Augment-related domains
	query := `DELETE FROM moz_cookies WHERE host LIKE '%augment%' OR name LIKE '%augment%'`
	result, err := db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete cookies: %w", err)
	}
	
	deleted, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get affected rows: %w", err)
	}
	
	return deleted, nil
}

// cleanFirefoxStorage cleans Augment-related storage from Firefox
func (bc *BrowserCleaner) cleanFirefoxStorage(storageDir string) (int64, error) {
	var deleted int64

	err := filepath.Walk(storageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && strings.Contains(strings.ToLower(info.Name()), "augment") {
			if err := os.RemoveAll(path); err == nil {
				deleted++
			}
			return filepath.SkipDir
		}

		return nil
	})

	return deleted, err
}

// cleanFirefoxCache cleans Augment-related cache from Firefox
func (bc *BrowserCleaner) cleanFirefoxCache(cacheDir string) (int64, error) {
	var deleted int64

	err := filepath.Walk(cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && bc.containsAugmentData(path) {
			if err := os.Remove(path); err == nil {
				deleted++
			}
		}

		return nil
	})

	return deleted, err
}
