# Módulos, archivos y plugins

[Índice](README.md)

## Estado actual

El proyecto se organiza como un conjunto de archivos `.joss`, pero cada comando
elige una superficie concreta. `joss analyze` carga la entrada indicada y todo
`app/`; no agrega automáticamente `routes.joss` o `api.joss`. El servidor carga
sus rutas por infraestructura propia. El runtime de `joss run` precarga sólo los
dominios estándar bajo `app/` y hoy omite `app/libs`, aunque el analizador sí la
ve. Esta asimetría está registrada como deuda.

Las formas históricas `import`, `use`, `@import` y `Namespace` fueron eliminadas del conjunto de tokens, AST, ejecutor y compilador de plugins. El parser las rechaza con un mensaje de migración. Esta ausencia es una decisión permanente del lenguaje: no habrá exports, namespaces fuente ni grafo de imports.

## Plugins

La extensibilidad modular vigente usa paquetes `.jp`:

- Se declaran en `joss.yaml` o se colocan en `plugins/`.
- El runtime los verifica y carga automáticamente.
- Cada paquete publica `META-INF/joss-symbols.json`.
- El analyzer consume ese índice para resolver clases, métodos y funciones exportadas sin ejecutar el plugin.

Un símbolo externo que no figure en el índice no se inventa ni se añade a una lista manual del analyzer: debe corregirse el paquete o su generación de símbolos.

## Archivos incluidos por infraestructura

Rutas, controladores, modelos y middleware se descubren según el layout de proyecto y la CLI/runtime. Esta inclusión física no crea un namespace por archivo; las funciones y clases top-level comparten el espacio de declaraciones del proyecto y sus duplicados son diagnosticados.

## Decisión frente a la tesis

El capítulo 11 de la tesis describe módulos fuente con interfaces, exports y un DAG. La implementación adopta una decisión distinta: “modular” en ALIM significa capacidades integradas desacopladas y plugins aislados, no imports escritos por la aplicación. Esta discrepancia es deliberada y el diseño de módulos fuente de la tesis no forma parte de la hoja de ruta de Joss.

Las funciones y clases top-level forman un único espacio de declaraciones del proyecto. Las variables fuente top-level no se heredan dinámicamente dentro de funciones con nombre; deben pasar como parámetros. Una closure sí conserva el entorno léxico que captura.
