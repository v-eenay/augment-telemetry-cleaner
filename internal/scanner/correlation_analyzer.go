package scanner

import (
	"crypto/md5"
	"fmt"
	"strings"
	"time"
)

// CorrelationAnalyzer analyzes data correlations between extensions
type CorrelationAnalyzer struct {
	correlationPatterns map[string]CorrelationPattern
	sharedDataTypes     map[string]SharedDataType
}

// CorrelationPattern represents a pattern for detecting shared data
type CorrelationPattern struct {
	Name        string   `json:"name"`
	KeyPatterns []string `json:"key_patterns"`
	ValuePatterns []string `json:"value_patterns"`
	Risk        TelemetryRisk `json:"risk"`
	Description string   `json:"description"`
}

// SharedDataType represents a type of data that might be shared between extensions
type SharedDataType struct {
	Name        string        `json:"name"`
	Risk        TelemetryRisk `json:"risk"`
	Description string        `json:"description"`
	Examples    []string      `json:"examples"`
}

// DataCorrelation represents a correlation between extension data
type DataCorrelation struct {
	CorrelationType string            `json:"correlation_type"`
	ExtensionIDs    []string          `json:"extension_ids"`
	SharedKeys      []string          `json:"shared_keys"`
	SharedValues    []CorrelatedValue `json:"shared_values"`
	Risk            TelemetryRisk     `json:"risk"`
	Confidence      float64           `json:"confidence"`
	Description     string            `json:"description"`
	DataSize        int64             `json:"data_size"`
	LastSeen        time.Time         `json:"last_seen"`
}

// CorrelatedValue represents a value that appears across multiple extensions
type CorrelatedValue struct {
	Value       interface{} `json:"value"`
	Hash        string      `json:"hash"`
	Extensions  []string    `json:"extensions"`
	Keys        []string    `json:"keys"`
	Risk        TelemetryRisk `json:"risk"`
	Description string      `json:"description"`
}

// NewCorrelationAnalyzer creates a new correlation analyzer
func NewCorrelationAnalyzer() *CorrelationAnalyzer {
	analyzer := &CorrelationAnalyzer{}
	analyzer.initializeCorrelationPatterns()
	analyzer.initializeSharedDataTypes()
	return analyzer
}

