package gui

import (
	"encoding/json"
	"fmt"

	"fyne.io/fyne/v2/dialog"

	"augment-telemetry-cleaner/internal/browser"
	"augment-telemetry-cleaner/internal/cleaner"
	"augment-telemetry-cleaner/internal/utils"
)

// runModifyTelemetry executes the telemetry modification operation
func (g *MainGUI) runModifyTelemetry() {
	g.setOperationState(true, "Modifying telemetry IDs...")
	defer g.setOperationState(false, "Ready")

	config := g.configManager.GetConfig()
	g.logger.LogOperation("Modify Telemetry IDs")

	if config.DryRunMode {
		g.logger.Info("DRY RUN MODE: Would modify telemetry IDs")
		g.setResults("DRY RUN: Telemetry IDs would be modified (no actual changes made)")
		return
	}

	result, err := cleaner.ModifyTelemetryIDs()
	if err != nil {
		g.logger.LogOperationResult("Modify Telemetry IDs", false, err.Error())
		g.showErrorDialog("Telemetry Modification Failed", err.Error())
		return
	}

	g.logger.LogOperationResult("Modify Telemetry IDs", true, "")
	g.logger.LogBackupCreated("storage.json", result.StorageBackupPath)
	if result.MachineIDBackupPath != "" {
		g.logger.LogBackupCreated("machineid", result.MachineIDBackupPath)
	}

	// Display results
	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	g.setResults(fmt.Sprintf("Telemetry IDs Modified Successfully:\n%s", string(resultJSON)))
}

// runCleanDatabase executes the database cleaning operation
func (g *MainGUI) runCleanDatabase() {
	g.setOperationState(true, "Cleaning database...")
	defer g.setOperationState(false, "Ready")

	config := g.configManager.GetConfig()
	g.logger.LogOperation("Clean Database")

	if config.DryRunMode {
		count, err := cleaner.GetAugmentDataCount()
		if err != nil {
			g.logger.Error("Failed to count database records: %v", err)
			g.showErrorDialog("Database Count Failed", err.Error())
			return
		}
		g.logger.Info("DRY RUN MODE: Would delete %d database records", count)
		g.setResults(fmt.Sprintf("DRY RUN: Would delete %d database records (no actual changes made)", count))
		return
	}

	result, err := cleaner.CleanAugmentData()
	if err != nil {
		g.logger.LogOperationResult("Clean Database", false, err.Error())
		g.showErrorDialog("Database Cleaning Failed", err.Error())
		return
	}

	g.logger.LogOperationResult("Clean Database", true, fmt.Sprintf("Deleted %d records", result.DeletedRows))
	g.logger.LogBackupCreated("database", result.DBBackupPath)

	// Display results
	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	g.setResults(fmt.Sprintf("Database Cleaned Successfully:\n%s", string(resultJSON)))
}

// runCleanWorkspace executes the workspace cleaning operation
func (g *MainGUI) runCleanWorkspace() {
	g.setOperationState(true, "Cleaning workspace...")
	defer g.setOperationState(false, "Ready")

	config := g.configManager.GetConfig()
	g.logger.LogOperation("Clean Workspace")

	if config.DryRunMode {
		workspacePath, err := utils.GetWorkspaceStoragePath()
		if err != nil {
			g.logger.Error("Failed to get workspace path: %v", err)
			g.showErrorDialog("Workspace Path Error", err.Error())
			return
		}
		g.logger.Info("DRY RUN MODE: Would clean workspace at %s", workspacePath)
		g.setResults(fmt.Sprintf("DRY RUN: Would clean workspace storage at %s (no actual changes made)", workspacePath))
		return
	}

	result, err := cleaner.CleanWorkspaceStorage()
	if err != nil {
		g.logger.LogOperationResult("Clean Workspace", false, err.Error())
		g.showErrorDialog("Workspace Cleaning Failed", err.Error())
		return
	}

	g.logger.LogOperationResult("Clean Workspace", true, fmt.Sprintf("Deleted %d files", result.DeletedFilesCount))
	g.logger.LogBackupCreated("workspace", result.BackupPath)

	// Log any failed operations
	for _, failed := range result.FailedOperations {
		g.logger.Warn("Failed to delete %s %s: %s", failed.Type, failed.Path, failed.Error)
	}

	// Display results
	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	g.setResults(fmt.Sprintf("Workspace Cleaned Successfully:\n%s", string(resultJSON)))
}

