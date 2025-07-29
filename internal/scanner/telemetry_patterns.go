package scanner

import (
	"regexp"
	"strings"
)

// TelemetryPatternDefinition represents a telemetry pattern with metadata
type TelemetryPatternDefinition struct {
	Name        string        `json:"name"`
	Pattern     string        `json:"pattern"`
	Risk        TelemetryRisk `json:"risk"`
	Category    string        `json:"category"`
	Description string        `json:"description"`
	Examples    []string      `json:"examples"`
	Regex       *regexp.Regexp `json:"-"`
}

// TelemetryPatternManager manages all telemetry detection patterns
type TelemetryPatternManager struct {
	patterns map[string]*TelemetryPatternDefinition
	compiled bool
}

// NewTelemetryPatternManager creates a new pattern manager
func NewTelemetryPatternManager() *TelemetryPatternManager {
	manager := &TelemetryPatternManager{
		patterns: make(map[string]*TelemetryPatternDefinition),
	}
	manager.initializePatterns()
	manager.compilePatterns()
	return manager
}

// initializePatterns sets up all telemetry detection patterns
func (tpm *TelemetryPatternManager) initializePatterns() {
	// Critical Risk Patterns - Direct telemetry implementation
	tpm.addPattern("telemetry_reporter_new", TelemetryRiskCritical, "Direct Telemetry",
		`new\s+TelemetryReporter\s*\(`,
		"Creates a new TelemetryReporter instance for data collection",
		[]string{
			"new TelemetryReporter(extensionId, extensionVersion, key)",
			"const reporter = new TelemetryReporter(...)",
		})

	tpm.addPattern("telemetry_reporter_import", TelemetryRiskCritical, "Direct Telemetry",
		`(?:import|require)\s*.*(?:@vscode/extension-telemetry|vscode-extension-telemetry)`,
		"Imports VS Code telemetry library",
		[]string{
			`import TelemetryReporter from '@vscode/extension-telemetry'`,
			`const TelemetryReporter = require('vscode-extension-telemetry')`,
		})

	tpm.addPattern("telemetry_send_event", TelemetryRiskCritical, "Direct Telemetry",
		`\.sendTelemetryEvent\s*\(`,
		"Actively sends telemetry events",
		[]string{
			"reporter.sendTelemetryEvent('eventName', properties)",
			"this.telemetryReporter.sendTelemetryEvent(...)",
		})

	tpm.addPattern("telemetry_send_exception", TelemetryRiskCritical, "Direct Telemetry",
		`\.sendTelemetryException\s*\(`,
		"Sends exception/error telemetry",
		[]string{
			"reporter.sendTelemetryException(error, properties)",
		})

	// High Risk Patterns - Machine/User identification
	tpm.addPattern("vscode_machine_id", TelemetryRiskHigh, "Machine Identification",
		`vscode\.env\.machineId`,
		"Accesses VS Code's unique machine identifier",
		[]string{
			"const machineId = vscode.env.machineId",
			"properties.machineId = vscode.env.machineId",
		})

	tpm.addPattern("vscode_session_id", TelemetryRiskHigh, "Session Identification",
		`vscode\.env\.sessionId`,
		"Accesses VS Code's session identifier",
		[]string{
			"const sessionId = vscode.env.sessionId",
		})

	tpm.addPattern("vscode_remote_name", TelemetryRiskHigh, "Environment Identification",
		`vscode\.env\.remoteName`,
		"Identifies remote development environment",
		[]string{
			"const remoteName = vscode.env.remoteName",
		})

	tpm.addPattern("os_hostname", TelemetryRiskHigh, "System Identification",
		`os\.hostname\s*\(\)`,
		"Gets system hostname for identification",
		[]string{
			"const hostname = os.hostname()",
			"properties.hostname = os.hostname()",
		})

	tpm.addPattern("process_env_user", TelemetryRiskHigh, "User Identification",
		`process\.env\.(?:USER|USERNAME|COMPUTERNAME)`,
		"Accesses system user/computer name",
		[]string{
			"process.env.USER",
			"process.env.USERNAME",
			"process.env.COMPUTERNAME",
		})

	// Medium Risk Patterns - Network communication
	tpm.addPattern("fetch_request", TelemetryRiskMedium, "Network Communication",
		`fetch\s*\(`,
		"Makes HTTP requests that could send data",
		[]string{
			"fetch('https://api.example.com/telemetry', options)",
			"await fetch(url, { method: 'POST', body: data })",
		})

	tpm.addPattern("axios_request", TelemetryRiskMedium, "Network Communication",
		`axios\s*\.(?:get|post|put|delete|request)`,
		"Makes HTTP requests using Axios library",
		[]string{
			"axios.post('https://analytics.com', data)",
			"axios.get(telemetryEndpoint)",
		})

	tpm.addPattern("http_request", TelemetryRiskMedium, "Network Communication",
		`https?\.request\s*\(`,
		"Makes HTTP requests using Node.js http module",
		[]string{
			"http.request(options, callback)",
			"https.request(url, options)",
		})

	tpm.addPattern("xmlhttprequest", TelemetryRiskMedium, "Network Communication",
		`new\s+XMLHttpRequest\s*\(\)`,
		"Creates XMLHttpRequest for web requests",
		[]string{
			"const xhr = new XMLHttpRequest()",
		})

	tpm.addPattern("navigator_useragent", TelemetryRiskMedium, "Browser Fingerprinting",
		`navigator\.userAgent`,
		"Accesses browser user agent string",
		[]string{
			"const userAgent = navigator.userAgent",
		})

	// Application Insights patterns
	tpm.addPattern("appinsights_import", TelemetryRiskHigh, "Analytics Service",
		`(?:import|require)\s*.*applicationinsights`,
		"Imports Microsoft Application Insights",
		[]string{
			`import * as appInsights from 'applicationinsights'`,
			`const appInsights = require('applicationinsights')`,
		})

	tpm.addPattern("appinsights_track", TelemetryRiskHigh, "Analytics Service",
		`\.track(?:Event|Exception|Metric|Request|Dependency)\s*\(`,
		"Tracks events using Application Insights",
		[]string{
			"client.trackEvent({ name: 'eventName' })",
			"appInsights.defaultClient.trackException({ exception: error })",
		})

	// Low Risk Patterns - General analytics references
	tpm.addPattern("analytics_reference", TelemetryRiskLow, "Analytics Reference",
		`(?:analytics|tracking|metrics|usage)\s*[:=]`,
		"References to analytics or tracking functionality",
		[]string{
			"const analytics = require('./analytics')",
			"tracking: true",
		})

	tpm.addPattern("performance_now", TelemetryRiskLow, "Performance Tracking",
		`performance\.now\s*\(\)`,
		"Measures performance timing",
		[]string{
			"const start = performance.now()",
		})

	tpm.addPattern("console_log_data", TelemetryRiskLow, "Data Logging",
		`console\.(?:log|info|warn|error)\s*\([^)]*(?:user|data|info|event)`,
		"Logs potentially sensitive data to console",
		[]string{
			"console.log('User data:', userData)",
			"console.info('Event:', eventData)",
		})

	// Configuration and storage patterns
	tpm.addPattern("localstorage_access", TelemetryRiskMedium, "Local Storage",
		`localStorage\.(?:getItem|setItem|removeItem)`,
		"Accesses browser local storage",
		[]string{
			"localStorage.setItem('telemetry', data)",
			"const stored = localStorage.getItem('analytics')",
		})

	tpm.addPattern("sessionstorage_access", TelemetryRiskMedium, "Session Storage",
		`sessionStorage\.(?:getItem|setItem|removeItem)`,
		"Accesses browser session storage",
		[]string{
			"sessionStorage.setItem('session', data)",
		})

	tpm.addPattern("document_cookie", TelemetryRiskMedium, "Cookie Access",
		`document\.cookie`,
		"Accesses browser cookies",
		[]string{
			"document.cookie = 'tracking=enabled'",
			"const cookies = document.cookie",
		})

	// Extension-specific patterns
	tpm.addPattern("vscode_workspace_config", TelemetryRiskLow, "Configuration Access",
		`vscode\.workspace\.getConfiguration\s*\([^)]*(?:telemetry|analytics|tracking)`,
		"Accesses telemetry-related configuration",
		[]string{
			"vscode.workspace.getConfiguration('telemetry')",
			"getConfiguration('myext.analytics')",
		})

	tpm.addPattern("extension_context_global", TelemetryRiskMedium, "Extension Storage",
		`context\.globalState\.(?:get|update)`,
		"Accesses extension global state storage",
		[]string{
			"context.globalState.update('telemetryData', data)",
			"const stored = context.globalState.get('analytics')",
		})

	tpm.addPattern("extension_context_workspace", TelemetryRiskLow, "Extension Storage",
		`context\.workspaceState\.(?:get|update)`,
		"Accesses extension workspace state storage",
		[]string{
			"context.workspaceState.update('usage', stats)",
		})
}

