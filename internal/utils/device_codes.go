package utils

import (
	"crypto/rand"
	"encoding/hex"
	"strings"

	"github.com/google/uuid"
)

// GenerateMachineID generates a random 64-character hex string for machine ID
// Similar to using /dev/urandom but using Go's cryptographic functions
func GenerateMachineID() (string, error) {
	// Generate 32 random bytes (which will become 64 hex characters)
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	
	// Convert to hexadecimal string
	return hex.EncodeToString(randomBytes), nil
}

// GenerateDeviceID generates a random UUID v4 for device ID
// Returns a lowercase UUID v4 string in the format: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx
// where x is any hexadecimal digit and y is one of 8, 9, A, or B
func GenerateDeviceID() string {
	// Generate a random UUID v4
	deviceID := uuid.New()
	return strings.ToLower(deviceID.String())
}
