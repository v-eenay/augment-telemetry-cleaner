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

// StorageAnalysisResult represents the result of comprehensive storage analysis
type StorageAnalysisResult struct {
	GlobalStorageAnalysis    GlobalStorageAnalysis    `json:"global_storage_analysis"`
	WorkspaceStorageAnalysis WorkspaceStorageAnalysis `json:"workspace_storage_analysis"`
	CacheAnalysis           CacheAnalysis            `json:"cache_analysis"`
	TempFileAnalysis        TempFileAnalysis         `json:"temp_file_analysis"`
	CrossExtensionData      []CrossExtensionData     `json:"cross_extension_data"`
	StorageStatistics       StorageStatistics        `json:"storage_statistics"`
	ScanDuration            time.Duration            `json:"scan_duration"`
}

// GlobalStorageAnalysis represents analysis of global storage
type GlobalStorageAnalysis struct {
	ExtensionStorages []ExtensionStorage `json:"extension_storages"`
	TotalSize         int64              `json:"total_size"`
	TelemetrySize     int64              `json:"telemetry_size"`
	ExtensionCount    int                `json:"extension_count"`
	TelemetryCount    int                `json:"telemetry_count"`
}

// WorkspaceStorageAnalysis represents analysis of workspace storage
type WorkspaceStorageAnalysis struct {
	WorkspaceStorages []WorkspaceStorage `json:"workspace_storages"`
	TotalSize         int64              `json:"total_size"`
	TelemetrySize     int64              `json:"telemetry_size"`
	WorkspaceCount    int                `json:"workspace_count"`
	ExtensionCount    int                `json:"extension_count"`
}

// ExtensionStorage represents storage data for a specific extension
type ExtensionStorage struct {
	ExtensionID       string              `json:"extension_id"`
	StoragePath       string              `json:"storage_path"`
	StorageItems      []StorageDataItem   `json:"storage_items"`
	TotalSize         int64               `json:"total_size"`
	TelemetrySize     int64               `json:"telemetry_size"`
	LastAccessed      time.Time           `json:"last_accessed"`
	DataCategories    []string            `json:"data_categories"`
	Risk              TelemetryRisk       `json:"risk"`
	RetentionPolicy   RetentionPolicy     `json:"retention_policy"`
}

// WorkspaceStorage represents storage data for a workspace
type WorkspaceStorage struct {
	WorkspaceHash     string            `json:"workspace_hash"`
	WorkspacePath     string            `json:"workspace_path,omitempty"`
	ExtensionStorages []ExtensionStorage `json:"extension_storages"`
	TotalSize         int64             `json:"total_size"`
	TelemetrySize     int64             `json:"telemetry_size"`
	LastAccessed      time.Time         `json:"last_accessed"`
}

// StorageDataItem represents a single item in extension storage
type StorageDataItem struct {
	Key             string        `json:"key"`
	Value           interface{}   `json:"value"`
	Size            int64         `json:"size"`
	Type            string        `json:"type"`
	Risk            TelemetryRisk `json:"risk"`
	Category        string        `json:"category"`
	Description     string        `json:"description"`
	LastModified    time.Time     `json:"last_modified"`
	AccessFrequency int           `json:"access_frequency"`
}

// CacheAnalysis represents analysis of extension cache files
type CacheAnalysis struct {
	CacheDirectories []CacheDirectory `json:"cache_directories"`
	TotalSize        int64            `json:"total_size"`
	TelemetrySize    int64            `json:"telemetry_size"`
	FileCount        int              `json:"file_count"`
	TelemetryCount   int              `json:"telemetry_count"`
}

// CacheDirectory represents a cache directory analysis
type CacheDirectory struct {
	ExtensionID   string          `json:"extension_id"`
	Path          string          `json:"path"`
	CacheFiles    []CacheFile     `json:"cache_files"`
	TotalSize     int64           `json:"total_size"`
	TelemetrySize int64           `json:"telemetry_size"`
	LastAccessed  time.Time       `json:"last_accessed"`
	CacheType     string          `json:"cache_type"`
	Risk          TelemetryRisk   `json:"risk"`
}

// CacheFile represents a single cache file
type CacheFile struct {
	Path         string        `json:"path"`
	Size         int64         `json:"size"`
	Type         string        `json:"type"`
	Risk         TelemetryRisk `json:"risk"`
	Description  string        `json:"description"`
	LastModified time.Time     `json:"last_modified"`
	LastAccessed time.Time     `json:"last_accessed"`
}

// TempFileAnalysis represents analysis of temporary files
type TempFileAnalysis struct {
	TempFiles     []TempFile `json:"temp_files"`
	TotalSize     int64      `json:"total_size"`
	TelemetrySize int64      `json:"telemetry_size"`
	FileCount     int        `json:"file_count"`
	TelemetryCount int       `json:"telemetry_count"`
}

// TempFile represents a temporary file
type TempFile struct {
	Path         string        `json:"path"`
	ExtensionID  string        `json:"extension_id,omitempty"`
	Size         int64         `json:"size"`
	Risk         TelemetryRisk `json:"risk"`
	Description  string        `json:"description"`
	LastModified time.Time     `json:"last_modified"`
	Age          time.Duration `json:"age"`
}

