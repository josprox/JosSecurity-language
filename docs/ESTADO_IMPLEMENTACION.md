# Estado de implementación

Este documento describe capacidades comprobables del código actual de **Joss v3.6.7**. No mezcla propuestas futuras con funciones terminadas.

## Implementado

- Intérprete Joss, inferencia fija en primera asignación, tipos explícitos, `mixed` explícito mediante `let`, coerción textual sin pérdida, clases, herencia (`extends`), visibilidad, métodos estáticos, funciones `func`, closures, ternarios, `match`, ciclos, excepciones, `async`/`await` y canales.
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

## Compatibilidad, no limitaciones

- `func` es la forma canónica. `function` sigue aceptándose para no romper código anterior.
- Los plugins declarados en `joss.yaml` o presentes en `plugins/` se cargan automáticamente sin necesidad de `use` ni `import`.
- Los plugins se ejecutan dentro del runtime de Joss con acceso seguro al entorno del proyecto (`r.Env`) y límites de instrucciones para evitar loops infinitos.

## No implementado todavía

- Declaraciones `const`, tipos nullable/union y anotaciones de retorno.
- Ownership, inmutabilidad por defecto, taint/escape formal o análisis sensible a ramas.
- Grafo de imports de fuentes y detección de ciclos; `import`, `use` y `@import` están obsoletos.
- Backend LLVM/Cranelift o AOT a código máquina para el lenguaje principal. `pkg/bytecode` serializa el AST; JPBC pertenece al pipeline separado de plugins.
- Firmas nativas completas con tipos de retorno, necesarias para inferir con precisión todas las cadenas de miembros.

Consulte [Arquitectura](ARQUITECTURA.md), [Sistema de tipos](SISTEMA_TIPOS.md) y [Auditoría técnica](AUDITORIA_TECNICA_2026.md).
