# Funciones, argumentos y ámbitos

[Índice](README.md)

Antes: [control de flujo](CONTROL_FLUJO.md). Después: [colecciones](COLECCIONES.md).
Profundizar: [recursión](RECURSION.md), [tipos](SISTEMA_TIPOS.md).

## Una operación con nombre

Una **función** reúne instrucciones para realizar una tarea. Evita repetir
código y da nombre a una intención: calcular un total o crear un saludo.
Una **llamada** pide ejecutar esa función.

<!-- joss-run: ["5", "12"] -->
```joss
public func sumar(int $a, int $b): int {
    return $a + $b
}
print(sumar(2, 3))
print(sumar(5, 7))
```

`func` declara la función; `sumar` es su nombre. `$a` y `$b` son **parámetros**:
los nombres que usará el cuerpo. `2` y `3` son **argumentos**: los valores que
entrega una llamada. `return` termina la llamada y devuelve su resultado.
El `: int` declara qué tipo de resultado espera quien llama.

`public` permite usar la función desde otros archivos del proyecto. `private`
limita su acceso al archivo. Las funciones globales requieren uno de ellos.
No existe la palabra `function`.

## Contratos y valores predeterminados

Todo parámetro requiere tipo, incluso en funciones anónimas. Si admite cualquier
valor, escribe `mixed`. El tipo de retorno es opcional, pero declararlo permite
detectar resultados incorrectos antes de ejecutar.

<!-- joss-run: ["Hola, visitante", "Hola, Ada"] -->
```joss
public func saludo(string $nombre = "visitante"): string {
    return "Hola, " . $nombre
}
print(saludo())
print(saludo("Ada"))
```

El valor después de `=` se usa cuando falta ese argumento. Coloca los parámetros
con default al final. Se permiten comas finales en firmas y llamadas. No hay
parámetros nombrados ni sintaxis de parámetros variádicos para funciones fuente.

## Retornar desde una decisión

<!-- joss-run: ["agotado", "disponible"] -->
```joss
public func disponibilidad(int $cantidad): string {
    ($cantidad <= 0) ? { return "agotado" } : {}
    return "disponible"
}
print(disponibilidad(0))
print(disponibilidad(2))
```

El primer `return` sale de la función completa, aunque esté dentro de un bloque.
Es una **guarda**: trata primero un caso que no debe seguir por el camino normal.
Una función con retorno anotado debe retornar o lanzar un error en todas las
rutas que el analizador puede demostrar. Escribir `return` en un único brazo
no basta si el otro puede llegar al final.

## Ámbito: dónde existe cada nombre

Un **ámbito** o *scope* determina qué nombres puedes usar en un lugar. Cada
llamada tiene sus propios parámetros y variables locales. La función con nombre
no hereda las variables de quien la llama ni las variables fuente de nivel superior.

<!-- joss-run: ["20", "10"] -->
```joss
public func duplicar(int $valor): int {
    $valor = $valor * 2
    return $valor
}
$numero = 10
print(duplicar($numero))
print($numero)
```

Reasignar el parámetro no cambia `$numero`. Pasar una colección o una instancia
no hace una copia profunda: modificar sus elementos puede verse desde fuera.
Consulta [colecciones](COLECCIONES.md) para distinguir el nombre del contenido.

## Funciones sin nombre y closures

Una función anónima es una función escrita como valor. Una **closure** conserva
el entorno que existía al crearla, para usarlo cuando se invoque después.
Esto permite guardar un comportamiento con sus datos:

<!-- joss-run: ["Hola, Ada"] -->
```joss
$prefijo = "Hola, "
$saludar = func(string $nombre): string {
    return $prefijo . $nombre
}
print($saludar("Ada"))
```

La closure no lleva `public` ni `private`. Su captura conserva bindings en un
entorno propio; reasignar un escalar capturado no equivale a modificar la
variable original exterior. Las invocaciones de una misma captura se serializan
mediante un mutex. No presupongas captura por referencia compartida entre
closures creadas por separado. La [arquitectura](ARQUITECTURA.md) explica el modelo.

Una **callback** es simplemente una función que entregas a otra API para que
la llame cuando corresponda, por ejemplo al recibir un mensaje WebSocket.

## Referencias temporales con ref

A veces la tarea consiste precisamente en actualizar una variable del llamador.
`ref` expresa esa intención en ambos extremos:

<!-- joss-run: ["2"] -->
```joss
public func incrementar(ref int $valor): int {
    $valor = $valor + 1
    return $valor
}
$contador = 1
incrementar(ref $contador)
print($contador)
```

Una **referencia** es un acceso temporal al mismo binding. No es una dirección de
memoria. Sólo acepta una variable mutable del mismo tipo exacto; no constantes,
expresiones, campos ni índices. No puede almacenarse, retornarse, capturarse,
tener default o cruzar llamadas nativas, plugins o async. Prefiere retornar un
nuevo valor cuando no necesitas actualizar el original.

## Pipelines y recursión

El operador `|>` entrega su izquierda como primer argumento de la derecha:

<!-- joss-run: ["ADA"] -->
```joss
print("  ada  " |> trim |> strtoupper)
```

Da `ADA`: primero recorta espacios y después convierte a mayúsculas. También
admite `valor |> funcion(otroArgumento)`. No implica procesamiento paralelo.

Una función **recursiva** se llama a sí misma. Debe tener un caso que termine
sin volver a llamar. Aprende a construirlo en [recursión](RECURSION.md); el
runtime limita por defecto la profundidad a 1024 llamadas.

Ejercicio: escribe `public func cuadrado(int $n): int` y prueba `cuadrado(4)`.
Debe devolver `16`. Después explica por qué imprimir dentro de la función no
es lo mismo que retornar el resultado a su llamador.
