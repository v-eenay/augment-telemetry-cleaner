package scanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"augment-telemetry-cleaner/internal/utils"
)

// ConfigAnalysisResult represents the result of analyzing configuration files
type ConfigAnalysisResult struct {
	VSCodeSettings      []ConfigFinding `json:"vscode_settings"`
	ExtensionSettings   []ConfigFinding `json:"extension_settings"`
	WorkspaceSettings   []ConfigFinding `json:"workspace_settings"`
	TelemetrySettings   []ConfigFinding `json:"telemetry_settings"`
	TotalFindings       int             `json:"total_findings"`
	HighRiskFindings    int             `json:"high_risk_findings"`
}

// ConfigFinding represents a telemetry-related finding in configuration files
type ConfigFinding struct {
	File            string        `json:"file"`
	Path            string        `json:"path"`
	Key             string        `json:"key"`
	Value           interface{}   `json:"value"`
	Risk            TelemetryRisk `json:"risk"`
	Category        string        `json:"category"`
	Description     string        `json:"description"`
	Recommendation  string        `json:"recommendation"`
}

// ConfigAnalyzer handles analysis of VS Code and extension configuration files
type ConfigAnalyzer struct {
	telemetryKeys    map[string]TelemetryRisk
	extensionPatterns []*regexp.Regexp
}

// NewConfigAnalyzer creates a new configuration analyzer
func NewConfigAnalyzer() *ConfigAnalyzer {
	analyzer := &ConfigAnalyzer{}
	analyzer.initializeTelemetryKeys()
	analyzer.initializeExtensionPatterns()
	return analyzer
}

// initializeTelemetryKeys sets up known telemetry-related configuration keys
func (ca *ConfigAnalyzer) initializeTelemetryKeys() {
	ca.telemetryKeys = map[string]TelemetryRisk{
		// VS Code core telemetry settings
		"telemetry.telemetryLevel":                    TelemetryRiskHigh,
		"telemetry.enableTelemetry":                   TelemetryRiskHigh,
		"telemetry.enableCrashReporter":               TelemetryRiskHigh,
		"telemetry.optInTelemetry":                    TelemetryRiskHigh,
		
		// Application Insights
		"applicationinsights.instrumentationkey":      TelemetryRiskCritical,
		"applicationinsights.connectionstring":        TelemetryRiskCritical,
		
		// Extension-specific telemetry
		"extensions.autoCheckUpdates":                 TelemetryRiskMedium,
		"extensions.autoUpdate":                       TelemetryRiskMedium,
		"extensions.ignoreRecommendations":            TelemetryRiskLow,
		
		// Update and feedback settings
		"update.enableWindowsBackgroundUpdates":      TelemetryRiskMedium,
		"update.showReleaseNotes":                     TelemetryRiskLow,
		"workbench.enableExperiments":                 TelemetryRiskMedium,
		"workbench.settings.enableNaturalLanguageSearch": TelemetryRiskMedium,
		
		// GitHub and remote settings
		"github.gitAuthentication":                    TelemetryRiskMedium,
		"remote.downloadExtensionsLocally":           TelemetryRiskLow,
		
		// Language server and IntelliSense
		"typescript.surveys.enabled":                  TelemetryRiskMedium,
		"typescript.updateImportsOnFileMove.enabled": TelemetryRiskLow,
		"python.analysis.autoImportCompletions":      TelemetryRiskLow,
		
		// Specific extension telemetry keys
		"csharp.semanticHighlighting.enabled":        TelemetryRiskLow,
		"java.configuration.checkProjectSettings":    TelemetryRiskLow,
		"go.toolsManagement.autoUpdate":              TelemetryRiskMedium,
	}
}

// initializeExtensionPatterns sets up regex patterns for extension-specific settings
func (ca *ConfigAnalyzer) initializeExtensionPatterns() {
	patterns := []string{
		`(?i).*\.telemetry\..*`,
		`(?i).*\.analytics\..*`,
		`(?i).*\.tracking\..*`,
		`(?i).*\.usage\..*`,
		`(?i).*\.metrics\..*`,
		`(?i).*\.crash.*report.*`,
		`(?i).*\.error.*report.*`,
		`(?i).*\.feedback\..*`,
		`(?i).*\.survey\..*`,
		`(?i).*\.experiment.*`,
		`(?i).*\.autoUpdate.*`,
		`(?i).*\.checkUpdate.*`,
		`(?i).*\.sendUsage.*`,
		`(?i).*\.collectData.*`,
	}

	for _, pattern := range patterns {
		if regex, err := regexp.Compile(pattern); err == nil {
			ca.extensionPatterns = append(ca.extensionPatterns, regex)
		}
	}
}

