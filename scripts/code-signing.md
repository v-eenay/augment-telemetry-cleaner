# Code Signing Setup Guide

This document outlines the code signing process for the Augment Telemetry Cleaner application to prevent security warnings on Windows and macOS.

## Windows Code Signing

### Requirements
1. **Code Signing Certificate**: Purchase from a trusted CA (DigiCert, Sectigo, etc.)
2. **SignTool**: Part of Windows SDK (available in GitHub Actions)
3. **Certificate Storage**: Store as GitHub Secrets

### Certificate Types
- **EV (Extended Validation)**: Best for avoiding SmartScreen warnings
- **OV (Organization Validation)**: Standard code signing
- **Self-signed**: For testing only (will show warnings)

### GitHub Secrets Required
```
WINDOWS_CERTIFICATE_BASE64: Base64 encoded .p12 certificate file
WINDOWS_CERTIFICATE_PASSWORD: Certificate password
WINDOWS_CERTIFICATE_NAME: Certificate subject name
```

### Implementation Steps
1. Convert .p12 certificate to base64: `base64 -i certificate.p12 -o certificate.txt`
2. Add to GitHub Secrets
3. Use in GitHub Actions workflow

## macOS Code Signing & Notarization

### Requirements
1. **Apple Developer Account**: $99/year
2. **Developer ID Certificate**: For distribution outside App Store
3. **App-specific Password**: For notarization
4. **Xcode Command Line Tools**: For codesign and notarytool

### GitHub Secrets Required
```
MACOS_CERTIFICATE_BASE64: Base64 encoded .p12 certificate
MACOS_CERTIFICATE_PASSWORD: Certificate password
MACOS_KEYCHAIN_PASSWORD: Temporary keychain password
APPLE_ID: Apple ID email
APPLE_ID_PASSWORD: App-specific password
APPLE_TEAM_ID: Developer team ID
```

### Implementation Steps
1. Export Developer ID certificate as .p12
2. Convert to base64 and add to GitHub Secrets
3. Create app-specific password in Apple ID settings
4. Use in GitHub Actions workflow

## Linux Code Signing

Linux doesn't have a standard code signing mechanism like Windows/macOS, but we can:
1. **GPG Signing**: Sign binaries with GPG keys
2. **Package Signing**: Sign .deb/.rpm packages
3. **Checksums**: Provide SHA256 checksums for verification

## Security Enhancements

### Windows Executable Metadata
- **File Version**: Matches application version
- **Product Version**: Application version
- **Company Name**: Developer/organization name
- **File Description**: Application description
- **Copyright**: Copyright notice
- **Original Filename**: Executable name

### SmartScreen Reputation
- **Consistent Signing**: Always sign with same certificate
- **Download Volume**: Higher download counts improve reputation
- **Time**: Reputation builds over time
- **No Malware Reports**: Keep clean reputation

### Best Practices
1. **Always Sign**: Sign all releases consistently
2. **Timestamp**: Use timestamp servers for long-term validity
3. **Strong Certificates**: Use EV certificates when possible
4. **Clean Builds**: Build in clean environments
5. **Virus Scanning**: Scan binaries before signing

## Cost Considerations

### Certificate Costs (Annual)
- **Windows EV Certificate**: $300-500/year
- **Windows OV Certificate**: $100-300/year
- **Apple Developer Account**: $99/year
- **Total**: $400-600/year for full code signing

### Free Alternatives
- **Self-signed certificates**: For testing/internal use
- **Open source certificates**: Limited trust
- **Checksums only**: Basic integrity verification

## Implementation Priority
1. **High Priority**: Windows code signing (most security warnings)
2. **Medium Priority**: macOS notarization (Gatekeeper warnings)
3. **Low Priority**: Linux GPG signing (less common)

## Testing
1. **Windows**: Test on clean Windows machines
2. **macOS**: Test with Gatekeeper enabled
3. **Antivirus**: Test with major antivirus software
4. **SmartScreen**: Monitor SmartScreen reputation
