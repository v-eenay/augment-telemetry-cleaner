package cleaner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"augment-telemetry-cleaner/internal/scanner"
)

// ExtensionCleanResult represents the result of extension data cleaning
type ExtensionCleanResult struct {
	ExtensionID         string                    `json:"extension_id"`
	CleanedStorageItems []CleanedStorageItem      `json:"cleaned_storage_items"`
	CleanedCacheFiles   []CleanedCacheFile        `json:"cleaned_cache_files"`
	CleanedTempFiles    []CleanedTempFile         `json:"cleaned_temp_files"`
	BackupPaths         []string                  `json:"backup_paths"`
	TotalSizeRemoved    int64                     `json:"total_size_removed"`
	TelemetrySizeRemoved int64                    `json:"telemetry_size_removed"`
	ItemsRemoved        int                       `json:"items_removed"`
	Errors              []string                  `json:"errors"`
	CleanupDuration     time.Duration             `json:"cleanup_duration"`
	SafetyChecks        SafetyCheckResult         `json:"safety_checks"`
}

// CleanedStorageItem represents a cleaned storage item
type CleanedStorageItem struct {
	Key          string                  `json:"key"`
	OriginalSize int64                   `json:"original_size"`
	Risk         scanner.TelemetryRisk   `json:"risk"`
	StorageType  string                  `json:"storage_type"`
	BackupPath   string                  `json:"backup_path"`
	RemovalTime  time.Time               `json:"removal_time"`
}

// CleanedCacheFile represents a cleaned cache file
type CleanedCacheFile struct {
	Path         string                `json:"path"`
	Size         int64                 `json:"size"`
	Risk         scanner.TelemetryRisk `json:"risk"`
	BackupPath   string                `json:"backup_path"`
	RemovalTime  time.Time             `json:"removal_time"`
}

// CleanedTempFile represents a cleaned temporary file
type CleanedTempFile struct {
	Path         string                `json:"path"`
	Size         int64                 `json:"size"`
	Risk         scanner.TelemetryRisk `json:"risk"`
	Age          time.Duration         `json:"age"`
	RemovalTime  time.Time             `json:"removal_time"`
}

// SafetyCheckResult represents the result of safety checks
type SafetyCheckResult struct {
	Passed           bool     `json:"passed"`
	Warnings         []string `json:"warnings"`
	BlockingIssues   []string `json:"blocking_issues"`
	BackupVerified   bool     `json:"backup_verified"`
	DependencyCheck  bool     `json:"dependency_check"`
	RollbackCapable  bool     `json:"rollback_capable"`
}

// RemovalPolicy represents policies for data removal
type RemovalPolicy struct {
	MinRiskLevel        scanner.TelemetryRisk `json:"min_risk_level"`
	MaxFileAge          time.Duration         `json:"max_file_age"`
	MaxFileSize         int64                 `json:"max_file_size"`
	PreserveRecent      bool                  `json:"preserve_recent"`
	RecentThreshold     time.Duration         `json:"recent_threshold"`
	CreateBackups       bool                  `json:"create_backups"`
	VerifyBackups       bool                  `json:"verify_backups"`
	DryRun              bool                  `json:"dry_run"`
	RequireConfirmation bool                  `json:"require_confirmation"`
	ExcludePatterns     []string              `json:"exclude_patterns"`
	IncludePatterns     []string              `json:"include_patterns"`
}

// ExtensionCleaner handles intelligent removal of extension data
type ExtensionCleaner struct {
	policy          RemovalPolicy
	backupManager   *BackupManager
	dependencyChecker *DependencyChecker
	safetyValidator *SafetyValidator
}

// NewExtensionCleaner creates a new extension cleaner
func NewExtensionCleaner(policy RemovalPolicy) *ExtensionCleaner {
	return &ExtensionCleaner{
		policy:            policy,
		backupManager:     NewBackupManager(),
		dependencyChecker: NewDependencyChecker(),
		safetyValidator:   NewSafetyValidator(),
	}
}

