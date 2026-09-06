# Sistema de tipos: contratos y límites

Aprender antes: [fundamentos](FUNDAMENTOS.md), [funciones](FUNCIONES.md),
[colecciones](COLECCIONES.md). Después: [clases](CLASES.md).
Consulta: [diagnósticos](DIAGNOSTICOS.md).

Un tipo describe qué valores acepta un lugar del programa. Un contrato de tipo
ayuda a detectar que una cantidad recibió un nombre o que una función devolvió
un dato distinto del esperado. **Inferencia** significa deducir ese contrato
del código; **conversión** significa producir un valor de otro tipo.

## Tipos fuente y representación

| Tipo fuente | Significado y ejemplo | Representación / límite |
|---|---|---|
| `int` | Entero: `42` | Habitualmente `int64` con signo; rango −2⁶³ a 2⁶³−1. |
| `float` | Aproximación numérica: `1.25` | `float64` binario. |
| `decimal` | Decimal base diez: `1.25m` | `shopspring/decimal.Decimal`, coeficiente arbitrario y escala. |
| `string` | Texto: `"hola"` | String Go UTF-8; indexación por grafemas. |
| `bool` | Lógica: `true`/`false` | Bool Go. |
| `array` | Secuencia: `[1, "a"]` | `[]interface{}`; puede contener valores heterogéneos. |
| `map` | Asociación: `{"id": 1}` | `map[string]interface{}`; claves string. |
| `object` | Instancia sin exigir clase concreta | Acepta clases; no equivale a «cualquier valor». |
| `channel` | Comunicación: `make_chan(1)` | `*core.Channel` con canal Go de valores dinámicos. |
| Nombre de clase | `Persona`, `GranDB` | Clase declarada/nativa/plugin que debe poder resolverse. |
| `mixed` | Dinamismo explícito | Acepta cambios de tipo, no prueba que una operación sea segura. |
| `null` / `nil` | Ausencia | Mismo valor; utilizable en uniones. |
| `var` | Solicitud de inferencia | No es una clase de valores ni dinamismo. |

`unknown` es un estado interno del checker para información insuficiente.
El parser de tipos lo reconoce, pero no es la elección para expresar una API
dinámica: usa `mixed`. Future, callable, slots y valores host no tienen
necesariamente un tipo fuente que reproduzca su representación Go.

## Elegir la declaración

<!-- joss-run: ["30", "Ada", "pendiente"] -->
```joss
var $edad = 20
$edad = 30
string $nombre = "Ada"
mixed $resultado = 10
$resultado = "pendiente"
print($edad)
print($nombre)
print($resultado)
```

`$edad = 20` también fija el tipo inferido; `let string $nombre` equivale
a la forma tipada. `let $resultado` equivale a `mixed`, no a constante.
No existe un modo de tipado configurable en `joss.yaml`.

Una inferencia iniciada con `nil`/nulo se pospone hasta el primer valor
concreto. Una anotación explícita no-nullable no acepta nulo.
Las declaraciones sin inicializador usan ceros para primitivas y contenedores;
para objetos/canales no hay una instancia válida construida automáticamente:
inicialízalos de forma explícita.

## Uniones y ausencia

Una **unión** admite alternativas concretas sin volverse dinámica:

<!-- joss-run: ["10", "A-10", "sin dato"] -->
```joss
int|string $id = 10
print($id)
$id = "A-10"
print($id)
int? $cantidad = null
print($cantidad ?? "sin dato")
```

`int?` se normaliza a `int|null` en el AST. Una fuente unión cabe en un
destino sólo si todas sus alternativas caben; un valor concreto cabe si una
alternativa del destino lo admite. El checker refina comparaciones directas
contra nulo (`==`, `===`, `!=`, `!==`) en las ramas del ternario.
No es análisis general de flujo, contratos SQL ni taint.

## Compatibilidad

| Origen → destino | Regla implementada |
|---|---|
| Mismo tipo | Aceptado; las clases comparan nombre. |
| `int → float` | Aceptado. No garantiza precisión de todos los enteros al hacer aritmética float. |
| `int/float → decimal` | Aceptado; el runtime convierte a representación decimal. |
| Clase → `object` | Aceptado. |
| Subclase → superclase | Analyzer/runtime consultan jerarquía; no basta la comparación canónica de nombres. |
| `mixed` o `unknown` involucrado | El checker evita acusar incompatibilidad sin información suficiente. |
| Otros tipos conocidos | Rechazados salvo conversión explícita/entrada textual admitida. |

Asignable no significa convertido: una variable anotada float puede conservar
un entero Go cuando recibe un entero, porque el runtime no convierte todos
los valores no-string. No deduzcas la representación ni la igualdad estricta
sólo de la anotación.

