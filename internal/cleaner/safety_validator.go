package cleaner

import (
	"fmt"
	"os"
	"strings"
	"time"

	"augment-telemetry-cleaner/internal/scanner"
)

// SafetyValidator handles validation of removal operations for safety
type SafetyValidator struct {
	criticalPaths    []string
	protectedPatterns []string
	safetyRules      []SafetyRule
}

// SafetyRule represents a safety rule for data removal
type SafetyRule struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	RuleType    string                `json:"rule_type"`
	Pattern     string                `json:"pattern"`
	Action      string                `json:"action"`
	Severity    string                `json:"severity"`
	Enabled     bool                  `json:"enabled"`
}

// SafetyValidationResult represents the result of safety validation
type SafetyValidationResult struct {
	Safe            bool          `json:"safe"`
	Warnings        []SafetyIssue `json:"warnings"`
	Errors          []SafetyIssue `json:"errors"`
	Recommendations []string      `json:"recommendations"`
	RiskScore       float64       `json:"risk_score"`
}

// SafetyIssue represents a safety issue found during validation
type SafetyIssue struct {
	Type        string                `json:"type"`
	Severity    string                `json:"severity"`
	Message     string                `json:"message"`
	Path        string                `json:"path,omitempty"`
	Rule        string                `json:"rule,omitempty"`
	Risk        scanner.TelemetryRisk `json:"risk,omitempty"`
	Suggestion  string                `json:"suggestion,omitempty"`
}

// NewSafetyValidator creates a new safety validator
func NewSafetyValidator() *SafetyValidator {
	validator := &SafetyValidator{}
	validator.initializeCriticalPaths()
	validator.initializeProtectedPatterns()
	validator.initializeSafetyRules()
	return validator
}

// initializeCriticalPaths sets up critical paths that should be protected
func (sv *SafetyValidator) initializeCriticalPaths() {
	sv.criticalPaths = []string{
		// VS Code core paths
		"settings.json",
		"keybindings.json",
		"tasks.json",
		"launch.json",
		
		// Extension manifest files
		"package.json",
		"extension.js",
		"main.js",
		
		// User data
		"user-data",
		"profiles",
		"workspaces",
		
		// System paths
		"system32",
		"program files",
		"applications",
	}
}

// initializeProtectedPatterns sets up patterns for protected data
func (sv *SafetyValidator) initializeProtectedPatterns() {
	sv.protectedPatterns = []string{
		// User configuration
		"config",
		"settings",
		"preferences",
		"profile",
		
		// Important user data
		"workspace",
		"project",
		"bookmark",
		"history",
		
		// Authentication data
		"auth",
		"token",
		"credential",
		"certificate",
		
		// Extension core files
		"manifest",
		"package",
		"main",
		"index",
	}
}

// initializeSafetyRules sets up safety rules for validation
func (sv *SafetyValidator) initializeSafetyRules() {
	sv.safetyRules = []SafetyRule{
		{
			Name:        "protect_user_settings",
			Description: "Protect user settings and configuration files",
			RuleType:    "path_protection",
			Pattern:     "*settings*",
			Action:      "warn",
			Severity:    "high",
			Enabled:     true,
		},
		{
			Name:        "protect_authentication",
			Description: "Protect authentication and credential data",
			RuleType:    "content_protection",
			Pattern:     "*auth*|*token*|*credential*",
			Action:      "block",
			Severity:    "critical",
			Enabled:     true,
		},
		{
			Name:        "protect_workspace_data",
			Description: "Protect workspace and project data",
			RuleType:    "path_protection",
			Pattern:     "*workspace*|*project*",
			Action:      "warn",
			Severity:    "medium",
			Enabled:     true,
		},
		{
			Name:        "protect_recent_data",
			Description: "Protect recently modified data",
			RuleType:    "temporal_protection",
			Pattern:     "age < 24h",
			Action:      "warn",
			Severity:    "medium",
			Enabled:     true,
		},
		{
			Name:        "protect_large_data",
			Description: "Warn about removing large amounts of data",
			RuleType:    "size_protection",
			Pattern:     "size > 100MB",
			Action:      "warn",
			Severity:    "medium",
			Enabled:     true,
		},
		{
			Name:        "protect_system_paths",
			Description: "Block removal from system paths",
			RuleType:    "path_protection",
			Pattern:     "*system*|*program files*|*applications*",
			Action:      "block",
			Severity:    "critical",
			Enabled:     true,
		},
	}
}

