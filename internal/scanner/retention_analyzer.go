package scanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// RetentionAnalyzer analyzes data retention policies for extensions
type RetentionAnalyzer struct {
	defaultRetentionPeriods map[string]time.Duration
	policyPatterns          map[string]RetentionPolicyType
}

// RetentionPolicyType represents different types of retention policies
type RetentionPolicyType int

const (
	RetentionPolicyNone RetentionPolicyType = iota
	RetentionPolicySession
	RetentionPolicyDaily
	RetentionPolicyWeekly
	RetentionPolicyMonthly
	RetentionPolicyPermanent
	RetentionPolicyCustom
)

// String returns the string representation of retention policy type
func (rpt RetentionPolicyType) String() string {
	switch rpt {
	case RetentionPolicyNone:
		return "None"
	case RetentionPolicySession:
		return "Session"
	case RetentionPolicyDaily:
		return "Daily"
	case RetentionPolicyWeekly:
		return "Weekly"
	case RetentionPolicyMonthly:
		return "Monthly"
	case RetentionPolicyPermanent:
		return "Permanent"
	case RetentionPolicyCustom:
		return "Custom"
	default:
		return "Unknown"
	}
}

// RetentionPolicyInfo represents detailed retention policy information
type RetentionPolicyInfo struct {
	Type            RetentionPolicyType `json:"type"`
	Period          time.Duration       `json:"period"`
	AutoCleanup     bool                `json:"auto_cleanup"`
	LastCleanup     time.Time           `json:"last_cleanup"`
	NextCleanup     time.Time           `json:"next_cleanup"`
	PolicySource    string              `json:"policy_source"`
	ConfigPath      string              `json:"config_path,omitempty"`
	IsEnforced      bool                `json:"is_enforced"`
	CleanupRules    []CleanupRule       `json:"cleanup_rules"`
}

// CleanupRule represents a specific cleanup rule
type CleanupRule struct {
	Name        string        `json:"name"`
	Pattern     string        `json:"pattern"`
	MaxAge      time.Duration `json:"max_age"`
	MaxSize     int64         `json:"max_size"`
	Priority    int           `json:"priority"`
	Enabled     bool          `json:"enabled"`
	Description string        `json:"description"`
}

// NewRetentionAnalyzer creates a new retention analyzer
func NewRetentionAnalyzer() *RetentionAnalyzer {
	analyzer := &RetentionAnalyzer{}
	analyzer.initializeDefaultRetentionPeriods()
	analyzer.initializePolicyPatterns()
	return analyzer
}

// initializeDefaultRetentionPeriods sets up default retention periods for different data types
func (ra *RetentionAnalyzer) initializeDefaultRetentionPeriods() {
	ra.defaultRetentionPeriods = map[string]time.Duration{
		// Telemetry data - typically short retention
		"telemetry":     7 * 24 * time.Hour,  // 1 week
		"analytics":     30 * 24 * time.Hour, // 1 month
		"tracking":      7 * 24 * time.Hour,  // 1 week
		
		// Usage data - medium retention
		"usage":         90 * 24 * time.Hour, // 3 months
		"metrics":       30 * 24 * time.Hour, // 1 month
		"performance":   14 * 24 * time.Hour, // 2 weeks
		
		// Error and diagnostic data - longer retention
		"error":         30 * 24 * time.Hour, // 1 month
		"crash":         90 * 24 * time.Hour, // 3 months
		"diagnostic":    30 * 24 * time.Hour, // 1 month
		
		// Cache data - short retention
		"cache":         7 * 24 * time.Hour,  // 1 week
		"temp":          1 * 24 * time.Hour,  // 1 day
		"log":           14 * 24 * time.Hour, // 2 weeks
		
		// User data - longer retention
		"preferences":   365 * 24 * time.Hour, // 1 year
		"settings":      365 * 24 * time.Hour, // 1 year
		"history":       90 * 24 * time.Hour,  // 3 months
		
		// Session data - very short retention
		"session":       1 * time.Hour,        // 1 hour
		"auth":          24 * time.Hour,       // 1 day
		"token":         24 * time.Hour,       // 1 day
	}
}

