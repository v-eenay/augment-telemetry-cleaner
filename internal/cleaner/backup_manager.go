package cleaner

import (
	"archive/zip"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"augment-telemetry-cleaner/internal/scanner"
)

// BackupManager handles creation, verification, and restoration of backups
type BackupManager struct {
	backupDirectory string
	maxBackupAge    time.Duration
	maxBackupSize   int64
}

// BackupMetadata represents metadata about a backup
type BackupMetadata struct {
	BackupID        string                    `json:"backup_id"`
	ExtensionID     string                    `json:"extension_id"`
	CreationTime    time.Time                 `json:"creation_time"`
	BackupType      string                    `json:"backup_type"`
	OriginalPath    string                    `json:"original_path"`
	BackupPath      string                    `json:"backup_path"`
	TotalSize       int64                     `json:"total_size"`
	FileCount       int                       `json:"file_count"`
	Checksum        string                    `json:"checksum"`
	BackupItems     []BackupItem              `json:"backup_items"`
	CompressionType string                    `json:"compression_type"`
	Verified        bool                      `json:"verified"`
	RestorationInfo *RestorationInfo          `json:"restoration_info,omitempty"`
}

// BackupItem represents an individual item in a backup
type BackupItem struct {
	RelativePath    string                `json:"relative_path"`
	OriginalPath    string                `json:"original_path"`
	Size            int64                 `json:"size"`
	ModTime         time.Time             `json:"mod_time"`
	Checksum        string                `json:"checksum"`
	ItemType        string                `json:"item_type"`
	Risk            scanner.TelemetryRisk `json:"risk"`
}

// RestorationInfo represents information about backup restoration
type RestorationInfo struct {
	RestoredTime    time.Time `json:"restored_time"`
	RestoredBy      string    `json:"restored_by"`
	RestorationPath string    `json:"restoration_path"`
	Success         bool      `json:"success"`
	ErrorMessage    string    `json:"error_message,omitempty"`
}

// BackupResult represents the result of a backup operation
type BackupResult struct {
	BackupPath      string        `json:"backup_path"`
	BackupSize      int64         `json:"backup_size"`
	FileCount       int           `json:"file_count"`
	BackupDuration  time.Duration `json:"backup_duration"`
	Verified        bool          `json:"verified"`
	Metadata        BackupMetadata `json:"metadata"`
	Errors          []string      `json:"errors"`
}

// RestoreResult represents the result of a restore operation
type RestoreResult struct {
	RestoredPath    string        `json:"restored_path"`
	RestoredSize    int64         `json:"restored_size"`
	FileCount       int           `json:"file_count"`
	RestoreDuration time.Duration `json:"restore_duration"`
	Success         bool          `json:"success"`
	Errors          []string      `json:"errors"`
}

// NewBackupManager creates a new backup manager
func NewBackupManager() *BackupManager {
	backupDir := filepath.Join("backups", "extensions")
	return &BackupManager{
		backupDirectory: backupDir,
		maxBackupAge:    90 * 24 * time.Hour, // 90 days
		maxBackupSize:   1024 * 1024 * 1024,  // 1GB
	}
}

// CreateExtensionBackup creates a comprehensive backup of extension data
func (bm *BackupManager) CreateExtensionBackup(extensionStorage scanner.ExtensionStorage, backupName string) (string, error) {
	// Ensure backup directory exists
	if err := os.MkdirAll(bm.backupDirectory, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Create backup path
	backupPath := filepath.Join(bm.backupDirectory, backupName+".zip")
	
	// Create backup metadata
	metadata := BackupMetadata{
		BackupID:        bm.generateBackupID(),
		ExtensionID:     extensionStorage.ExtensionID,
		CreationTime:    time.Now(),
		BackupType:      "extension_full",
		OriginalPath:    extensionStorage.StoragePath,
		BackupPath:      backupPath,
		CompressionType: "zip",
		BackupItems:     make([]BackupItem, 0),
	}

	// Create zip file
	zipFile, err := os.Create(backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Backup storage directory
	err = filepath.Walk(extensionStorage.StoragePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue despite errors
		}

		if info.IsDir() {
			return nil
		}

		// Calculate relative path
		relPath, err := filepath.Rel(extensionStorage.StoragePath, path)
		if err != nil {
			return nil // Skip files we can't process
		}

		// Create backup item
		backupItem, err := bm.createBackupItem(path, relPath, info, extensionStorage.StorageItems)
		if err != nil {
			return nil // Skip files we can't backup
		}

		// Add file to zip
		if err := bm.addFileToZip(zipWriter, path, relPath); err != nil {
			return nil // Skip files we can't add
		}

		metadata.BackupItems = append(metadata.BackupItems, *backupItem)
		metadata.TotalSize += info.Size()
		metadata.FileCount++

		return nil
	})

	if err != nil {
		os.Remove(backupPath)
		return "", fmt.Errorf("failed to create backup: %w", err)
	}

	// Calculate backup checksum
	zipWriter.Close()
	zipFile.Close()

	checksum, err := bm.calculateFileChecksum(backupPath)
	if err != nil {
		os.Remove(backupPath)
		return "", fmt.Errorf("failed to calculate backup checksum: %w", err)
	}
	metadata.Checksum = checksum

	// Save metadata
	metadataPath := strings.TrimSuffix(backupPath, ".zip") + ".metadata.json"
	if err := bm.saveBackupMetadata(metadata, metadataPath); err != nil {
		os.Remove(backupPath)
		return "", fmt.Errorf("failed to save backup metadata: %w", err)
	}

	return backupPath, nil
}