// CleanExtensionData performs intelligent cleaning of extension data
func (ec *ExtensionCleaner) CleanExtensionData(extensionStorage scanner.ExtensionStorage) (*ExtensionCleanResult, error) {
	startTime := time.Now()
	
	result := &ExtensionCleanResult{
		ExtensionID:         extensionStorage.ExtensionID,
		CleanedStorageItems: make([]CleanedStorageItem, 0),
		CleanedCacheFiles:   make([]CleanedCacheFile, 0),
		CleanedTempFiles:    make([]CleanedTempFile, 0),
		BackupPaths:         make([]string, 0),
		Errors:              make([]string, 0),
	}

	// Perform safety checks
	safetyResult, err := ec.performSafetyChecks(extensionStorage)
	if err != nil {
		return nil, fmt.Errorf("safety checks failed: %w", err)
	}
	result.SafetyChecks = *safetyResult

	if !safetyResult.Passed {
		return result, fmt.Errorf("safety checks failed, aborting cleanup")
	}

	// Create backup if required
	if ec.policy.CreateBackups {
		backupPath, err := ec.createExtensionBackup(extensionStorage)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Backup creation failed: %v", err))
			if ec.policy.VerifyBackups {
				return result, fmt.Errorf("backup creation failed and verification is required")
			}
		} else {
			result.BackupPaths = append(result.BackupPaths, backupPath)
			result.SafetyChecks.BackupVerified = true
		}
	}

	// Clean storage items
	if err := ec.cleanStorageItems(extensionStorage.StorageItems, result); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Storage cleaning failed: %v", err))
	}

	// Clean cache files (if available)
	// This would integrate with cache analysis from Phase 3
	// For now, we'll add a placeholder

	// Clean temporary files (if available)
	// This would integrate with temp file analysis from Phase 3
	// For now, we'll add a placeholder

	result.CleanupDuration = time.Since(startTime)
	return result, nil
}

// performSafetyChecks performs comprehensive safety checks before cleaning
func (ec *ExtensionCleaner) performSafetyChecks(extensionStorage scanner.ExtensionStorage) (*SafetyCheckResult, error) {
	result := &SafetyCheckResult{
		Passed:         true,
		Warnings:       make([]string, 0),
		BlockingIssues: make([]string, 0),
	}

	// Check if extension is currently active
	if ec.isExtensionActive(extensionStorage.ExtensionID) {
		result.BlockingIssues = append(result.BlockingIssues, 
			fmt.Sprintf("Extension %s is currently active", extensionStorage.ExtensionID))
		result.Passed = false
	}

	// Check for critical dependencies
	dependencies, err := ec.dependencyChecker.CheckDependencies(extensionStorage.ExtensionID)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Dependency check failed: %v", err))
	} else if len(dependencies) > 0 {
		result.Warnings = append(result.Warnings, 
			fmt.Sprintf("Extension has %d dependencies that may be affected", len(dependencies)))
		result.DependencyCheck = true
	}

	// Check storage path accessibility
	if _, err := os.Stat(extensionStorage.StoragePath); err != nil {
		result.BlockingIssues = append(result.BlockingIssues, 
			fmt.Sprintf("Storage path not accessible: %v", err))
		result.Passed = false
	}

	// Validate backup capability
	if ec.policy.CreateBackups {
		if err := ec.validateBackupCapability(extensionStorage); err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Backup validation failed: %v", err))
		} else {
			result.RollbackCapable = true
		}
	}

	// Check for high-risk data that requires special handling
	criticalItems := ec.findCriticalItems(extensionStorage.StorageItems)
	if len(criticalItems) > 0 {
		result.Warnings = append(result.Warnings, 
			fmt.Sprintf("Found %d critical telemetry items", len(criticalItems)))
	}

	return result, nil
}

