# Concurrencia

## Async y await

Joss permite ejecutar bloques de código de forma asíncrona en goroutines nativas utilizando la sintaxis de bloque `async`:

```joss
// Lanzar una goroutine en segundo plano
async {
    $downloader = new DocDownloader()
    $downloader->sync(false)
}

// O capturar el resultado en un Future para await
$future = async {
    return 20 + 22
}
$result = await($future)
```

`async` crea el fork del runtime antes de iniciar la goroutine. El fork copia variables, mapas y listas de primer nivel y comparte definiciones de clases, funciones, conexión SQL y tablas de dispatch. No es aislamiento de proceso.

Una excepción en la tarea se guarda en el `Future` y se muestra como diagnóstico; `await()` devuelve el resultado almacenado.

## Canales

```joss
$channel = make_chan(1)
send($channel, "hola")
$value = recv($channel)
close($channel)
```

El operador `$channel << $value` también envía. Un canal sin buffer bloquea hasta que exista receptor; uno con tamaño positivo admite esa cantidad de elementos pendientes.

Las tareas `Cron::schedule($name, $schedule, { ... })` y `Task::on_request($name, $interval, { ... })` reciben bloques, no closures. El motor de Cron funciona dinámicamente en memoria aun cuando no haya conexión a base de datos disponible (sincronizándose automáticamente al conectar una BD) y evalúa cada minuto expresiones de cinco campos; soporta `*`, `*/n`, listas numéricas, valores exactos y los alias `hourly`, `daily`, `weekly` y `monthly`. `Task::on_request` ejecuta el bloque registrado en una goroutine background en cada petición.
