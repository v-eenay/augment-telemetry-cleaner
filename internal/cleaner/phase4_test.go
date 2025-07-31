package cleaner

import (
	"os"
	"testing"
	"time"

	"augment-telemetry-cleaner/internal/scanner"
)

func TestNewExtensionCleaner(t *testing.T) {
	policy := GetDefaultRemovalPolicy()
	cleaner := NewExtensionCleaner(policy)
	
	if cleaner == nil {
		t.Fatal("NewExtensionCleaner() returned nil")
	}
	
	if cleaner.backupManager == nil {
		t.Error("Expected backup manager to be initialized")
	}
	
	if cleaner.dependencyChecker == nil {
		t.Error("Expected dependency checker to be initialized")
	}
	
	if cleaner.safetyValidator == nil {
		t.Error("Expected safety validator to be initialized")
	}
}

func TestGetDefaultRemovalPolicy(t *testing.T) {
	policy := GetDefaultRemovalPolicy()
	
	if policy.MinRiskLevel != scanner.TelemetryRiskMedium {
		t.Errorf("Expected MinRiskLevel to be Medium, got %v", policy.MinRiskLevel)
	}
	
	if !policy.PreserveRecent {
		t.Error("Expected PreserveRecent to be true")
	}
	
	if !policy.CreateBackups {
		t.Error("Expected CreateBackups to be true")
	}
	
	if !policy.VerifyBackups {
		t.Error("Expected VerifyBackups to be true")
	}
	
	if policy.RecentThreshold != 24*time.Hour {
		t.Errorf("Expected RecentThreshold to be 24h, got %v", policy.RecentThreshold)
	}
}

func TestGetAggressiveRemovalPolicy(t *testing.T) {
	policy := GetAggressiveRemovalPolicy()
	
	if policy.MinRiskLevel != scanner.TelemetryRiskLow {
		t.Errorf("Expected MinRiskLevel to be Low, got %v", policy.MinRiskLevel)
	}
	
	if policy.PreserveRecent {
		t.Error("Expected PreserveRecent to be false for aggressive policy")
	}
	
	if len(policy.IncludePatterns) == 0 {
		t.Error("Expected IncludePatterns to be specified for aggressive policy")
	}
}

func TestGetConservativeRemovalPolicy(t *testing.T) {
	policy := GetConservativeRemovalPolicy()
	
	if policy.MinRiskLevel != scanner.TelemetryRiskHigh {
		t.Errorf("Expected MinRiskLevel to be High, got %v", policy.MinRiskLevel)
	}
	
	if policy.MaxFileAge != 30*24*time.Hour {
		t.Errorf("Expected MaxFileAge to be 30 days, got %v", policy.MaxFileAge)
	}
	
	if policy.RecentThreshold != 7*24*time.Hour {
		t.Errorf("Expected RecentThreshold to be 7 days, got %v", policy.RecentThreshold)
	}
}

func TestExtensionCleanerShouldCleanItem(t *testing.T) {
	policy := GetDefaultRemovalPolicy()
	cleaner := NewExtensionCleaner(policy)
	
	tests := []struct {
		name     string
		item     scanner.StorageDataItem
		expected bool
	}{
		{
			name: "high_risk_item",
			item: scanner.StorageDataItem{
				Key:          "telemetryData",
				Risk:         scanner.TelemetryRiskHigh,
				LastModified: time.Now().Add(-48 * time.Hour), // 2 days old
			},
			expected: true,
		},
		{
			name: "low_risk_item",
			item: scanner.StorageDataItem{
				Key:          "config",
				Risk:         scanner.TelemetryRiskLow,
				LastModified: time.Now().Add(-48 * time.Hour),
			},
			expected: false, // Below minimum risk level
		},
		{
			name: "recent_item",
			item: scanner.StorageDataItem{
				Key:          "telemetryData",
				Risk:         scanner.TelemetryRiskHigh,
				LastModified: time.Now().Add(-1 * time.Hour), // 1 hour old
			},
			expected: false, // Too recent
		},
		{
			name: "excluded_pattern",
			item: scanner.StorageDataItem{
				Key:          "settings",
				Risk:         scanner.TelemetryRiskHigh,
				LastModified: time.Now().Add(-48 * time.Hour),
			},
			expected: false, // Matches exclude pattern
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := cleaner.shouldCleanItem(test.item)
			if result != test.expected {
				t.Errorf("shouldCleanItem() = %v, want %v", result, test.expected)
			}
		})
	}
}

