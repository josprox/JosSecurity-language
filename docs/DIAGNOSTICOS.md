# Diagnósticos de Joss

El modelo común vive en `pkg/diagnostics` y contiene:

- `Code`
- `Severity` (`error`, `warning`, `info`)
- `Message`
- `File`
- `Range` con línea y columna
- `Explanation`
- `Suggestion`

Ejemplo:

```text
error[JOSS-TYPE-001] app/example.joss:3:2: Cannot use `string` as assignment for `$age` of type `int`.
  suggestion: Convert the value explicitly or use `let $name` only when dynamic typing is intentional.
```

## Familias actuales

| Código | Significado |
|---|---|
| `JOSS-PARSE-001` | Error sintáctico. |
| `JOSS-IO-001` | Fuente no legible. |
| `JOSS-SYM-001` | Variable no definida. |
| `JOSS-SYM-002` | Redeclaración en el mismo scope. |
| `JOSS-SYM-003` | Función inexistente. |
| `JOSS-SYM-004` | Clase inexistente. |
| `JOSS-SYM-005` | Superclase inexistente. |
| `JOSS-DECL-001..003` | Declaraciones duplicadas de función, clase o método. |
| `JOSS-TYPE-001` | Asignación incompatible. |
| `JOSS-TYPE-002` | Inicializador/default incompatible. |
| `JOSS-TYPE-003` | Argumento incompatible. |
| `JOSS-TYPE-004..007` | Operador, clave o índice incompatible. |
| `JOSS-CALL-001` | Aridad incorrecta. |
| `JOSS-MEMBER-001` | Método inexistente en clase conocida. |
| `JOSS-FLOW-001` | Código inalcanzable. |
| `JOSS-LINT-001` | Variable local sin uso. |

## Política contra falsos positivos

`unknown` significa información todavía no inferida; `mixed` significa decisión dinámica explícita. Ninguno es evidencia de invalidez. Las firmas nativas que sólo exponen nombres de método no generan diagnósticos de aridad. `isset` y `empty` son probes de existencia y no reportan una variable ausente. Los errores de miembros sólo se emiten cuando el receptor es una clase conocida y su tabla fue resuelta.

## Cómo añadir un diagnóstico

1. Reutilizar una familia existente o asignar un código estable nuevo.
2. Emitir `diagnostics.Diagnostic`, nunca concatenar strings en el analyzer.
3. Incluir archivo y token/rango.
4. Explicar por qué existe evidencia suficiente.
5. Añadir un test positivo y uno que proteja contra el falso positivo vecino.
6. Documentar el código aquí si forma parte de la interfaz pública.
