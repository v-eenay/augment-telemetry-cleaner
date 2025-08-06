@echo off

REM Build script for Augment Telemetry Cleaner CLI

echo Building Augment Telemetry Cleaner CLI...

REM Create build directory
if not exist build mkdir build

REM Build for different platforms
echo Building for Windows (amd64)...
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w" -o build\augment-telemetry-cleaner-cli-windows-amd64.exe ./cmd/cli/

echo Building for Linux (amd64)...
set GOOS=linux
set GOARCH=amd64
go build -ldflags="-s -w" -o build\augment-telemetry-cleaner-cli-linux-amd64 ./cmd/cli/

echo Building for macOS (amd64)...
set GOOS=darwin
set GOARCH=amd64
go build -ldflags="-s -w" -o build\augment-telemetry-cleaner-cli-darwin-amd64 ./cmd/cli/

echo Building for macOS (arm64)...
set GOOS=darwin
set GOARCH=arm64
go build -ldflags="-s -w" -o build\augment-telemetry-cleaner-cli-darwin-arm64 ./cmd/cli/

echo Build completed! Binaries are in the 'build' directory:
dir build

echo.
echo Usage examples:
echo   build\augment-telemetry-cleaner-cli-windows-amd64.exe --operation run-all --dry-run
echo   build\augment-telemetry-cleaner-cli-windows-amd64.exe --operation clean-database --verbose
echo   build\augment-telemetry-cleaner-cli-windows-amd64.exe --help

pause
