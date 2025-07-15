package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"augment-telemetry-cleaner/internal/utils"
)

// ScanResult represents the result of scanning for Augment-related files
type ScanResult struct {
	VSCodeFiles      []FileInfo `json:"vscode_files"`
	AugmentFiles     []FileInfo `json:"augment_files"`
	ConfigFiles      []FileInfo `json:"config_files"`
	LogFiles         []FileInfo `json:"log_files"`
	TotalFiles       int        `json:"total_files"`
	TotalSize        int64      `json:"total_size_bytes"`
	ScanDuration     time.Duration `json:"scan_duration"`
}

// FileInfo represents information about a discovered file
type FileInfo struct {
	Path         string    `json:"path"`
	Size         int64     `json:"size"`
	ModTime      time.Time `json:"mod_time"`
	Type         string    `json:"type"`
	Description  string    `json:"description"`
	Confidence   float64   `json:"confidence"` // 0.0 to 1.0, how confident we are this is Augment-related
}

// AugmentScanner scans the system for Augment-related files and directories
type AugmentScanner struct {
	// Patterns for detecting Augment-related content
	augmentPatterns []*regexp.Regexp
	pathPatterns    []*regexp.Regexp
}

// NewAugmentScanner creates a new scanner instance
func NewAugmentScanner() *AugmentScanner {
	scanner := &AugmentScanner{}
	scanner.initializePatterns()
	return scanner
}

// initializePatterns sets up regex patterns for detecting Augment-related content
func (s *AugmentScanner) initializePatterns() {
	// Content patterns (case-insensitive)
	contentPatterns := []string{
		`(?i)augment`,
		`(?i)augmentcode`,
		`(?i)augment\.code`,
		`(?i)telemetry\.machineId`,
		`(?i)telemetry\.devDeviceId`,
		`(?i)vscode-augment`,
		`(?i)augment-vscode`,
	}

	// Path patterns
	pathPatterns := []string{
		`(?i).*augment.*`,
		`(?i).*telemetry.*`,
		`(?i).*machine.*id.*`,
		`(?i).*device.*id.*`,
	}

	// Compile content patterns
	for _, pattern := range contentPatterns {
		if regex, err := regexp.Compile(pattern); err == nil {
			s.augmentPatterns = append(s.augmentPatterns, regex)
		}
	}

	// Compile path patterns
	for _, pattern := range pathPatterns {
		if regex, err := regexp.Compile(pattern); err == nil {
			s.pathPatterns = append(s.pathPatterns, regex)
		}
	}
}

// ScanSystem performs a comprehensive scan for Augment-related files
func (s *AugmentScanner) ScanSystem() (*ScanResult, error) {
	startTime := time.Now()
	result := &ScanResult{
		VSCodeFiles:  make([]FileInfo, 0),
		AugmentFiles: make([]FileInfo, 0),
		ConfigFiles:  make([]FileInfo, 0),
		LogFiles:     make([]FileInfo, 0),
	}

	// Scan VS Code directories
	if err := s.scanVSCodeDirectories(result); err != nil {
		return nil, fmt.Errorf("failed to scan VS Code directories: %w", err)
	}

	// Scan common application directories
	if err := s.scanCommonDirectories(result); err != nil {
		return nil, fmt.Errorf("failed to scan common directories: %w", err)
	}

	// Calculate totals
	result.TotalFiles = len(result.VSCodeFiles) + len(result.AugmentFiles) + 
					   len(result.ConfigFiles) + len(result.LogFiles)
	
	for _, files := range [][]FileInfo{result.VSCodeFiles, result.AugmentFiles, result.ConfigFiles, result.LogFiles} {
		for _, file := range files {
			result.TotalSize += file.Size
		}
	}

	result.ScanDuration = time.Since(startTime)
	return result, nil
}

// scanVSCodeDirectories scans VS Code specific directories
func (s *AugmentScanner) scanVSCodeDirectories(result *ScanResult) error {
	// Scan storage.json
	if storagePath, err := utils.GetStoragePath(); err == nil {
		if info := s.analyzeFile(storagePath, "VS Code Storage"); info != nil {
			result.VSCodeFiles = append(result.VSCodeFiles, *info)
		}
	}

	// Scan database
	if dbPath, err := utils.GetDBPath(); err == nil {
		if info := s.analyzeFile(dbPath, "VS Code Database"); info != nil {
			result.VSCodeFiles = append(result.VSCodeFiles, *info)
		}
	}

	// Scan machine ID
	if machineIDPath, err := utils.GetMachineIDPath(); err == nil {
		if info := s.analyzeFile(machineIDPath, "VS Code Machine ID"); info != nil {
			result.VSCodeFiles = append(result.VSCodeFiles, *info)
		}
	}

	// Scan workspace storage
	if workspacePath, err := utils.GetWorkspaceStoragePath(); err == nil {
		s.scanDirectory(workspacePath, result, "VS Code Workspace")
	}

	return nil
}

