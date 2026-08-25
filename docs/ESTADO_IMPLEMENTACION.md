# Estado de implementación

Este documento describe capacidades comprobables del código actual de **Joss v3.6.7.2**. No mezcla propuestas futuras con funciones terminadas.

## Implementado

- Intérprete Joss, inferencia fija, tipos explícitos y unión/nullable, constantes, retornos anotados y exhaustivos, recursión con frames léxicos aislados, `mixed` explícito mediante `let`, coerción textual sin pérdida, clases, herencia (`extends`), visibilidad, métodos estáticos, funciones `func`, closures, ternarios, `match`, ciclos, excepciones, `async`/`await` y canales.
- **Analizador semántico (`joss analyze`)**: unidades fuente, scopes por callable, símbolos de proyecto/plugins, inferencia, asignaciones, argumentos, aridad, operadores, índices, miembros conocidos, inalcanzable y diagnósticos estructurados.
- **Guardias de Sintaxis Educacionales**: Detección amigable de sintaxis foránea (`if`, `else`, `elif`, `switch`, `for`) sugiriendo el operador ternario `$cond ? $a : $b`, `match`, `while` y `foreach`.
- **Sandbox WASI y Permisos en Plugins (`PermissionGuard`)**: Control granular de permisos para I/O, red y variables de entorno (`http_get`, `file_read`, `env_read`, `db_query`) en la máquina virtual `JPBCVM`.
- **Fuentes canónicas**: built-ins en `pkg/core/builtins.go`, clases desde el registro nativo real, tipos en `pkg/typesystem` y catálogo del editor generado por `tools/cataloggen`.
- Motor de plantillas y vistas (`View::render`, `View::exists`, `View::share`, directiva `@json()`, `@foreach` anidado sobre arrays asociativos/expresiones complejas, comentarios `{{-- --}}`, layouts con `@extends` y `@yield`).
- Aplicaciones web y de consola, rutas HTTP y WebSocket dinámicas (`Router::any`, `Router::query`, `Router::match`), respuestas JSON/raw/stream, sesiones persistentes, CSRF, CORS, rate limit configurable y TLS integrado.
- Cliente HTTP nativo universal (`Http::get`, `Http::post`, `Http::put`, `Http::patch`, `Http::delete`, `Http::head`, `Http::options`, `Http::query`, `Http::json`, `Http::request`).
- Registro Pub con resolución dinámica de repositorios en tiempo real y fallback automático a GitHub Releases.
- SQLite, MySQL, PostgreSQL y MS SQL Server mediante GranDB ORM, migraciones y Schema Builder.
- Paquetes JP v2 con bytecode optimizado, carga automática desde `plugins/`, lockfile, índice de símbolos para IntelliSense (`META-INF/joss-symbols.json`), firma criptográfica Ed25519 y validación determinista (`joss plugin verify`).
- Runtime dual de plugins (`pkg/pluginruntime`): ejecución nativa de AST (`JOSSBC2Z`) y máquina virtual JPBC de 17 opcodes (`JPBC`) con Sandbox WASI.
- Compilador multilenguaje integrado (`joss plugin compile`): traduce Java, Python, PHP y Rust/Wasm a Bytecode Joss con tree shaking automático y cero dependencias para el usuario final.
- Distribuciones de Windows, Linux y macOS, SDK y extensión VSIX mediante el script de compilación y workflow manual.

## Contratos vigentes

- `func` es la única forma de declarar funciones y closures. `function` produce un error de migración.
- Los plugins declarados en `joss.yaml` o presentes en `plugins/` se cargan automáticamente. `import`, `@import`, `use` y namespaces fuente fueron eliminados del lexer, parser, AST y runtime.
- La ausencia de imports es permanente: no existe ni se proyecta sintaxis de módulos fuente, exports o namespaces. La modularidad se obtiene mediante layout, runtime integrado y plugins aislados.
- GranDB recibe inserts como mapas y Schema Builder recibe una función de blueprint; las variantes históricas con arrays/mapas paralelos fueron retiradas.
- Los plugins se ejecutan dentro del runtime de Joss con acceso seguro al entorno del proyecto (`r.Env`) y límites de instrucciones para evitar loops infinitos.

## No implementado todavía

- Ownership, inmutabilidad por defecto, taint/escape formal o análisis sensible a ramas.
- Backend LLVM/Cranelift o AOT a código máquina para el lenguaje principal. `pkg/bytecode` serializa el AST; JPBC pertenece al pipeline separado de plugins.
- Metadatos precisos de parámetros para todas las APIs nativas; sus retornos ya son explícitos y usan `mixed` sólo cuando el contrato es polimórfico.

Consulte [Arquitectura](ARQUITECTURA.md), [Sistema de tipos](SISTEMA_TIPOS.md) y [Auditoría técnica](AUDITORIA_TECNICA_2026.md).
