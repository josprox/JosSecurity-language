# Análisis estático (`joss analyze`)

[Índice](README.md)

`joss analyze [archivo.joss]` analiza la entrada y los archivos `.joss` bajo `app/` sin ejecutar la aplicación. El pipeline carga cada archivo como una unidad fuente, registra declaraciones globales y después analiza cada callable con su propio scope.

```bash
joss analyze
joss analyze main.joss
```

El comando devuelve código de salida distinto de cero si hay errores. Los warnings se muestran, pero no bloquean. `joss run` aplica el mismo análisis antes de ejecutar y sólo continúa si no hay errores.

## Comprobaciones actuales

- Variables indefinidas, redeclaraciones y locales sin uso.
- Funciones, clases, superclases y métodos inexistentes cuando su receptor es conocido.
- Scope de parámetros, funciones, métodos, `Init` y closures.
- Inferencia fija en primera asignación y compatibilidad de reasignaciones.
- Inicializadores tipados, uniones/nullables, constantes, defaults, argumentos, retornos, rutas exhaustivas y aridad de funciones Joss.
- Tipos de clase inexistentes en variables, parámetros y retornos.
- Operadores e índices incompatibles.
- Código posterior a un `return` incondicional.
- Declaraciones duplicadas a nivel de proyecto.
- Símbolos de clases nativas y plugins JP v2 cargados por el proyecto.

## Evidencia e información desconocida

El analizador diferencia `unknown` de inválido y de `mixed`. Los retornos de todas las APIs core tienen metadata explícita, pero algunas entradas no coinciden con todos los resultados runtime (véase [auditoría](DOCUMENTATION_AUDIT.md)). `mixed` representa polimorfismo intencional. Una API nativa sin metadatos de parámetros no produce un error de aridad especulativo; un receptor dinámico no produce un error de miembro; `isset` y `empty` pueden consultar una variable ausente.

## Salida

Los diagnósticos usan el modelo de `pkg/diagnostics`:

```text
error[JOSS-TYPE-001] app/example.joss:3:2: Cannot use `string` as assignment for `$age` of type `int`.
  suggestion: Convert the value explicitly or use `let $name` only when dynamic typing is intentional.
```

Consulte [Diagnósticos](DIAGNOSTICOS.md) para códigos y severidades, y [Sistema de tipos](SISTEMA_TIPOS.md) para las reglas de inferencia.

## Arquitectura

`pkg/analyzer` no depende de `pkg/core`. El adaptador de `pkg/core/analyzer.go` construye su `Environment` desde los registros reales de built-ins, clases nativas y plugins. La CLI usa `analyzer.LoadProject`, por lo que conserva archivo, línea y columna en lugar de concatenar ASTs.

## Límites conocidos

- La prueba de retorno exhaustivo cubre bloques, ternarios, `match` con `default` y `try/catch`; no demuestra todavía terminación matemática de loops.
- Los parámetros de muchas APIs nativas siguen siendo variádicos/desconocidos para evitar errores de aridad; sus retornos sí son explícitos.
- No hay análisis de refinamiento sensible a ramas, taint/escape formal ni contratos de esquema de base de datos.
- Joss no tendrá grafo de imports fuente: el proyecto usa carga automática y un espacio único de declaraciones.
- La recuperación del parser todavía puede emitir más de un diagnóstico derivado tras un token inválido; cada hallazgo conserva ya línea y columna estructuradas desde el token original.

Estas limitaciones no se reportan como errores del usuario.
