# Async, espera y comunicación entre tareas

Antes: [errores](ERRORES.md). Después: [proyecto de consola](PROYECTO_CONSOLA.md).
Referencia complementaria: [biblioteca](MODULOS_NATIVOS.md), [arquitectura](ARQUITECTURA.md).

## Tres conceptos diferentes

Una ejecución **síncrona** realiza una operación y espera a que termine antes
de continuar. Una operación **asíncrona** permite empezar un trabajo y recoger
su resultado después. **Concurrencia** significa que varias tareas pueden
avanzar durante el mismo período. No garantiza que todas se ejecuten al mismo
instante ni que el programa sea más rápido.

Un **Future** es el objeto que representa un resultado pendiente. Joss lo crea
con `async { ... }`. No es un tipo fuente genérico `Future<T>` que puedas
declarar. El contrato publicado de `async` y `await` es `mixed`.

## Lanzar y esperar

<!-- joss-run: ["Preparando resultado", "42"] -->
```joss
$futuro = async {
    return 20 + 22
}
print("Preparando resultado")
$resultado = await($futuro)
print($resultado)
```

`async` inicia una tarea en una goroutine de Go y devuelve su Future.
`await($futuro)` significa «espera aquí hasta que termine y dame su resultado».
Bloquea la tarea que espera; no es un event loop ni una suspensión cooperativa
como la de algunos otros lenguajes. Puede usarse en código de nivel superior
o dentro de una función; no exige declarar la función como `async func`.

Para obtener un resultado claro, escribe `return` en el bloque. Puedes esperar
un Future ya completado, incluso más de una vez. `await` sobre un valor que no
sea Future devuelve `null`; el analizador no prueba por completo esa condición.

`async(func() { ... })` fue eliminado y el parser lo rechaza. La forma
`async expresión` todavía se reconoce, pero evalúa el argumento antes de
entrar al built-in: no la uses para intentar diferir una llamada costosa.
Utiliza siempre el bloque.

## Trabajo independiente y errores

Lanza ambas tareas antes de esperar si son independientes:

<!-- joss-run: ["30"] -->
```joss
$uno = async { return 10 }
$dos = async { return 20 }
$a = await($uno)
$b = await($dos)
print($a + $b)
```

El orden en que terminan no está garantizado. Si imprimes desde dos tareas,
sus mensajes pueden aparecer en distinto orden. Esperar inmediatamente después
de lanzar cada una puede eliminar el solapamiento que buscabas.

Si el bloque lanza un error, el Future lo guarda. `await` vuelve a lanzarlo y
un `try/catch` alrededor de la espera puede recuperarlo. La implementación
también imprime `[ASYNC PANIC]` al fallar la tarea; puede perder la estructura
original al convertir el error a texto. No interpreta el fallo como éxito nulo.

No hay cancelación, timeout de Future, `select`, `await all` ni grupo estructurado
de tareas en la sintaxis fuente. Espera las tareas que necesitas antes de salir:
lanzarlas no garantiza que el proceso siga vivo hasta que terminen.

## Qué datos se comparten

Antes de iniciar la goroutine, el runtime hace un `Fork()`. Copia bindings,
tipos, constantes y varios contenedores; comparte registros y recursos como
conexiones de base de datos y canales. La closure que ejecuta el bloque conserva
su entorno capturado. No es un proceso aislado ni una copia profunda general
de todos los valores posibles.

No coordines tareas reasignando una variable escalar del exterior. Usa el
valor retornado o un canal. Los recursos compartidos y las colecciones anidadas
requieren revisar su contrato; las pruebas de carreras cubren casos concretos,
no una garantía universal.

## Canales: entregar valores

Un **canal** es un conducto por el que una tarea envía valores y otra los recibe.
`make_chan(n)` crea un canal con espacio para hasta `n` valores pendientes.
Sin argumento el espacio es cero: envío y recepción deben encontrarse.

<!-- joss-run: ["hola"] -->
```joss
$canal = make_chan(1)
send($canal, "hola")
print(recv($canal))
close($canal)
```

El buffer de uno permite enviar antes de recibir en la misma tarea. Con
`make_chan()`, ese orden bloquearía indefinidamente: nadie llegaría a recibir.

Un productor concurrente permite usar un canal sin buffer:

<!-- joss-run: ["10", "20"] -->
```joss
$canal = make_chan()
$productor = async {
    send($canal, 10)
    send($canal, 20)
    close($canal)
}
foreach ($canal as $valor) {
    print($valor)
}
await($productor)
```

`foreach` recibe hasta que el canal se cierra y se vacía. El productor cierra
porque sabe que ya no enviará más. Cerrar dos veces o enviar a un canal cerrado
produce un fallo. Un tamaño negativo tampoco es válido.

| Operación | Comportamiento |
|---|---|
| `make_chan([capacidad])` | Crea un `channel`; elementos dinámicos, sin contrato de tipo de elemento. |
| `send(canal, valor)` | Bloquea si no hay receptor/espacio; retorna `null`. |
| `canal << valor` | Envía y devuelve el propio canal. |
| `recv(canal)` | Bloquea hasta recibir; al cerrarse y vaciarse retorna `null`. |
| `close(canal)` | Cierra; no devuelve un resultado útil. |

`recv` no devuelve una bandera que distinga fin de canal de un mensaje
`null`. Diseña los mensajes teniendo en cuenta esa ambigüedad. Canales no son
colas persistentes: viven dentro del proceso.

## Cron y Task no son sinónimos de async

`Cron::schedule(nombre, expresión, { bloque })` registra un bloque para el
planificador en memoria. Comprueba cada minuto cinco campos, `*`, `*/n`,
listas, valores y aliases `hourly`, `daily`, `weekly`, `monthly`. Puede
registrar estado en SQL si hay conexión. No es una garantía de entrega durable
ni de ejecución única entre varias réplicas.

`Task::on_request(nombre, intervalo, { bloque })` tiene un nombre engañoso:
el handler actual ignora el intervalo y lanza una goroutine inmediatamente al
invocarse. No registra por sí mismo una tarea recurrente por petición.
Ambas APIs reciben bloques, no closures. Consulta el [informe](DOCUMENTATION_AUDIT.md).

Ejercicio: cambia el productor para enviar tres nombres; recibe e imprime cada
uno y explica por qué el `close` permite terminar el `foreach`.


[Índice](README.md)
