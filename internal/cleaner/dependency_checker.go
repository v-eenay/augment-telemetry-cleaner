package cleaner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"augment-telemetry-cleaner/internal/utils"
)

// DependencyChecker handles checking extension dependencies
type DependencyChecker struct {
	extensionRegistry map[string]*ExtensionInfo
	dependencyGraph   map[string][]string
}

// ExtensionInfo represents information about an installed extension
type ExtensionInfo struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Publisher       string            `json:"publisher"`
	Dependencies    []string          `json:"dependencies"`
	ExtensionDependencies []string    `json:"extension_dependencies"`
	IsActive        bool              `json:"is_active"`
	InstallPath     string            `json:"install_path"`
	Manifest        map[string]interface{} `json:"manifest"`
}

// DependencyInfo represents dependency relationship information
type DependencyInfo struct {
	DependentExtension string   `json:"dependent_extension"`
	DependencyType     string   `json:"dependency_type"`
	Required           bool     `json:"required"`
	Description        string   `json:"description"`
	Impact             string   `json:"impact"`
}

// NewDependencyChecker creates a new dependency checker
func NewDependencyChecker() *DependencyChecker {
	return &DependencyChecker{
		extensionRegistry: make(map[string]*ExtensionInfo),
		dependencyGraph:   make(map[string][]string),
	}
}

// CheckDependencies checks what extensions depend on the given extension
func (dc *DependencyChecker) CheckDependencies(extensionID string) ([]DependencyInfo, error) {
	// Load extension registry if not already loaded
	if len(dc.extensionRegistry) == 0 {
		if err := dc.loadExtensionRegistry(); err != nil {
			return nil, fmt.Errorf("failed to load extension registry: %w", err)
		}
	}

	var dependencies []DependencyInfo

	// Check direct dependencies
	for _, ext := range dc.extensionRegistry {
		// Check extension dependencies
		for _, dep := range ext.ExtensionDependencies {
			if strings.EqualFold(dep, extensionID) {
				dependencies = append(dependencies, DependencyInfo{
					DependentExtension: ext.ID,
					DependencyType:     "extension",
					Required:           true,
					Description:        fmt.Sprintf("%s depends on %s", ext.Name, extensionID),
					Impact:             "Extension may not function properly without this dependency",
				})
			}
		}

		// Check if extension is referenced in configuration or settings
		if dc.hasConfigurationDependency(ext, extensionID) {
			dependencies = append(dependencies, DependencyInfo{
				DependentExtension: ext.ID,
				DependencyType:     "configuration",
				Required:           false,
				Description:        fmt.Sprintf("%s has configuration references to %s", ext.Name, extensionID),
				Impact:             "Some configuration settings may be affected",
			})
		}
	}

	// Check for shared data dependencies
	sharedDataDeps := dc.checkSharedDataDependencies(extensionID)
	dependencies = append(dependencies, sharedDataDeps...)

	return dependencies, nil
}

