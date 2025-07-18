# Augment Telemetry Cleaner v1.1.0 Release Notes

## 🎉 Major Feature Release

This release introduces three significant enhancements that transform the Augment Telemetry Cleaner into a more professional, secure, and comprehensive privacy tool.

---

## 🎨 **1. Professional Application Icon & Branding**

### New Features
- **Custom SVG Icon**: Professional icon design featuring:
  - Security shield representing privacy protection
  - Cleaning brush symbolizing data removal
  - Data particles being cleaned away
  - VS Code-inspired color accents
  - Motion lines showing active cleaning

- **GUI Integration**: 
  - Icon appears in window title bar
  - Icon shows in taskbar/dock
  - Embedded directly in compiled binaries

- **Cross-Platform Support**:
  - Windows: ICO format with multiple sizes
  - macOS: ICNS format for Retina displays
  - Linux: PNG format in standard sizes

### Technical Implementation
- `assets/icon.svg` - Master SVG icon file
- `internal/assets/icon.go` - Embedded icon resource
- `scripts/generate-icons.go` - Platform-specific icon generation
- `FyneApp.toml` - Application metadata configuration

---

## 🔒 **2. Code Signing & Security Infrastructure**

### Windows Code Signing
- **PowerShell Script**: `scripts/sign-windows.ps1`
- **Features**:
  - Automatic certificate import and cleanup
  - SignTool.exe detection across Windows SDK versions
  - SHA256 signing with timestamp verification
  - Comprehensive error handling and logging

### macOS Code Signing & Notarization
- **Shell Script**: `scripts/sign-macos.sh`
- **Features**:
  - Developer ID Application certificate support
  - Automatic keychain management
  - Apple notarization service integration
  - Gatekeeper compatibility verification

### GitHub Actions Integration
- **Automated Signing**: Integrated into release workflow
- **Security**: Uses GitHub Secrets for certificate storage
- **Validation**: Binary verification before release creation

### Documentation
- `docs/SETUP-CODE-SIGNING.md` - Complete setup guide
- `scripts/code-signing.md` - Technical implementation details
- Certificate requirements and cost analysis
- Troubleshooting guides for common issues

### Benefits
- **Windows**: Reduces SmartScreen warnings
- **macOS**: Eliminates Gatekeeper security warnings
- **Professional**: Builds user trust and credibility

---

## 🌐 **3. Browser Data Cleaning Feature**

### Supported Browsers
- **Google Chrome**: All profiles and user data directories
- **Microsoft Edge**: Chromium-based Edge installations
- **Mozilla Firefox**: Profile-based installations
- **Safari**: macOS native browser (limited support)

### Cross-Platform Detection
- **Windows**: AppData and Program Files locations
- **macOS**: Library/Application Support directories
- **Linux**: .config and .mozilla directories

### Data Types Cleaned
- **Cookies**: Augment-related domain cookies
- **Local Storage**: Browser local storage data
- **Session Storage**: Temporary session data
- **Cache**: Browser cache files (selective)

### Safety Features
- **Process Detection**: Warns if browsers are running
- **Backup Creation**: Automatic backup of critical files
- **Dry-Run Mode**: Preview what would be cleaned
- **Confirmation Dialogs**: User consent for destructive operations
- **Error Handling**: Graceful failure with detailed logging

### Technical Implementation
- `internal/browser/browser_detector.go` - Browser discovery
- `internal/browser/browser_cleaner.go` - Data cleaning logic
- `internal/browser/browser_utils.go` - Utility functions
- GUI integration with new "Clean Browser Data" button
- Integrated into "Run All Operations" workflow

### Browser-Specific Logic
- **Chromium Browsers**: SQLite database cleaning for cookies
- **Firefox**: Profile.ini parsing and SQLite operations
- **Safari**: Limited support due to proprietary formats

---

## 🔧 **Technical Improvements**

### Enhanced Build Process
- **Icon Generation**: Automated platform-specific icon creation
- **Version Management**: Centralized version information
- **Resource Embedding**: Improved asset management

### Code Quality
- **Error Handling**: Comprehensive error management
- **Logging**: Detailed operation logging
- **Documentation**: Extensive inline and external documentation

### GitHub Actions Enhancements
- **Parallel Builds**: Matrix-based builds for faster releases
- **Icon Tools**: ImageMagick and Inkscape integration
- **Validation**: Binary verification and testing

---

## 📋 **Installation & Usage**

### Download
- Visit [GitHub Releases](https://github.com/v-eenay/augment-telemetry-cleaner/releases)
- Download the appropriate binary for your platform:
  - `augment-telemetry-cleaner-windows-amd64.exe` (Windows)
  - `augment-telemetry-cleaner-linux-amd64` (Linux)
  - `augment-telemetry-cleaner-darwin-amd64` (macOS Intel)
  - `augment-telemetry-cleaner-darwin-arm64` (macOS Apple Silicon)

### New Features Usage
1. **Browser Cleaning**: Click "Clean Browser Data" button
2. **Safety First**: Close all browsers before cleaning
3. **Dry Run**: Enable dry-run mode to preview changes
4. **Backups**: Automatic backups are created by default

---

## 🛡️ **Security & Privacy**

### Code Signing Status
- **Ready for Implementation**: Complete infrastructure in place
- **GitHub Secrets Required**: Certificate configuration needed
- **Documentation**: Step-by-step setup guides provided

### Privacy Enhancements
- **Browser Data**: Comprehensive cleaning across all major browsers
- **Local Processing**: All operations performed locally
- **No Network**: No data transmitted to external servers

---

## 🔄 **Backward Compatibility**

- **Existing Features**: All v1.0.0 features remain unchanged
- **Configuration**: Existing settings and preferences preserved
- **Safety Features**: All safety mechanisms enhanced, not replaced

---

## 🚀 **Future Roadmap**

### Planned Enhancements
- **Certificate Acquisition**: Guidance for obtaining code signing certificates
- **Additional Browsers**: Support for more browser types
- **Advanced Cleaning**: More granular cleaning options
- **Scheduling**: Automated cleaning schedules

---

## 📞 **Support & Feedback**

- **Issues**: [GitHub Issues](https://github.com/v-eenay/augment-telemetry-cleaner/issues)
- **Documentation**: See `docs/` directory for detailed guides
- **Developer**: Vinay Koirala (koiralavinay@gmail.com)

---

## 🙏 **Acknowledgments**

This release represents a significant evolution of the Augment Telemetry Cleaner, transforming it from a simple utility into a comprehensive privacy tool with professional-grade features and security infrastructure.
