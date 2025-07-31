package scanner

import (
	"os"
	"testing"
	"time"
)

func TestNewStorageAnalyzer(t *testing.T) {
	analyzer := NewStorageAnalyzer()
	
	if analyzer == nil {
		t.Fatal("NewStorageAnalyzer() returned nil")
	}
	
	if len(analyzer.telemetryPatterns) == 0 {
		t.Error("Expected telemetry patterns to be initialized")
	}
	
	if len(analyzer.cachePatterns) == 0 {
		t.Error("Expected cache patterns to be initialized")
	}
	
	if analyzer.retentionAnalyzer == nil {
		t.Error("Expected retention analyzer to be initialized")
	}
	
	if analyzer.correlationAnalyzer == nil {
		t.Error("Expected correlation analyzer to be initialized")
	}
}

func TestNewRetentionAnalyzer(t *testing.T) {
	analyzer := NewRetentionAnalyzer()
	
	if analyzer == nil {
		t.Fatal("NewRetentionAnalyzer() returned nil")
	}
	
	if len(analyzer.defaultRetentionPeriods) == 0 {
		t.Error("Expected default retention periods to be initialized")
	}
	
	if len(analyzer.policyPatterns) == 0 {
		t.Error("Expected policy patterns to be initialized")
	}
}

func TestNewCorrelationAnalyzer(t *testing.T) {
	analyzer := NewCorrelationAnalyzer()
	
	if analyzer == nil {
		t.Fatal("NewCorrelationAnalyzer() returned nil")
	}
	
	if len(analyzer.correlationPatterns) == 0 {
		t.Error("Expected correlation patterns to be initialized")
	}
	
	if len(analyzer.sharedDataTypes) == 0 {
		t.Error("Expected shared data types to be initialized")
	}
}

func TestRetentionPolicyTypeString(t *testing.T) {
	tests := []struct {
		policy   RetentionPolicyType
		expected string
	}{
		{RetentionPolicyNone, "None"},
		{RetentionPolicySession, "Session"},
		{RetentionPolicyDaily, "Daily"},
		{RetentionPolicyWeekly, "Weekly"},
		{RetentionPolicyMonthly, "Monthly"},
		{RetentionPolicyPermanent, "Permanent"},
		{RetentionPolicyCustom, "Custom"},
	}
	
	for _, test := range tests {
		if got := test.policy.String(); got != test.expected {
			t.Errorf("RetentionPolicyType(%d).String() = %s, want %s", test.policy, got, test.expected)
		}
	}
}

func TestRetentionAnalyzerParseRetentionPeriod(t *testing.T) {
	analyzer := NewRetentionAnalyzer()
	
	tests := []struct {
		input    interface{}
		expected time.Duration
	}{
		{"24h", 24 * time.Hour},
		{"1d", 0}, // Should not parse "1d" format
		{"day", 24 * time.Hour},
		{"week", 7 * 24 * time.Hour},
		{"month", 30 * 24 * time.Hour},
		{"year", 365 * 24 * time.Hour},
		{24.0, 24 * time.Hour},
		{24, 24 * time.Hour},
		{"invalid", 0},
	}
	
	for _, test := range tests {
		got := analyzer.parseRetentionPeriod(test.input)
		if got != test.expected {
			t.Errorf("parseRetentionPeriod(%v) = %v, want %v", test.input, got, test.expected)
		}
	}
}

func TestRetentionAnalyzerAnalyzeRetentionPolicy(t *testing.T) {
	analyzer := NewRetentionAnalyzer()
	
	// Test with non-existent path
	policy := analyzer.AnalyzeRetentionPolicy("test.extension", "/non/existent/path")
	
	if policy.HasPolicy {
		t.Error("Expected HasPolicy to be false for non-existent path")
	}
	
	if policy.PolicySource != "default" {
		t.Errorf("Expected PolicySource to be 'default', got '%s'", policy.PolicySource)
	}
	
	if policy.RetentionPeriod <= 0 {
		t.Error("Expected default retention period to be set")
	}
}

func TestCorrelationAnalyzerMatchesKeyPattern(t *testing.T) {
	analyzer := NewCorrelationAnalyzer()
	
	tests := []struct {
		key     string
		pattern string
		matches bool
	}{
		{"machineId", "machineid", true},
		{"machine_id", "machine", true}, // Partial match works
		{"deviceIdentifier", "device", true}, // Partial match works
		{"userId", "user", true}, // Partial match works
		{"randomKey", "machineid", false},
		{"", "pattern", false},
	}
	
	for _, test := range tests {
		got := analyzer.matchesKeyPattern(test.key, test.pattern)
		if got != test.matches {
			t.Errorf("matchesKeyPattern(%s, %s) = %v, want %v", test.key, test.pattern, got, test.matches)
		}
	}
}

func TestCorrelationAnalyzerHashValue(t *testing.T) {
	analyzer := NewCorrelationAnalyzer()
	
	tests := []struct {
		value    interface{}
		hasHash  bool
	}{
		{"test-value", true},
		{"", false},
		{"ab", false}, // Too short
		{nil, false},
		{"true", false}, // Trivial value
		{"false", false}, // Trivial value
		{"null", false}, // Trivial value
		{123, true},
	}
	
	for _, test := range tests {
		hash := analyzer.hashValue(test.value)
		hasHash := hash != ""
		
		if hasHash != test.hasHash {
			t.Errorf("hashValue(%v) hash existence = %v, want %v", test.value, hasHash, test.hasHash)
		}
	}
}

