@echo off
echo Augment Telemetry Cleaner - Build Script
echo ========================================

echo.
echo Building for Windows...
go build -ldflags="-s -w" -o augment-telemetry-cleaner.exe .
if %ERRORLEVEL% neq 0 (
    echo Build failed!
    pause
    exit /b 1
)

echo.
echo Build completed successfully!
echo Executable: augment-telemetry-cleaner.exe

echo.
echo Running tests...
go test ./...
if %ERRORLEVEL% neq 0 (
    echo Tests failed!
    pause
    exit /b 1
)

echo.
echo All tests passed!

echo.
echo Build Summary:
echo - Executable: augment-telemetry-cleaner.exe
echo - Tests: All passed
echo - Ready to use!

echo.
echo To clean up old Python files, run:
echo   go run cleanup.go

pause
