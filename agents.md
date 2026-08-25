# Guía de arquitectura para agentes

Este archivo es una guía operativa del repositorio, no un changelog. Antes de modificar semántica, lea `docs/ARQUITECTURA.md`, `docs/SISTEMA_TIPOS.md`, `docs/DIAGNOSTICOS.md` y los tests del subsistema afectado.

## Pipeline real

```text
.joss → lexer → parser Pratt → AST → semantic analyzer → diagnostics
                                      ↓ sin errores
                                intérprete/runtime Go
```

- `pkg/parser`: tokens, lexer, parser y AST.
- `pkg/typesystem`: tipos canónicos, asignabilidad, inferencia y coerción.
- `pkg/analyzer`: scopes, símbolos, firmas, inferencia de expresiones y flujo.
- `pkg/diagnostics`: formato estructurado de errores y warnings.
- `pkg/core`: intérprete, runtime y clases/built-ins integrados; adapta sus registros al analyzer.
- `pkg/bytecode`: AST serializado y comprimido, no IR de código máquina.
- `pkg/pluginruntime`, `pkg/pluginpkg`, `pkg/plugincompiler`: runtime, paquetes y JPBC de plugins.
- `cmd/joss`: CLI y orquestación del proyecto.
- `vscode-joss`: LSP; consume el catálogo generado del lenguaje.

La dirección de dependencias es `parser/typesystem/diagnostics → analyzer → core adapter`. `pkg/analyzer` no debe importar `pkg/core`.

## Fuentes de verdad que no deben duplicarse

- Keywords: tabla de tokens de `pkg/parser/token.go`, proyectada por `parser.KeywordNames()`.
- Tipos y compatibilidad: `pkg/typesystem`; no reintroducir aliases fuera de la lista canónica.
- Built-ins globales: nombres en `pkg/core/builtins.go` y retornos en `pkg/core/native_signatures.go`; toda entrada debe tener un caso real en el dispatcher.
- Clases y métodos nativos: registros ejecutados por `Runtime.RegisterNativeClasses()` y retornos en `pkg/core/native_signatures.go`; usar `GetNativeClassMethods()` para inspección.
- Plugins: índice `pluginpkg.SymbolIndex` del paquete `.jp`.
- Diagnósticos: `diagnostics.Diagnostic` y códigos estables del analyzer.
- VS Code: `vscode-joss/src/server/generated/languageCatalog.json`, generado por `go run ./tools/cataloggen`; nunca editarlo manualmente.

## Semántica de variables y tipos

- `$x = 1`: la primera asignación declara e infiere `int`; las siguientes deben ser compatibles.
- `var $x = 1`: inferencia explícita, también fija.
- `int $x = 1` o `let int $x = 1`: tipo explícito.
- `let $x = 1`: `mixed` explícito; permite cambiar de tipo.
- Una inicialización `nil` pospone la inferencia hasta un valor concreto.
- `T|null` declara una unión nullable; `T?` es sólo su atajo sintáctico y el AST lo normaliza a `T|null`.
- La coerción de string a número/bool debe ser completa y sin pérdida; use `typesystem.CoerceString` tanto en análisis como en runtime.
- Los parámetros tipados conservan su tipo durante el cuerpo.
- `const $x = ...` infiere un tipo fijo; `const int $x = ...` lo declara. También se protegen propiedades constantes.
- `func name(...): Type` declara el retorno; analyzer y runtime validan cada retorno explícito y el analyzer exige retorno/throw en todas las rutas demostrables.
- Cada invocación de función/método usa un frame aislado. Los callables con nombre sólo ven parámetros, locales, `this` y bindings del host/plugin; no heredan locales del caller. Las closures sí conservan su entorno capturado. La recursión está limitada por `Runtime.MaxCallDepth` (1024 por defecto).

No implemente reglas paralelas en parser, CLI o evaluator. Añada primero la regla y tests a `pkg/typesystem`, después consúmala desde analyzer/runtime.

## Scopes, símbolos y falsos positivos

- Funciones, métodos, `Init` y closures tienen scope de callable independiente.
- Las declaraciones de funciones y clases top-level se resuelven en dos pasadas a nivel de proyecto. Las variables fuente top-level no son globals implícitos de funciones; use parámetros o closures.
- `foreach` puede reutilizar un binding existente; el runtime lo trata como asignación.
- `isset` y `empty` consultan existencia y no deben acusar una variable ausente.
- `unknown` representa falta de información; `mixed` representa dinamismo explícito. Ninguno prueba invalidez.
- No valide aridad de una API nativa si sus parámetros no están publicados. Todo retorno nativo debe ser explícito; use `mixed`, no `unknown`, cuando el contrato sea polimórfico.
- Sólo emita error de miembro cuando el tipo receptor y la tabla de miembros estén resueltos.

## Cómo añadir una característica

1. Defina la semántica y sus invariantes, contrastándolas con la tesis y el comportamiento actual.
2. Si cambia sintaxis, añada token/lexer/parser/AST y tests positivos y negativos.
3. Añada tipos/compatibilidad a `pkg/typesystem` cuando corresponda.
4. Añada resolución y diagnóstico a `pkg/analyzer`, conservando unidad fuente y scope.
5. Implemente el mismo contrato en `pkg/core`; el runtime es defensa, no el primer detector.
6. Actualice el catálogo generado si cambia un símbolo compartido.
7. Añada regresión e integración, documentación y valide JosSecurity.

## Cómo añadir un tipo

