# Colecciones: Arrays, Maps y Manipulación de Texto Unicode

Antes: [Funciones y closures](FUNCIONES.md). Después: [Sistema de tipos e inferencia](SISTEMA_TIPOS.md).
Referencia técnica: [Módulos nativos](MODULOS_NATIVOS.md), [Funciones globales](FUNCIONES_GLOBALES.md).

---

## ¿Qué vas a aprender aquí?

Hasta ahora hemos trabajado con variables que almacenan un solo dato a la vez: un número, un nombre o un booleano. Pero en la vida real los datos casi nunca vienen aislados:
- Una lista de productos en una tienda online.
- Los comentarios de una publicación.
- La ficha de un usuario con su nombre, correo, teléfono y dirección.

Para agrupar y organizar múltiples datos en una sola estructura, existen las **colecciones**.

En esta guía aprenderás:
1. Qué es un **array**, cómo funciona la numeración desde cero (índices) y cómo añadir elementos.
2. Qué es un **map** (diccionario clave-valor) y cómo estructurar registros de datos.
3. Cómo recorrer colecciones y por qué los mapas se recorren a través de sus claves.
4. Cómo se comportan las colecciones en la memoria: copia de referencias vs duplicación de datos.
5. Las peculiaridades de funciones como `array_pop`, `array_push` y `array_shift` en Joss.
6. El texto como colección: la diferencia fundamental entre **bytes**, **puntos de código Unicode** y **grafemas (caracteres visibles)**.

---

## 1. Arrays: Secuencias ordenadas de elementos

Un **array** es una lista ordenada de valores. Cada valor ocupa una casilla numerada llamada **índice**.

En Joss (y en la inmensa mayoría de lenguajes modernos), **los índices comienzan a contar desde cero (`0`)**, no desde uno:
- El primer elemento está en el índice `0`.
- El segundo elemento está en el índice `1`.
- El tercer elemento está en el índice `2`.

<!-- joss-run: ["pan", "3", "fruta"] -->
```joss
$compras = ["pan", "leche"]
print($compras[0])
$compras[] = "fruta"
print(count($compras))
print($compras[2])
```

### Operaciones básicas con arrays:

1. **Creación**: Se delimitan con corchetes `[` y `]`, separando los elementos por comas: `["pan", "leche"]`.
2. **Lectura por índice**: `$compras[0]` accede al primer elemento (`"pan"`).
3. **Añadir al final con `[]`**: Escribir `$compras[] = "fruta"` agrega automáticamente el nuevo elemento al final de la lista.
4. **Contar elementos**: `count($compras)` (o `len($compras)`) devuelve la cantidad total de elementos (en este caso, `3`).
5. **Protección de límites**: Si intentas acceder a un índice que no existe (por ejemplo `$compras[99]` o un índice negativo `$compras[-1]`), Joss detiene la ejecución de inmediato con el error de seguridad `JOSS-INDEX-001` (Index Out of Range), protegiendo tu programa de leer memoria basura.

### Tipado de colecciones: Arrays homogéneos
Por defecto, un array `array` puede contener tipos mezclados. Si quieres garantizar que todos los elementos sean números enteros, puedes usar la sintaxis parametrizada:

```joss
array<int> $edades = [18, 25, 30]
```

---

## 2. Maps: Diccionarios de Clave → Valor

Un array es perfecto cuando el orden de los elementos importa (como una lista de espera). Pero si quieres representar una entidad con propiedades etiquetadas (como un usuario), recordar que "el nombre está en el índice 0 y el correo en el 1" es frágil y confuso.

Para eso existen los **maps** (también conocidos como diccionarios, tablas hash o mapas asociativos). En un map, cada valor se guarda y se recupera mediante una **clave de texto**:

<!-- joss-run: ["Ada", "21", "sin teléfono"] -->
```joss
$persona = {"nombre": "Ada", "edad": 20}
print($persona["nombre"])
$persona["edad"] = 21
print($persona["edad"])
print($persona["telefono"] ?? "sin teléfono")
```

### Características de los Maps en Joss:
1. **Creación**: Se delimitan con llaves `{` y `}`, asociando cada clave con su valor mediante dos puntos `:`: `{"clave": valor}`.
2. **Mapa vacío**: `{}` crea un mapa vacío.
3. **Acceso y modificación**: Se utilizan corchetes con el nombre de la clave entre comillas: `$persona["nombre"]`.
4. **Claves inexistentes**: Si intentas leer una clave que no existe (como `$persona["telefono"]`), Joss devuelve `null` de forma segura en lugar de fallar. Puedes usar el operador `??` para proveer un valor predeterminado elegante.
5. **Verificación de existencia**: Puedes usar `array_key_exists("telefono", $persona)` para saber con certeza si una clave fue definida, incluso si su valor asociado es `null`.

---

## 3. Recorrer Maps con `foreach` y `keys()`

En Joss, la sentencia `foreach` recorre directamente secuencias ordenadas (arrays) y canales de comunicación concurrentes (`channel`). Como los mapas son tablas asociativas internas en Go sin orden secuencial fijo, para recorrer un map se utiliza la función nativa `keys()`:

<!-- joss-run: ["nombre: Ada"] -->
```joss
$persona = {"nombre": "Ada"}
foreach (keys($persona) as $clave) {
    print($clave . ": " . $persona[$clave])
}
```

- `keys($persona)`: Devuelve un array con todos los nombres de las claves del mapa.
- Luego, `foreach` recorre esa lista de nombres y podemos consultar `$persona[$clave]`.
- También existe `values($persona)`, que devuelve un array únicamente con los valores contenidos en el mapa.