// CrossExtensionData represents data shared between extensions
type CrossExtensionData struct {
	DataType        string   `json:"data_type"`
	ExtensionIDs    []string `json:"extension_ids"`
	SharedKeys      []string `json:"shared_keys"`
	Risk            TelemetryRisk `json:"risk"`
	Description     string   `json:"description"`
	DataSize        int64    `json:"data_size"`
	CorrelationHash string   `json:"correlation_hash"`
}

// RetentionPolicy represents data retention information
type RetentionPolicy struct {
	HasPolicy       bool          `json:"has_policy"`
	RetentionPeriod time.Duration `json:"retention_period,omitempty"`
	LastCleanup     time.Time     `json:"last_cleanup,omitempty"`
	AutoCleanup     bool          `json:"auto_cleanup"`
	PolicySource    string        `json:"policy_source"`
}

// StorageStatistics represents overall storage statistics
type StorageStatistics struct {
	TotalStorageSize    int64   `json:"total_storage_size"`
	TelemetryStorageSize int64  `json:"telemetry_storage_size"`
	TelemetryPercentage float64 `json:"telemetry_percentage"`
	ExtensionCount      int     `json:"extension_count"`
	WorkspaceCount      int     `json:"workspace_count"`
	CacheSize           int64   `json:"cache_size"`
	TempFileSize        int64   `json:"temp_file_size"`
	OldestData          time.Time `json:"oldest_data"`
	NewestData          time.Time `json:"newest_data"`
}

// StorageAnalyzer handles comprehensive analysis of extension storage
type StorageAnalyzer struct {
	telemetryPatterns    map[string]TelemetryRisk
	cachePatterns        map[string]TelemetryRisk
	retentionAnalyzer    *RetentionAnalyzer
	correlationAnalyzer  *CorrelationAnalyzer
}

// NewStorageAnalyzer creates a new storage analyzer
func NewStorageAnalyzer() *StorageAnalyzer {
	analyzer := &StorageAnalyzer{
		retentionAnalyzer:   NewRetentionAnalyzer(),
		correlationAnalyzer: NewCorrelationAnalyzer(),
	}
	analyzer.initializeTelemetryPatterns()
	analyzer.initializeCachePatterns()
	return analyzer
}

// initializeTelemetryPatterns sets up patterns for telemetry data detection
func (sa *StorageAnalyzer) initializeTelemetryPatterns() {
	sa.telemetryPatterns = map[string]TelemetryRisk{
		// High-risk telemetry data
		"telemetryData":        TelemetryRiskCritical,
		"analyticsData":        TelemetryRiskCritical,
		"trackingData":         TelemetryRiskCritical,
		"machineId":            TelemetryRiskCritical,
		"deviceId":             TelemetryRiskCritical,
		"sessionId":            TelemetryRiskHigh,
		"userId":               TelemetryRiskHigh,
		"installId":            TelemetryRiskHigh,
		
		// Usage and metrics data
		"usageStats":           TelemetryRiskHigh,
		"userMetrics":          TelemetryRiskHigh,
		"performanceMetrics":   TelemetryRiskMedium,
		"featureUsage":         TelemetryRiskMedium,
		"commandUsage":         TelemetryRiskMedium,
		"activationCount":      TelemetryRiskLow,
		
		// Error and diagnostic data
		"crashReports":         TelemetryRiskMedium,
		"errorLogs":            TelemetryRiskMedium,
		"diagnosticData":       TelemetryRiskMedium,
		"debugInfo":            TelemetryRiskLow,
		
		// User behavior data
		"searchHistory":        TelemetryRiskMedium,
		"commandHistory":       TelemetryRiskMedium,
		"navigationHistory":    TelemetryRiskMedium,
		"recentFiles":          TelemetryRiskLow,
		"preferences":          TelemetryRiskLow,
		
		// Experiment and survey data
		"experimentData":       TelemetryRiskMedium,
		"surveyResponses":      TelemetryRiskMedium,
		"feedbackData":         TelemetryRiskLow,
		"betaFeatures":         TelemetryRiskLow,
		
		// Network and communication data
		"apiKeys":              TelemetryRiskHigh,
		"authTokens":           TelemetryRiskHigh,
		"serverEndpoints":      TelemetryRiskMedium,
		"networkLogs":          TelemetryRiskMedium,
	}
}

// initializeCachePatterns sets up patterns for cache file analysis
func (sa *StorageAnalyzer) initializeCachePatterns() {
	sa.cachePatterns = map[string]TelemetryRisk{
		// High-risk cache patterns
		"telemetry":            TelemetryRiskHigh,
		"analytics":            TelemetryRiskHigh,
		"tracking":             TelemetryRiskHigh,
		"usage":                TelemetryRiskMedium,
		"metrics":              TelemetryRiskMedium,
		
		// Network cache patterns
		"http":                 TelemetryRiskMedium,
		"api":                  TelemetryRiskMedium,
		"request":              TelemetryRiskMedium,
		"response":             TelemetryRiskLow,
		
		// User data cache patterns
		"user":                 TelemetryRiskMedium,
		"session":              TelemetryRiskMedium,
		"auth":                 TelemetryRiskHigh,
		"token":                TelemetryRiskHigh,
		
		// General cache patterns
		"cache":                TelemetryRiskLow,
		"temp":                 TelemetryRiskLow,
		"tmp":                  TelemetryRiskLow,
		"log":                  TelemetryRiskMedium,
	}
}

