# Plan de Actualización - JosSecurity v3.0.0

## 1. Documentación
- [ ] Actualizar `docs/SINTAXIS.md` para incluir el Operador Pipe (`|>`).
- [ ] Actualizar `docs/README.md` mencionando la versión 3.0.1.

## 2. VS Code Extension (vscode-joss)
- [ ] Actualizar `package.json`:
    - Cambiar versión a `3.0.0`.
    - Actualizar descripción si es necesario.
- [ ] Actualizar Gramática (`syntaxes/joss.tmLanguage.json`):
    - Agregar regla para el operador `|>`.
    - Asegurar que se resalte como `keyword.operator`.
- [ ] Actualizar `README.md` de la extensión:
    - Mencionar soporte para Pipe Operator.
    - Actualizar versión en la sección de versión.

## 3. Limpieza
- [ ] Verificar que no queden archivos temporales (`examples/pipe_test.joss` ya fue borrado).
