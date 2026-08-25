# Módulos, archivos y plugins

## Estado actual

El proyecto de aplicación se carga como un conjunto de archivos `.joss`: `main.joss`, rutas y directorios convencionales como `app/`. `joss analyze` descubre esos archivos, conserva su identidad y resuelve sus declaraciones top-level a nivel de proyecto.

Las formas históricas `import`, `use` y `@import` están marcadas como obsoletas. No existe actualmente un grafo de módulos fuente con exports, visibilidad entre módulos o detección de ciclos. No debe asumirse esa semántica en código nuevo.

## Plugins

La extensibilidad modular vigente usa paquetes `.jp`:

- Se declaran en `joss.yaml` o se colocan en `plugins/`.
- El runtime los verifica y carga automáticamente.
- Cada paquete publica `META-INF/joss-symbols.json`.
- El analyzer consume ese índice para resolver clases, métodos y funciones exportadas sin ejecutar el plugin.

Un símbolo externo que no figure en el índice no se inventa ni se añade a una lista manual del analyzer: debe corregirse el paquete o su generación de símbolos.

## Archivos incluidos por infraestructura

Rutas, controladores, modelos y middleware se descubren según el layout de proyecto y la CLI/runtime. Esta inclusión física no crea un namespace por archivo; las funciones y clases top-level comparten el espacio de declaraciones del proyecto y sus duplicados son diagnosticados.

## Limitación conocida

La tesis describe un sistema de módulos con grafo y ciclos. Es una meta arquitectónica, no una capacidad implementada. Antes de incorporarla deben definirse semántica de exportación, resolución determinista, identidad canónica, ciclos permitidos y compatibilidad con plugins; no se debe reactivar `import` sólo como inclusión textual.
