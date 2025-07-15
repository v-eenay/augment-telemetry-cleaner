package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	// General settings
	DryRunMode          bool   `json:"dry_run_mode"`
	CreateBackups       bool   `json:"create_backups"`
	LogLevel            string `json:"log_level"`
	
	// Paths (can be overridden by user)
	CustomStoragePath      string `json:"custom_storage_path,omitempty"`
	CustomDBPath           string `json:"custom_db_path,omitempty"`
	CustomWorkspacePath    string `json:"custom_workspace_path,omitempty"`
	CustomMachineIDPath    string `json:"custom_machine_id_path,omitempty"`
	
	// Backup settings
	BackupDirectory        string `json:"backup_directory"`
	MaxBackupAge           int    `json:"max_backup_age_days"`
	
	// Safety settings
	RequireConfirmation    bool   `json:"require_confirmation"`
	ShowPreviewBeforeRun   bool   `json:"show_preview_before_run"`
	
	// Advanced settings
	DatabaseTimeout        int    `json:"database_timeout_seconds"`
	FileOperationRetries   int    `json:"file_operation_retries"`
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	return &Config{
		DryRunMode:             true,  // Start in safe mode
		CreateBackups:          true,
		LogLevel:               "INFO",
		BackupDirectory:        "",    // Will be set to user's documents folder
		MaxBackupAge:           30,    // Keep backups for 30 days
		RequireConfirmation:    true,
		ShowPreviewBeforeRun:   true,
		DatabaseTimeout:        30,
		FileOperationRetries:   3,
	}
}

// ConfigManager manages application configuration
type ConfigManager struct {
	configPath string
	config     *Config
}

// NewConfigManager creates a new configuration manager
func NewConfigManager() (*ConfigManager, error) {
	// Get user's config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback to home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user directories: %w", err)
		}
		configDir = filepath.Join(homeDir, ".config")
	}
	
	// Create application config directory
	appConfigDir := filepath.Join(configDir, "augment-telemetry-cleaner")
	if err := os.MkdirAll(appConfigDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}
	
	configPath := filepath.Join(appConfigDir, "config.json")
	
	cm := &ConfigManager{
		configPath: configPath,
		config:     DefaultConfig(),
	}
	
	// Set default backup directory
	if cm.config.BackupDirectory == "" {
		documentsDir, err := getDocumentsDir()
		if err == nil {
			cm.config.BackupDirectory = filepath.Join(documentsDir, "Augment-Telemetry-Backups")
		} else {
			// Fallback to config directory
			cm.config.BackupDirectory = filepath.Join(appConfigDir, "backups")
		}
	}
	
	// Load existing config if it exists
	if err := cm.Load(); err != nil {
		// If loading fails, save the default config
		if saveErr := cm.Save(); saveErr != nil {
			return nil, fmt.Errorf("failed to save default config: %w", saveErr)
		}
	}
	
	return cm, nil
}

// Load loads the configuration from file
func (cm *ConfigManager) Load() error {
	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Config file doesn't exist, use defaults
			return nil
		}
		return fmt.Errorf("failed to read config file: %w", err)
	}
	
	if err := json.Unmarshal(data, cm.config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}
	
	return nil
}

// Save saves the configuration to file
func (cm *ConfigManager) Save() error {
	data, err := json.MarshalIndent(cm.config, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	if err := os.WriteFile(cm.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}

// GetConfig returns the current configuration
func (cm *ConfigManager) GetConfig() *Config {
	return cm.config
}

// UpdateConfig updates the configuration and saves it
func (cm *ConfigManager) UpdateConfig(updater func(*Config)) error {
	updater(cm.config)
	return cm.Save()
}

// GetBackupDirectory returns the backup directory, creating it if necessary
func (cm *ConfigManager) GetBackupDirectory() (string, error) {
	backupDir := cm.config.BackupDirectory
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}
	return backupDir, nil
}

// getDocumentsDir attempts to get the user's documents directory
func getDocumentsDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	
	// Try common documents directory locations
	possiblePaths := []string{
		filepath.Join(homeDir, "Documents"),
		filepath.Join(homeDir, "My Documents"),
		homeDir, // Fallback to home directory
	}
	
	for _, path := range possiblePaths {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			return path, nil
		}
	}
	
	return homeDir, nil
}
