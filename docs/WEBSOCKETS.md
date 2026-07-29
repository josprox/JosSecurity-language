# WebSockets

Las rutas aceptan parámetros dinámicos. El primer argumento del handler es la conexión y después se inyectan los parámetros en orden.

```joss
Router::ws("/rooms/{room}/users/{id}", "ChatController@connect")

class ChatController {
    func connect($ws, string $room, string $id) {
        $ws->onMessage(func($message) {
            $ws->send($message)
        })
    }
}
```

Los callbacks registrados con `onMessage` capturan el entorno léxico del
handler. `$ws`, los parámetros de ruta y las variables locales siguen
disponibles después de que el handler termina. El estado capturado se conserva
entre mensajes y sus invocaciones se serializan:

```joss
func connect($ws) {
    $count = 0
    $ws->onMessage(func($message) {
        $count = $count + 1
        $ws->send("Mensaje " . $count . ": " . $message)
    })
}
```

Los callbacks registrados durante la misma ejecución comparten su entorno. Por
ejemplo, un `onClose` puede leer el estado que modificó `onMessage`.

Cada conexión se ejecuta sobre su propio `Runtime.Fork()`, por lo que el
contexto capturado no se comparte con otras conexiones.

La conexión expone:

- `send($message)`: envía un mensaje y retorna si la escritura tuvo éxito.
- `onMessage($callback)`: registra el callback de mensajes.
- `onClose($callback)`: registra limpieza que siempre se ejecuta al terminar.
- `subscribe($channel)` / `unsubscribe($channel)`: administra canales locales.
- `publish($channel, $message)`: publica a los demás miembros del canal.
- `WebSocket::subscriberCount($channel)`: devuelve el número de conexiones
  activas suscritas al canal.
- `close()`: cierra realmente el socket.

`WebSocket::publish($channel, $message)` publica a todos los miembros y puede
invocarse desde handlers HTTP. Las suscripciones se eliminan automáticamente
al cerrar la conexión. `WebSocket::broadcast()` conserva el hub global
histórico.

Los canales son locales al proceso. Un despliegue con varias réplicas necesita
sticky sessions y un backplane pub/sub externo, o debe mantener una sola
réplica para funciones de sala.

El servidor aplica keepalive y límites configurables:

```env
WS_MAX_MESSAGE_BYTES="8388608"
WS_IDLE_TIMEOUT_SECONDS="120"
WS_PING_INTERVAL_SECONDS="30"
```

El upgrade ocurre antes del middleware HTTP normal. Para autenticación dentro del socket, valida el JWT con `Auth::validateToken($token)`; el runtime repuebla la sesión usada por `Auth::user()`.

Con `TLS_CERT_FILE` y `TLS_KEY_FILE`, el servidor integrado ofrece `wss`. Un proxy inverso sigue siendo válido y debe reenviar `Upgrade` y `Connection`.
