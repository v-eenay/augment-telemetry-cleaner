package scanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"augment-telemetry-cleaner/internal/utils"
)

// ExtensionSettingsResult represents the result of scanning extension settings
type ExtensionSettingsResult struct {
	ExtensionSettings   []ExtensionSetting `json:"extension_settings"`
	GlobalStorageItems  []StorageItem      `json:"global_storage_items"`
	WorkspaceStorageItems []StorageItem    `json:"workspace_storage_items"`
	TotalSettings       int                `json:"total_settings"`
	TelemetrySettings   int                `json:"telemetry_settings"`
	ScanDuration        time.Duration      `json:"scan_duration"`
}

// ExtensionSetting represents a setting for a specific extension
type ExtensionSetting struct {
	ExtensionID     string        `json:"extension_id"`
	SettingKey      string        `json:"setting_key"`
	SettingValue    interface{}   `json:"setting_value"`
	Source          string        `json:"source"` // "user", "workspace", "default"
	Risk            TelemetryRisk `json:"risk"`
	Category        string        `json:"category"`
	Description     string        `json:"description"`
	LastModified    time.Time     `json:"last_modified"`
}

// StorageItem represents an item stored by an extension
type StorageItem struct {
	ExtensionID     string        `json:"extension_id"`
	StorageType     string        `json:"storage_type"` // "global", "workspace"
	Key             string        `json:"key"`
	Value           interface{}   `json:"value"`
	Size            int64         `json:"size"`
	Risk            TelemetryRisk `json:"risk"`
	Description     string        `json:"description"`
	FilePath        string        `json:"file_path"`
	LastModified    time.Time     `json:"last_modified"`
}

// ExtensionSettingsScanner handles scanning of extension-specific settings and storage
type ExtensionSettingsScanner struct {
	telemetryKeyPatterns map[string]TelemetryRisk
	storageKeyPatterns   map[string]TelemetryRisk
}

// NewExtensionSettingsScanner creates a new extension settings scanner
func NewExtensionSettingsScanner() *ExtensionSettingsScanner {
	scanner := &ExtensionSettingsScanner{}
	scanner.initializeTelemetryKeyPatterns()
	scanner.initializeStorageKeyPatterns()
	return scanner
}

// initializeTelemetryKeyPatterns sets up patterns for telemetry-related setting keys
func (ess *ExtensionSettingsScanner) initializeTelemetryKeyPatterns() {
	ess.telemetryKeyPatterns = map[string]TelemetryRisk{
		// Common telemetry setting patterns
		"telemetry":                    TelemetryRiskHigh,
		"analytics":                    TelemetryRiskHigh,
		"tracking":                     TelemetryRiskHigh,
		"usage":                        TelemetryRiskMedium,
		"metrics":                      TelemetryRiskMedium,
		"statistics":                   TelemetryRiskMedium,
		"crash":                        TelemetryRiskMedium,
		"error":                        TelemetryRiskMedium,
		"feedback":                     TelemetryRiskLow,
		"survey":                       TelemetryRiskLow,
		"experiment":                   TelemetryRiskMedium,
		"autoUpdate":                   TelemetryRiskMedium,
		"checkUpdate":                  TelemetryRiskMedium,
		"sendUsage":                    TelemetryRiskHigh,
		"collectData":                  TelemetryRiskHigh,
		"reportErrors":                 TelemetryRiskMedium,
		"enableLogging":                TelemetryRiskLow,
		"diagnostics":                  TelemetryRiskMedium,
		"performance":                  TelemetryRiskLow,
		
		// Specific extension patterns
		"python.analysis.autoImportCompletions": TelemetryRiskLow,
		"typescript.surveys.enabled":   TelemetryRiskMedium,
		"go.toolsManagement.autoUpdate": TelemetryRiskMedium,
		"java.configuration.checkProjectSettings": TelemetryRiskLow,
		"csharp.semanticHighlighting.enabled": TelemetryRiskLow,
		"eslint.autoFixOnSave":         TelemetryRiskLow,
		"prettier.requireConfig":       TelemetryRiskLow,
	}
}