// scanCommonDirectories scans common directories where Augment files might be found
func (s *AugmentScanner) scanCommonDirectories(result *ScanResult) error {
	// Get common directories to scan
	directories := s.getCommonDirectories()

	for _, dir := range directories {
		if _, err := os.Stat(dir); err == nil {
			s.scanDirectory(dir, result, "System Directory")
		}
	}

	return nil
}

// getCommonDirectories returns a list of common directories to scan
func (s *AugmentScanner) getCommonDirectories() []string {
	directories := make([]string, 0)

	// Add user directories
	if homeDir, err := utils.GetHomeDir(); err == nil {
		directories = append(directories, 
			filepath.Join(homeDir, "Documents"),
			filepath.Join(homeDir, "Downloads"),
			filepath.Join(homeDir, "Desktop"),
		)
	}

	// Add application data directories
	if appDataDir, err := utils.GetAppDataDir(); err == nil {
		directories = append(directories, appDataDir)
	}

	// Add temporary directories
	directories = append(directories, os.TempDir())

	return directories
}

// scanDirectory recursively scans a directory for Augment-related files
func (s *AugmentScanner) scanDirectory(dirPath string, result *ScanResult, category string) {
	filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue scanning despite errors
		}

		if info.IsDir() {
			return nil // Continue into subdirectories
		}

		// Analyze the file
		if fileInfo := s.analyzeFile(path, category); fileInfo != nil {
			// Categorize based on file type and content
			switch {
			case strings.Contains(strings.ToLower(path), "log"):
				result.LogFiles = append(result.LogFiles, *fileInfo)
			case strings.Contains(strings.ToLower(path), "config"):
				result.ConfigFiles = append(result.ConfigFiles, *fileInfo)
			case fileInfo.Confidence > 0.7:
				result.AugmentFiles = append(result.AugmentFiles, *fileInfo)
			default:
				result.VSCodeFiles = append(result.VSCodeFiles, *fileInfo)
			}
		}

		return nil
	})
}

// analyzeFile analyzes a single file to determine if it's Augment-related
func (s *AugmentScanner) analyzeFile(filePath, category string) *FileInfo {
	info, err := os.Stat(filePath)
	if err != nil {
		return nil
	}

	// Check if file path matches any patterns
	pathConfidence := s.calculatePathConfidence(filePath)
	
	// For small files, also check content
	contentConfidence := 0.0
	if info.Size() < 10*1024*1024 { // Only scan files smaller than 10MB
		contentConfidence = s.calculateContentConfidence(filePath)
	}

	// Calculate overall confidence
	confidence := (pathConfidence + contentConfidence) / 2.0

	// Only return files with some confidence of being Augment-related
	if confidence > 0.1 {
		return &FileInfo{
			Path:        filePath,
			Size:        info.Size(),
			ModTime:     info.ModTime(),
			Type:        category,
			Description: s.generateDescription(filePath, confidence),
			Confidence:  confidence,
		}
	}

	return nil
}

// calculatePathConfidence calculates confidence based on file path
func (s *AugmentScanner) calculatePathConfidence(filePath string) float64 {
	confidence := 0.0
	
	for _, pattern := range s.pathPatterns {
		if pattern.MatchString(filePath) {
			confidence += 0.3
		}
	}

	// Boost confidence for known VS Code paths
	if strings.Contains(filePath, "Code") && 
	   (strings.Contains(filePath, "globalStorage") || 
		strings.Contains(filePath, "workspaceStorage") ||
		strings.Contains(filePath, "machineid")) {
		confidence += 0.5
	}

	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// calculateContentConfidence calculates confidence based on file content
func (s *AugmentScanner) calculateContentConfidence(filePath string) float64 {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return 0.0
	}

	confidence := 0.0
	contentStr := string(content)

	for _, pattern := range s.augmentPatterns {
		if pattern.MatchString(contentStr) {
			confidence += 0.2
		}
	}

	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// generateDescription generates a human-readable description of the file
func (s *AugmentScanner) generateDescription(filePath string, confidence float64) string {
	switch {
	case confidence > 0.8:
		return "Highly likely to be Augment-related"
	case confidence > 0.5:
		return "Likely to be Augment-related"
	case confidence > 0.3:
		return "Possibly Augment-related"
	default:
		return "May contain Augment references"
	}
}