// initializeCorrelationPatterns sets up patterns for detecting correlated data
func (ca *CorrelationAnalyzer) initializeCorrelationPatterns() {
	ca.correlationPatterns = map[string]CorrelationPattern{
		"machine_identification": {
			Name: "Machine Identification",
			KeyPatterns: []string{
				"machineId", "machine_id", "deviceId", "device_id",
				"installId", "install_id", "sessionId", "session_id",
			},
			ValuePatterns: []string{
				// Patterns for UUID-like values
				`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`,
				// Patterns for hex strings
				`[0-9a-f]{32,64}`,
			},
			Risk:        TelemetryRiskCritical,
			Description: "Machine or device identification data shared between extensions",
		},
		
		"user_identification": {
			Name: "User Identification",
			KeyPatterns: []string{
				"userId", "user_id", "username", "userEmail", "user_email",
				"accountId", "account_id", "profileId", "profile_id",
			},
			ValuePatterns: []string{
				// Email patterns
				`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`,
				// Username patterns
				`[a-zA-Z0-9_-]{3,}`,
			},
			Risk:        TelemetryRiskHigh,
			Description: "User identification data shared between extensions",
		},
		
		"telemetry_endpoints": {
			Name: "Telemetry Endpoints",
			KeyPatterns: []string{
				"telemetryUrl", "telemetry_url", "analyticsUrl", "analytics_url",
				"trackingUrl", "tracking_url", "endpoint", "apiEndpoint",
			},
			ValuePatterns: []string{
				// URL patterns
				`https?://[a-zA-Z0-9.-]+/.*`,
				// Domain patterns
				`[a-zA-Z0-9.-]+\.(com|net|org|io)`,
			},
			Risk:        TelemetryRiskHigh,
			Description: "Telemetry or analytics endpoints shared between extensions",
		},
		
		"api_keys": {
			Name: "API Keys",
			KeyPatterns: []string{
				"apiKey", "api_key", "authKey", "auth_key", "token",
				"accessToken", "access_token", "secretKey", "secret_key",
			},
			ValuePatterns: []string{
				// API key patterns
				`[A-Za-z0-9]{20,}`,
				// JWT token patterns
				`eyJ[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+\.[A-Za-z0-9-_]*`,
			},
			Risk:        TelemetryRiskHigh,
			Description: "API keys or authentication tokens shared between extensions",
		},
		
		"usage_statistics": {
			Name: "Usage Statistics",
			KeyPatterns: []string{
				"usageCount", "usage_count", "activationCount", "activation_count",
				"commandCount", "command_count", "featureUsage", "feature_usage",
			},
			ValuePatterns: []string{
				// Numeric patterns
				`\d+`,
			},
			Risk:        TelemetryRiskMedium,
			Description: "Usage statistics data shared between extensions",
		},
		
		"performance_metrics": {
			Name: "Performance Metrics",
			KeyPatterns: []string{
				"performanceData", "performance_data", "metrics", "timing",
				"loadTime", "load_time", "responseTime", "response_time",
			},
			ValuePatterns: []string{
				// Numeric patterns with decimals
				`\d+\.?\d*`,
			},
			Risk:        TelemetryRiskMedium,
			Description: "Performance metrics shared between extensions",
		},
		
		"error_tracking": {
			Name: "Error Tracking",
			KeyPatterns: []string{
				"errorCount", "error_count", "crashCount", "crash_count",
				"errorLog", "error_log", "exception", "stackTrace",
			},
			ValuePatterns: []string{
				// Error message patterns
				`Error:.*`,
				`Exception:.*`,
			},
			Risk:        TelemetryRiskMedium,
			Description: "Error tracking data shared between extensions",
		},
	}
}

// initializeSharedDataTypes sets up known shared data types
func (ca *CorrelationAnalyzer) initializeSharedDataTypes() {
	ca.sharedDataTypes = map[string]SharedDataType{
		"vscode_machine_id": {
			Name:        "VS Code Machine ID",
			Risk:        TelemetryRiskCritical,
			Description: "VS Code's unique machine identifier",
			Examples:    []string{"vscode.env.machineId", "machineId"},
		},
		
		"vscode_session_id": {
			Name:        "VS Code Session ID",
			Risk:        TelemetryRiskHigh,
			Description: "VS Code's session identifier",
			Examples:    []string{"vscode.env.sessionId", "sessionId"},
		},
		
		"extension_host_id": {
			Name:        "Extension Host ID",
			Risk:        TelemetryRiskHigh,
			Description: "Extension host process identifier",
			Examples:    []string{"extensionHostId", "hostId"},
		},
		
		"workspace_hash": {
			Name:        "Workspace Hash",
			Risk:        TelemetryRiskMedium,
			Description: "Workspace folder hash identifier",
			Examples:    []string{"workspaceHash", "workspace_hash"},
		},
		
		"user_preferences": {
			Name:        "User Preferences",
			Risk:        TelemetryRiskLow,
			Description: "Shared user preference data",
			Examples:    []string{"preferences", "settings", "config"},
		},
	}
}

// AnalyzeCrossExtensionData analyzes data correlations between extensions
func (ca *CorrelationAnalyzer) AnalyzeCrossExtensionData(globalStorages []ExtensionStorage, workspaceStorages []WorkspaceStorage) []CrossExtensionData {
	var crossExtensionData []CrossExtensionData
	
	// Collect all storage items from all extensions
	allStorageItems := ca.collectAllStorageItems(globalStorages, workspaceStorages)
	
	// Analyze correlations by key patterns
	keyCorrelations := ca.analyzeKeyCorrelations(allStorageItems)
	crossExtensionData = append(crossExtensionData, keyCorrelations...)
	
	// Analyze correlations by value patterns
	valueCorrelations := ca.analyzeValueCorrelations(allStorageItems)
	crossExtensionData = append(crossExtensionData, valueCorrelations...)
	
	// Analyze shared data types
	sharedDataCorrelations := ca.analyzeSharedDataTypes(allStorageItems)
	crossExtensionData = append(crossExtensionData, sharedDataCorrelations...)
	
	return crossExtensionData
}

