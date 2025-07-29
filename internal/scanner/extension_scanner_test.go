package scanner

import (
	"testing"
)

func TestNewExtensionScanner(t *testing.T) {
	scanner := NewExtensionScanner()
	
	if scanner == nil {
		t.Fatal("NewExtensionScanner() returned nil")
	}
	
	if len(scanner.telemetryPatterns) == 0 {
		t.Error("Expected telemetry patterns to be initialized")
	}
	
	if len(scanner.riskPatterns) == 0 {
		t.Error("Expected risk patterns to be initialized")
	}
}

func TestTelemetryRiskString(t *testing.T) {
	tests := []struct {
		risk     TelemetryRisk
		expected string
	}{
		{TelemetryRiskNone, "None"},
		{TelemetryRiskLow, "Low"},
		{TelemetryRiskMedium, "Medium"},
		{TelemetryRiskHigh, "High"},
		{TelemetryRiskCritical, "Critical"},
	}
	
	for _, test := range tests {
		if got := test.risk.String(); got != test.expected {
			t.Errorf("TelemetryRisk(%d).String() = %s, want %s", test.risk, got, test.expected)
		}
	}
}

func TestExtensionScannerPatterns(t *testing.T) {
	scanner := NewExtensionScanner()
	
	// Test that critical patterns are properly categorized
	criticalPatterns := scanner.riskPatterns[TelemetryRiskCritical]
	if len(criticalPatterns) == 0 {
		t.Error("Expected critical risk patterns to be defined")
	}
	
	// Check for specific critical patterns
	found := false
	for _, pattern := range criticalPatterns {
		if pattern == "TelemetryReporter" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'TelemetryReporter' to be in critical patterns")
	}
}

func TestGetExtensionDirectories(t *testing.T) {
	scanner := NewExtensionScanner()
	
	directories, err := scanner.getExtensionDirectories()
	if err != nil {
		t.Fatalf("getExtensionDirectories() failed: %v", err)
	}
	
	if len(directories) == 0 {
		t.Error("Expected at least one extension directory")
	}
	
	// Check that .vscode/extensions is included
	found := false
	for _, dir := range directories {
		if dir != "" && (dir[len(dir)-len(".vscode/extensions"):] == ".vscode/extensions" || 
						 dir[len(dir)-len(".vscode\\extensions"):] == ".vscode\\extensions") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected .vscode/extensions directory to be included")
	}
}