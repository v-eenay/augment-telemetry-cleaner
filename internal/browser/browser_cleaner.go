package browser

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	processManager := NewProcessManager()
	
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
			// Try to force close the browser
			if err := processManager.ForceCloseBrowser(profile.Type); err != nil {
				result := BrowserCleanResult{
					Profile: profile,
					Errors:  []string{fmt.Sprintf("Failed to close %s processes: %v", profile.Type.String(), err)},
				}
				results = append(results, result)
				continue
			}
			
			// Wait for processes to close
			if err := processManager.WaitForProcessesToClose(profile.Type, 10*time.Second); err != nil {
				result := BrowserCleanResult{
					Profile: profile,
					Errors:  []string{fmt.Sprintf("%s processes did not close in time. Please close manually and try again.", profile.Type.String())},
				}
				results = append(results, result)
				continue
			}
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
	// Handle WAL mode files
	walFile := cookiesDBPath + "-wal"
	shmFile := cookiesDBPath + "-shm"
	
	// Remove WAL and SHM files if they exist (they prevent database access)
	if _, err := os.Stat(walFile); err == nil {
		os.Remove(walFile)
	}
	if _, err := os.Stat(shmFile); err == nil {
		os.Remove(shmFile)
	}

	// Open database with retry mechanism and timeout
	connectionString := fmt.Sprintf("%s?_timeout=30000&_journal_mode=DELETE&_synchronous=NORMAL", cookiesDBPath)
	db, err := sql.Open("sqlite3", connectionString)
	if err != nil {
		return 0, fmt.Errorf("failed to open cookies database: %w", err)
	}
	defer db.Close()

	// Set connection pool settings
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	// Test connection with retry
	var connectionErr error
	for i := 0; i < 3; i++ {
		if connectionErr = db.Ping(); connectionErr == nil {
			break
		}
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	if connectionErr != nil {
		return 0, fmt.Errorf("failed to connect to database after retries: %w", connectionErr)
	}

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
		"%augmentai%",
		"%augment-ai%",
	}

	var totalDeleted int64

	// Begin transaction for better performance and atomicity
	tx, err := db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete cookies with Augment-related domains or names
	for _, pattern := range augmentPatterns {
		query := `DELETE FROM cookies WHERE host_key LIKE ? OR name LIKE ? OR value LIKE ?`
		result, err := tx.Exec(query, pattern, pattern, pattern)
		if err != nil {
			return totalDeleted, fmt.Errorf("failed to delete cookies with pattern %s: %w", pattern, err)
		}

		deleted, err := result.RowsAffected()
		if err != nil {
			return totalDeleted, fmt.Errorf("failed to get affected rows for pattern %s: %w", pattern, err)
		}

		totalDeleted += deleted
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return totalDeleted, fmt.Errorf("failed to commit transaction: %w", err)
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
		"augmentai",
		"augment-ai",
	}

	// First, try to remove any lock files that might prevent access
	lockFiles := []string{
		filepath.Join(storageDir, "LOCK"),
		filepath.Join(storageDir, "LOG"),
		filepath.Join(storageDir, "LOG.old"),
	}
	
	for _, lockFile := range lockFiles {
		if _, err := os.Stat(lockFile); err == nil {
			// Try to remove lock files, but don't fail if we can't
			os.Remove(lockFile)
		}
	}

	// LevelDB files containing Augment data
	err := filepath.Walk(storageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip files we can't access instead of failing
			return nil
		}

		if !info.IsDir() {
			fileName := strings.ToLower(info.Name())

			// Check if file contains any Augment-related patterns
			for _, pattern := range augmentPatterns {
				if strings.Contains(fileName, pattern) {
					// Try multiple times to remove the file
					for i := 0; i < 3; i++ {
						if err := os.Remove(path); err == nil {
							deleted++
							break
						}
						time.Sleep(100 * time.Millisecond)
					}
					break
				}
			}

			// Also check for files that might contain Augment data in their content
			// This is more thorough but slower
			if bc.shouldCheckFileContent(fileName) {
				if bc.fileContainsAugmentData(path) {
					for i := 0; i < 3; i++ {
						if err := os.Remove(path); err == nil {
							deleted++
							break
						}
						time.Sleep(100 * time.Millisecond)
					}
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
	
	// Remove lock files first
	lockFiles := []string{
		filepath.Join(storageDir, "LOCK"),
		filepath.Join(storageDir, "LOG"),
		filepath.Join(storageDir, "LOG.old"),
	}
	
	for _, lockFile := range lockFiles {
		if _, err := os.Stat(lockFile); err == nil {
			os.Remove(lockFile)
		}
	}
	
	err := filepath.Walk(storageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}
		
		if !info.IsDir() {
			fileName := strings.ToLower(info.Name())
			
			// Check filename for Augment patterns
			augmentPatterns := []string{
				"augment",
				"augmentcode",
				"augment-code",
				"vscode-augment",
				"augment.code",
				"augmentai",
				"augment-ai",
			}
			
			for _, pattern := range augmentPatterns {
				if strings.Contains(fileName, pattern) {
					// Try multiple times to remove the file
					for i := 0; i < 3; i++ {
						if err := os.Remove(path); err == nil {
							deleted++
							break
						}
						time.Sleep(100 * time.Millisecond)
					}
					break
				}
			}
		}
		
		return nil
	})
	
	return deleted, err
}