// initializeStorageKeyPatterns sets up patterns for telemetry-related storage keys
func (ess *ExtensionSettingsScanner) initializeStorageKeyPatterns() {
	ess.storageKeyPatterns = map[string]TelemetryRisk{
		// Storage keys that might contain telemetry data
		"telemetryData":                TelemetryRiskCritical,
		"analyticsData":                TelemetryRiskCritical,
		"usageStats":                   TelemetryRiskHigh,
		"userMetrics":                  TelemetryRiskHigh,
		"sessionData":                  TelemetryRiskHigh,
		"machineId":                    TelemetryRiskCritical,
		"deviceId":                     TelemetryRiskCritical,
		"userId":                       TelemetryRiskHigh,
		"installId":                    TelemetryRiskHigh,
		"crashReports":                 TelemetryRiskMedium,
		"errorLogs":                    TelemetryRiskMedium,
		"performanceData":              TelemetryRiskMedium,
		"featureUsage":                 TelemetryRiskMedium,
		"lastUsed":                     TelemetryRiskLow,
		"activationCount":              TelemetryRiskLow,
		"commandHistory":               TelemetryRiskMedium,
		"searchHistory":                TelemetryRiskMedium,
		"recentFiles":                  TelemetryRiskLow,
		"preferences":                  TelemetryRiskLow,
		"configuration":                TelemetryRiskLow,
		"cache":                        TelemetryRiskLow,
		"temp":                         TelemetryRiskLow,
		"logs":                         TelemetryRiskMedium,
		"diagnostics":                  TelemetryRiskMedium,
		"experiments":                  TelemetryRiskMedium,
		"surveys":                      TelemetryRiskMedium,
		"feedback":                     TelemetryRiskLow,
	}
}

// ScanExtensionSettings performs comprehensive scanning of extension settings and storage
func (ess *ExtensionSettingsScanner) ScanExtensionSettings() (*ExtensionSettingsResult, error) {
	startTime := time.Now()
	
	result := &ExtensionSettingsResult{
		ExtensionSettings:     make([]ExtensionSetting, 0),
		GlobalStorageItems:    make([]StorageItem, 0),
		WorkspaceStorageItems: make([]StorageItem, 0),
	}

	// Scan user settings for extension configurations
	if err := ess.scanUserSettings(result); err != nil {
		// Continue even if user settings scan fails
	}

	// Scan workspace settings for extension configurations
	if err := ess.scanWorkspaceSettings(result); err != nil {
		// Continue even if workspace settings scan fails
	}

	// Scan global storage
	if err := ess.scanGlobalStorage(result); err != nil {
		// Continue even if global storage scan fails
	}

	// Scan workspace storage
	if err := ess.scanWorkspaceStorage(result); err != nil {
		// Continue even if workspace storage scan fails
	}

	// Calculate totals
	ess.calculateTotals(result)
	result.ScanDuration = time.Since(startTime)

	return result, nil
}

// scanUserSettings scans VS Code user settings for extension configurations
func (ess *ExtensionSettingsScanner) scanUserSettings(result *ExtensionSettingsResult) error {
	settingsPath, err := ess.getVSCodeSettingsPath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		return nil // Settings file doesn't exist
	}

	settings, err := ess.loadJSONConfig(settingsPath)
	if err != nil {
		return err
	}

	info, _ := os.Stat(settingsPath)
	lastModified := time.Now()
	if info != nil {
		lastModified = info.ModTime()
	}

	ess.extractExtensionSettings(settings, "user", settingsPath, lastModified, result)
	return nil
}

