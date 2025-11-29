# Reinstalación Manual de Extensión JosSecurity v2.0

## Pasos Realizados

1. ✅ Limpieza de archivos antiguos (`rm -rf out/`)
2. ✅ Recompilación TypeScript (`npm run compile`)
3. ✅ Eliminación de extensión antigua
4. ✅ Copia directa a directorio de extensiones

## Ubicación de la Extensión

```
$HOME/.vscode/extensions/jossecurity.joss-language-2.0.0/
```

## Verificación

Para verificar que la extensión v2.0 está instalada:

```bash
code --list-extensions --show-versions | grep joss
```

Debería mostrar: `jossecurity.joss-language@2.0.0`

## Activación

1. **Reiniciar VS Code** completamente
2. Abrir un archivo `.joss`
3. Verificar en la barra de estado: `$(database) Joss`
4. Verificar Output panel: "JosSecurity Language Server"

## Probar Funcionalidades

### Test 1: Hover
Abrir `ejemplos/web/routes.joss` y hacer hover sobre un controlador

### Test 2: Comandos
`Ctrl+Shift+P` → Buscar "Joss" → Deberían aparecer 6 comandos

### Test 3: Diagnósticos
Escribir una ruta con controlador inexistente → Debería mostrar error

## Si No Funciona

### Opción 1: Reinstalar desde Carpeta
```bash
cd vscode-joss
code --install-extension . --force
```

### Opción 2: Desarrollo Mode
```bash
cd vscode-joss
code .
# Presionar F5 para abrir Extension Development Host
```

### Opción 3: Verificar Logs
1. Abrir Output panel (Ctrl+Shift+U)
2. Seleccionar "JosSecurity Language Server"
3. Buscar errores

## Archivos Compilados

La extensión v2.0 incluye:
- `out/extension.js` - Cliente LSP
- `out/server/server.js` - Servidor LSP
- `out/server/parser/routeParser.js`
- `out/server/indexer/indexer.js`
- `out/server/analyzer/securityAnalyzer.js`

Todos estos archivos deben existir en `out/` para que funcione.
