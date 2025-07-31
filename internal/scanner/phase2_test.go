package scanner

import (
	"testing"
)

func TestNewConfigAnalyzer(t *testing.T) {
	analyzer := NewConfigAnalyzer()
	
	if analyzer == nil {
		t.Fatal("NewConfigAnalyzer() returned nil")
	}
	
	if len(analyzer.telemetryKeys) == 0 {
		t.Error("Expected telemetry keys to be initialized")
	}
	
	if len(analyzer.extensionPatterns) == 0 {
		t.Error("Expected extension patterns to be initialized")
	}
}

func TestNewExtensionSettingsScanner(t *testing.T) {
	scanner := NewExtensionSettingsScanner()
	
	if scanner == nil {
		t.Fatal("NewExtensionSettingsScanner() returned nil")
	}
	
	if len(scanner.telemetryKeyPatterns) == 0 {
		t.Error("Expected telemetry key patterns to be initialized")
	}
	
	if len(scanner.storageKeyPatterns) == 0 {
		t.Error("Expected storage key patterns to be initialized")
	}
}

func TestNewAdvancedPatternMatcher(t *testing.T) {
	matcher := NewAdvancedPatternMatcher()
	
	if matcher == nil {
		t.Fatal("NewAdvancedPatternMatcher() returned nil")
	}
	
	if len(matcher.contextPatterns) == 0 {
		t.Error("Expected context patterns to be initialized")
	}
	
	if len(matcher.semanticPatterns) == 0 {
		t.Error("Expected semantic patterns to be initialized")
	}
	
	if len(matcher.combinationRules) == 0 {
		t.Error("Expected combination rules to be initialized")
	}
}

func TestNewDatabaseAnalyzer(t *testing.T) {
	analyzer := NewDatabaseAnalyzer()
	
	if analyzer == nil {
		t.Fatal("NewDatabaseAnalyzer() returned nil")
	}
	
	if len(analyzer.telemetryKeyPatterns) == 0 {
		t.Error("Expected telemetry key patterns to be initialized")
	}
	
	if len(analyzer.extensionPatterns) == 0 {
		t.Error("Expected extension patterns to be initialized")
	}
	
	if len(analyzer.tableAnalyzers) == 0 {
		t.Error("Expected table analyzers to be initialized")
	}
}

func TestAdvancedPatternMatcherAnalyzeCode(t *testing.T) {
	matcher := NewAdvancedPatternMatcher()
	
	// Test code with telemetry patterns
	testCode := `
import TelemetryReporter from '@vscode/extension-telemetry';

function activate(context) {
    const reporter = new TelemetryReporter('test', '1.0.0', 'key');
    const machineId = vscode.env.machineId;
    reporter.sendTelemetryEvent('activation', { machineId });
}
`
	
	matches := matcher.AnalyzeCode(testCode, "test.js")
	
	if len(matches) == 0 {
		t.Error("Expected to find telemetry patterns in test code")
	}
	
	// Check for high-risk patterns
	foundHighRisk := false
	for _, match := range matches {
		if match.Risk >= TelemetryRiskHigh {
			foundHighRisk = true
			break
		}
	}
	
	if !foundHighRisk {
		t.Error("Expected to find high-risk telemetry patterns")
	}
}

func TestPatternMatcherStatistics(t *testing.T) {
	matcher := NewAdvancedPatternMatcher()
	
	stats := matcher.GetPatternStatistics()
	
	if len(stats) == 0 {
		t.Error("Expected pattern statistics to be available")
	}
	
	expectedKeys := []string{"function_calls", "assignments", "imports", "config_access", "semantic_patterns", "combination_rules", "exclusion_patterns"}
	
	for _, key := range expectedKeys {
		if _, exists := stats[key]; !exists {
			t.Errorf("Expected statistics key '%s' to exist", key)
		}
	}
}

func TestTelemetryPatternManager(t *testing.T) {
	manager := NewTelemetryPatternManager()
	
	if manager == nil {
		t.Fatal("NewTelemetryPatternManager() returned nil")
	}
	
	patterns := manager.GetPatterns()
	if len(patterns) == 0 {
		t.Error("Expected telemetry patterns to be initialized")
	}
	
	// Test pattern matching
	testLine := "new TelemetryReporter('test', '1.0.0', 'key')"
	matches := manager.MatchLine(testLine)
	
	if len(matches) == 0 {
		t.Error("Expected to match telemetry pattern in test line")
	}
}

func TestGetRiskDescription(t *testing.T) {
	tests := []struct {
		risk     TelemetryRisk
		expected string
	}{
		{TelemetryRiskNone, "None: No telemetry patterns detected"},
		{TelemetryRiskLow, "Low: Contains references to analytics/tracking but limited data collection"},
		{TelemetryRiskMedium, "Medium: Has network communication capabilities or accesses local storage"},
		{TelemetryRiskHigh, "High: Accesses machine/user identification or sends data to external services"},
		{TelemetryRiskCritical, "Critical: Actively collects and transmits telemetry data"},
	}
	
	for _, test := range tests {
		if got := GetRiskDescription(test.risk); got != test.expected {
			t.Errorf("GetRiskDescription(%v) = %s, want %s", test.risk, got, test.expected)
		}
	}
}

func TestGetRiskColor(t *testing.T) {
	tests := []struct {
		risk     TelemetryRisk
		expected string
	}{
		{TelemetryRiskNone, "#CCCCCC"},
		{TelemetryRiskLow, "#00CC00"},
		{TelemetryRiskMedium, "#FFCC00"},
		{TelemetryRiskHigh, "#FF6600"},
		{TelemetryRiskCritical, "#FF0000"},
	}
	
	for _, test := range tests {
		if got := GetRiskColor(test.risk); got != test.expected {
			t.Errorf("GetRiskColor(%v) = %s, want %s", test.risk, got, test.expected)
		}
	}
}