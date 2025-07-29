package scanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"augment-telemetry-cleaner/internal/utils"
)

// ExtensionInfo represents information about a VS Code extension
type ExtensionInfo struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Publisher       string            `json:"publisher"`
	DisplayName     string            `json:"display_name"`
	Description     string            `json:"description"`
	InstallPath     string            `json:"install_path"`
	ManifestPath    string            `json:"manifest_path"`
	HasTelemetry    bool              `json:"has_telemetry"`
	TelemetryRisk   TelemetryRisk     `json:"telemetry_risk"`
	TelemetryTypes  []string          `json:"telemetry_types"`
	Dependencies    []string          `json:"dependencies"`
	ActivationEvents []string         `json:"activation_events"`
	Commands        []string          `json:"commands"`
	StorageSize     int64             `json:"storage_size"`
	LastModified    time.Time         `json:"last_modified"`
	Manifest        *ExtensionManifest `json:"manifest,omitempty"`
}

// ExtensionManifest represents the package.json structure of a VS Code extension
type ExtensionManifest struct {
	Name                string                 `json:"name"`
	DisplayName         string                 `json:"displayName"`
	Description         string                 `json:"description"`
	Version             string                 `json:"version"`
	Publisher           string                 `json:"publisher"`
	Engines             map[string]string      `json:"engines"`
	Categories          []string               `json:"categories"`
	Keywords            []string               `json:"keywords"`
	ActivationEvents    []string               `json:"activationEvents"`
	Main                string                 `json:"main"`
	Contributes         map[string]interface{} `json:"contributes"`
	Scripts             map[string]string      `json:"scripts"`
	Dependencies        map[string]string      `json:"dependencies"`
	DevDependencies     map[string]string      `json:"devDependencies"`
	ExtensionDependencies []string             `json:"extensionDependencies"`
}

// TelemetryRisk represents the risk level of telemetry in an extension
type TelemetryRisk int

const (
	TelemetryRiskNone TelemetryRisk = iota
	TelemetryRiskLow
	TelemetryRiskMedium
	TelemetryRiskHigh
	TelemetryRiskCritical
)

// String returns the string representation of telemetry risk
func (tr TelemetryRisk) String() string {
	switch tr {
	case TelemetryRiskNone:
		return "None"
	case TelemetryRiskLow:
		return "Low"
	case TelemetryRiskMedium:
		return "Medium"
	case TelemetryRiskHigh:
		return "High"
	case TelemetryRiskCritical:
		return "Critical"
	default:
		return "Unknown"
	}
}

// ExtensionScanResult represents the result of scanning for VS Code extensions
type ExtensionScanResult struct {
	Extensions          []ExtensionInfo `json:"extensions"`
	TotalExtensions     int             `json:"total_extensions"`
	TelemetryExtensions int             `json:"telemetry_extensions"`
	HighRiskExtensions  int             `json:"high_risk_extensions"`
	TotalStorageSize    int64           `json:"total_storage_size"`
	ScanDuration        time.Duration   `json:"scan_duration"`
	ExtensionPaths      []string        `json:"extension_paths"`
}

// ExtensionScanner handles scanning for VS Code extensions and their telemetry capabilities
type ExtensionScanner struct {
	telemetryPatterns []string
	riskPatterns      map[TelemetryRisk][]string
}

// NewExtensionScanner creates a new extension scanner
func NewExtensionScanner() *ExtensionScanner {
	scanner := &ExtensionScanner{}
	scanner.initializeTelemetryPatterns()
	return scanner
}

