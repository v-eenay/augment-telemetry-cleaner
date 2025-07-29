package browser

import (
	"fmt"
	"os/exec"
	"runtime"
	"time"
)

// ProcessManager handles browser process management
type ProcessManager struct{}

// NewProcessManager creates a new process manager
func NewProcessManager() *ProcessManager {
	return &ProcessManager{}
}

// ForceCloseBrowser attempts to forcefully close all browser processes
func (pm *ProcessManager) ForceCloseBrowser(browserType BrowserType) error {
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

	return pm.terminateProcesses(processNames)
}

// terminateProcesses terminates the specified processes
func (pm *ProcessManager) terminateProcesses(processNames []string) error {
	if len(processNames) == 0 {
		return nil
	}

	switch runtime.GOOS {
	case "windows":
		return pm.terminateWindowsProcesses(processNames)
	case "darwin":
		return pm.terminateMacProcesses(processNames)
	case "linux":
		return pm.terminateLinuxProcesses(processNames)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// terminateWindowsProcesses terminates processes on Windows
func (pm *ProcessManager) terminateWindowsProcesses(processNames []string) error {
	for _, name := range processNames {
		cmd := exec.Command("taskkill", "/F", "/IM", name)
		cmd.Run() // Ignore errors as process might not be running
	}
	
	// Wait a moment for processes to terminate
	time.Sleep(2 * time.Second)
	
	return nil
}

// terminateMacProcesses terminates processes on macOS
func (pm *ProcessManager) terminateMacProcesses(processNames []string) error {
	for _, name := range processNames {
		// Try graceful termination first
		cmd := exec.Command("pkill", "-f", name)
		cmd.Run()
		
		// Wait a moment
		time.Sleep(1 * time.Second)
		
		// Force kill if still running
		cmd = exec.Command("pkill", "-9", "-f", name)
		cmd.Run()
	}
	
	// Wait for processes to terminate
	time.Sleep(2 * time.Second)
	
	return nil
}

// terminateLinuxProcesses terminates processes on Linux
func (pm *ProcessManager) terminateLinuxProcesses(processNames []string) error {
	for _, name := range processNames {
		// Try graceful termination first
		cmd := exec.Command("pkill", "-f", name)
		cmd.Run()
		
		// Wait a moment
		time.Sleep(1 * time.Second)
		
		// Force kill if still running
		cmd = exec.Command("pkill", "-9", "-f", name)
		cmd.Run()
	}
	
	// Wait for processes to terminate
	time.Sleep(2 * time.Second)
	
	return nil
}

// WaitForProcessesToClose waits for browser processes to close with timeout
func (pm *ProcessManager) WaitForProcessesToClose(browserType BrowserType, timeout time.Duration) error {
	detector, err := NewBrowserDetector()
	if err != nil {
		return fmt.Errorf("failed to create browser detector: %w", err)
	}

	start := time.Now()
	for time.Since(start) < timeout {
		isRunning, err := detector.IsProcessRunning(browserType)
		if err != nil {
			return fmt.Errorf("failed to check if browser is running: %w", err)
		}
		
		if !isRunning {
			return nil // Processes have closed
		}
		
		time.Sleep(500 * time.Millisecond)
	}
	
	return fmt.Errorf("timeout waiting for %s processes to close", browserType.String())
}