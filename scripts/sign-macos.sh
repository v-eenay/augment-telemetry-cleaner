#!/bin/bash

# macOS Code Signing and Notarization Script
# This script signs macOS binaries and submits them for notarization

set -e

BINARY_PATH="$1"
CERTIFICATE_BASE64="$2"
CERTIFICATE_PASSWORD="$3"
KEYCHAIN_PASSWORD="$4"
APPLE_ID="$5"
APPLE_ID_PASSWORD="$6"
APPLE_TEAM_ID="$7"

echo "macOS Code Signing and Notarization Script"
echo "=========================================="

# Check if binary exists
if [ ! -f "$BINARY_PATH" ]; then
    echo "❌ Binary not found: $BINARY_PATH"
    exit 1
fi

echo "Binary to sign: $BINARY_PATH"

# Check if we have certificate data
if [ -z "$CERTIFICATE_BASE64" ]; then
    echo "⚠️ No certificate provided. Skipping code signing."
    echo "To enable code signing, set the following GitHub Secrets:"
    echo "- MACOS_CERTIFICATE_BASE64"
    echo "- MACOS_CERTIFICATE_PASSWORD"
    echo "- MACOS_KEYCHAIN_PASSWORD"
    echo "- APPLE_ID"
    echo "- APPLE_ID_PASSWORD"
    echo "- APPLE_TEAM_ID"
    exit 0
fi

# Create temporary keychain
KEYCHAIN_NAME="build-$(date +%s)"
KEYCHAIN_PATH="$HOME/Library/Keychains/$KEYCHAIN_NAME.keychain-db"

echo "Creating temporary keychain..."
security create-keychain -p "$KEYCHAIN_PASSWORD" "$KEYCHAIN_NAME"
security set-keychain-settings -lut 21600 "$KEYCHAIN_NAME"
security unlock-keychain -p "$KEYCHAIN_PASSWORD" "$KEYCHAIN_NAME"

# Add keychain to search list
security list-keychains -d user -s "$KEYCHAIN_NAME" $(security list-keychains -d user | sed s/\"//g)

# Import certificate
echo "Importing certificate..."
CERT_FILE=$(mktemp)
echo "$CERTIFICATE_BASE64" | base64 --decode > "$CERT_FILE"
security import "$CERT_FILE" -k "$KEYCHAIN_NAME" -P "$CERTIFICATE_PASSWORD" -T /usr/bin/codesign
rm "$CERT_FILE"

# Set partition list for codesign
security set-key-partition-list -S apple-tool:,apple:,codesign: -s -k "$KEYCHAIN_PASSWORD" "$KEYCHAIN_NAME"

# Find certificate identity
CERT_IDENTITY=$(security find-identity -v -p codesigning "$KEYCHAIN_NAME" | grep "Developer ID Application" | head -1 | grep -o '"[^"]*"' | sed 's/"//g')

if [ -z "$CERT_IDENTITY" ]; then
    echo "❌ No Developer ID Application certificate found"
    exit 1
fi

echo "Using certificate: $CERT_IDENTITY"

# Sign the binary
echo "Signing binary..."
codesign --force --sign "$CERT_IDENTITY" --timestamp --options runtime "$BINARY_PATH"

# Verify signature
echo "Verifying signature..."
codesign --verify --verbose "$BINARY_PATH"
spctl --assess --type execute --verbose "$BINARY_PATH"

echo "✅ Binary signed successfully!"

# Notarization (if Apple ID credentials are provided)
if [ -n "$APPLE_ID" ] && [ -n "$APPLE_ID_PASSWORD" ] && [ -n "$APPLE_TEAM_ID" ]; then
    echo "Starting notarization process..."
    
    # Create ZIP file for notarization
    ZIP_FILE="${BINARY_PATH}.zip"
    zip -r "$ZIP_FILE" "$BINARY_PATH"
    
    # Submit for notarization
    echo "Submitting for notarization..."
    NOTARIZATION_ID=$(xcrun notarytool submit "$ZIP_FILE" \
        --apple-id "$APPLE_ID" \
        --password "$APPLE_ID_PASSWORD" \
        --team-id "$APPLE_TEAM_ID" \
        --wait \
        --output-format json | jq -r '.id')
    
    if [ "$NOTARIZATION_ID" != "null" ] && [ -n "$NOTARIZATION_ID" ]; then
        echo "Notarization submitted with ID: $NOTARIZATION_ID"
        
        # Wait for notarization to complete
        echo "Waiting for notarization to complete..."
        xcrun notarytool wait "$NOTARIZATION_ID" \
            --apple-id "$APPLE_ID" \
            --password "$APPLE_ID_PASSWORD" \
            --team-id "$APPLE_TEAM_ID"
        
        # Check notarization status
        STATUS=$(xcrun notarytool info "$NOTARIZATION_ID" \
            --apple-id "$APPLE_ID" \
            --password "$APPLE_ID_PASSWORD" \
            --team-id "$APPLE_TEAM_ID" \
            --output-format json | jq -r '.status')
        
        if [ "$STATUS" = "Accepted" ]; then
            echo "✅ Notarization successful!"
            
            # Staple the notarization
            echo "Stapling notarization..."
            xcrun stapler staple "$BINARY_PATH"
            echo "✅ Notarization stapled!"
        else
            echo "❌ Notarization failed with status: $STATUS"
            # Get detailed log
            xcrun notarytool log "$NOTARIZATION_ID" \
                --apple-id "$APPLE_ID" \
                --password "$APPLE_ID_PASSWORD" \
                --team-id "$APPLE_TEAM_ID"
        fi
    else
        echo "❌ Failed to submit for notarization"
    fi
    
    # Clean up ZIP file
    rm -f "$ZIP_FILE"
else
    echo "⚠️ Skipping notarization (Apple ID credentials not provided)"
fi

# Clean up keychain
echo "Cleaning up..."
security delete-keychain "$KEYCHAIN_NAME" || true

echo "Code signing and notarization completed."