// collectAllStorageItems collects storage items from all extensions
func (ca *CorrelationAnalyzer) collectAllStorageItems(globalStorages []ExtensionStorage, workspaceStorages []WorkspaceStorage) map[string][]ExtensionStorageItem {
	allItems := make(map[string][]ExtensionStorageItem)
	
	// Collect from global storage
	for _, storage := range globalStorages {
		for _, item := range storage.StorageItems {
			storageItem := ExtensionStorageItem{
				ExtensionID:  storage.ExtensionID,
				StorageType:  "global",
				StorageItem:  item,
			}
			allItems[storage.ExtensionID] = append(allItems[storage.ExtensionID], storageItem)
		}
	}
	
	// Collect from workspace storage
	for _, workspace := range workspaceStorages {
		for _, storage := range workspace.ExtensionStorages {
			for _, item := range storage.StorageItems {
				storageItem := ExtensionStorageItem{
					ExtensionID:   storage.ExtensionID,
					StorageType:   "workspace",
					WorkspaceHash: workspace.WorkspaceHash,
					StorageItem:   item,
				}
				allItems[storage.ExtensionID] = append(allItems[storage.ExtensionID], storageItem)
			}
		}
	}
	
	return allItems
}

// ExtensionStorageItem represents a storage item with extension context
type ExtensionStorageItem struct {
	ExtensionID   string          `json:"extension_id"`
	StorageType   string          `json:"storage_type"`
	WorkspaceHash string          `json:"workspace_hash,omitempty"`
	StorageItem   StorageDataItem `json:"storage_item"`
}

// analyzeKeyCorrelations analyzes correlations based on key patterns
func (ca *CorrelationAnalyzer) analyzeKeyCorrelations(allItems map[string][]ExtensionStorageItem) []CrossExtensionData {
	var correlations []CrossExtensionData
	
	// Group items by key patterns
	keyGroups := make(map[string]map[string][]ExtensionStorageItem)
	
	for extensionID, items := range allItems {
		for _, item := range items {
			for patternName, pattern := range ca.correlationPatterns {
				for _, keyPattern := range pattern.KeyPatterns {
					if ca.matchesKeyPattern(item.StorageItem.Key, keyPattern) {
						if keyGroups[patternName] == nil {
							keyGroups[patternName] = make(map[string][]ExtensionStorageItem)
						}
						keyGroups[patternName][extensionID] = append(keyGroups[patternName][extensionID], item)
					}
				}
			}
		}
	}
	
	// Create correlations for patterns found in multiple extensions
	for patternName, extensionGroups := range keyGroups {
		if len(extensionGroups) > 1 { // Found in multiple extensions
			pattern := ca.correlationPatterns[patternName]
			
			var extensionIDs []string
			var sharedKeys []string
			var totalSize int64
			
			for extensionID, items := range extensionGroups {
				extensionIDs = append(extensionIDs, extensionID)
				for _, item := range items {
					sharedKeys = append(sharedKeys, item.StorageItem.Key)
					totalSize += item.StorageItem.Size
				}
			}
			
			correlation := CrossExtensionData{
				DataType:        pattern.Name,
				ExtensionIDs:    extensionIDs,
				SharedKeys:      ca.uniqueStrings(sharedKeys),
				Risk:            pattern.Risk,
				Description:     fmt.Sprintf("%s found in %d extensions", pattern.Description, len(extensionIDs)),
				DataSize:        totalSize,
				CorrelationHash: ca.generateCorrelationHash(patternName, extensionIDs),
			}
			
			correlations = append(correlations, correlation)
		}
	}
	
	return correlations
}

