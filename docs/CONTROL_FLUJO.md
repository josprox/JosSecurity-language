# Decisiones y repetición

[Índice](README.md)

Antes: [fundamentos](FUNDAMENTOS.md). Después: [funciones](FUNCIONES.md).
Referencia: [operadores y sentencias](SINTAXIS.md).

El **flujo** es el orden en que se ejecutan las instrucciones. Hasta ahora
hemos seguido de arriba abajo. Una condición permite elegir y un ciclo permite
repetir sin copiar el mismo código muchas veces.

## Elegir entre dos resultados

Una condición suele ser una pregunta con resultado `bool`:

<!-- joss-run: ["true", "Entrada permitida"] -->
```joss
$edad = 20
print($edad >= 18)
print(($edad >= 18) ? "Entrada permitida" : "Debes esperar")
```

`>=` pregunta «¿es mayor o igual?». El **ternario** tiene tres partes:
`condición ? resultado_si_verdadera : resultado_si_falsa`. Sólo evalúa el
resultado seleccionado. Joss no tiene `if`, `else`, `elif` ni `switch`.

Para ejecutar varias instrucciones, escribe un bloque entre `{` y `}`:

<!-- joss-run: ["Hay existencias", "Preparando pedido"] -->
```joss
$existencias = 4
($existencias > 0) ? {
    print("Hay existencias")
    print("Preparando pedido")
} : {
    print("Producto agotado")
}
```

Si una rama no necesita hacer nada, usa `{}`. Evita cadenas enormes de ternarios:
una función con nombre o un `match` hará más legible una decisión compleja.

## Comparar y combinar

| Operador | Pregunta |
|---|---|
| `==`, `!=` | ¿Coinciden?, ¿son diferentes? |
| `===`, `!==` | ¿Coinciden también sus tipos?, ¿no coinciden estrictamente? |
| `<`, `<=`, `>`, `>=` | ¿Es menor, menor o igual, mayor, mayor o igual? |
| `&&` | ¿Se cumplen ambas condiciones? |
| `\|\|` | ¿Se cumple al menos una? |
| `!` | Invierte la respuesta. |

`&&` y `||` evitan evaluar la derecha cuando la izquierda decide el resultado.
En Joss tienen la misma precedencia: usa paréntesis cuando los combines.
Aunque el runtime acepta otros valores como condición, empieza por comparar
explícitamente: `$importe > 0` es más claro que depender de cómo se interpreta
un importe. La [referencia](SINTAXIS.md) detalla esa interpretación.

## Elegir entre varios valores con match

<!-- joss-run: ["En camino"] -->
```joss
$estado = "enviado"
$mensaje = match ($estado) {
    "nuevo" => "Preparando",
    "enviado", "reparto" => "En camino",
    default => "Consulta el pedido"
}
print($mensaje)
```

Cada **brazo** relaciona valores con un resultado usando `=>`. `default` cubre
los demás. El primer brazo que coincide es el elegido; no continúa al siguiente.
Es selección de valores, no desestructuración de objetos ni patrones de tipos.

## Repetir mientras se cumpla una condición

<!-- joss-run: ["1", "2", "3"] -->
```joss
$numero = 1
while ($numero <= 3) {
    print($numero)
    $numero++
}
```

`while` comprueba antes de cada vuelta. `++` aumenta el contador en uno.
Sin esa actualización, la condición seguiría verdadera y el ciclo no terminaría.
Si necesitas ejecutar el cuerpo al menos una vez, usa `do … while`:

<!-- joss-run: ["Intento 1"] -->
```joss
$intento = 0
do {
    $intento++
    print("Intento " . $intento)
} while ($intento < 1)
```

## Recorrer una colección

Un `array` reúne varios valores en orden. `foreach` asigna cada uno a una
variable y ejecuta el bloque:

<!-- joss-run: ["pan", "leche"] -->
```joss
$compras = ["pan", "leche"]
foreach ($compras as $producto) {
    print($producto)
}
```

La sintaxis admite una variable después de `as`; no admite `clave => valor`.
Para recorrer un mapa, recorre primero `keys($mapa)` y consulta `$mapa[$clave]`.
El orden de sus claves no está garantizado. Para posiciones de un array, usa
un contador o `while` con índices que empiezan en cero.

`break` termina el ciclo; `continue` salta a la siguiente vuelta:

<!-- joss-run: ["1"] -->
```joss
foreach ([1, 2, 3, 4] as $numero) {
    print($numero)
    break
}
```

Este programa termina tras el primer elemento. Hay una limitación actual:
el plan de ejecución no detecta `break`/`continue` que aparecen únicamente
dentro de un ternario o `match`; pueden escapar como un fallo interno.
Hasta corregirla, controla la salida condicional con la condición de `while`
y filtra con ramas que elijan qué trabajo realizar. No dependas de ese fallo
para controlar el programa. Está registrado en la [auditoría](DOCUMENTATION_AUDIT.md).

La variable de `foreach` puede reutilizar un nombre existente y debe respetar
su tipo. No uses `break` ni `continue` fuera de un ciclo. No existe `for` clásico.

Ejercicio: recorre `[3, 7, 2]`, acumula su suma en `$total = 0` y muestra `12`.
Elige `foreach` porque ya tienes los elementos; elige `while` cuando lo que
controla la repetición es una condición que cambia.
