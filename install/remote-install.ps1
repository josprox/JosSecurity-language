# JosSecurity Remote Installer
# Usage: iwr -useb https://raw.githubusercontent.com/USER/REPO/main/install/remote-install.ps1 | iex

$ErrorActionPreference = "Stop"

Write-Host "=======================================" -ForegroundColor Blue
Write-Host "   JosSecurity Remote Installer        " -ForegroundColor Blue
Write-Host "=======================================" -ForegroundColor Blue
Write-Host ""

# Configuration
$RepoUrl = "https://github.com/josprox/JosSecurity-language"
$RawUrl = "https://raw.githubusercontent.com/josprox/JosSecurity-language/main/install"
$TempDir = "$env:TEMP\jossecurity-install"

# Create temp directory
if (Test-Path $TempDir) {
    Remove-Item $TempDir -Recurse -Force
}
New-Item -ItemType Directory -Path $TempDir -Force | Out-Null

Write-Host "[1/4] Downloading installer..." -ForegroundColor Yellow

try {
    # Download main installer
    Invoke-WebRequest -Uri "$RawUrl/install.ps1" -OutFile "$TempDir\install.ps1" -UseBasicParsing
    
    # Download binaries zip
    Write-Host "[2/4] Downloading JosSecurity binaries..." -ForegroundColor Yellow
    $ZipPath = "$TempDir\jossecurity-binaries.zip"
    Invoke-WebRequest -Uri "$RepoUrl/releases/latest/download/jossecurity-binaries.zip" -OutFile $ZipPath -UseBasicParsing
    
    # Extract binaries
    Write-Host "[3/4] Extracting files..." -ForegroundColor Yellow
    Expand-Archive -Path $ZipPath -DestinationPath $TempDir -Force
    Remove-Item $ZipPath -Force
    
    Write-Host "[4/4] Starting installation..." -ForegroundColor Yellow
    Write-Host ""
    
    # Change to temp directory and run installer
    Set-Location $TempDir
    & .\install.ps1
    
} catch {
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host ""
    Write-Host "Manual installation:" -ForegroundColor Yellow
    Write-Host "  1. Visit: $RepoUrl" -ForegroundColor White
    Write-Host "  2. Download install folder" -ForegroundColor White
    Write-Host "  3. Run install.ps1" -ForegroundColor White
    exit 1
}