// AnalyzeStorage performs comprehensive storage analysis
func (sa *StorageAnalyzer) AnalyzeStorage() (*StorageAnalysisResult, error) {
	startTime := time.Now()
	
	result := &StorageAnalysisResult{
		CrossExtensionData: make([]CrossExtensionData, 0),
	}

	// Analyze global storage
	globalAnalysis, err := sa.analyzeGlobalStorage()
	if err != nil {
		return nil, fmt.Errorf("failed to analyze global storage: %w", err)
	}
	result.GlobalStorageAnalysis = *globalAnalysis

	// Analyze workspace storage
	workspaceAnalysis, err := sa.analyzeWorkspaceStorage()
	if err != nil {
		return nil, fmt.Errorf("failed to analyze workspace storage: %w", err)
	}
	result.WorkspaceStorageAnalysis = *workspaceAnalysis

	// Analyze cache files
	cacheAnalysis, err := sa.analyzeCacheFiles()
	if err != nil {
		// Continue even if cache analysis fails
		result.CacheAnalysis = CacheAnalysis{}
	} else {
		result.CacheAnalysis = *cacheAnalysis
	}

	// Analyze temporary files
	tempAnalysis, err := sa.analyzeTempFiles()
	if err != nil {
		// Continue even if temp file analysis fails
		result.TempFileAnalysis = TempFileAnalysis{}
	} else {
		result.TempFileAnalysis = *tempAnalysis
	}

	// Perform cross-extension correlation analysis
	crossExtensionData := sa.correlationAnalyzer.AnalyzeCrossExtensionData(
		result.GlobalStorageAnalysis.ExtensionStorages,
		result.WorkspaceStorageAnalysis.WorkspaceStorages,
	)
	result.CrossExtensionData = crossExtensionData

	// Calculate overall statistics
	result.StorageStatistics = sa.calculateStorageStatistics(result)
	result.ScanDuration = time.Since(startTime)

	return result, nil
}

// analyzeGlobalStorage analyzes global storage for all extensions
func (sa *StorageAnalyzer) analyzeGlobalStorage() (*GlobalStorageAnalysis, error) {
	globalStoragePath, err := sa.getGlobalStoragePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get global storage path: %w", err)
	}

	analysis := &GlobalStorageAnalysis{
		ExtensionStorages: make([]ExtensionStorage, 0),
	}

	if _, err := os.Stat(globalStoragePath); os.IsNotExist(err) {
		return analysis, nil // No global storage directory
	}

	entries, err := os.ReadDir(globalStoragePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read global storage directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		extensionID := entry.Name()
		extensionStoragePath := filepath.Join(globalStoragePath, extensionID)
		
		extensionStorage, err := sa.analyzeExtensionStorage(extensionID, extensionStoragePath, "global")
		if err != nil {
			continue // Skip extensions we can't analyze
		}

		analysis.ExtensionStorages = append(analysis.ExtensionStorages, *extensionStorage)
		analysis.TotalSize += extensionStorage.TotalSize
		analysis.TelemetrySize += extensionStorage.TelemetrySize
		
		if extensionStorage.Risk >= TelemetryRiskMedium {
			analysis.TelemetryCount++
		}
	}

	analysis.ExtensionCount = len(analysis.ExtensionStorages)
	return analysis, nil
}

// analyzeWorkspaceStorage analyzes workspace storage for all workspaces
func (sa *StorageAnalyzer) analyzeWorkspaceStorage() (*WorkspaceStorageAnalysis, error) {
	workspaceStoragePath, err := utils.GetWorkspaceStoragePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace storage path: %w", err)
	}

	analysis := &WorkspaceStorageAnalysis{
		WorkspaceStorages: make([]WorkspaceStorage, 0),
	}

	if _, err := os.Stat(workspaceStoragePath); os.IsNotExist(err) {
		return analysis, nil // No workspace storage directory
	}

	workspaceEntries, err := os.ReadDir(workspaceStoragePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read workspace storage directory: %w", err)
	}

	for _, workspaceEntry := range workspaceEntries {
		if !workspaceEntry.IsDir() {
			continue
		}

		workspaceHash := workspaceEntry.Name()
		workspaceHashPath := filepath.Join(workspaceStoragePath, workspaceHash)
		
		workspaceStorage, err := sa.analyzeWorkspaceStorageDirectory(workspaceHash, workspaceHashPath)
		if err != nil {
			continue // Skip workspaces we can't analyze
		}

		analysis.WorkspaceStorages = append(analysis.WorkspaceStorages, *workspaceStorage)
		analysis.TotalSize += workspaceStorage.TotalSize
		analysis.TelemetrySize += workspaceStorage.TelemetrySize
	}

	analysis.WorkspaceCount = len(analysis.WorkspaceStorages)
	
	// Count unique extensions across all workspaces
	extensionSet := make(map[string]bool)
	for _, workspace := range analysis.WorkspaceStorages {
		for _, ext := range workspace.ExtensionStorages {
			extensionSet[ext.ExtensionID] = true
		}
	}
	analysis.ExtensionCount = len(extensionSet)

	return analysis, nil
}