// initializePolicyPatterns sets up patterns for detecting retention policies
func (ra *RetentionAnalyzer) initializePolicyPatterns() {
	ra.policyPatterns = map[string]RetentionPolicyType{
		"session":    RetentionPolicySession,
		"daily":      RetentionPolicyDaily,
		"weekly":     RetentionPolicyWeekly,
		"monthly":    RetentionPolicyMonthly,
		"permanent":  RetentionPolicyPermanent,
		"never":      RetentionPolicyPermanent,
		"cleanup":    RetentionPolicyCustom,
		"retention":  RetentionPolicyCustom,
		"expire":     RetentionPolicyCustom,
		"ttl":        RetentionPolicyCustom,
	}
}

// AnalyzeRetentionPolicy analyzes the retention policy for an extension
func (ra *RetentionAnalyzer) AnalyzeRetentionPolicy(extensionID, storagePath string) RetentionPolicy {
	policy := RetentionPolicy{
		HasPolicy:    false,
		AutoCleanup:  false,
		PolicySource: "default",
	}

	// Look for explicit retention policy configuration
	policyInfo := ra.findExplicitPolicy(extensionID, storagePath)
	if policyInfo != nil {
		policy.HasPolicy = true
		policy.RetentionPeriod = policyInfo.Period
		policy.AutoCleanup = policyInfo.AutoCleanup
		policy.LastCleanup = policyInfo.LastCleanup
		policy.PolicySource = policyInfo.PolicySource
		return policy
	}

	// Infer policy from data patterns
	inferredPolicy := ra.inferPolicyFromData(storagePath)
	if inferredPolicy != nil {
		policy.HasPolicy = true
		policy.RetentionPeriod = inferredPolicy.Period
		policy.AutoCleanup = inferredPolicy.AutoCleanup
		policy.PolicySource = "inferred"
		return policy
	}

	// Use default policy based on extension type
	defaultPeriod := ra.getDefaultRetentionPeriod(extensionID, storagePath)
	if defaultPeriod > 0 {
		policy.RetentionPeriod = defaultPeriod
		policy.PolicySource = "default"
	}

	return policy
}

// findExplicitPolicy looks for explicit retention policy configuration
func (ra *RetentionAnalyzer) findExplicitPolicy(extensionID, storagePath string) *RetentionPolicyInfo {
	// Check for retention policy files
	policyFiles := []string{
		"retention.json",
		"cleanup.json",
		"policy.json",
		"config.json",
		"settings.json",
	}

	for _, fileName := range policyFiles {
		policyPath := filepath.Join(storagePath, fileName)
		if policy := ra.parseRetentionPolicyFile(policyPath); policy != nil {
			return policy
		}
	}

	// Check for retention settings in main storage files
	return ra.findRetentionInStorageFiles(storagePath)
}

// parseRetentionPolicyFile parses a retention policy configuration file
func (ra *RetentionAnalyzer) parseRetentionPolicyFile(filePath string) *RetentionPolicyInfo {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil
	}

	return ra.extractRetentionPolicyFromConfig(config, filePath)
}

