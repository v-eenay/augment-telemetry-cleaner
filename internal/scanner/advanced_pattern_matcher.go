package scanner

import (
	"fmt"
	"regexp"
	"strings"
)

// AdvancedPatternMatcher provides sophisticated pattern matching for telemetry detection
type AdvancedPatternMatcher struct {
	contextPatterns    map[string][]*regexp.Regexp
	semanticPatterns   map[string]TelemetryRisk
	combinationRules   []CombinationRule
	exclusionPatterns  []*regexp.Regexp
}

// CombinationRule defines rules for combining multiple pattern matches
type CombinationRule struct {
	Name        string        `json:"name"`
	Patterns    []string      `json:"patterns"`
	MinMatches  int           `json:"min_matches"`
	Risk        TelemetryRisk `json:"risk"`
	Description string        `json:"description"`
}

// PatternMatch represents a sophisticated pattern match with context
type PatternMatch struct {
	Pattern     string        `json:"pattern"`
	Match       string        `json:"match"`
	Context     string        `json:"context"`
	Risk        TelemetryRisk `json:"risk"`
	Confidence  float64       `json:"confidence"`
	Category    string        `json:"category"`
	Line        int           `json:"line"`
	Column      int           `json:"column"`
	Surrounding []string      `json:"surrounding"`
}

// NewAdvancedPatternMatcher creates a new advanced pattern matcher
func NewAdvancedPatternMatcher() *AdvancedPatternMatcher {
	matcher := &AdvancedPatternMatcher{
		contextPatterns:  make(map[string][]*regexp.Regexp),
		semanticPatterns: make(map[string]TelemetryRisk),
	}
	matcher.initializeContextPatterns()
	matcher.initializeSemanticPatterns()
	matcher.initializeCombinationRules()
	matcher.initializeExclusionPatterns()
	return matcher
}

// initializeContextPatterns sets up context-aware patterns
func (apm *AdvancedPatternMatcher) initializeContextPatterns() {
	// Function call context patterns
	functionPatterns := []string{
		`(?i)new\s+TelemetryReporter\s*\([^)]*\)`,
		`(?i)\.sendTelemetryEvent\s*\([^)]*\)`,
		`(?i)\.sendTelemetryException\s*\([^)]*\)`,
		`(?i)\.trackEvent\s*\([^)]*\)`,
		`(?i)\.trackException\s*\([^)]*\)`,
		`(?i)fetch\s*\(\s*['"][^'"]*telemetry[^'"]*['"]`,
		`(?i)axios\.\w+\s*\(\s*['"][^'"]*analytics[^'"]*['"]`,
		`(?i)http\.request\s*\([^)]*telemetry[^)]*\)`,
	}

	// Variable assignment context patterns
	assignmentPatterns := []string{
		`(?i)(?:const|let|var)\s+\w*(?:telemetry|analytics|tracking)\w*\s*=`,
		`(?i)\w*(?:machineId|deviceId|sessionId)\w*\s*=\s*vscode\.env\.\w+`,
		`(?i)\w*hostname\w*\s*=\s*os\.hostname\s*\(\)`,
		`(?i)\w*userAgent\w*\s*=\s*navigator\.userAgent`,
	}

	// Import/require context patterns
	importPatterns := []string{
		`(?i)(?:import|require)\s*\([^)]*(?:telemetry|analytics|applicationinsights)[^)]*\)`,
		`(?i)from\s+['"][^'"]*(?:telemetry|analytics)[^'"]*['"]`,
		`(?i)import\s+.*\s+from\s+['"][^'"]*(?:telemetry|analytics)[^'"]*['"]`,
	}

	// Configuration access patterns
	configPatterns := []string{
		`(?i)vscode\.workspace\.getConfiguration\s*\([^)]*(?:telemetry|analytics)[^)]*\)`,
		`(?i)context\.globalState\.(?:get|update)\s*\([^)]*(?:telemetry|usage|analytics)[^)]*\)`,
		`(?i)context\.workspaceState\.(?:get|update)\s*\([^)]*(?:telemetry|usage)[^)]*\)`,
	}

	apm.compileContextPatterns("function_calls", functionPatterns)
	apm.compileContextPatterns("assignments", assignmentPatterns)
	apm.compileContextPatterns("imports", importPatterns)
	apm.compileContextPatterns("config_access", configPatterns)
}

