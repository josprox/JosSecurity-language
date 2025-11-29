# Guía Paso a Paso: Instalar Extensión JosSecurity en VS Code

## Problema Actual

La extensión NO está instalada en VS Code. Por eso:
- ❌ No aparece en lista de extensiones
- ❌ No detecta archivos `.joss`
- ❌ Aparece como "texto sin formato"

## Solución: Developer Mode (Método Correcto)

### Paso 1: Cerrar VS Code Completamente

Cierra TODAS las ventanas de VS Code.

### Paso 2: Abrir Carpeta de Extensión

```bash
cd c:\Users\joss\Documents\Proyectos\JosSecurity\vscode-joss
code .
```

**IMPORTANTE**: Debes abrir la carpeta `vscode-joss`, NO el proyecto JosSecurity.

### Paso 3: Verificar que Estás en la Carpeta Correcta

En VS Code, verifica que ves:
- `package.json`
- `src/`
- `out/`
- `.vscode/`

### Paso 4: Presionar F5

1. Presiona **F5** (o Run → Start Debugging)
2. Se abrirá una NUEVA ventana de VS Code con título "[Extension Development Host]"
3. Esta ventana tiene la extensión cargada

### Paso 5: Abrir Proyecto JosSecurity en la Nueva Ventana

En la ventana "[Extension Development Host]":
1. File → Open Folder
2. Navegar a: `c:\Users\joss\Documents\Proyectos\JosSecurity\ejemplos\web`
3. Abrir carpeta

### Paso 6: Abrir Archivo .joss

1. Abrir `main.joss` o `routes.joss`
2. Verificar en la esquina inferior derecha: debería decir "JosSecurity" o "joss"
3. Verificar barra de estado: debería aparecer "$(database) Joss"

## Verificación

### ✅ Funciona Si:
- Archivo .joss se detecta como lenguaje "JosSecurity"
- Aparece "$(database) Joss" en barra de estado
- Hover sobre controladores muestra información
- Ctrl+Shift+P → "Joss" muestra 6 comandos

### ❌ NO Funciona Si:
- Archivo aparece como "Texto sin formato"
- No hay "$(database) Joss" en barra de estado
- No hay syntax highlighting

## Troubleshooting

### "No puedo presionar F5"

**Solución**: 
1. Menú Run → Start Debugging
2. O Ctrl+Shift+D → Presionar botón verde "Run Extension"

### "Dice 'Cannot find module'"

**Solución**:
```bash
cd vscode-joss
npm install
npm run compile
```

### "La nueva ventana no se abre"

**Solución**:
1. Ver Output panel (Ctrl+Shift+U)
2. Seleccionar "Extension Host"
3. Buscar errores

### "Sigue sin detectar .joss"

**Solución**:
1. En ventana Extension Development Host
2. Ctrl+Shift+P → "Developer: Reload Window"
3. Abrir archivo .joss de nuevo

## Método Alternativo: Enlace Simbólico

Si Developer Mode no funciona:

### Windows (PowerShell como Admin)

```powershell
cd $env:USERPROFILE\.vscode\extensions
New-Item -ItemType SymbolicLink -Name "jossecurity.joss-language-2.0.0" -Target "c:\Users\joss\Documents\Proyectos\JosSecurity\vscode-joss"
```

### Linux/macOS

```bash
cd ~/.vscode/extensions
ln -s /ruta/completa/a/vscode-joss jossecurity.joss-language-2.0.0
```

Luego reiniciar VS Code.

## Notas Importantes

1. **NO copiar carpeta** - Usar enlace simbólico o Developer Mode
2. **NO usar `code --install-extension .`** - No funciona sin VSIX
3. **Compilar primero** - Siempre ejecutar `npm run compile`
4. **Ventana correcta** - La extensión solo funciona en Extension Development Host

## Resumen Visual

```
┌─────────────────────────────────────┐
│  Ventana 1: vscode-joss/            │
│  (Aquí presionas F5)                │
└──────────────┬──────────────────────┘
               │ F5
               ▼
┌─────────────────────────────────────┐
│  Ventana 2: [Extension Dev Host]    │
│  (Aquí abres ejemplos/web)          │
│  (Aquí funciona la extensión)       │
└─────────────────────────────────────┘
```

---

**Siguiente paso**: Abrir `vscode-joss/` en VS Code y presionar F5
