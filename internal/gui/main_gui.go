package gui

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"augment-telemetry-cleaner/internal/config"
	"augment-telemetry-cleaner/internal/logger"
)

// MainGUI represents the main GUI application
type MainGUI struct {
	window fyne.Window

	// Core components
	configManager *config.ConfigManager
	logger        *logger.Logger

	// UI Components
	statusLabel    *widget.Label
	progressBar    *widget.ProgressBar
	logText        *widget.Entry

	// Operation buttons
	modifyTelemetryBtn  *widget.Button
	cleanDatabaseBtn    *widget.Button
	cleanWorkspaceBtn   *widget.Button
	cleanBrowserBtn     *widget.Button
	runAllBtn          *widget.Button

	// Mode selection
	dryRunCheck        *widget.Check
	backupCheck        *widget.Check
	confirmCheck       *widget.Check

	// Results display
	resultsText        *widget.Entry

	// Operation state
	isRunning          bool
}



// NewMainGUI creates a new instance of the main GUI
func NewMainGUI(window fyne.Window) *MainGUI {
	// Initialize configuration manager
	configManager, err := config.NewConfigManager()
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to initialize configuration: %w", err), window)
		return nil
	}

	// Initialize logger
	logDir := "logs"
	logger, err := logger.NewLogger(logDir, nil)
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to initialize logger: %w", err), window)
		return nil
	}

	gui := &MainGUI{
		window:        window,
		configManager: configManager,
		logger:        logger,
		isRunning:     false,
	}

	// Set up logger callback for GUI updates
	gui.logger = logger // This will be updated with callback after GUI initialization

	gui.initializeComponents()
	return gui
}

// initializeComponents initializes all GUI components
func (g *MainGUI) initializeComponents() {
	config := g.configManager.GetConfig()

	// Status and progress components
	g.statusLabel = widget.NewLabel("Ready to clean Augment telemetry data")
	g.progressBar = widget.NewProgressBar()
	g.progressBar.Hide()

	// Log display
	g.logText = widget.NewMultiLineEntry()
	g.logText.SetText("Application started. Ready to perform operations.\n")
	g.logText.Wrapping = fyne.TextWrapWord
	g.logText.MultiLine = true

	// Set up logger callback for real-time updates
	g.logger.SetGUICallback(g.appendToLog)

	// Operation buttons with improved sizing
	g.modifyTelemetryBtn = widget.NewButton("Modify Telemetry IDs", g.onModifyTelemetry)
	g.cleanDatabaseBtn = widget.NewButton("Clean Database", g.onCleanDatabase)
	g.cleanWorkspaceBtn = widget.NewButton("Clean Workspace", g.onCleanWorkspace)
	g.cleanBrowserBtn = widget.NewButton("Clean Browser Data", g.onCleanBrowser)
	g.runAllBtn = widget.NewButton("Run All Operations", g.onRunAll)

	// Make the main action button more prominent
	g.runAllBtn.Importance = widget.HighImportance

	// Mode selection
	g.dryRunCheck = widget.NewCheck("Dry Run Mode (Preview only)", g.onDryRunToggle)
	g.dryRunCheck.SetChecked(config.DryRunMode)

	g.backupCheck = widget.NewCheck("Create Backups", g.onBackupToggle)
	g.backupCheck.SetChecked(config.CreateBackups)

	g.confirmCheck = widget.NewCheck("Require Confirmation", g.onConfirmToggle)
	g.confirmCheck.SetChecked(config.RequireConfirmation)

	// Results display
	g.resultsText = widget.NewMultiLineEntry()
	g.resultsText.SetText("Operation results will appear here...")
	g.resultsText.Wrapping = fyne.TextWrapWord
	g.resultsText.MultiLine = true

	// Update logger with GUI callback
	logDir := "logs"
	var err error
	g.logger, err = logger.NewLogger(logDir, g.onLogMessage)
	if err != nil {
		g.appendLog(fmt.Sprintf("Warning: Failed to reinitialize logger: %v", err))
	}
}

// BuildUI constructs and returns the main UI layout
func (g *MainGUI) BuildUI() fyne.CanvasObject {
	// Top section - status and controls in a compact row
	topSection := container.NewVBox(
		g.statusLabel,
		g.progressBar,
		container.NewHBox(
			g.dryRunCheck,
			g.backupCheck,
			g.confirmCheck,
		),
	)

	// Operation buttons in a compact grid
	buttonsGrid := container.NewGridWithColumns(2,
		g.modifyTelemetryBtn,
		g.cleanDatabaseBtn,
		g.cleanWorkspaceBtn,
		g.cleanBrowserBtn,
	)

	// Main action button
	mainActionContainer := container.NewVBox(
		buttonsGrid,
		g.runAllBtn,
	)

	// Log and results areas with optimized heights
	logScroll := container.NewScroll(g.logText)
	logScroll.SetMinSize(fyne.NewSize(800, 180))

	resultsScroll := container.NewScroll(g.resultsText)
	resultsScroll.SetMinSize(fyne.NewSize(800, 120))

	// Single-panel layout for maximum space efficiency
	mainContent := container.NewVBox(
		topSection,
		widget.NewSeparator(),
		mainActionContainer,
		widget.NewSeparator(),
		widget.NewLabel("Log:"),
		logScroll,
		widget.NewLabel("Results:"),
		resultsScroll,
	)

	// Simplified footer with only essential elements
	footer := container.NewHBox(
		widget.NewLabel("© 2025 Augment Telemetry Cleaner v2.0.0 - Vinay Koirala"),
		widget.NewButton("Exit", g.onExit),
	)

	return container.NewBorder(
		nil,
		footer,
		nil,
		nil,
		mainContent,
	)
}