// extractRetentionPolicyFromConfig extracts retention policy from configuration
func (ra *RetentionAnalyzer) extractRetentionPolicyFromConfig(config map[string]interface{}, configPath string) *RetentionPolicyInfo {
	policy := &RetentionPolicyInfo{
		Type:         RetentionPolicyNone,
		PolicySource: configPath,
		ConfigPath:   configPath,
		CleanupRules: make([]CleanupRule, 0),
	}

	// Look for retention-related keys
	for key, value := range config {
		lowerKey := strings.ToLower(key)
		
		// Check for retention period
		if strings.Contains(lowerKey, "retention") || strings.Contains(lowerKey, "ttl") {
			if period := ra.parseRetentionPeriod(value); period > 0 {
				policy.Period = period
				policy.Type = RetentionPolicyCustom
			}
		}
		
		// Check for cleanup settings
		if strings.Contains(lowerKey, "cleanup") || strings.Contains(lowerKey, "clean") {
			if cleanupConfig, ok := value.(map[string]interface{}); ok {
				policy.AutoCleanup = ra.parseCleanupConfig(cleanupConfig, policy)
			} else if enabled, ok := value.(bool); ok {
				policy.AutoCleanup = enabled
			}
		}
		
		// Check for last cleanup timestamp
		if strings.Contains(lowerKey, "lastcleanup") || strings.Contains(lowerKey, "last_cleanup") {
			if timestamp := ra.parseTimestamp(value); !timestamp.IsZero() {
				policy.LastCleanup = timestamp
			}
		}
		
		// Check for policy type
		if strings.Contains(lowerKey, "policy") || strings.Contains(lowerKey, "type") {
			if policyType := ra.parsePolicyType(value); policyType != RetentionPolicyNone {
				policy.Type = policyType
			}
		}
	}

	if policy.Type != RetentionPolicyNone || policy.Period > 0 || policy.AutoCleanup {
		return policy
	}

	return nil
}

// parseRetentionPeriod parses a retention period from various formats
func (ra *RetentionAnalyzer) parseRetentionPeriod(value interface{}) time.Duration {
	switch v := value.(type) {
	case string:
		// Try to parse as duration string
		if duration, err := time.ParseDuration(v); err == nil {
			return duration
		}
		
		// Try to parse common formats
		lowerValue := strings.ToLower(v)
		if strings.Contains(lowerValue, "day") {
			return 24 * time.Hour
		}
		if strings.Contains(lowerValue, "week") {
			return 7 * 24 * time.Hour
		}
		if strings.Contains(lowerValue, "month") {
			return 30 * 24 * time.Hour
		}
		if strings.Contains(lowerValue, "year") {
			return 365 * 24 * time.Hour
		}
		
	case float64:
		// Assume value is in hours
		return time.Duration(v) * time.Hour
		
	case int:
		// Assume value is in hours
		return time.Duration(v) * time.Hour
	}
	
	return 0
}

// parseCleanupConfig parses cleanup configuration
func (ra *RetentionAnalyzer) parseCleanupConfig(config map[string]interface{}, policy *RetentionPolicyInfo) bool {
	autoCleanup := false
	
	for key, value := range config {
		lowerKey := strings.ToLower(key)
		
		if strings.Contains(lowerKey, "enabled") || strings.Contains(lowerKey, "auto") {
			if enabled, ok := value.(bool); ok {
				autoCleanup = enabled
			}
		}
		
		if strings.Contains(lowerKey, "rule") {
			if rules, ok := value.([]interface{}); ok {
				for _, ruleData := range rules {
					if ruleMap, ok := ruleData.(map[string]interface{}); ok {
						rule := ra.parseCleanupRule(ruleMap)
						if rule != nil {
							policy.CleanupRules = append(policy.CleanupRules, *rule)
						}
					}
				}
			}
		}
		
		if strings.Contains(lowerKey, "maxage") || strings.Contains(lowerKey, "max_age") {
			if period := ra.parseRetentionPeriod(value); period > 0 {
				rule := CleanupRule{
					Name:        "MaxAge",
					MaxAge:      period,
					Priority:    1,
					Enabled:     true,
					Description: "Maximum age cleanup rule",
				}
				policy.CleanupRules = append(policy.CleanupRules, rule)
			}
		}
	}
	
	return autoCleanup
}

