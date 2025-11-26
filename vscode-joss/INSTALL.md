# Guía de Instalación - Extensión JosSecurity para VS Code

## Método 1: Instalación Manual (Recomendado)

### Windows
1. Abre el Explorador de Archivos
2. Navega a: `%USERPROFILE%\.vscode\extensions\`
   - O pega en la barra de direcciones: `C:\Users\TU_USUARIO\.vscode\extensions\`
3. Copia la carpeta `vscode-joss` completa a esa ubicación
4. Reinicia VS Code
5. Abre cualquier archivo `.joss` para verificar

### macOS/Linux
```bash
cp -r vscode-joss ~/.vscode/extensions/
code --list-extensions | grep joss
```

## Método 2: Empaquetar como VSIX

### Requisitos
```bash
npm install -g @vscode/vsce
```

### Crear VSIX
```bash
cd vscode-joss
vsce package
```

Esto creará `joss-language-1.0.0.vsix`

### Instalar VSIX
1. En VS Code: `Ctrl+Shift+P` (o `Cmd+Shift+P` en Mac)
2. Escribe: "Extensions: Install from VSIX"
3. Selecciona el archivo `.vsix`

## Verificación

1. Abre VS Code
2. Crea un archivo `test.joss`
3. Escribe:
```joss
class Main {
    Init main() {
        print("Test")
    }
}
```

4. Verifica que:
   - ✅ `class`, `Init` están en morado/rosa
   - ✅ `print` está en amarillo
   - ✅ Los strings están en naranja
   - ✅ Los comentarios `//` están en verde

## Snippets de Prueba

Escribe y presiona `Tab`:
- `class` → Genera clase completa
- `tern` → Genera ternario
- `foreach` → Genera bucle

## Solución de Problemas

### La extensión no aparece
```bash
# Verificar instalación
code --list-extensions | grep joss

# Si no aparece, reinstalar
rm -rf ~/.vscode/extensions/vscode-joss
cp -r vscode-joss ~/.vscode/extensions/
```

### Syntax highlighting no funciona
1. Verifica que el archivo tenga extensión `.joss`
2. Presiona `Ctrl+Shift+P` → "Change Language Mode" → "JosSecurity"
3. Reinicia VS Code

### Snippets no funcionan
1. Verifica que estés en un archivo `.joss`
2. Presiona `Ctrl+Space` para ver sugerencias
3. Escribe el prefijo completo antes de `Tab`

## Desinstalación

```bash
# Windows
rmdir /s %USERPROFILE%\.vscode\extensions\vscode-joss

# macOS/Linux
rm -rf ~/.vscode/extensions/vscode-joss
```

## Actualización

1. Elimina la versión anterior
2. Copia la nueva versión
3. Reinicia VS Code
