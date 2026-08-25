# Auditoría técnica del núcleo y alineación con la tesis (agosto de 2026)

## Línea base

Antes de los cambios, `go test ./...` y `go build ./...` pasaban; `go vet ./...` fallaba por construir manualmente una dirección SMTP incompatible con IPv6. `joss analyze` sobre JosSecurity producía 10 errores falsos y 14 warnings sin archivo: dos nombres built-in implementados pero ausentes del catálogo (`html_escape`, `unlink`) y siete usos de clases exportadas por plugins que el analyzer no consultaba (`BrevoClient`, `Notify`).

## Causas raíz encontradas

- Analyzer global basado en `map[string]int`, sin scopes, tipos ni unidad fuente.
- CLI concatenaba ASTs y eliminaba identidad de archivo.
- Catálogo de built-ins divergente del dispatcher: incluía funciones inexistentes y omitía funciones reales.
- Clases nativas y plugins se resolvían por caminos diferentes.
- `let $name` construía por error un símbolo llamado `$`.
- `VarTypes` sólo protegía declaraciones tipadas; primeras asignaciones y parámetros perdían su tipo.
- `CallMethod` y `CallMethodEvaluated` duplicaban binding y validación.
- El editor mantenía listas manuales separadas de keywords y clases/métodos nativos.
- CI sólo existía para distribución manual; no había workflow de push/PR.
- `go vet` reveló el uso de `fmt.Sprintf("%s:%s")` para SMTP en lugar de `net.JoinHostPort`.
- El pool conservaba `PluginRegistry` mientras borraba los símbolos expuestos, de modo que una reutilización podía omitir la recarga del plugin.

## Decisiones aplicadas

- Capas nuevas y consumidas: `pkg/typesystem`, `pkg/diagnostics`, `pkg/analyzer`.
- Scopes por callable y resolución de declaraciones a nivel de proyecto.
- Inferencia fija en primera asignación; `var` inferido; `let $x` dinámico explícito.
- Diagnósticos estructurados y deterministas por archivo/línea/columna.
- Entorno del analyzer adaptado desde registros reales de runtime y símbolos JP v2.
- Catálogo generado para VS Code, validado por CI.
- Las firmas ricas del editor se filtran contra ese catálogo; metadatos obsoletos ya no pueden publicar símbolos que el runtime no registra.
- Binding de métodos unificado en `CallMethodEvaluated`.
- Frames de invocación aislados, recursión directa/mutua/de método, límite de profundidad y contratos de retorno opcionales.
- Frames léxicos sin scope dinámico: las funciones con nombre no heredan locales del caller; las closures conservan su captura.
- Uniones `T|U`, nullable `T?`, retornos exhaustivos demostrables y firmas explícitas de retorno para todo el núcleo nativo.
- `const` y propiedades tipadas/constantes validadas por analyzer y runtime.
- Errores del parser almacenados directamente como diagnósticos estructurados; se eliminó la extracción de líneas desde strings.
- Eliminación física de `ImportStatement`, tablas/tokens de imports, linker textual de plugins y formato bytecode sin compresión.
- Retiro de APIs de compatibilidad sin consumidores canónicos: rutas crudas, inserts por arrays, Schema por mapas y `where(..., "json")`.
- Reset completo del registro de plugins al devolver un runtime al pool.
- Prueba de integración sobre todo JosSecurity.
- Fixture de proyecto versionado en `testdata/analyzer-project`; JosSecurity permanece como repositorio externo ignorado y su test se omite sólo cuando no está disponible.

## Resultado en JosSecurity

Al eliminar las causas de falsos positivos aparecieron seis problemas reales antes ocultos:

- `Math::length` no existe; se reemplazó por `count`.
- `Str::endsWith` no existe; se reemplazó por `str_ends_with`.
- `Str::upper` no existe; se reemplazó por `strtoupper`.
- Tres modelos heredaban de la clase eliminada `GranMySQL`; ahora heredan de `GranDB`.

Una segunda revisión detectó cinco warnings falsos: variables de modelos con el mismo nombre que su clase se confundían con acceso estático y no se marcaban como usadas al ser receptoras de `->`. Se corrigió la precedencia para que el símbolo léxico sombree a la clase y se añadió una regresión.

El resultado final es cero errores y cinco warnings, todos inspeccionados contra el código: `$id`, `$licenseData`, `$offset`, `$user` y `$domain` se inicializan pero no vuelven a leerse en sus respectivos callables. Son deuda real de JosSecurity, no bloquean la ejecución y no se modificaron sólo para obtener una salida vacía.

## Comparación con la tesis

La implementación coincide con la visión ALIM en el runtime integrado, parser Pratt, AST compartido, plugins aislados y toolchain único. La tesis, sin embargo, mezcla estado real, sintaxis conceptual y hoja de ruta:

| Afirmación de la tesis | Estado comprobado del repositorio |
|---|---|
| Pipeline incluye type checker | Existe un checker semántico inicial con inferencia fija, constantes, firmas/retornos y llamadas recursivas; aún no cubre taint, escape ni esquemas DB. |
| AOT/LLVM/Cranelift y código máquina | El build principal empaqueta AST serializado y el intérprete Go. LLVM/Cranelift no están implementados. |
| Lexer/parser en Rust | La implementación actual está en Go. |
| Inmutabilidad por defecto/ownership | No existe semántica de ownership ni inmutabilidad por defecto. |
| Imports/módulos con grafo y ciclos | La sintaxis histórica fue eliminada completamente y no volverá. Plugins y archivos convencionales se cargan automáticamente; Joss adopta deliberadamente un proyecto zero-imports. |
| Rutas/DB como nodos AST de primer orden | Hoy son llamadas a clases nativas (`Router`, `GranDB`), no nodos específicos. |
| 1,420 tests y 91.4% de cobertura | El repositorio contiene una suite Go mucho menor. Medición focalizada actual: parser 52.3%, typesystem 44.9%, analyzer 47.7% y core 14.8%; CI valida ejecución y no afirma una cobertura inexistente. |
| Taint analysis y 83% de vulnerabilidades | Existe un analizador de seguridad heurístico en el LSP, no un taint engine formal en el compilador. |

Estas diferencias no se “corrigieron” inventando características. La propuesta de módulos fuente del capítulo 11 queda expresamente descartada para Joss; la modularidad ALIM se conserva mediante componentes del runtime y plugins aislados. Las demás deben resolverse en la tesis distinguiendo implementación validada, sintaxis conceptual y trabajo futuro.

## Riesgos pendientes

- La recuperación del parser puede producir diagnósticos derivados después del primer token inválido, aunque ahora todos usan el modelo estructurado y conservan columna sin extraerla de mensajes.
- Los retornos nativos son explícitos, pero muchas APIs conservan parámetros variádicos hasta publicar contratos de aridad confiables.
- No existe refinamiento sensible a ramas, contratos de infraestructura, taint o escape formal. Los ciclos de módulos no aplican porque no existen módulos fuente.
- `pkg/core` sigue siendo amplio; dividir subsistemas de infraestructura requiere pruebas específicas y no se realizó sólo por estética.
- El catálogo de firmas ricas del LSP es metadato manual; la existencia de nombres sí proviene ya del catálogo generado.
