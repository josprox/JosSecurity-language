# Análisis estático (`joss analyze`)

`joss analyze [archivo.joss]` analiza la entrada y todos los `.joss` del proyecto sin ejecutar la aplicación. El pipeline carga cada archivo como una unidad fuente, registra declaraciones globales y después analiza cada callable con su propio scope.

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
- Inicializadores tipados, defaults, argumentos y aridad de funciones Joss.
- Operadores e índices incompatibles.
- Código posterior a un `return` incondicional.
- Declaraciones duplicadas a nivel de proyecto.
- Símbolos de clases nativas y plugins JP v2 cargados por el proyecto.

## Evidencia e información desconocida

El analizador diferencia `unknown` de inválido. Una API nativa sin firma formal no produce un error de aridad especulativo; un receptor dinámico no produce un error de miembro; `isset` y `empty` pueden consultar una variable ausente. Esta política evita convertir una limitación del checker en un supuesto error del programa.

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

- No hay sintaxis de tipo de retorno, nullable, union o constante.
- Las firmas nativas no describen todavía tipos de retorno; una cadena de miembros puede perder precisión.
- No hay análisis sensible a ramas, taint/escape formal, grafo de imports fuente ni contratos de esquema de base de datos.
- Los errores del parser siguen originándose como strings y el loader los adapta a diagnósticos; su rango es menos preciso que el de errores semánticos.

Estas limitaciones no se reportan como errores del usuario.