// Event handlers for operations
func (g *MainGUI) onModifyTelemetry() {
	if g.isRunning {
		return
	}

	config := g.configManager.GetConfig()
	if config.RequireConfirmation && !g.showConfirmationDialog("Modify Telemetry IDs", "This will modify VS Code's telemetry IDs. Continue?") {
		return
	}

	go g.runModifyTelemetry()
}

func (g *MainGUI) onCleanDatabase() {
	if g.isRunning {
		return
	}

	config := g.configManager.GetConfig()
	if config.RequireConfirmation && !g.showConfirmationDialog("Clean Database", "This will remove Augment-related data from VS Code's database. Continue?") {
		return
	}

	go g.runCleanDatabase()
}

func (g *MainGUI) onCleanWorkspace() {
	if g.isRunning {
		return
	}

	config := g.configManager.GetConfig()
	if config.RequireConfirmation && !g.showConfirmationDialog("Clean Workspace", "This will clean VS Code's workspace storage. Continue?") {
		return
	}

	go g.runCleanWorkspace()
}

func (g *MainGUI) onCleanBrowser() {
	if g.isRunning {
		return
	}

	config := g.configManager.GetConfig()
	if config.RequireConfirmation && !g.showConfirmationDialog("Clean Browser Data",
		"This will remove Augment-related data from your browsers (cookies, storage, cache).\n\n"+
		"⚠️ WARNING: Please close all browsers before proceeding.\n\n"+
		"Continue?") {
		return
	}

	go g.runCleanBrowser()
}

func (g *MainGUI) onRunAll() {
	if g.isRunning {
		return
	}

	config := g.configManager.GetConfig()
	if config.RequireConfirmation && !g.showConfirmationDialog("Run All Operations", "This will run all cleaning operations. Continue?") {
		return
	}

	go g.runAllOperations()
}



func (g *MainGUI) onExit() {
	if g.logger != nil {
		g.logger.Close()
	}
	g.window.Close()
}

// Configuration event handlers
func (g *MainGUI) onDryRunToggle(checked bool) {
	g.configManager.UpdateConfig(func(config *config.Config) {
		config.DryRunMode = checked
	})
}

func (g *MainGUI) onBackupToggle(checked bool) {
	g.configManager.UpdateConfig(func(config *config.Config) {
		config.CreateBackups = checked
	})
}

func (g *MainGUI) onConfirmToggle(checked bool) {
	g.configManager.UpdateConfig(func(config *config.Config) {
		config.RequireConfirmation = checked
	})
}

// Logger callback
func (g *MainGUI) onLogMessage(level logger.LogLevel, message string) {
	// This runs in a goroutine, so we need to update the GUI safely
	timestamp := time.Now().Format("15:04:05")
	logEntry := fmt.Sprintf("[%s] %s: %s", timestamp, level.String(), message)
	g.appendLog(logEntry)
}

// Helper methods
func (g *MainGUI) appendLog(message string) {
	current := g.logText.Text
	g.logText.SetText(current + message + "\n")
	// Auto-scroll to bottom
	g.logText.CursorRow = len(g.logText.Text)
}

func (g *MainGUI) setStatus(status string) {
	g.statusLabel.SetText(status)
}

func (g *MainGUI) showProgress() {
	g.progressBar.Show()
}

func (g *MainGUI) hideProgress() {
	g.progressBar.Hide()
}

func (g *MainGUI) setProgress(value float64) {
	g.progressBar.SetValue(value)
}

func (g *MainGUI) setResults(results string) {
	g.resultsText.SetText(results)
}



// appendToLog adds a log entry to the log display
func (g *MainGUI) appendToLog(level, message string) {
	// Format timestamp
	timestamp := time.Now().Format("15:04:05")

	// Format the log message
	logMessage := fmt.Sprintf("[%s] %s: %s\n", timestamp, strings.ToUpper(level), message)

	// Append to log text
	currentText := g.logText.Text
	g.logText.SetText(currentText + logMessage)

	// Auto-scroll to bottom by moving cursor to end
	g.logText.CursorRow = len(strings.Split(g.logText.Text, "\n")) - 1
	g.logText.CursorColumn = 0
}
