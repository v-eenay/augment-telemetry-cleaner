package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"augment-telemetry-cleaner/internal/config"
)

// SettingsDialog represents the settings configuration dialog
type SettingsDialog struct {
	parent        fyne.Window
	configManager *config.ConfigManager
	
	// UI components
	dryRunCheck       *widget.Check
	backupCheck       *widget.Check
	confirmCheck      *widget.Check
	previewCheck      *widget.Check
	
	logLevelSelect    *widget.Select
	backupDirEntry    *widget.Entry
	maxBackupEntry    *widget.Entry
	dbTimeoutEntry    *widget.Entry
	retriesEntry      *widget.Entry
	
	dialog            dialog.Dialog
}

// NewSettingsDialog creates a new settings dialog
func NewSettingsDialog(parent fyne.Window, configManager *config.ConfigManager) *SettingsDialog {
	sd := &SettingsDialog{
		parent:        parent,
		configManager: configManager,
	}
	
	sd.createComponents()
	sd.loadCurrentSettings()
	
	return sd
}

// createComponents creates all the UI components for the settings dialog
func (sd *SettingsDialog) createComponents() {
	config := sd.configManager.GetConfig()
	
	// Safety settings
	sd.dryRunCheck = widget.NewCheck("Enable Dry Run Mode by default", nil)
	sd.backupCheck = widget.NewCheck("Create backups before operations", nil)
	sd.confirmCheck = widget.NewCheck("Require confirmation for operations", nil)
	sd.previewCheck = widget.NewCheck("Show preview before running operations", nil)
	
	// Log level selection
	sd.logLevelSelect = widget.NewSelect([]string{"DEBUG", "INFO", "WARN", "ERROR"}, nil)
	sd.logLevelSelect.SetSelected(config.LogLevel)
	
	// Backup directory
	sd.backupDirEntry = widget.NewEntry()
	sd.backupDirEntry.SetText(config.BackupDirectory)
	
	// Numeric settings
	sd.maxBackupEntry = widget.NewEntry()
	sd.maxBackupEntry.SetText(fmt.Sprintf("%d", config.MaxBackupAge))
	
	sd.dbTimeoutEntry = widget.NewEntry()
	sd.dbTimeoutEntry.SetText(fmt.Sprintf("%d", config.DatabaseTimeout))
	
	sd.retriesEntry = widget.NewEntry()
	sd.retriesEntry.SetText(fmt.Sprintf("%d", config.FileOperationRetries))
}

// loadCurrentSettings loads the current configuration into the UI
func (sd *SettingsDialog) loadCurrentSettings() {
	config := sd.configManager.GetConfig()
	
	sd.dryRunCheck.SetChecked(config.DryRunMode)
	sd.backupCheck.SetChecked(config.CreateBackups)
	sd.confirmCheck.SetChecked(config.RequireConfirmation)
	sd.previewCheck.SetChecked(config.ShowPreviewBeforeRun)
}

// Show displays the settings dialog
func (sd *SettingsDialog) Show() {
	content := sd.createDialogContent()
	
	sd.dialog = dialog.NewCustom("Settings", "Close", content, sd.parent)
	sd.dialog.Resize(fyne.NewSize(500, 600))
	sd.dialog.Show()
}

// createDialogContent creates the main content for the settings dialog
func (sd *SettingsDialog) createDialogContent() fyne.CanvasObject {
	// Safety settings section
	safetyCard := widget.NewCard("Safety Settings", "", container.NewVBox(
		sd.dryRunCheck,
		sd.backupCheck,
		sd.confirmCheck,
		sd.previewCheck,
	))
	
	// Logging settings section
	loggingCard := widget.NewCard("Logging Settings", "", container.NewVBox(
		widget.NewLabel("Log Level:"),
		sd.logLevelSelect,
	))
	
	// Backup settings section
	backupDirContainer := container.NewBorder(
		nil, nil, nil, widget.NewButton("Browse", sd.onBrowseBackupDir),
		sd.backupDirEntry,
	)
	
	backupCard := widget.NewCard("Backup Settings", "", container.NewVBox(
		widget.NewLabel("Backup Directory:"),
		backupDirContainer,
		widget.NewLabel("Maximum Backup Age (days):"),
		sd.maxBackupEntry,
	))
	
	// Advanced settings section
	advancedCard := widget.NewCard("Advanced Settings", "", container.NewVBox(
		widget.NewLabel("Database Timeout (seconds):"),
		sd.dbTimeoutEntry,
		widget.NewLabel("File Operation Retries:"),
		sd.retriesEntry,
	))
	
	// Action buttons
	saveBtn := widget.NewButton("Save Settings", sd.onSave)
	saveBtn.Importance = widget.HighImportance
	
	resetBtn := widget.NewButton("Reset to Defaults", sd.onReset)
	
	buttonsContainer := container.NewHBox(
		saveBtn,
		resetBtn,
	)
	
	// Main content
	content := container.NewVBox(
		safetyCard,
		loggingCard,
		backupCard,
		advancedCard,
		widget.NewSeparator(),
		buttonsContainer,
	)
	
	return container.NewScroll(content)
}

