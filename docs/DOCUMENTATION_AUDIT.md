# Auditoría de implementación y documentación

[Índice](README.md) · [Estado actual](ESTADO_IMPLEMENTACION.md) · [Contribuir](CONTRIBUIR.md)

Auditoría realizada el 5 de septiembre de 2026. Su objetivo fue reconstruir la
documentación desde el código, separar aprendizaje/referencia/internos y dejar
controles contra la degradación. No cambia deliberadamente la semántica del lenguaje.

## Alcance y método

Se inventariaron los archivos rastreados y los árboles locales relevantes; lexer,
tokens, parser Pratt y AST; tipos, analyzer y diagnósticos; evaluator, frames,
planes, caches, bytecode y VM experimental; las 42 clases y 393 métodos registrados;
los 117 built-ins y sus dispatchers; servidor, router, vistas, auth, datos,
concurrencia, archivos y procesos; contenedores, compiladores y runtime de plugins;
CLI, formatter, linter, fixer, tests, plantillas, editor, workflows, ejemplos y
documentación existente. Los archivos generados/dependencias se trataron como
artefactos, verificando sus generadores o contratos en lugar de atribuirles
semántica fuente.

El repositorio tenía 429 archivos rastreados en la revisión inicial. `pkg/core`
concentra la superficie runtime; el inventario por paquetes y el catálogo generado
permitieron detectar nombres públicos ausentes o afirmados de más. Se ejecutó la
suite Go de base antes de reescribir y se añadieron pruebas que leen los snippets
publicados directamente.

Para la estructura se contrastaron índices oficiales de Rust (Book/Reference/API),
Go (tutoriales/referencia/paquetes), Python (tutorial/referencia/biblioteca), Dart,
Kotlin y TypeScript. Se adoptó la separación de audiencias, no su sintaxis ni su
modelo semántico.