// AnalyzeConfigurations performs comprehensive analysis of configuration files
func (ca *ConfigAnalyzer) AnalyzeConfigurations() (*ConfigAnalysisResult, error) {
	result := &ConfigAnalysisResult{
		VSCodeSettings:    make([]ConfigFinding, 0),
		ExtensionSettings: make([]ConfigFinding, 0),
		WorkspaceSettings: make([]ConfigFinding, 0),
		TelemetrySettings: make([]ConfigFinding, 0),
	}

	// Analyze VS Code user settings
	if err := ca.analyzeVSCodeSettings(result); err != nil {
		// Continue even if user settings analysis fails
	}

	// Analyze workspace settings
	if err := ca.analyzeWorkspaceSettings(result); err != nil {
		// Continue even if workspace settings analysis fails
	}

	// Analyze extension-specific configurations
	if err := ca.analyzeExtensionConfigurations(result); err != nil {
		// Continue even if extension config analysis fails
	}

	// Calculate totals
	ca.calculateTotals(result)

	return result, nil
}

// analyzeVSCodeSettings analyzes VS Code user settings.json
func (ca *ConfigAnalyzer) analyzeVSCodeSettings(result *ConfigAnalysisResult) error {
	settingsPath, err := ca.getVSCodeSettingsPath()
	if err != nil {
		return fmt.Errorf("failed to get settings path: %w", err)
	}

	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		return nil // Settings file doesn't exist, which is normal
	}

	settings, err := ca.loadJSONConfig(settingsPath)
	if err != nil {
		return fmt.Errorf("failed to load settings: %w", err)
	}

	ca.analyzeConfigObject(settings, settingsPath, "VS Code Settings", result)
	return nil
}

// analyzeWorkspaceSettings analyzes workspace-specific settings
func (ca *ConfigAnalyzer) analyzeWorkspaceSettings(result *ConfigAnalysisResult) error {
	// Look for .vscode/settings.json in common locations
	workspacePaths := ca.getWorkspaceSettingsPaths()

	for _, workspacePath := range workspacePaths {
		if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
			continue
		}

		settings, err := ca.loadJSONConfig(workspacePath)
		if err != nil {
			continue // Skip files we can't parse
		}

		ca.analyzeConfigObject(settings, workspacePath, "Workspace Settings", result)
	}

	return nil
}

// analyzeExtensionConfigurations analyzes extension-specific configuration files
func (ca *ConfigAnalyzer) analyzeExtensionConfigurations(result *ConfigAnalysisResult) error {
	// Analyze global storage configurations
	globalStoragePath, err := ca.getGlobalStoragePath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(globalStoragePath); err == nil {
		ca.analyzeGlobalStorageConfigs(globalStoragePath, result)
	}

	// Analyze workspace storage configurations
	workspaceStoragePath, err := utils.GetWorkspaceStoragePath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(workspaceStoragePath); err == nil {
		ca.analyzeWorkspaceStorageConfigs(workspaceStoragePath, result)
	}

	return nil
}

// getVSCodeSettingsPath returns the path to VS Code user settings
func (ca *ConfigAnalyzer) getVSCodeSettingsPath() (string, error) {
	homeDir, err := utils.GetHomeDir()
	if err != nil {
		return "", err
	}

	switch utils.GetOS() {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(homeDir, "AppData", "Roaming")
		}
		return filepath.Join(appData, "Code", "User", "settings.json"), nil

	case "darwin":
		return filepath.Join(homeDir, "Library", "Application Support", "Code", "User", "settings.json"), nil

	default: // Linux and other Unix-like systems
		return filepath.Join(homeDir, ".config", "Code", "User", "settings.json"), nil
	}
}

// getWorkspaceSettingsPaths returns possible workspace settings paths
func (ca *ConfigAnalyzer) getWorkspaceSettingsPaths() []string {
	var paths []string

	// Common workspace locations
	homeDir, err := utils.GetHomeDir()
	if err != nil {
		return paths
	}

	// Check common project directories
	commonDirs := []string{
		filepath.Join(homeDir, "Documents"),
		filepath.Join(homeDir, "Projects"),
		filepath.Join(homeDir, "Development"),
		filepath.Join(homeDir, "Code"),
		filepath.Join(homeDir, "Desktop"),
	}

	for _, dir := range commonDirs {
		if _, err := os.Stat(dir); err == nil {
			// Look for .vscode/settings.json in subdirectories
			ca.findWorkspaceSettings(dir, &paths, 2) // Max depth of 2
		}
	}

	return paths
}

// findWorkspaceSettings recursively finds workspace settings files
func (ca *ConfigAnalyzer) findWorkspaceSettings(dir string, paths *[]string, maxDepth int) {
	if maxDepth <= 0 {
		return
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		entryPath := filepath.Join(dir, entry.Name())

		// Check if this directory has .vscode/settings.json
		settingsPath := filepath.Join(entryPath, ".vscode", "settings.json")
		if _, err := os.Stat(settingsPath); err == nil {
			*paths = append(*paths, settingsPath)
		}

		// Recurse into subdirectories
		ca.findWorkspaceSettings(entryPath, paths, maxDepth-1)
	}
}