## Colecciones parametrizadas

`array<T>` y `map<K, V>` existen; no hay funciones/clases genéricas fuente.

<!-- joss-run: ["6", "2"] -->
```joss
array<int> $cantidades = [2, 4, 6]
map<string, int> $inventario = {"pan": 2}
print($cantidades[2])
print($inventario["pan"])
```

El analyzer infiere elementos de arrays homogéneos y retorna el tipo del elemento
al indexar una colección anotada. Para maps literales conserva información
menos precisa. El runtime comprueba elementos al validar un array/map anotado,
pero sus escrituras por índice no hacen la misma comprobación; los aliases
pueden invalidar el contrato. En maps valida valores, no el parámetro de clave:
la representación real sigue exigiendo strings. No declares `map<int,V>`.

La forma compacta de genéricos anidados `array<array<int>>` puede tropezar con
el token `>>`; separa cierres (`array<array<int> >`) y verifica con el parser.
Las uniones de colecciones y conversiones decimales tienen comprobaciones
menos profundas que una colección simple. No son contratos de memoria seguros.

## Conversión textual tipada

El código común `typesystem.CoerceString` recorta espacios y admite:

| Destino | Entrada admitida por la implementación |
|---|---|
| `int` | Entero decimal, o float textual integral y dentro de rango. |
| `float` | Texto aceptado por `strconv.ParseFloat`. |
| `decimal` | Texto numérico; el runtime usa además `decimal.NewFromString`. |
| `bool` | Sin distinguir mayúsculas: `true/1/yes`; `false/0/no/""`. |

<!-- joss-run: ["9000", "49.99", "true"] -->
```joss
int $puerto = "9000"
decimal $precio = "49.99"
bool $activo = "yes"
print($puerto)
print($precio)
print($activo)
```

Un texto literal incompatible permite diagnóstico antes de ejecutar. Con entrada
calculada, la defensa runtime decide. `int $x = "90.5"` no trunca, mientras
`intval("90.5")` sí retorna `90`. `intval("texto")` y `decimal("texto")`
retornan cero. Esos helpers no sustituyen una validación.

La intención declarada es conversión completa, pero no hay garantía universal
de ausencia de pérdida: el camino textual integral vía float puede redondear
y ParseFloat admite valores especiales. Decimal tiene caminos distintos en
checker/runtime. Es una discrepancia de implementación registrada en la
[auditoría](DOCUMENTATION_AUDIT.md), no una promesa de exactitud absoluta.

## Precisión numérica

Los operadores enteros `+`, `-`, `*`, `%`, negación y `++` se
comprueban contra overflow (`JOSS-ARITH-001`). Dividir o tomar módulo por
cero produce `JOSS-ARITH-002`. El literal positivo 2⁶³ no cabe; para construir
el mínimo entero usa `-9223372036854775807 - 1`.

`float` usa representación binaria de 64 bits. Mezclar enteros grandes con
float puede perder precisión por encima de 2⁵³. No es correcto afirmar que todo
`int → float` conserva el valor exacto.

`decimal` evita el redondeo binario de una suma como `0.10m + 0.20m`.
Su división usa `Div` de la dependencia (precisión de división predeterminada:
16 lugares decimales en la versión fijada); no promete representar 1/3 exactamente.
Convertir un float ya redondeado a decimal no recupera su intención original.

## Funciones, constantes y referencias

Los parámetros mantienen el tipo durante el cuerpo. Todo parámetro requiere
tipo (`JOSS-TYPE-011`); el retorno anotado se valida en cada `return`
(`JOSS-TYPE-008`) y debe cubrir rutas demostrables (`JOSS-TYPE-010`).
Una función sin anotación tiene retorno desconocido para el checker.

`const` protege contra reasignación; no implementa inmutabilidad profunda.
`ref T $x` y `f(ref $v)` exigen el mismo tipo exacto, variable no constante
y vida temporal limitada a la llamada. No admiten defaults, almacenamiento,
retorno, captura, índices/campos ni fronteras native/async/plugin.

Los errores de miembro se emiten sólo cuando el receptor y la tabla de miembros
están resueltos. Las APIs nativas sin metadatos de parámetros no tienen
comprobación especulativa de aridad. Algunas firmas de retorno publicadas
todavía discrepan del handler: consulta [biblioteca](MODULOS_NATIVOS.md).

Fuente de reglas: `pkg/typesystem/types.go`, `pkg/analyzer/infer.go`,
`pkg/core/evaluator_utils.go` y `pkg/core/frame_runtime.go`.


[Índice](README.md)
