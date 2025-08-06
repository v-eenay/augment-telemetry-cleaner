# Augment Telemetry Cleaner

A comprehensive tool for cleaning Augment telemetry data from VS Code, available in both GUI and CLI versions for different use cases.

## 🚀 Features

### Core Functionality
- 🔧 **Telemetry ID Modification**: Reset device and machine IDs with cryptographically secure random generation
- 🗃️ **Database Cleanup**: Remove Augment-related records from VS Code's SQLite database
- 💾 **Workspace Storage Management**: Clean workspace storage with comprehensive backup creation
- 🔍 **Comprehensive File Scanning**: Advanced detection of Augment-related files across the system

### Safety & Security
- 🛡️ **Dry-Run Mode**: Preview changes before execution
- 💾 **Automatic Backups**: Create timestamped backups before any modifications
- ✅ **Confirmation Dialogs**: Require user confirmation for destructive operations
- 🔐 **Backup Verification**: Verify backup integrity before proceeding
- 📊 **Pre-Operation Checks**: Comprehensive safety checks before running operations

### User Experience
- 🖥️ **Modern GUI**: Clean, intuitive desktop interface built with Fyne
- ⚙️ **Configurable Settings**: Customizable safety settings and preferences
- 📝 **Comprehensive Logging**: Detailed operation logs with multiple severity levels
- 📈 **Progress Tracking**: Real-time progress indicators for long-running operations
- 🎯 **Cross-Platform**: Works on Windows, macOS, and Linux

## 🖥️ Two Versions Available

### GUI Version (Desktop Application)
- **Interactive Interface**: Point-and-click operation with real-time feedback
- **Visual Progress**: Progress bars and status indicators
- **Settings Management**: Configurable preferences and safety settings
- **Best for**: Interactive use, one-time operations, users who prefer graphical interfaces

### CLI Version (Command-Line Tool)
- **Automation Ready**: Perfect for scripts and CI/CD pipelines
- **Batch Operations**: Run multiple operations in sequence
- **JSON Output**: Machine-readable output for integration
- **Dry-Run Support**: Preview operations without making changes
- **Best for**: Automation, scripting, headless environments, power users

See [CLI-README.md](CLI-README.md) for detailed CLI documentation.

## 📋 Requirements

- Go 1.21 or higher
- VS Code (for the files to be cleaned)

## 🛠️ Installation