// getGlobalStoragePath returns the global storage path
func (ca *ConfigAnalyzer) getGlobalStoragePath() (string, error) {
	homeDir, err := utils.GetHomeDir()
	if err != nil {
		return "", err
	}

	switch utils.GetOS() {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(homeDir, "AppData", "Roaming")
		}
		return filepath.Join(appData, "Code", "User", "globalStorage"), nil

	case "darwin":
		return filepath.Join(homeDir, "Library", "Application Support", "Code", "User", "globalStorage"), nil

	default: // Linux and other Unix-like systems
		return filepath.Join(homeDir, ".config", "Code", "User", "globalStorage"), nil
	}
}

// loadJSONConfig loads and parses a JSON configuration file
func (ca *ConfigAnalyzer) loadJSONConfig(filePath string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return config, nil
}

// analyzeConfigObject analyzes a configuration object for telemetry settings
func (ca *ConfigAnalyzer) analyzeConfigObject(config map[string]interface{}, filePath, category string, result *ConfigAnalysisResult) {
	ca.analyzeConfigRecursive(config, filePath, category, "", result)
}

// analyzeConfigRecursive recursively analyzes configuration objects
func (ca *ConfigAnalyzer) analyzeConfigRecursive(obj interface{}, filePath, category, keyPath string, result *ConfigAnalysisResult) {
	switch v := obj.(type) {
	case map[string]interface{}:
		for key, value := range v {
			currentPath := key
			if keyPath != "" {
				currentPath = keyPath + "." + key
			}
			
			// Check if this key is telemetry-related
			if risk, found := ca.telemetryKeys[currentPath]; found {
				finding := ConfigFinding{
					File:        filePath,
					Path:        currentPath,
					Key:         key,
					Value:       value,
					Risk:        risk,
					Category:    category,
					Description: ca.getKeyDescription(currentPath, risk),
					Recommendation: ca.getKeyRecommendation(currentPath, value),
				}
				
				ca.addFinding(finding, result)
			}

			// Check against extension patterns
			for _, pattern := range ca.extensionPatterns {
				if pattern.MatchString(currentPath) {
					risk := ca.determinePatternRisk(currentPath, value)
					finding := ConfigFinding{
						File:        filePath,
						Path:        currentPath,
						Key:         key,
						Value:       value,
						Risk:        risk,
						Category:    category,
						Description: ca.getPatternDescription(currentPath, risk),
						Recommendation: ca.getPatternRecommendation(currentPath, value),
					}
					
					ca.addFinding(finding, result)
					break // Only match first pattern to avoid duplicates
				}
			}

			// Recurse into nested objects
			ca.analyzeConfigRecursive(value, filePath, category, currentPath, result)
		}
	}
}

// addFinding adds a finding to the appropriate category in results
func (ca *ConfigAnalyzer) addFinding(finding ConfigFinding, result *ConfigAnalysisResult) {
	// Categorize the finding
	if strings.Contains(strings.ToLower(finding.Path), "telemetry") {
		result.TelemetrySettings = append(result.TelemetrySettings, finding)
	} else if strings.Contains(finding.Category, "Workspace") {
		result.WorkspaceSettings = append(result.WorkspaceSettings, finding)
	} else if ca.isExtensionSetting(finding.Path) {
		result.ExtensionSettings = append(result.ExtensionSettings, finding)
	} else {
		result.VSCodeSettings = append(result.VSCodeSettings, finding)
	}
}

// isExtensionSetting determines if a setting path belongs to an extension
func (ca *ConfigAnalyzer) isExtensionSetting(path string) bool {
	// Extension settings typically have a publisher.extension format
	parts := strings.Split(path, ".")
	if len(parts) >= 2 {
		// Check if it looks like an extension setting (has at least 2 parts)
		return !ca.isCoreSetting(parts[0])
	}
	return false
}

// isCoreSetting checks if a setting belongs to VS Code core
func (ca *ConfigAnalyzer) isCoreSetting(prefix string) bool {
	coreSettings := []string{
		"editor", "workbench", "window", "files", "search", "debug",
		"extensions", "terminal", "scm", "problems", "breadcrumbs",
		"telemetry", "update", "security", "remote", "merge-conflict",
	}

	for _, core := range coreSettings {
		if strings.EqualFold(prefix, core) {
			return true
		}
	}
	return false
}

