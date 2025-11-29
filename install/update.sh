#!/bin/bash
# JosSecurity Updater v1.0 - Linux/macOS
# Automatic update checker and installer

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Variables
REPO_URL="https://api.github.com/repos/josprox/JosSecurity-language/releases/latest"
INSTALL_DIR="/usr/local/bin"
CURRENT_VERSION="3.0.0"
LOG_FILE="/tmp/jossecurity-update.log"

# Logging
log() {
    local level=$1
    shift
    local message="$@"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[$timestamp] [$level] $message" >> "$LOG_FILE"
    
    case $level in
        ERROR)
            echo -e "${RED}$message${NC}"
            ;;
        SUCCESS)
            echo -e "${GREEN}$message${NC}"
            ;;
        WARNING)
            echo -e "${YELLOW}$message${NC}"
            ;;
        *)
            echo -e "$message"
            ;;
    esac
}

# Banner
echo -e "${BLUE}"
echo "======================================="
echo "   JosSecurity Updater v1.0"
echo "======================================="
echo -e "${NC}"
log INFO "Update check started"

# Detect OS
OS=$(uname -s)
case "$OS" in
    Linux*)
        BINARY_NAME="joss-linux"
        ;;
    Darwin*)
        BINARY_NAME="joss-macos"
        ;;
    *)
        log ERROR "Unsupported OS: $OS"
        exit 1
        ;;
esac

# Check for updates
check_update() {
    log INFO "Checking for updates..."
    log INFO "Current version: $CURRENT_VERSION"
    
    # Get latest release
    RELEASE_INFO=$(curl -s "$REPO_URL")
    LATEST_VERSION=$(echo "$RELEASE_INFO" | grep '"tag_name"' | sed -E 's/.*"v?([^"]+)".*/\1/')
    
    log INFO "Latest version: $LATEST_VERSION"
    
    # Compare versions
    if [ "$(printf '%s\n' "$LATEST_VERSION" "$CURRENT_VERSION" | sort -V | head -n1)" != "$LATEST_VERSION" ]; then
        log WARNING "[!] Update available: $LATEST_VERSION"
        
        # Get download URLs
        BINARY_URL=$(echo "$RELEASE_INFO" | grep "browser_download_url.*$BINARY_NAME" | cut -d '"' -f 4)
        EXTENSION_URL=$(echo "$RELEASE_INFO" | grep "browser_download_url.*joss-language.*vsix" | cut -d '"' -f 4)
        
        return 0  # Update available
    else
        log SUCCESS "[OK] You have the latest version"
        return 1  # No update
    fi
}

# Update JosSecurity
update_jossecurity() {
    log INFO "Updating JosSecurity..."
    
    # Download new binary
    TEMP_FILE="/tmp/joss-new"
    log INFO "Downloading version $LATEST_VERSION..."
    curl -fsSL "$BINARY_URL" -o "$TEMP_FILE"
    
    # Backup current version
    if [ -f "$INSTALL_DIR/joss" ]; then
        sudo cp "$INSTALL_DIR/joss" "$INSTALL_DIR/joss.bak"
        log INFO "Backup created"
    fi
    
    # Replace binary
    sudo cp "$TEMP_FILE" "$INSTALL_DIR/joss"
    sudo chmod +x "$INSTALL_DIR/joss"
    rm "$TEMP_FILE"
    
    # Verify
    if [ -f "$INSTALL_DIR/joss" ] && [ -x "$INSTALL_DIR/joss" ]; then
        log SUCCESS "[OK] JosSecurity updated to $LATEST_VERSION"
    else
        log ERROR "[X] Update verification failed"
        
        # Restore backup
        if [ -f "$INSTALL_DIR/joss.bak" ]; then
            sudo cp "$INSTALL_DIR/joss.bak" "$INSTALL_DIR/joss"
            log WARNING "Restored from backup"
        fi
    fi
}

# Update Extension
update_extension() {
    log INFO "Updating VS Code extension..."
    
    # Download new VSIX
    TEMP_FILE="/tmp/joss-language-new.vsix"
    log INFO "Downloading extension..."
    curl -fsSL "$EXTENSION_URL" -o "$TEMP_FILE"
    
    # Install extension
    code --install-extension "$TEMP_FILE" --force &> /dev/null
    rm "$TEMP_FILE"
    
    # Verify
    sleep 2
    if code --list-extensions | grep -q "joss-language"; then
        log SUCCESS "[OK] Extension updated"
    else
        log ERROR "[X] Extension update verification failed"
    fi
}

# Main update flow
if check_update; then
    echo ""
    echo -e "${GREEN}Update available: v$LATEST_VERSION${NC}"
    echo ""
    echo "What do you want to update?"
    echo ""
    echo "  [1] Update JosSecurity only"
    echo "  [2] Update VS Code Extension only"
    echo "  [3] Update Everything (Recommended)"
    echo "  [0] Cancel"
    echo ""
    read -p "Option: " option
    
    case $option in
        1)
            update_jossecurity
            ;;
        2)
            update_extension
            ;;
        3)
            update_jossecurity
            update_extension
            ;;
        0)
            log INFO "Update cancelled"
            exit 0
            ;;
        *)
            log ERROR "Invalid option"
            exit 1
            ;;
    esac
    
    echo ""
    echo -e "${GREEN}=======================================${NC}"
    echo -e "${GREEN}Update completed${NC}"
    echo -e "${GREEN}=======================================${NC}"
    echo ""
    echo -e "${YELLOW}Please restart terminal and VS Code${NC}"
else
    echo ""
    echo -e "${GREEN}No updates available${NC}"
fi

log INFO "Log saved to: $LOG_FILE"