// analyzeValueCorrelations analyzes correlations based on value patterns
func (ca *CorrelationAnalyzer) analyzeValueCorrelations(allItems map[string][]ExtensionStorageItem) []CrossExtensionData {
	var correlations []CrossExtensionData
	
	// Group items by value hashes
	valueGroups := make(map[string][]ExtensionStorageItem)
	
	for _, items := range allItems {
		for _, item := range items {
			valueHash := ca.hashValue(item.StorageItem.Value)
			if valueHash != "" {
				valueGroups[valueHash] = append(valueGroups[valueHash], item)
			}
		}
	}
	
	// Create correlations for values found in multiple extensions
	for valueHash, items := range valueGroups {
		if len(items) > 1 {
			// Check if items are from different extensions
			extensionSet := make(map[string]bool)
			for _, item := range items {
				extensionSet[item.ExtensionID] = true
			}
			
			if len(extensionSet) > 1 { // Found in multiple extensions
				var extensionIDs []string
				var sharedKeys []string
				var totalSize int64
				var maxRisk TelemetryRisk
				
				for extensionID := range extensionSet {
					extensionIDs = append(extensionIDs, extensionID)
				}
				
				for _, item := range items {
					sharedKeys = append(sharedKeys, item.StorageItem.Key)
					totalSize += item.StorageItem.Size
					if item.StorageItem.Risk > maxRisk {
						maxRisk = item.StorageItem.Risk
					}
				}
				
				correlation := CrossExtensionData{
					DataType:        "Shared Value",
					ExtensionIDs:    extensionIDs,
					SharedKeys:      ca.uniqueStrings(sharedKeys),
					Risk:            maxRisk,
					Description:     fmt.Sprintf("Identical value found in %d extensions", len(extensionIDs)),
					DataSize:        totalSize,
					CorrelationHash: valueHash,
				}
				
				correlations = append(correlations, correlation)
			}
		}
	}
	
	return correlations
}

// analyzeSharedDataTypes analyzes known shared data types
func (ca *CorrelationAnalyzer) analyzeSharedDataTypes(allItems map[string][]ExtensionStorageItem) []CrossExtensionData {
	var correlations []CrossExtensionData
	
	for dataTypeName, dataType := range ca.sharedDataTypes {
		extensionMatches := make(map[string][]ExtensionStorageItem)
		
		// Find extensions that have this data type
		for extensionID, items := range allItems {
			for _, item := range items {
				if ca.matchesSharedDataType(item.StorageItem, dataType) {
					extensionMatches[extensionID] = append(extensionMatches[extensionID], item)
				}
			}
		}
		
		if len(extensionMatches) > 1 { // Found in multiple extensions
			var extensionIDs []string
			var sharedKeys []string
			var totalSize int64
			
			for extensionID, items := range extensionMatches {
				extensionIDs = append(extensionIDs, extensionID)
				for _, item := range items {
					sharedKeys = append(sharedKeys, item.StorageItem.Key)
					totalSize += item.StorageItem.Size
				}
			}
			
			correlation := CrossExtensionData{
				DataType:        dataType.Name,
				ExtensionIDs:    extensionIDs,
				SharedKeys:      ca.uniqueStrings(sharedKeys),
				Risk:            dataType.Risk,
				Description:     fmt.Sprintf("%s found in %d extensions", dataType.Description, len(extensionIDs)),
				DataSize:        totalSize,
				CorrelationHash: ca.generateCorrelationHash(dataTypeName, extensionIDs),
			}
			
			correlations = append(correlations, correlation)
		}
	}
	
	return correlations
}

// matchesKeyPattern checks if a key matches a pattern
func (ca *CorrelationAnalyzer) matchesKeyPattern(key, pattern string) bool {
	lowerKey := strings.ToLower(key)
	lowerPattern := strings.ToLower(pattern)
	
	return strings.Contains(lowerKey, lowerPattern)
}