// initializeSemanticPatterns sets up semantic analysis patterns
func (apm *AdvancedPatternMatcher) initializeSemanticPatterns() {
	apm.semanticPatterns = map[string]TelemetryRisk{
		// High-confidence telemetry indicators
		"telemetryreporter":           TelemetryRiskCritical,
		"sendtelemetryevent":          TelemetryRiskCritical,
		"sendtelemetryexception":      TelemetryRiskCritical,
		"applicationinsights":         TelemetryRiskCritical,
		"trackevent":                  TelemetryRiskCritical,
		"trackexception":              TelemetryRiskCritical,
		
		// Machine/user identification
		"vscode.env.machineid":        TelemetryRiskHigh,
		"vscode.env.sessionid":        TelemetryRiskHigh,
		"os.hostname":                 TelemetryRiskHigh,
		"navigator.useragent":         TelemetryRiskHigh,
		"process.env.user":            TelemetryRiskHigh,
		"process.env.username":        TelemetryRiskHigh,
		"process.env.computername":    TelemetryRiskHigh,
		
		// Network communication with telemetry endpoints
		"fetch.*telemetry":            TelemetryRiskHigh,
		"axios.*analytics":            TelemetryRiskHigh,
		"http.*telemetry":             TelemetryRiskHigh,
		
		// Data collection and storage
		"globalstate.*telemetry":      TelemetryRiskMedium,
		"workspacestate.*usage":       TelemetryRiskMedium,
		"localstorage.*analytics":     TelemetryRiskMedium,
		"sessionstorage.*tracking":    TelemetryRiskMedium,
		
		// Performance and usage tracking
		"performance.now":             TelemetryRiskLow,
		"performance.mark":            TelemetryRiskLow,
		"performance.measure":         TelemetryRiskLow,
		"console.time":                TelemetryRiskLow,
		
		// Error and crash reporting
		"crashreporter":               TelemetryRiskMedium,
		"errorreporter":               TelemetryRiskMedium,
		"uncaughtexception":           TelemetryRiskMedium,
		"unhandledrejection":          TelemetryRiskMedium,
	}
}

// initializeCombinationRules sets up rules for combining pattern matches
func (apm *AdvancedPatternMatcher) initializeCombinationRules() {
	apm.combinationRules = []CombinationRule{
		{
			Name:        "Active Telemetry Implementation",
			Patterns:    []string{"telemetryreporter", "sendtelemetryevent", "machineid"},
			MinMatches:  2,
			Risk:        TelemetryRiskCritical,
			Description: "Extension actively implements telemetry with machine identification",
		},
		{
			Name:        "Network Telemetry",
			Patterns:    []string{"fetch.*telemetry", "axios.*analytics", "http.*telemetry"},
			MinMatches:  1,
			Risk:        TelemetryRiskHigh,
			Description: "Extension makes network requests to telemetry endpoints",
		},
		{
			Name:        "User Identification Combo",
			Patterns:    []string{"machineid", "hostname", "useragent", "username"},
			MinMatches:  2,
			Risk:        TelemetryRiskHigh,
			Description: "Extension collects multiple user/machine identifiers",
		},
		{
			Name:        "Data Collection and Storage",
			Patterns:    []string{"globalstate.*telemetry", "localstorage.*analytics", "performance"},
			MinMatches:  2,
			Risk:        TelemetryRiskMedium,
			Description: "Extension collects and stores usage/performance data",
		},
	}
}

