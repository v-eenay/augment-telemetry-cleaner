package browser

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// BrowserType represents different browser types
type BrowserType int

const (
	Chrome BrowserType = iota
	Edge
	Firefox
	Safari
)

// String returns the string representation of the browser type
func (bt BrowserType) String() string {
	switch bt {
	case Chrome:
		return "Google Chrome"
	case Edge:
		return "Microsoft Edge"
	case Firefox:
		return "Mozilla Firefox"
	case Safari:
		return "Safari"
	default:
		return "Unknown"
	}
}

// BrowserProfile represents a browser profile/installation
type BrowserProfile struct {
	Type        BrowserType `json:"type"`
	Name        string      `json:"name"`
	ProfilePath string      `json:"profile_path"`
	DataPath    string      `json:"data_path"`
	IsDefault   bool        `json:"is_default"`
	Version     string      `json:"version,omitempty"`
}

// BrowserDetector handles detection of installed browsers and their profiles
type BrowserDetector struct {
	homeDir string
}

// NewBrowserDetector creates a new browser detector
func NewBrowserDetector() (*BrowserDetector, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}
	
	return &BrowserDetector{
		homeDir: homeDir,
	}, nil
}

// DetectBrowsers detects all installed browsers and their profiles
func (bd *BrowserDetector) DetectBrowsers() ([]BrowserProfile, error) {
	var profiles []BrowserProfile
	
	// Detect Chrome profiles
	chromeProfiles, err := bd.detectChromeProfiles()
	if err == nil {
		profiles = append(profiles, chromeProfiles...)
	}
	
	// Detect Edge profiles
	edgeProfiles, err := bd.detectEdgeProfiles()
	if err == nil {
		profiles = append(profiles, edgeProfiles...)
	}
	
	// Detect Firefox profiles
	firefoxProfiles, err := bd.detectFirefoxProfiles()
	if err == nil {
		profiles = append(profiles, firefoxProfiles...)
	}
	
	// Detect Safari profiles (macOS only)
	if runtime.GOOS == "darwin" {
		safariProfiles, err := bd.detectSafariProfiles()
		if err == nil {
			profiles = append(profiles, safariProfiles...)
		}
	}
	
	return profiles, nil
}

// detectChromeProfiles detects Google Chrome profiles
func (bd *BrowserDetector) detectChromeProfiles() ([]BrowserProfile, error) {
	var profiles []BrowserProfile
	var chromePaths []string
	
	switch runtime.GOOS {
	case "windows":
		chromePaths = []string{
			filepath.Join(bd.homeDir, "AppData", "Local", "Google", "Chrome", "User Data"),
		}
	case "darwin":
		chromePaths = []string{
			filepath.Join(bd.homeDir, "Library", "Application Support", "Google", "Chrome"),
		}
	case "linux":
		chromePaths = []string{
			filepath.Join(bd.homeDir, ".config", "google-chrome"),
			filepath.Join(bd.homeDir, ".config", "chromium"),
		}
	}
	
	for _, basePath := range chromePaths {
		if _, err := os.Stat(basePath); os.IsNotExist(err) {
			continue
		}
		
		// Default profile
		defaultProfile := filepath.Join(basePath, "Default")
		if _, err := os.Stat(defaultProfile); err == nil {
			profiles = append(profiles, BrowserProfile{
				Type:        Chrome,
				Name:        "Chrome - Default",
				ProfilePath: defaultProfile,
				DataPath:    basePath,
				IsDefault:   true,
			})
		}
		
		// Additional profiles
		entries, err := os.ReadDir(basePath)
		if err != nil {
			continue
		}
		
		for _, entry := range entries {
			if entry.IsDir() && strings.HasPrefix(entry.Name(), "Profile ") {
				profilePath := filepath.Join(basePath, entry.Name())
				profiles = append(profiles, BrowserProfile{
					Type:        Chrome,
					Name:        fmt.Sprintf("Chrome - %s", entry.Name()),
					ProfilePath: profilePath,
					DataPath:    basePath,
					IsDefault:   false,
				})
			}
		}
	}
	
	return profiles, nil
}

