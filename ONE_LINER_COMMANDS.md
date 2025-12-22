# JosSecurity v3.0.3 - Quick Commands

## 🚀 Installation

### Windows (PowerShell as Admin)
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope Process; iwr -useb https://raw.githubusercontent.com/josprox/JosSecurity-language/main/install/remote-install.ps1 | iex
```

### Linux/macOS
```bash
curl -fsSL https://raw.githubusercontent.com/josprox/JosSecurity-language/main/install/remote-install.sh | bash
```

---

## 🔄 Update / 🗑️ Uninstall

Run the **Installation command above**. The script will detect your existing installation and show a menu:

```text
[1] Install
[2] Update
[3] Uninstall
```

Simply select the option you need.

---

## 📝 Manual Installation

If you prefer to download and run locally:

```bash
# Clone repository
git clone https://github.com/josprox/JosSecurity-language.git
cd JosSecurity-language/install

# Run the menu-driven installer
bash remote-install.sh    # Linux/macOS
# or
.\remote-install.ps1   # Windows
```

---

**No dependencies required!** 🎉