// initializeExclusionPatterns sets up patterns to exclude false positives
func (apm *AdvancedPatternMatcher) initializeExclusionPatterns() {
	exclusionPatterns := []string{
		// Comments and documentation
		`(?i)//.*(?:telemetry|analytics|tracking)`,
		`(?i)/\*.*(?:telemetry|analytics|tracking).*\*/`,
		`(?i)\*.*(?:telemetry|analytics|tracking)`,
		
		// String literals that are just labels/messages
		`(?i)['"].*(?:disable|turn off|opt out).*(?:telemetry|analytics).*['"]`,
		`(?i)['"].*(?:telemetry|analytics).*(?:disabled|off|false).*['"]`,
		
		// Configuration descriptions
		`(?i)description.*['"].*(?:telemetry|analytics).*['"]`,
		`(?i)title.*['"].*(?:telemetry|analytics).*['"]`,
		
		// Test files and mock data
		`(?i)test.*(?:telemetry|analytics)`,
		`(?i)mock.*(?:telemetry|analytics)`,
		`(?i)spec.*(?:telemetry|analytics)`,
	}

	for _, pattern := range exclusionPatterns {
		if regex, err := regexp.Compile(pattern); err == nil {
			apm.exclusionPatterns = append(apm.exclusionPatterns, regex)
		}
	}
}

// compileContextPatterns compiles patterns for a specific context
func (apm *AdvancedPatternMatcher) compileContextPatterns(context string, patterns []string) {
	for _, pattern := range patterns {
		if regex, err := regexp.Compile(pattern); err == nil {
			apm.contextPatterns[context] = append(apm.contextPatterns[context], regex)
		}
	}
}

// AnalyzeCode performs advanced pattern analysis on code content
func (apm *AdvancedPatternMatcher) AnalyzeCode(content string, filePath string) []PatternMatch {
	var matches []PatternMatch
	lines := strings.Split(content, "\n")

	for lineNum, line := range lines {
		lineMatches := apm.analyzeLine(line, lineNum+1, lines, filePath)
		matches = append(matches, lineMatches...)
	}

	// Apply combination rules
	combinationMatches := apm.applyCombinationRules(matches, content, filePath)
	matches = append(matches, combinationMatches...)

	// Filter out exclusions
	matches = apm.filterExclusions(matches)

	// Calculate confidence scores
	matches = apm.calculateConfidence(matches)

	return matches
}

// analyzeLine analyzes a single line of code
func (apm *AdvancedPatternMatcher) analyzeLine(line string, lineNum int, allLines []string, filePath string) []PatternMatch {
	var matches []PatternMatch

	// Check context patterns
	for context, patterns := range apm.contextPatterns {
		for _, pattern := range patterns {
			if submatches := pattern.FindAllStringSubmatch(line, -1); len(submatches) > 0 {
				for _, submatch := range submatches {
					match := PatternMatch{
						Pattern:     pattern.String(),
						Match:       submatch[0],
						Context:     line,
						Risk:        apm.determineContextRisk(context, submatch[0]),
						Category:    context,
						Line:        lineNum,
						Column:      strings.Index(line, submatch[0]),
						Surrounding: apm.getSurroundingLines(allLines, lineNum, 2),
					}
					matches = append(matches, match)
				}
			}
		}
	}

	// Check semantic patterns
	lowerLine := strings.ToLower(line)
	for pattern, risk := range apm.semanticPatterns {
		if strings.Contains(lowerLine, strings.ToLower(pattern)) {
			match := PatternMatch{
				Pattern:     pattern,
				Match:       pattern,
				Context:     line,
				Risk:        risk,
				Category:    "semantic",
				Line:        lineNum,
				Column:      strings.Index(lowerLine, strings.ToLower(pattern)),
				Surrounding: apm.getSurroundingLines(allLines, lineNum, 2),
			}
			matches = append(matches, match)
		}
	}

	return matches
}

