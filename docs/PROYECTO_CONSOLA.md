# Proyecto completo: informe de compras

[Índice](README.md) · Antes: [funciones](FUNCIONES.md), [colecciones](COLECCIONES.md), [errores](ERRORES.md) · Después: [concurrencia](CONCURRENCIA.md)

Vamos a guardar compras en un archivo y calcular el total. Un archivo conserva
los datos después de cerrar el programa; una variable sólo vive durante la ejecución.
Usaremos JSON, un formato de texto para intercambiar arrays y mapas.

## Preparar el directorio

Crea una carpeta vacía llamada `compras` y ábrela en la terminal. Guarda el
siguiente programa completo como `main.joss`. No requiere servidor, base de
datos ni plugins. Escribe únicamente `compras.json` dentro del directorio actual.

<!-- joss-run: ["Articulos: 2", "Total: 42.5", "Archivo guardado"] -->
```joss
public func totalizar(array $compras): decimal {
    decimal $total = 0m
    foreach ($compras as $compra) {
        $total = $total + decimal($compra["precio"]) * intval($compra["cantidad"])
    }
    return $total
}

$compras = [
    {"nombre": "Cuaderno", "precio": "12.50", "cantidad": 2},
    {"nombre": "Lapiz", "precio": "3.50", "cantidad": 5}
]

$guardado = file_put_contents("compras.json", json_encode($compras))
$guardado ? {} : { throw "No se pudo guardar compras.json" }

$texto = file_get_contents("compras.json")
$texto == null ? { throw "No se pudo leer compras.json" } : {}
json_verify($texto) ? {} : { throw "El archivo no contiene JSON valido" }

$leidas = json_decode($texto)
print("Articulos: " . count($leidas))
print("Total: " . totalizar($leidas))
print("Archivo guardado")
```

## Ejecutar y comprender

```sh
joss analyze main.joss
joss run main.joss
```

La salida del programa es:

```text
Articulos: 2
Total: 42.5
Archivo guardado
```

El CLI puede añadir mensajes de preparación. `totalizar` recibe las compras
como parámetro: una función con nombre no hereda variables locales del caller.
Cada vuelta obtiene un mapa; `precio` se guarda como texto para no introducir
redondeo binario al decodificar JSON. `decimal(...)` lo convierte para calcular.

Las funciones de archivo devuelven `false` o `null` al fallar. El programa
comprueba esas señales antes de continuar. `json_verify` sólo valida sintaxis
JSON: no garantiza que los registros tengan nombre, precio y cantidad.
Para aceptar archivos ajenos, añade validación de estructura y valores antes
de llamar a `totalizar`; los convertidores permisivos pueden retornar cero
ante texto inválido.

## Cambiar el programa

1. Añade otra compra y predice el total antes de ejecutar.
2. Cambia el nombre del archivo en las dos operaciones y observa dónde aparece.
3. Separa la carga en una función que reciba la ruta.
4. Para conservar compras editadas manualmente, elimina la escritura inicial:
   ahora el archivo será la entrada del programa, no un ejemplo que se regenera.

No uses `run("archivo.joss")` para dividir código: ese helper inicia Python/PHP.
Para crecer dentro de Joss, consulta [organización sin imports](MODULOS_IMPORTS.md).
