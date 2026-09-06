# Referencia de diagnósticos

[Índice](README.md) · Antes: [manejo de errores](ERRORES.md) · Después: [analizador](ANALIZADOR.md)

Un diagnóstico explica dónde se detectó un problema y qué regla lo origina.
Empieza por el primero: un token incorrecto puede causar varios mensajes derivados.
`error` bloquea el análisis/ejecución del CLI; `warning` informa sin cambiar a
salida fallida en `joss analyze`. Los avisos del linter o del editor no prueban
por sí solos un fallo de ejecución.

El modelo `diagnostics.Diagnostic` incluye Code, Severity, Message, File,
Range (línea/columna), Explanation y Suggestion. El runtime también tiene errores
con stack de llamadas, aunque su adaptación a catch no conserva todos los campos.

## Parser, carga y símbolos

En las tablas, los fragmentos son ilustraciones de la causa; cuando se necesitan
varios archivos o contexto se indica. Las pruebas ejecutables de abajo comprueban
casos completos con su código real.

| Código | Causa / ejemplo incorrecto | Solución / caso correcto |
|---|---|---|
| JOSS-IO-001 | Archivo fuente que no se puede leer. | Corregir ruta/permisos; usar un archivo existente. |
| JOSS-PARSE-001 | Token o estructura inválida; `function...`, falta de llave o visibilidad. | Usar sintaxis canónica, por ejemplo `public func f() {}`. |
| JOSS-SYM-001 | Variable no definida: `print($x)`. | Declarar `$x = 1` antes o pasarla como parámetro. |
| JOSS-SYM-002 | Redeclaración en el mismo ámbito: dos `int $x`. | Asignar con `$x = ...` si ya existe, o elegir otro nombre. |
| JOSS-SYM-003 | Función no resuelta: `noExiste()`. | Declarar/cargar la función o corregir nombre. |
| JOSS-SYM-004 | `new Ausente()`. | Declarar clase pública o instalar plugin que la publique. |
| JOSS-SYM-005 | `extends Ausente`. | Usar una clase base existente. |
| JOSS-SYM-006 | `const $x = 1; $x = 2`. | Mantener constante o declarar variable mutable desde el inicio. |
| JOSS-DECL-001 | Dos funciones globales con igual nombre. | Renombrar o eliminar duplicado; carpetas no crean namespaces. |
| JOSS-DECL-002 | Dos clases globales con igual nombre. | Mantener una declaración en el proyecto. |
| JOSS-DECL-003 | Dos métodos homónimos en una clase. | No hay sobrecarga por firma; usar nombres diferentes. |

## Tipos y llamadas

| Código | Causa / ejemplo incorrecto | Solución / caso correcto |
|---|---|---|
| JOSS-TYPE-001 | Reasignación incompatible: `$x=1; $x="hola"`. | `$x=2`; o mixed explícito si el dominio es variable. |
| JOSS-TYPE-002 | Inicializador/default incompatible: `int $x="hola"`. | Inicializar con entero o texto numérico aceptado por coerción. |
| JOSS-TYPE-003 | Argumento incompatible con parámetro conocido. | Pasar int a `func f(int $x)`; validar antes de convertir. |
| JOSS-TYPE-004 | Operador no admite operandos conocidos. | Usar números para aritmética y punto para concatenación. |
| JOSS-TYPE-005 | Clave map con tipo no admitido. | Usar string/int según contrato, no array como clave. |
| JOSS-TYPE-006 | Índice de tipo incorrecto para colección conocida. | Array/string con entero; map con clave admitida. |
| JOSS-TYPE-007 | Indexación de un valor no indexable, como entero. | Indexar array/map/string o retirar el índice. |
| JOSS-TYPE-008 | Retorno incompatible con anotación. | `: int { return 1 }`, no texto no numérico. |
| JOSS-TYPE-009 | Tipo/clase desconocido, incluidos aliases retirados. | Usar int/float/bool/mixed/array o clase declarada. |
| JOSS-TYPE-010 | Callable anotado puede terminar sin return/throw. | Cubrir cada ruta demostrable con un resultado del tipo declarado. |
| JOSS-TYPE-011 | Parámetro sin tipo: `func($x) {}`. | `func(mixed $x) {}` o tipo concreto. |
| JOSS-CALL-001 | Cantidad incorrecta de argumentos en firma conocida. | Respetar parámetros obligatorios/defaults. |
| JOSS-MEMBER-001 | Método ausente en receptor con clase resuelta. | Corregir método; consultar registro real para clases nativas. |
| JOSS-ACCESS-001 | Función/clase private usada desde otro archivo. | API public o mantener el uso dentro de su archivo. |
| JOSS-ACCESS-002 | Miembro private/protected inaccesible. | Acceder mediante método público autorizado o desde ámbito permitido. |

Los aliases fuente eliminados `integer`, `double`, `boolean`, `dynamic`,
`any` y `list` no se convierten automáticamente en tipos canónicos. Una clase
declarada con uno de esos nombres se resuelve como clase, no como alias.

## Referencias temporales

