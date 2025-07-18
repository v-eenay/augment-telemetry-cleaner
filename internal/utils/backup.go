package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// CreateBackup creates a backup of the specified file with timestamp
// Format: <filename>.bak.<timestamp>
func CreateBackup(filePath string) (string, error) {
	// Check if source file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("source file does not exist: %s", filePath)
	}

	// Generate backup path with timestamp
	timestamp := time.Now().Unix()
	backupPath := fmt.Sprintf("%s.bak.%d", filePath, timestamp)

	// Open source file
	sourceFile, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	// Create backup file
	backupFile, err := os.Create(backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %w", err)
	}
	defer backupFile.Close()

	// Copy file contents
	_, err = io.Copy(backupFile, sourceFile)
	if err != nil {
		// Clean up incomplete backup file
		os.Remove(backupPath)
		return "", fmt.Errorf("failed to copy file contents: %w", err)
	}

	// Copy file permissions
	sourceInfo, err := sourceFile.Stat()
	if err == nil {
		backupFile.Chmod(sourceInfo.Mode())
	}

	return backupPath, nil
}

// VerifyBackup verifies that a backup file exists and is readable
func VerifyBackup(backupPath string) error {
	info, err := os.Stat(backupPath)
	if err != nil {
		return fmt.Errorf("backup file not accessible: %w", err)
	}

	if info.Size() == 0 {
		return fmt.Errorf("backup file is empty")
	}

	// Try to open the file to ensure it's readable
	file, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("backup file not readable: %w", err)
	}
	file.Close()

	return nil
}

// CopyFile copies a file from source to destination
func CopyFile(src, dst string) error {
	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Open source file
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	// Create destination file
	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Copy file contents
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		// Clean up incomplete file
		os.Remove(dst)
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	// Copy file permissions
	sourceInfo, err := sourceFile.Stat()
	if err == nil {
		destFile.Chmod(sourceInfo.Mode())
	}

	return nil
}
