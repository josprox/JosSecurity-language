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
| Pipeline incluye type checker | Ahora existe un checker semántico inicial; aún no cubre retornos anotados, taint, escape ni esquemas DB. |
| AOT/LLVM/Cranelift y código máquina | El build principal empaqueta AST serializado y el intérprete Go. LLVM/Cranelift no están implementados. |
| Lexer/parser en Rust | La implementación actual está en Go. |
| Inmutabilidad por defecto/ownership | No existe semántica de ownership ni inmutabilidad por defecto. |
| Imports/módulos con grafo y ciclos | `import`, `use` y `@import` están obsoletos; plugins de `joss.yaml` se cargan automáticamente. No hay grafo de módulos fuente. |
| Rutas/DB como nodos AST de primer orden | Hoy son llamadas a clases nativas (`Router`, `GranDB`), no nodos específicos. |
| 1,420 tests y 91.4% de cobertura | El repositorio contiene una suite Go mucho menor. Medición focalizada actual: parser 52.3%, typesystem 44.9%, analyzer 47.7% y core 14.8%; CI valida ejecución y no afirma una cobertura inexistente. |
| Taint analysis y 83% de vulnerabilidades | Existe un analizador de seguridad heurístico en el LSP, no un taint engine formal en el compilador. |

Estas diferencias no se “corrigieron” inventando características. Deben resolverse en la tesis distinguiendo con claridad implementación validada, sintaxis conceptual y trabajo futuro.

## Riesgos pendientes

- La gramática no declara tipos de retorno, nullables, unions ni constantes.
- El parser todavía expone errores como strings; el loader los adapta al modelo estructurado y sólo garantiza columna para diagnósticos semánticos.
- El análisis de miembros encadenados pierde tipo cuando las APIs nativas no publican retorno formal.
- No existe análisis de flujo sensible a ramas, contratos de infraestructura, taint, escape o ciclos de módulos.
- `pkg/core` sigue siendo amplio; dividir subsistemas de infraestructura requiere pruebas específicas y no se realizó sólo por estética.
- El catálogo de firmas ricas del LSP es metadato manual; la existencia de nombres sí proviene ya del catálogo generado.