func TestCorrelationAnalyzerUniqueStrings(t *testing.T) {
	analyzer := NewCorrelationAnalyzer()
	
	input := []string{"a", "b", "a", "c", "b", "d"}
	expected := []string{"a", "b", "c", "d"}
	
	result := analyzer.uniqueStrings(input)
	
	if len(result) != len(expected) {
		t.Errorf("uniqueStrings() returned %d items, want %d", len(result), len(expected))
	}
	
	// Check that all expected items are present
	resultMap := make(map[string]bool)
	for _, item := range result {
		resultMap[item] = true
	}
	
	for _, expectedItem := range expected {
		if !resultMap[expectedItem] {
			t.Errorf("uniqueStrings() missing expected item: %s", expectedItem)
		}
	}
}

func TestStorageAnalyzerAssessFileRisk(t *testing.T) {
	analyzer := NewStorageAnalyzer()
	
	tests := []struct {
		fileName string
		filePath string
		minRisk  TelemetryRisk
	}{
		{"telemetryData.json", "/path/to/telemetryData.json", TelemetryRiskCritical},
		{"analyticsData.log", "/path/to/analyticsData.log", TelemetryRiskCritical},
		{"machineId.txt", "/path/to/machineId.txt", TelemetryRiskCritical},
		{"config.json", "/path/to/config.json", TelemetryRiskNone},
		{"usageStats.data", "/path/to/usageStats.data", TelemetryRiskHigh},
		{"cache.tmp", "/path/to/cache.tmp", TelemetryRiskNone}, // cache alone doesn't trigger risk
	}
	
	for _, test := range tests {
		risk := analyzer.assessFileRisk(test.fileName, test.filePath)
		if risk < test.minRisk {
			t.Errorf("assessFileRisk(%s, %s) = %v, want at least %v", 
				test.fileName, test.filePath, risk, test.minRisk)
		}
	}
}

func TestStorageAnalyzerInferExtensionFromPath(t *testing.T) {
	analyzer := NewStorageAnalyzer()
	
	tests := []struct {
		path     string
		expected string
	}{
		{"/path/to/cpptools/cache", "ms-vscode.cpptools"},
		{"/path/to/eslint/data", "dbaeumer.vscode-eslint"},
		{"/path/to/typescript/logs", "vscode.typescript-language-features"},
		{"/path/to/python/cache", "ms-python.python"},
		{"/path/to/unknown/cache", "unknown"},
		{"", "unknown"},
	}
	
	for _, test := range tests {
		result := analyzer.inferExtensionFromPath(test.path)
		if result != test.expected {
			t.Errorf("inferExtensionFromPath(%s) = %s, want %s", test.path, result, test.expected)
		}
	}
}

func TestStorageAnalyzerInferCacheType(t *testing.T) {
	analyzer := NewStorageAnalyzer()
	
	tests := []struct {
		path     string
		expected string
	}{
		{"/path/to/logs/cache", "logs"},
		{"/path/to/temp/data", "temporary"},
		{"/path/to/data/store", "data"},
		{"/path/to/cache/files", "cache"},
		{"/path/to/other/stuff", "general"},
	}
	
	for _, test := range tests {
		result := analyzer.inferCacheType(test.path)
		if result != test.expected {
			t.Errorf("inferCacheType(%s) = %s, want %s", test.path, result, test.expected)
		}
	}
}

func TestStorageAnalyzerInferFileType(t *testing.T) {
	analyzer := NewStorageAnalyzer()
	
	tests := []struct {
		fileName string
		expected string
	}{
		{"test.log", "log"},
		{"config.json", "json"},
		{"temp.tmp", "temporary"},
		{"data.cache", "cache"},
		{"binary.exe", "binary"},
		{"noextension", "binary"},
	}
	
	for _, test := range tests {
		result := analyzer.inferFileType(test.fileName)
		if result != test.expected {
			t.Errorf("inferFileType(%s) = %s, want %s", test.fileName, result, test.expected)
		}
	}
}

func TestStorageAnalyzerIsExtensionRelated(t *testing.T) {
	analyzer := NewStorageAnalyzer()
	
	tests := []struct {
		fileName string
		filePath string
		expected bool
	}{
		{"vscode-temp.log", "/tmp/vscode-temp.log", true},
		{"extension-data.json", "/path/extension-data.json", true},
		{"ms-python.cache", "/cache/ms-python.cache", true},
		{"eslint.tmp", "/tmp/eslint.tmp", true},
		{"random-file.txt", "/path/random-file.txt", false},
		{"system.log", "/var/log/system.log", false},
	}
	
	for _, test := range tests {
		result := analyzer.isExtensionRelated(test.fileName, test.filePath)
		if result != test.expected {
			t.Errorf("isExtensionRelated(%s, %s) = %v, want %v", 
				test.fileName, test.filePath, result, test.expected)
		}
	}
}

func TestStorageAnalyzerEstimateAccessFrequency(t *testing.T) {
	analyzer := NewStorageAnalyzer()
	
	// Create mock file info with different ages
	now := time.Now()
	
	tests := []struct {
		age      time.Duration
		minFreq  int
		maxFreq  int
	}{
		{1 * time.Hour, 10, 10},    // Recent file - high frequency
		{3 * 24 * time.Hour, 5, 5}, // 3 days old - medium frequency
		{15 * 24 * time.Hour, 2, 2}, // 15 days old - low frequency
		{60 * 24 * time.Hour, 1, 1}, // 60 days old - very low frequency
	}
	
	for _, test := range tests {
		// Create a mock file info
		mockTime := now.Add(-test.age)
		mockInfo := &mockFileInfo{modTime: mockTime}
		
		freq := analyzer.estimateAccessFrequency(mockInfo)
		if freq < test.minFreq || freq > test.maxFreq {
			t.Errorf("estimateAccessFrequency(age: %v) = %d, want between %d and %d", 
				test.age, freq, test.minFreq, test.maxFreq)
		}
	}
}

// mockFileInfo is a simple mock implementation of os.FileInfo for testing
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