// analyzeWorkspaceStorageDirectory analyzes a specific workspace storage directory
func (sa *StorageAnalyzer) analyzeWorkspaceStorageDirectory(workspaceHash, workspaceHashPath string) (*WorkspaceStorage, error) {
	workspaceStorage := &WorkspaceStorage{
		WorkspaceHash:     workspaceHash,
		ExtensionStorages: make([]ExtensionStorage, 0),
	}

	// Try to determine workspace path from hash (this is complex and may not always work)
	workspaceStorage.WorkspacePath = sa.resolveWorkspacePath(workspaceHash)

	extensionEntries, err := os.ReadDir(workspaceHashPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read workspace hash directory: %w", err)
	}

	var latestAccess time.Time
	for _, extensionEntry := range extensionEntries {
		if !extensionEntry.IsDir() {
			continue
		}

		extensionID := extensionEntry.Name()
		extensionStoragePath := filepath.Join(workspaceHashPath, extensionID)
		
		extensionStorage, err := sa.analyzeExtensionStorage(extensionID, extensionStoragePath, "workspace")
		if err != nil {
			continue // Skip extensions we can't analyze
		}

		workspaceStorage.ExtensionStorages = append(workspaceStorage.ExtensionStorages, *extensionStorage)
		workspaceStorage.TotalSize += extensionStorage.TotalSize
		workspaceStorage.TelemetrySize += extensionStorage.TelemetrySize
		
		if extensionStorage.LastAccessed.After(latestAccess) {
			latestAccess = extensionStorage.LastAccessed
		}
	}

	workspaceStorage.LastAccessed = latestAccess
	return workspaceStorage, nil
}

// analyzeExtensionStorage analyzes storage for a specific extension
func (sa *StorageAnalyzer) analyzeExtensionStorage(extensionID, storagePath, storageType string) (*ExtensionStorage, error) {
	storage := &ExtensionStorage{
		ExtensionID:    extensionID,
		StoragePath:    storagePath,
		StorageItems:   make([]StorageDataItem, 0),
		DataCategories: make([]string, 0),
	}

	// Get directory info
	dirInfo, err := os.Stat(storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage directory info: %w", err)
	}
	storage.LastAccessed = dirInfo.ModTime()

	// Analyze retention policy
	storage.RetentionPolicy = sa.retentionAnalyzer.AnalyzeRetentionPolicy(extensionID, storagePath)

	// Walk through all files in the storage directory
	err = filepath.Walk(storagePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue despite errors
		}

		if info.IsDir() {
			return nil
		}

		// Analyze the file
		sa.analyzeStorageFile(path, info, storage)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk storage directory: %w", err)
	}

	// Determine overall risk level
	storage.Risk = sa.calculateStorageRisk(storage.StorageItems)

	// Extract unique data categories
	categorySet := make(map[string]bool)
	for _, item := range storage.StorageItems {
		if item.Category != "" {
			categorySet[item.Category] = true
		}
	}
	for category := range categorySet {
		storage.DataCategories = append(storage.DataCategories, category)
	}

	return storage, nil
}

// analyzeStorageFile analyzes a single storage file
func (sa *StorageAnalyzer) analyzeStorageFile(filePath string, info os.FileInfo, storage *ExtensionStorage) {
	storage.TotalSize += info.Size()

	fileName := strings.ToLower(info.Name())
	
	// Determine file risk based on name and content
	risk := sa.assessFileRisk(fileName, filePath)
	
	if risk == TelemetryRiskNone {
		return // Skip files with no telemetry risk
	}

	// For JSON files, analyze content in detail
	if strings.HasSuffix(fileName, ".json") {
		sa.analyzeJSONStorageFile(filePath, info, storage)
	} else {
		// For non-JSON files, create basic storage item
		item := StorageDataItem{
			Key:             info.Name(),
			Value:           fmt.Sprintf("Binary file (%d bytes)", info.Size()),
			Size:            info.Size(),
			Type:            "file",
			Risk:            risk,
			Category:        sa.categorizeFile(fileName),
			Description:     sa.getFileDescription(fileName, risk),
			LastModified:    info.ModTime(),
			AccessFrequency: sa.estimateAccessFrequency(info),
		}
		
		storage.StorageItems = append(storage.StorageItems, item)
		
		if risk >= TelemetryRiskMedium {
			storage.TelemetrySize += info.Size()
		}
	}
}

