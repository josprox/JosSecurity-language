#!/bin/bash
# JosSecurity Uninstaller v1.0 - Linux/macOS
# Selective uninstallation

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Variables
INSTALL_DIR="/usr/local/bin"
LOG_FILE="/tmp/jossecurity-uninstall.log"

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
echo -e "${RED}"
echo "======================================="
echo "   JosSecurity Uninstaller v1.0"
echo "======================================="
echo -e "${NC}"
log INFO "Uninstallation started"

# Uninstall JosSecurity
uninstall_jossecurity() {
    log INFO "Uninstalling JosSecurity..."
    
    # Remove binary
    if [ -f "$INSTALL_DIR/joss" ]; then
        sudo rm "$INSTALL_DIR/joss"
        log SUCCESS "[OK] Binary removed"
    else
        log INFO "[OK] Binary not found"
    fi
    
    # Remove from PATH (if added to shell config)
    for rc in ~/.bashrc ~/.zshrc ~/.profile; do
        if [ -f "$rc" ] && grep -q "JosSecurity" "$rc"; then
            sed -i.bak '/JosSecurity/d' "$rc"
            log SUCCESS "[OK] Removed from $rc"
        fi
    done
    
    log SUCCESS "[OK] JosSecurity uninstalled"
}

# Uninstall VS Code Extension
uninstall_extension() {
    log INFO "Uninstalling VS Code extension..."
    
    if command -v code &> /dev/null; then
        if code --list-extensions | grep -q "joss-language"; then
            code --uninstall-extension jossecurity.joss-language &> /dev/null
            
            # Verify
            sleep 2
            if ! code --list-extensions | grep -q "joss-language"; then
                log SUCCESS "[OK] Extension uninstalled"
            else
                log WARNING "[X] Extension still present"
            fi
        else
            log INFO "[OK] Extension not installed"
        fi
    else
        log WARNING "[X] VS Code not found"
    fi
}

# Uninstall VS Code
uninstall_vscode() {
    log WARNING "Uninstalling VS Code..."
    echo ""
    echo -e "${YELLOW}WARNING: This will uninstall VS Code completely!${NC}"
    echo -e "${YELLOW}All your VS Code settings and extensions will be removed.${NC}"
    echo ""
    read -p "Are you sure? (yes/no): " confirm
    
    if [[ $confirm != "yes" ]]; then
        log INFO "VS Code uninstallation cancelled"
        return
    fi
    
    # Detect OS
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        if command -v apt-get &> /dev/null; then
            sudo apt-get remove -y code
            sudo apt-get autoremove -y
        elif command -v dnf &> /dev/null; then
            sudo dnf remove -y code
        fi
        log SUCCESS "[OK] VS Code uninstalled"
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        if command -v brew &> /dev/null; then
            brew uninstall --cask visual-studio-code
            log SUCCESS "[OK] VS Code uninstalled"
        else
            log ERROR "[X] Homebrew not found"
        fi
    fi
}

# Main menu
uninstall_menu() {
    echo ""
    echo "What do you want to uninstall?"
    echo ""
    echo "  [1] Uninstall JosSecurity only"
    echo "  [2] Uninstall VS Code Extension only"
    echo "  [3] Uninstall JosSecurity + Extension"
    echo "  [4] Uninstall Everything (JosSecurity + Extension + VS Code)"
    echo "  [0] Cancel"
    echo ""
    read -p "Option: " option
    
    case $option in
        1)
            uninstall_jossecurity
            ;;
        2)
            uninstall_extension
            ;;
        3)
            uninstall_jossecurity
            uninstall_extension
            ;;
        4)
            uninstall_jossecurity
            uninstall_extension
            uninstall_vscode
            ;;
        0)
            log INFO "Uninstallation cancelled"
            exit 0
            ;;
        *)
            log ERROR "Invalid option"
            uninstall_menu
            ;;
    esac
    
    echo ""
    echo -e "${GREEN}=======================================${NC}"
    echo -e "${GREEN}Uninstallation completed${NC}"
    echo -e "${GREEN}=======================================${NC}"
    echo ""
    log INFO "Log saved to: $LOG_FILE"
}

# Execute
uninstall_menu
