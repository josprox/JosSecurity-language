# Contexto para Agentes de IA

Este documento sirve como memoria persistente para futuros agentes que trabajen en este proyecto.

## Lecciones Aprendidas (Sesión 21/12/2025)

### 1. Intérprete JosSecurity - Comportamientos Clave
- **Retornos Estrictos**: El intérprete detiene la ejecución de un bloque inmediatamente al encontrar un `ReturnStatement`.
- **JSON Parsing**: `JSON::parse()` requiere estrictamente un `string`. Si pasas un objeto (como una lista de BD), retornará `nil` o fallará.
- **Base de Datos**: `GranMySQL::get()` retorna un `[]map[string]interface{}` (Lista Nativa), NO un string JSON. No es necesario parsearlo.
- **Foreach**: Soporta iteración sobre `[]interface{}` y `[]map[string]interface{}`.
- **Prefix Expressions**: Operadores como `!` y `-` funcionan correctamente en el evaluador (`evaluatePrefix`).

### 2. Manejo de Archivos y Descargas
- **Uploads**: Los archivos subidos se encuentran en `$file["content"]`, no en `tmp_name`. El servidor lee el contenido en memoria.
- **Downloads**: Para descargar archivos binarios sin corrupción:
  1. Usar `Response::raw($content, $status, $mime, $headers)`.
  2. Forzar headers: `Content-Disposition: attachment; filename="..."`.
  3. **IMPORTANTE**: No retornar strings simples para binarios, ya que el servidor podría inyectar scripts (Hot Reload) y corromper el archivo.

### 3. Estructura de Controladores
- **Sintaxis**: JosSecurity no usa `if/else`, usa ternarios `cond ? { ... } : { ... }`.
- **Métodos**: Asegurarse de que cada método esté correctamente cerrado con `}`. Un error de anidamiento puede hacer que el Dispatcher no encuentre el método.

### 4. Estilo de Código
- Usar bloques `{ ... }` explícitos dentro de los ternarios para flujos complejos.
- Para concatenar strings usar `.`.

### 5. Autenticación y Sesiones (JWT Update)
- **Stateless**: La autenticación ya no depende de `storage/sessions.json`.
- **JWT Cookie**: El login exitoso setea una cookie `joss_token` (HTTP-Only).
- **Validación**: El servidor (`handler.go`) valida el JWT en cada petición y restaura la sesión (`user_id`, `email`, etc.) desde los claims del token.
- **API**: El endpoint `/api/login` retorna el JWT en el JSON para clientes externos.
- **Uso**: Usar `Response::redirect(...)->withCookie("joss_token", $token)` en el login.
- **Gotcha: Roles**: El Token JWT DEBE incluir el rol del usuario (claim `role`). Si no, al restaurar sesión tras un reinicio, se pierden los permisos de admin.
- **Gotcha: Logout**: `Auth::logout()` solo limpia memoria. Para invalidar realmente la sesión, SE DEBE setear la cookie con valor vacío: `withCookie("joss_token", "")`. El servidor procesará esto (`handler.go`) seteando `MaxAge: -1`.

### 6. Integración Flutter & Backups (Sesión 27/12/2025)
- **API Standard**: Flutter debe usar siempre el prefijo `/api/` (ej: `/api/listfiles`) y autenticación `Authorization: Bearer <token>`. Headers viejos como `X-JossRed-Auth` son obsoletos.
- **Backups**:
  - `listfiles` retorna los paths completos.
  - Para descargas, el path puede ser de 2 partes (`appName/file`) o 3 partes. El cliente debe manejar ambos casos.
  - **Borrado**: `UserStorage::delete($token, $path)` funciona correctamente. Se implementó `DELETE /api/backup/{id}`.
- **Flutter UI**:
  - Migración de widgets legacy a componentes modernos y aislados (ej: `JossChips`).
### 7. IA Nativa, WebSockets y CLI (Sesión 28/12/2025)
- **IA Nativa**:
  - Implementada abstracción fluida `AI::client()->user(...)->call()`.
  - Soporte de Streaming Token-by-Token (`streamTo($ws)`).
  - Documentación en `docs/IA_NATIVA.md`.
- **WebSockets**:
  - Implementado `Router::ws("/path", "Controller@method")`.
  - Manejo de conexiones crudas mediante actualización en `MainHandler`. **Critico**: Los WebSockets actualmente se ejecutan *antes* del middleware de sesión estándar en `handler.go`, por lo que `Auth::user()` puede no estar disponible automáticamente. Se recomienda enviar el token en el primer mensaje o headers y validarlo manualmente si es crítico.
  - Documentación en `docs/WEBSOCKETS.md`.
- **Flutter Integration**:
  - Usar `web_socket_channel` para chat en tiempo real.
  - El protocolo actual usa JSON events: `{type: "chunk", content: "..."}`.
- **CLI**:
  - Nuevos comandos se registran en `cmd/joss/main.go`.
  - Implementado `joss ai:activate` con prompts interactivos (`bufio`).
  - **Gotcha Environment**: El Runtime de Joss carga `env.joss` en memoria (`r.Env`). Los módulos nativos deben preferir `r.Env["KEY"]` antes que `os.Getenv("KEY")`, ya que `joss server start` no siempre exporta las variables al entorno del SO.