// BackupStorageItem creates a backup of a single storage item
func (bm *BackupManager) BackupStorageItem(item scanner.StorageDataItem) (string, error) {
	timestamp := time.Now().Unix()
	backupName := fmt.Sprintf("storage-item-%s-%d", 
		strings.ReplaceAll(item.Key, "/", "-"), 
		timestamp)

	backupPath := filepath.Join(bm.backupDirectory, "items", backupName+".json")
	
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(backupPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Create backup data
	backupData := map[string]interface{}{
		"key":           item.Key,
		"value":         item.Value,
		"size":          item.Size,
		"type":          item.Type,
		"risk":          item.Risk,
		"category":      item.Category,
		"description":   item.Description,
		"last_modified": item.LastModified,
		"backup_time":   time.Now(),
	}

	// Save backup
	data, err := json.MarshalIndent(backupData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal backup data: %w", err)
	}

	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write backup file: %w", err)
	}

	return backupPath, nil
}

// VerifyBackup verifies the integrity of a backup
func (bm *BackupManager) VerifyBackup(backupPath string) error {
	// Load metadata
	metadataPath := strings.TrimSuffix(backupPath, ".zip") + ".metadata.json"
	metadata, err := bm.loadBackupMetadata(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to load backup metadata: %w", err)
	}

	// Verify backup file exists
	if _, err := os.Stat(backupPath); err != nil {
		return fmt.Errorf("backup file not found: %w", err)
	}

	// Verify checksum
	currentChecksum, err := bm.calculateFileChecksum(backupPath)
	if err != nil {
		return fmt.Errorf("failed to calculate current checksum: %w", err)
	}

	if currentChecksum != metadata.Checksum {
		return fmt.Errorf("backup checksum mismatch: expected %s, got %s", 
			metadata.Checksum, currentChecksum)
	}

	// Verify zip file integrity
	if err := bm.verifyZipIntegrity(backupPath); err != nil {
		return fmt.Errorf("zip file integrity check failed: %w", err)
	}

	// Mark as verified
	metadata.Verified = true
	if err := bm.saveBackupMetadata(*metadata, metadataPath); err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}

	return nil
}

// RestoreBackup restores a backup to the specified location
func (bm *BackupManager) RestoreBackup(backupPath, restorePath string) (*RestoreResult, error) {
	startTime := time.Now()
	
	result := &RestoreResult{
		RestoredPath: restorePath,
		Errors:       make([]string, 0),
	}

	// Load metadata
	metadataPath := strings.TrimSuffix(backupPath, ".zip") + ".metadata.json"
	metadata, err := bm.loadBackupMetadata(metadataPath)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to load metadata: %v", err))
		return result, fmt.Errorf("failed to load backup metadata: %w", err)
	}

	// Verify backup before restoration
	if !metadata.Verified {
		if err := bm.VerifyBackup(backupPath); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Backup verification failed: %v", err))
			return result, fmt.Errorf("backup verification failed: %w", err)
		}
	}

	// Create restore directory
	if err := os.MkdirAll(restorePath, 0755); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to create restore directory: %v", err))
		return result, fmt.Errorf("failed to create restore directory: %w", err)
	}

	// Extract zip file
	if err := bm.extractZipFile(backupPath, restorePath); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to extract backup: %v", err))
		return result, fmt.Errorf("failed to extract backup: %w", err)
	}

	// Calculate restored size and file count
	err = filepath.Walk(restorePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			result.RestoredSize += info.Size()
			result.FileCount++
		}
		return nil
	})

	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to calculate restore statistics: %v", err))
	}

	// Update metadata with restoration info
	metadata.RestorationInfo = &RestorationInfo{
		RestoredTime:    time.Now(),
		RestoredBy:      "extension_cleaner",
		RestorationPath: restorePath,
		Success:         len(result.Errors) == 0,
	}

	if len(result.Errors) > 0 {
		metadata.RestorationInfo.ErrorMessage = strings.Join(result.Errors, "; ")
	}

	// Save updated metadata
	if err := bm.saveBackupMetadata(*metadata, metadataPath); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to update metadata: %v", err))
	}

	result.Success = len(result.Errors) == 0
	result.RestoreDuration = time.Since(startTime)

	return result, nil
}

// ListBackups returns a list of available backups
func (bm *BackupManager) ListBackups() ([]BackupMetadata, error) {
	var backups []BackupMetadata

	if _, err := os.Stat(bm.backupDirectory); os.IsNotExist(err) {
		return backups, nil // No backups directory
	}

	err := filepath.Walk(bm.backupDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue despite errors
		}

		if strings.HasSuffix(info.Name(), ".metadata.json") {
			metadata, err := bm.loadBackupMetadata(path)
			if err != nil {
				return nil // Skip invalid metadata files
			}
			backups = append(backups, *metadata)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list backups: %w", err)
	}

	return backups, nil
}

