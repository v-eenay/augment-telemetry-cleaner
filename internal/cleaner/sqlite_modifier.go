package cleaner

import (
	"database/sql"
	"fmt"
	"os"

	"augment-telemetry-cleaner/internal/utils"
	_ "github.com/mattn/go-sqlite3"
)

// DatabaseCleanResult contains the results of database cleaning operation
type DatabaseCleanResult struct {
	DBBackupPath string `json:"db_backup_path"`
	DeletedRows  int64  `json:"deleted_rows"`
}

// CleanAugmentData cleans augment-related data from the SQLite database
// Creates a backup before modification
//
// This function:
// 1. Gets the SQLite database path
// 2. Creates a backup of the database file
// 3. Opens the database connection
// 4. Deletes records where key contains 'augment'
func CleanAugmentData() (*DatabaseCleanResult, error) {
	dbPath, err := utils.GetDBPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get database path: %w", err)
	}

	// Check if database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("database file not found at: %s", dbPath)
	}

	// Create backup before modification
	dbBackupPath, err := utils.CreateBackup(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create database backup: %w", err)
	}

	// Verify backup was created successfully
	if err := utils.VerifyBackup(dbBackupPath); err != nil {
		return nil, fmt.Errorf("backup verification failed: %w", err)
	}

	// Connect to the database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Begin transaction for safety
	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Will be ignored if tx.Commit() succeeds

	// Execute the delete query
	result, err := tx.Exec("DELETE FROM ItemTable WHERE key LIKE '%augment%'")
	if err != nil {
		return nil, fmt.Errorf("failed to execute delete query: %w", err)
	}

	// Get the number of affected rows
	deletedRows, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get affected rows count: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &DatabaseCleanResult{
		DBBackupPath: dbBackupPath,
		DeletedRows:  deletedRows,
	}, nil
}

// GetAugmentDataCount returns the count of records containing 'augment' in their keys
// This can be used for dry-run mode to show what would be deleted
func GetAugmentDataCount() (int64, error) {
	dbPath, err := utils.GetDBPath()
	if err != nil {
		return 0, fmt.Errorf("failed to get database path: %w", err)
	}

	// Check if database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return 0, fmt.Errorf("database file not found at: %s", dbPath)
	}

	// Connect to the database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Test the connection
	if err := db.Ping(); err != nil {
		return 0, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Count records that would be deleted
	var count int64
	err = db.QueryRow("SELECT COUNT(*) FROM ItemTable WHERE key LIKE '%augment%'").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count records: %w", err)
	}

	return count, nil
}
