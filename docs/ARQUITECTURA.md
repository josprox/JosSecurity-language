# Arquitectura del lenguaje Joss

## Pipeline real

```text
fuentes .joss
  → lexer (`pkg/parser/lexer.go`)
  → parser Pratt (`pkg/parser`)
  → AST (`pkg/parser/ast*.go`)
  → análisis semántico (`pkg/analyzer`)
  → diagnósticos (`pkg/diagnostics`)
  → intérprete AST (`pkg/core/evaluator*.go`, `executor.go`)
  → runtime integrado (`pkg/core`, `pkg/server`)
```

`joss analyze` conserva cada archivo como una `analyzer.SourceUnit`; no concatena ASTs perdiendo el origen. Primero registra declaraciones globales de funciones y clases, después analiza cada método con un scope léxico independiente. Las clases nativas provienen de `Runtime.RegisterNativeClasses`; los plugins aportan sus índices de símbolos JP v2.

## Responsabilidades

| Paquete | Responsabilidad |
|---|---|
| `pkg/parser` | Tokens, lexer, precedencias, parser y nodos AST. |
| `pkg/typesystem` | Nombres canónicos, aliases, inferencia, coerción explícita y compatibilidad de asignación. |
| `pkg/analyzer` | Unidades fuente, scopes, símbolos, inferencia de expresiones, firmas y flujo alcanzable. No depende del runtime. |
| `pkg/diagnostics` | Modelo común: código, severidad, mensaje, archivo, rango, explicación y sugerencia. |
| `pkg/core` | Adaptación de catálogos reales al analizador, intérprete y primitivas integradas. |
| `pkg/pluginruntime`, `pkg/pluginpkg` | Carga aislada, verificación y resolución de símbolos de plugins JP v2. |
| `pkg/bytecode` | Serialización comprimida del AST. No es código máquina ni LLVM IR. |
| `cmd/joss` | CLI, análisis de proyecto, ejecución, build y administración. |
| `vscode-joss` | LSP/editor. Consume el catálogo generado del núcleo. |

## Fuentes de verdad

- Keywords: `pkg/parser/token.go`; `parser.KeywordNames()` es la proyección para tooling.
- Tipos y compatibilidad: `pkg/typesystem`.
- Built-ins globales: `pkg/core/builtins.go`. El dispatcher rechaza nombres fuera de ese catálogo.
- Clases/métodos nativos: llamadas a `registerNative` dentro de `Runtime.RegisterNativeClasses()`.
- Símbolos de plugins: `pluginpkg.SymbolIndex` incluido en cada `.jp`.
- Diagnósticos: `pkg/diagnostics.Diagnostic` y códigos emitidos por `pkg/analyzer`.
- Catálogo de VS Code: `vscode-joss/src/server/generated/languageCatalog.json`, generado mediante `go run ./tools/cataloggen`.

CI ejecuta `go run ./tools/cataloggen --check`; editar a mano el catálogo generado no es válido.

## Scopes y símbolos

- Cada función, método, `Init` y closure tiene un scope propio.
- Los parámetros pertenecen únicamente a su callable.
- Los bloques de control usan el scope del callable para reflejar el runtime actual.
- El binding de `foreach` puede reutilizarse en otro loop; el runtime lo trata como asignación.
- Las clases y funciones top-level se resuelven a nivel de proyecto.
- Los globals nativos y símbolos de plugins se inyectan mediante `analyzer.Environment`.

## Build y ejecución

El modo de desarrollo interpreta el AST. `pkg/bytecode` codifica el AST con `gob` y compresión. Los builds nativos empaquetan ese bytecode junto con el runner Go; actualmente no existe un backend LLVM/Cranelift ni traducción AOT del programa Joss a código máquina. El compilador de plugins sí posee un IR JPBC separado; no debe confundirse con el pipeline del lenguaje principal.

## Regla de dependencia

Las capas de lenguaje (`parser`, `typesystem`, `diagnostics`, `analyzer`) no importan `core`. `core` adapta sus registros al analizador. Esta dirección evita que el type checker dependa de efectos secundarios del servidor o de la base de datos.
