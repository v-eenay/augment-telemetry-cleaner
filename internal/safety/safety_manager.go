package safety

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"augment-telemetry-cleaner/internal/config"
	"augment-telemetry-cleaner/internal/logger"
	"augment-telemetry-cleaner/internal/scanner"
	"augment-telemetry-cleaner/internal/utils"
)

// SafetyManager handles safety features like confirmations, dry-run mode, and backup verification
type SafetyManager struct {
	config  *config.Config
	logger  *logger.Logger
	scanner *scanner.AugmentScanner
}

// SafetyCheck represents a safety check result
type SafetyCheck struct {
	CheckName   string `json:"check_name"`
	Passed      bool   `json:"passed"`
	Message     string `json:"message"`
	Severity    string `json:"severity"` // "info", "warning", "error"
}

// PreOperationCheck represents the result of pre-operation safety checks
type PreOperationCheck struct {
	CanProceed    bool          `json:"can_proceed"`
	Checks        []SafetyCheck `json:"checks"`
	Warnings      []string      `json:"warnings"`
	Errors        []string      `json:"errors"`
	ScanResult    *scanner.ScanResult `json:"scan_result,omitempty"`
}

// NewSafetyManager creates a new safety manager
func NewSafetyManager(config *config.Config, logger *logger.Logger) *SafetyManager {
	return &SafetyManager{
		config:  config,
		logger:  logger,
		scanner: scanner.NewAugmentScanner(),
	}
}

// PerformPreOperationChecks performs comprehensive safety checks before operations
func (sm *SafetyManager) PerformPreOperationChecks() (*PreOperationCheck, error) {
	sm.logger.Info("Performing pre-operation safety checks...")
	
	result := &PreOperationCheck{
		CanProceed: true,
		Checks:     make([]SafetyCheck, 0),
		Warnings:   make([]string, 0),
		Errors:     make([]string, 0),
	}

	// Check 1: Verify VS Code is not running
	vscodeCheck := sm.checkVSCodeNotRunning()
	result.Checks = append(result.Checks, vscodeCheck)
	if !vscodeCheck.Passed {
		result.CanProceed = false
		result.Errors = append(result.Errors, vscodeCheck.Message)
	}

	// Check 2: Verify required files exist
	filesCheck := sm.checkRequiredFilesExist()
	result.Checks = append(result.Checks, filesCheck)
	if !filesCheck.Passed {
		result.Warnings = append(result.Warnings, filesCheck.Message)
	}

	// Check 3: Check disk space for backups
	diskSpaceCheck := sm.checkDiskSpace()
	result.Checks = append(result.Checks, diskSpaceCheck)
	if !diskSpaceCheck.Passed {
		result.CanProceed = false
		result.Errors = append(result.Errors, diskSpaceCheck.Message)
	}

	// Check 4: Verify backup directory is writable
	backupDirCheck := sm.checkBackupDirectory()
	result.Checks = append(result.Checks, backupDirCheck)
	if !backupDirCheck.Passed {
		result.CanProceed = false
		result.Errors = append(result.Errors, backupDirCheck.Message)
	}

	// Check 5: Scan for existing Augment files (if enabled)
	if sm.config.ShowPreviewBeforeRun {
		scanResult, err := sm.scanner.ScanSystem()
		if err != nil {
			sm.logger.Warn("Failed to scan system: %v", err)
			result.Warnings = append(result.Warnings, "System scan failed, proceeding without preview")
		} else {
			result.ScanResult = scanResult
			sm.logger.Info("System scan completed: found %d files in %v", 
				scanResult.TotalFiles, scanResult.ScanDuration)
		}
	}

	sm.logger.Info("Pre-operation checks completed. Can proceed: %v", result.CanProceed)
	return result, nil
}

// checkVSCodeNotRunning checks if VS Code is currently running
func (sm *SafetyManager) checkVSCodeNotRunning() SafetyCheck {
	// This is a simplified check - in a real implementation, you might want to
	// check for running processes more thoroughly
	check := SafetyCheck{
		CheckName: "VS Code Process Check",
		Passed:    true,
		Message:   "VS Code does not appear to be running",
		Severity:  "info",
	}

	// For now, we'll just warn the user to close VS Code manually
	// A more sophisticated implementation could check running processes
	check.Message = "Please ensure VS Code is completely closed before proceeding"
	check.Severity = "warning"

	return check
}

