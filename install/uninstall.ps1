# JosSecurity Uninstaller v1.0 - Windows PowerShell
# Selective uninstallation

$ErrorActionPreference = "Stop"
$Host.UI.RawUI.ForegroundColor = "White"

# Variables
$InstallDir = "C:\Program Files\JosSecurity"
$LogFile = "$env:TEMP\jossecurity-uninstall.log"

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
Write-Host "=======================================" -ForegroundColor Red
Write-Host "   JosSecurity Uninstaller v1.0        " -ForegroundColor Red
Write-Host "=======================================" -ForegroundColor Red
Write-Host ""
Write-Log "Uninstallation started" "INFO"

# Uninstall JosSecurity
function Uninstall-JosSecurity {
    Write-Log "Uninstalling JosSecurity..." "INFO"
    
    try {
        # Remove binary
        if (Test-Path "$InstallDir\joss.exe") {
            Remove-Item "$InstallDir\joss.exe" -Force
            Write-Log "[OK] Binary removed" "SUCCESS"
        }
        
        # Remove directory
        if (Test-Path $InstallDir) {
            Remove-Item $InstallDir -Recurse -Force
            Write-Log "[OK] Installation directory removed" "SUCCESS"
        }
        
        # Remove from PATH
        $currentPath = [Environment]::GetEnvironmentVariable("Path", "Machine")
        if ($currentPath -like "*$InstallDir*") {
            $newPath = $currentPath -replace [regex]::Escape(";$InstallDir"), ""
            $newPath = $newPath -replace [regex]::Escape("$InstallDir;"), ""
            [Environment]::SetEnvironmentVariable("Path", $newPath, "Machine")
            Write-Log "[OK] Removed from PATH" "SUCCESS"
        }
        
        Write-Log "[OK] JosSecurity uninstalled" "SUCCESS"
        return $true
    } catch {
        Write-Log "[X] Error: $($_.Exception.Message)" "ERROR"
        return $false
    }
}

# Uninstall VS Code Extension
function Uninstall-Extension {
    Write-Log "Uninstalling VS Code extension..." "INFO"
    
    try {
        $extensions = & code --list-extensions 2>&1
        if ($extensions -match "joss-language") {
            & code --uninstall-extension jossecurity.joss-language 2>&1 | Out-Null
            
            # Verify
            Start-Sleep -Seconds 2
            $extensions = & code --list-extensions 2>&1
            if ($extensions -notmatch "joss-language") {
                Write-Log "[OK] Extension uninstalled" "SUCCESS"
                return $true
            } else {
                Write-Log "[X] Extension still present" "WARNING"
                return $false
            }
        } else {
            Write-Log "[OK] Extension not installed" "INFO"
            return $true
        }
    } catch {
        Write-Log "[X] Error: $($_.Exception.Message)" "ERROR"
        return $false
    }
}

# Uninstall VS Code
function Uninstall-VSCode {
    Write-Log "Uninstalling VS Code..." "WARNING"
    Write-Host ""
    Write-Host "WARNING: This will uninstall VS Code completely!" -ForegroundColor Yellow
    Write-Host "All your VS Code settings and extensions will be removed." -ForegroundColor Yellow
    Write-Host ""
    $confirm = Read-Host "Are you sure? (yes/no)"
    
    if ($confirm -ne "yes") {
        Write-Log "VS Code uninstallation cancelled" "INFO"
        return $false
    }
    
    try {
        # Find VS Code uninstaller
        $uninstallers = @(
            "$env:LOCALAPPDATA\Programs\Microsoft VS Code\unins000.exe",
            "$env:ProgramFiles\Microsoft VS Code\unins000.exe",
            "$env:ProgramFiles(x86)\Microsoft VS Code\unins000.exe"
        )
        
        $uninstaller = $null
        foreach ($path in $uninstallers) {
            if (Test-Path $path) {
                $uninstaller = $path
                break
            }
        }
        
        if ($uninstaller) {
            Write-Log "Running VS Code uninstaller..." "INFO"
            Start-Process -FilePath $uninstaller -ArgumentList "/VERYSILENT" -Wait
            Write-Log "[OK] VS Code uninstalled" "SUCCESS"
            return $true
        } else {
            Write-Log "[X] VS Code uninstaller not found" "ERROR"
            return $false
        }
    } catch {
        Write-Log "[X] Error: $($_.Exception.Message)" "ERROR"
        return $false
    }
}

# Main menu
function Show-UninstallMenu {
    Write-Host ""
    Write-Host "What do you want to uninstall?" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "  [1] Uninstall JosSecurity only" -ForegroundColor White
    Write-Host "  [2] Uninstall VS Code Extension only" -ForegroundColor White
    Write-Host "  [3] Uninstall JosSecurity + Extension" -ForegroundColor White
    Write-Host "  [4] Uninstall Everything (JosSecurity + Extension + VS Code)" -ForegroundColor White
    Write-Host "  [0] Cancel" -ForegroundColor White
    Write-Host ""
    
    $option = Read-Host "Option"
    
    switch ($option) {
        "1" {
            Uninstall-JosSecurity
        }
        "2" {
            Uninstall-Extension
        }
        "3" {
            Uninstall-JosSecurity
            Uninstall-Extension
        }
        "4" {
            Uninstall-JosSecurity
            Uninstall-Extension
            Uninstall-VSCode
        }
        "0" {
            Write-Log "Uninstallation cancelled" "INFO"
            exit
        }
        default {
            Write-Log "Invalid option" "ERROR"
            Show-UninstallMenu
        }
    }
    
    Write-Host ""
    Write-Host "=======================================" -ForegroundColor Green
    Write-Host "Uninstallation completed" -ForegroundColor Green
    Write-Host "=======================================" -ForegroundColor Green
    Write-Host ""
    Write-Log "Log saved to: $LogFile" "INFO"
}

# Verify admin
$currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
$principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
if (-not $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
    Write-Host "WARNING: Not running as Administrator" -ForegroundColor Yellow
    Write-Host "Some operations may fail" -ForegroundColor Yellow
    Write-Host ""
    $continue = Read-Host "Continue anyway? (y/n)"
    if ($continue -ne "y") {
        exit
    }
}

# Execute
Show-UninstallMenu
