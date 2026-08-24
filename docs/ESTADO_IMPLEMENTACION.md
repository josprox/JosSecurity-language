# Estado de implementación

Este documento describe capacidades comprobables del código actual de **Joss v3.6.7**. No mezcla propuestas futuras con funciones terminadas.

## Implementado

- Intérprete Joss, tipos opcionales con coerción automática, clases, herencia (`extends`), modificadores de visibilidad (`public`, `private`, `protected`), métodos estáticos (`static`), funciones `func`, closures, ternarios, `match`, ciclos, excepciones (`try`/`catch`/`throw`), `panic()`, `async`/`await` y canales.
- **Analizador Estático AST (`joss analyze`)**: Inspección estática automática en `joss run` y `joss server start`, y ejecutable independiente con `joss analyze`.
- **Guardias de Sintaxis Educacionales**: Detección amigable de sintaxis foránea (`if`, `else`, `elif`, `switch`, `for`) sugiriendo el operador ternario `$cond ? $a : $b`, `match`, `while` y `foreach`.
- **Sandbox WASI y Permisos en Plugins (`PermissionGuard`)**: Control granular de permisos para I/O, red y variables de entorno (`http_get`, `file_read`, `env_read`, `db_query`) en la máquina virtual `JPBCVM`.
- **Registro Centralizado de Builtins (`pkg/core/builtins.go`)**: Registro único y canónico para funciones integradas del lenguaje.
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
