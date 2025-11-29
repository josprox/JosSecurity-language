#!/bin/bash
# JosSecurity Remote Installer - Linux/macOS
# Usage: curl -fsSL https://raw.githubusercontent.com/USER/REPO/main/install/remote-install.sh | bash

set -e

echo "======================================="
echo "   JosSecurity Remote Installer"
echo "======================================="
echo ""

# Configuration
REPO_URL="https://github.com/josprox/JosSecurity-language"
RAW_URL="https://raw.githubusercontent.com/josprox/JosSecurity-language/main/install"
TEMP_DIR="/tmp/jossecurity-install"

# Create temp directory
rm -rf "$TEMP_DIR"
mkdir -p "$TEMP_DIR"
cd "$TEMP_DIR"

echo "[1/5] Downloading installer..."

# Download main installer
curl -fsSL "$RAW_URL/install.sh" -o install.sh
chmod +x install.sh

# Detect OS
OS=$(uname -s)
case "$OS" in
    Linux*)
        BINARY="joss-linux"
        ;;
    Darwin*)
        BINARY="joss-macos"
        ;;
    *)
        echo "Error: Unsupported OS: $OS"
        exit 1
        ;;
esac

echo "[2/5] Downloading JosSecurity binaries..."
curl -fsSL "$REPO_URL/releases/latest/download/jossecurity-binaries.zip" -o jossecurity-binaries.zip

echo "[3/5] Extracting files..."
unzip -o jossecurity-binaries.zip
rm jossecurity-binaries.zip

echo "[4/5] Starting installation..."
echo ""

# Run installer
./install.sh

echo ""
echo "[5/5] Cleaning up..."
cd ~
rm -rf "$TEMP_DIR"

echo "Done!"
