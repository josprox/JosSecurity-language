# WebSockets Nativos 🔌

JosSecurity soporta WebSockets de forma nativa, permitiendo comunicación bidireccional en tiempo real.

## Definición de Rutas (`routes.joss` / `api.joss`)

Usa el método `Router::ws` para definir un endpoint WebSocket.

```javascript
Router.ws("/api/chat-ws", "ChatController@handler")
```

> **Nota**: Este endpoint intercepta la petición HTTP y realiza el "Upgrade" a WebSocket automáticamente.

## Controladores

El manejador recibe una instancia nativa de `WebSocket` (`$ws`).

```javascript
class ChatController {
    func handler($ws) {
        // Evento: Al conectar (opcional, el código se ejecuta al conectar)
        $ws.send("¡Bienvenido!")

        // Evento: Al recibir mensaje
        $ws.onMessage(func($msg) {
            print("Mensaje recibido: " . $msg)
            
            // Responder
            $ws.send("Eco: " . $msg)
        })
    }
}
```

## Integración con IA

Puedes usar `streamTo` para canalizar la IA al socket:

```javascript
$ws.onMessage(func($msg) {
    AI::client()->user($msg)->streamTo($ws)
})
```

> **Importante**: `streamTo` usa un protocolo JSON específico (`type: chunk/start/done`). Revisa `docs/IA_NATIVA.md` para más detalles.

## Protocolo en el Cliente (Frontend)

Desde JavaScript en el navegador o Flutter:

```javascript
const socket = new WebSocket("ws://localhost:8000/api/chat-ws");

socket.onopen = () => {
    socket.send(JSON.stringify({content: "Hola"}));
};

socket.onmessage = (event) => {
    console.log("Recibido:", event.data);
};
```

## Despliegue en Producción (Nginx/Apache)

Si usas un proxy reverso como Nginx (por ejemplo con HestiaCP), es **CRÍTICO** asegurar que las cabeceras `Upgrade` y `Connection` pasen correctamente.

### Nginx ("Missing Upgrade Header")

Si recibes errores de handshake, verifica que tu configuración de Nginx NO tenga:

```nginx
proxy_hide_header Upgrade; # ELIMINAR ESTA LÍNEA
```

Y asegúrate de incluir:

```nginx
proxy_set_header Upgrade $http_upgrade;
proxy_set_header Connection "Upgrade";
```

Esto es común en plantillas por defecto de paneles de control.
