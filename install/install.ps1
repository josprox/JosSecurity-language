# JosSecurity Installer v1.0 - Windows PowerShell
# Enhanced with comprehensive verification

$ErrorActionPreference = "Stop"
$Host.UI.RawUI.ForegroundColor = "White"

# Variables
$JossVersion = "3.0.0"
$InstallDir = "C:\Program Files\JosSecurity"
$VSCodeInstaller = "https://code.visualstudio.com/sha/download?build=stable&os=win32-x64-user"
$LogFile = "$env:TEMP\jossecurity-install.log"

# Logging function
function Write-Log {
    param([string]$Message, [string]$Level = "INFO")
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $logMessage = "[$timestamp] [$Level] $Message"
    Add-Content -Path $LogFile -Value $logMessage
    
    switch ($Level) {
        "ERROR" { Write-Host $Message -ForegroundColor Red }
        "SUCCESS" { Write-Host $Message -ForegroundColor Green }
        "WARNING" { Write-Host $Message -ForegroundColor Yellow }
        default { Write-Host $Message }
    }
}

# Banner
Write-Host "=======================================" -ForegroundColor Blue
Write-Host "   JosSecurity Installer v1.0          " -ForegroundColor Blue
Write-Host "   Enhanced Installation System         " -ForegroundColor Blue
Write-Host "=======================================" -ForegroundColor Blue
Write-Host ""
Write-Log "Installation started" "INFO"

