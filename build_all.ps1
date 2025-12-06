# JosSecurity Build Script for Windows PowerShell
# Must be run from the root of the project (next to go.mod).

$ErrorActionPreference = "Stop"

$SourcePackage = "./cmd/joss"
$InstallerDir = "installer"
$VSIXSourceDir = "vscode-joss"

Write-Host "==================" -ForegroundColor Blue
Write-Host "  JosSecurity Build" -ForegroundColor Blue
Write-Host "==================" -ForegroundColor Blue

if (-not (Test-Path $InstallerDir)) {
    New-Item -ItemType Directory -Path $InstallerDir | Out-Null
}
Write-Host "Installer directory ready: $InstallerDir"

# Build VSIX
Write-Host "Building VSIX..."
Push-Location $VSIXSourceDir
try {
    # Remove old vsix
    Get-ChildItem -Filter "*.vsix" | Remove-Item -Force -ErrorAction SilentlyContinue
    
    # Build
    Write-Host "Running 'npm run package'..."
    cmd /c "npm run package"
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "npm run package failed. Is Node.js installed?" -ForegroundColor Red
        # Don't exit, maybe continuing with just binaries is desired? 
        # But user requested this. Let's flag it.
    }
} catch {
    Write-Host "VSIX Build Error: $($_.Exception.Message)" -ForegroundColor Red
}
Pop-Location

# Copy new VSIX
$LatestVSIX = Get-ChildItem -Path $VSIXSourceDir -Filter "*.vsix" | Sort-Object LastWriteTime -Descending | Select-Object -First 1

if ($LatestVSIX) {
    Copy-Item -Path $LatestVSIX.FullName -Destination $InstallerDir -Force
    Write-Host "VSIX Built and Copied: $($LatestVSIX.Name)" -ForegroundColor Green
} else {
    Write-Host "VSIX Build Failed or Not Found." -ForegroundColor Yellow
}

# Define Targets
$Targets = @(
    @{GOOS="windows"; GOARCH="amd64"; OutputName="joss.exe"},
    @{GOOS="windows"; GOARCH="arm64"; OutputName="joss-windows-arm64.exe"},
    @{GOOS="linux"; GOARCH="amd64"; OutputName="joss-linux-amd64"},
    @{GOOS="linux"; GOARCH="arm64"; OutputName="joss-linux-arm64"},
    @{GOOS="linux"; GOARCH="arm"; OutputName="joss-linux-armv7"},
    @{GOOS="darwin"; GOARCH="amd64"; OutputName="joss-macos-amd64"},
    @{GOOS="darwin"; GOARCH="arm64"; OutputName="joss-macos-arm64"}
)

Write-Host "Starting compilation..."

foreach ($Target in $Targets) {
    $GOOS = $Target.GOOS
    $GOARCH = $Target.GOARCH
    $OutputName = $Target.OutputName
    
    Write-Host "Compiling for $GOOS/$GOARCH ($OutputName)..."

    $env:GOOS = $GOOS
    $env:GOARCH = $GOARCH
    $env:CGO_ENABLED = 0

    try {
        go build -o "$InstallerDir/$OutputName" $SourcePackage
        Write-Host "  [OK]" -ForegroundColor Green
    } catch {
        Write-Host "  [ERROR] Compilation failed: $($_.Exception.Message)" -ForegroundColor Red
    }
    
    Remove-Item Env:GOOS
    Remove-Item Env:GOARCH
    Remove-Item Env:CGO_ENABLED
}

# Zip Files
Write-Host "Packaging..." -ForegroundColor Cyan

function Compress-Files {
    param($Files, $ZipName)
    if (Test-Path $ZipName) { Remove-Item $ZipName -Force | Out-Null }
    
    $TempZipDir = New-Item -ItemType Directory -Path "$env:TEMP/joss_zip_temp_$(Get-Random)" -Force
    
    foreach ($File in $Files) {
        if (Test-Path "$InstallerDir/$File") {
            Copy-Item "$InstallerDir/$File" $TempZipDir -Force
        }
    }
    
    if ((Get-ChildItem $TempZipDir).Count -gt 0) {
        Get-ChildItem -Path $TempZipDir -Recurse | Compress-Archive -DestinationPath $ZipName -Force
        Write-Host "Created $ZipName" -ForegroundColor Green
    } else {
        Write-Host "Skipped $ZipName (files not found)" -ForegroundColor Yellow
    }
    
    Remove-Item $TempZipDir -Recurse -Force
}

try {
    # 1. Extension
    $VSIXFiles = Get-ChildItem -Path $InstallerDir -Filter "*.vsix" | Select-Object -ExpandProperty Name
    Compress-Files -Files $VSIXFiles -ZipName "jossecurity-vscode.zip"

    # 2. Windows
    Compress-Files -Files @("joss.exe", "joss-windows-arm64.exe") -ZipName "jossecurity-windows.zip"

    # 3. Linux
    Compress-Files -Files @("joss-linux-amd64", "joss-linux-arm64", "joss-linux-armv7") -ZipName "jossecurity-linux.zip"

    # 4. macOS
    Compress-Files -Files @("joss-macos-amd64", "joss-macos-arm64") -ZipName "jossecurity-macos.zip"

    Write-Host "Ready for GitHub Releases." -ForegroundColor Yellow

} catch {
    Write-Host "Compression failed: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "Done." -ForegroundColor Green