func TestExtensionCleanerMatchesPattern(t *testing.T) {
	policy := GetDefaultRemovalPolicy()
	cleaner := NewExtensionCleaner(policy)
	
	tests := []struct {
		text     string
		pattern  string
		expected bool
	}{
		{"telemetryData", "telemetry", true},
		{"userSettings", "settings", true},
		{"configFile", "config", true}, // Simple contains matching
		{"randomData", "telemetry", false},
		{"anything", "*", true},
		{"test.config.json", "config", true}, // Simple contains matching
	}
	
	for _, test := range tests {
		result := cleaner.matchesPattern(test.text, test.pattern)
		if result != test.expected {
			t.Errorf("matchesPattern(%s, %s) = %v, want %v", 
				test.text, test.pattern, result, test.expected)
		}
	}
}

func TestNewBackupManager(t *testing.T) {
	manager := NewBackupManager()
	
	if manager == nil {
		t.Fatal("NewBackupManager() returned nil")
	}
	
	if manager.backupDirectory == "" {
		t.Error("Expected backup directory to be set")
	}
	
	if manager.maxBackupAge <= 0 {
		t.Error("Expected maxBackupAge to be positive")
	}
	
	if manager.maxBackupSize <= 0 {
		t.Error("Expected maxBackupSize to be positive")
	}
}

func TestBackupManagerGenerateBackupID(t *testing.T) {
	manager := NewBackupManager()
	
	id1 := manager.generateBackupID()
	time.Sleep(1 * time.Second) // Ensure different timestamp (Unix timestamp precision)
	id2 := manager.generateBackupID()
	
	if id1 == id2 {
		t.Error("Expected different backup IDs")
	}
	
	if id1 == "" || id2 == "" {
		t.Error("Expected non-empty backup IDs")
	}
}

func TestNewDependencyChecker(t *testing.T) {
	checker := NewDependencyChecker()
	
	if checker == nil {
		t.Fatal("NewDependencyChecker() returned nil")
	}
	
	if checker.extensionRegistry == nil {
		t.Error("Expected extension registry to be initialized")
	}
	
	if checker.dependencyGraph == nil {
		t.Error("Expected dependency graph to be initialized")
	}
}

func TestDependencyCheckerIsExtensionActive(t *testing.T) {
	checker := NewDependencyChecker()
	
	tests := []struct {
		extensionID string
		expected    bool
	}{
		{"vscode.git", true},
		{"vscode.typescript-language-features", true},
		{"unknown.extension", false},
		{"", false},
	}
	
	for _, test := range tests {
		result := checker.isExtensionActive(test.extensionID)
		if result != test.expected {
			t.Errorf("isExtensionActive(%s) = %v, want %v", 
				test.extensionID, result, test.expected)
		}
	}
}

func TestNewSafetyValidator(t *testing.T) {
	validator := NewSafetyValidator()
	
	if validator == nil {
		t.Fatal("NewSafetyValidator() returned nil")
	}
	
	if len(validator.criticalPaths) == 0 {
		t.Error("Expected critical paths to be initialized")
	}
	
	if len(validator.protectedPatterns) == 0 {
		t.Error("Expected protected patterns to be initialized")
	}
	
	if len(validator.safetyRules) == 0 {
		t.Error("Expected safety rules to be initialized")
	}
}

func TestSafetyValidatorMatchesPathPattern(t *testing.T) {
	validator := NewSafetyValidator()
	
	tests := []struct {
		path     string
		pattern  string
		expected bool
	}{
		{"user/settings.json", "*settings*", true},
		{"config/app.json", "*config*", true},
		{"data/telemetry.log", "*telemetry*", true},
		{"random/file.txt", "*settings*", false},
		{"", "*pattern*", false},
		{"auth/token.json", "*auth*|*token*", true},
	}
	
	for _, test := range tests {
		result := validator.matchesPathPattern(test.path, test.pattern)
		if result != test.expected {
			t.Errorf("matchesPathPattern(%s, %s) = %v, want %v", 
				test.path, test.pattern, result, test.expected)
		}
	}
}

func TestSafetyValidatorMatchesTemporalPattern(t *testing.T) {
	validator := NewSafetyValidator()
	
	now := time.Now()
	
	tests := []struct {
		name     string
		item     scanner.StorageDataItem
		pattern  string
		expected bool
	}{
		{
			name: "recent_file",
			item: scanner.StorageDataItem{
				LastModified: now.Add(-1 * time.Hour),
			},
			pattern:  "age < 24h",
			expected: true,
		},
		{
			name: "old_file",
			item: scanner.StorageDataItem{
				LastModified: now.Add(-48 * time.Hour),
			},
			pattern:  "age < 24h",
			expected: false,
		},
		{
			name: "week_old_file",
			item: scanner.StorageDataItem{
				LastModified: now.Add(-3 * 24 * time.Hour),
			},
			pattern:  "age < 7d",
			expected: true,
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := validator.matchesTemporalPattern(test.item, test.pattern)
			if result != test.expected {
				t.Errorf("matchesTemporalPattern() = %v, want %v", result, test.expected)
			}
		})
	}
}