// applyCombinationRules applies combination rules to enhance detection
func (apm *AdvancedPatternMatcher) applyCombinationRules(matches []PatternMatch, content, filePath string) []PatternMatch {
	var combinationMatches []PatternMatch

	for _, rule := range apm.combinationRules {
		matchCount := 0
		var ruleMatches []PatternMatch

		for _, match := range matches {
			for _, rulePattern := range rule.Patterns {
				if strings.Contains(strings.ToLower(match.Pattern), strings.ToLower(rulePattern)) ||
				   strings.Contains(strings.ToLower(match.Match), strings.ToLower(rulePattern)) {
					matchCount++
					ruleMatches = append(ruleMatches, match)
					break
				}
			}
		}

		if matchCount >= rule.MinMatches {
			combinationMatch := PatternMatch{
				Pattern:     rule.Name,
				Match:       fmt.Sprintf("Combination rule matched (%d patterns)", matchCount),
				Context:     rule.Description,
				Risk:        rule.Risk,
				Category:    "combination",
				Confidence:  0.9, // High confidence for combination matches
			}
			combinationMatches = append(combinationMatches, combinationMatch)
		}
	}

	return combinationMatches
}

// filterExclusions removes false positives based on exclusion patterns
func (apm *AdvancedPatternMatcher) filterExclusions(matches []PatternMatch) []PatternMatch {
	var filtered []PatternMatch

	for _, match := range matches {
		excluded := false
		
		for _, exclusionPattern := range apm.exclusionPatterns {
			if exclusionPattern.MatchString(match.Context) {
				excluded = true
				break
			}
		}

		if !excluded {
			filtered = append(filtered, match)
		}
	}

	return filtered
}

// calculateConfidence calculates confidence scores for matches
func (apm *AdvancedPatternMatcher) calculateConfidence(matches []PatternMatch) []PatternMatch {
	for i := range matches {
		match := &matches[i]
		
		// Base confidence based on risk level
		switch match.Risk {
		case TelemetryRiskCritical:
			match.Confidence = 0.95
		case TelemetryRiskHigh:
			match.Confidence = 0.85
		case TelemetryRiskMedium:
			match.Confidence = 0.70
		case TelemetryRiskLow:
			match.Confidence = 0.50
		default:
			match.Confidence = 0.30
		}

		// Adjust confidence based on context
		if match.Category == "function_calls" {
			match.Confidence += 0.05
		}
		if match.Category == "combination" {
			match.Confidence += 0.10
		}

		// Adjust confidence based on match specificity
		if len(match.Match) > 20 {
			match.Confidence += 0.05
		}

		// Cap confidence at 1.0
		if match.Confidence > 1.0 {
			match.Confidence = 1.0
		}
	}

	return matches
}

// determineContextRisk determines risk level based on context
func (apm *AdvancedPatternMatcher) determineContextRisk(context, match string) TelemetryRisk {
	switch context {
	case "function_calls":
		if strings.Contains(strings.ToLower(match), "telemetryreporter") {
			return TelemetryRiskCritical
		}
		return TelemetryRiskHigh
	case "assignments":
		if strings.Contains(strings.ToLower(match), "machineid") {
			return TelemetryRiskHigh
		}
		return TelemetryRiskMedium
	case "imports":
		return TelemetryRiskHigh
	case "config_access":
		return TelemetryRiskMedium
	default:
		return TelemetryRiskLow
	}
}

// getSurroundingLines gets surrounding lines for context
func (apm *AdvancedPatternMatcher) getSurroundingLines(lines []string, lineNum, radius int) []string {
	var surrounding []string
	
	start := lineNum - radius - 1
	end := lineNum + radius - 1
	
	if start < 0 {
		start = 0
	}
	if end >= len(lines) {
		end = len(lines) - 1
	}
	
	for i := start; i <= end; i++ {
		surrounding = append(surrounding, lines[i])
	}
	
	return surrounding
}

// GetPatternStatistics returns statistics about pattern matching
func (apm *AdvancedPatternMatcher) GetPatternStatistics() map[string]int {
	stats := make(map[string]int)
	
	for context, patterns := range apm.contextPatterns {
		stats[context] = len(patterns)
	}
	
	stats["semantic_patterns"] = len(apm.semanticPatterns)
	stats["combination_rules"] = len(apm.combinationRules)
	stats["exclusion_patterns"] = len(apm.exclusionPatterns)
	
	return stats
}