// analyzeJSONStorageFile analyzes a JSON storage file in detail
func (sa *StorageAnalyzer) analyzeJSONStorageFile(filePath string, info os.FileInfo, storage *ExtensionStorage) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return // Skip files we can't read
	}

	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return // Skip files we can't parse
	}

	// Analyze JSON structure recursively
	sa.analyzeJSONData(jsonData, filepath.Base(filePath), "", info, storage)
}

// analyzeJSONData recursively analyzes JSON data structure
func (sa *StorageAnalyzer) analyzeJSONData(data interface{}, fileName, keyPath string, info os.FileInfo, storage *ExtensionStorage) {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			currentPath := key
			if keyPath != "" {
				currentPath = keyPath + "." + key
			}
			
			risk := sa.assessKeyRisk(key, currentPath, value)
			if risk > TelemetryRiskNone {
				item := StorageDataItem{
					Key:             currentPath,
					Value:           sa.sanitizeValue(value),
					Size:            sa.estimateValueSize(value),
					Type:            "json_key",
					Risk:            risk,
					Category:        sa.categorizeKey(key),
					Description:     sa.getKeyDescription(key, risk),
					LastModified:    info.ModTime(),
					AccessFrequency: sa.estimateAccessFrequency(info),
				}
				
				storage.StorageItems = append(storage.StorageItems, item)
				
				if risk >= TelemetryRiskMedium {
					storage.TelemetrySize += item.Size
				}
			}
			
			// Recurse into nested objects
			sa.analyzeJSONData(value, fileName, currentPath, info, storage)
		}
	case []interface{}:
		// For arrays, analyze each element
		for i, item := range v {
			arrayPath := fmt.Sprintf("%s[%d]", keyPath, i)
			sa.analyzeJSONData(item, fileName, arrayPath, info, storage)
		}
	}
}

// Helper methods for storage analysis

