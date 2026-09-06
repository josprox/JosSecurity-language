# Valores, variables y primeras operaciones

[Índice](README.md)

Antes: [primer programa](PRIMEROS_PASOS.md). Después: [decisiones y ciclos](CONTROL_FLUJO.md).
Profundizar: [sistema de tipos](SISTEMA_TIPOS.md), [biblioteca](MODULOS_NATIVOS.md).

## Valores y tipos

Un **valor** es un dato: `3`, `"Hola"` o `true`. Su **tipo** indica qué representa
y qué operaciones tienen sentido. Sumar cantidades es distinto de unir nombres.

| Tipo | Qué representa | Ejemplo | Uso habitual |
|---|---|---|---|
| `int` | Entero, sin parte decimal | `12`, `-3` | Cantidades y contadores. |
| `float` | Número con aproximación binaria | `1.5`, `0.1` | Mediciones y cálculos aproximados. |
| `decimal` | Número representado en base diez | `0.10m` | Importes cuya suma decimal debe conservarse. |
| `string` | Texto | `"Ada"` | Nombres, mensajes y contenido de archivos. |
| `bool` | Una respuesta lógica | `true`, `false` | Decidir si realizar una acción. |
| `array` | Secuencia ordenada | `["pan", "leche"]` | Varios elementos que recorrer. |
| `map` | Valores identificados por claves de texto | `{"nombre": "Ada"}` | Datos con campos. |
| `object` | Una instancia de clase | Se aprende en [clases](CLASES.md) | Datos junto con operaciones. |
| `channel` | Conducto para comunicar tareas | `make_chan(1)` | [Concurrencia](CONCURRENCIA.md). |

`null` y `nil` expresan ausencia de valor. No son texto vacío ni cero. `mixed`
permite distintos tipos en un mismo nombre; `var` pide que Joss deduzca el tipo.

## Variables: dar nombre a un dato

Una **variable** es un nombre asociado a un valor. Así puedes recordarlo y
actualizarlo. En Joss el nombre lleva `$` y distingue mayúsculas: `$edad` y
`$Edad` son nombres diferentes. Usa letras sin acentos, números después de la
primera letra y `_`.

<!-- joss-run: ["21"] -->
```joss
$edad = 20
$edad = $edad + 1
print($edad)
```

`=` guarda el valor de la derecha en el nombre de la izquierda. La segunda
línea lee `20`, suma `1` y guarda `21`. No significa igualdad matemática;
para preguntar si dos valores coinciden existe `==`.

La primera asignación **infiere** (deduce) el tipo `int`. Las posteriores deben
ser compatibles. También puedes escribir `var $edad = 20` o `int $edad = 20`.
No declares las tres variantes con el mismo nombre en el mismo ejemplo.

Ejemplo incorrecto: la cantidad empezó siendo un entero y ahora recibe texto.

<!-- joss-error: JOSS-TYPE-001 -->
```joss-invalid
$cantidad = 2
$cantidad = "muchas"
```

Solución: conserva una cantidad numérica. Si tu problema realmente requiere
alternar tipos, haz visible esa decisión:

<!-- joss-run: ["pendiente"] -->
```joss
mixed $resultado = 2
$resultado = "pendiente"
print($resultado)
```

`let $resultado = 2` tiene ese mismo dinamismo. `let` sin tipo **no** significa
constante. Empieza con inferencia y usa `mixed` cuando sea necesario.

## Constantes

Una constante es un nombre cuyo valor no puedes volver a asignar. Sirve para
una regla que quieres preservar, como el número máximo de intentos:

<!-- joss-run: ["3"] -->
```joss
const int $maximo = 3
print($maximo)
```

`const` requiere un valor inicial. Protege el nombre; no promete que todos los
datos de una colección anidada se vuelvan inmutables.

## Texto, comentarios y salida

<!-- joss-run: ["Hola, Ada", "Primera línea", "Segunda línea"] -->
```joss
// Este comentario explica el código; no se ejecuta.
$nombre = 'Ada'
print("Hola, " . $nombre)
/* Un comentario también puede
   ocupar varias líneas. */
print("Primera línea\nSegunda línea")
```

Las comillas simples y dobles delimitan texto; ambas procesan escapes como
`\n` (salto de línea) y `\t` (tabulación). No sustituyen automáticamente
`$nombre` dentro de una cadena. Usa `.` para unir texto. No uses `+`: suma números.

`print` imprime cada argumento en una línea. `printf` permite un formato,
por ejemplo `printf("Cantidad: %d\n", 3)`. `%d` es el lugar de un entero y
`%s` el de un string. No agregues `%` si no necesitas formato.

## Números

<!-- joss-run: ["5", "2.5", "1", "0.3"] -->
```joss
print(2 + 3)
print(5 / 2)
print(5 % 2)
print(0.10m + 0.20m)
```

`+`, `-` y `*` suman, restan y multiplican. `/` divide y da un `float` cuando
los operandos son enteros. `%` obtiene el resto: al dividir cinco objetos en
parejas sobra uno. Los paréntesis cambian el orden: `(2 + 3) * 4` da `20`.

`float` aproxima algunos números; `0.1 + 0.2` puede diferir de `0.3`. Usa
literales `decimal` desde el principio cuando necesites suma decimal exacta.
Esto no hace exacta una división infinita: `1m / 3m` necesita redondeo.
Consulta [precisión y conversiones](SISTEMA_TIPOS.md).

## Entrada del usuario

La entrada llega como texto y debe interpretarse. El runtime integra el flujo
`cin`, explicado en la [referencia de biblioteca](MODULOS_NATIVOS.md).
Este ejemplo requiere una persona escribiendo en la terminal:

```joss
print("Escribe tu nombre (una palabra):")
string $nombre = ""
cin >> $nombre
print("Hola, " . $nombre)
```

`cin` lee una palabra mediante `Scanln`; no es una API para leer líneas completas.
No existe una clase nativa `Console` registrada en esta revisión.
No confundas convertir con validar. `intval("hola")` retorna cero; no demuestra
que la persona haya escrito un número. Una declaración como `int $edad = "20"`
aplica la conversión de entrada tipada. Los casos límite están en la referencia.

Ejercicio: guarda precio y cantidad, calcula el total y muestra un mensaje.
Solución posible: `decimal $precio = 2.50m`, `int $cantidad = 4`, y
`print($precio * $cantidad)` muestran `10`.
