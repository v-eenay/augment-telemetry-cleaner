#!/bin/bash

echo "Augment Telemetry Cleaner - Build Script"
echo "========================================"

echo
echo "Building for current platform..."
go build -ldflags="-s -w" -o augment-telemetry-cleaner .
if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi

echo
echo "Build completed successfully!"
echo "Executable: augment-telemetry-cleaner"

echo
echo "Running tests..."
go test ./...
if [ $? -ne 0 ]; then
    echo "Tests failed!"
    exit 1
fi

echo
echo "All tests passed!"

echo
echo "Build Summary:"
echo "- Executable: augment-telemetry-cleaner"
echo "- Tests: All passed"
echo "- Ready to use!"

echo
echo "To clean up old Python files, run:"
echo "  go run cleanup.go"

# Make the script executable
chmod +x augment-telemetry-cleaner