// scanWorkspaceSettings scans workspace settings for extension configurations
func (ess *ExtensionSettingsScanner) scanWorkspaceSettings(result *ExtensionSettingsResult) error {
	workspacePaths := ess.getWorkspaceSettingsPaths()

	for _, workspacePath := range workspacePaths {
		if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
			continue
		}

		settings, err := ess.loadJSONConfig(workspacePath)
		if err != nil {
			continue
		}

		info, _ := os.Stat(workspacePath)
		lastModified := time.Now()
		if info != nil {
			lastModified = info.ModTime()
		}

		ess.extractExtensionSettings(settings, "workspace", workspacePath, lastModified, result)
	}

	return nil
}

// scanGlobalStorage scans extension global storage directories
func (ess *ExtensionSettingsScanner) scanGlobalStorage(result *ExtensionSettingsResult) error {
	globalStoragePath, err := ess.getGlobalStoragePath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(globalStoragePath); os.IsNotExist(err) {
		return nil // Global storage doesn't exist
	}

	entries, err := os.ReadDir(globalStoragePath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		extensionID := entry.Name()
		extensionStoragePath := filepath.Join(globalStoragePath, extensionID)
		
		ess.scanExtensionStorageDirectory(extensionID, extensionStoragePath, "global", result)
	}

	return nil
}

// scanWorkspaceStorage scans extension workspace storage directories
func (ess *ExtensionSettingsScanner) scanWorkspaceStorage(result *ExtensionSettingsResult) error {
	workspaceStoragePath, err := utils.GetWorkspaceStoragePath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(workspaceStoragePath); os.IsNotExist(err) {
		return nil // Workspace storage doesn't exist
	}

	// Scan workspace hash directories
	workspaceEntries, err := os.ReadDir(workspaceStoragePath)
	if err != nil {
		return err
	}

	for _, workspaceEntry := range workspaceEntries {
		if !workspaceEntry.IsDir() {
			continue
		}

		workspaceHashPath := filepath.Join(workspaceStoragePath, workspaceEntry.Name())
		
		// Scan extension directories within this workspace
		extensionEntries, err := os.ReadDir(workspaceHashPath)
		if err != nil {
			continue
		}

		for _, extensionEntry := range extensionEntries {
			if !extensionEntry.IsDir() {
				continue
			}

			extensionID := extensionEntry.Name()
			extensionStoragePath := filepath.Join(workspaceHashPath, extensionID)
			
			ess.scanExtensionStorageDirectory(extensionID, extensionStoragePath, "workspace", result)
		}
	}

	return nil
}

// scanExtensionStorageDirectory scans a specific extension's storage directory
func (ess *ExtensionSettingsScanner) scanExtensionStorageDirectory(extensionID, dirPath, storageType string, result *ExtensionSettingsResult) {
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue despite errors
		}

		if info.IsDir() {
			return nil
		}

		// Skip very large files
		if info.Size() > 10*1024*1024 { // 10MB limit
			return nil
		}

		// Analyze the file
		ess.analyzeStorageFile(extensionID, path, storageType, info, result)
		return nil
	})

	if err != nil {
		// Continue despite walk errors
	}
}

// analyzeStorageFile analyzes a single storage file for telemetry data
func (ess *ExtensionSettingsScanner) analyzeStorageFile(extensionID, filePath, storageType string, info os.FileInfo, result *ExtensionSettingsResult) {
	// Determine risk based on file name and path
	fileName := strings.ToLower(info.Name())
	risk := ess.assessFileRisk(fileName, filePath)

	if risk == TelemetryRiskNone {
		return // Skip files with no telemetry risk
	}

	// Try to parse JSON files for more detailed analysis
	if strings.HasSuffix(fileName, ".json") {
		ess.analyzeJSONStorageFile(extensionID, filePath, storageType, info, risk, result)
	} else {
		// For non-JSON files, create a basic storage item
		storageItem := StorageItem{
			ExtensionID:  extensionID,
			StorageType:  storageType,
			Key:          info.Name(),
			Value:        fmt.Sprintf("Binary file (%d bytes)", info.Size()),
			Size:         info.Size(),
			Risk:         risk,
			Description:  ess.getFileDescription(fileName, risk),
			FilePath:     filePath,
			LastModified: info.ModTime(),
		}

		if storageType == "global" {
			result.GlobalStorageItems = append(result.GlobalStorageItems, storageItem)
		} else {
			result.WorkspaceStorageItems = append(result.WorkspaceStorageItems, storageItem)
		}
	}
}