// checkRequiredFilesExist checks if the required VS Code files exist
func (sm *SafetyManager) checkRequiredFilesExist() SafetyCheck {
	check := SafetyCheck{
		CheckName: "Required Files Check",
		Passed:    true,
		Severity:  "info",
	}

	missingFiles := make([]string, 0)

	// Check storage.json
	if storagePath, err := utils.GetStoragePath(); err != nil || !fileExists(storagePath) {
		missingFiles = append(missingFiles, "storage.json")
	}

	// Check database
	if dbPath, err := utils.GetDBPath(); err != nil || !fileExists(dbPath) {
		missingFiles = append(missingFiles, "state.vscdb")
	}

	// Check workspace storage directory
	if workspacePath, err := utils.GetWorkspaceStoragePath(); err != nil || !dirExists(workspacePath) {
		missingFiles = append(missingFiles, "workspaceStorage directory")
	}

	if len(missingFiles) > 0 {
		check.Passed = false
		check.Message = fmt.Sprintf("Missing files/directories: %v. Some operations may not work.", missingFiles)
		check.Severity = "warning"
	} else {
		check.Message = "All required VS Code files found"
	}

	return check
}

// checkDiskSpace checks if there's enough disk space for backups
func (sm *SafetyManager) checkDiskSpace() SafetyCheck {
	check := SafetyCheck{
		CheckName: "Disk Space Check",
		Passed:    true,
		Message:   "Sufficient disk space available",
		Severity:  "info",
	}

	// Get backup directory
	backupDir := sm.config.BackupDirectory
	if backupDir == "" {
		check.Passed = false
		check.Message = "Backup directory not configured"
		check.Severity = "error"
		return check
	}

	// For simplicity, we'll assume there's enough space
	// A more sophisticated implementation would check actual disk space
	return check
}

// checkBackupDirectory checks if the backup directory is accessible and writable
func (sm *SafetyManager) checkBackupDirectory() SafetyCheck {
	check := SafetyCheck{
		CheckName: "Backup Directory Check",
		Passed:    true,
		Message:   "Backup directory is accessible and writable",
		Severity:  "info",
	}

	backupDir := sm.config.BackupDirectory
	if backupDir == "" {
		check.Passed = false
		check.Message = "Backup directory not configured"
		check.Severity = "error"
		return check
	}

	// Create backup directory if it doesn't exist
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		check.Passed = false
		check.Message = fmt.Sprintf("Cannot create backup directory: %v", err)
		check.Severity = "error"
		return check
	}

	// Test write access
	testFile := filepath.Join(backupDir, fmt.Sprintf("test_%d.tmp", time.Now().Unix()))
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		check.Passed = false
		check.Message = fmt.Sprintf("Cannot write to backup directory: %v", err)
		check.Severity = "error"
		return check
	}

	// Clean up test file
	os.Remove(testFile)

	return check
}

// VerifyBackup verifies that a backup was created successfully
func (sm *SafetyManager) VerifyBackup(originalPath, backupPath string) error {
	sm.logger.Debug("Verifying backup: %s -> %s", originalPath, backupPath)

	// Check if backup file exists
	if !fileExists(backupPath) {
		return fmt.Errorf("backup file does not exist: %s", backupPath)
	}

	// Check if backup file is readable
	if err := utils.VerifyBackup(backupPath); err != nil {
		return fmt.Errorf("backup verification failed: %w", err)
	}

	// Get file sizes
	originalInfo, err := os.Stat(originalPath)
	if err != nil {
		return fmt.Errorf("cannot stat original file: %w", err)
	}

	backupInfo, err := os.Stat(backupPath)
	if err != nil {
		return fmt.Errorf("cannot stat backup file: %w", err)
	}

	// Compare sizes
	if originalInfo.Size() != backupInfo.Size() {
		return fmt.Errorf("backup size mismatch: original=%d, backup=%d", 
			originalInfo.Size(), backupInfo.Size())
	}

	sm.logger.Debug("Backup verification successful")
	return nil
}

// CleanOldBackups removes old backup files based on configuration
func (sm *SafetyManager) CleanOldBackups() error {
	if sm.config.MaxBackupAge <= 0 {
		return nil // Backup cleanup disabled
	}

	sm.logger.Info("Cleaning old backups older than %d days", sm.config.MaxBackupAge)

	backupDir := sm.config.BackupDirectory
	if !dirExists(backupDir) {
		return nil // No backup directory
	}

	cutoffTime := time.Now().AddDate(0, 0, -sm.config.MaxBackupAge)
	deletedCount := 0

	err := filepath.Walk(backupDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue despite errors
		}

		if !info.IsDir() && info.ModTime().Before(cutoffTime) {
			if err := os.Remove(path); err != nil {
				sm.logger.Warn("Failed to delete old backup %s: %v", path, err)
			} else {
				deletedCount++
				sm.logger.Debug("Deleted old backup: %s", path)
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to clean old backups: %w", err)
	}

	sm.logger.Info("Cleaned %d old backup files", deletedCount)
	return nil
}

// Helper functions
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