func TestSafetyValidatorMatchesSizePattern(t *testing.T) {
	validator := NewSafetyValidator()
	
	tests := []struct {
		name     string
		item     scanner.StorageDataItem
		pattern  string
		expected bool
	}{
		{
			name: "large_file",
			item: scanner.StorageDataItem{
				Size: 200 * 1024 * 1024, // 200MB
			},
			pattern:  "size > 100MB",
			expected: true,
		},
		{
			name: "small_file",
			item: scanner.StorageDataItem{
				Size: 50 * 1024 * 1024, // 50MB
			},
			pattern:  "size > 100MB",
			expected: false,
		},
		{
			name: "medium_file",
			item: scanner.StorageDataItem{
				Size: 5 * 1024 * 1024, // 5MB
			},
			pattern:  "size > 1MB",
			expected: true,
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := validator.matchesSizePattern(test.item, test.pattern)
			if result != test.expected {
				t.Errorf("matchesSizePattern() = %v, want %v", result, test.expected)
			}
		})
	}
}

func TestSafetyValidatorCalculateRiskScore(t *testing.T) {
	validator := NewSafetyValidator()
	
	// Create test items
	items := []scanner.StorageDataItem{
		{
			Risk:         scanner.TelemetryRiskCritical,
			Size:         10 * 1024 * 1024, // 10MB
			LastModified: time.Now().Add(-1 * time.Hour), // Recent
		},
		{
			Risk:         scanner.TelemetryRiskLow,
			Size:         1024, // 1KB
			LastModified: time.Now().Add(-48 * time.Hour), // Old
		},
		{
			Risk:         scanner.TelemetryRiskMedium,
			Size:         5 * 1024 * 1024, // 5MB
			LastModified: time.Now().Add(-12 * time.Hour), // Recent
		},
	}
	
	totalSize := int64(15*1024*1024 + 1024) // ~15MB
	criticalItems := 1
	recentItems := 2
	
	score := validator.calculateRiskScore(items, totalSize, criticalItems, recentItems)
	
	if score < 0.0 || score > 1.0 {
		t.Errorf("Risk score should be between 0.0 and 1.0, got %f", score)
	}
	
	// Should have some risk due to critical and recent items
	if score == 0.0 {
		t.Error("Expected non-zero risk score for items with critical and recent data")
	}
}

func TestSafetyValidatorGetSafetyRules(t *testing.T) {
	validator := NewSafetyValidator()
	
	rules := validator.GetSafetyRules()
	
	if len(rules) == 0 {
		t.Error("Expected safety rules to be returned")
	}
	
	// Check for specific important rules
	foundProtectSettings := false
	foundProtectAuth := false
	
	for _, rule := range rules {
		if rule.Name == "protect_user_settings" {
			foundProtectSettings = true
		}
		if rule.Name == "protect_authentication" {
			foundProtectAuth = true
		}
	}
	
	if !foundProtectSettings {
		t.Error("Expected to find protect_user_settings rule")
	}
	
	if !foundProtectAuth {
		t.Error("Expected to find protect_authentication rule")
	}
}

func TestSafetyValidatorUpdateSafetyRule(t *testing.T) {
	validator := NewSafetyValidator()
	
	// Create a new rule
	newRule := SafetyRule{
		Name:        "test_rule",
		Description: "Test rule for unit testing",
		RuleType:    "test",
		Pattern:     "*test*",
		Action:      "warn",
		Severity:    "low",
		Enabled:     true,
	}
	
	// Add the rule
	validator.UpdateSafetyRule(newRule)
	
	// Verify it was added
	rules := validator.GetSafetyRules()
	found := false
	for _, rule := range rules {
		if rule.Name == "test_rule" {
			found = true
			if rule.Description != newRule.Description {
				t.Error("Rule was not updated correctly")
			}
			break
		}
	}
	
	if !found {
		t.Error("New rule was not added")
	}
	
	// Update the rule
	newRule.Description = "Updated test rule"
	validator.UpdateSafetyRule(newRule)
	
	// Verify it was updated
	rules = validator.GetSafetyRules()
	for _, rule := range rules {
		if rule.Name == "test_rule" {
			if rule.Description != "Updated test rule" {
				t.Error("Rule was not updated correctly")
			}
			break
		}
	}
}

func TestSafetyValidatorDisableSafetyRule(t *testing.T) {
	validator := NewSafetyValidator()
	
	// Disable a rule
	validator.DisableSafetyRule("protect_user_settings")
	
	// Verify it was disabled
	rules := validator.GetSafetyRules()
	for _, rule := range rules {
		if rule.Name == "protect_user_settings" {
			if rule.Enabled {
				t.Error("Rule was not disabled")
			}
			break
		}
	}
}

// Mock file info for testing
type mockFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return m.size }
func (m *mockFileInfo) Mode() os.FileMode  { return m.mode }
func (m *mockFileInfo) ModTime() time.Time { return m.modTime }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }
func (m *mockFileInfo) Sys() interface{}   { return nil }