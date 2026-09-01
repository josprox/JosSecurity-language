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
| `JOSS-SYM-006` | Reasignación de una constante. |
| `JOSS-DECL-001..003` | Declaraciones duplicadas de función, clase o método. |
| `JOSS-TYPE-001` | Asignación incompatible. |
| `JOSS-TYPE-002` | Inicializador/default incompatible. |
| `JOSS-TYPE-003` | Argumento incompatible. |
| `JOSS-TYPE-004..007` | Operador, clave o índice incompatible. |
| `JOSS-TYPE-008` | Retorno incompatible con la anotación. |
| `JOSS-TYPE-009` | Tipo fuente o clase de tipo inexistente. |
| `JOSS-TYPE-010` | Una función anotada puede terminar sin retornar ni lanzar. |
| `JOSS-TYPE-011` | Parámetro sin tipo explícito; escribir el tipo o `mixed`. |
| `JOSS-REF-001..006` | Contrato `ref` inválido: marcador, l-value, constante, tipo, escape o default. |
| `JOSS-ACCESS-001` | Declaración privada usada desde otro archivo. |
| `JOSS-ACCESS-002` | Miembro privado/protegido inaccesible. |
| `JOSS-CALL-001` | Aridad incorrecta. |
| `JOSS-MEMBER-001` | Método inexistente en clase conocida. |
| `JOSS-FLOW-001` | Código inalcanzable. |
| `JOSS-LINT-001` | Variable local sin uso. |
| `JOSS-ARITH-001` | Una operación entera desborda el rango signed de 64 bits. |
| `JOSS-ARITH-002` | División o módulo entre cero. |
| `JOSS-INDEX-001` | Índice negativo o fuera de rango. |
| `JOSS-INDEX-002` | Tipo de índice incompatible con la colección/string. |

## Política contra falsos positivos

`unknown` significa información todavía no inferida; `mixed` significa decisión dinámica explícita. Ninguno es evidencia de invalidez. Los retornos core siempre tienen firma explícita; las APIs sin metadatos de parámetros no generan diagnósticos de aridad. `isset` y `empty` son probes de existencia y no reportan una variable ausente. Los errores de miembros sólo se emiten cuando el receptor es una clase conocida y su tabla fue resuelta.

## Cómo añadir un diagnóstico

1. Reutilizar una familia existente o asignar un código estable nuevo.
2. Emitir `diagnostics.Diagnostic`, nunca concatenar strings en el analyzer.
3. Incluir archivo y token/rango.
4. Explicar por qué existe evidencia suficiente.
5. Añadir un test positivo y uno que proteja contra el falso positivo vecino.
6. Documentar el código aquí si forma parte de la interfaz pública.
