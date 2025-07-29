package scanner

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// TelemetryPattern represents a pattern found in extension source code
type TelemetryPattern struct {
	Type        string `json:"type"`
	Pattern     string `json:"pattern"`
	File        string `json:"file"`
	LineNumber  int    `json:"line_number"`
	Context     string `json:"context"`
	Risk        TelemetryRisk `json:"risk"`
	Description string `json:"description"`
}

// ExtensionAnalyzer handles deep analysis of extension source code for telemetry patterns
type ExtensionAnalyzer struct {
	telemetryRegexes map[TelemetryRisk][]*regexp.Regexp
	fileExtensions   []string
}

// NewExtensionAnalyzer creates a new extension analyzer
func NewExtensionAnalyzer() *ExtensionAnalyzer {
	analyzer := &ExtensionAnalyzer{
		fileExtensions: []string{".js", ".ts", ".json"},
	}
	analyzer.initializeTelemetryRegexes()
	return analyzer
}

// initializeTelemetryRegexes sets up regex patterns for detecting telemetry in source code
func (ea *ExtensionAnalyzer) initializeTelemetryRegexes() {
	ea.telemetryRegexes = make(map[TelemetryRisk][]*regexp.Regexp)

	// Critical risk patterns - Direct telemetry usage
	criticalPatterns := []string{
		`new\s+TelemetryReporter\s*\(`,
		`TelemetryReporter\s*\(`,
		`@vscode/extension-telemetry`,
		`vscode-extension-telemetry`,
		`telemetryReporter\s*\.\s*(sendTelemetryEvent|sendTelemetryException)`,
	}

	// High risk patterns - Machine/environment identification
	highPatterns := []string{
		`vscode\.env\.machineId`,
		`vscode\.env\.sessionId`,
		`vscode\.env\.remoteName`,
		`os\.hostname\s*\(\)`,
		`process\.env\.COMPUTERNAME`,
		`process\.env\.USER`,
		`process\.env\.USERNAME`,
		`require\s*\(\s*['"]os['"]`,
	}

	// Medium risk patterns - Network requests and data collection
	mediumPatterns := []string{
		`fetch\s*\(`,
		`axios\s*\.`,
		`http\.request\s*\(`,
		`https\.request\s*\(`,
		`XMLHttpRequest`,
		`navigator\.userAgent`,
		`window\.location`,
		`document\.cookie`,
		`localStorage\.`,
		`sessionStorage\.`,
	}

	// Low risk patterns - General analytics and tracking
	lowPatterns := []string{
		`analytics`,
		`tracking`,
		`metrics`,
		`usage`,
		`statistics`,
		`performance`,
		`error.*report`,
		`crash.*report`,
		`log.*event`,
	}

	// Compile all patterns
	ea.compilePatterns(TelemetryRiskCritical, criticalPatterns)
	ea.compilePatterns(TelemetryRiskHigh, highPatterns)
	ea.compilePatterns(TelemetryRiskMedium, mediumPatterns)
	ea.compilePatterns(TelemetryRiskLow, lowPatterns)
}

// compilePatterns compiles regex patterns for a specific risk level
func (ea *ExtensionAnalyzer) compilePatterns(risk TelemetryRisk, patterns []string) {
	for _, pattern := range patterns {
		if regex, err := regexp.Compile(`(?i)` + pattern); err == nil {
			ea.telemetryRegexes[risk] = append(ea.telemetryRegexes[risk], regex)
		}
	}
}

// AnalyzeExtensionSourceCode performs deep analysis of extension source code
func (ea *ExtensionAnalyzer) AnalyzeExtensionSourceCode(extension *ExtensionInfo) ([]TelemetryPattern, error) {
	var patterns []TelemetryPattern

	// Analyze main entry point if specified
	if extension.Manifest != nil && extension.Manifest.Main != "" {
		mainFile := filepath.Join(extension.InstallPath, extension.Manifest.Main)
		if filePatterns, err := ea.analyzeFile(mainFile); err == nil {
			patterns = append(patterns, filePatterns...)
		}
	}

	// Analyze all JavaScript/TypeScript files in the extension
	err := filepath.Walk(extension.InstallPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue despite errors
		}

		if info.IsDir() {
			// Skip node_modules and other irrelevant directories
			if info.Name() == "node_modules" || info.Name() == ".git" || 
			   info.Name() == "test" || info.Name() == "tests" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file has relevant extension
		if ea.isRelevantFile(path) {
			if filePatterns, err := ea.analyzeFile(path); err == nil {
				patterns = append(patterns, filePatterns...)
			}
		}

		return nil
	})

	if err != nil {
		return patterns, fmt.Errorf("failed to walk extension directory: %w", err)
	}

	// Update extension telemetry information based on findings
	ea.updateExtensionTelemetryInfo(extension, patterns)

	return patterns, nil
}

