# Colecciones y texto Unicode

[Índice](README.md)

Antes: [funciones](FUNCIONES.md). Después: [tipos](SISTEMA_TIPOS.md) y [clases](CLASES.md).
Referencia de operaciones: [biblioteca](MODULOS_NATIVOS.md).

## Arrays: varios valores en orden

Un **array** guarda una secuencia. Cada posición tiene un **índice** entero que
empieza en cero. Usa arrays para tareas, productos o resultados de una consulta.

<!-- joss-run: ["pan", "3", "fruta"] -->
```joss
$compras = ["pan", "leche"]
print($compras[0])
$compras[] = "fruta"
print(count($compras))
print($compras[2])
```

Los corchetes vacíos en una asignación añaden al final. Leer una posición fuera
de rango o negativa produce `JOSS-INDEX-001`. Para recorrerlo basta
`foreach ($compras as $producto) { print($producto) }`.

Un array sin tipo de elemento puede reunir valores distintos. Cuando deseas
homogeneidad, `array<int> $cantidades = [1, 2, 3]` expresa una colección de enteros.
Lee sus límites de validación en [tipos](SISTEMA_TIPOS.md): la anotación no convierte
la colección en inmutable ni garantiza todas las escrituras mediante aliases.

## Maps: acceder por una clave

Un **map** asocia claves de texto con valores. A diferencia de una posición,
una clave describe qué dato buscas:

<!-- joss-run: ["Ada", "21", "sin teléfono"] -->
```joss
$persona = {"nombre": "Ada", "edad": 20}
print($persona["nombre"])
$persona["edad"] = 21
print($persona["edad"])
print($persona["telefono"] ?? "sin teléfono")
```

Una clave ausente devuelve `null`. `??` elige la derecha si la izquierda es
nula. Actualmente también oculta panics al evaluar la izquierda: mantenla
simple y no lo uses para ocultar fallos de operaciones importantes.

En contexto de expresión, `{}` es un map vacío. También puedes declarar
`map $datos` para obtenerlo como valor inicial. En lugares que exigen un cuerpo,
como una función o un ciclo, las llaves delimitan el bloque de instrucciones.
Las claves del mapa son strings, incluso si sus valores son variados.

El runtime de `foreach` recorre arrays y channels, no maps directamente:

<!-- joss-run: ["nombre: Ada"] -->
```joss
$persona = {"nombre": "Ada"}
foreach (keys($persona) as $clave) {
    print($clave . ": " . $persona[$clave])
}
```

No dependas del orden de `keys` o `values`: el mapa no tiene orden estable.
`array_key_exists("campo", $mapa)` comprueba la clave incluso si su valor es nulo.
`isset` tiene un tratamiento distinto y limitaciones para índices de mapas.

## Cambiar un nombre frente a cambiar el contenido

Los arrays usan slices de Go y los mapas usan mapas compartidos. Asignarlos a
otro nombre puede compartir los datos. No asumas una copia profunda:

<!-- joss-run: ["9", "nuevo"] -->
```joss
$original = [1, 2]
$copia = $original
$copia[0] = 9
print($original[0])
$datos = {"estado": "inicial"}
$alias = $datos
$alias["estado"] = "nuevo"
print($datos["estado"])
```

Al ampliar un array, su almacenamiento puede realojarse; dos nombres no tienen
por qué observar la misma longitud después. Para producir un array nuevo usa
`merge` o `array_merge`, que copian el contenedor, aunque los elementos anidados
todavía pueden compartirse. `const` protege el binding, no el grafo de contenido.

## Nombres que pueden confundirte

<!-- joss-run: ["2", "2", "3"] -->
```joss
$numeros = [1, 2]
print(array_pop($numeros))
print(count($numeros))
$numeros = array_push($numeros, 3)
print(count($numeros))
```

`array_pop` devuelve el último elemento y **no lo elimina**. `array_shift`
devuelve el primero sin eliminarlo. `array_push` y `append` devuelven el array
ampliado: conserva ese retorno. Para eliminar el último elemento, calcula una
porción con `array_slice($numeros, 0, count($numeros) - 1)` y reasigna.

## Texto: bytes, puntos Unicode y caracteres percibidos

Un texto UTF-8 no siempre utiliza un byte por letra. Además, un carácter que
ves puede estar compuesto por varios puntos Unicode. Por ejemplo, `e` más un
acento combinante se ve como una sola letra.

<!-- joss-run: ["3", "2", "é"] -->
```joss
$texto = "é"
print(len($texto))
print(strlen($texto))
print($texto[0])
```

| Operación | Unidad actual |
|---|---|
| `len`, `count` sobre string | Bytes UTF-8. |
| `strlen`, `substr`, `strpos` | Puntos Unicode; `strpos` devuelve `false` si no encuentra. |
| `$texto[indice]` | Grafemas extendidos: caracteres percibidos completos. |
| `Str::length` | Bytes UTF-8. |
| `Str::substring`, `Str::indexOf` | Puntos Unicode; `Str::indexOf` devuelve `-1` si no encuentra. |

Por eso `strlen($texto)` no siempre es el límite correcto para recorrerlo por
índices. La indexación protege contra cortes parciales de emojis y acentos;
las demás APIs no comparten necesariamente esa unidad. No uses una posición
obtenida con una API como índice de otra sin comprobar su unidad.

Ejercicio: crea un array de tres mapas con `nombre` y `cantidad`, recórrelo y
muestra cada nombre. Para orden predecible, conserva los registros en el array.
