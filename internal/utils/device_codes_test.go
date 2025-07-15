package utils

import (
	"regexp"
	"testing"
)

func TestGenerateMachineID(t *testing.T) {
	// Test that machine ID is generated
	machineID, err := GenerateMachineID()
	if err != nil {
		t.Fatalf("GenerateMachineID() failed: %v", err)
	}

	// Test that machine ID is 64 characters long
	if len(machineID) != 64 {
		t.Errorf("Expected machine ID length 64, got %d", len(machineID))
	}

	// Test that machine ID contains only hex characters
	hexPattern := regexp.MustCompile("^[a-f0-9]+$")
	if !hexPattern.MatchString(machineID) {
		t.Errorf("Machine ID contains non-hex characters: %s", machineID)
	}

	// Test that multiple calls generate different IDs
	machineID2, err := GenerateMachineID()
	if err != nil {
		t.Fatalf("Second GenerateMachineID() failed: %v", err)
	}

	if machineID == machineID2 {
		t.Errorf("GenerateMachineID() generated the same ID twice: %s", machineID)
	}
}

func TestGenerateDeviceID(t *testing.T) {
	// Test that device ID is generated
	deviceID := GenerateDeviceID()

	// Test that device ID is in UUID format
	uuidPattern := regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$")
	if !uuidPattern.MatchString(deviceID) {
		t.Errorf("Device ID is not a valid UUID v4: %s", deviceID)
	}

	// Test that device ID is lowercase
	if deviceID != deviceID {
		t.Errorf("Device ID should be lowercase: %s", deviceID)
	}

	// Test that multiple calls generate different IDs
	deviceID2 := GenerateDeviceID()
	if deviceID == deviceID2 {
		t.Errorf("GenerateDeviceID() generated the same ID twice: %s", deviceID)
	}
}

func BenchmarkGenerateMachineID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateMachineID()
		if err != nil {
			b.Fatalf("GenerateMachineID() failed: %v", err)
		}
	}
}

func BenchmarkGenerateDeviceID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateDeviceID()
	}
}