// detectEdgeProfiles detects Microsoft Edge profiles
func (bd *BrowserDetector) detectEdgeProfiles() ([]BrowserProfile, error) {
	var profiles []BrowserProfile
	var edgePaths []string
	
	switch runtime.GOOS {
	case "windows":
		edgePaths = []string{
			filepath.Join(bd.homeDir, "AppData", "Local", "Microsoft", "Edge", "User Data"),
		}
	case "darwin":
		edgePaths = []string{
			filepath.Join(bd.homeDir, "Library", "Application Support", "Microsoft Edge"),
		}
	case "linux":
		edgePaths = []string{
			filepath.Join(bd.homeDir, ".config", "microsoft-edge"),
		}
	}
	
	for _, basePath := range edgePaths {
		if _, err := os.Stat(basePath); os.IsNotExist(err) {
			continue
		}
		
		// Default profile
		defaultProfile := filepath.Join(basePath, "Default")
		if _, err := os.Stat(defaultProfile); err == nil {
			profiles = append(profiles, BrowserProfile{
				Type:        Edge,
				Name:        "Edge - Default",
				ProfilePath: defaultProfile,
				DataPath:    basePath,
				IsDefault:   true,
			})
		}
		
		// Additional profiles
		entries, err := os.ReadDir(basePath)
		if err != nil {
			continue
		}
		
		for _, entry := range entries {
			if entry.IsDir() && strings.HasPrefix(entry.Name(), "Profile ") {
				profilePath := filepath.Join(basePath, entry.Name())
				profiles = append(profiles, BrowserProfile{
					Type:        Edge,
					Name:        fmt.Sprintf("Edge - %s", entry.Name()),
					ProfilePath: profilePath,
					DataPath:    basePath,
					IsDefault:   false,
				})
			}
		}
	}
	
	return profiles, nil
}

// detectFirefoxProfiles detects Mozilla Firefox profiles
func (bd *BrowserDetector) detectFirefoxProfiles() ([]BrowserProfile, error) {
	var profiles []BrowserProfile
	var firefoxPath string
	
	switch runtime.GOOS {
	case "windows":
		firefoxPath = filepath.Join(bd.homeDir, "AppData", "Roaming", "Mozilla", "Firefox")
	case "darwin":
		firefoxPath = filepath.Join(bd.homeDir, "Library", "Application Support", "Firefox")
	case "linux":
		firefoxPath = filepath.Join(bd.homeDir, ".mozilla", "firefox")
	}
	
	if _, err := os.Stat(firefoxPath); os.IsNotExist(err) {
		return profiles, nil
	}
	
	// Read profiles.ini
	profilesIni := filepath.Join(firefoxPath, "profiles.ini")
	if _, err := os.Stat(profilesIni); err != nil {
		return profiles, nil
	}
	
	// Parse profiles.ini to find profile directories
	content, err := os.ReadFile(profilesIni)
	if err != nil {
		return profiles, nil
	}
	
	lines := strings.Split(string(content), "\n")
	var currentProfile map[string]string
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if strings.HasPrefix(line, "[Profile") {
			if currentProfile != nil {
				// Process previous profile
				if path, ok := currentProfile["Path"]; ok {
					isDefault := currentProfile["Default"] == "1"
					name := currentProfile["Name"]
					if name == "" {
						name = "Firefox Profile"
					}
					
					profilePath := filepath.Join(firefoxPath, path)
					if _, err := os.Stat(profilePath); err == nil {
						profiles = append(profiles, BrowserProfile{
							Type:        Firefox,
							Name:        fmt.Sprintf("Firefox - %s", name),
							ProfilePath: profilePath,
							DataPath:    firefoxPath,
							IsDefault:   isDefault,
						})
					}
				}
			}
			currentProfile = make(map[string]string)
		} else if strings.Contains(line, "=") && currentProfile != nil {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				currentProfile[parts[0]] = parts[1]
			}
		}
	}
	
	// Process last profile
	if currentProfile != nil {
		if path, ok := currentProfile["Path"]; ok {
			isDefault := currentProfile["Default"] == "1"
			name := currentProfile["Name"]
			if name == "" {
				name = "Firefox Profile"
			}
			
			profilePath := filepath.Join(firefoxPath, path)
			if _, err := os.Stat(profilePath); err == nil {
				profiles = append(profiles, BrowserProfile{
					Type:        Firefox,
					Name:        fmt.Sprintf("Firefox - %s", name),
					ProfilePath: profilePath,
					DataPath:    firefoxPath,
					IsDefault:   isDefault,
				})
			}
		}
	}
	
	return profiles, nil
}