// initializeTelemetryPatterns sets up patterns for detecting telemetry in extensions
func (es *ExtensionScanner) initializeTelemetryPatterns() {
	// Basic telemetry dependency patterns
	es.telemetryPatterns = []string{
		"@vscode/extension-telemetry",
		"vscode-extension-telemetry",
		"applicationinsights",
		"@microsoft/applicationinsights",
		"telemetry",
		"analytics",
		"tracking",
		"metrics",
	}

	// Risk-based patterns for more detailed analysis
	es.riskPatterns = map[TelemetryRisk][]string{
		TelemetryRiskCritical: {
			"@vscode/extension-telemetry",
			"vscode-extension-telemetry",
			"TelemetryReporter",
		},
		TelemetryRiskHigh: {
			"vscode.env.machineId",
			"vscode.env.sessionId",
			"os.hostname",
			"applicationinsights",
		},
		TelemetryRiskMedium: {
			"fetch(",
			"axios.",
			"http.request",
			"https.request",
			"XMLHttpRequest",
		},
		TelemetryRiskLow: {
			"analytics",
			"tracking",
			"metrics",
			"usage",
		},
	}
}

// ScanExtensions performs a comprehensive scan for VS Code extensions
func (es *ExtensionScanner) ScanExtensions() (*ExtensionScanResult, error) {
	startTime := time.Now()
	
	result := &ExtensionScanResult{
		Extensions:     make([]ExtensionInfo, 0),
		ExtensionPaths: make([]string, 0),
	}

	// Get all extension directories
	extensionDirs, err := es.getExtensionDirectories()
	if err != nil {
		return nil, fmt.Errorf("failed to get extension directories: %w", err)
	}

	result.ExtensionPaths = extensionDirs

	// Scan each extension directory
	for _, dir := range extensionDirs {
		if _, err := os.Stat(dir); err != nil {
			continue // Skip directories that don't exist
		}

		extensions, err := es.scanExtensionDirectory(dir)
		if err != nil {
			// Log error but continue with other directories
			continue
		}

		result.Extensions = append(result.Extensions, extensions...)
	}

	// Calculate statistics
	result.TotalExtensions = len(result.Extensions)
	for _, ext := range result.Extensions {
		if ext.HasTelemetry {
			result.TelemetryExtensions++
		}
		if ext.TelemetryRisk >= TelemetryRiskHigh {
			result.HighRiskExtensions++
		}
		result.TotalStorageSize += ext.StorageSize
	}

	result.ScanDuration = time.Since(startTime)
	return result, nil
}

// getExtensionDirectories returns all possible VS Code extension directories
func (es *ExtensionScanner) getExtensionDirectories() ([]string, error) {
	var directories []string

	homeDir, err := utils.GetHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	switch runtime.GOOS {
	case "windows":
		// User extensions
		directories = append(directories, filepath.Join(homeDir, ".vscode", "extensions"))
		
		// System extensions (if accessible)
		programFiles := os.Getenv("PROGRAMFILES")
		if programFiles != "" {
			directories = append(directories, 
				filepath.Join(programFiles, "Microsoft VS Code", "resources", "app", "extensions"))
		}
		
		// Insiders version
		directories = append(directories, filepath.Join(homeDir, ".vscode-insiders", "extensions"))

	case "darwin":
		// User extensions
		directories = append(directories, filepath.Join(homeDir, ".vscode", "extensions"))
		
		// System extensions
		directories = append(directories, "/Applications/Visual Studio Code.app/Contents/Resources/app/extensions")
		
		// Insiders version
		directories = append(directories, filepath.Join(homeDir, ".vscode-insiders", "extensions"))

	default: // Linux and other Unix-like systems
		// User extensions
		directories = append(directories, filepath.Join(homeDir, ".vscode", "extensions"))
		
		// System extensions (common locations)
		directories = append(directories, 
			"/usr/share/code/resources/app/extensions",
			"/opt/visual-studio-code/resources/app/extensions")
		
		// Insiders version
		directories = append(directories, filepath.Join(homeDir, ".vscode-insiders", "extensions"))
	}

	return directories, nil
}

// scanExtensionDirectory scans a specific directory for extensions
func (es *ExtensionScanner) scanExtensionDirectory(dirPath string) ([]ExtensionInfo, error) {
	var extensions []ExtensionInfo

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read extension directory %s: %w", dirPath, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		extensionPath := filepath.Join(dirPath, entry.Name())
		manifestPath := filepath.Join(extensionPath, "package.json")

		// Check if package.json exists
		if _, err := os.Stat(manifestPath); err != nil {
			continue // Skip directories without package.json
		}

		// Parse extension
		extension, err := es.parseExtension(extensionPath, manifestPath)
		if err != nil {
			// Log error but continue with other extensions
			continue
		}

		extensions = append(extensions, *extension)
	}

	return extensions, nil
}