// getGlobalStoragePath returns the global storage path
func (sa *StorageAnalyzer) getGlobalStoragePath() (string, error) {
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

// resolveWorkspacePath attempts to resolve workspace path from hash
func (sa *StorageAnalyzer) resolveWorkspacePath(workspaceHash string) string {
	// This is a simplified implementation
	// In practice, VS Code uses a complex hashing algorithm
	// and the reverse mapping is not straightforward
	return fmt.Sprintf("Unknown workspace (hash: %s)", workspaceHash[:8])
}

// assessFileRisk assesses the telemetry risk of a file
func (sa *StorageAnalyzer) assessFileRisk(fileName, filePath string) TelemetryRisk {
	lowerName := strings.ToLower(fileName)
	lowerPath := strings.ToLower(filePath)
	
	// Check against telemetry patterns
	maxRisk := TelemetryRiskNone
	for pattern, risk := range sa.telemetryPatterns {
		if strings.Contains(lowerName, strings.ToLower(pattern)) ||
		   strings.Contains(lowerPath, strings.ToLower(pattern)) {
			if risk > maxRisk {
				maxRisk = risk
			}
		}
	}
	
	return maxRisk
}

// assessKeyRisk assesses the telemetry risk of a JSON key
func (sa *StorageAnalyzer) assessKeyRisk(key, fullPath string, value interface{}) TelemetryRisk {
	lowerKey := strings.ToLower(key)
	lowerPath := strings.ToLower(fullPath)
	
	// Check against telemetry patterns
	maxRisk := TelemetryRiskNone
	for pattern, risk := range sa.telemetryPatterns {
		if strings.Contains(lowerKey, strings.ToLower(pattern)) ||
		   strings.Contains(lowerPath, strings.ToLower(pattern)) {
			if risk > maxRisk {
				maxRisk = risk
			}
		}
	}
	
	// Check value content for additional patterns
	if valueStr, ok := value.(string); ok {
		lowerValue := strings.ToLower(valueStr)
		for pattern, risk := range sa.telemetryPatterns {
			if strings.Contains(lowerValue, strings.ToLower(pattern)) {
				if risk > maxRisk {
					maxRisk = risk
				}
			}
		}
	}
	
	return maxRisk
}

// calculateStorageRisk calculates overall risk for extension storage
func (sa *StorageAnalyzer) calculateStorageRisk(items []StorageDataItem) TelemetryRisk {
	maxRisk := TelemetryRiskNone
	for _, item := range items {
		if item.Risk > maxRisk {
			maxRisk = item.Risk
		}
	}
	return maxRisk
}

// categorizeFile categorizes a file based on its name
func (sa *StorageAnalyzer) categorizeFile(fileName string) string {
	lowerName := strings.ToLower(fileName)
	
	if strings.Contains(lowerName, "telemetry") {
		return "Telemetry"
	}
	if strings.Contains(lowerName, "analytics") {
		return "Analytics"
	}
	if strings.Contains(lowerName, "cache") {
		return "Cache"
	}
	if strings.Contains(lowerName, "log") {
		return "Logging"
	}
	if strings.Contains(lowerName, "config") {
		return "Configuration"
	}
	
	return "General"
}

// categorizeKey categorizes a JSON key
func (sa *StorageAnalyzer) categorizeKey(key string) string {
	lowerKey := strings.ToLower(key)
	
	if strings.Contains(lowerKey, "telemetry") {
		return "Telemetry"
	}
	if strings.Contains(lowerKey, "usage") {
		return "Usage Tracking"
	}
	if strings.Contains(lowerKey, "performance") {
		return "Performance"
	}
	if strings.Contains(lowerKey, "error") {
		return "Error Reporting"
	}
	
	return "Data"
}

// getFileDescription returns a description for a file
func (sa *StorageAnalyzer) getFileDescription(fileName string, risk TelemetryRisk) string {
	return fmt.Sprintf("Storage file with %s telemetry risk", risk.String())
}

// getKeyDescription returns a description for a JSON key
func (sa *StorageAnalyzer) getKeyDescription(key string, risk TelemetryRisk) string {
	return fmt.Sprintf("Storage key with %s telemetry risk", risk.String())
}

// sanitizeValue sanitizes a value for safe display
func (sa *StorageAnalyzer) sanitizeValue(value interface{}) interface{} {
	if str, ok := value.(string); ok {
		if len(str) > 100 {
			return str[:100] + "... (truncated)"
		}
		
		// Mask potentially sensitive data
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

// estimateValueSize estimates the size of a JSON value
func (sa *StorageAnalyzer) estimateValueSize(value interface{}) int64 {
	data, err := json.Marshal(value)
	if err != nil {
		return 0
	}
	return int64(len(data))
}

// estimateAccessFrequency estimates how frequently a file is accessed
func (sa *StorageAnalyzer) estimateAccessFrequency(info os.FileInfo) int {
	// Simple heuristic based on file age and size
	age := time.Since(info.ModTime())
	
	if age < 24*time.Hour {
		return 10 // High frequency
	} else if age < 7*24*time.Hour {
		return 5 // Medium frequency
	} else if age < 30*24*time.Hour {
		return 2 // Low frequency
	}
	
	return 1 // Very low frequency
}

// calculateStorageStatistics calculates overall storage statistics
func (sa *StorageAnalyzer) calculateStorageStatistics(result *StorageAnalysisResult) StorageStatistics {
	stats := StorageStatistics{}
	
	// Calculate totals from all storage types
	stats.TotalStorageSize = result.GlobalStorageAnalysis.TotalSize +
		result.WorkspaceStorageAnalysis.TotalSize +
		result.CacheAnalysis.TotalSize +
		result.TempFileAnalysis.TotalSize
	
	stats.TelemetryStorageSize = result.GlobalStorageAnalysis.TelemetrySize +
		result.WorkspaceStorageAnalysis.TelemetrySize +
		result.CacheAnalysis.TelemetrySize +
		result.TempFileAnalysis.TelemetrySize
	
	if stats.TotalStorageSize > 0 {
		stats.TelemetryPercentage = float64(stats.TelemetryStorageSize) / float64(stats.TotalStorageSize) * 100
	}
	
	stats.ExtensionCount = result.GlobalStorageAnalysis.ExtensionCount
	stats.WorkspaceCount = result.WorkspaceStorageAnalysis.WorkspaceCount
	stats.CacheSize = result.CacheAnalysis.TotalSize
	stats.TempFileSize = result.TempFileAnalysis.TotalSize
	
	// Find oldest and newest data
	stats.OldestData = time.Now()
	stats.NewestData = time.Time{}
	
	for _, ext := range result.GlobalStorageAnalysis.ExtensionStorages {
		if ext.LastAccessed.Before(stats.OldestData) {
			stats.OldestData = ext.LastAccessed
		}
		if ext.LastAccessed.After(stats.NewestData) {
			stats.NewestData = ext.LastAccessed
		}
	}
	
	return stats
}

// analyzeCacheFiles analyzes extension cache files
func (sa *StorageAnalyzer) analyzeCacheFiles() (*CacheAnalysis, error) {
	analysis := &CacheAnalysis{
		CacheDirectories: make([]CacheDirectory, 0),
	}

	// Get common cache directories
	cacheDirectories := sa.getCacheDirectories()

	for _, cacheDir := range cacheDirectories {
		if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
			continue
		}

		// Analyze each cache directory
		cacheAnalysis, err := sa.analyzeCacheDirectory(cacheDir)
		if err != nil {
			continue // Skip directories we can't analyze
		}

		if cacheAnalysis != nil {
			analysis.CacheDirectories = append(analysis.CacheDirectories, *cacheAnalysis)
			analysis.TotalSize += cacheAnalysis.TotalSize
			analysis.TelemetrySize += cacheAnalysis.TelemetrySize
			analysis.FileCount += len(cacheAnalysis.CacheFiles)
			
			if cacheAnalysis.Risk >= TelemetryRiskMedium {
				analysis.TelemetryCount++
			}
		}
	}

	return analysis, nil
}

// analyzeTempFiles analyzes temporary files created by extensions
func (sa *StorageAnalyzer) analyzeTempFiles() (*TempFileAnalysis, error) {
	analysis := &TempFileAnalysis{
		TempFiles: make([]TempFile, 0),
	}

	// Get common temp directories
	tempDirectories := sa.getTempDirectories()

	for _, tempDir := range tempDirectories {
		if _, err := os.Stat(tempDir); os.IsNotExist(err) {
			continue
		}

		// Analyze temp files in directory
		err := filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Continue despite errors
			}

			if info.IsDir() {
				return nil
			}

			// Check if file is extension-related and has telemetry risk
			if tempFile := sa.analyzeTempFile(path, info); tempFile != nil {
				analysis.TempFiles = append(analysis.TempFiles, *tempFile)
				analysis.TotalSize += tempFile.Size
				analysis.FileCount++
				
				if tempFile.Risk >= TelemetryRiskMedium {
					analysis.TelemetrySize += tempFile.Size
					analysis.TelemetryCount++
				}
			}

			return nil
		})

		if err != nil {
			continue // Skip directories we can't walk
		}
	}

	return analysis, nil
}