// loadExtensionRegistry loads information about all installed extensions
func (dc *DependencyChecker) loadExtensionRegistry() error {
	extensionsPath, err := utils.GetExtensionsPath()
	if err != nil {
		return fmt.Errorf("failed to get extensions path: %w", err)
	}

	if _, err := os.Stat(extensionsPath); os.IsNotExist(err) {
		return nil // No extensions directory
	}

	entries, err := os.ReadDir(extensionsPath)
	if err != nil {
		return fmt.Errorf("failed to read extensions directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		extensionPath := filepath.Join(extensionsPath, entry.Name())
		manifestPath := filepath.Join(extensionPath, "package.json")

		// Check if package.json exists
		if _, err := os.Stat(manifestPath); err != nil {
			continue
		}

		// Load extension info
		extInfo, err := dc.loadExtensionInfo(extensionPath, manifestPath)
		if err != nil {
			continue // Skip extensions we can't load
		}

		dc.extensionRegistry[extInfo.ID] = extInfo
		
		// Build dependency graph
		for _, dep := range extInfo.ExtensionDependencies {
			dc.dependencyGraph[dep] = append(dc.dependencyGraph[dep], extInfo.ID)
		}
	}

	return nil
}

// loadExtensionInfo loads information about a specific extension
func (dc *DependencyChecker) loadExtensionInfo(extensionPath, manifestPath string) (*ExtensionInfo, error) {
	// Read package.json
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest map[string]interface{}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Extract basic information
	extInfo := &ExtensionInfo{
		InstallPath: extensionPath,
		Manifest:    manifest,
	}

	// Extract ID
	if name, ok := manifest["name"].(string); ok {
		if publisher, ok := manifest["publisher"].(string); ok {
			extInfo.ID = fmt.Sprintf("%s.%s", publisher, name)
			extInfo.Name = name
			extInfo.Publisher = publisher
		}
	}

	// Extract version
	if version, ok := manifest["version"].(string); ok {
		extInfo.Version = version
	}

	// Extract dependencies
	if deps, ok := manifest["dependencies"].(map[string]interface{}); ok {
		for dep := range deps {
			extInfo.Dependencies = append(extInfo.Dependencies, dep)
		}
	}

	// Extract extension dependencies
	if extDeps, ok := manifest["extensionDependencies"].([]interface{}); ok {
		for _, dep := range extDeps {
			if depStr, ok := dep.(string); ok {
				extInfo.ExtensionDependencies = append(extInfo.ExtensionDependencies, depStr)
			}
		}
	}

	// Check if extension is active (simplified check)
	extInfo.IsActive = dc.isExtensionActive(extInfo.ID)

	return extInfo, nil
}

// hasConfigurationDependency checks if an extension has configuration dependencies
func (dc *DependencyChecker) hasConfigurationDependency(ext *ExtensionInfo, targetExtensionID string) bool {
	// Check if the extension's configuration mentions the target extension
	if contributes, ok := ext.Manifest["contributes"].(map[string]interface{}); ok {
		// Check configuration contributions
		if config, ok := contributes["configuration"].(map[string]interface{}); ok {
			configStr := fmt.Sprintf("%v", config)
			if strings.Contains(strings.ToLower(configStr), strings.ToLower(targetExtensionID)) {
				return true
			}
		}

		// Check command contributions
		if commands, ok := contributes["commands"].([]interface{}); ok {
			commandsStr := fmt.Sprintf("%v", commands)
			if strings.Contains(strings.ToLower(commandsStr), strings.ToLower(targetExtensionID)) {
				return true
			}
		}
	}

	return false
}

// checkSharedDataDependencies checks for shared data dependencies
func (dc *DependencyChecker) checkSharedDataDependencies(extensionID string) []DependencyInfo {
	var dependencies []DependencyInfo

	// This would integrate with the correlation analyzer from Phase 3
	// to check for shared data between extensions
	
	// For now, we'll add some common shared data patterns
	sharedDataPatterns := map[string]string{
		"machineId":     "Machine identification data",
		"sessionId":     "Session identification data",
		"telemetryData": "Telemetry collection data",
		"analyticsData": "Analytics data",
	}

	for pattern, description := range sharedDataPatterns {
		// Check if other extensions might be using the same data
		affectedExtensions := dc.findExtensionsUsingPattern(pattern, extensionID)
		
		for _, affectedExt := range affectedExtensions {
			dependencies = append(dependencies, DependencyInfo{
				DependentExtension: affectedExt,
				DependencyType:     "shared_data",
				Required:           false,
				Description:        fmt.Sprintf("Shares %s with %s", description, extensionID),
				Impact:             "Shared data may be affected if removed",
			})
		}
	}

	return dependencies
}

// findExtensionsUsingPattern finds extensions that might be using a specific data pattern
func (dc *DependencyChecker) findExtensionsUsingPattern(pattern, excludeExtensionID string) []string {
	var extensions []string

	// This is a simplified implementation
	// In practice, this would integrate with the storage analyzer
	// to check actual storage data for shared patterns

	for extID := range dc.extensionRegistry {
		if extID != excludeExtensionID {
			// Check if extension name or ID suggests it might use this pattern
			lowerExtID := strings.ToLower(extID)
			lowerPattern := strings.ToLower(pattern)
			
			if strings.Contains(lowerExtID, "telemetry") && strings.Contains(lowerPattern, "telemetry") {
				extensions = append(extensions, extID)
			}
			if strings.Contains(lowerExtID, "analytics") && strings.Contains(lowerPattern, "analytics") {
				extensions = append(extensions, extID)
			}
		}
	}

	return extensions
}

// isExtensionActive checks if an extension is currently active
func (dc *DependencyChecker) isExtensionActive(extensionID string) bool {
	// This would integrate with VS Code's extension management API
	// For now, we'll use a simple heuristic based on common active extensions
	
	activePatterns := []string{
		"vscode.git",
		"vscode.typescript-language-features",
		"vscode.json-language-features",
		"vscode.html-language-features",
		"vscode.css-language-features",
	}

	lowerExtID := strings.ToLower(extensionID)
	for _, pattern := range activePatterns {
		if strings.Contains(lowerExtID, strings.ToLower(pattern)) {
			return true
		}
	}

	return false
}

// GetExtensionInfo returns information about a specific extension
func (dc *DependencyChecker) GetExtensionInfo(extensionID string) (*ExtensionInfo, error) {
	if len(dc.extensionRegistry) == 0 {
		if err := dc.loadExtensionRegistry(); err != nil {
			return nil, fmt.Errorf("failed to load extension registry: %w", err)
		}
	}

	if info, exists := dc.extensionRegistry[extensionID]; exists {
		return info, nil
	}

	return nil, fmt.Errorf("extension not found: %s", extensionID)
}

// GetAllExtensions returns information about all installed extensions
func (dc *DependencyChecker) GetAllExtensions() (map[string]*ExtensionInfo, error) {
	if len(dc.extensionRegistry) == 0 {
		if err := dc.loadExtensionRegistry(); err != nil {
			return nil, fmt.Errorf("failed to load extension registry: %w", err)
		}
	}

	return dc.extensionRegistry, nil
}

// GetDependencyGraph returns the complete dependency graph
func (dc *DependencyChecker) GetDependencyGraph() map[string][]string {
	return dc.dependencyGraph
}

// ValidateRemovalSafety validates if it's safe to remove an extension's data
func (dc *DependencyChecker) ValidateRemovalSafety(extensionID string) (bool, []string, error) {
	dependencies, err := dc.CheckDependencies(extensionID)
	if err != nil {
		return false, nil, fmt.Errorf("failed to check dependencies: %w", err)
	}

	var warnings []string
	safe := true

	// Check for required dependencies
	for _, dep := range dependencies {
		if dep.Required {
			safe = false
			warnings = append(warnings, 
				fmt.Sprintf("Required by %s: %s", dep.DependentExtension, dep.Description))
		} else {
			warnings = append(warnings, 
				fmt.Sprintf("May affect %s: %s", dep.DependentExtension, dep.Description))
		}
	}

	// Check if extension is currently active
	if extInfo, err := dc.GetExtensionInfo(extensionID); err == nil && extInfo.IsActive {
		warnings = append(warnings, "Extension is currently active")
	}

	return safe, warnings, nil
}