// runCleanBrowser executes the browser data cleaning operation
func (g *MainGUI) runCleanBrowser() {
	g.setOperationState(true, "Cleaning browser data...")
	defer g.setOperationState(false, "Ready")

	config := g.configManager.GetConfig()
	g.logger.LogOperation("Clean Browser Data")

	if config.DryRunMode {
		browserCleaner, err := browser.NewBrowserCleaner()
		if err != nil {
			g.logger.Error("Failed to create browser cleaner: %v", err)
			g.showErrorDialog("Browser Cleaner Failed", err.Error())
			return
		}

		counts, err := browserCleaner.GetBrowserDataCount()
		if err != nil {
			g.logger.Error("Failed to count browser data: %v", err)
			g.showErrorDialog("Browser Count Failed", err.Error())
			return
		}

		totalCount := int64(0)
		for _, count := range counts {
			totalCount += count
		}

		g.logger.Info("DRY RUN MODE: Would clean %d browser data items", totalCount)

		countsJSON, _ := json.MarshalIndent(counts, "", "  ")
		g.setResults(fmt.Sprintf("DRY RUN: Would clean browser data:\n%s\n\nTotal items: %d", string(countsJSON), totalCount))
		return
	}

	browserCleaner, err := browser.NewBrowserCleaner()
	if err != nil {
		g.logger.LogOperationResult("Clean Browser Data", false, err.Error())
		g.showErrorDialog("Browser Cleaner Failed", err.Error())
		return
	}

	results, err := browserCleaner.CleanBrowserData(config.CreateBackups)
	if err != nil {
		g.logger.LogOperationResult("Clean Browser Data", false, err.Error())
		g.showErrorDialog("Browser Cleaning Failed", err.Error())
		return
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
			g.logger.LogBackupCreated("browser-"+result.Profile.Name, result.BackupPath)
		}

		for _, err := range result.Errors {
			allErrors = append(allErrors, fmt.Sprintf("%s: %s", result.Profile.Name, err))
		}
	}

	// Log results
	successMsg := fmt.Sprintf("Cleaned %d cookies, %d storage items, %d cache items", totalCookies, totalStorage, totalCache)
	g.logger.LogOperationResult("Clean Browser Data", len(allErrors) == 0, successMsg)

	// Log any errors
	for _, err := range allErrors {
		g.logger.Error("Browser cleaning error: %s", err)
	}

	// Display results
	resultJSON, _ := json.MarshalIndent(results, "", "  ")
	g.setResults(fmt.Sprintf("Browser Data Cleaned:\n%s", string(resultJSON)))
}

// runAllOperations executes all cleaning operations in sequence
func (g *MainGUI) runAllOperations() {
	g.setOperationState(true, "Running all operations...")
	defer g.setOperationState(false, "Ready")

	g.logger.LogOperation("Run All Operations")

	// Step 1: Modify Telemetry IDs
	g.setStatus("Step 1/4: Modifying telemetry IDs...")
	g.setProgress(0.1)
	g.runModifyTelemetryInternal()
	g.setProgress(0.25)

	// Step 2: Clean Database
	g.setStatus("Step 2/4: Cleaning database...")
	g.runCleanDatabaseInternal()
	g.setProgress(0.5)

	// Step 3: Clean Workspace
	g.setStatus("Step 3/4: Cleaning workspace...")
	g.runCleanWorkspaceInternal()
	g.setProgress(0.75)

	// Step 4: Clean Browser Data
	g.setStatus("Step 4/4: Cleaning browser data...")
	g.runCleanBrowserInternal()
	g.setProgress(1.0)

	g.logger.LogOperationResult("Run All Operations", true, "All operations completed")
	g.setResults("All operations completed successfully! You can now restart VS Code and login with a new account.")
}