// ValidateRemovalSafety validates the safety of removing specific data
func (sv *SafetyValidator) ValidateRemovalSafety(items []scanner.StorageDataItem, extensionPath string) (*SafetyValidationResult, error) {
	result := &SafetyValidationResult{
		Safe:            true,
		Warnings:        make([]SafetyIssue, 0),
		Errors:          make([]SafetyIssue, 0),
		Recommendations: make([]string, 0),
	}

	var totalSize int64
	var criticalItems int
	var recentItems int

	// Validate each item
	for _, item := range items {
		totalSize += item.Size
		
		// Check against safety rules
		issues := sv.validateItem(item, extensionPath)
		for _, issue := range issues {
			if issue.Severity == "critical" || issue.Severity == "high" {
				result.Errors = append(result.Errors, issue)
				result.Safe = false
			} else {
				result.Warnings = append(result.Warnings, issue)
			}
		}

		// Count critical items
		if item.Risk == scanner.TelemetryRiskCritical {
			criticalItems++
		}

		// Count recent items
		if time.Since(item.LastModified) < 24*time.Hour {
			recentItems++
		}
	}

	// Calculate risk score
	result.RiskScore = sv.calculateRiskScore(items, totalSize, criticalItems, recentItems)

	// Generate recommendations
	result.Recommendations = sv.generateRecommendations(result, totalSize, criticalItems, recentItems)

	// Final safety check
	if result.RiskScore > 0.8 {
		result.Safe = false
		result.Errors = append(result.Errors, SafetyIssue{
			Type:     "risk_assessment",
			Severity: "critical",
			Message:  fmt.Sprintf("Overall risk score too high: %.2f", result.RiskScore),
			Suggestion: "Consider using a more conservative removal policy",
		})
	}

	return result, nil
}

// validateItem validates a single storage item against safety rules
func (sv *SafetyValidator) validateItem(item scanner.StorageDataItem, extensionPath string) []SafetyIssue {
	var issues []SafetyIssue

	for _, rule := range sv.safetyRules {
		if !rule.Enabled {
			continue
		}

		violation := false
		var message string

		switch rule.RuleType {
		case "path_protection":
			if sv.matchesPathPattern(item.Key, rule.Pattern) {
				violation = true
				message = fmt.Sprintf("Item matches protected path pattern: %s", rule.Pattern)
			}

		case "content_protection":
			if sv.matchesContentPattern(item, rule.Pattern) {
				violation = true
				message = fmt.Sprintf("Item contains protected content: %s", rule.Pattern)
			}

		case "temporal_protection":
			if sv.matchesTemporalPattern(item, rule.Pattern) {
				violation = true
				message = fmt.Sprintf("Item matches temporal protection rule: %s", rule.Pattern)
			}

		case "size_protection":
			if sv.matchesSizePattern(item, rule.Pattern) {
				violation = true
				message = fmt.Sprintf("Item matches size protection rule: %s", rule.Pattern)
			}
		}

		if violation {
			issue := SafetyIssue{
				Type:     rule.RuleType,
				Severity: rule.Severity,
				Message:  message,
				Path:     item.Key,
				Rule:     rule.Name,
				Risk:     item.Risk,
				Suggestion: sv.getSuggestionForRule(rule),
			}
			issues = append(issues, issue)
		}
	}

	// Additional custom validations
	customIssues := sv.performCustomValidations(item, extensionPath)
	issues = append(issues, customIssues...)

	return issues
}

// matchesPathPattern checks if a path matches a protection pattern
func (sv *SafetyValidator) matchesPathPattern(path, pattern string) bool {
	lowerPath := strings.ToLower(path)
	
	// Handle multiple patterns separated by |
	patterns := strings.Split(pattern, "|")
	for _, p := range patterns {
		p = strings.TrimSpace(strings.ToLower(p))
		
		// Remove * wildcards for simple contains matching
		p = strings.ReplaceAll(p, "*", "")
		
		if strings.Contains(lowerPath, p) {
			return true
		}
	}
	
	return false
}

// matchesContentPattern checks if item content matches a protection pattern
func (sv *SafetyValidator) matchesContentPattern(item scanner.StorageDataItem, pattern string) bool {
	// Check item key and category
	content := strings.ToLower(item.Key + " " + item.Category + " " + item.Description)
	
	patterns := strings.Split(pattern, "|")
	for _, p := range patterns {
		p = strings.TrimSpace(strings.ToLower(p))
		p = strings.ReplaceAll(p, "*", "")
		
		if strings.Contains(content, p) {
			return true
		}
	}
	
	return false
}

