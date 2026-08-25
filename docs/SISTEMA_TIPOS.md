# Sistema de tipos, variables e inferencia

## Reglas

Joss permite sintaxis breve sin convertir las variables inferidas en dinámicas:

```joss
$age = 20        // primera asignación: infiere int
$age = 30        // válido
$age = "twenty"  // JOSS-TYPE-001
```

También puede hacerse explícita la inferencia:

```joss
var $age = 20
```

Una declaración tipada fija el tipo desde el inicio:

```joss
int $age = 20
string $name = "Ada"
let int $port = 9000
```

`let $value` es el escape dinámico explícito y se representa como `mixed`:

```joss
let $value = 20
$value = "twenty" // válido por decisión explícita
```

Una inicialización inferida con `nil` pospone la inferencia hasta el primer valor concreto. `nil` no es asignable a un tipo explícito no-nullable. Las uniones usan `|` y el atajo postfix `?` se normaliza en el AST:

```joss
int|null $count = null
int? $page = null       // exactamente el mismo tipo: int|null
func find(int|string $id): User|null { return null }
```

Una fuente unión es asignable a un destino sólo si todas sus alternativas caben en él; un valor concreto cabe en una unión cuando al menos una alternativa lo acepta. Las uniones no convierten una variable en dinámica.

## Tipos reconocidos

Los tipos fuente canónicos son `int`, `float`, `string`, `bool`, `array`, `map`, `object`, `channel` y nombres de clase declarada/nativa. `mixed` es dinámico; `var` solicita inferencia. Los antiguos aliases `integer`, `double`, `boolean`, `dynamic`, `any` y `list` ya no se normalizan: si no existe una clase con ese nombre, el analyzer emite `JOSS-TYPE-009`.

Compatibilidad relevante:

- Mismo tipo: válido.
- `int` hacia `float`: válido y sin pérdida.
- Clase concreta hacia `object`: válido.
- `mixed`: contrato dinámico explícito. `unknown`: información aún insuficiente; ninguno genera un error especulativo.
- Cualquier otro cambio conocido: error antes de ejecutar.

## Conversión de entrada

Para conservar el uso de entradas textuales en declaraciones explícitas, una cadena puede convertirse a `int`, `float` o `bool` sólo si la conversión es completa y no pierde información. La misma función `typesystem.CoerceString` es usada por el analizador y el runtime.

```joss
int $port = "9000"   // válido
int $port = "90.5"   // inválido: no se trunca
int $port = "nine"   // inválido
```

## Funciones

Los parámetros pueden tiparse:

```joss
func add(int $a, int $b): int {
    return $a + $b
}
```

El analizador valida aridad, defaults, argumentos, cada `return` y que toda ruta demostrable de una función anotada termine con `return` o `throw`. El runtime repite el contrato como defensa. La anotación posterior a `:` es opcional; sin ella el retorno permanece `unknown` y no se inventa un error. Las firmas se registran antes de analizar cuerpos, por lo que una llamada recursiva o mutuamente recursiva ve su tipo de retorno declarado.

## Constantes

`const` exige inicializador y puede inferir o declarar su tipo:

```joss
const $maximum = 10
const string $application = "Joss"
```

No puede reasignarse ni con `=`, ni con `++`, ni como propiedad constante de una instancia. El analyzer emite `JOSS-SYM-006` cuando puede resolver el símbolo y el runtime aplica la misma invariante.

## Accesos e índices

- `array` y `string` requieren índice `int`.
- `map` requiere índice `string`.
- Cuando el receptor es `mixed` o desconocido, el analizador no acusa un error sin evidencia.
- Cuando una clase está resuelta, las llamadas a métodos inexistentes producen `JOSS-MEMBER-001`.

## Fuente canónica

Toda decisión de compatibilidad debe añadirse a `pkg/typesystem` y probarse allí. El analyzer y el runtime no deben crear tablas paralelas de tipos o reglas de asignación.