// createExtensionBackup creates a comprehensive backup of extension data
func (ec *ExtensionCleaner) createExtensionBackup(extensionStorage scanner.ExtensionStorage) (string, error) {
	timestamp := time.Now().Unix()
	backupName := fmt.Sprintf("%s-backup-%d", 
		strings.ReplaceAll(extensionStorage.ExtensionID, ".", "-"), 
		timestamp)

	backupPath, err := ec.backupManager.CreateExtensionBackup(extensionStorage, backupName)
	if err != nil {
		return "", fmt.Errorf("failed to create backup: %w", err)
	}

	// Verify backup if required
	if ec.policy.VerifyBackups {
		if err := ec.backupManager.VerifyBackup(backupPath); err != nil {
			// Clean up failed backup
			os.RemoveAll(backupPath)
			return "", fmt.Errorf("backup verification failed: %w", err)
		}
	}

	return backupPath, nil
}

// cleanStorageItems cleans individual storage items based on policy
func (ec *ExtensionCleaner) cleanStorageItems(items []scanner.StorageDataItem, result *ExtensionCleanResult) error {
	for _, item := range items {
		// Check if item should be cleaned based on policy
		if !ec.shouldCleanItem(item) {
			continue
		}

		// Perform the cleaning
		if ec.policy.DryRun {
			// Dry run - just record what would be cleaned
			cleanedItem := CleanedStorageItem{
				Key:          item.Key,
				OriginalSize: item.Size,
				Risk:         item.Risk,
				StorageType:  "global", // This would be determined from context
				RemovalTime:  time.Now(),
			}
			result.CleanedStorageItems = append(result.CleanedStorageItems, cleanedItem)
			result.TotalSizeRemoved += item.Size
			if item.Risk >= scanner.TelemetryRiskMedium {
				result.TelemetrySizeRemoved += item.Size
			}
		} else {
			// Actually remove the item
			if err := ec.removeStorageItem(item, result); err != nil {
				result.Errors = append(result.Errors, 
					fmt.Sprintf("Failed to remove item %s: %v", item.Key, err))
				continue
			}
		}

		result.ItemsRemoved++
	}

	return nil
}