// parseExtension parses a single extension from its directory
func (es *ExtensionScanner) parseExtension(extensionPath, manifestPath string) (*ExtensionInfo, error) {
	// Read and parse package.json
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest ExtensionManifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Get directory info
	dirInfo, err := os.Stat(extensionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get directory info: %w", err)
	}

	// Calculate storage size
	storageSize, err := es.calculateStorageSize(extensionPath)
	if err != nil {
		storageSize = 0 // Continue even if we can't calculate size
	}

	// Create extension info
	extension := &ExtensionInfo{
		ID:               fmt.Sprintf("%s.%s", manifest.Publisher, manifest.Name),
		Name:             manifest.Name,
		Version:          manifest.Version,
		Publisher:        manifest.Publisher,
		DisplayName:      manifest.DisplayName,
		Description:      manifest.Description,
		InstallPath:      extensionPath,
		ManifestPath:     manifestPath,
		ActivationEvents: manifest.ActivationEvents,
		Dependencies:     es.extractDependencies(manifest),
		StorageSize:      storageSize,
		LastModified:     dirInfo.ModTime(),
		Manifest:         &manifest,
	}

	// Extract commands from contributes section
	if contributes, ok := manifest.Contributes["commands"].([]interface{}); ok {
		for _, cmd := range contributes {
			if cmdMap, ok := cmd.(map[string]interface{}); ok {
				if command, ok := cmdMap["command"].(string); ok {
					extension.Commands = append(extension.Commands, command)
				}
			}
		}
	}

	// Analyze telemetry capabilities
	es.analyzeTelemetryCapabilities(extension)

	return extension, nil
}

// extractDependencies extracts all dependencies from the manifest
func (es *ExtensionScanner) extractDependencies(manifest ExtensionManifest) []string {
	var deps []string

	// Add runtime dependencies
	for dep := range manifest.Dependencies {
		deps = append(deps, dep)
	}

	// Add dev dependencies (they might contain telemetry tools)
	for dep := range manifest.DevDependencies {
		deps = append(deps, dep)
	}

	// Add extension dependencies
	deps = append(deps, manifest.ExtensionDependencies...)

	return deps
}

// calculateStorageSize calculates the total size of an extension directory
func (es *ExtensionScanner) calculateStorageSize(extensionPath string) (int64, error) {
	var totalSize int64

	err := filepath.Walk(extensionPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue despite errors
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})

	return totalSize, err
}

// analyzeTelemetryCapabilities analyzes an extension for telemetry capabilities
func (es *ExtensionScanner) analyzeTelemetryCapabilities(extension *ExtensionInfo) {
	var telemetryTypes []string
	maxRisk := TelemetryRiskNone

	// Check dependencies for telemetry packages
	for _, dep := range extension.Dependencies {
		for _, pattern := range es.telemetryPatterns {
			if strings.Contains(strings.ToLower(dep), strings.ToLower(pattern)) {
				telemetryTypes = append(telemetryTypes, fmt.Sprintf("Dependency: %s", dep))
				extension.HasTelemetry = true
				
				// Determine risk level
				for risk, patterns := range es.riskPatterns {
					for _, riskPattern := range patterns {
						if strings.Contains(strings.ToLower(dep), strings.ToLower(riskPattern)) {
							if risk > maxRisk {
								maxRisk = risk
							}
						}
					}
				}
			}
		}
	}

	// Check activation events for suspicious patterns
	for _, event := range extension.ActivationEvents {
		if strings.Contains(event, "*") || strings.Contains(event, "onStartup") {
			telemetryTypes = append(telemetryTypes, fmt.Sprintf("Activation: %s", event))
			if maxRisk < TelemetryRiskLow {
				maxRisk = TelemetryRiskLow
			}
		}
	}

	extension.TelemetryTypes = telemetryTypes
	extension.TelemetryRisk = maxRisk
}