// getCacheDirectories returns common cache directories for extensions
func (sa *StorageAnalyzer) getCacheDirectories() []string {
	var directories []string

	homeDir, err := utils.GetHomeDir()
	if err != nil {
		return directories
	}

	switch utils.GetOS() {
	case "windows":
		// Windows cache locations
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			localAppData = filepath.Join(homeDir, "AppData", "Local")
		}
		
		directories = append(directories,
			filepath.Join(localAppData, "Microsoft", "vscode-cpptools"),
			filepath.Join(localAppData, "Microsoft", "vscode-eslint"),
			filepath.Join(localAppData, "Microsoft", "vscode-typescript"),
			filepath.Join(localAppData, "vscode-extensions-cache"),
			filepath.Join(os.TempDir(), "vscode-extensions"),
		)

	case "darwin":
		// macOS cache locations
		directories = append(directories,
			filepath.Join(homeDir, "Library", "Caches", "com.microsoft.VSCode"),
			filepath.Join(homeDir, "Library", "Caches", "vscode-extensions"),
			filepath.Join("/tmp", "vscode-extensions"),
		)

	default: // Linux
		// Linux cache locations
		xdgCache := os.Getenv("XDG_CACHE_HOME")
		if xdgCache == "" {
			xdgCache = filepath.Join(homeDir, ".cache")
		}
		
		directories = append(directories,
			filepath.Join(xdgCache, "vscode-extensions"),
			filepath.Join("/tmp", "vscode-extensions"),
			filepath.Join(homeDir, ".vscode-server", "data", "logs"),
		)
	}

	return directories
}

// getTempDirectories returns common temporary directories
func (sa *StorageAnalyzer) getTempDirectories() []string {
	var directories []string

	// System temp directory
	directories = append(directories, os.TempDir())

	// User-specific temp directories
	homeDir, err := utils.GetHomeDir()
	if err == nil {
		switch utils.GetOS() {
		case "windows":
			directories = append(directories,
				filepath.Join(homeDir, "AppData", "Local", "Temp"),
			)
		case "darwin":
			directories = append(directories,
				filepath.Join(homeDir, "Library", "Caches", "TemporaryItems"),
			)
		default: // Linux
			directories = append(directories,
				filepath.Join("/var", "tmp"),
				filepath.Join(homeDir, ".tmp"),
			)
		}
	}

	return directories
}

