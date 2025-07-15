package cleaner

import (
	"encoding/json"
	"fmt"
	"os"

	"augment-telemetry-cleaner/internal/utils"
)

// TelemetryModifyResult contains the results of telemetry ID modification
type TelemetryModifyResult struct {
	OldMachineID         string `json:"old_machine_id"`
	NewMachineID         string `json:"new_machine_id"`
	OldDeviceID          string `json:"old_device_id"`
	NewDeviceID          string `json:"new_device_id"`
	StorageBackupPath    string `json:"storage_backup_path"`
	MachineIDBackupPath  string `json:"machine_id_backup_path,omitempty"`
}

// ModifyTelemetryIDs modifies the telemetry IDs in the VS Code storage.json file and machine ID file
// Creates backups before modification
//
// This function:
// 1. Creates backups of the storage.json and machine ID files
// 2. Reads the storage.json file
// 3. Generates new machine and device IDs
// 4. Updates the telemetry.machineId and telemetry.devDeviceId values in storage.json
// 5. Updates the machine ID file with the new machine ID
// 6. Saves the modified files
func ModifyTelemetryIDs() (*TelemetryModifyResult, error) {
	storagePath, err := utils.GetStoragePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get storage path: %w", err)
	}

	machineIDPath, err := utils.GetMachineIDPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get machine ID path: %w", err)
	}

	// Check if storage file exists
	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("storage file not found at: %s", storagePath)
	}

	// Create backup of storage.json
	storageBackupPath, err := utils.CreateBackup(storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage backup: %w", err)
	}

	// Create backup of machine ID file if it exists
	var machineIDBackupPath string
	if _, err := os.Stat(machineIDPath); err == nil {
		machineIDBackupPath, err = utils.CreateBackup(machineIDPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create machine ID backup: %w", err)
		}
	}

	// Read the current JSON content
	data, err := os.ReadFile(storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read storage file: %w", err)
	}

	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Store old values
	oldMachineID, _ := jsonData["telemetry.machineId"].(string)
	oldDeviceID, _ := jsonData["telemetry.devDeviceId"].(string)

	// Generate new IDs
	newMachineID, err := utils.GenerateMachineID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate machine ID: %w", err)
	}

	newDeviceID := utils.GenerateDeviceID()

	// Update the values in storage.json
	jsonData["telemetry.machineId"] = newMachineID
	jsonData["telemetry.devDeviceId"] = newDeviceID

	// Write the modified content back to storage.json
	modifiedData, err := json.MarshalIndent(jsonData, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(storagePath, modifiedData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write storage file: %w", err)
	}

	// Write the new device ID to the machine ID file
	if err := os.WriteFile(machineIDPath, []byte(newDeviceID), 0644); err != nil {
		return nil, fmt.Errorf("failed to write machine ID file: %w", err)
	}

	return &TelemetryModifyResult{
		OldMachineID:        oldMachineID,
		NewMachineID:        newMachineID,
		OldDeviceID:         oldDeviceID,
		NewDeviceID:         newDeviceID,
		StorageBackupPath:   storageBackupPath,
		MachineIDBackupPath: machineIDBackupPath,
	}, nil
}