// detectSafariProfiles detects Safari profiles (macOS only)
func (bd *BrowserDetector) detectSafariProfiles() ([]BrowserProfile, error) {
	var profiles []BrowserProfile
	
	safariPath := filepath.Join(bd.homeDir, "Library", "Safari")
	if _, err := os.Stat(safariPath); os.IsNotExist(err) {
		return profiles, nil
	}
	
	profiles = append(profiles, BrowserProfile{
		Type:        Safari,
		Name:        "Safari - Default",
		ProfilePath: safariPath,
		DataPath:    safariPath,
		IsDefault:   true,
	})
	
	return profiles, nil
}

// IsProcessRunning checks if a browser process is currently running
func (bd *BrowserDetector) IsProcessRunning(browserType BrowserType) (bool, error) {
	var processNames []string

	switch browserType {
	case Chrome:
		switch runtime.GOOS {
		case "windows":
			processNames = []string{"chrome.exe", "chrome_proxy.exe", "chrome_crashpad_handler.exe"}
		case "darwin":
			processNames = []string{"Google Chrome", "Google Chrome Helper", "chrome"}
		case "linux":
			processNames = []string{"chrome", "chromium", "google-chrome", "chrome-sandbox"}
		}
	case Edge:
		switch runtime.GOOS {
		case "windows":
			processNames = []string{"msedge.exe", "msedge_proxy.exe", "msedgewebview2.exe"}
		case "darwin":
			processNames = []string{"Microsoft Edge", "Microsoft Edge Helper"}
		case "linux":
			processNames = []string{"microsoft-edge", "msedge"}
		}
	case Firefox:
		switch runtime.GOOS {
		case "windows":
			processNames = []string{"firefox.exe", "plugin-container.exe", "crashreporter.exe"}
		case "darwin":
			processNames = []string{"Firefox", "firefox", "plugin-container"}
		case "linux":
			processNames = []string{"firefox", "firefox-bin", "plugin-container"}
		}
	case Safari:
		if runtime.GOOS == "darwin" {
			processNames = []string{"Safari", "com.apple.WebKit.WebContent", "SafariForWebKitDevelopment"}
		}
	}

	return bd.checkProcesses(processNames)
}

// checkProcesses checks if any of the given process names are running
func (bd *BrowserDetector) checkProcesses(processNames []string) (bool, error) {
	if len(processNames) == 0 {
		return false, nil
	}

	switch runtime.GOOS {
	case "windows":
		return bd.checkWindowsProcesses(processNames)
	case "darwin":
		return bd.checkMacProcesses(processNames)
	case "linux":
		return bd.checkLinuxProcesses(processNames)
	default:
		return false, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// checkWindowsProcesses checks if processes are running on Windows
func (bd *BrowserDetector) checkWindowsProcesses(processNames []string) (bool, error) {
	cmd := exec.Command("tasklist", "/fo", "csv", "/nh")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to execute tasklist: %w", err)
	}

	outputStr := strings.ToLower(string(output))
	for _, name := range processNames {
		if strings.Contains(outputStr, strings.ToLower(name)) {
			return true, nil
		}
	}

	return false, nil
}

// checkMacProcesses checks if processes are running on macOS
func (bd *BrowserDetector) checkMacProcesses(processNames []string) (bool, error) {
	cmd := exec.Command("ps", "-A")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to execute ps: %w", err)
	}

	outputStr := strings.ToLower(string(output))
	for _, name := range processNames {
		if strings.Contains(outputStr, strings.ToLower(name)) {
			return true, nil
		}
	}

	return false, nil
}

// checkLinuxProcesses checks if processes are running on Linux
func (bd *BrowserDetector) checkLinuxProcesses(processNames []string) (bool, error) {
	cmd := exec.Command("ps", "-A")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to execute ps: %w", err)
	}

	outputStr := strings.ToLower(string(output))
	for _, name := range processNames {
		if strings.Contains(outputStr, strings.ToLower(name)) {
			return true, nil
		}
	}

	return false, nil
}
