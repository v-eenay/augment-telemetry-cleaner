# Windows Code Signing Script
# This script signs Windows executables using a code signing certificate

param(
    [Parameter(Mandatory=$true)]
    [string]$BinaryPath,
    
    [Parameter(Mandatory=$false)]
    [string]$CertificateBase64,
    
    [Parameter(Mandatory=$false)]
    [string]$CertificatePassword,
    
    [Parameter(Mandatory=$false)]
    [string]$CertificateName,
    
    [Parameter(Mandatory=$false)]
    [string]$TimestampUrl = "http://timestamp.digicert.com"
)

Write-Host "Windows Code Signing Script"
Write-Host "============================"

# Check if binary exists
if (-not (Test-Path $BinaryPath)) {
    Write-Error "Binary not found: $BinaryPath"
    exit 1
}

Write-Host "Binary to sign: $BinaryPath"

# Check if we have certificate data
if ([string]::IsNullOrEmpty($CertificateBase64)) {
    Write-Warning "No certificate provided. Skipping code signing."
    Write-Host "To enable code signing, set the following GitHub Secrets:"
    Write-Host "- WINDOWS_CERTIFICATE_BASE64"
    Write-Host "- WINDOWS_CERTIFICATE_PASSWORD"
    Write-Host "- WINDOWS_CERTIFICATE_NAME"
    exit 0
}

try {
    # Create temporary directory for certificate
    $tempDir = New-TemporaryFile | ForEach-Object { Remove-Item $_; New-Item -ItemType Directory -Path $_ }
    $certPath = Join-Path $tempDir "certificate.p12"
    
    Write-Host "Preparing certificate..."
    
    # Decode base64 certificate
    $certBytes = [System.Convert]::FromBase64String($CertificateBase64)
    [System.IO.File]::WriteAllBytes($certPath, $certBytes)
    
    # Import certificate to temporary store
    $securePassword = ConvertTo-SecureString -String $CertificatePassword -AsPlainText -Force
    $cert = Import-PfxCertificate -FilePath $certPath -Password $securePassword -CertStoreLocation "Cert:\CurrentUser\My"
    
    Write-Host "Certificate imported: $($cert.Subject)"
    
    # Find SignTool.exe
    $signToolPaths = @(
        "${env:ProgramFiles(x86)}\Windows Kits\10\bin\*\x64\signtool.exe",
        "${env:ProgramFiles}\Windows Kits\10\bin\*\x64\signtool.exe",
        "${env:ProgramFiles(x86)}\Microsoft SDKs\Windows\*\bin\signtool.exe",
        "${env:ProgramFiles}\Microsoft SDKs\Windows\*\bin\signtool.exe"
    )
    
    $signTool = $null
    foreach ($path in $signToolPaths) {
        $found = Get-ChildItem -Path $path -ErrorAction SilentlyContinue | Select-Object -First 1
        if ($found) {
            $signTool = $found.FullName
            break
        }
    }
    
    if (-not $signTool) {
        Write-Error "SignTool.exe not found. Please install Windows SDK."
        exit 1
    }
    
    Write-Host "Using SignTool: $signTool"
    
    # Sign the binary
    Write-Host "Signing binary..."
    
    $signArgs = @(
        "sign",
        "/fd", "SHA256",
        "/tr", $TimestampUrl,
        "/td", "SHA256",
        "/a"
    )
    
    if (-not [string]::IsNullOrEmpty($CertificateName)) {
        $signArgs += "/n"
        $signArgs += $CertificateName
    }
    
    $signArgs += $BinaryPath
    
    $process = Start-Process -FilePath $signTool -ArgumentList $signArgs -Wait -PassThru -NoNewWindow
    
    if ($process.ExitCode -eq 0) {
        Write-Host "✅ Binary signed successfully!" -ForegroundColor Green
        
        # Verify signature
        Write-Host "Verifying signature..."
        $verifyArgs = @("verify", "/pa", "/v", $BinaryPath)
        $verifyProcess = Start-Process -FilePath $signTool -ArgumentList $verifyArgs -Wait -PassThru -NoNewWindow
        
        if ($verifyProcess.ExitCode -eq 0) {
            Write-Host "✅ Signature verification successful!" -ForegroundColor Green
        } else {
            Write-Warning "⚠️ Signature verification failed"
        }
    } else {
        Write-Error "❌ Code signing failed with exit code: $($process.ExitCode)"
        exit 1
    }
    
} catch {
    Write-Error "❌ Code signing error: $($_.Exception.Message)"
    exit 1
} finally {
    # Clean up
    if ($cert) {
        Remove-Item "Cert:\CurrentUser\My\$($cert.Thumbprint)" -ErrorAction SilentlyContinue
    }
    if ($tempDir -and (Test-Path $tempDir)) {
        Remove-Item $tempDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}

Write-Host "Code signing completed."