// parseCleanupRule parses a single cleanup rule
func (ra *RetentionAnalyzer) parseCleanupRule(ruleConfig map[string]interface{}) *CleanupRule {
	rule := &CleanupRule{
		Priority: 1,
		Enabled:  true,
	}
	
	for key, value := range ruleConfig {
		lowerKey := strings.ToLower(key)
		
		switch lowerKey {
		case "name":
			if name, ok := value.(string); ok {
				rule.Name = name
			}
		case "pattern":
			if pattern, ok := value.(string); ok {
				rule.Pattern = pattern
			}
		case "maxage", "max_age":
			rule.MaxAge = ra.parseRetentionPeriod(value)
		case "maxsize", "max_size":
			if size, ok := value.(float64); ok {
				rule.MaxSize = int64(size)
			}
		case "priority":
			if priority, ok := value.(float64); ok {
				rule.Priority = int(priority)
			}
		case "enabled":
			if enabled, ok := value.(bool); ok {
				rule.Enabled = enabled
			}
		case "description":
			if desc, ok := value.(string); ok {
				rule.Description = desc
			}
		}
	}
	
	if rule.Name != "" || rule.Pattern != "" || rule.MaxAge > 0 || rule.MaxSize > 0 {
		return rule
	}
	
	return nil
}

// parseTimestamp parses a timestamp from various formats
func (ra *RetentionAnalyzer) parseTimestamp(value interface{}) time.Time {
	switch v := value.(type) {
	case string:
		// Try common timestamp formats
		formats := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02T15:04:05Z",
			"2006-01-02 15:04:05",
			"2006-01-02",
		}
		
		for _, format := range formats {
			if t, err := time.Parse(format, v); err == nil {
				return t
			}
		}
		
	case float64:
		// Assume Unix timestamp
		return time.Unix(int64(v), 0)
		
	case int64:
		// Unix timestamp
		return time.Unix(v, 0)
	}
	
	return time.Time{}
}

// parsePolicyType parses a policy type from string
func (ra *RetentionAnalyzer) parsePolicyType(value interface{}) RetentionPolicyType {
	if str, ok := value.(string); ok {
		lowerStr := strings.ToLower(str)
		
		for pattern, policyType := range ra.policyPatterns {
			if strings.Contains(lowerStr, pattern) {
				return policyType
			}
		}
	}
	
	return RetentionPolicyNone
}

// findRetentionInStorageFiles looks for retention settings in storage files
func (ra *RetentionAnalyzer) findRetentionInStorageFiles(storagePath string) *RetentionPolicyInfo {
	var policy *RetentionPolicyInfo
	
	err := filepath.Walk(storagePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue despite errors
		}
		
		if info.IsDir() || !strings.HasSuffix(strings.ToLower(info.Name()), ".json") {
			return nil
		}
		
		// Skip very large files
		if info.Size() > 1024*1024 { // 1MB limit
			return nil
		}
		
		if foundPolicy := ra.parseRetentionPolicyFile(path); foundPolicy != nil {
			policy = foundPolicy
			return filepath.SkipDir // Stop searching once we find a policy
		}
		
		return nil
	})
	
	if err != nil {
		return nil
	}
	
	return policy
}

// inferPolicyFromData infers retention policy from data patterns
func (ra *RetentionAnalyzer) inferPolicyFromData(storagePath string) *RetentionPolicyInfo {
	policy := &RetentionPolicyInfo{
		Type:         RetentionPolicyNone,
		PolicySource: "inferred",
	}
	
	var oldestFile, newestFile time.Time
	var totalFiles int
	var hasCleanupPattern bool
	
	err := filepath.Walk(storagePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		if info.IsDir() {
			return nil
		}
		
		totalFiles++
		modTime := info.ModTime()
		
		if oldestFile.IsZero() || modTime.Before(oldestFile) {
			oldestFile = modTime
		}
		if newestFile.IsZero() || modTime.After(newestFile) {
			newestFile = modTime
		}
		
		// Check for cleanup patterns in file names
		fileName := strings.ToLower(info.Name())
		if strings.Contains(fileName, "cleanup") || 
		   strings.Contains(fileName, "clean") ||
		   strings.Contains(fileName, "temp") ||
		   strings.Contains(fileName, "tmp") {
			hasCleanupPattern = true
		}
		
		return nil
	})
	
	if err != nil {
		return nil
	}
	
	if totalFiles == 0 {
		return nil
	}
	
	// Infer policy based on file age distribution
	if !oldestFile.IsZero() && !newestFile.IsZero() {
		dataSpan := newestFile.Sub(oldestFile)
		
		if dataSpan < 24*time.Hour {
			policy.Type = RetentionPolicySession
			policy.Period = 24 * time.Hour
		} else if dataSpan < 7*24*time.Hour {
			policy.Type = RetentionPolicyDaily
			policy.Period = 7 * 24 * time.Hour
		} else if dataSpan < 30*24*time.Hour {
			policy.Type = RetentionPolicyWeekly
			policy.Period = 30 * 24 * time.Hour
		} else if dataSpan < 365*24*time.Hour {
			policy.Type = RetentionPolicyMonthly
			policy.Period = 365 * 24 * time.Hour
		} else {
			policy.Type = RetentionPolicyPermanent
		}
		
		policy.AutoCleanup = hasCleanupPattern
		
		return policy
	}
	
	return nil
}