// matchesTemporalPattern checks if item matches temporal protection rules
func (sv *SafetyValidator) matchesTemporalPattern(item scanner.StorageDataItem, pattern string) bool {
	if pattern == "age < 24h" {
		return time.Since(item.LastModified) < 24*time.Hour
	}
	if pattern == "age < 7d" {
		return time.Since(item.LastModified) < 7*24*time.Hour
	}
	if pattern == "age < 30d" {
		return time.Since(item.LastModified) < 30*24*time.Hour
	}
	
	return false
}

// matchesSizePattern checks if item matches size protection rules
func (sv *SafetyValidator) matchesSizePattern(item scanner.StorageDataItem, pattern string) bool {
	if pattern == "size > 100MB" {
		return item.Size > 100*1024*1024
	}
	if pattern == "size > 10MB" {
		return item.Size > 10*1024*1024
	}
	if pattern == "size > 1MB" {
		return item.Size > 1024*1024
	}
	
	return false
}

// performCustomValidations performs additional custom safety validations
func (sv *SafetyValidator) performCustomValidations(item scanner.StorageDataItem, extensionPath string) []SafetyIssue {
	var issues []SafetyIssue

	// Check for critical paths
	for _, criticalPath := range sv.criticalPaths {
		if strings.Contains(strings.ToLower(item.Key), strings.ToLower(criticalPath)) {
			issues = append(issues, SafetyIssue{
				Type:     "critical_path",
				Severity: "high",
				Message:  fmt.Sprintf("Item is in critical path: %s", criticalPath),
				Path:     item.Key,
				Risk:     item.Risk,
				Suggestion: "Consider excluding this item from removal",
			})
		}
	}

	// Check for protected patterns
	for _, pattern := range sv.protectedPatterns {
		if strings.Contains(strings.ToLower(item.Key), strings.ToLower(pattern)) {
			issues = append(issues, SafetyIssue{
				Type:     "protected_pattern",
				Severity: "medium",
				Message:  fmt.Sprintf("Item matches protected pattern: %s", pattern),
				Path:     item.Key,
				Risk:     item.Risk,
				Suggestion: "Verify this item should be removed",
			})
		}
	}

	// Check for high-risk items
	if item.Risk == scanner.TelemetryRiskCritical {
		issues = append(issues, SafetyIssue{
			Type:     "high_risk_data",
			Severity: "high",
			Message:  "Item contains critical telemetry data",
			Path:     item.Key,
			Risk:     item.Risk,
			Suggestion: "Ensure this critical data should be removed",
		})
	}

	// Check for very recent modifications
	if time.Since(item.LastModified) < 1*time.Hour {
		issues = append(issues, SafetyIssue{
			Type:     "recent_modification",
			Severity: "medium",
			Message:  "Item was modified very recently",
			Path:     item.Key,
			Risk:     item.Risk,
			Suggestion: "Consider waiting before removing recently modified data",
		})
	}

	return issues
}

// calculateRiskScore calculates an overall risk score for the removal operation
func (sv *SafetyValidator) calculateRiskScore(items []scanner.StorageDataItem, totalSize int64, criticalItems, recentItems int) float64 {
	score := 0.0
	totalItems := len(items)

	if totalItems == 0 {
		return 0.0
	}

	// Factor in critical items (0.0 - 0.4)
	criticalRatio := float64(criticalItems) / float64(totalItems)
	score += criticalRatio * 0.4

	// Factor in recent items (0.0 - 0.2)
	recentRatio := float64(recentItems) / float64(totalItems)
	score += recentRatio * 0.2

	// Factor in total size (0.0 - 0.2)
	if totalSize > 1024*1024*1024 { // > 1GB
		score += 0.2
	} else if totalSize > 100*1024*1024 { // > 100MB
		score += 0.1
	}

	// Factor in item count (0.0 - 0.2)
	if totalItems > 1000 {
		score += 0.2
	} else if totalItems > 100 {
		score += 0.1
	}

	return score
}