// isRelevantFile checks if a file should be analyzed for telemetry patterns
func (ea *ExtensionAnalyzer) isRelevantFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	for _, relevantExt := range ea.fileExtensions {
		if ext == relevantExt {
			return true
		}
	}
	return false
}

// analyzeFile analyzes a single file for telemetry patterns
func (ea *ExtensionAnalyzer) analyzeFile(filePath string) ([]TelemetryPattern, error) {
	var patterns []TelemetryPattern

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		// Check line against all telemetry patterns
		for risk, regexes := range ea.telemetryRegexes {
			for _, regex := range regexes {
				if matches := regex.FindAllString(line, -1); len(matches) > 0 {
					for _, match := range matches {
						pattern := TelemetryPattern{
							Type:        ea.getPatternType(match),
							Pattern:     match,
							File:        filePath,
							LineNumber:  lineNumber,
							Context:     strings.TrimSpace(line),
							Risk:        risk,
							Description: ea.getPatternDescription(match, risk),
						}
						patterns = append(patterns, pattern)
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return patterns, fmt.Errorf("error reading file %s: %w", filePath, err)
	}

	return patterns, nil
}

// getPatternType determines the type of telemetry pattern
func (ea *ExtensionAnalyzer) getPatternType(pattern string) string {
	pattern = strings.ToLower(pattern)
	
	switch {
	case strings.Contains(pattern, "telemetryreporter"):
		return "TelemetryReporter"
	case strings.Contains(pattern, "machineid"):
		return "MachineID Access"
	case strings.Contains(pattern, "sessionid"):
		return "SessionID Access"
	case strings.Contains(pattern, "hostname"):
		return "Hostname Access"
	case strings.Contains(pattern, "fetch") || strings.Contains(pattern, "http"):
		return "Network Request"
	case strings.Contains(pattern, "analytics") || strings.Contains(pattern, "tracking"):
		return "Analytics"
	default:
		return "General Telemetry"
	}
}

// getPatternDescription provides a human-readable description of the pattern
func (ea *ExtensionAnalyzer) getPatternDescription(pattern string, risk TelemetryRisk) string {
	pattern = strings.ToLower(pattern)
	
	switch risk {
	case TelemetryRiskCritical:
		return "Direct telemetry implementation - actively collects and sends data"
	case TelemetryRiskHigh:
		return "Accesses machine/user identification - potential privacy concern"
	case TelemetryRiskMedium:
		return "Network communication capability - may send data externally"
	case TelemetryRiskLow:
		return "General analytics/tracking reference - low privacy impact"
	default:
		return "Unknown telemetry pattern"
	}
}

// updateExtensionTelemetryInfo updates extension telemetry information based on analysis
func (ea *ExtensionAnalyzer) updateExtensionTelemetryInfo(extension *ExtensionInfo, patterns []TelemetryPattern) {
	if len(patterns) == 0 {
		return
	}

	extension.HasTelemetry = true
	maxRisk := extension.TelemetryRisk

	// Track unique telemetry types found in source code
	telemetryTypes := make(map[string]bool)
	for _, existing := range extension.TelemetryTypes {
		telemetryTypes[existing] = true
	}

	for _, pattern := range patterns {
		// Update maximum risk level
		if pattern.Risk > maxRisk {
			maxRisk = pattern.Risk
		}

		// Add telemetry type if not already present
		telemetryType := fmt.Sprintf("Source: %s", pattern.Type)
		if !telemetryTypes[telemetryType] {
			extension.TelemetryTypes = append(extension.TelemetryTypes, telemetryType)
			telemetryTypes[telemetryType] = true
		}
	}

	extension.TelemetryRisk = maxRisk
}

// AnalyzeActivationFunction specifically analyzes the activate() function for telemetry
func (ea *ExtensionAnalyzer) AnalyzeActivationFunction(extension *ExtensionInfo) ([]TelemetryPattern, error) {
	var patterns []TelemetryPattern

	if extension.Manifest == nil || extension.Manifest.Main == "" {
		return patterns, nil
	}

	mainFile := filepath.Join(extension.InstallPath, extension.Manifest.Main)
	file, err := os.Open(mainFile)
	if err != nil {
		return patterns, fmt.Errorf("failed to open main file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0
	inActivateFunction := false
	braceCount := 0

	activateFunctionRegex := regexp.MustCompile(`(?i)function\s+activate\s*\(|exports\.activate\s*=|activate\s*:\s*function`)

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		// Check if we're entering the activate function
		if !inActivateFunction && activateFunctionRegex.MatchString(line) {
			inActivateFunction = true
			braceCount = 0
		}

		if inActivateFunction {
			// Count braces to track function scope
			braceCount += strings.Count(line, "{") - strings.Count(line, "}")

			// Analyze line for telemetry patterns with higher weight
			for risk, regexes := range ea.telemetryRegexes {
				for _, regex := range regexes {
					if matches := regex.FindAllString(line, -1); len(matches) > 0 {
						for _, match := range matches {
							pattern := TelemetryPattern{
								Type:        "Activation: " + ea.getPatternType(match),
								Pattern:     match,
								File:        mainFile,
								LineNumber:  lineNumber,
								Context:     strings.TrimSpace(line),
								Risk:        risk,
								Description: "Found in activate() function - " + ea.getPatternDescription(match, risk),
							}
							patterns = append(patterns, pattern)
						}
					}
				}
			}

			// Exit activate function when braces are balanced
			if braceCount <= 0 && strings.Contains(line, "}") {
				inActivateFunction = false
			}
		}
	}

	return patterns, scanner.Err()
}

// AnalyzeCommandHandlers analyzes command handlers for telemetry patterns
func (ea *ExtensionAnalyzer) AnalyzeCommandHandlers(extension *ExtensionInfo) ([]TelemetryPattern, error) {
	var patterns []TelemetryPattern

	for _, command := range extension.Commands {
		commandPatterns, err := ea.findCommandHandler(extension, command)
		if err != nil {
			continue // Skip commands we can't analyze
		}
		patterns = append(patterns, commandPatterns...)
	}

	return patterns, nil
}

// findCommandHandler finds and analyzes the handler for a specific command
func (ea *ExtensionAnalyzer) findCommandHandler(extension *ExtensionInfo, command string) ([]TelemetryPattern, error) {
	var patterns []TelemetryPattern

	// Search for command registration patterns
	commandRegex := regexp.MustCompile(fmt.Sprintf(`(?i)registerCommand\s*\(\s*['"]%s['"]`, regexp.QuoteMeta(command)))

	err := filepath.Walk(extension.InstallPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !ea.isRelevantFile(path) {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNumber := 0

		for scanner.Scan() {
			lineNumber++
			line := scanner.Text()

			if commandRegex.MatchString(line) {
				// Found command registration, analyze surrounding context
				contextPatterns := ea.analyzeCommandContext(path, lineNumber, 10)
				for _, pattern := range contextPatterns {
					pattern.Type = fmt.Sprintf("Command Handler: %s", pattern.Type)
					pattern.Description = fmt.Sprintf("Found in handler for command '%s' - %s", command, pattern.Description)
				}
				patterns = append(patterns, contextPatterns...)
			}
		}

		return nil
	})

	return patterns, err
}

// analyzeCommandContext analyzes lines around a command registration for telemetry
func (ea *ExtensionAnalyzer) analyzeCommandContext(filePath string, centerLine, contextLines int) []TelemetryPattern {
	var patterns []TelemetryPattern

	file, err := os.Open(filePath)
	if err != nil {
		return patterns
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string

	// Read all lines
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Analyze context around the center line
	start := centerLine - contextLines - 1
	end := centerLine + contextLines - 1

	if start < 0 {
		start = 0
	}
	if end >= len(lines) {
		end = len(lines) - 1
	}

	for i := start; i <= end; i++ {
		line := lines[i]
		
		for risk, regexes := range ea.telemetryRegexes {
			for _, regex := range regexes {
				if matches := regex.FindAllString(line, -1); len(matches) > 0 {
					for _, match := range matches {
						pattern := TelemetryPattern{
							Type:        ea.getPatternType(match),
							Pattern:     match,
							File:        filePath,
							LineNumber:  i + 1,
							Context:     strings.TrimSpace(line),
							Risk:        risk,
							Description: ea.getPatternDescription(match, risk),
						}
						patterns = append(patterns, pattern)
					}
				}
			}
		}
	}

	return patterns
}