// getKeyDescription returns a description for a known telemetry key
func (ca *ConfigAnalyzer) getKeyDescription(key string, risk TelemetryRisk) string {
	descriptions := map[string]string{
		"telemetry.telemetryLevel":     "Controls the level of telemetry data sent to Microsoft",
		"telemetry.enableTelemetry":    "Enables or disables telemetry data collection",
		"telemetry.enableCrashReporter": "Controls crash report submission",
		"applicationinsights.instrumentationkey": "Application Insights instrumentation key for telemetry",
		"extensions.autoCheckUpdates":  "Automatically checks for extension updates",
		"workbench.enableExperiments":  "Enables experimental features that may collect data",
	}

	if desc, found := descriptions[key]; found {
		return desc
	}

	return fmt.Sprintf("Telemetry-related setting with %s risk level", risk.String())
}

// getKeyRecommendation returns a recommendation for a telemetry setting
func (ca *ConfigAnalyzer) getKeyRecommendation(key string, value interface{}) string {
	switch key {
	case "telemetry.telemetryLevel":
		if value == "off" {
			return "Good: Telemetry is disabled"
		}
		return "Consider setting to 'off' to disable telemetry"
	case "telemetry.enableTelemetry":
		if value == false {
			return "Good: Telemetry is disabled"
		}
		return "Consider setting to false to disable telemetry"
	case "telemetry.enableCrashReporter":
		if value == false {
			return "Good: Crash reporting is disabled"
		}
		return "Consider setting to false to disable crash reporting"
	default:
		return "Review this setting and disable if not needed"
	}
}

// determinePatternRisk determines the risk level for a pattern match
func (ca *ConfigAnalyzer) determinePatternRisk(path string, value interface{}) TelemetryRisk {
	lowerPath := strings.ToLower(path)
	
	if strings.Contains(lowerPath, "telemetry") || strings.Contains(lowerPath, "analytics") {
		return TelemetryRiskHigh
	}
	if strings.Contains(lowerPath, "tracking") || strings.Contains(lowerPath, "usage") {
		return TelemetryRiskMedium
	}
	if strings.Contains(lowerPath, "update") || strings.Contains(lowerPath, "experiment") {
		return TelemetryRiskMedium
	}
	
	return TelemetryRiskLow
}

// getPatternDescription returns a description for a pattern match
func (ca *ConfigAnalyzer) getPatternDescription(path string, risk TelemetryRisk) string {
	return fmt.Sprintf("Extension setting that may be related to telemetry or data collection (%s risk)", risk.String())
}

// getPatternRecommendation returns a recommendation for a pattern match
func (ca *ConfigAnalyzer) getPatternRecommendation(path string, value interface{}) string {
	return "Review this extension setting and disable if it relates to unwanted data collection"
}

// analyzeGlobalStorageConfigs analyzes global storage configuration files
func (ca *ConfigAnalyzer) analyzeGlobalStorageConfigs(globalStoragePath string, result *ConfigAnalysisResult) {
	entries, err := os.ReadDir(globalStoragePath)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		extensionPath := filepath.Join(globalStoragePath, entry.Name())
		ca.analyzeExtensionStorageDir(extensionPath, "Global Storage", result)
	}
}

// analyzeWorkspaceStorageConfigs analyzes workspace storage configuration files
func (ca *ConfigAnalyzer) analyzeWorkspaceStorageConfigs(workspaceStoragePath string, result *ConfigAnalysisResult) {
	entries, err := os.ReadDir(workspaceStoragePath)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		workspacePath := filepath.Join(workspaceStoragePath, entry.Name())
		ca.analyzeExtensionStorageDir(workspacePath, "Workspace Storage", result)
	}
}

// analyzeExtensionStorageDir analyzes an extension's storage directory
func (ca *ConfigAnalyzer) analyzeExtensionStorageDir(dirPath, category string, result *ConfigAnalysisResult) {
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue despite errors
		}

		if info.IsDir() {
			return nil
		}

		// Only analyze JSON files
		if strings.ToLower(filepath.Ext(path)) != ".json" {
			return nil
		}

		// Skip very large files
		if info.Size() > 1024*1024 { // 1MB limit
			return nil
		}

		config, err := ca.loadJSONConfig(path)
		if err != nil {
			return nil // Skip files we can't parse
		}

		ca.analyzeConfigObject(config, path, category, result)
		return nil
	})

	if err != nil {
		// Continue despite walk errors
	}
}

// calculateTotals calculates summary statistics for the analysis result
func (ca *ConfigAnalyzer) calculateTotals(result *ConfigAnalysisResult) {
	allFindings := [][]ConfigFinding{
		result.VSCodeSettings,
		result.ExtensionSettings,
		result.WorkspaceSettings,
		result.TelemetrySettings,
	}

	for _, findings := range allFindings {
		result.TotalFindings += len(findings)
		for _, finding := range findings {
			if finding.Risk >= TelemetryRiskHigh {
				result.HighRiskFindings++
			}
		}
	}
}