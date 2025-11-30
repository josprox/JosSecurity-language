# -----------------------------------------------------------
# JosSecurity Build Script for Windows PowerShell
# Genera todos los binarios de Go y copia la última extensión VSIX.
# Debe ejecutarse desde la raíz del proyecto (junto a go.mod).
# -----------------------------------------------------------
$ErrorActionPreference = "Stop"

$SourcePackage = "./cmd/joss"
$InstallerDir = "installer"
$VSIXSourceDir = "vscode-joss"

Write-Host "=======================================" -ForegroundColor Blue
Write-Host "  🚀 JosSecurity Go Compilation" -ForegroundColor Blue
Write-Host "=======================================" -ForegroundColor Blue

# 1. Crear el directorio de instalación
if (-not (Test-Path $InstallerDir)) {
    New-Item -ItemType Directory -Path $InstallerDir | Out-Null
}
Write-Host "[INFO] Directorio '$InstallerDir' listo."

# 2. Copiar el último archivo VSIX
Write-Host "[INFO] Buscando el último archivo VSIX en '$VSIXSourceDir'..."
$LatestVSIX = Get-ChildItem -Path $VSIXSourceDir -Filter "*.vsix" | Sort-Object LastWriteTime -Descending | Select-Object -First 1

if ($LatestVSIX) {
    Copy-Item -Path $LatestVSIX.FullName -Destination $InstallerDir -Force
    Write-Host "[SUCCESS] VSIX copiado: $($LatestVSIX.Name)" -ForegroundColor Green
} else {
    Write-Host "[WARNING] No se encontró ningún archivo .vsix. Continuar sin VSIX." -ForegroundColor Yellow
}

# 3. Definir y ejecutar la compilación cruzada
$Targets = @(
    # Windows
    @{GOOS="windows"; GOARCH="amd64"; OutputName="joss.exe"},
    @{GOOS="windows"; GOARCH="arm64"; OutputName="joss-windows-arm64.exe"},
    # Linux
    @{GOOS="linux"; GOARCH="amd64"; OutputName="joss-linux-amd64"},
    @{GOOS="linux"; GOARCH="arm64"; OutputName="joss-linux-arm64"},
    @{GOOS="linux"; GOARCH="arm"; OutputName="joss-linux-armv7"},
    # macOS (Darwin)
    @{GOOS="darwin"; GOARCH="amd64"; OutputName="joss-macos-amd64"},
    @{GOOS="darwin"; GOARCH="arm64"; OutputName="joss-macos-arm64"}
)

Write-Host ""
Write-Host "[INFO] Iniciando compilación de binarios..."

foreach ($Target in $Targets) {
    $GOOS = $Target.GOOS
    $GOARCH = $Target.GOARCH
    $OutputName = $Target.OutputName
    
    # Línea 63 corregida con ${}
    Write-Host "  -> Compilando para ${GOOS}/${GOARCH} ($OutputName)..."

    # Establecer variables de entorno temporalmente
    $env:GOOS = $GOOS
    $env:GOARCH = $GOARCH
    $env:CGO_ENABLED = 0

    # Ejecutar la compilación
    try {
        # Línea corregida (antes línea 67)
        go build -o "$InstallerDir/$OutputName" $SourcePackage
        Write-Host "  [OK] Creado." -ForegroundColor Green
    } catch {
        # Línea 68 corregida con ${}
        Write-Host "  [ERROR] Falló la compilación para ${GOOS}/${GOARCH}: $($_.Exception.Message)" -ForegroundColor Red
    }
    
    # Limpiar variables de entorno
    Remove-Item Env:GOOS
    Remove-Item Env:GOARCH
    Remove-Item Env:CGO_ENABLED
}

# 4. Comprimir los binarios y el VSIX
Write-Host ""
Write-Host "=======================================" -ForegroundColor Cyan
Write-Host "[INFO] 📦 Empaquetando para distribución..." -ForegroundColor Cyan

try {
    $ZipFileName = "jossecurity-binaries.zip"

    # Eliminar archivo viejo si existe
    if (Test-Path $ZipFileName) {
        Remove-Item $ZipFileName -Force | Out-Null
    }
    
    # Comprimir el contenido de la carpeta 'installer' a jossecurity-binaries.zip en la raíz
    # Esto asegura que el ZIP contenga solo los archivos, no la carpeta padre 'installer/'
    Get-ChildItem -Path $InstallerDir -Recurse | Compress-Archive -DestinationPath $ZipFileName -Force
    
    Write-Host "[SUCCESS] Archivo de distribución creado: $ZipFileName" -ForegroundColor Green
    Write-Host "[INFO] ¡Listo para subir a GitHub Releases!" -ForegroundColor Yellow

} catch {
    Write-Host "[ERROR] Falló la compresión del archivo ZIP: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "=======================================" -ForegroundColor Green
Write-Host "✅ Tarea de compilación y empaquetado finalizada." -ForegroundColor Green
Write-Host "=======================================" -ForegroundColor Green