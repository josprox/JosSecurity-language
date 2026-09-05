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

Joss es fuertemente tipado por defecto; no existe una bandera en `joss.yaml`. Sólo `let $value` y una declaración explícita `mixed $value` crean bindings dinámicos. Los parámetros nunca reciben `mixed` implícitamente: deben declarar un tipo o escribir `mixed` de forma visible.

Una inicialización inferida con `nil` pospone la inferencia hasta el primer valor concreto. `nil` no es asignable a un tipo explícito no-nullable. Las uniones usan `|` y el atajo postfix `?` se normaliza en el AST:

```joss
int|null $count = null
int? $page = null       // exactamente el mismo tipo: int|null
public func find(int|string $id): User|null { return null }
```

Una fuente unión es asignable a un destino sólo si todas sus alternativas caben en él; un valor concreto cabe en una unión cuando al menos una alternativa lo acepta. Las uniones no convierten una variable en dinámica.

## Tipos reconocidos

Los tipos fuente canónicos son `int`, `float`, `decimal`, `string`, `bool`, `array`, `map`, `object`, `channel` y nombres de clase declarada/nativa. `mixed` es dinámico; `var` solicita inferencia. Los antiguos aliases `integer`, `double`, `boolean`, `dynamic`, `any` y `list` ya no se normalizan: si no existe una clase con ese nombre, el analyzer emite `JOSS-TYPE-009`.

Compatibilidad relevante:

- Mismo tipo: válido.
- `int` hacia `float`: válido y sin pérdida.
- `int` y `float` hacia `decimal`: válido (conversión exacta en base 10).
- Clase concreta hacia `object`: válido.
- `mixed`: contrato dinámico explícito. `unknown`: información aún insuficiente; ninguno genera un error especulativo.
- Cualquier otro cambio conocido: error antes de ejecutar.

## Conversión de entrada

Para conservar el uso de entradas textuales en declaraciones explícitas, una cadena puede convertirse a `int`, `float`, `decimal` o `bool` sólo si la conversión es completa y no pierde información. La misma función `typesystem.CoerceString` es usada por el analizador y el runtime.

```joss
int $port = "9000"       // válido
decimal $monto = "49.99" // válido
int $port = "90.5"       // inválido: no se trunca
int $port = "nine"       // inválido
```

## Funciones

Los parámetros siempre deben tiparse. `mixed` se permite sólo cuando el contrato realmente es dinámico:

```joss
public func add(int $a, int $b): int {
    return $a + $b
}

public func passthrough(mixed $value): mixed {
    return $value
}
```

Una firma `func passthrough($value)` es inválida (`JOSS-TYPE-011`).

## Referencias seguras

`ref` permite que una función modifique el binding del llamador sin direcciones de memoria ni desreferenciación manual:

```joss
public func increment(ref int $value): int {
    $value = $value + 1
    return $value
}

$count = 1
increment(ref $count)
```

La referencia es temporal y mutable: requiere una variable no constante, coincidencia exacta e invariante de tipo y la marca `ref` tanto en la firma como en la llamada. No puede guardarse, retornarse, capturarse ni enviarse a una API que no declare un parámetro `ref`. No admite defaults, `nil`, aritmética de punteros ni memoria manual.

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
- `string[index]` cuenta _extended grapheme clusters_ de Unicode (caracteres
  percibidos), no bytes ni code points aislados. Acentos combinados, flags y
  emojis unidos por ZWJ se retornan completos.
- Un índice negativo o fuera de rango produce `JOSS-INDEX-001`; nunca retorna
  bytes UTF-8 parciales ni continúa silenciosamente con `nil`.
- Cuando el receptor es `mixed` o desconocido, el analizador no acusa un error sin evidencia.
- Cuando una clase está resuelta, las llamadas a métodos inexistentes producen `JOSS-MEMBER-001`.

## Aritmética entera

Los enteros Joss son valores signed de 64 bits. `+`, `-`, `*`, `%`, negación y
`++` se evalúan de forma exacta. Un resultado fuera de rango produce
`JOSS-ARITH-001`; división o módulo entre cero produce `JOSS-ARITH-002`. No hay
wrapping ni saturación implícitos. `/` conserva el resultado `float` histórico,
pero también rechaza divisor cero.

## Aritmética de punto flotante (`float`) y límites IEEE 754

El tipo `float` en Joss se implementa bajo el estándar IEEE 754 de 64 bits (doble precisión binaria).

Al trabajar en base 2, los números fraccionarios decimales que no son potencias de dos (como `0.1 = 1/10` o `0.2 = 1/5`) no poseen una representación binaria finita y sufren pérdidas residuales por redondeo:

```joss
float $a = 0.1 + 0.2
float $b = 0.3

($a == $b) ? {
    print("Iguales")
} : {
    print("Diferentes") // Se ejecuta esta rama: diferencia de 5.551115123125783e-17
}
```

**Regla de oro:** No use `float` para montos monetarios, balances bancarios o contabilidad crítica. En comparaciones financieras como `$saldo >= $precio`, pequeñas desviaciones de redondeo pueden producir decisiones erróneas (por ejemplo, con `$saldo = 0.30` y `$precio = 0.10 + 0.20`, Joss evaluará `$precio` como `0.30000000000000004` y la condición `$saldo >= $precio` resultará falsa).

## Precisión fija y transacciones financieras (`decimal`)

Para cálculos monetarios, balances bancarios, liquidación de impuestos o cualquier operación donde las pérdidas por redondeo binario sean inadmisibles, Joss proporciona el tipo nativo `decimal`:

- **Aritmética exacta en Base 10**: Implementado con precisión arbitraria decimal de punto fijo/escala dinámica (respaldado por `shopspring/decimal`).
- **Sintaxis de literales**: Sufijo `m` o `M` (ej. `0.10m`, `0.20m`, `1500.50M`).
- **Coerción sin pérdida**: Acepta asignaciones desde enteros, floats o cadenas numéricas (ej. `decimal $saldo = 100` o `decimal $tasa = "0.05"`).
- **Operaciones completas**: Soporta `+`, `-`, `*`, `/`, `%`, negación unaria `-$d`, operadores relacionales `<`, `>`, `<=`, `>=`, `==`, `!=`, y comparador de nave espacial `<=>`.
- **Funciones integradas**:
  - `decimal($val)`: convierte o construye un valor `decimal`.
  - `is_decimal($val)`: verifica si una variable contiene un valor `decimal`.

```joss
// Aritmética bancaria exacta
decimal $saldo = 0.60m
decimal $precio = 0.10m + 0.20m

// La comparación es 100% exacta: 0.10m + 0.20m da exactamente 0.30m
($saldo >= $precio) ? {
    decimal $resto = $saldo - $precio
    print("Compra exitosa. Saldo restante: " . $resto) // Muestra exactamente: 0.3
} : {
    print("Fondos insuficientes")
}
```

## Fuente canónica

Toda decisión de compatibilidad debe añadirse a `pkg/typesystem` y probarse allí. El analyzer y el runtime no deben crear tablas paralelas de tipos o reglas de asignación.