// Event handlers
func (sd *SettingsDialog) onBrowseBackupDir() {
	folderDialog := dialog.NewFolderOpen(func(folder fyne.ListableURI, err error) {
		if err == nil && folder != nil {
			sd.backupDirEntry.SetText(folder.Path())
		}
	}, sd.parent)

	folderDialog.Show()
}

func (sd *SettingsDialog) onSave() {
	// Validate inputs
	if err := sd.validateInputs(); err != nil {
		dialog.ShowError(err, sd.parent)
		return
	}
	
	// Update configuration
	err := sd.configManager.UpdateConfig(func(config *config.Config) {
		config.DryRunMode = sd.dryRunCheck.Checked
		config.CreateBackups = sd.backupCheck.Checked
		config.RequireConfirmation = sd.confirmCheck.Checked
		config.ShowPreviewBeforeRun = sd.previewCheck.Checked
		config.LogLevel = sd.logLevelSelect.Selected
		config.BackupDirectory = sd.backupDirEntry.Text
		
		// Parse numeric values
		if maxAge, err := parseIntSafe(sd.maxBackupEntry.Text); err == nil {
			config.MaxBackupAge = maxAge
		}
		if timeout, err := parseIntSafe(sd.dbTimeoutEntry.Text); err == nil {
			config.DatabaseTimeout = timeout
		}
		if retries, err := parseIntSafe(sd.retriesEntry.Text); err == nil {
			config.FileOperationRetries = retries
		}
	})
	
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to save settings: %w", err), sd.parent)
		return
	}
	
	dialog.ShowInformation("Settings Saved", "Settings have been saved successfully!", sd.parent)
	sd.dialog.Hide()
}

func (sd *SettingsDialog) onReset() {
	dialog.ShowConfirm("Reset Settings", 
		"Are you sure you want to reset all settings to their default values?", 
		func(confirmed bool) {
			if confirmed {
				sd.resetToDefaults()
			}
		}, sd.parent)
}

// resetToDefaults resets all settings to their default values
func (sd *SettingsDialog) resetToDefaults() {
	defaultConfig := config.DefaultConfig()
	
	sd.dryRunCheck.SetChecked(defaultConfig.DryRunMode)
	sd.backupCheck.SetChecked(defaultConfig.CreateBackups)
	sd.confirmCheck.SetChecked(defaultConfig.RequireConfirmation)
	sd.previewCheck.SetChecked(defaultConfig.ShowPreviewBeforeRun)
	sd.logLevelSelect.SetSelected(defaultConfig.LogLevel)
	sd.backupDirEntry.SetText(defaultConfig.BackupDirectory)
	sd.maxBackupEntry.SetText(fmt.Sprintf("%d", defaultConfig.MaxBackupAge))
	sd.dbTimeoutEntry.SetText(fmt.Sprintf("%d", defaultConfig.DatabaseTimeout))
	sd.retriesEntry.SetText(fmt.Sprintf("%d", defaultConfig.FileOperationRetries))
}

// validateInputs validates all user inputs
func (sd *SettingsDialog) validateInputs() error {
	// Validate backup directory
	if sd.backupDirEntry.Text == "" {
		return fmt.Errorf("backup directory cannot be empty")
	}
	
	// Validate numeric inputs
	if _, err := parseIntSafe(sd.maxBackupEntry.Text); err != nil {
		return fmt.Errorf("invalid maximum backup age: must be a number")
	}
	
	if timeout, err := parseIntSafe(sd.dbTimeoutEntry.Text); err != nil || timeout <= 0 {
		return fmt.Errorf("invalid database timeout: must be a positive number")
	}
	
	if retries, err := parseIntSafe(sd.retriesEntry.Text); err != nil || retries < 0 {
		return fmt.Errorf("invalid file operation retries: must be a non-negative number")
	}
	
	return nil
}

// Helper function to safely parse integers
func parseIntSafe(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}