1. Incorpore el `Kind` y nombre fuente canónico en `pkg/typesystem`; no añada aliases sólo por compatibilidad.
2. Defina `Assignable`, inferencia, conversión y tests de bordes.
3. Enseñe al analyzer a inferir sus literales/operadores.
4. Enseñe al runtime a reconocer y validar el mismo tipo usando el paquete común.
5. Actualice parser sólo si requiere sintaxis nueva.
6. Regenerar: `go run ./tools/cataloggen`.

## Cómo añadir un diagnóstico

1. Use una familia/código estable `JOSS-...`.
2. Emita `diagnostics.Diagnostic`, no strings ad hoc.
3. Incluya severidad, archivo, rango, explicación y sugerencia útil.
4. Exija evidencia suficiente; una limitación del checker no es error del usuario.
5. Añada un caso inválido y el caso válido vecino que no debe diagnosticarse.
6. Documente el código en `docs/DIAGNOSTICOS.md`.

## Convenciones del lenguaje que suelen causar errores

- No hay `if/else`; usar ternarios con bloques. `return` burbujea y permite guard clauses.
- Sólo existe `func`; `function`, `import`, `@import`, `use` y namespaces fuente son sintaxis eliminada. Zero-imports es una decisión permanente: no añadir módulos fuente, exports ni un DAG de imports.
- `foreach`, `while` y `match` son las estructuras soportadas; `await($future)` evita ambigüedad.
- Clases estáticas usan `::`; instancias usan `->`; concatenación usa `.`.
- `Auth::user()` retorna `*Instance`: acceder con `$u->id`; preferir `Auth::id()`.
- No pasar `Auth::user()` directamente a vistas; extraer campos escalares.
- `GranDB::get()` retorna lista nativa, no JSON string.
- GranDB `insert`/`insertGetId` reciben un único mapa; no reintroducir arrays paralelos. Schema `create`/`table` reciben una closure de blueprint.
- Uploads están en `$file["content"]`. Binarios deben usar `Response::raw` con `Content-Disposition`.
- El motor de vistas procesa herencia, includes, ternarios de bloque y luego `@foreach`; no usar un ternario de bloque dependiente del item dentro del loop. Precomputar en controller.
- Las rutas WebSocket son estáticas; autenticar el JWT manualmente antes de usar la sesión.
- Los módulos nativos deben preferir `r.Env` sobre `os.Getenv`.
- Servicios locales deben usar `127.0.0.1` para evitar resolución IPv6 inesperada.

## Plugins y lifecycle

Los plugins declarados en `joss.yaml` o presentes en `plugins/` se cargan automáticamente; no requieren imports fuente. `Runtime.Free()` debe borrar `PluginRegistry` junto con símbolos/clases: conservar sólo uno de esos estados rompe la reutilización del pool. Los forks comparten recursos inmutables/seguros y copian variables, tipos, constantes e instancias; cualquier nuevo estado mutable debe tener una decisión explícita de copia o compartición y un test de concurrencia.

## Comandos de validación

Desde la raíz:

```bash
gofmt -w <archivos-go-modificados>
go run ./tools/cataloggen --check
go vet ./...
go test ./...
go test -race ./pkg/parser ./pkg/typesystem ./pkg/analyzer ./pkg/core
go build ./...
```

Extensión:

```bash
cd vscode-joss
npm ci
npm run compile
```

Integración real:

```bash
go build -o <temporal>/joss ./cmd/joss
cd ejemplos/Joss-Red-JosSecurity
<temporal>/joss analyze main.joss
```

`joss analyze` debe finalizar con código 0 cuando sólo haya warnings. No edite JosSecurity para silenciar un diagnóstico sin comprobar el símbolo contra el runtime real. CI en `.github/workflows/ci.yml` ejecuta estas familias en push y pull request.

Los cambios en `joss new` deben pasar `pkg/template.TestGeneratedProjectsUseCanonicalParsableJoss` para web/consola y `cmd/joss.TestNewPackageAndPluginTemplatesCompileEndToEnd` para package/plugin. Estas regresiones verifican parser, analyzer, ejecución de consola y archivos JP firmados/decodificables; no añada un template sin extender esa matriz.

Los cambios en migraciones/CRUD deben pasar `cmd/joss.TestMigrationAndCRUDGeneratorsWorkTogether`. `make:migration` normaliza `create_x`, `create_x_table` y `x` hacia una tabla lógica plural sin prefijo; `Schema` agrega el prefijo. El runtime crea sus tablas internas mediante `ensureInternalSchemaTable`, no mediante mapas enviados a la API pública `Schema::create/table` (esa compatibilidad fue eliminada). `LogMigration` retorna el error y el runner no puede anunciar éxito si el registro del batch falla. `make:crud` es sólo web, valida identificadores, genera mapas de campos permitidos, usa `POST` para borrar y debe mantener idempotentes rutas/navbar.

`docs/*.md` es la fuente canónica de documentación. La copia publicable de JosSecurity vive en `ejemplos/Joss-Red-JosSecurity/assets/docs/` y debe coincidir archivo por archivo; su menú y el mapa de `DocsController.pageHeading()` deben cubrir exactamente el mismo conjunto. No reintroducir descargas de documentación durante el arranque: la publicación debe ser determinista y versionada.

## Límites que deben declararse con honestidad

El build principal empaqueta AST serializado comprimido (`JOSSBC2Z`) con el runner Go y continúa interpretándolo; el formato anterior sin compresión ya no se acepta y no hay backend LLVM/Cranelift. Tampoco existen todavía ownership, inmutabilidad por defecto ni taint/escape formal. No habrá grafo de módulos fuente: la modularidad de ALIM se implementa mediante capacidades integradas, organización física y plugins aislados, todos con carga automática. La tesis combina arquitectura objetivo y estado implementado; documente explícitamente esta discrepancia.
