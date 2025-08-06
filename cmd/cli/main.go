package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"augment-telemetry-cleaner/internal/browser"
	"augment-telemetry-cleaner/internal/cleaner"
	"augment-telemetry-cleaner/internal/config"
	"augment-telemetry-cleaner/internal/logger"
)

// CLI represents the command-line interface
type CLI struct {
	configManager *config.ConfigManager
	logger        *logger.Logger
	fileLogger    *log.Logger
	logLevel      int
	config        *CLIConfig
}

// CLIConfig holds CLI-specific configuration
type CLIConfig struct {
	DryRun         bool
	Verbose        bool
	CreateBackups  bool
	NoConfirm      bool
	TargetBrowser  string
	Operation      string
	OutputFormat   string
	LogLevel       string
}

// Operation constants
const (
	OpModifyTelemetry = "modify-telemetry"
	OpCleanDatabase   = "clean-database"
	OpCleanWorkspace  = "clean-workspace"
	OpCleanBrowser    = "clean-browser"
	OpRunAll          = "run-all"
)

func main() {
	cli := &CLI{
		config: &CLIConfig{},
	}

	if err := cli.parseFlags(); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	if err := cli.initialize(); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing CLI: %v\n", err)
		os.Exit(1)
	}

	if err := cli.run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running operation: %v\n", err)
		os.Exit(1)
	}
}