// getDefaultRetentionPeriod gets the default retention period for an extension
func (ra *RetentionAnalyzer) getDefaultRetentionPeriod(extensionID, storagePath string) time.Duration {
	// Analyze storage path and extension ID for data type hints
	lowerPath := strings.ToLower(storagePath)
	lowerID := strings.ToLower(extensionID)
	
	// Check against known patterns
	for pattern, period := range ra.defaultRetentionPeriods {
		if strings.Contains(lowerPath, pattern) || strings.Contains(lowerID, pattern) {
			return period
		}
	}
	
	// Default retention period
	return 30 * 24 * time.Hour // 1 month
}

// GetRetentionRecommendations provides recommendations for retention policies
func (ra *RetentionAnalyzer) GetRetentionRecommendations(extensionStorage ExtensionStorage) []RetentionRecommendation {
	var recommendations []RetentionRecommendation
	
	// Analyze storage items for recommendations
	for _, item := range extensionStorage.StorageItems {
		if rec := ra.getItemRetentionRecommendation(item); rec != nil {
			recommendations = append(recommendations, *rec)
		}
	}
	
	// Add general recommendations based on storage size and age
	if extensionStorage.TotalSize > 100*1024*1024 { // > 100MB
		recommendations = append(recommendations, RetentionRecommendation{
			Type:        "storage_size",
			Priority:    "high",
			Description: "Large storage size detected - consider implementing cleanup policies",
			Action:      "Enable automatic cleanup for old data",
		})
	}
	
	if extensionStorage.Risk >= TelemetryRiskHigh {
		recommendations = append(recommendations, RetentionRecommendation{
			Type:        "privacy",
			Priority:    "critical",
			Description: "High-risk telemetry data detected - implement short retention period",
			Action:      "Set retention period to 7 days or less",
		})
	}
	
	return recommendations
}

// RetentionRecommendation represents a retention policy recommendation
type RetentionRecommendation struct {
	Type        string `json:"type"`
	Priority    string `json:"priority"`
	Description string `json:"description"`
	Action      string `json:"action"`
}

// getItemRetentionRecommendation gets retention recommendation for a storage item
func (ra *RetentionAnalyzer) getItemRetentionRecommendation(item StorageDataItem) *RetentionRecommendation {
	if item.Risk >= TelemetryRiskHigh {
		return &RetentionRecommendation{
			Type:        "high_risk_data",
			Priority:    "high",
			Description: fmt.Sprintf("High-risk data item: %s", item.Key),
			Action:      "Consider removing or implementing short retention period",
		}
	}
	
	// Check for old data
	if time.Since(item.LastModified) > 90*24*time.Hour { // > 3 months
		return &RetentionRecommendation{
			Type:        "old_data",
			Priority:    "medium",
			Description: fmt.Sprintf("Old data item: %s (age: %v)", item.Key, time.Since(item.LastModified)),
			Action:      "Consider removing old data",
		}
	}
	
	return nil
}