package cleaner

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"augment-telemetry-cleaner/internal/utils"
)

// WorkspaceCleanResult contains the results of workspace cleaning operation
type WorkspaceCleanResult struct {
	BackupPath           string                    `json:"backup_path"`
	DeletedFilesCount    int                       `json:"deleted_files_count"`
	FailedOperations     []FailedOperation         `json:"failed_operations,omitempty"`
	FailedCompressions   []FailedCompression       `json:"failed_compressions,omitempty"`
}

// FailedOperation represents a failed file/directory operation
type FailedOperation struct {
	Type  string `json:"type"`  // "file" or "directory"
	Path  string `json:"path"`
	Error string `json:"error"`
}

// FailedCompression represents a failed compression operation
type FailedCompression struct {
	File  string `json:"file"`
	Error string `json:"error"`
}

// CleanWorkspaceStorage cleans the workspace storage directory after creating a backup
//
// This function:
// 1. Gets the workspace storage path
// 2. Creates a zip backup of all files in the directory
// 3. Deletes all files in the directory
func CleanWorkspaceStorage() (*WorkspaceCleanResult, error) {
	workspacePath, err := utils.GetWorkspaceStoragePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace storage path: %w", err)
	}

	// Check if workspace directory exists
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("workspace storage directory not found at: %s", workspacePath)
	}

	// Create backup filename with timestamp
	timestamp := time.Now().Unix()
	backupPath := fmt.Sprintf("%s_backup_%d.zip", workspacePath, timestamp)

	// Create zip backup
	failedCompressions, err := createZipBackup(workspacePath, backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup: %w", err)
	}

	// Count files before deletion
	totalFiles, err := countFiles(workspacePath)
	if err != nil {
		return nil, fmt.Errorf("failed to count files: %w", err)
	}

	// Delete all files in the directory
	failedOperations, err := deleteWorkspaceContents(workspacePath)
	if err != nil {
		return nil, fmt.Errorf("failed to delete workspace contents: %w", err)
	}

	return &WorkspaceCleanResult{
		BackupPath:         backupPath,
		DeletedFilesCount:  totalFiles,
		FailedOperations:   failedOperations,
		FailedCompressions: failedCompressions,
	}, nil
}

// createZipBackup creates a zip backup of the workspace directory
func createZipBackup(workspacePath, backupPath string) ([]FailedCompression, error) {
	var failedCompressions []FailedCompression

	zipFile, err := os.Create(backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	err = filepath.Walk(workspacePath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			failedCompressions = append(failedCompressions, FailedCompression{
				File:  filePath,
				Error: err.Error(),
			})
			return nil // Continue walking
		}

		// Skip the root directory itself
		if filePath == workspacePath {
			return nil
		}

		// Get relative path for the zip archive
		relPath, err := filepath.Rel(workspacePath, filePath)
		if err != nil {
			failedCompressions = append(failedCompressions, FailedCompression{
				File:  filePath,
				Error: fmt.Sprintf("failed to get relative path: %v", err),
			})
			return nil
		}

		// Normalize path separators for zip archive
		relPath = strings.ReplaceAll(relPath, "\\", "/")

		if info.IsDir() {
			// Create directory entry in zip
			_, err := zipWriter.Create(relPath + "/")
			if err != nil {
				failedCompressions = append(failedCompressions, FailedCompression{
					File:  filePath,
					Error: fmt.Sprintf("failed to create directory entry: %v", err),
				})
			}
			return nil
		}

		// Add file to zip
		err = addFileToZip(zipWriter, filePath, relPath)
		if err != nil {
			failedCompressions = append(failedCompressions, FailedCompression{
				File:  filePath,
				Error: err.Error(),
			})
		}

		return nil
	})

	if err != nil {
		return failedCompressions, fmt.Errorf("failed to walk directory: %w", err)
	}

	return failedCompressions, nil
}

// addFileToZip adds a single file to the zip archive
func addFileToZip(zipWriter *zip.Writer, filePath, relPath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	zipEntry, err := zipWriter.Create(relPath)
	if err != nil {
		return fmt.Errorf("failed to create zip entry: %w", err)
	}

	_, err = io.Copy(zipEntry, file)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}

// countFiles counts the total number of files in the directory
func countFiles(dirPath string) (int, error) {
	count := 0
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue counting despite errors
		}
		if !info.IsDir() {
			count++
		}
		return nil
	})
	return count, err
}

// deleteWorkspaceContents deletes all contents of the workspace directory
func deleteWorkspaceContents(workspacePath string) ([]FailedOperation, error) {
	var failedOperations []FailedOperation

	// First, try to remove the entire directory tree
	err := os.RemoveAll(workspacePath)
	if err == nil {
		// If successful, recreate the empty directory
		return failedOperations, os.MkdirAll(workspacePath, 0755)
	}

	// If bulk removal failed, try file-by-file approach
	err = filepath.Walk(workspacePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			failedOperations = append(failedOperations, FailedOperation{
				Type:  "unknown",
				Path:  path,
				Error: err.Error(),
			})
			return nil // Continue walking
		}

		// Skip the root directory itself
		if path == workspacePath {
			return nil
		}

		if info.IsDir() {
			// We'll handle directories after files
			return nil
		}

		// Delete file
		err = deleteFile(path)
		if err != nil {
			failedOperations = append(failedOperations, FailedOperation{
				Type:  "file",
				Path:  path,
				Error: err.Error(),
			})
		}

		return nil
	})

	if err != nil {
		return failedOperations, fmt.Errorf("failed to walk directory for deletion: %w", err)
	}

	// Now delete directories from deepest to shallowest
	var directories []string
	filepath.Walk(workspacePath, func(path string, info os.FileInfo, err error) error {
		if err == nil && info.IsDir() && path != workspacePath {
			directories = append(directories, path)
		}
		return nil
	})

	// Sort directories by depth (deepest first)
	for i := len(directories) - 1; i >= 0; i-- {
		err := os.Remove(directories[i])
		if err != nil {
			failedOperations = append(failedOperations, FailedOperation{
				Type:  "directory",
				Path:  directories[i],
				Error: err.Error(),
			})
		}
	}

	return failedOperations, nil
}

// deleteFile attempts to delete a file, handling read-only files on Windows
func deleteFile(filePath string) error {
	err := os.Remove(filePath)
	if err != nil && runtime.GOOS == "windows" {
		// Try to remove read-only attribute and delete again
		if err := os.Chmod(filePath, 0666); err == nil {
			err = os.Remove(filePath)
		}
	}
	return err
}
