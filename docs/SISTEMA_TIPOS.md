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

Una inicialización con `nil` pospone la inferencia hasta el primer valor concreto. `nil` no es asignable a un tipo explícito no-nullable. Joss todavía no implementa tipos unión ni sintaxis nullable.

## Tipos reconocidos

Los tipos fuente canónicos son `int`, `float`, `string`, `bool`, `array`, `map`, `object`, `channel` y nombres de clase. `mixed` es dinámico; `var` solicita inferencia. Los aliases históricos `integer`, `double` y `boolean` se normalizan en `pkg/typesystem`.

Compatibilidad relevante:

- Mismo tipo: válido.
- `int` hacia `float`: válido y sin pérdida.
- Clase concreta hacia `object`: válido.
- `mixed` o información desconocida: no genera un error especulativo.
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
func add(int $a, int $b) {
    return $a + $b
}
```

El analizador valida aridad, defaults y tipos de argumentos cuando la firma es conocida. El runtime aplica la misma validación y conserva el tipo del parámetro durante todo el cuerpo. La gramática actual no incluye anotaciones de tipo de retorno; por ello el analizador inspecciona expresiones de retorno, pero no puede compararlas con una firma declarada.

## Constantes

No existe todavía un nodo o keyword de declaración `const` implementado. Los identificadores en mayúsculas pueden provenir del entorno/runtime y se conservan como tipo desconocido para evitar falsos positivos. No se documenta inmutabilidad de constantes porque el lenguaje aún no la garantiza.

## Accesos e índices

- `array` y `string` requieren índice `int`.
- `map` requiere índice `string`.
- Cuando el receptor es `mixed` o desconocido, el analizador no acusa un error sin evidencia.
- Cuando una clase está resuelta, las llamadas a métodos inexistentes producen `JOSS-MEMBER-001`.

## Fuente canónica

Toda decisión de compatibilidad debe añadirse a `pkg/typesystem` y probarse allí. El analyzer y el runtime no deben crear tablas paralelas de aliases o reglas de asignación.