// generateRecommendations generates safety recommendations based on validation results
func (sv *SafetyValidator) generateRecommendations(result *SafetyValidationResult, totalSize int64, criticalItems, recentItems int) []string {
	var recommendations []string

	// Recommendations based on errors
	if len(result.Errors) > 0 {
		recommendations = append(recommendations, "Address all critical safety issues before proceeding")
		recommendations = append(recommendations, "Consider using a more conservative removal policy")
	}

	// Recommendations based on warnings
	if len(result.Warnings) > 0 {
		recommendations = append(recommendations, "Review all warnings carefully before proceeding")
		if len(result.Warnings) > 10 {
			recommendations = append(recommendations, "Consider filtering items to reduce the number of warnings")
		}
	}

	// Recommendations based on data characteristics
	if criticalItems > 0 {
		recommendations = append(recommendations, 
			fmt.Sprintf("Found %d critical telemetry items - ensure these should be removed", criticalItems))
	}

	if recentItems > 0 {
		recommendations = append(recommendations, 
			fmt.Sprintf("Found %d recently modified items - consider preserving recent data", recentItems))
	}

	if totalSize > 100*1024*1024 { // > 100MB
		recommendations = append(recommendations, 
			fmt.Sprintf("Large amount of data to remove (%.2f MB) - ensure adequate backup", 
				float64(totalSize)/(1024*1024)))
	}

	// General recommendations
	if result.RiskScore > 0.5 {
		recommendations = append(recommendations, "High risk operation - consider running in dry-run mode first")
		recommendations = append(recommendations, "Ensure comprehensive backups are created and verified")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Operation appears safe to proceed")
	}

	return recommendations
}

// getSuggestionForRule returns a suggestion for a specific safety rule
func (sv *SafetyValidator) getSuggestionForRule(rule SafetyRule) string {
	suggestions := map[string]string{
		"protect_user_settings":   "Consider excluding user settings from removal",
		"protect_authentication": "Never remove authentication data without explicit user consent",
		"protect_workspace_data":  "Verify workspace data should be removed",
		"protect_recent_data":     "Consider preserving recently modified data",
		"protect_large_data":      "Ensure adequate backup for large data removal",
		"protect_system_paths":    "System paths should never be modified",
	}

	if suggestion, exists := suggestions[rule.Name]; exists {
		return suggestion
	}

	return "Review this item carefully before removal"
}

// ValidateBackupIntegrity validates that a backup can be used for restoration
func (sv *SafetyValidator) ValidateBackupIntegrity(backupPath string) error {
	// Check if backup file exists
	if _, err := os.Stat(backupPath); err != nil {
		return fmt.Errorf("backup file not accessible: %w", err)
	}

	// Check if backup is not empty
	info, err := os.Stat(backupPath)
	if err != nil {
		return fmt.Errorf("cannot get backup file info: %w", err)
	}

	if info.Size() == 0 {
		return fmt.Errorf("backup file is empty")
	}

	// For zip files, try to open and verify structure
	if strings.HasSuffix(strings.ToLower(backupPath), ".zip") {
		return sv.validateZipBackup(backupPath)
	}

	return nil
}

// validateZipBackup validates a zip backup file
func (sv *SafetyValidator) validateZipBackup(zipPath string) error {
	// This would use the same logic as in backup_manager.go
	// For now, we'll do a basic check
	
	file, err := os.Open(zipPath)
	if err != nil {
		return fmt.Errorf("cannot open zip file: %w", err)
	}
	defer file.Close()

	// Read first few bytes to check zip signature
	header := make([]byte, 4)
	_, err = file.Read(header)
	if err != nil {
		return fmt.Errorf("cannot read zip header: %w", err)
	}

	// Check for zip file signature (PK)
	if header[0] != 0x50 || header[1] != 0x4B {
		return fmt.Errorf("invalid zip file signature")
	}

	return nil
}

// GetSafetyRules returns the current safety rules
func (sv *SafetyValidator) GetSafetyRules() []SafetyRule {
	return sv.safetyRules
}

// UpdateSafetyRule updates or adds a safety rule
func (sv *SafetyValidator) UpdateSafetyRule(rule SafetyRule) {
	for i, existingRule := range sv.safetyRules {
		if existingRule.Name == rule.Name {
			sv.safetyRules[i] = rule
			return
		}
	}
	
	// Add new rule if not found
	sv.safetyRules = append(sv.safetyRules, rule)
}

// DisableSafetyRule disables a specific safety rule
func (sv *SafetyValidator) DisableSafetyRule(ruleName string) {
	for i, rule := range sv.safetyRules {
		if rule.Name == ruleName {
			sv.safetyRules[i].Enabled = false
			return
		}
	}
}