// analyzeCacheDirectory analyzes a specific cache directory
func (sa *StorageAnalyzer) analyzeCacheDirectory(cacheDir string) (*CacheDirectory, error) {
	// Try to determine which extension this cache belongs to
	extensionID := sa.inferExtensionFromPath(cacheDir)
	
	cacheDirectory := &CacheDirectory{
		ExtensionID: extensionID,
		Path:        cacheDir,
		CacheFiles:  make([]CacheFile, 0),
		CacheType:   sa.inferCacheType(cacheDir),
	}

	// Get directory info
	dirInfo, err := os.Stat(cacheDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get cache directory info: %w", err)
	}
	cacheDirectory.LastAccessed = dirInfo.ModTime()

	// Walk through cache files
	err = filepath.Walk(cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue despite errors
		}

		if info.IsDir() {
			return nil
		}

		// Analyze cache file
		if cacheFile := sa.analyzeCacheFile(path, info); cacheFile != nil {
			cacheDirectory.CacheFiles = append(cacheDirectory.CacheFiles, *cacheFile)
			cacheDirectory.TotalSize += cacheFile.Size
			
			if cacheFile.Risk >= TelemetryRiskMedium {
				cacheDirectory.TelemetrySize += cacheFile.Size
			}
			
			// Update directory risk based on files
			if cacheFile.Risk > cacheDirectory.Risk {
				cacheDirectory.Risk = cacheFile.Risk
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk cache directory: %w", err)
	}

	return cacheDirectory, nil
}

// analyzeCacheFile analyzes a single cache file
func (sa *StorageAnalyzer) analyzeCacheFile(filePath string, info os.FileInfo) *CacheFile {
	fileName := strings.ToLower(info.Name())
	
	// Assess risk based on file name and path
	risk := sa.assessCacheFileRisk(fileName, filePath)
	
	if risk == TelemetryRiskNone {
		return nil // Skip files with no telemetry risk
	}

	cacheFile := &CacheFile{
		Path:         filePath,
		Size:         info.Size(),
		Type:         sa.inferFileType(fileName),
		Risk:         risk,
		Description:  sa.getCacheFileDescription(fileName, risk),
		LastModified: info.ModTime(),
		LastAccessed: info.ModTime(), // Approximation
	}

	return cacheFile
}

// analyzeTempFile analyzes a single temporary file
func (sa *StorageAnalyzer) analyzeTempFile(filePath string, info os.FileInfo) *TempFile {
	fileName := strings.ToLower(info.Name())
	
	// Check if file is extension-related
	if !sa.isExtensionRelated(fileName, filePath) {
		return nil
	}

	// Assess risk
	risk := sa.assessTempFileRisk(fileName, filePath)
	
	if risk == TelemetryRiskNone {
		return nil
	}

	tempFile := &TempFile{
		Path:         filePath,
		ExtensionID:  sa.inferExtensionFromPath(filePath),
		Size:         info.Size(),
		Risk:         risk,
		Description:  sa.getTempFileDescription(fileName, risk),
		LastModified: info.ModTime(),
		Age:          time.Since(info.ModTime()),
	}

	return tempFile
}

// Helper methods for cache and temp file analysis

// inferExtensionFromPath tries to infer extension ID from file path
func (sa *StorageAnalyzer) inferExtensionFromPath(path string) string {
	lowerPath := strings.ToLower(path)
	
	// Common extension patterns in paths
	extensionPatterns := map[string]string{
		"cpptools":    "ms-vscode.cpptools",
		"eslint":      "dbaeumer.vscode-eslint",
		"typescript":  "vscode.typescript-language-features",
		"python":      "ms-python.python",
		"java":        "redhat.java",
		"go":          "golang.go",
		"docker":      "ms-azuretools.vscode-docker",
		"git":         "vscode.git",
		"markdown":    "vscode.markdown-language-features",
	}

	for pattern, extensionID := range extensionPatterns {
		if strings.Contains(lowerPath, pattern) {
			return extensionID
		}
	}

	return "unknown"
}

// inferCacheType infers the type of cache from directory path
func (sa *StorageAnalyzer) inferCacheType(path string) string {
	lowerPath := strings.ToLower(path)
	
	if strings.Contains(lowerPath, "log") {
		return "logs"
	}
	if strings.Contains(lowerPath, "temp") {
		return "temporary"
	}
	if strings.Contains(lowerPath, "data") {
		return "data"
	}
	if strings.Contains(lowerPath, "cache") {
		return "cache"
	}
	
	return "general"
}

// inferFileType infers file type from name
func (sa *StorageAnalyzer) inferFileType(fileName string) string {
	if strings.HasSuffix(fileName, ".log") {
		return "log"
	}
	if strings.HasSuffix(fileName, ".json") {
		return "json"
	}
	if strings.HasSuffix(fileName, ".tmp") {
		return "temporary"
	}
	if strings.HasSuffix(fileName, ".cache") {
		return "cache"
	}
	
	return "binary"
}

// assessCacheFileRisk assesses the telemetry risk of a cache file
func (sa *StorageAnalyzer) assessCacheFileRisk(fileName, filePath string) TelemetryRisk {
	lowerName := strings.ToLower(fileName)
	lowerPath := strings.ToLower(filePath)
	
	// Check against cache patterns
	maxRisk := TelemetryRiskNone
	for pattern, risk := range sa.cachePatterns {
		if strings.Contains(lowerName, strings.ToLower(pattern)) ||
		   strings.Contains(lowerPath, strings.ToLower(pattern)) {
			if risk > maxRisk {
				maxRisk = risk
			}
		}
	}
	
	return maxRisk
}

// assessTempFileRisk assesses the telemetry risk of a temporary file
func (sa *StorageAnalyzer) assessTempFileRisk(fileName, filePath string) TelemetryRisk {
	// Use same logic as cache files for now
	return sa.assessCacheFileRisk(fileName, filePath)
}

// isExtensionRelated checks if a file is related to VS Code extensions
func (sa *StorageAnalyzer) isExtensionRelated(fileName, filePath string) bool {
	lowerName := strings.ToLower(fileName)
	lowerPath := strings.ToLower(filePath)
	
	// Check for VS Code or extension-related patterns
	patterns := []string{
		"vscode", "extension", "code-server", "ms-", "redhat", "golang",
		"python", "typescript", "eslint", "prettier", "docker", "git",
	}
	
	for _, pattern := range patterns {
		if strings.Contains(lowerName, pattern) || strings.Contains(lowerPath, pattern) {
			return true
		}
	}
	
	return false
}

// getCacheFileDescription returns a description for a cache file
func (sa *StorageAnalyzer) getCacheFileDescription(fileName string, risk TelemetryRisk) string {
	return fmt.Sprintf("Cache file with %s telemetry risk", risk.String())
}

// getTempFileDescription returns a description for a temporary file
func (sa *StorageAnalyzer) getTempFileDescription(fileName string, risk TelemetryRisk) string {
	return fmt.Sprintf("Temporary file with %s telemetry risk", risk.String())
}