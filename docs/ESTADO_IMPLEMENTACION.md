# Estado y límites de implementación

[Índice](README.md) · [Arquitectura](ARQUITECTURA.md) · [Auditoría](DOCUMENTATION_AUDIT.md)

Esta página separa capacidades disponibles, implementaciones parciales y
objetivos de diseño. Corresponde al código auditado, no a garantías de una
versión descargada previamente.

| Área | Estado real | Referencia |
|---|---|---|
| Lenguaje | Parser Pratt, variables de tipo fijo o mixed explícito, clases/herencia/visibilidad, funciones/closures/ref, ternarios, loops, match de valores, try/catch. | [Sintaxis](SINTAXIS.md) |
| Tipos | int64, float64, decimal, strings, arrays, maps, object, channel, clases, uniones y nullable. Análisis y defensa runtime con diferencias registradas. | [Tipos](SISTEMA_TIPOS.md) |
| Concurrencia | Goroutines mediante async, Future, espera bloqueante y channels. Sin cancelación estructurada ni aislamiento profundo. | [Concurrencia](CONCURRENCIA.md) |
| Analizador | Símbolos en dos pasadas, scopes, asignabilidad, miembros conocidos, retornos y diagnósticos. Sin prueba general de terminación ni refinamiento por ramas. | [Analizador](ANALIZADOR.md) |
| Ejecución principal | AST interpretado con planes de callable, frames y caches. Build nativo empaqueta AST comprimido JOSSBC2Z con runner Go. | [Arquitectura](ARQUITECTURA.md) |
| VM experimental | pkg/vm contiene compilador y VM independientes; no es backend por defecto de CLI/core. | [Internos](ARQUITECTURA.md) |
| Web | Router HTTP/WS, vistas, Request/Response, sesión, CSRF, CORS, TLS y límites configurables. | [Proyecto web](PROYECTO_WEB.md) |
| SQL | Adaptadores SQLite/MySQL/PostgreSQL/SQL Server, builder, Schema y migraciones. Portabilidad parcial por operación; transaction no vincula consultas al Tx. | [Modelos](MODELOS.md) |
| Plugins | Contenedor firmado e índice, AST y JPBC; compiladores parciales. Ruta Wasm genera stubs de texto, no ejecuta Wasm. | [Plugins](PLUGINS.md) |
| Permisos de plugins | Guard para llamadas host mapeadas; no sandbox WASI/OS ni consentimiento por paquete. | [Plugins](PLUGINS.md) |
| Herramientas | CLI, formatter, linter/fix, runner de tests y extensión VS Code con catálogo generado. No hay comando REPL ni debugger integrado. | [CLI](CLI.md) |

No existen ownership, inmutabilidad por defecto, punteros generales, interfaces,
traits, protocolos, generics de funciones/clases, defer/finally, select de canales,
ni backend LLVM/Cranelift. Las anotaciones de colecciones no equivalen a generics
universales ni garantizan que cada mutación revalide elementos.

La ausencia de imports/exports/namespaces fuente es una decisión permanente,
no una función pendiente. La modularidad de ALIM utiliza capacidades integradas,
organización física y plugins con carga automática.

La tesis combina arquitectura objetivo e implementación. Sus afirmaciones de
backend, aislamiento y módulos deben contrastarse con este estado y con la
[auditoría técnica histórica](AUDITORIA_TECNICA_2026.md). No presentes objetivos
como prestaciones ya terminadas.
