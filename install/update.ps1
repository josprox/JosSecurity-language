# JosSecurity Updater v1.0 - Windows PowerShell
# Automatic update checker and installer

$ErrorActionPreference = "Stop"
$Host.UI.RawUI.ForegroundColor = "White"

# Variables
$RepoUrl = "https://api.github.com/repos/josprox/JosSecurity-language/releases/latest"
$InstallDir = "C:\Program Files\JosSecurity"
$CurrentVersion = "3.0.0"
$LogFile = "$env:TEMP\jossecurity-update.log"

# Logging
function Write-Log {
    param([string]$Message, [string]$Level = "INFO")
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    Add-Content -Path $LogFile -Value "[$timestamp] [$Level] $Message"
    
    switch ($Level) {
        "ERROR" { Write-Host $Message -ForegroundColor Red }
        "SUCCESS" { Write-Host $Message -ForegroundColor Green }
        "WARNING" { Write-Host $Message -ForegroundColor Yellow }
        default { Write-Host $Message }
    }
}

# Banner
Write-Host "=======================================" -ForegroundColor Blue
Write-Host "   JosSecurity Updater v1.0            " -ForegroundColor Blue
Write-Host "=======================================" -ForegroundColor Blue
Write-Host ""
Write-Log "Update check started" "INFO"

# Check for updates
function Test-Update {
    Write-Log "Checking for updates..." "INFO"
    Write-Log "Current version: $CurrentVersion" "INFO"
    
    try {
        # Get latest release info
        $release = Invoke-RestMethod -Uri $RepoUrl -UseBasicParsing
        $latestVersion = $release.tag_name -replace '^v', ''
        
        Write-Log "Latest version: $latestVersion" "INFO"
        
        # Compare versions
        if ([version]$latestVersion -gt [version]$CurrentVersion) {
            Write-Log "[!] Update available: $latestVersion" "WARNING"
            return @{
                Available = $true
                Version = $latestVersion
                DownloadUrl = $release.assets | Where-Object { $_.name -eq "joss.exe" } | Select-Object -ExpandProperty browser_download_url
                ExtensionUrl = $release.assets | Where-Object { $_.name -like "joss-language-*.vsix" } | Select-Object -ExpandProperty browser_download_url
            }
        } else {
            Write-Log "[OK] You have the latest version" "SUCCESS"
            return @{ Available = $false }
        }
    } catch {
        Write-Log "[X] Error checking for updates: $($_.Exception.Message)" "ERROR"
        return @{ Available = $false }
    }
}

# Update JosSecurity
function Update-JosSecurity {
    param($UpdateInfo)
    
    Write-Log "Updating JosSecurity..." "INFO"
    
    try {
        # Download new binary
        $tempFile = "$env:TEMP\joss-new.exe"
        Write-Log "Downloading version $($UpdateInfo.Version)..." "INFO"
        Invoke-WebRequest -Uri $UpdateInfo.DownloadUrl -OutFile $tempFile -UseBasicParsing
        
        # Backup current version
        if (Test-Path "$InstallDir\joss.exe") {
            Copy-Item "$InstallDir\joss.exe" "$InstallDir\joss.exe.bak" -Force
            Write-Log "Backup created" "INFO"
        }
        
        # Replace binary
        Copy-Item $tempFile "$InstallDir\joss.exe" -Force
        Remove-Item $tempFile -Force
        
        # Verify
        if (Test-Path "$InstallDir\joss.exe") {
            Write-Log "[OK] JosSecurity updated to $($UpdateInfo.Version)" "SUCCESS"
            return $true
        } else {
            Write-Log "[X] Update verification failed" "ERROR"
            
            # Restore backup
            if (Test-Path "$InstallDir\joss.exe.bak") {
                Copy-Item "$InstallDir\joss.exe.bak" "$InstallDir\joss.exe" -Force
                Write-Log "Restored from backup" "WARNING"
            }
            return $false
        }
    } catch {
        Write-Log "[X] Update failed: $($_.Exception.Message)" "ERROR"
        
        # Restore backup
        if (Test-Path "$InstallDir\joss.exe.bak") {
            Copy-Item "$InstallDir\joss.exe.bak" "$InstallDir\joss.exe" -Force
            Write-Log "Restored from backup" "WARNING"
        }
        return $false
    }
}

# Update Extension
function Update-Extension {
    param($UpdateInfo)
    
    Write-Log "Updating VS Code extension..." "INFO"
    
    try {
        # Download new VSIX
        $tempFile = "$env:TEMP\joss-language-new.vsix"
        Write-Log "Downloading extension..." "INFO"
        Invoke-WebRequest -Uri $UpdateInfo.ExtensionUrl -OutFile $tempFile -UseBasicParsing
        
        # Install extension
        & code --install-extension $tempFile --force 2>&1 | Out-Null
        Remove-Item $tempFile -Force
        
        # Verify
        Start-Sleep -Seconds 2
        $extensions = & code --list-extensions 2>&1
        if ($extensions -match "joss-language") {
            Write-Log "[OK] Extension updated" "SUCCESS"
            return $true
        } else {
            Write-Log "[X] Extension update verification failed" "ERROR"
            return $false
        }
    } catch {
        Write-Log "[X] Extension update failed: $($_.Exception.Message)" "ERROR"
        return $false
    }
}

# Main update flow
$updateInfo = Test-Update

if ($updateInfo.Available) {
    Write-Host ""
    Write-Host "Update available: v$($updateInfo.Version)" -ForegroundColor Green
    Write-Host ""
    Write-Host "What do you want to update?" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "  [1] Update JosSecurity only" -ForegroundColor White
    Write-Host "  [2] Update VS Code Extension only" -ForegroundColor White
    Write-Host "  [3] Update Everything (Recommended)" -ForegroundColor White
    Write-Host "  [0] Cancel" -ForegroundColor White
    Write-Host ""
    
    $option = Read-Host "Option"
    
    switch ($option) {
        "1" {
            Update-JosSecurity -UpdateInfo $updateInfo
        }
        "2" {
            Update-Extension -UpdateInfo $updateInfo
        }
        "3" {
            Update-JosSecurity -UpdateInfo $updateInfo
            Update-Extension -UpdateInfo $updateInfo
        }
        "0" {
            Write-Log "Update cancelled" "INFO"
            exit
        }
        default {
            Write-Log "Invalid option" "ERROR"
            exit
        }
    }
    
    Write-Host ""
    Write-Host "=======================================" -ForegroundColor Green
    Write-Host "Update completed" -ForegroundColor Green
    Write-Host "=======================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "Please restart terminal and VS Code" -ForegroundColor Yellow
} else {
    Write-Host ""
    Write-Host "No updates available" -ForegroundColor Green
}

Write-Log "Log saved to: $LogFile" "INFO"
