# Augment Telemetry Cleaner CLI

A command-line interface for the Augment Telemetry Cleaner, providing the same powerful cleaning functionality as the GUI version but optimized for automation, scripting, and headless environments.

## 🚀 Features

- **Same Core Functionality**: All cleaning operations from the GUI version
- **Command-Line Interface**: Perfect for automation and scripting
- **Dry-Run Mode**: Preview operations without making changes
- **Flexible Output**: Text or JSON output formats
- **Comprehensive Logging**: Configurable log levels with file output
- **Safety Features**: Backup creation and confirmation prompts
- **Cross-Platform**: Works on Windows, macOS, and Linux

## 📦 Installation

### Option 1: Build from Source

```bash
# Clone the repository
git clone https://github.com/v-eenay/augment-telemetry-cleaner.git
cd augment-telemetry-cleaner

# Build the CLI version
go build -o augment-telemetry-cleaner-cli ./cmd/cli/

# Or use the build scripts for multiple platforms
./build-cli.sh        # Linux/macOS
build-cli.bat          # Windows
```

### Option 2: Download Pre-built Binaries

Download the appropriate binary for your platform from the [releases page](https://github.com/v-eenay/augment-telemetry-cleaner/releases).

## 🔧 Usage

### Basic Syntax

```bash
augment-telemetry-cleaner-cli --operation <operation> [options]
```

### Available Operations

- `modify-telemetry` - Modify VS Code telemetry IDs
- `clean-database` - Clean Augment data from VS Code database
- `clean-workspace` - Clean VS Code workspace storage
- `clean-browser` - Clean Augment data from browsers
- `run-all` - Run all cleaning operations

### Command-Line Options

| Option | Description | Default |
|--------|-------------|---------|
| `--operation <op>` | Operation to perform (required) | - |
| `--dry-run` | Preview operations without making changes | false |
| `--verbose` | Enable verbose output | false |
| `--backup` | Create backups before operations | true |
| `--no-backup` | Disable backup creation | false |
| `--no-confirm` | Skip confirmation prompts | false |
| `--browser <browser>` | Target specific browser | all |
| `--output <format>` | Output format: text, json | text |
| `--log-level <level>` | Log level: DEBUG, INFO, WARN, ERROR | INFO |
| `--help` | Show help message | - |

## 📋 Examples

### Preview All Operations (Safe)
```bash
# See what would be cleaned without making changes
augment-telemetry-cleaner-cli --operation run-all --dry-run --verbose
```

### Clean Database with Verbose Output
```bash
# Clean VS Code database with detailed logging
augment-telemetry-cleaner-cli --operation clean-database --verbose
```

### Clean Browser Data (Chrome Only)
```bash
# Clean only Chrome browser data without confirmation
augment-telemetry-cleaner-cli --operation clean-browser --browser chrome --no-confirm
```

### Modify Telemetry IDs (No Backup)
```bash
# Modify telemetry IDs without creating backups
augment-telemetry-cleaner-cli --operation modify-telemetry --no-backup
```

### Automated Cleaning (Scripting)
```bash
# Run all operations without prompts, JSON output for parsing
augment-telemetry-cleaner-cli --operation run-all --no-confirm --output json > results.json
```

### Debug Mode
```bash
# Run with maximum logging for troubleshooting
augment-telemetry-cleaner-cli --operation clean-workspace --log-level DEBUG --verbose
```

## 🛡️ Safety Features

### Dry-Run Mode
Always test operations first:
```bash
augment-telemetry-cleaner-cli --operation run-all --dry-run
```

### Automatic Backups
Backups are created by default before any destructive operations:
- VS Code storage files
- Database files
- Workspace storage
- Browser data (when possible)

### Confirmation Prompts
Interactive confirmation for destructive operations (can be disabled with `--no-confirm`).

### Comprehensive Logging
All operations are logged to files in the `logs/` directory with timestamps.

## 📊 Output Formats

### Text Output (Default)
Human-readable format perfect for interactive use:
```
=== Augment Telemetry Cleaner CLI v2.0.0 ===
Operation: clean-database
Mode: LIVE (Making actual changes)
Backups: true
==========================================

🗃️ Cleaning VS Code database...
✅ Database Cleaning completed successfully!

Result Details:
  Records Deleted: 42
  Database Backup: /path/to/backup.db
```

### JSON Output
Machine-readable format for automation:
```bash
augment-telemetry-cleaner-cli --operation clean-database --output json
```

```json
{
  "deleted_rows": 42,
  "db_backup_path": "/path/to/backup.db",
  "operation_time": "2025-01-01T12:00:00Z"
}
```

## 🔄 Integration with CI/CD

The CLI version is perfect for automation:

```yaml
# GitHub Actions example
- name: Clean Augment Telemetry Data
  run: |
    ./augment-telemetry-cleaner-cli --operation run-all --no-confirm --output json > results.json
    cat results.json
```

```bash
# Shell script example
#!/bin/bash
set -e

echo "Cleaning Augment telemetry data..."
./augment-telemetry-cleaner-cli --operation run-all --no-confirm --verbose

if [ $? -eq 0 ]; then
    echo "Cleaning completed successfully"
else
    echo "Cleaning failed"
    exit 1
fi
```

## ⚠️ Important Notes

- **Browser Warning**: Close all browsers before running browser cleaning operations
- **Backup Location**: Backups are stored in your Documents folder under `Augment-Telemetry-Backups`
- **Permissions**: May require elevated permissions on some systems
- **VS Code**: Close VS Code before running operations for best results

## 🆘 Troubleshooting

### Common Issues

1. **Permission Denied**
   ```bash
   # Run with appropriate permissions
   sudo ./augment-telemetry-cleaner-cli --operation clean-database  # Linux/macOS
   ```

2. **Database Locked**
   ```bash
   # Close VS Code and try again
   ./augment-telemetry-cleaner-cli --operation clean-database --verbose
   ```

3. **Browser Still Running**
   ```bash
   # Use dry-run to check what would be cleaned
   ./augment-telemetry-cleaner-cli --operation clean-browser --dry-run
   ```

### Debug Mode
For detailed troubleshooting:
```bash
./augment-telemetry-cleaner-cli --operation run-all --log-level DEBUG --verbose --dry-run
```

## 📝 Logs

Logs are automatically created in the `logs/` directory:
- Timestamped log files
- Operation results
- Error details
- Backup locations

## 🔗 Related

- [GUI Version](README.md) - Desktop application with graphical interface
- [GitHub Repository](https://github.com/v-eenay/augment-telemetry-cleaner)
- [Issues & Support](https://github.com/v-eenay/augment-telemetry-cleaner/issues)

---

**⚠️ User Warning**: This application may log you out of other browser extensions and accounts, but Augment will continue to work properly even with a new email account after running this tool.
