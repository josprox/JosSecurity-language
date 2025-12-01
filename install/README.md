# JosSecurity v3.0 - Installation Package

**One-liner installation available!**

## 🚀 Quick Install (Recommended)

### Windows (PowerShell as Admin)
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope Process; iwr -useb https://raw.githubusercontent.com/josprox/JosSecurity-language/main/install/remote-install.ps1 | iex
```

### Linux/macOS
```bash
curl -fsSL https://raw.githubusercontent.com/josprox/JosSecurity-language/main/install/remote-install.sh | bash
```

## 📦 Manual Installation

If you prefer to download and run locally:

### Windows
1. Download `jossecurity-binaries.zip` from the latest GitHub Release.
2. Extract the zip file to this `install` folder.
3. Run the remote installer script locally (it handles everything):
```powershell
.\remote-install.ps1
```

### Linux/macOS
1. Download `jossecurity-binaries.zip` from the latest GitHub Release.
2. Extract the zip file to this `install` folder.
3. Run:
```bash
chmod +x remote-install.sh
./remote-install.sh
```

## ✨ What Gets Installed

- **JosSecurity Compiler** (~15 MB)
  - Installed to system PATH
  - Ready to use: `joss version`

- **VS Code Extension** (29 KB)
  - Syntax highlighting
  - IntelliSense
  - Diagnostics
  - Security analysis

- **VS Code** (optional)
  - Downloaded and installed if not present

## 📋 Package Contents

- `joss.exe` - Windows binary
- `joss-linux` - Linux binary  
- `joss-macos` - macOS binary
- `joss-language-3.0.1.vsix` - VS Code extension
- `remote-install.ps1` - Windows installer/updater/uninstaller
- `remote-install.sh` - Linux/macOS installer/updater/uninstaller

## 🔒 Enhanced Features

### Comprehensive Verification
- ✅ Binary size verification
- ✅ PATH verification
- ✅ Installation verification
- ✅ Extension verification
- ✅ Detailed logging

### Error Handling
- ✅ Try-catch blocks
- ✅ Rollback on failure
- ✅ Clear error messages
- ✅ Installation logs

### Logging
- Windows: `%TEMP%\jossecurity-install.log`
- Linux/macOS: `/tmp/jossecurity-install.log`

## 📝 Installation Options
```bash
# Check version
joss version

# Check extension
code --list-extensions | grep joss

# Create first project
joss new my_app
cd my_app
joss server start
```

## 🐛 Troubleshooting

### Check installation log
**Windows**: `type %TEMP%\jossecurity-install.log`
**Linux/macOS**: `cat /tmp/jossecurity-install.log`

### "joss: command not found"
Restart terminal or reload shell:
```bash
source ~/.bashrc  # Linux/macOS
```

### Permission errors
**Windows**: Run PowerShell as Administrator
**Linux/macOS**: Run with sudo

## 🌐 Remote Installation

The remote installers:
1. Download all required files
2. Run local installer
3. Clean up temporary files
4. Verify installation

**Requirements**:
- Internet connection
- Administrator/sudo privileges
- PowerShell 5.1+ (Windows) or Bash (Linux/macOS)

## 📊 File Sizes

- Windows binary: ~15 MB
- Linux binary: ~15 MB
- macOS binary: ~15 MB
- VS Code extension: 29 KB
- Installers: 2 scripts (Unified Install/Update/Uninstall)
- **Total**: ~45 MB

## 🆘 Support

- Documentation: `../docs/`
- Version: 3.0.1
- License: MIT

---

**Ready to install?** Choose your method above! 🚀