| Código | Causa | Solución |
|---|---|---|
| JOSS-REF-001 | Falta/sobra marcador bilateral ref o cruce a nativo/async no permitido. | Declarar `ref int $x` y llamar con `f(ref $x)` a función fuente compatible. |
| JOSS-REF-002 | Se pasa expresión, campo o índice como referencia. | Pasar una variable local mutable simple. |
| JOSS-REF-003 | La variable referida es constante. | Mantener const y pasar por valor, o diseñar salida retornada. |
| JOSS-REF-004 | Tipo referido no es exactamente igual al parámetro. | Usar tipo invariante, sin ampliación int→float para ref. |
| JOSS-REF-005 | Intento de almacenar, retornar o escapar referencia. | Utilizar ref sólo durante una llamada y devolver un valor ordinario. |
| JOSS-REF-006 | Parámetro ref con default. | Exigir argumento explícito; quitar default. |

## Flujo, linter y vistas

| Código | Severidad / causa | Corrección |
|---|---|---|
| JOSS-FLOW-001 | Warning: código posterior a terminación incondicional. | Mover antes del return si debe ejecutarse o eliminarlo. |
| JOSS-LINT-001 | Warning: variable local sin uso. | Usarla o retirar declaración sin efecto necesario. |
| JOSS-SYNTAX-001 | Error sintáctico envuelto por linter. | Corregir primer error del parser; no es una gramática distinta. |
| JOSS-LINT-002 | Error de linter: parámetro sin tipo explícito. | Escribir el tipo; puede aparecer junto a TYPE-011. |
| JOSS-LINT-007 | Warning de nombres fuera de convención. | Clases PascalCase, funciones camelCase según regla del linter. |
| JOSS-SEC-001 | Warning heurístico: posible secreto literal en fuente. | Leer configuración; comprobar si el hallazgo es real sin publicar el valor. |
| JOSS-VIEW-001 | Error al compilar plantilla. | Corregir directivas, bloques o archivo requerido. |
| JOSS-VIEW-SYNTAX | Error sintáctico del Joss embebido en vista. | Corregir la expresión y su contexto de plantilla. |
| JOSS-VIEW-UNDEF | Warning heurístico de variable sin protección en ternario de vista. | Pasar dato explícitamente o proteger existencia cuando sea opcional. |

El LSP incluye heurísticas adicionales sobre texto (eval, SQL interpolado,
coste bcrypt) que no constituyen nuevos códigos estables JOSS. Una cadena
que mencione una API de otro lenguaje no demuestra que esa API exista en Joss.

## Errores aritméticos e indexación runtime

| Código | Operación inválida | Corrección |
|---|---|---|
| JOSS-ARITH-001 | Operación entera fuera del rango signed de 64 bits. | Comprobar límites o elegir decimal/otro dominio antes del cálculo. |
| JOSS-ARITH-002 | División o módulo por cero. | Comprobar divisor antes de operar. |
| JOSS-INDEX-001 | Índice negativo o >= longitud admitida. | Usar rango válido; en strings la indexación cuenta grafemas. |
| JOSS-INDEX-002 | Tipo de índice no compatible en runtime. | Convertir/validar índice entero o clave según colección. |

Estas guardias no cubren automáticamente todas las funciones nativas ni las
escrituras por índice. Otros errores runtime tienen tipo/mensaje sin código
JOSS estable; no inventes un código para ellos.

## Casos completos verificados

Tipo inexistente:

<!-- joss-error: JOSS-TYPE-009 -->
```joss-invalid
integer $edad = 20
```

Corrección:

<!-- joss-run: ["20"] -->
```joss
int $edad = 20
print($edad)
```

Parámetro sin tipo:

<!-- joss-error: JOSS-TYPE-011 -->
```joss-invalid
public func duplicar($valor) { return $valor * 2 }
```

Corrección:

<!-- joss-run: ["6"] -->
```joss
public func duplicar(int $valor): int { return $valor * 2 }
print(duplicar(3))
```

Reasignación incompatible:

<!-- joss-error: JOSS-TYPE-001 -->
```joss-invalid
$edad = 20
$edad = "veinte"
```

Corrección dinámica sólo cuando se desea ese contrato:

<!-- joss-run: ["veinte"] -->
```joss
mixed $dato = 20
$dato = "veinte"
print($dato)
```

Retorno ausente:

<!-- joss-error: JOSS-TYPE-010 -->
```joss-invalid
public func signo(int $n): string {
    $n > 0 ? { return "positivo" } : {}
}
```

Corrección:

<!-- joss-run: ["no positivo"] -->
```joss
public func signo(int $n): string {
    return $n > 0 ? "positivo" : "no positivo"
}
print(signo(0))
```

## Política para contribuidores

No inferir invalidez de unknown/mixed, metadatos nativos ausentes o tablas de
miembros no resueltas. Cada código nuevo exige un caso inválido y su vecino
válido, archivo/rango, explicación y sugerencia. Conserva códigos estables:
no uses el mensaje traducido como identificador.

Fuentes: [analizador](../pkg/analyzer/), [modelo](../pkg/diagnostics/),
[linter](../pkg/linter/linter.go), [errores runtime](../pkg/runtime/errors/).
