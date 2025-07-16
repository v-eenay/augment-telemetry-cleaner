# Code Signing Setup Instructions

This document provides step-by-step instructions for setting up code signing for the Augment Telemetry Cleaner application.

## GitHub Secrets Configuration

To enable code signing, you need to configure the following GitHub Secrets in your repository settings:

### Windows Code Signing Secrets

1. **WINDOWS_CERTIFICATE_BASE64**
   - Your code signing certificate (.p12 file) encoded in base64
   - To generate: `base64 -i your-certificate.p12 -o certificate.txt`
   - Copy the contents of certificate.txt

2. **WINDOWS_CERTIFICATE_PASSWORD**
   - The password for your .p12 certificate file

3. **WINDOWS_CERTIFICATE_NAME**
   - The subject name of your certificate (e.g., "Your Company Name")

### macOS Code Signing Secrets

1. **MACOS_CERTIFICATE_BASE64**
   - Your Developer ID Application certificate (.p12 file) encoded in base64
   - Export from Keychain Access and encode with base64

2. **MACOS_CERTIFICATE_PASSWORD**
   - The password for your macOS .p12 certificate file

3. **MACOS_KEYCHAIN_PASSWORD**
   - A temporary password for the build keychain (can be any secure string)

4. **APPLE_ID**
   - Your Apple ID email address

5. **APPLE_ID_PASSWORD**
   - App-specific password for your Apple ID
   - Generate at: https://appleid.apple.com/account/manage

6. **APPLE_TEAM_ID**
   - Your Apple Developer Team ID (found in Apple Developer portal)

## Setting Up GitHub Secrets

1. Go to your GitHub repository
2. Click on "Settings" tab
3. In the left sidebar, click "Secrets and variables" → "Actions"
4. Click "New repository secret"
5. Add each secret with the exact name and value

## Certificate Requirements

### Windows Certificate
- **Type**: Code Signing Certificate
- **Recommended**: Extended Validation (EV) certificate for best SmartScreen reputation
- **Providers**: DigiCert, Sectigo, GlobalSign, etc.
- **Cost**: $100-500/year depending on type and provider

### macOS Certificate
- **Type**: Developer ID Application certificate
- **Requirement**: Apple Developer Program membership ($99/year)
- **Steps**:
  1. Join Apple Developer Program
  2. Create Developer ID Application certificate in Apple Developer portal
  3. Download and install in Keychain Access
  4. Export as .p12 file

## Testing Code Signing

### Without Certificates (Development)
The build will work without certificates but binaries won't be signed:
- Windows: Will show "Unknown publisher" warnings
- macOS: Will show Gatekeeper warnings

### With Certificates (Production)
- Windows: Reduced SmartScreen warnings (reputation builds over time)
- macOS: No Gatekeeper warnings after notarization

## Troubleshooting

### Windows Issues
- **SignTool not found**: Install Windows SDK
- **Certificate import fails**: Check password and certificate format
- **Signing fails**: Verify certificate is valid and not expired

### macOS Issues
- **Certificate not found**: Ensure Developer ID Application certificate is used
- **Notarization fails**: Check Apple ID credentials and team ID
- **Keychain issues**: Verify keychain password and permissions

### GitHub Actions Issues
- **Secret not found**: Verify secret names match exactly
- **Base64 decode fails**: Ensure certificate is properly encoded
- **Permission denied**: Check script permissions and paths

## Security Best Practices

1. **Protect Certificates**: Never commit certificates to version control
2. **Use Strong Passwords**: Use complex passwords for certificate files
3. **Rotate Secrets**: Regularly update app-specific passwords
4. **Monitor Usage**: Review GitHub Actions logs for signing attempts
5. **Backup Certificates**: Keep secure backups of certificate files

## Cost-Benefit Analysis

### With Code Signing
- **Pros**: Professional appearance, reduced security warnings, user trust
- **Cons**: Annual certificate costs, setup complexity

### Without Code Signing
- **Pros**: No additional costs, simpler setup
- **Cons**: Security warnings, reduced user trust, potential antivirus flags

## Recommended Approach

1. **Start without signing**: Get the application working first
2. **Add Windows signing**: Most users are on Windows, biggest impact
3. **Add macOS signing**: If you have significant macOS users
4. **Monitor feedback**: See if users report security warnings

## Support

If you encounter issues with code signing setup:
1. Check the GitHub Actions logs for detailed error messages
2. Verify all secrets are set correctly
3. Test certificate validity locally before using in CI/CD
4. Consult certificate provider documentation for specific requirements