### Option 1: Download Pre-built Binary
1. Download the latest release from the [Releases](https://github.com/v-eenay/augment-telemetry-cleaner/releases) page
2. Extract the archive
3. Run the executable

### Option 2: Build from Source
1. Clone this repository:
   ```bash
   git clone https://github.com/v-eenay/augment-telemetry-cleaner.git
   cd augment-telemetry-cleaner
   ```

2. Build the application:

   **GUI Version (Desktop Application):**
   ```bash
   # Windows
   go build -o augment-cleaner.exe .

   # Linux/macOS
   go build -o augment-cleaner .
   ```

   **CLI Version (Command-Line Tool):**
   ```bash
   # Windows
   go build -o augment-cleaner-cli.exe ./cmd/cli/

   # Linux/macOS
   go build -o augment-cleaner-cli ./cmd/cli/
   ```

   **Build All Versions (Cross-Platform):**
   ```bash
   # Use the provided build scripts
   ./build-cli.sh        # Linux/macOS
   build-cli.bat         # Windows
   ```

3. Run the application:

   **GUI Version:**
   ```bash
   ./augment-cleaner.exe    # Windows
   ./augment-cleaner        # Linux/macOS
   ```

   **CLI Version:**
   ```bash
   ./augment-cleaner-cli.exe --help    # Windows
   ./augment-cleaner-cli --help        # Linux/macOS
   ```

## 🎯 Usage

### Important: Pre-Operation Steps
1. **Close VS Code completely** - Ensure all VS Code processes are terminated
2. **Exit Augment plugin** - Make sure the Augment extension is not running

### Using the GUI Application
1. Launch the Augment Telemetry Cleaner
2. Configure your preferences in the Settings dialog (optional)
3. Choose your operation mode:
   - **Dry Run Mode** (recommended first): Preview what will be changed
   - **Full Operation**: Actually perform the cleaning operations
4. Select individual operations or use "Run All Operations"
5. Review the results and backup information
6. Restart VS Code when ready

### Using the CLI Application
```bash
# Preview all operations (safe)
./augment-cleaner-cli --operation run-all --dry-run

# Clean database with verbose output
./augment-cleaner-cli --operation clean-database --verbose

# Run all operations without confirmation (automation)
./augment-cleaner-cli --operation run-all --no-confirm

# Get help
./augment-cleaner-cli --help
```

For detailed CLI documentation, see [CLI-README.md](CLI-README.md).

### Operation Types
- **Modify Telemetry IDs**: Changes machine and device IDs in VS Code's configuration
- **Clean Database**: Removes Augment-related entries from the SQLite database
- **Clean Workspace**: Clears workspace storage files and directories
- **Run All**: Executes all operations in sequence

## 📁 Project Structure

```
augment-telemetry-cleaner/
├── main.go                    # Application entry point
├── internal/
│   ├── cleaner/              # Core cleaning operations
│   │   ├── json_modifier.go     # JSON file modification
│   │   ├── sqlite_modifier.go   # Database cleaning
│   │   └── workspace_cleaner.go # Workspace management
│   ├── config/               # Configuration management
│   │   └── config.go            # Settings and preferences
│   ├── gui/                  # User interface
│   │   ├── main_gui.go          # Main application window
│   │   ├── operations.go        # Operation handlers
│   │   └── settings_dialog.go   # Settings configuration
│   ├── logger/               # Logging system
│   │   └── logger.go            # Structured logging
│   ├── safety/               # Safety features
│   │   └── safety_manager.go    # Pre-operation checks
│   ├── scanner/              # File system scanning
│   │   └── augment_scanner.go   # Augment file detection
│   └── utils/                # Utility functions
│       ├── backup.go            # Backup operations
│       ├── device_codes.go      # ID generation
│       └── paths.go             # Cross-platform paths
└── go.mod                    # Go module definition
```

## ⚙️ Configuration

The application stores its configuration in:
- **Windows**: `%APPDATA%\augment-telemetry-cleaner\config.json`
- **macOS**: `~/Library/Application Support/augment-telemetry-cleaner/config.json`
- **Linux**: `~/.config/augment-telemetry-cleaner/config.json`

### Configurable Options
- Dry-run mode default setting
- Backup creation preferences
- Confirmation dialog requirements
- Log level settings
- Backup directory location
- Maximum backup age
- Database operation timeouts

## 🔒 Safety Features

This application prioritizes data safety:

1. **Automatic Backups**: All original files are backed up before modification
2. **Dry-Run Mode**: Test operations without making actual changes
3. **Verification Checks**: Backup integrity is verified before proceeding
4. **Rollback Capability**: Backups can be used to restore original state
5. **Comprehensive Logging**: All operations are logged for audit purposes

## 🤝 Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

### Development Setup
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## 👨‍💻 Developer

**Vinay Koirala**
- 📧 Email: [koiralavinay@gmail.com](mailto:koiralavinay@gmail.com)
- 🐙 GitHub: [github.com/v-eenay](https://github.com/v-eenay)
- 💼 LinkedIn: [linkedin.com/in/veenay](https://linkedin.com/in/veenay)

## ⚠️ Disclaimer

This tool is designed to clean telemetry data for legitimate development purposes. Users are responsible for ensuring compliance with their organization's policies and applicable terms of service.

## 🙏 Acknowledgments

- Built with [Fyne](https://fyne.io/) for the cross-platform GUI
- Uses [SQLite](https://www.sqlite.org/) for database operations
- Inspired by the need for clean development environments