// Internal operation methods (without UI state management)
func (g *MainGUI) runModifyTelemetryInternal() {
	config := g.configManager.GetConfig()
	if config.DryRunMode {
		g.logger.Info("DRY RUN: Skipping telemetry modification")
		return
	}

	result, err := cleaner.ModifyTelemetryIDs()
	if err != nil {
		g.logger.Error("Telemetry modification failed: %v", err)
		return
	}
	g.logger.Info("Telemetry IDs modified successfully")
	g.logger.LogBackupCreated("storage.json", result.StorageBackupPath)
}

func (g *MainGUI) runCleanDatabaseInternal() {
	config := g.configManager.GetConfig()
	if config.DryRunMode {
		g.logger.Info("DRY RUN: Skipping database cleaning")
		return
	}

	result, err := cleaner.CleanAugmentData()
	if err != nil {
		g.logger.Error("Database cleaning failed: %v", err)
		return
	}
	g.logger.Info("Database cleaned successfully, deleted %d records", result.DeletedRows)
	g.logger.LogBackupCreated("database", result.DBBackupPath)
}

func (g *MainGUI) runCleanWorkspaceInternal() {
	config := g.configManager.GetConfig()
	if config.DryRunMode {
		g.logger.Info("DRY RUN: Skipping workspace cleaning")
		return
	}

	result, err := cleaner.CleanWorkspaceStorage()
	if err != nil {
		g.logger.Error("Workspace cleaning failed: %v", err)
		return
	}
	g.logger.Info("Workspace cleaned successfully, deleted %d files", result.DeletedFilesCount)
	g.logger.LogBackupCreated("workspace", result.BackupPath)
}

func (g *MainGUI) runCleanBrowserInternal() {
	config := g.configManager.GetConfig()
	if config.DryRunMode {
		g.logger.Info("DRY RUN: Skipping browser cleaning")
		return
	}

	browserCleaner, err := browser.NewBrowserCleaner()
	if err != nil {
		g.logger.Error("Browser cleaner creation failed: %v", err)
		return
	}

	results, err := browserCleaner.CleanBrowserData(config.CreateBackups)
	if err != nil {
		g.logger.Error("Browser cleaning failed: %v", err)
		return
	}

	// Count total items cleaned
	totalItems := int64(0)
	for _, result := range results {
		totalItems += result.CookiesDeleted + result.StorageDeleted + result.CacheDeleted
		if result.BackupPath != "" {
			g.logger.LogBackupCreated("browser-"+result.Profile.Name, result.BackupPath)
		}
	}

	g.logger.Info("Browser data cleaned successfully, processed %d items", totalItems)
}

// Helper methods for UI state management
func (g *MainGUI) setOperationState(running bool, status string) {
	g.isRunning = running
	g.setStatus(status)
	
	if running {
		g.showProgress()
		g.disableButtons()
	} else {
		g.hideProgress()
		g.enableButtons()
	}
}

func (g *MainGUI) disableButtons() {
	g.modifyTelemetryBtn.Disable()
	g.cleanDatabaseBtn.Disable()
	g.cleanWorkspaceBtn.Disable()
	g.cleanBrowserBtn.Disable()
	g.runAllBtn.Disable()
}

func (g *MainGUI) enableButtons() {
	g.modifyTelemetryBtn.Enable()
	g.cleanDatabaseBtn.Enable()
	g.cleanWorkspaceBtn.Enable()
	g.cleanBrowserBtn.Enable()
	g.runAllBtn.Enable()
}

// Dialog helpers
func (g *MainGUI) showConfirmationDialog(title, message string) bool {
	result := make(chan bool, 1)
	
	dialog.ShowConfirm(title, message, func(confirmed bool) {
		result <- confirmed
	}, g.window)
	
	return <-result
}

func (g *MainGUI) showErrorDialog(title, message string) {
	dialog.ShowError(fmt.Errorf(message), g.window)
}
