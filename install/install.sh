#!/bin/bash
# JosSecurity Installer v1.0 - Linux/macOS
# Enhanced with comprehensive verification

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Variables
JOSS_VERSION="3.0.0"
INSTALL_DIR="/usr/local/bin"
LOG_FILE="/tmp/jossecurity-install.log"

# Logging function
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
echo "   JosSecurity Installer v1.0"
echo "   Enhanced Installation System"
echo "======================================="
echo -e "${NC}"
log INFO "Installation started"

# Detect VS Code
detect_vscode() {
    if command -v code &> /dev/null; then
        log SUCCESS "[OK] VS Code detected"
        return 0
    else
        log WARNING "[X] VS Code not detected"
        return 1
    fi
}

# Install JosSecurity with verification
install_jossecurity() {
    log INFO "[1/3] Installing JosSecurity..."
    
    # Detect OS and select binary
    OS=$(uname -s)
    case "$OS" in
        Linux*)
            BINARY="joss-linux"
            ;;
        Darwin*)
            BINARY="joss-macos"
            ;;
        *)
            log ERROR "[X] Unsupported OS: $OS"
            return 1
            ;;
    esac
    
    # Check if binary exists
    if [ ! -f "$BINARY" ]; then
        log ERROR "[X] Binary $BINARY not found"
        log WARNING "Please ensure pre-compiled binaries are in install folder"
        return 1
    fi
    
    log INFO "Found binary for $OS: $BINARY"
    
    # Get binary size
    SIZE=$(du -h "$BINARY" | cut -f1)
    log INFO "Binary size: $SIZE"
    
    # Copy to /usr/local/bin
    log INFO "Installing to $INSTALL_DIR..."
    if sudo cp "$BINARY" "$INSTALL_DIR/joss"; then
        sudo chmod +x "$INSTALL_DIR/joss"
        
        # Verify installation
        if [ -f "$INSTALL_DIR/joss" ] && [ -x "$INSTALL_DIR/joss" ]; then
            log SUCCESS "[OK] Binary installed and executable"
        else
            log ERROR "[X] Binary installation verification failed"
            return 1
        fi
    else
        log ERROR "[X] Failed to copy binary"
        return 1
    fi
    
    # Verify PATH
    if echo "$PATH" | grep -q "$INSTALL_DIR"; then
        log SUCCESS "[OK] Already in PATH"
    else
        log INFO "Adding to PATH..."
        
        # Detect shell
        if [ -n "$ZSH_VERSION" ]; then
            SHELL_RC="$HOME/.zshrc"
        elif [ -n "$BASH_VERSION" ]; then
            SHELL_RC="$HOME/.bashrc"
        else
            SHELL_RC="$HOME/.profile"
        fi
        
        echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$SHELL_RC"
        log SUCCESS "[OK] PATH updated in $SHELL_RC"
    fi
    
    log SUCCESS "[OK] JosSecurity $JOSS_VERSION installed"
    return 0
}

# Install VS Code with verification
install_vscode() {
    log INFO "[2/3] Installing VS Code..."
    
    # Detect OS
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Linux
        if command -v apt-get &> /dev/null; then
            # Debian/Ubuntu
            log INFO "Detected Debian/Ubuntu"
            wget -qO- https://packages.microsoft.com/keys/microsoft.asc | gpg --dearmor > packages.microsoft.gpg
            sudo install -D -o root -g root -m 644 packages.microsoft.gpg /etc/apt/keyrings/packages.microsoft.gpg
            sudo sh -c 'echo "deb [arch=amd64,arm64,armhf signed-by=/etc/apt/keyrings/packages.microsoft.gpg] https://packages.microsoft.com/repos/code stable main" > /etc/apt/sources.list.d/vscode.list'
            rm -f packages.microsoft.gpg
            sudo apt-get update
            sudo apt-get install -y code
        elif command -v dnf &> /dev/null; then
            # Fedora/RHEL
            log INFO "Detected Fedora/RHEL"
            sudo rpm --import https://packages.microsoft.com/keys/microsoft.asc
            sudo sh -c 'echo -e "[code]\nname=Visual Studio Code\nbaseurl=https://packages.microsoft.com/yumrepos/vscode\nenabled=1\ngpgcheck=1\ngpgkey=https://packages.microsoft.com/keys/microsoft.asc" > /etc/yum.repos.d/vscode.repo'
            sudo dnf check-update
            sudo dnf install -y code
        fi
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        log INFO "Detected macOS"
        if command -v brew &> /dev/null; then
            brew install --cask visual-studio-code
        else
            log ERROR "Homebrew not installed"
            log WARNING "Install from https://code.visualstudio.com"
            return 1
        fi
    fi
    
    # Verify installation
    sleep 2
    if detect_vscode; then
        log SUCCESS "[OK] VS Code installed successfully"
        return 0
    else
        log WARNING "[X] VS Code installation could not be verified"
        return 1
    fi
}