// parseFlags parses command-line flags
func (c *CLI) parseFlags() error {
	var noBackup bool

	flag.StringVar(&c.config.Operation, "operation", "", "Operation to perform: modify-telemetry, clean-database, clean-workspace, clean-browser, run-all")
	flag.BoolVar(&c.config.DryRun, "dry-run", false, "Preview operations without making changes")
	flag.BoolVar(&c.config.Verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&c.config.CreateBackups, "backup", true, "Create backups before operations")
	flag.BoolVar(&noBackup, "no-backup", false, "Disable backup creation")
	flag.BoolVar(&c.config.NoConfirm, "no-confirm", false, "Skip confirmation prompts")
	flag.StringVar(&c.config.TargetBrowser, "browser", "", "Target specific browser: chrome, firefox, edge, safari (for browser operations)")
	flag.StringVar(&c.config.OutputFormat, "output", "text", "Output format: text, json")
	flag.StringVar(&c.config.LogLevel, "log-level", "INFO", "Log level: DEBUG, INFO, WARN, ERROR")

	// Custom help
	flag.Usage = c.printUsage

	flag.Parse()

	// Handle no-backup flag
	if noBackup {
		c.config.CreateBackups = false
	}

	// Validate operation
	if c.config.Operation == "" {
		return fmt.Errorf("operation is required. Use --help for usage information")
	}

	validOps := []string{OpModifyTelemetry, OpCleanDatabase, OpCleanWorkspace, OpCleanBrowser, OpRunAll}
	valid := false
	for _, op := range validOps {
		if c.config.Operation == op {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid operation: %s. Valid operations: %s", c.config.Operation, strings.Join(validOps, ", "))
	}

	return nil
}

// printUsage prints usage information
func (c *CLI) printUsage() {
	fmt.Fprintf(os.Stderr, `Augment Telemetry Cleaner CLI v2.0.0

USAGE:
    augment-telemetry-cleaner-cli --operation <operation> [options]

OPERATIONS:
    modify-telemetry    Modify VS Code telemetry IDs
    clean-database      Clean Augment data from VS Code database
    clean-workspace     Clean VS Code workspace storage
    clean-browser       Clean Augment data from browsers
    run-all            Run all cleaning operations

OPTIONS:
    --operation <op>        Operation to perform (required)
    --dry-run              Preview operations without making changes
    --verbose              Enable verbose output
    --backup               Create backups before operations (default: true)
    --no-backup            Disable backup creation
    --no-confirm           Skip confirmation prompts
    --browser <browser>    Target specific browser for browser operations
    --output <format>      Output format: text, json (default: text)
    --log-level <level>    Log level: DEBUG, INFO, WARN, ERROR (default: INFO)
    --help                 Show this help message

EXAMPLES:
    # Preview all operations
    augment-telemetry-cleaner-cli --operation run-all --dry-run

    # Clean database with verbose output
    augment-telemetry-cleaner-cli --operation clean-database --verbose

    # Clean Chrome browser data without confirmation
    augment-telemetry-cleaner-cli --operation clean-browser --browser chrome --no-confirm

    # Modify telemetry IDs without creating backups
    augment-telemetry-cleaner-cli --operation modify-telemetry --no-backup

SAFETY FEATURES:
    - Dry-run mode for safe preview
    - Automatic backup creation (unless disabled)
    - Confirmation prompts (unless disabled)
    - Comprehensive logging

WARNING:
    This application may log you out of other browser extensions and accounts,
    but Augment will continue to work properly even with a new email account
    after running this tool.
`)
}

// initialize initializes the CLI components
func (c *CLI) initialize() error {
	// Initialize configuration manager
	configManager, err := config.NewConfigManager()
	if err != nil {
		return fmt.Errorf("failed to initialize configuration: %w", err)
	}
	c.configManager = configManager

	// Update config based on CLI flags
	err = c.configManager.UpdateConfig(func(config *config.Config) {
		config.DryRunMode = c.config.DryRun
		config.CreateBackups = c.config.CreateBackups
		config.RequireConfirmation = !c.config.NoConfirm
	})
	if err != nil {
		return fmt.Errorf("failed to update configuration: %w", err)
	}

	// Initialize simple file logger
	logDir := "logs"
	fileLogger, err := c.createSimpleFileLogger(logDir)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	c.fileLogger = fileLogger
	c.logger = nil // Keep this nil to avoid stdout conflicts

	// Store log level for our simple logger
	c.logLevel = c.parseLogLevel(c.config.LogLevel)

	return nil
}

// onLogMessage handles log messages
func (c *CLI) onLogMessage(level logger.LogLevel, message string) {
	if c.config.Verbose || level >= 2 { // WARN level
		fmt.Printf("[%s] %s\n", level.String(), message)
	}
}

// Helper methods for safe logging
func (c *CLI) logOperation(operation string) {
	if c.fileLogger != nil {
		c.fileLogger.Printf("[INFO] === Starting operation: %s ===", operation)
	}
}

func (c *CLI) logInfo(format string, args ...interface{}) {
	if c.fileLogger != nil && c.logLevel <= 1 { // INFO level
		c.fileLogger.Printf("[INFO] "+format, args...)
	}
}

func (c *CLI) logError(format string, args ...interface{}) {
	if c.fileLogger != nil && c.logLevel <= 3 { // ERROR level
		c.fileLogger.Printf("[ERROR] "+format, args...)
	}
}

func (c *CLI) logOperationResult(operation string, success bool, details string) {
	if c.fileLogger != nil {
		if success {
			c.fileLogger.Printf("[INFO] === Operation completed successfully: %s ===", operation)
			if details != "" {
				c.fileLogger.Printf("[INFO] Details: %s", details)
			}
		} else {
			c.fileLogger.Printf("[ERROR] === Operation failed: %s ===", operation)
			if details != "" {
				c.fileLogger.Printf("[ERROR] Error details: %s", details)
			}
		}
	}
}

func (c *CLI) logBackupCreated(originalPath, backupPath string) {
	if c.fileLogger != nil {
		c.fileLogger.Printf("[INFO] Backup created: %s -> %s", originalPath, backupPath)
	}
}

// run executes the specified operation
func (c *CLI) run() error {
	// Note: fileLogger doesn't need explicit closing as it's handled by the OS
	// when the program exits, but we could add it if needed

	c.printHeader()

	switch c.config.Operation {
	case OpModifyTelemetry:
		return c.runModifyTelemetry()
	case OpCleanDatabase:
		return c.runCleanDatabase()
	case OpCleanWorkspace:
		return c.runCleanWorkspace()
	case OpCleanBrowser:
		return c.runCleanBrowser()
	case OpRunAll:
		return c.runAllOperations()
	default:
		return fmt.Errorf("unknown operation: %s", c.config.Operation)
	}
}

// printHeader prints the application header
func (c *CLI) printHeader() {
	fmt.Println("=== Augment Telemetry Cleaner CLI v2.0.0 ===")
	fmt.Printf("Operation: %s\n", c.config.Operation)
	if c.config.DryRun {
		fmt.Println("Mode: DRY RUN (Preview only)")
	} else {
		fmt.Println("Mode: LIVE (Making actual changes)")
	}
	fmt.Printf("Backups: %t\n", c.config.CreateBackups)
	fmt.Println("==========================================")
	fmt.Println()
}

// runModifyTelemetry executes the telemetry modification operation
func (c *CLI) runModifyTelemetry() error {
	c.logOperation("Modify Telemetry IDs")
	fmt.Println("🔧 Modifying VS Code telemetry IDs...")

	if c.config.DryRun {
		fmt.Println("DRY RUN: Would modify telemetry IDs in VS Code storage")
		c.logInfo("DRY RUN MODE: Would modify telemetry IDs")
		return nil
	}

	if !c.config.NoConfirm {
		if !c.confirmOperation("modify VS Code telemetry IDs") {
			fmt.Println("Operation cancelled by user")
			return nil
		}
	}

	result, err := cleaner.ModifyTelemetryIDs()
	if err != nil {
		c.logOperationResult("Modify Telemetry IDs", false, err.Error())
		return fmt.Errorf("telemetry modification failed: %w", err)
	}

	c.logOperationResult("Modify Telemetry IDs", true, "Telemetry IDs modified successfully")
	c.logBackupCreated("storage.json", result.StorageBackupPath)

	return c.printResult("Telemetry Modification", result)
}

// runCleanDatabase executes the database cleaning operation
func (c *CLI) runCleanDatabase() error {
	c.logOperation("Clean Database")
	fmt.Println("🗃️ Cleaning VS Code database...")

	if c.config.DryRun {
		count, err := cleaner.GetAugmentDataCount()
		if err != nil {
			return fmt.Errorf("failed to count database records: %w", err)
		}
		fmt.Printf("DRY RUN: Would delete %d database records\n", count)
		c.logInfo("DRY RUN MODE: Would delete %d database records", count)
		return nil
	}

	if !c.config.NoConfirm {
		if !c.confirmOperation("clean Augment data from VS Code database") {
			fmt.Println("Operation cancelled by user")
			return nil
		}
	}

	result, err := cleaner.CleanAugmentData()
	if err != nil {
		c.logOperationResult("Clean Database", false, err.Error())
		return fmt.Errorf("database cleaning failed: %w", err)
	}

	c.logOperationResult("Clean Database", true, fmt.Sprintf("Deleted %d records", result.DeletedRows))
	c.logBackupCreated("database", result.DBBackupPath)

	return c.printResult("Database Cleaning", result)
}

// runCleanWorkspace executes the workspace cleaning operation
func (c *CLI) runCleanWorkspace() error {
	c.logOperation("Clean Workspace")
	fmt.Println("💾 Cleaning VS Code workspace storage...")

	if c.config.DryRun {
		fmt.Println("DRY RUN: Would clean VS Code workspace storage")
		c.logInfo("DRY RUN MODE: Would clean workspace storage")
		return nil
	}

	if !c.config.NoConfirm {
		if !c.confirmOperation("clean VS Code workspace storage") {
			fmt.Println("Operation cancelled by user")
			return nil
		}
	}

	result, err := cleaner.CleanWorkspaceStorage()
	if err != nil {
		c.logOperationResult("Clean Workspace", false, err.Error())
		return fmt.Errorf("workspace cleaning failed: %w", err)
	}

	c.logOperationResult("Clean Workspace", true, fmt.Sprintf("Deleted %d files", result.DeletedFilesCount))
	c.logBackupCreated("workspace", result.BackupPath)

	return c.printResult("Workspace Cleaning", result)
}

// runCleanBrowser executes the browser cleaning operation
func (c *CLI) runCleanBrowser() error {
	c.logOperation("Clean Browser Data")
	fmt.Println("🌐 Cleaning browser data...")

	if c.config.DryRun {
		browserCleaner, err := browser.NewBrowserCleaner()
		if err != nil {
			return fmt.Errorf("failed to create browser cleaner: %w", err)
		}

		counts, err := browserCleaner.GetBrowserDataCount()
		if err != nil {
			return fmt.Errorf("failed to count browser data: %w", err)
		}

		totalCount := int64(0)
		for _, count := range counts {
			totalCount += count
		}

		fmt.Printf("DRY RUN: Would clean %d browser data items\n", totalCount)
		c.logInfo("DRY RUN MODE: Would clean %d browser data items", totalCount)
		return nil
	}

	if !c.config.NoConfirm {
		fmt.Println("⚠️  WARNING: Please close all browsers before proceeding.")
		fmt.Println("This operation will clean:")
		fmt.Println("  • Augment-related cookies and domains")
		fmt.Println("  • Local storage data containing Augment patterns")
		fmt.Println("  • Session storage with Augment identifiers")
		fmt.Println("  • Cache files with Augment references")
		fmt.Println()

		if !c.confirmOperation("clean browser data") {
			fmt.Println("Operation cancelled by user")
			return nil
		}
	}

	browserCleaner, err := browser.NewBrowserCleaner()
	if err != nil {
		c.logOperationResult("Clean Browser Data", false, err.Error())
		return fmt.Errorf("browser cleaner creation failed: %w", err)
	}

	results, err := browserCleaner.CleanBrowserData(c.config.CreateBackups)
	if err != nil {
		c.logOperationResult("Clean Browser Data", false, err.Error())
		return fmt.Errorf("browser cleaning failed: %w", err)
	}

	// Process results
	totalCookies := int64(0)
	totalStorage := int64(0)
	totalCache := int64(0)
	var allErrors []string

	for _, result := range results {
		totalCookies += result.CookiesDeleted
		totalStorage += result.StorageDeleted
		totalCache += result.CacheDeleted

		if result.BackupPath != "" {
			c.logBackupCreated("browser-"+result.Profile.Name, result.BackupPath)
		}

		for _, err := range result.Errors {
			allErrors = append(allErrors, fmt.Sprintf("%s: %s", result.Profile.Name, err))
		}
	}

	// Log results
	successMsg := fmt.Sprintf("Cleaned %d cookies, %d storage items, %d cache items", totalCookies, totalStorage, totalCache)
	c.logOperationResult("Clean Browser Data", len(allErrors) == 0, successMsg)

	// Log any errors
	for _, err := range allErrors {
		c.logError("Browser cleaning error: %s", err)
	}

	return c.printResult("Browser Cleaning", results)
}

// runAllOperations executes all cleaning operations in sequence
func (c *CLI) runAllOperations() error {
	c.logOperation("Run All Operations")
	fmt.Println("🚀 Running all cleaning operations...")

	if c.config.DryRun {
		fmt.Println("DRY RUN: Would run all cleaning operations")
		c.logInfo("DRY RUN MODE: Would run all operations")
		return nil
	}

	if !c.config.NoConfirm {
		fmt.Println("This will run all cleaning operations:")
		fmt.Println("  1. Modify telemetry IDs")
		fmt.Println("  2. Clean database")
		fmt.Println("  3. Clean workspace")
		fmt.Println("  4. Clean browser data")
		fmt.Println()

		if !c.confirmOperation("run all cleaning operations") {
			fmt.Println("Operation cancelled by user")
			return nil
		}
	}

	operations := []struct {
		name string
		fn   func() error
	}{
		{"Modify Telemetry IDs", c.runModifyTelemetryInternal},
		{"Clean Database", c.runCleanDatabaseInternal},
		{"Clean Workspace", c.runCleanWorkspaceInternal},
		{"Clean Browser Data", c.runCleanBrowserInternal},
	}

	for i, op := range operations {
		fmt.Printf("Step %d/4: %s...\n", i+1, op.name)
		if err := op.fn(); err != nil {
			c.logError("Operation failed: %s - %v", op.name, err)
			fmt.Printf("❌ %s failed: %v\n", op.name, err)
			continue
		}
		fmt.Printf("✅ %s completed\n", op.name)
	}

	fmt.Println("\n🎉 All operations completed!")
	c.logOperationResult("Run All Operations", true, "All operations completed")
	return nil
}

// Internal operation methods (without confirmation prompts)
func (c *CLI) runModifyTelemetryInternal() error {
	result, err := cleaner.ModifyTelemetryIDs()
	if err != nil {
		c.logError("Telemetry modification failed: %v", err)
		return err
	}
	c.logInfo("Telemetry IDs modified successfully")
	c.logBackupCreated("storage.json", result.StorageBackupPath)
	return nil
}

func (c *CLI) runCleanDatabaseInternal() error {
	result, err := cleaner.CleanAugmentData()
	if err != nil {
		c.logError("Database cleaning failed: %v", err)
		return err
	}
	c.logInfo("Database cleaned successfully, deleted %d records", result.DeletedRows)
	c.logBackupCreated("database", result.DBBackupPath)
	return nil
}

func (c *CLI) runCleanWorkspaceInternal() error {
	result, err := cleaner.CleanWorkspaceStorage()
	if err != nil {
		c.logError("Workspace cleaning failed: %v", err)
		return err
	}
	c.logInfo("Workspace cleaned successfully, deleted %d files", result.DeletedFilesCount)
	c.logBackupCreated("workspace", result.BackupPath)
	return nil
}

func (c *CLI) runCleanBrowserInternal() error {
	browserCleaner, err := browser.NewBrowserCleaner()
	if err != nil {
		c.logError("Browser cleaner creation failed: %v", err)
		return err
	}

	results, err := browserCleaner.CleanBrowserData(c.config.CreateBackups)
	if err != nil {
		c.logError("Browser cleaning failed: %v", err)
		return err
	}

	// Count total items cleaned
	totalItems := int64(0)
	for _, result := range results {
		totalItems += result.CookiesDeleted + result.StorageDeleted + result.CacheDeleted
		if result.BackupPath != "" {
			c.logBackupCreated("browser-"+result.Profile.Name, result.BackupPath)
		}
	}

	c.logInfo("Browser data cleaned successfully, processed %d items", totalItems)
	return nil
}

// confirmOperation prompts the user for confirmation
func (c *CLI) confirmOperation(operation string) bool {
	fmt.Printf("Are you sure you want to %s? [y/N]: ", operation)
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// printResult prints the operation result
func (c *CLI) printResult(operationName string, result interface{}) error {
	fmt.Printf("\n✅ %s completed successfully!\n", operationName)

	if c.config.OutputFormat == "json" {
		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal result to JSON: %w", err)
		}
		fmt.Println("\nResult Details (JSON):")
		fmt.Println(string(jsonData))
	} else {
		fmt.Println("\nResult Details:")
		c.printTextResult(result)
	}

	return nil
}

// printTextResult prints the result in human-readable text format
func (c *CLI) printTextResult(result interface{}) {
	switch r := result.(type) {
	case *cleaner.TelemetryModifyResult:
		fmt.Printf("  Old Machine ID: %s\n", r.OldMachineID)
		fmt.Printf("  New Machine ID: %s\n", r.NewMachineID)
		fmt.Printf("  Old Device ID: %s\n", r.OldDeviceID)
		fmt.Printf("  New Device ID: %s\n", r.NewDeviceID)
		if r.StorageBackupPath != "" {
			fmt.Printf("  Storage Backup: %s\n", r.StorageBackupPath)
		}
		if r.MachineIDBackupPath != "" {
			fmt.Printf("  Machine ID Backup: %s\n", r.MachineIDBackupPath)
		}

	case *cleaner.DatabaseCleanResult:
		fmt.Printf("  Records Deleted: %d\n", r.DeletedRows)
		if r.DBBackupPath != "" {
			fmt.Printf("  Database Backup: %s\n", r.DBBackupPath)
		}

	case *cleaner.WorkspaceCleanResult:
		fmt.Printf("  Files Deleted: %d\n", r.DeletedFilesCount)
		if r.BackupPath != "" {
			fmt.Printf("  Workspace Backup: %s\n", r.BackupPath)
		}
		if len(r.FailedOperations) > 0 {
			fmt.Printf("  Failed Operations: %d\n", len(r.FailedOperations))
		}

	case []browser.BrowserCleanResult:
		totalCookies := int64(0)
		totalStorage := int64(0)
		totalCache := int64(0)
		totalErrors := 0

		for _, result := range r {
			totalCookies += result.CookiesDeleted
			totalStorage += result.StorageDeleted
			totalCache += result.CacheDeleted
			totalErrors += len(result.Errors)

			fmt.Printf("  Browser: %s (%s)\n", result.Profile.Name, result.Profile.Type.String())
			fmt.Printf("    Cookies Deleted: %d\n", result.CookiesDeleted)
			fmt.Printf("    Storage Items Deleted: %d\n", result.StorageDeleted)
			fmt.Printf("    Cache Items Deleted: %d\n", result.CacheDeleted)
			if result.BackupPath != "" {
				fmt.Printf("    Backup: %s\n", result.BackupPath)
			}
			if len(result.Errors) > 0 {
				fmt.Printf("    Errors: %d\n", len(result.Errors))
			}
		}

		fmt.Printf("  Total Summary:\n")
		fmt.Printf("    Total Cookies Deleted: %d\n", totalCookies)
		fmt.Printf("    Total Storage Items Deleted: %d\n", totalStorage)
		fmt.Printf("    Total Cache Items Deleted: %d\n", totalCache)
		if totalErrors > 0 {
			fmt.Printf("    Total Errors: %d\n", totalErrors)
		}

	default:
		fmt.Printf("  Result: %+v\n", result)
	}
}

// createSimpleFileLogger creates a simple file-only logger
func (c *CLI) createSimpleFileLogger(logDir string) (*log.Logger, error) {
	// Create log directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create log file with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFile := filepath.Join(logDir, fmt.Sprintf("augment_cleaner_cli_%s.log", timestamp))

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Create logger that only writes to file
	fileLogger := log.New(file, "", log.LstdFlags)
	fileLogger.Println("[INFO] CLI Logger initialized")

	return fileLogger, nil
}

// parseLogLevel converts string log level to integer
func (c *CLI) parseLogLevel(level string) int {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return 0
	case "INFO":
		return 1
	case "WARN":
		return 2
	case "ERROR":
		return 3
	default:
		return 1 // Default to INFO
	}
}


