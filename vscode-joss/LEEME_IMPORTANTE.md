# ⚠️ IMPORTANTE: INSTALACIÓN DE EXTENSIÓN VS CODE

## 🔴 REGLA DE ORO

**SIEMPRE instalar la extensión usando VSIX**

```bash
# ✅ CORRECTO
npx vsce package
code --install-extension joss-language-2.0.0.vsix

# ❌ INCORRECTO - NO USAR
code --install-extension .
cp -r . ~/.vscode/extensions/
```

## ¿Por Qué VSIX?

1. **Método Oficial**: Es el método estándar de VS Code
2. **Empaquetado Correcto**: Incluye solo archivos necesarios
3. **Instalación Limpia**: VS Code gestiona correctamente la extensión
4. **Actualizaciones**: Fácil de actualizar y desinstalar
5. **Distribución**: Mismo archivo para todos los usuarios

## Proceso Correcto

### 1. Compilar TypeScript
```bash
cd vscode-joss
npm run compile
```

### 2. Empaquetar VSIX
```bash
npx vsce package --allow-star-activation
```

Esto genera: `joss-language-2.0.0.vsix`

### 3. Instalar VSIX
```bash
code --install-extension joss-language-2.0.0.vsix --force
```

### 4. Reiniciar VS Code
Cerrar y volver a abrir VS Code completamente.

## Verificación

```bash
# Ver extensiones instaladas
code --list-extensions --show-versions | grep joss

# Debería mostrar:
# jossecurity.joss-language@2.0.0
```

## Actualizar Extensión

```bash
# 1. Recompilar
npm run compile

# 2. Empaquetar nueva versión
npx vsce package --allow-star-activation

# 3. Reinstalar
code --install-extension joss-language-2.0.0.vsix --force

# 4. Reiniciar VS Code
```

## Desinstalar

```bash
code --uninstall-extension jossecurity.joss-language
```

## Troubleshooting

### "Extension not found"
- Verificar que el archivo .vsix existe
- Usar ruta absoluta si es necesario

### "Extension already installed"
- Agregar `--force` para sobrescribir

### Cambios no se reflejan
- Recompilar: `npm run compile`
- Empaquetar de nuevo
- Reinstalar con `--force`
- Reiniciar VS Code

---

**RECORDATORIO**: Nunca copiar carpetas manualmente. Siempre usar VSIX.