# Install extension with verification
install_extension() {
    log INFO "[3/3] Installing VS Code extension..."
    
    # Find VSIX file
    VSIX_FILE=$(find . -name "joss-language-*.vsix" -type f | head -n 1)
    
    if [ -z "$VSIX_FILE" ]; then
        log ERROR "[X] VSIX file not found"
        log WARNING "Ensure joss-language-2.0.0.vsix is in install folder"
        return 1
    fi
    
    log INFO "Found VSIX: $(basename $VSIX_FILE)"
    
    # Get VSIX size
    SIZE=$(du -h "$VSIX_FILE" | cut -f1)
    log INFO "VSIX size: $SIZE"
    
    # Install extension
    if code --install-extension "$VSIX_FILE" --force &> /dev/null; then
        # Verify installation
        sleep 2
        if code --list-extensions | grep -q "joss-language"; then
            log SUCCESS "[OK] Extension installed successfully"
            return 0
        else
            log WARNING "[X] Extension installation could not be verified"
            return 1
        fi
    else
        log ERROR "[X] Extension installation failed"
        return 1
    fi
}

# Comprehensive verification
verify_installation() {
    echo ""
    log INFO "=== Verifying Installation ==="
    echo ""
    
    ALL_GOOD=true
    
    # Verify joss
    if command -v joss &> /dev/null; then
        VERSION=$(joss version 2>&1 || echo "unknown")
        log SUCCESS "[OK] joss: $VERSION"
    else
        log WARNING "[X] joss not found (restart terminal)"
        ALL_GOOD=false
    fi
    
    # Verify VS Code
    if detect_vscode; then
        # Verify extension
        if code --list-extensions | grep -q "joss-language"; then
            log SUCCESS "[OK] Extension: jossecurity.joss-language@2.0.0"
        else
            log WARNING "[X] Extension not detected"
            ALL_GOOD=false
        fi
    else
        ALL_GOOD=false
    fi
    
    echo ""
    if [ "$ALL_GOOD" = true ]; then
        echo -e "${GREEN}=======================================${NC}"
        echo -e "${GREEN}Installation completed successfully!${NC}"
        echo -e "${GREEN}=======================================${NC}"
    else
        echo -e "${YELLOW}=======================================${NC}"
        echo -e "${YELLOW}Installation completed with warnings${NC}"
        echo -e "${YELLOW}=======================================${NC}"
    fi
    echo ""
    echo "Next steps:"
    echo "  1. Restart your terminal: source ~/.bashrc"
    echo "  2. Restart VS Code"
    echo "  3. Run: joss version"
    echo ""
    log INFO "Installation log saved to: $LOG_FILE"
}

# Main menu
main_menu() {
    echo ""
    echo "Select installation option:"
    echo ""
    echo "  [1] Install JosSecurity only"
    echo "  [2] Install VS Code + Extension"
    echo "  [3] Install Everything (Recommended)"
    echo "  [4] Install Extension only"
    echo "  [0] Exit"
    echo ""
    read -p "Option: " option
    
    case $option in
        1)
            if install_jossecurity; then
                verify_installation
            fi
            ;;
        2)
            if detect_vscode; then
                echo -e "${YELLOW}VS Code already installed${NC}"
                read -p "Reinstall? (y/n): " reinstall
                if [[ $reinstall == "y" ]]; then
                    install_vscode
                fi
            else
                install_vscode
            fi
            install_extension
            verify_installation
            ;;
        3)
            SUCCESS=true
            if ! install_jossecurity; then SUCCESS=false; fi
            if ! detect_vscode; then
                if ! install_vscode; then SUCCESS=false; fi
            fi
            if ! install_extension; then SUCCESS=false; fi
            verify_installation
            ;;
        4)
            if detect_vscode; then
                install_extension
                verify_installation
            else
                log ERROR "VS Code not installed. Use option 2 or 3"
            fi
            ;;
        0)
            log INFO "Installation cancelled by user"
            exit 0
            ;;
        *)
            log ERROR "Invalid option"
            main_menu
            ;;
    esac
}

# Execute main menu
main_menu
