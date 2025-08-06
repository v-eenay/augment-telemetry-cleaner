package utils

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCreateBackup(t *testing.T) {
	// Create a temporary file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "This is a test file for backup testing"
	
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create backup
	backupPath, err := CreateBackup(testFile)
	if err != nil {
		t.Fatalf("CreateBackup() failed: %v", err)
	}

	// Verify backup file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Errorf("Backup file was not created: %s", backupPath)
	}

	// Verify backup content matches original
	backupContent, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup file: %v", err)
	}

	if string(backupContent) != testContent {
		t.Errorf("Backup content doesn't match original. Expected: %s, Got: %s", testContent, string(backupContent))
	}

	// Verify backup path format
	expectedPrefix := testFile + ".bak."
	if !startsWith(backupPath, expectedPrefix) {
		t.Errorf("Backup path doesn't have expected format. Expected prefix: %s, Got: %s", expectedPrefix, backupPath)
	}
}

func TestCreateBackupNonExistentFile(t *testing.T) {
	// Try to backup a non-existent file
	nonExistentFile := "/path/that/does/not/exist/file.txt"
	
	_, err := CreateBackup(nonExistentFile)
	if err == nil {
		t.Error("CreateBackup() should fail for non-existent file")
	}
}

func TestVerifyBackup(t *testing.T) {
	// Create a temporary file and backup
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "Test content for verification"
	
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	backupPath, err := CreateBackup(testFile)
	if err != nil {
		t.Fatalf("CreateBackup() failed: %v", err)
	}

	// Verify the backup
	err = VerifyBackup(backupPath)
	if err != nil {
		t.Errorf("VerifyBackup() failed: %v", err)
	}
}

func TestVerifyBackupNonExistentFile(t *testing.T) {
	// Try to verify a non-existent backup
	nonExistentBackup := "/path/that/does/not/exist/backup.bak"
	
	err := VerifyBackup(nonExistentBackup)
	if err == nil {
		t.Error("VerifyBackup() should fail for non-existent file")
	}
}

func TestVerifyBackupEmptyFile(t *testing.T) {
	// Create an empty backup file
	tempDir := t.TempDir()
	emptyBackup := filepath.Join(tempDir, "empty.bak")
	
	err := os.WriteFile(emptyBackup, []byte{}, 0644)
	if err != nil {
		t.Fatalf("Failed to create empty backup file: %v", err)
	}

	// Verify should fail for empty file
	err = VerifyBackup(emptyBackup)
	if err == nil {
		t.Error("VerifyBackup() should fail for empty file")
	}
}

func TestBackupTimestamp(t *testing.T) {
	// Create two backups with a small delay
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	
	err := os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	backup1, err := CreateBackup(testFile)
	if err != nil {
		t.Fatalf("First CreateBackup() failed: %v", err)
	}

	// Small delay to ensure different timestamps
	time.Sleep(time.Second)

	backup2, err := CreateBackup(testFile)
	if err != nil {
		t.Fatalf("Second CreateBackup() failed: %v", err)
	}

	// Verify backups have different names (due to timestamps)
	if backup1 == backup2 {
		t.Errorf("Two backups created at different times should have different names")
	}
}

// Helper function
func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