---

## 4. Comportamiento en memoria: ¿Copia o referencia?

Este es un concepto fundamental en la arquitectura de Joss:

- Cuando asignas un número o un texto a otra variable (`$b = $a`), se copia el valor de forma independiente.
- En cambio, los **arrays** y los **maps** se gestionan internamente mediante punteros y estructuras compartidas (slices y maps de Go).

Si asignas un array o map existente a una nueva variable, **ambas variables apuntan a la misma estructura de datos en memoria**:

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

Modificar `$copia[0]` alteró también a `$original[0]`, porque ambas son dos nombres distintos para el mismo array físico.

> [!TIP]
> **¿Cómo crear una copia independiente?**
> Para duplicar un array sin compartir cambios futuros, utiliza la función `merge`:

```joss
$clon = merge([], $original)
```

---

## 5. Nombres de funciones que debes conocer bien

Algunas funciones para manipular arrays tienen contratos específicos en Joss que difieren de lenguajes como PHP o JavaScript:

<!-- joss-run: ["2", "2", "3"] -->
```joss
$numeros = [1, 2]
print(array_pop($numeros))
print(count($numeros))
$numeros = array_push($numeros, 3)
print(count($numeros))
```

Presta atención a estos detalles:
1. `array_pop($arr)`: Devuelve el último elemento, pero **no lo elimina del array original** (no muta la longitud).
2. `array_shift($arr)`: Devuelve el primer elemento sin eliminarlo.
3. `array_push($arr, $item)` y `append($arr, $item)`: Toman el array, le agregan el nuevo elemento y **retornan el nuevo array resultante**. Por eso debes reasignar: `$numeros = array_push($numeros, 3)` o usar la sintaxis directa `$numeros[] = 3`.
4. `array_slice($arr, $inicio, $longitud)`: Extrae una porción del array sin modificar el original.

---

## 6. Manipulación de Texto: Bytes, Runas y Grafemas

El texto digital moderno es mucho más complejo que las letras en inglés del teclado ASCII. Cuando manejas texto con acentos (`á`, `é`), caracteres asiáticos o emojis (`😀`, `👨‍👩‍👧‍👦`), una sola letra visual puede estar compuesta por varios bytes e incluso por varios caracteres Unicode combinados.

En Joss:

<!-- joss-run: ["3", "2", "é"] -->
```joss
$texto = "é"
print(len($texto))
print(strlen($texto))
print($texto[0])
```

Observa la diferencia de las tres líneas para la letra `é` (letra `e` con tilde combinada):
1. `len($texto)`: Devuelve **3 bytes** en formato UTF-8 físico.
2. `strlen($texto)`: Devuelve **2 puntos de código Unicode** (la `e` base + el acento combinante).
3. `$texto[0]`: Devuelve el **grafema visual completo** (`é`).

| Operación | Unidad que mide | Uso recomendado |
|---|---|---|
| `len($texto)` | Bytes físicos | Tamaños de archivo, transferencias de red, buffers en disco. |
| `strlen($texto)` | Puntos de código (Runas) | Algoritmos de análisis de texto estándar. |
| `$texto[$i]` | Grafemas extendidos (caracteres visibles) | **Manipulación de texto orientada al usuario**: cortar nombres, mostrar avatares, indexar sin partir emojis por la mitad. |

---

## 7. Resumen de funciones útiles para colecciones

| Función | Propósito | Ejemplo |
|---|---|---|
| `count($arr)` / `len($arr)` | Devuelve la longitud de una colección o texto. | `count([1, 2])` → `2` |
| `in_array($val, $arr)` | Verifica si un valor existe en el array. | `in_array(2, [1, 2, 3])` → `true` |
| `keys($map)` | Obtiene la lista de claves de un mapa. | `keys({"a": 1})` → `["a"]` |
| `values($map)` | Obtiene la lista de valores de un mapa. | `values({"a": 1})` → `[1]` |
| `explode($sep, $str)` | Divide un texto en un array usando un separador. | `explode(",", "a,b,c")` → `["a", "b", "c"]` |
| `implode($sep, $arr)` | Une un array de textos en una sola cadena. | `implode("-", ["2026", "09", "05"])` → `"2026-09-05"` |
| `array_reverse($arr)` | Invierte el orden de los elementos. | `array_reverse([1, 2, 3])` → `[3, 2, 1]` |

---

## 8. Ejercicios prácticos

1. **Gestión de inventario**:
   - Crea un mapa llamado `$producto` con las claves `"nombre"` (`"Laptop"`), `"precio"` (`1200.00m`) y `"stock"` (`5`).
   - Muestra en la terminal: `"Producto: Laptop | Precio: $1200.00 | Disponibles: 5"`.
   - Simula una compra restándole 1 al stock y muestra el nuevo valor.
2. **Filtro de palabras con `explode` e `implode`**:
   - Crea un texto `$frase = "manzana,pera,uva,platano"`.
   - Conviértelo en un array usando `explode`.
   - Recorre el array con `foreach` e imprime cada fruta en mayúsculas usando el operador pipeline: `$fruta |> strtoupper`.

---

## Siguiente paso

Ahora que dominas las estructuras de datos en memoria, es momento de profundizar en cómo el analizador de tipos de Joss verifica contratos, cómo funciona la inferencia y cómo se manejan valores nulos de forma segura.

Continúa con: [Sistema de tipos, inferencia y conversiones](SISTEMA_TIPOS.md).