// shouldCleanItem determines if a storage item should be cleaned based on policy
func (ec *ExtensionCleaner) shouldCleanItem(item scanner.StorageDataItem) bool {
	// Check minimum risk level
	if item.Risk < ec.policy.MinRiskLevel {
		return false
	}

	// Check file age
	if ec.policy.MaxFileAge > 0 {
		age := time.Since(item.LastModified)
		if age < ec.policy.MaxFileAge {
			return false
		}
	}

	// Check file size
	if ec.policy.MaxFileSize > 0 && item.Size > ec.policy.MaxFileSize {
		return false
	}

	// Preserve recent files if policy requires
	if ec.policy.PreserveRecent {
		age := time.Since(item.LastModified)
		if age < ec.policy.RecentThreshold {
			return false
		}
	}

	// Check exclude patterns
	for _, pattern := range ec.policy.ExcludePatterns {
		if ec.matchesPattern(item.Key, pattern) {
			return false
		}
	}

	// Check include patterns (if specified, item must match at least one)
	if len(ec.policy.IncludePatterns) > 0 {
		matched := false
		for _, pattern := range ec.policy.IncludePatterns {
			if ec.matchesPattern(item.Key, pattern) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

// removeStorageItem removes a single storage item
func (ec *ExtensionCleaner) removeStorageItem(item scanner.StorageDataItem, result *ExtensionCleanResult) error {
	// This is a simplified implementation
	// In practice, this would need to handle different storage types
	// and integrate with the actual VS Code storage mechanisms

	cleanedItem := CleanedStorageItem{
		Key:          item.Key,
		OriginalSize: item.Size,
		Risk:         item.Risk,
		StorageType:  "global", // Would be determined from context
		RemovalTime:  time.Now(),
	}

	// Create individual backup if needed
	if ec.policy.CreateBackups {
		backupPath, err := ec.backupManager.BackupStorageItem(item)
		if err != nil {
			return fmt.Errorf("failed to backup item: %w", err)
		}
		cleanedItem.BackupPath = backupPath
	}

	// Add to results
	result.CleanedStorageItems = append(result.CleanedStorageItems, cleanedItem)
	result.TotalSizeRemoved += item.Size
	if item.Risk >= scanner.TelemetryRiskMedium {
		result.TelemetrySizeRemoved += item.Size
	}

	return nil
}

// Helper methods

// isExtensionActive checks if an extension is currently active
func (ec *ExtensionCleaner) isExtensionActive(extensionID string) bool {
	// This would integrate with VS Code's extension management
	// For now, return false as a placeholder
	return false
}

// findCriticalItems finds items with critical telemetry risk
func (ec *ExtensionCleaner) findCriticalItems(items []scanner.StorageDataItem) []scanner.StorageDataItem {
	var critical []scanner.StorageDataItem
	for _, item := range items {
		if item.Risk == scanner.TelemetryRiskCritical {
			critical = append(critical, item)
		}
	}
	return critical
}

// validateBackupCapability validates that backups can be created
func (ec *ExtensionCleaner) validateBackupCapability(extensionStorage scanner.ExtensionStorage) error {
	// Check if backup directory is writable
	backupDir := ec.backupManager.GetBackupDirectory()
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("cannot create backup directory: %w", err)
	}

	// Test write permissions
	testFile := filepath.Join(backupDir, "test-write-permissions")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("backup directory not writable: %w", err)
	}
	os.Remove(testFile)

	return nil
}

// matchesPattern checks if a string matches a pattern (supports wildcards)
func (ec *ExtensionCleaner) matchesPattern(text, pattern string) bool {
	// Simple pattern matching with * wildcard support
	if pattern == "*" {
		return true
	}
	
	if strings.Contains(pattern, "*") {
		// Convert to simple regex-like matching
		parts := strings.Split(pattern, "*")
		if len(parts) == 2 {
			return strings.HasPrefix(text, parts[0]) && strings.HasSuffix(text, parts[1])
		}
	}
	
	return strings.Contains(strings.ToLower(text), strings.ToLower(pattern))
}

// GetDefaultRemovalPolicy returns a default removal policy
func GetDefaultRemovalPolicy() RemovalPolicy {
	return RemovalPolicy{
		MinRiskLevel:        scanner.TelemetryRiskMedium,
		MaxFileAge:          0, // No age limit
		MaxFileSize:         0, // No size limit
		PreserveRecent:      true,
		RecentThreshold:     24 * time.Hour, // Preserve files modified in last 24 hours
		CreateBackups:       true,
		VerifyBackups:       true,
		DryRun:              false,
		RequireConfirmation: true,
		ExcludePatterns:     []string{"config", "settings", "preferences"},
		IncludePatterns:     []string{}, // Empty means include all (subject to other filters)
	}
}

// GetAggressiveRemovalPolicy returns an aggressive removal policy for high-risk data
func GetAggressiveRemovalPolicy() RemovalPolicy {
	return RemovalPolicy{
		MinRiskLevel:        scanner.TelemetryRiskLow,
		MaxFileAge:          0,
		MaxFileSize:         0,
		PreserveRecent:      false,
		RecentThreshold:     0,
		CreateBackups:       true,
		VerifyBackups:       true,
		DryRun:              false,
		RequireConfirmation: true,
		ExcludePatterns:     []string{},
		IncludePatterns:     []string{"telemetry", "analytics", "tracking", "usage", "metrics"},
	}
}

// GetConservativeRemovalPolicy returns a conservative removal policy
func GetConservativeRemovalPolicy() RemovalPolicy {
	return RemovalPolicy{
		MinRiskLevel:        scanner.TelemetryRiskHigh,
		MaxFileAge:          30 * 24 * time.Hour, // Only remove files older than 30 days
		MaxFileSize:         0,
		PreserveRecent:      true,
		RecentThreshold:     7 * 24 * time.Hour, // Preserve files modified in last 7 days
		CreateBackups:       true,
		VerifyBackups:       true,
		DryRun:              false,
		RequireConfirmation: true,
		ExcludePatterns:     []string{"config", "settings", "preferences", "cache"},
		IncludePatterns:     []string{"telemetry", "analytics"},
	}
}