# Verify administrator
function Test-Administrator {
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

# Verify VS Code
function Test-VSCode {
    try {
        $null = Get-Command code -ErrorAction Stop
        Write-Log "[OK] VS Code detected" "SUCCESS"
        return $true
    } catch {
        Write-Log "[X] VS Code not detected" "WARNING"
        return $false
    }
}

# Add to PATH with verification
function Add-ToPath {
    param([string]$PathToAdd)
    
    try {
        $currentPath = [Environment]::GetEnvironmentVariable("Path", "Machine")
        
        if ($currentPath -notlike "*$PathToAdd*") {
            Write-Log "Adding $PathToAdd to system PATH..." "INFO"
            
            if (Test-Administrator) {
                $newPath = "$currentPath;$PathToAdd"
                [Environment]::SetEnvironmentVariable("Path", $newPath, "Machine")
                
                # Verify PATH was updated
                $updatedPath = [Environment]::GetEnvironmentVariable("Path", "Machine")
                if ($updatedPath -like "*$PathToAdd*") {
                    Write-Log "[OK] PATH updated successfully" "SUCCESS"
                    return $true
                } else {
                    Write-Log "[X] Failed to verify PATH update" "ERROR"
                    return $false
                }
            } else {
                Write-Log "[X] Administrator permissions required" "ERROR"
                return $false
            }
        } else {
            Write-Log "[OK] Already in PATH" "SUCCESS"
            return $true
        }
    } catch {
        Write-Log "[X] Error updating PATH: $($_.Exception.Message)" "ERROR"
        return $false
    }
}

# Install JosSecurity with verification
function Install-JosSecurity {
    Write-Log "[1/3] Installing JosSecurity..." "INFO"
    
    try {
        # Create installation directory
        if (-not (Test-Path $InstallDir)) {
            New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
            Write-Log "Created directory: $InstallDir" "INFO"
        }
        
        # Find binary
        $binaryPath = Get-ChildItem -Path . -Filter "joss.exe" -Recurse -ErrorAction SilentlyContinue | Select-Object -First 1
        
        if ($null -eq $binaryPath) {
            Write-Log "[X] joss.exe not found in current directory" "ERROR"
            Write-Log "Please ensure joss.exe is in the install folder" "WARNING"
            return $false
        }
        
        Write-Log "Found binary: $($binaryPath.FullName)" "INFO"
        
        # Copy binary
        Copy-Item -Path $binaryPath.FullName -Destination "$InstallDir\joss.exe" -Force
        
        # Verify copy
        if (Test-Path "$InstallDir\joss.exe") {
            $size = (Get-Item "$InstallDir\joss.exe").Length / 1MB
            Write-Log "[OK] Binary installed ($([math]::Round($size, 2)) MB)" "SUCCESS"
        } else {
            Write-Log "[X] Failed to copy binary" "ERROR"
            return $false
        }
        
        # Add to PATH
        if (-not (Add-ToPath -PathToAdd $InstallDir)) {
            Write-Log "[X] Failed to add to PATH" "ERROR"
            return $false
        }
        
        Write-Log "[OK] JosSecurity $JossVersion installed" "SUCCESS"
        return $true
    } catch {
        Write-Log "[X] Installation failed: $($_.Exception.Message)" "ERROR"
        return $false
    }
}

# Install VS Code with verification
function Install-VSCode {
    Write-Log "[2/3] Installing VS Code..." "INFO"
    
    try {
        $installerPath = "$env:TEMP\VSCodeSetup.exe"
        
        Write-Log "Downloading VS Code..." "INFO"
        Invoke-WebRequest -Uri $VSCodeInstaller -OutFile $installerPath -UseBasicParsing
        
        # Verify download
        if (-not (Test-Path $installerPath)) {
            Write-Log "[X] Failed to download VS Code" "ERROR"
            return $false
        }
        
        $downloadSize = (Get-Item $installerPath).Length / 1MB
        Write-Log "Downloaded $([math]::Round($downloadSize, 2)) MB" "INFO"
        
        Write-Log "Installing VS Code..." "INFO"
        Start-Process -FilePath $installerPath -ArgumentList "/VERYSILENT /MERGETASKS=!runcode,addcontextmenufiles,addcontextmenufolders,addtopath" -Wait
        
        Remove-Item $installerPath -Force
        
        # Update PATH in current session
        $machinePath = [System.Environment]::GetEnvironmentVariable("Path","Machine")
        $userPath = [System.Environment]::GetEnvironmentVariable("Path","User")
        $env:Path = "$machinePath;$userPath"
        
        # Verify installation
        Start-Sleep -Seconds 2
        if (Test-VSCode) {
            Write-Log "[OK] VS Code installed successfully" "SUCCESS"
            return $true
        } else {
            Write-Log "[X] VS Code installation could not be verified" "WARNING"
            return $false
        }
    } catch {
        Write-Log "[X] VS Code installation failed: $($_.Exception.Message)" "ERROR"
        return $false
    }
}

# Install extension with verification
function Install-Extension {
    Write-Log "[3/3] Installing VS Code extension..." "INFO"
    
    try {
        # Find VSIX file
        $vsixFile = Get-ChildItem -Path . -Filter "joss-language-*.vsix" -Recurse -ErrorAction SilentlyContinue | Select-Object -First 1
        
        if ($null -eq $vsixFile) {
            Write-Log "[X] VSIX file not found" "ERROR"
            Write-Log "Please ensure joss-language-2.0.0.vsix is in install folder" "WARNING"
            return $false
        }
        
        Write-Log "Found VSIX: $($vsixFile.Name)" "INFO"
        
        # Install extension
        & code --install-extension $vsixFile.FullName --force 2>&1 | Out-Null
        
        # Verify installation
        Start-Sleep -Seconds 2
        $extensions = & code --list-extensions 2>&1
        if ($extensions -match "joss-language") {
            Write-Log "[OK] Extension installed successfully" "SUCCESS"
            return $true
        } else {
            Write-Log "[X] Extension installation could not be verified" "WARNING"
            return $false
        }
    } catch {
        Write-Log "[X] Extension installation failed: $($_.Exception.Message)" "ERROR"
        return $false
    }
}

# Comprehensive verification
function Test-Installation {
    Write-Host ""
    Write-Log "=== Verifying Installation ===" "INFO"
    Write-Host ""
    
    $allGood = $true
    
    # Verify joss
    try {
        if (Get-Command joss -ErrorAction SilentlyContinue) {
            $version = & joss version 2>&1
            Write-Log "[OK] joss: $version" "SUCCESS"
        } else {
            Write-Log "[X] joss not found in PATH (restart terminal)" "WARNING"
            $allGood = $false
        }
    } catch {
        Write-Log "[X] joss verification failed" "ERROR"
        $allGood = $false
    }
    
    # Verify VS Code
    if (Test-VSCode) {
        # Verify extension
        $extensions = & code --list-extensions 2>&1
        if ($extensions -match "joss-language") {
            Write-Log "[OK] Extension: jossecurity.joss-language@2.0.0" "SUCCESS"
        } else {
            Write-Log "[X] Extension not detected" "WARNING"
            $allGood = $false
        }
    } else {
        $allGood = $false
    }
    
    Write-Host ""
    if ($allGood) {
        Write-Host "=======================================" -ForegroundColor Green
        Write-Host "Installation completed successfully!" -ForegroundColor Green
        Write-Host "=======================================" -ForegroundColor Green
    } else {
        Write-Host "=======================================" -ForegroundColor Yellow
        Write-Host "Installation completed with warnings" -ForegroundColor Yellow
        Write-Host "=======================================" -ForegroundColor Yellow
    }
    Write-Host ""
    Write-Host "Next steps:"
    Write-Host "  1. Restart your terminal"
    Write-Host "  2. Restart VS Code"
    Write-Host "  3. Run: joss version"
    Write-Host ""
    Write-Log "Installation log saved to: $LogFile" "INFO"
}

# Main menu
function Show-MainMenu {
    Write-Host ""
    Write-Host "Select installation option:" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "  [1] Install JosSecurity only" -ForegroundColor White
    Write-Host "  [2] Install VS Code + Extension" -ForegroundColor White
    Write-Host "  [3] Install Everything (Recommended)" -ForegroundColor White
    Write-Host "  [4] Install Extension only" -ForegroundColor White
    Write-Host "  [0] Exit" -ForegroundColor White
    Write-Host ""
    
    $option = Read-Host "Option"
    
    switch ($option) {
        "1" {
            if (Install-JosSecurity) {
                Test-Installation
            }
        }
        "2" {
            if (Test-VSCode) {
                Write-Host "VS Code already installed" -ForegroundColor Yellow
                $reinstall = Read-Host "Reinstall? (y/n)"
                if ($reinstall -eq "y") {
                    Install-VSCode
                }
            } else {
                Install-VSCode
            }
            Install-Extension
            Test-Installation
        }
        "3" {
            $success = $true
            if (-not (Install-JosSecurity)) { $success = $false }
            if (-not (Test-VSCode)) {
                if (-not (Install-VSCode)) { $success = $false }
            }
            if (-not (Install-Extension)) { $success = $false }
            Test-Installation
        }
        "4" {
            if (Test-VSCode) {
                Install-Extension
                Test-Installation
            } else {
                Write-Log "VS Code not installed. Use option 2 or 3" "ERROR"
            }
        }
        "0" {
            Write-Log "Installation cancelled by user" "INFO"
            exit
        }
        default {
            Write-Log "Invalid option" "ERROR"
            Show-MainMenu
        }
    }
}

# Verify permissions
if (-not (Test-Administrator)) {
    Write-Host "WARNING: Not running as Administrator" -ForegroundColor Yellow
    Write-Host "Some functions require elevated permissions" -ForegroundColor Yellow
    Write-Host ""
    $continue = Read-Host "Continue anyway? (y/n)"
    if ($continue -ne "y") {
        Write-Host "Run as Administrator (right-click -> Run as administrator)" -ForegroundColor Yellow
        exit
    }
}

# Execute main menu
Show-MainMenu