// addPattern adds a new telemetry pattern definition
func (tpm *TelemetryPatternManager) addPattern(name string, risk TelemetryRisk, category, pattern, description string, examples []string) {
	tpm.patterns[name] = &TelemetryPatternDefinition{
		Name:        name,
		Pattern:     pattern,
		Risk:        risk,
		Category:    category,
		Description: description,
		Examples:    examples,
	}
}

// compilePatterns compiles all regex patterns
func (tpm *TelemetryPatternManager) compilePatterns() {
	for _, pattern := range tpm.patterns {
		if regex, err := regexp.Compile(`(?i)` + pattern.Pattern); err == nil {
			pattern.Regex = regex
		}
	}
	tpm.compiled = true
}

// GetPatterns returns all pattern definitions
func (tpm *TelemetryPatternManager) GetPatterns() map[string]*TelemetryPatternDefinition {
	return tpm.patterns
}

// GetPatternsByRisk returns patterns filtered by risk level
func (tpm *TelemetryPatternManager) GetPatternsByRisk(risk TelemetryRisk) []*TelemetryPatternDefinition {
	var filtered []*TelemetryPatternDefinition
	for _, pattern := range tpm.patterns {
		if pattern.Risk == risk {
			filtered = append(filtered, pattern)
		}
	}
	return filtered
}

