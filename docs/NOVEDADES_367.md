# Novedades de Joss v3.6.7.2

[Índice](README.md)

Esta versión consolida el motor alrededor de garantías comprobables: análisis
antes de ejecutar, tipos estables después de inferirlos, carga automática sin
imports fuente y tooling que valida los proyectos generados.

## Lenguaje y sistema de tipos

- `$x = valor` y `var $x = valor` infieren un tipo fijo. Una reasignación debe
  conservarlo.
- `let $x = valor` declara dinamismo explícito mediante `mixed`.
- Se implementaron constantes, tipos nullable/unión y tipos de retorno.
- Las funciones y métodos usan frames léxicos aislados; la recursión directa y
  mutua funciona con un límite configurable de 1024 llamadas por defecto.
- `func` es la única palabra clave de función. `function`, `import`, `@import`,
  `use` y namespaces fuente fueron eliminados.

## Análisis estático

- `joss analyze` carga el mismo proyecto y superficie nativa que el runtime.
- Resuelve clases y funciones top-level en dos pasadas, incluidas referencias
  adelantadas necesarias para recursión mutua.
- Comprueba símbolos, scopes, aridad conocida, operadores, asignaciones,
  argumentos, retornos y flujo alcanzable.
- Los diagnósticos estructurados incluyen código, severidad, archivo, rango,
  explicación y sugerencia; la información desconocida no se convierte en un
  error sin evidencia.

## Runtime, bytecode y plugins

- Los frames de ejecución evitan el scope dinámico accidental y el runtime
  protege el límite de recursión.
- El formato principal `JOSSBC2Z` contiene AST serializado y comprimido. El
  runner continúa interpretándolo; no se presenta como código máquina.
- Los paquetes JP v2 contienen manifiesto, tabla de símbolos, bytecode y firma
  Ed25519 verificable.
- Los backends Python, Java, PHP y WASM dependen de sus hosts/protocolos reales
  cuando corresponde; no se promete eliminar runtimes externos.
- Los plugins declarados en `joss.yaml` o instalados en `plugins/` se cargan
  automáticamente, sin sentencias import en el código Joss.

## CLI y proyectos generados

- `joss new web`, `console`, `package` y `plugin` producen proyectos que pasan
  parser, análisis y sus flujos de ejecución/compilación representativos.
- `make:migration` normaliza nombres como `create_products_table`, aplica la
  migración y sólo informa éxito después de registrar el batch.
- `make:crud` valida el esquema, distingue relaciones existentes, limita los
  campos escribibles, evita borrados por GET y no duplica rutas ni navegación.

## Validación

El repositorio valida build, tests, análisis, catálogo del lenguaje y ejemplos
representativos mediante CI. Para el estado exacto y sus límites consulte
[Estado de implementación](ESTADO_IMPLEMENTACION.md) y la
[Auditoría técnica](AUDITORIA_TECNICA_2026.md).