// cleanChromiumCache cleans Augment-related cache files
func (bc *BrowserCleaner) cleanChromiumCache(cacheDir string) (int64, error) {
	var deleted int64
	
	// Remove cache lock files first
	lockFiles := []string{
		filepath.Join(cacheDir, "index"),
		filepath.Join(cacheDir, "data_0"),
		filepath.Join(cacheDir, "data_1"),
		filepath.Join(cacheDir, "data_2"),
		filepath.Join(cacheDir, "data_3"),
	}
	
	// Try to remove cache index and data files that might be locked
	for _, lockFile := range lockFiles {
		if _, err := os.Stat(lockFile); err == nil {
			// Don't remove these core files, but check if they're accessible
			if file, err := os.OpenFile(lockFile, os.O_RDONLY, 0); err == nil {
				file.Close()
			}
		}
	}
	
	err := filepath.Walk(cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}
		
		if !info.IsDir() {
			fileName := strings.ToLower(info.Name())
			
			// Check filename for Augment patterns first (faster)
			augmentPatterns := []string{
				"augment",
				"augmentcode",
				"augment-code",
				"vscode-augment",
				"augment.code",
				"augmentai",
				"augment-ai",
			}
			
			for _, pattern := range augmentPatterns {
				if strings.Contains(fileName, pattern) {
					// Try multiple times to remove the file
					for i := 0; i < 3; i++ {
						if err := os.Remove(path); err == nil {
							deleted++
							break
						}
						time.Sleep(100 * time.Millisecond)
					}
					return nil
				}
			}
			
			// For cache files, also check content if it's a reasonable size
			if info.Size() < 10*1024*1024 && bc.shouldCheckFileContent(fileName) { // Only check files < 10MB
				if bc.fileContainsAugmentData(path) {
					for i := 0; i < 3; i++ {
						if err := os.Remove(path); err == nil {
							deleted++
							break
						}
						time.Sleep(100 * time.Millisecond)
					}
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
	// Handle WAL mode files for Firefox too
	walFile := cookiesDBPath + "-wal"
	shmFile := cookiesDBPath + "-shm"
	
	if _, err := os.Stat(walFile); err == nil {
		os.Remove(walFile)
	}
	if _, err := os.Stat(shmFile); err == nil {
		os.Remove(shmFile)
	}

	// Open database with retry mechanism and timeout
	connectionString := fmt.Sprintf("%s?_timeout=30000&_journal_mode=DELETE&_synchronous=NORMAL", cookiesDBPath)
	db, err := sql.Open("sqlite3", connectionString)
	if err != nil {
		return 0, fmt.Errorf("failed to open cookies database: %w", err)
	}
	defer db.Close()

	// Set connection pool settings
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	// Test connection with retry
	var connectionErr error
	for i := 0; i < 3; i++ {
		if connectionErr = db.Ping(); connectionErr == nil {
			break
		}
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	if connectionErr != nil {
		return 0, fmt.Errorf("failed to connect to database after retries: %w", connectionErr)
	}

	// Enhanced patterns for Firefox
	augmentPatterns := []string{
		"%augment%",
		"%augmentcode%",
		"%augment-code%",
		"%vscode-augment%",
		"%augment.code%",
		"%augment_telemetry%",
		"%augment_session%",
		"%augment_user%",
		"%augmentai%",
		"%augment-ai%",
	}

	var totalDeleted int64

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete cookies with Augment-related domains or names
	for _, pattern := range augmentPatterns {
		query := `DELETE FROM moz_cookies WHERE host LIKE ? OR name LIKE ? OR value LIKE ?`
		result, err := tx.Exec(query, pattern, pattern, pattern)
		if err != nil {
			return totalDeleted, fmt.Errorf("failed to delete cookies with pattern %s: %w", pattern, err)
		}

		deleted, err := result.RowsAffected()
		if err != nil {
			return totalDeleted, fmt.Errorf("failed to get affected rows for pattern %s: %w", pattern, err)
		}

		totalDeleted += deleted
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return totalDeleted, fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return totalDeleted, nil
}

// cleanFirefoxStorage cleans Augment-related storage from Firefox
func (bc *BrowserCleaner) cleanFirefoxStorage(storageDir string) (int64, error) {
	var deleted int64

	// Enhanced patterns for Firefox storage
	augmentPatterns := []string{
		"augment",
		"augmentcode",
		"augment-code",
		"vscode-augment",
		"augment.code",
		"augmentai",
		"augment-ai",
	}

	err := filepath.Walk(storageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		if info.IsDir() {
			dirName := strings.ToLower(info.Name())
			for _, pattern := range augmentPatterns {
				if strings.Contains(dirName, pattern) {
					// Try multiple times to remove the directory
					for i := 0; i < 3; i++ {
						if err := os.RemoveAll(path); err == nil {
							deleted++
							break
						}
						time.Sleep(100 * time.Millisecond)
					}
					return filepath.SkipDir
				}
			}
		} else {
			// Also check individual files
			fileName := strings.ToLower(info.Name())
			for _, pattern := range augmentPatterns {
				if strings.Contains(fileName, pattern) {
					for i := 0; i < 3; i++ {
						if err := os.Remove(path); err == nil {
							deleted++
							break
						}
						time.Sleep(100 * time.Millisecond)
					}
					break
				}
			}
		}

		return nil
	})

	return deleted, err
}

// cleanFirefoxCache cleans Augment-related cache from Firefox
func (bc *BrowserCleaner) cleanFirefoxCache(cacheDir string) (int64, error) {
	var deleted int64

	// Enhanced patterns for Firefox cache
	augmentPatterns := []string{
		"augment",
		"augmentcode",
		"augment-code",
		"vscode-augment",
		"augment.code",
		"augmentai",
		"augment-ai",
	}

	err := filepath.Walk(cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		if !info.IsDir() {
			fileName := strings.ToLower(info.Name())
			
			// Check filename for Augment patterns first
			for _, pattern := range augmentPatterns {
				if strings.Contains(fileName, pattern) {
					for i := 0; i < 3; i++ {
						if err := os.Remove(path); err == nil {
							deleted++
							break
						}
						time.Sleep(100 * time.Millisecond)
					}
					return nil
				}
			}
			
			// Also check content for smaller files
			if info.Size() < 5*1024*1024 && bc.shouldCheckFileContent(fileName) { // Only check files < 5MB
				if bc.fileContainsAugmentData(path) {
					for i := 0; i < 3; i++ {
						if err := os.Remove(path); err == nil {
							deleted++
							break
						}
						time.Sleep(100 * time.Millisecond)
					}
				}
			}
		}

		return nil
	})

	return deleted, err
}
// shouldCheckFileContent determines if we should scan file content for Augment data
func (bc *BrowserCleaner) shouldCheckFileContent(fileName string) bool {
	// Only check certain file types to avoid performance issues
	checkExtensions := []string{".ldb", ".log", ".sst", ".manifest"}
	
	for _, ext := range checkExtensions {
		if strings.HasSuffix(fileName, ext) {
			return true
		}
	}
	
	// Also check files without extensions (common in LevelDB)
	return !strings.Contains(fileName, ".")
}

// fileContainsAugmentData checks if a file contains Augment-related data in its content
func (bc *BrowserCleaner) fileContainsAugmentData(filePath string) bool {
	// Read first 1KB of file to check for Augment patterns
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()
	
	buffer := make([]byte, 1024)
	n, err := file.Read(buffer)
	if err != nil && err.Error() != "EOF" {
		return false
	}
	
	content := strings.ToLower(string(buffer[:n]))
	
	augmentPatterns := []string{
		"augment",
		"augmentcode",
		"augment-code",
		"vscode-augment",
		"augment.code",
		"augmentai",
		"augment-ai",
	}
	
	for _, pattern := range augmentPatterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}
	
	return false
}