// GetPatternsByCategory returns patterns filtered by category
func (tpm *TelemetryPatternManager) GetPatternsByCategory(category string) []*TelemetryPatternDefinition {
	var filtered []*TelemetryPatternDefinition
	for _, pattern := range tpm.patterns {
		if strings.EqualFold(pattern.Category, category) {
			filtered = append(filtered, pattern)
		}
	}
	return filtered
}

// MatchLine checks if a line matches any telemetry patterns
func (tpm *TelemetryPatternManager) MatchLine(line string) []*TelemetryPatternDefinition {
	var matches []*TelemetryPatternDefinition
	
	if !tpm.compiled {
		tpm.compilePatterns()
	}

	for _, pattern := range tpm.patterns {
		if pattern.Regex != nil && pattern.Regex.MatchString(line) {
			matches = append(matches, pattern)
		}
	}

	return matches
}

// GetRiskDescription returns a description for a risk level
func GetRiskDescription(risk TelemetryRisk) string {
	switch risk {
	case TelemetryRiskCritical:
		return "Critical: Actively collects and transmits telemetry data"
	case TelemetryRiskHigh:
		return "High: Accesses machine/user identification or sends data to external services"
	case TelemetryRiskMedium:
		return "Medium: Has network communication capabilities or accesses local storage"
	case TelemetryRiskLow:
		return "Low: Contains references to analytics/tracking but limited data collection"
	case TelemetryRiskNone:
		return "None: No telemetry patterns detected"
	default:
		return "Unknown risk level"
	}
}

// GetRiskColor returns a color code for UI display of risk levels
func GetRiskColor(risk TelemetryRisk) string {
	switch risk {
	case TelemetryRiskCritical:
		return "#FF0000" // Red
	case TelemetryRiskHigh:
		return "#FF6600" // Orange
	case TelemetryRiskMedium:
		return "#FFCC00" // Yellow
	case TelemetryRiskLow:
		return "#00CC00" // Green
	case TelemetryRiskNone:
		return "#CCCCCC" // Gray
	default:
		return "#000000" // Black
	}
}