Referencias estructurales consultadas: [Rust](https://doc.rust-lang.org/stable/),
[Go](https://go.dev/doc/), [Python](https://docs.python.org/3/),
[Dart](https://dart.dev/docs), [Kotlin](https://kotlinlang.org/docs/home.html) y
[TypeScript](https://www.typescriptlang.org/docs/).

## Brecha inicial corregida

| Clase de deuda | Evidencia encontrada | Tratamiento |
|---|---|---|
| Incorrecta | Se anunciaban `Http::query`, `System::change_db` y `SEO::twitter` aunque no están registrados. | Eliminados como API pública y documentadas alternativas. |
| Incorrecta | Plugins Wasm se describían como ejecución multilenguaje y sandbox WASI. | Se documentó que sólo valida magic y genera stubs; PermissionGuard cubre llamadas host mapeadas. |
| Incorrecta | `GranDB::transaction` se mostraba como transferencia atómica. | Se documentó que las consultas del callback no usan el `sql.Tx`. |
| Incorrecta | `array_pop`/`array_shift` se presentaban como mutaciones. | Se aclaró que sólo retornan el elemento. |
| Incorrecta | Longitudes Unicode se trataban como una unidad única. | Tabla de bytes, puntos Unicode y grafemas por API. |
| Obsoleta | Parámetros sin tipo en closures, aliases fuente y formas async antiguas. | Ejemplos tipados; sintaxis retirada explicada explícitamente. |
| Incompleta | No había recorrido para principiantes, glosario ni proyectos completos. | Se añadió aprendizaje progresivo de niveles 0–13. |
| Incompleta | Built-ins y clases carecían de retorno/error/contexto. | Referencia manual por familias y catálogo generado completo. |
| Mezclada | Objetivos de tesis/marketing aparecían como funciones terminadas. | Página de estado separa implementado, parcial y ausente. |
| Frágil | Fences sólo se parseaban y la lista nativa era manual. | Contratos `joss-run/check/error` y `tools/docgen --check`. |

Otros hallazgos corregidos: `json()` retorna WebResponse; JSON decodifica números
a float; TOON es texto simplificado; `await` bloquea y propaga el fallo; closures
separadas no comparten automáticamente reasignaciones; rutas WS aceptan parámetros;
el compilador de vistas procesa cuerpos foreach recursivamente; un build nativo
sigue interpretando AST comprimido.

## Arquitectura resultante

La raíz conserva nombres planos para no romper URLs públicas. El índice crea cuatro
capas: **Aprender**, **Referencia**, **Guías de aplicación** e **Internos**. El
recorrido principiante enlaza antes/después; las páginas de referencia remiten a
fuentes y catálogos. `docs/*.md` sigue siendo la fuente canónica.

Documentos nuevos: PRIMEROS_PASOS, FUNDAMENTOS, CONTROL_FLUJO, FUNCIONES,
COLECCIONES, CLASES, ERRORES, GRAMATICA, GLOSARIO, FUNCIONES_GLOBALES,
CATALOGO_NATIVO, PROYECTO_CONSOLA, PROYECTO_WEB, AUTENTICACION, CONTRIBUIR y este
informe. README raíz, índice, sintaxis, tipos, concurrencia, plugins, modelos,
Schema, migraciones, HTTP, WebSockets, analyzer, estado y guías relacionadas se
revisaron contra la implementación.

## Problemas de experiencia del lenguaje/tooling

Estos son defectos o contratos confusos observados, no propuestas ya aplicadas:

1. `break`/`continue` dentro de ciertos ternarios o match no es detectado por el
   plan del loop y puede escapar como panic.
2. Un brazo `match` cuyo valor es bloque puede devolver el nodo del bloque sin
   ejecutarlo, mientras el analyzer lo cuenta para flujo de retorno.
3. `GranDB::transaction` no dirige las consultas al Tx abierto.
4. Registro y handlers divergen (`Http::query`, `System::change_db`,
   `Sitemap::provider` con CapturedFunction, alias `firstofail`).
5. Retornos publicados difieren del runtime para boolval, json, microtime,
   strpos, base64_decode, array_merge y pluck.
6. Escrituras por índice y aliases pueden romper tipos de colecciones; algunos
   errores imprimen texto o hacen panic en vez de diagnóstico estructurado.
7. Truthiness difiere para float cero y map vacío; `??` recupera cualquier panic;
   `isset` sobre binding null cuenta existencia. Son difíciles de descubrir.
8. `joss analyze` carga entrada + app, pero el runtime precarga sólo dominios
   estándar y el servidor carga rutas por otra vía; `app/libs` y `routes.joss`
   revelan asimetrías.
9. El formatter de un solo archivo modifica por defecto si falta `--check`; el
   orden de `--filter` del test runner depende de Go flag.
10. El lexer omite silenciosamente bytes no ASCII fuera de strings; literales con
    cero inicial heredan base 0; faltan exponentes numéricos.
11. TOTP usa `math/rand` y la URL QR externa contiene el secreto.
12. `cmd/runner` existe localmente pero está ignorado por `.gitignore`; esto pone
    en riesgo reproducibilidad de build si no está presente en un clon.

## Riesgos y decisiones humanas

Requieren decisión de producto/compatibilidad: semántica de match y saltos en
ternarios; si alinear metadata nativa con uniones reales; si hacer transaction
verdaderamente transaccional; si unificar descubrimiento de archivos; si endurecer
TOTP/permisos de plugins; y si rastrear `cmd/runner`. Corregirlos puede cambiar
programas existentes, por eso esta auditoría sólo los hace visibles.

La documentación publicable se sincroniza con JosSecurity y debe mantener el menú,
headings y páginas en la misma lista. Los cambios locales previos del perfil de
JosSecurity no pertenecen a esta auditoría y se preservan.

## Verificación realizada

- `go run ./tools/cataloggen --check` y `go run ./tools/docgen --check`.
- `go vet ./...`, `go test ./...`, `go build ./...`.
- `go test -race ./pkg/parser ./pkg/typesystem ./pkg/analyzer ./pkg/core`.
- `npm ci` y `npm run compile` en `vscode-joss`; npm informó 1 vulnerabilidad
  moderada y 1 alta en el árbol instalado, sin aplicar una actualización automática.
- Binario temporal de `cmd/joss` y `joss analyze main.joss` sobre JosSecurity:
  análisis completado sin problemas y cinco plugins locales cargados.
- 76 fences Joss pasan la prueba sintáctica existente. De ellos, 62 contratos
  marcados se analizan; los `joss-run` se ejecutan con salida exacta y los
  `joss-error` exigen el código indicado.
- 181 enlaces Markdown locales fueron recorridos por el test (los externos no se
  descargan); 45 documentos canónicos coinciden byte por byte con 45 públicos,
  y menú/mapa de títulos contienen exactamente esas 45 páginas.
- `git diff --check` en el repositorio principal y el proyecto publicable no mostró
  errores de whitespace después de normalizar finales de archivo.