// analyzeJSONStorageFile analyzes a JSON storage file in detail
func (ess *ExtensionSettingsScanner) analyzeJSONStorageFile(extensionID, filePath, storageType string, info os.FileInfo, baseRisk TelemetryRisk, result *ExtensionSettingsResult) {
	data, err := ess.loadJSONConfig(filePath)
	if err != nil {
		// If we can't parse as JSON, treat as regular file
		storageItem := StorageItem{
			ExtensionID:  extensionID,
			StorageType:  storageType,
			Key:          info.Name(),
			Value:        "Invalid JSON file",
			Size:         info.Size(),
			Risk:         baseRisk,
			Description:  "JSON file that couldn't be parsed",
			FilePath:     filePath,
			LastModified: info.ModTime(),
		}

		if storageType == "global" {
			result.GlobalStorageItems = append(result.GlobalStorageItems, storageItem)
		} else {
			result.WorkspaceStorageItems = append(result.WorkspaceStorageItems, storageItem)
		}
		return
	}

	// Analyze each key-value pair in the JSON
	ess.analyzeJSONData(data, extensionID, filePath, storageType, info.ModTime(), result)
}

// analyzeJSONData recursively analyzes JSON data for telemetry patterns
func (ess *ExtensionSettingsScanner) analyzeJSONData(data interface{}, extensionID, filePath, storageType string, lastModified time.Time, result *ExtensionSettingsResult) {
	ess.analyzeJSONRecursive(data, extensionID, filePath, storageType, "", lastModified, result)
}

// analyzeJSONRecursive recursively analyzes JSON structures
func (ess *ExtensionSettingsScanner) analyzeJSONRecursive(obj interface{}, extensionID, filePath, storageType, keyPath string, lastModified time.Time, result *ExtensionSettingsResult) {
	switch v := obj.(type) {
	case map[string]interface{}:
		for key, value := range v {
			currentPath := key
			if keyPath != "" {
				currentPath = keyPath + "." + key
			}

			// Check if this key matches telemetry patterns
			risk := ess.assessKeyRisk(key, currentPath, value)
			
			if risk > TelemetryRiskNone {
				// Calculate size estimate for this value
				size := ess.estimateValueSize(value)
				
				storageItem := StorageItem{
					ExtensionID:  extensionID,
					StorageType:  storageType,
					Key:          currentPath,
					Value:        ess.sanitizeValue(value),
					Size:         size,
					Risk:         risk,
					Description:  ess.getKeyDescription(key, risk),
					FilePath:     filePath,
					LastModified: lastModified,
				}

				if storageType == "global" {
					result.GlobalStorageItems = append(result.GlobalStorageItems, storageItem)
				} else {
					result.WorkspaceStorageItems = append(result.WorkspaceStorageItems, storageItem)
				}
			}

			// Recurse into nested objects
			ess.analyzeJSONRecursive(value, extensionID, filePath, storageType, currentPath, lastModified, result)
		}
	case []interface{}:
		// For arrays, analyze each element
		for i, item := range v {
			arrayPath := fmt.Sprintf("%s[%d]", keyPath, i)
			ess.analyzeJSONRecursive(item, extensionID, filePath, storageType, arrayPath, lastModified, result)
		}
	}
}

// Helper methods