// matchesSharedDataType checks if a storage item matches a shared data type
func (ca *CorrelationAnalyzer) matchesSharedDataType(item StorageDataItem, dataType SharedDataType) bool {
	lowerKey := strings.ToLower(item.Key)
	
	for _, example := range dataType.Examples {
		if strings.Contains(lowerKey, strings.ToLower(example)) {
			return true
		}
	}
	
	return false
}

// hashValue creates a hash of a value for comparison
func (ca *CorrelationAnalyzer) hashValue(value interface{}) string {
	if value == nil {
		return ""
	}
	
	valueStr := fmt.Sprintf("%v", value)
	
	// Only hash non-trivial values
	if len(valueStr) < 3 || valueStr == "true" || valueStr == "false" || valueStr == "null" {
		return ""
	}
	
	// Don't hash very long values (likely to be unique)
	if len(valueStr) > 1000 {
		return ""
	}
	
	hash := md5.Sum([]byte(valueStr))
	return fmt.Sprintf("%x", hash)
}

// generateCorrelationHash generates a hash for a correlation
func (ca *CorrelationAnalyzer) generateCorrelationHash(dataType string, extensionIDs []string) string {
	combined := dataType + ":" + strings.Join(extensionIDs, ",")
	hash := md5.Sum([]byte(combined))
	return fmt.Sprintf("%x", hash)
}

// uniqueStrings returns unique strings from a slice
func (ca *CorrelationAnalyzer) uniqueStrings(strings []string) []string {
	seen := make(map[string]bool)
	var unique []string
	
	for _, str := range strings {
		if !seen[str] {
			seen[str] = true
			unique = append(unique, str)
		}
	}
	
	return unique
}

// GetCorrelationStatistics returns statistics about data correlations
func (ca *CorrelationAnalyzer) GetCorrelationStatistics(correlations []CrossExtensionData) CorrelationStatistics {
	stats := CorrelationStatistics{
		TotalCorrelations: len(correlations),
	}
	
	// Count by risk level
	for _, correlation := range correlations {
		switch correlation.Risk {
		case TelemetryRiskCritical:
			stats.CriticalRiskCorrelations++
		case TelemetryRiskHigh:
			stats.HighRiskCorrelations++
		case TelemetryRiskMedium:
			stats.MediumRiskCorrelations++
		case TelemetryRiskLow:
			stats.LowRiskCorrelations++
		}
		
		stats.TotalDataSize += correlation.DataSize
		
		// Track unique extensions involved
		for _, extensionID := range correlation.ExtensionIDs {
			stats.AffectedExtensions[extensionID] = true
		}
	}
	
	stats.AffectedExtensionCount = len(stats.AffectedExtensions)
	
	// Find most common correlation types
	typeCount := make(map[string]int)
	for _, correlation := range correlations {
		typeCount[correlation.DataType]++
	}
	
	for dataType, count := range typeCount {
		stats.CommonCorrelationTypes = append(stats.CommonCorrelationTypes, CorrelationTypeCount{
			Type:  dataType,
			Count: count,
		})
	}
	
	return stats
}

// CorrelationStatistics represents statistics about data correlations
type CorrelationStatistics struct {
	TotalCorrelations         int                     `json:"total_correlations"`
	CriticalRiskCorrelations  int                     `json:"critical_risk_correlations"`
	HighRiskCorrelations      int                     `json:"high_risk_correlations"`
	MediumRiskCorrelations    int                     `json:"medium_risk_correlations"`
	LowRiskCorrelations       int                     `json:"low_risk_correlations"`
	TotalDataSize             int64                   `json:"total_data_size"`
	AffectedExtensionCount    int                     `json:"affected_extension_count"`
	AffectedExtensions        map[string]bool         `json:"affected_extensions"`
	CommonCorrelationTypes    []CorrelationTypeCount  `json:"common_correlation_types"`
}

// CorrelationTypeCount represents a count of correlation types
type CorrelationTypeCount struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}