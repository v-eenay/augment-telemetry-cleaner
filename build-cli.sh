#!/bin/bash

# Build script for Augment Telemetry Cleaner CLI

set -e

echo "Building Augment Telemetry Cleaner CLI..."

# Create build directory
mkdir -p build

# Build for different platforms
echo "Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o build/augment-telemetry-cleaner-cli-windows-amd64.exe ./cmd/cli/

echo "Building for Linux (amd64)..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o build/augment-telemetry-cleaner-cli-linux-amd64 ./cmd/cli/

echo "Building for macOS (amd64)..."
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o build/augment-telemetry-cleaner-cli-darwin-amd64 ./cmd/cli/

echo "Building for macOS (arm64)..."
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o build/augment-telemetry-cleaner-cli-darwin-arm64 ./cmd/cli/

echo "Build completed! Binaries are in the 'build' directory:"
ls -la build/

echo ""
echo "Usage examples:"
echo "  ./build/augment-telemetry-cleaner-cli-linux-amd64 --operation run-all --dry-run"
echo "  ./build/augment-telemetry-cleaner-cli-linux-amd64 --operation clean-database --verbose"
echo "  ./build/augment-telemetry-cleaner-cli-linux-amd64 --help"