// getVSCodeSettingsPath returns the path to VS Code user settings
func (ess *ExtensionSettingsScanner) getVSCodeSettingsPath() (string, error) {
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
func (ess *ExtensionSettingsScanner) getWorkspaceSettingsPaths() []string {
	// This is a simplified implementation - in practice, you might want to
	// scan more locations or use VS Code's workspace detection
	var paths []string
	
	homeDir, err := utils.GetHomeDir()
	if err != nil {
		return paths
	}

	// Check common project directories for .vscode/settings.json
	commonDirs := []string{
		filepath.Join(homeDir, "Documents"),
		filepath.Join(homeDir, "Projects"),
		filepath.Join(homeDir, "Development"),
		filepath.Join(homeDir, "Code"),
	}

	for _, dir := range commonDirs {
		if _, err := os.Stat(dir); err == nil {
			// This is a simplified search - could be expanded
			settingsPath := filepath.Join(dir, ".vscode", "settings.json")
			if _, err := os.Stat(settingsPath); err == nil {
				paths = append(paths, settingsPath)
			}
		}
	}

	return paths
}

// getGlobalStoragePath returns the global storage path
func (ess *ExtensionSettingsScanner) getGlobalStoragePath() (string, error) {
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
func (ess *ExtensionSettingsScanner) loadJSONConfig(filePath string) (map[string]interface{}, error) {
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

// extractExtensionSettings extracts extension settings from a configuration object
func (ess *ExtensionSettingsScanner) extractExtensionSettings(settings map[string]interface{}, source, filePath string, lastModified time.Time, result *ExtensionSettingsResult) {
	for key, value := range settings {
		// Check if this is an extension setting (typically has format: publisher.extension.setting)
		if ess.isExtensionSetting(key) {
			risk := ess.assessSettingRisk(key, value)
			
			if risk > TelemetryRiskNone {
				setting := ExtensionSetting{
					ExtensionID:  ess.extractExtensionID(key),
					SettingKey:   key,
					SettingValue: ess.sanitizeValue(value),
					Source:       source,
					Risk:         risk,
					Category:     ess.getSettingCategory(key),
					Description:  ess.getSettingDescription(key, risk),
					LastModified: lastModified,
				}
				
				result.ExtensionSettings = append(result.ExtensionSettings, setting)
			}
		}
	}
}

// isExtensionSetting determines if a setting key belongs to an extension
func (ess *ExtensionSettingsScanner) isExtensionSetting(key string) bool {
	parts := strings.Split(key, ".")
	if len(parts) < 2 {
		return false
	}

	// VS Code core settings typically start with known prefixes
	coreSettings := []string{
		"editor", "workbench", "window", "files", "search", "debug",
		"extensions", "terminal", "scm", "problems", "breadcrumbs",
		"telemetry", "update", "security", "remote", "merge-conflict",
	}

	firstPart := parts[0]
	for _, core := range coreSettings {
		if strings.EqualFold(firstPart, core) {
			return false
		}
	}

	return true
}

// extractExtensionID extracts the extension ID from a setting key
func (ess *ExtensionSettingsScanner) extractExtensionID(key string) string {
	parts := strings.Split(key, ".")
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	return parts[0]
}

// assessSettingRisk assesses the telemetry risk of a setting
func (ess *ExtensionSettingsScanner) assessSettingRisk(key string, value interface{}) TelemetryRisk {
	lowerKey := strings.ToLower(key)
	
	// Check against known telemetry patterns
	for pattern, risk := range ess.telemetryKeyPatterns {
		if strings.Contains(lowerKey, strings.ToLower(pattern)) {
			return risk
		}
	}

	return TelemetryRiskNone
}

// assessKeyRisk assesses the telemetry risk of a storage key
func (ess *ExtensionSettingsScanner) assessKeyRisk(key, fullPath string, value interface{}) TelemetryRisk {
	lowerKey := strings.ToLower(key)
	lowerPath := strings.ToLower(fullPath)
	
	// Check against storage key patterns
	for pattern, risk := range ess.storageKeyPatterns {
		if strings.Contains(lowerKey, strings.ToLower(pattern)) || 
		   strings.Contains(lowerPath, strings.ToLower(pattern)) {
			return risk
		}
	}

	// Check value content for additional patterns
	if valueStr, ok := value.(string); ok {
		lowerValue := strings.ToLower(valueStr)
		if strings.Contains(lowerValue, "telemetry") || 
		   strings.Contains(lowerValue, "analytics") ||
		   strings.Contains(lowerValue, "tracking") {
			return TelemetryRiskMedium
		}
	}

	return TelemetryRiskNone
}

// assessFileRisk assesses the telemetry risk of a file based on its name and path
func (ess *ExtensionSettingsScanner) assessFileRisk(fileName, filePath string) TelemetryRisk {
	lowerName := strings.ToLower(fileName)
	lowerPath := strings.ToLower(filePath)
	
	// High risk file patterns
	highRiskPatterns := []string{"telemetry", "analytics", "tracking", "usage", "metrics"}
	for _, pattern := range highRiskPatterns {
		if strings.Contains(lowerName, pattern) || strings.Contains(lowerPath, pattern) {
			return TelemetryRiskHigh
		}
	}

	// Medium risk file patterns
	mediumRiskPatterns := []string{"crash", "error", "log", "diagnostic", "performance"}
	for _, pattern := range mediumRiskPatterns {
		if strings.Contains(lowerName, pattern) || strings.Contains(lowerPath, pattern) {
			return TelemetryRiskMedium
		}
	}

	// Low risk file patterns
	lowRiskPatterns := []string{"cache", "temp", "config", "settings", "preferences"}
	for _, pattern := range lowRiskPatterns {
		if strings.Contains(lowerName, pattern) || strings.Contains(lowerPath, pattern) {
			return TelemetryRiskLow
		}
	}

	return TelemetryRiskNone
}

// sanitizeValue sanitizes a value for safe display (removes sensitive data)
func (ess *ExtensionSettingsScanner) sanitizeValue(value interface{}) interface{} {
	if str, ok := value.(string); ok {
		// Sanitize potential sensitive strings
		if len(str) > 100 {
			return str[:100] + "... (truncated)"
		}
		
		// Check for potential sensitive patterns and mask them
		lowerStr := strings.ToLower(str)
		if strings.Contains(lowerStr, "key") || 
		   strings.Contains(lowerStr, "token") || 
		   strings.Contains(lowerStr, "secret") ||
		   strings.Contains(lowerStr, "password") {
			return "[SENSITIVE DATA MASKED]"
		}
	}
	
	return value
}

// estimateValueSize estimates the size of a JSON value in bytes
func (ess *ExtensionSettingsScanner) estimateValueSize(value interface{}) int64 {
	data, err := json.Marshal(value)
	if err != nil {
		return 0
	}
	return int64(len(data))
}

// getSettingCategory returns the category of a setting
func (ess *ExtensionSettingsScanner) getSettingCategory(key string) string {
	if strings.Contains(strings.ToLower(key), "telemetry") {
		return "Telemetry"
	}
	if strings.Contains(strings.ToLower(key), "analytics") {
		return "Analytics"
	}
	if strings.Contains(strings.ToLower(key), "tracking") {
		return "Tracking"
	}
	return "Extension Setting"
}

// getSettingDescription returns a description for a setting
func (ess *ExtensionSettingsScanner) getSettingDescription(key string, risk TelemetryRisk) string {
	return fmt.Sprintf("Extension setting with %s telemetry risk", risk.String())
}

// getKeyDescription returns a description for a storage key
func (ess *ExtensionSettingsScanner) getKeyDescription(key string, risk TelemetryRisk) string {
	return fmt.Sprintf("Extension storage key with %s telemetry risk", risk.String())
}

// getFileDescription returns a description for a file
func (ess *ExtensionSettingsScanner) getFileDescription(fileName string, risk TelemetryRisk) string {
	return fmt.Sprintf("Extension file with %s telemetry risk", risk.String())
}

// calculateTotals calculates summary statistics
func (ess *ExtensionSettingsScanner) calculateTotals(result *ExtensionSettingsResult) {
	result.TotalSettings = len(result.ExtensionSettings)
	
	for _, setting := range result.ExtensionSettings {
		if setting.Risk >= TelemetryRiskMedium {
			result.TelemetrySettings++
		}
	}
}