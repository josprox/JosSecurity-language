# Subir JosSecurity a GitHub

## Repositorio
**URL**: https://github.com/josprox/JosSecurity-language

## Pasos para Subir

### 1. Inicializar Git (si no está inicializado)

```bash
cd c:\Users\joss\Documents\Proyectos\JosSecurity
git init
```

### 2. Configurar Remote

```bash
git remote add origin https://github.com/josprox/JosSecurity-language.git
```

### 3. Crear .gitignore

Ya existe, pero verifica que incluya:
```
# Build artifacts
*.exe
joss
joss-linux
joss-macos

# Node modules
node_modules/
out/

# Logs
*.log

# OS files
.DS_Store
Thumbs.db

# IDE
.vscode/
.idea/

# Temporary files
*.tmp
*.temp
```

### 4. Agregar Archivos

```bash
# Agregar todo excepto lo ignorado
git add .

# Verificar qué se agregará
git status
```

### 5. Crear Commit

```bash
git commit -m "feat: JosSecurity v3.0 - Complete installation package

- Pre-compiled binaries for Windows, Linux, macOS
- VS Code extension v2.0 with LSP
- Enhanced installers with verification
- One-liner remote installation
- Zero dependencies installation"
```

### 6. Subir a GitHub

```bash
# Primera vez
git push -u origin main

# O si ya existe
git push origin main
```

## Archivos Importantes a Subir

### Carpeta `install/` (CRÍTICO)
- ✅ `joss.exe` (~15 MB)
- ✅ `joss-linux` (~15 MB)
- ✅ `joss-macos` (~15 MB)
- ✅ `joss-language-2.0.0.vsix` (29 KB)
- ✅ `install.ps1`
- ✅ `install.sh`
- ✅ `remote-install.ps1`
- ✅ `remote-install.sh`
- ✅ `README.md`

### Código Fuente
- ✅ `cmd/` - Código del compilador
- ✅ `pkg/` - Paquetes Go
- ✅ `docs/` - Documentación
- ✅ `ejemplos/` - Ejemplos de código
- ✅ `vscode-joss/` - Código fuente de extensión

### Documentación
- ✅ `README.md`
- ✅ `LICENSE`
- ✅ `docs/`

## NO Subir

- ❌ `node_modules/`
- ❌ `out/`
- ❌ `*.log`
- ❌ `.DS_Store`
- ❌ Archivos temporales

## Verificar Después de Subir

### 1. Verificar que los archivos estén en GitHub

Visita: https://github.com/josprox/JosSecurity-language/tree/main/install

Deberías ver:
- joss.exe
- joss-linux
- joss-macos
- joss-language-2.0.0.vsix
- install.ps1
- install.sh
- remote-install.ps1
- remote-install.sh
- README.md

### 2. Probar Instalación Remota

**Windows**:
```powershell
iwr -useb https://raw.githubusercontent.com/josprox/JosSecurity-language/main/install/remote-install.ps1 | iex
```

**Linux/macOS**:
```bash
curl -fsSL https://raw.githubusercontent.com/josprox/JosSecurity-language/main/install/remote-install.sh | bash
```

### 3. Verificar URLs Raw

Verifica que estos archivos sean accesibles:
- https://raw.githubusercontent.com/josprox/JosSecurity-language/main/install/joss.exe
- https://raw.githubusercontent.com/josprox/JosSecurity-language/main/install/joss-linux
- https://raw.githubusercontent.com/josprox/JosSecurity-language/main/install/joss-macos
- https://raw.githubusercontent.com/josprox/JosSecurity-language/main/install/joss-language-2.0.0.vsix

## Crear Release (Opcional)

### 1. Crear Tag

```bash
git tag -a v3.0.0 -m "JosSecurity v3.0.0 - Complete Installation Package"
git push origin v3.0.0
```

### 2. Crear Release en GitHub

1. Ir a: https://github.com/josprox/JosSecurity-language/releases/new
2. Tag: `v3.0.0`
3. Title: `JosSecurity v3.0.0`
4. Description:
```markdown
# JosSecurity v3.0.0

## One-Liner Installation

**Windows**:
```powershell
iwr -useb https://raw.githubusercontent.com/josprox/JosSecurity-language/main/install/remote-install.ps1 | iex
```

**Linux/macOS**:
```bash
curl -fsSL https://raw.githubusercontent.com/josprox/JosSecurity-language/main/install/remote-install.sh | bash
```

## Features

- ✅ Pre-compiled binaries (Windows, Linux, macOS)
- ✅ VS Code extension with LSP
- ✅ Zero dependencies installation
- ✅ Enhanced installers with verification
- ✅ One-liner remote installation

## Downloads

- Windows: `joss.exe` (~15 MB)
- Linux: `joss-linux` (~15 MB)
- macOS: `joss-macos` (~15 MB)
- VS Code Extension: `joss-language-2.0.0.vsix` (29 KB)
```

5. Adjuntar archivos:
   - `joss.exe`
   - `joss-linux`
   - `joss-macos`
   - `joss-language-2.0.0.vsix`

## Actualizar README Principal

Agrega al README.md principal:

```markdown
## Installation

### Quick Install (One-Liner)

**Windows (PowerShell as Admin)**:
```powershell
iwr -useb https://raw.githubusercontent.com/josprox/JosSecurity-language/main/install/remote-install.ps1 | iex
```

**Linux/macOS**:
```bash
curl -fsSL https://raw.githubusercontent.com/josprox/JosSecurity-language/main/install/remote-install.sh | bash
```

### Manual Install

Download the [latest release](https://github.com/josprox/JosSecurity-language/releases) or clone the repository:

```bash
git clone https://github.com/josprox/JosSecurity-language.git
cd JosSecurity-language/install
./install.sh  # Linux/macOS
# or
.\install.ps1  # Windows
```
```

## Checklist Final

- [ ] Git inicializado
- [ ] Remote configurado
- [ ] .gitignore verificado
- [ ] Archivos agregados
- [ ] Commit creado
- [ ] Push a GitHub
- [ ] Verificar archivos en GitHub
- [ ] Probar instalación remota
- [ ] Crear release (opcional)
- [ ] Actualizar README principal

---

**Listo para distribuir!** 🚀
