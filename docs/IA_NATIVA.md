# JosSecurity AI Nativa 🧠

JosSecurity incluye ahora soporte nativo para Inteligencia Artificial ("Spring AI Style"), permitiendo interactuar con LLMs (Groq, OpenAI, Gemini) directamente desde el lenguaje sin dependencias externas.

## Configuración (.env)

Debes configurar tus claves de API en el archivo `.env`:

```ini
GROQ_API_KEY="gsk_..."
OPENAI_API_KEY="sk-..."
GEMINI_API_KEY="AIzr..."

# Proveedor por defecto
AI_PROVIDER="groq" 
# Modelo por defecto
AI_MODEL="llama3-70b-8192"
```

## API Fluida (Spring AI Style)

La clase `AI` proporciona un constructor fluido `client()` para facilitar la creación de chats.

### Ejemplo Básico

```javascript
$response = AI::client()
    ->system("Eres un experto en seguridad.")
    ->user("¿Qué es XSS?")
    ->call()

print($response)
```

### Ejemplo con Historial

```javascript
$client = AI::client()
    ->system("Eres un asistente útil.")
    ->user("Hola, mi nombre es Joss.")
    ->assistant("Hola Joss, ¿en qué puedo ayudarte?")
    ->user("¿Cuál es mi nombre?")

$respuesta = $client->call()
```

### Streaming (Server-Sent Events)

Para recibir la respuesta trozo a trozo (token by token):

```javascript
AI::client()
    ->user("Genera un cuento largo...")
    ->stream(func($chunk) {
        print($chunk) // Se ejecuta por cada token recibido
    })
```

### Streaming a WebSockets ⚡

Esta es la característica más potente. Conecta la salida de la IA directamente a un WebSocket cliente.

```javascript
func ws($ws) {
    $ws.onMessage(func($msg) {
        $client = AI::client()->user($msg)
        
        // La IA escribe directamente en el socket del cliente
        // Devuelve el texto completo al finalizar para guardarlo en BD si quieres
        $fullText = $client->streamTo($ws) 
    })
}
```

#### Protocolo de Streaming (Frontend)

Cuando usas `streamTo($ws)`, el cliente recibirá mensajes JSON con la siguiente estructura:

1. **Inicio**: `{ "type": "start" }`
2. **Contenido**: `{ "type": "chunk", "content": "Hola..." }`
3. **Fin**: `{ "type": "done" }`
4. **Error**: `{ "type": "error", "content": "..." }`

Asegúrate de que tu cliente (JS/Flutter) parsee estos eventos.

## Métodos Disponibles

| Método | Descripción |
|--------|-------------|
| `system($msg)` | Define el mensaje del sistema (instrucciones). |
| `user($msg)` | Añade un mensaje de usuario. |
| `assistant($msg)` | Añade una respuesta del asistente (para contexto/historial). |
| `call()` | Ejecuta la petición de forma síncrona y devuelve el texto. |
| `stream($cb)` | Ejecuta en streaming y llama a `$cb($chunk)` por cada trozo. |
| `streamTo($ws)` | Envía los trozos directamente a una instancia `WebSocket`. |