// CleanupOldBackups removes old backups based on age and size limits
func (bm *BackupManager) CleanupOldBackups() error {
	backups, err := bm.ListBackups()
	if err != nil {
		return fmt.Errorf("failed to list backups: %w", err)
	}

	now := time.Now()
	var totalSize int64

	// Calculate total backup size
	for _, backup := range backups {
		totalSize += backup.TotalSize
	}

	// Remove backups that are too old
	for _, backup := range backups {
		age := now.Sub(backup.CreationTime)
		
		shouldRemove := false
		
		// Remove if too old
		if age > bm.maxBackupAge {
			shouldRemove = true
		}
		
		// Remove if total size exceeds limit (remove oldest first)
		if totalSize > bm.maxBackupSize {
			shouldRemove = true
			totalSize -= backup.TotalSize
		}

		if shouldRemove {
			if err := bm.removeBackup(backup); err != nil {
				// Log error but continue with other backups
				continue
			}
		}
	}

	return nil
}

// Helper methods

// generateBackupID generates a unique backup ID
func (bm *BackupManager) generateBackupID() string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("backup-%d", timestamp)
}

// createBackupItem creates a backup item from file info
func (bm *BackupManager) createBackupItem(filePath, relativePath string, info os.FileInfo, storageItems []scanner.StorageDataItem) (*BackupItem, error) {
	checksum, err := bm.calculateFileChecksum(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate checksum: %w", err)
	}

	// Find corresponding storage item for risk assessment
	var risk scanner.TelemetryRisk = scanner.TelemetryRiskNone
	var itemType string = "file"

	for _, item := range storageItems {
		if strings.Contains(relativePath, item.Key) {
			risk = item.Risk
			itemType = item.Type
			break
		}
	}

	return &BackupItem{
		RelativePath: relativePath,
		OriginalPath: filePath,
		Size:         info.Size(),
		ModTime:      info.ModTime(),
		Checksum:     checksum,
		ItemType:     itemType,
		Risk:         risk,
	}, nil
}

// addFileToZip adds a file to a zip archive
func (bm *BackupManager) addFileToZip(zipWriter *zip.Writer, filePath, relativePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file info
	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Create zip file header
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return fmt.Errorf("failed to create zip header: %w", err)
	}

	header.Name = relativePath
	header.Method = zip.Deflate

	// Create writer for this file
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("failed to create zip writer: %w", err)
	}

	// Copy file content
	_, err = io.Copy(writer, file)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}

// calculateFileChecksum calculates MD5 checksum of a file
func (bm *BackupManager) calculateFileChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate hash: %w", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// saveBackupMetadata saves backup metadata to a JSON file
func (bm *BackupManager) saveBackupMetadata(metadata BackupMetadata, metadataPath string) error {
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}

// loadBackupMetadata loads backup metadata from a JSON file
func (bm *BackupManager) loadBackupMetadata(metadataPath string) (*BackupMetadata, error) {
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var metadata BackupMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &metadata, nil
}

// verifyZipIntegrity verifies that a zip file can be opened and read
func (bm *BackupManager) verifyZipIntegrity(zipPath string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer reader.Close()

	// Try to read each file in the zip
	for _, file := range reader.File {
		rc, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open file %s in zip: %w", file.Name, err)
		}
		
		// Read a small amount to verify the file is readable
		buffer := make([]byte, 1024)
		_, err = rc.Read(buffer)
		rc.Close()
		
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read file %s in zip: %w", file.Name, err)
		}
	}

	return nil
}

// extractZipFile extracts a zip file to the specified directory
func (bm *BackupManager) extractZipFile(zipPath, destPath string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer reader.Close()

	// Extract files
	for _, file := range reader.File {
		path := filepath.Join(destPath, file.Name)
		
		// Ensure the file path is within the destination directory
		if !strings.HasPrefix(path, filepath.Clean(destPath)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.FileInfo().Mode())
			continue
		}

		// Create directory for file
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		// Extract file
		if err := bm.extractFile(file, path); err != nil {
			return fmt.Errorf("failed to extract file %s: %w", file.Name, err)
		}
	}

	return nil
}

// extractFile extracts a single file from a zip archive
func (bm *BackupManager) extractFile(file *zip.File, destPath string) error {
	rc, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open file in zip: %w", err)
	}
	defer rc.Close()

	outFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.FileInfo().Mode())
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, rc)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}

// removeBackup removes a backup and its metadata
func (bm *BackupManager) removeBackup(backup BackupMetadata) error {
	// Remove backup file
	if err := os.Remove(backup.BackupPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove backup file: %w", err)
	}

	// Remove metadata file
	metadataPath := strings.TrimSuffix(backup.BackupPath, ".zip") + ".metadata.json"
	if err := os.Remove(metadataPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove metadata file: %w", err)
	}

	return nil
}

// GetBackupDirectory returns the backup directory path
func (bm *BackupManager) GetBackupDirectory() string {
	return bm.backupDirectory
}