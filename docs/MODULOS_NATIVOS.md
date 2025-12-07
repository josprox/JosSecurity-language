# Módulos Nativos de JosSecurity

Documentación completa de todos los módulos nativos disponibles en JosSecurity.

## Índice
- [Auth](#auth) - Autenticación y autorización
- [GranMySQL](#granmysql) - Base de datos
- [Router](#router) - Sistema de rutas
- [View](#view) - Motor de plantillas
- [SmtpClient](#smtpclient) - Correo electrónico
- [Response](#response) - Respuestas HTTP
- [Request](#request) - Peticiones HTTP
- [Cron](#cron) - Tareas programadas
- [Task](#task) - Tareas hit-based
- [Schema](#schema) - Esquemas de base de datos
- [System](#system) - Utilidades del sistema
- [Redis](#redis) - Cache y sesiones
- [Queue](#queue) - Colas de trabajo
- [WebSocket](#websocket) - Comunicación en tiempo real
- [Math](#math) - Funciones matemáticas
- [Session](#session) - Gestión de sesiones


---

## Auth

Módulo de autenticación con Bcrypt y JWT.

### Métodos

#### `Auth::create(array $datos)`
Registra un nuevo usuario con hash automático de contraseña.

```joss
// Registro simple
Auth::create(["user@example.com", "password123", "Juan Pérez"])

// Con rol personalizado
Auth::create(["admin@example.com", "admin123", "Admin", 1])
```

**Parámetros**:
- `[0]` email (string)
- `[1]` password (string) - Se hashea automáticamente con Bcrypt
- `[2]` name (string)
- `[3]` role_id (int, opcional) - Default: 2 (Client)

#### `Auth::attempt(string $email, string $password)`
Intenta autenticar un usuario.

```joss
$success = Auth::attempt("user@example.com", "password123")

($success) ? {
    print("Login exitoso")
} : {
    print("Credenciales inválidas")
}
```

**Retorna**: `string|bool` - Token JWT si es exitoso, `false` si falla.

#### `Auth::check()`
Verifica si hay un usuario autenticado.

```joss
($Auth::check()) ? {
    print("Usuario autenticado")
} : {
    print("No autenticado")
}
```

#### `Auth::guest()`
Verifica si el usuario es invitado (no autenticado).

```joss
($Auth::guest()) ? {
    print("Bienvenido, invitado")
} : {
    print("Ya estás autenticado")
}
```

#### `Auth::user()`
Obtiene el nombre del usuario autenticado.

```joss
$nombre = Auth::user()
print("Hola, " . $nombre)
```

#### `Auth::id()`
Obtiene el ID del usuario autenticado.

```joss
$userId = Auth::id()
```

#### `Auth::hasRole(string $rol)`
Verifica si el usuario tiene un rol específico.

```joss
($Auth::hasRole("admin")) ? {
    print("Acceso de administrador")
} : {
    print("Acceso denegado")
}
```

#### `Auth::logout()`
Cierra la sesión del usuario.

```joss
Auth::logout()
print("Sesión cerrada")
```

#### `Auth::verify(string $token)`
Verifica una cuenta de usuario mediante su token.
Retorna `true` si la verificación fue exitosa.

```joss
$verificado = Auth::verify($token)
```

#### `Auth::refresh(int $userId)`
Genera un nuevo token JWT para el usuario especificado.

```joss
$newToken = Auth::refresh(Auth::id())
```

#### `Auth::delete(int $userId)`
Elimina un usuario y sus datos de la base de datos.
Retorna `true` si la eliminación fue exitosa.

```joss
Auth::delete(Auth::id())
```

#### `Auth::validateToken(string $token)`
Valida un token Bearer y establece la sesión del usuario si es válido.

```joss
$valid = Auth::validateToken("Bearer eyJhb...")
```

### Base de Datos (Automática)
El módulo Auth gestiona automáticamente una tabla `users` (con prefijo opcional `js_`) con **17 columnas** optimizadas, incluyendo:
- `user_token` (UUID)
- `verificado` (Control de email)
- `role_id` (RBAC)
- Timestamps: `created_at`, `updated_at`, `last_login_at`, `last_refresh_at`, `last_logout_at`.

El sistema incluye "Self-Healing": si faltan columnas, se agregan automáticamente sin perder datos.
Las tablas se sincronizan automáticamente con las correcciones del motor (e.g. adición de `last_login_at`, `verificado`, etc).

---

## GranMySQL

ORM nativo con protección contra SQL injection.

### API Fluida

```joss
$db = new GranMySQL()

// Seleccionar tabla
$db->table("users")

// Condiciones
$db->where("edad", ">", 18)
$db->where("activo", 1)

// Obtener resultados
$usuarios = $db->get()  // JSON string
```

### Métodos

#### `table(string $nombre)`
Selecciona la tabla (agrega prefijo automáticamente).

```joss
$db->table("users")  // Usa js_users
```

#### `select(string|array $columnas)`
Especifica columnas a seleccionar.

```joss
$db->select("nombre, email")
$db->select(["nombre", "email"])
```

#### `where(string $columna, mixed $valor)`
#### `where(string $columna, string $operador, mixed $valor)`
Agrega condición WHERE.

```joss
$db->where("id", 1)
$db->where("edad", ">", 18)
$db->where("nombre", "LIKE", "%Juan%")
```

#### `get()`
Ejecuta la consulta y retorna resultados como JSON.

```joss
$resultados = $db->table("users")->get()
```

#### `first()`
Obtiene el primer resultado.

```joss
$usuario = $db->table("users")->where("email", "user@example.com")->first()
```

#### `insert(array $columnas, array $valores)`
Inserta un nuevo registro.

```joss
$db->table("users")->insert(
    ["nombre", "email"],
    ["Juan", "juan@example.com"]
)
```

#### `innerJoin(string $tabla, string $col1, string $op, string $col2)`
Realiza un INNER JOIN.

```joss
$db->table("users")
   ->innerJoin("roles", "users.role_id", "=", "roles.id")
   ->get()
```

#### `leftJoin()`, `rightJoin()`
Joins izquierdo y derecho.

```joss
$db->table("posts")
   ->leftJoin("users", "posts.user_id", "=", "users.id")
   ->get()
```

### API Legacy

```joss
$consulta = new GranMySQL()
$consulta->tabla = "users"
$consulta->comparar = "email"
$consulta->comparable = "user@example.com"
$resultado = $consulta->where("json")
```

---

## Router

Sistema de rutas con middleware.

### Métodos

#### `Router::get(string $path, string $handler)`
Ruta GET.

```joss
Router::get("/", "HomeController@index")
Router::get("/about", "PageController@about")
```

#### `Router::post(string $path, string $handler)`
Ruta POST.

```joss
Router::post("/login", "AuthController@login")
Router::post("/register", "AuthController@register")
```

#### `Router::put()`, `Router::delete()`
Rutas PUT y DELETE.

```joss
Router::put("/users/:id", "UserController@update")
Router::delete("/users/:id", "UserController@delete")
```

#### `Router::match(string $methods, string $path, string $handler)`
Múltiples métodos HTTP.

```joss
// Mismo handler para GET y POST
Router::match("GET|POST", "/contact", "ContactController@handle")

// Handlers diferentes por método
Router::match("GET|POST", "/form", "FormController@show@submit")
```

#### `Router::api(string $path, string $handler)`
Ruta API (sin CSRF, retorno JSON).

```joss
Router::api("/users", "ApiController@getUsers")
```

#### `Router::group(string $prefix)`
Agrupa rutas bajo un prefijo común.

```joss
Router::group("/admin")
Router::get("/dashboard", "AdminController@dashboard") // /admin/dashboard
Router::end()
```

#### `Router::middleware(string $nombre)`
Inicia grupo de middleware.

```joss
Router::middleware("auth")
Router::get("/dashboard", "DashboardController@index")
Router::get("/profile", "ProfileController@show")
Router::end()
```

#### `Router::end()`
Finaliza grupo de middleware.

**Middleware disponibles**:
- `auth` - Requiere autenticación
- `guest` - Solo invitados

---

## View

Motor de plantillas HTML con herencia.

### Métodos

#### `View::render(string $nombre, map $datos)`
Renderiza una vista.

```joss
return View::render("welcome", {"nombre": "Juan"})
```

### Sintaxis de Plantillas

#### Variables
```html
<!-- Escapado (seguro) -->
<h1>Hola {{nombre}}</h1>
<p>{{ $email }}</p>

<!-- Sin escapar (raw) -->
<div>{{! contenido_html }}</div>
```

#### Condicionales
```html
<!-- Ternario -->
<p>{{ $activo ? 'Activo' : 'Inactivo' }}</p>

<!-- Null coalescing -->
<p>{{ $nombre ?? "Anónimo" }}</p>

<!-- Expresiones Complejas -->
<div class="{{ ($error) ? 'alert-danger' : 'alert-success' }}">
    {{ $mensaje }}
</div>

<!-- Lógica y Matemáticas -->
<p>Total: {{ $precio * $cantidad }}</p>
<p>Estado: {{ ($activo && !$banned) ? "OK" : "Bloqueado" }}</p>
```

#### Herencia
```html
<!-- layout.joss.html -->
<!DOCTYPE html>
<html>
<head>
    <title>@yield('title')</title>
</head>
<body>
    @yield('content')
</body>
</html>

<!-- page.joss.html -->
@extends('layouts.layout')

@section('title')
Mi Página
@endsection

@section('content')
<h1>Contenido</h1>
@endsection
```

#### Helpers
```html
<!-- CSRF Token -->
<form method="POST">
    {{ csrf_field() }}
    <!-- Genera: <input type="hidden" name="_token" value="..."> -->
</form>
```

#### Variables de Auth
```html
<!-- Disponibles automáticamente -->
{{ auth_check }}  <!-- true/false -->
{{ auth_guest }}  <!-- true/false -->
{{ auth_user }}   <!-- Nombre del usuario -->
{{ auth_role }}   <!-- Rol del usuario -->
```

---

## SmtpClient

Cliente de correo con SSL/TLS.

### Uso

```joss
$mail = new SmtpClient()
$mail->auth(env("MAIL_USER"), env("MAIL_PASS"))
$mail->secure(true)  // SSL/TLS
$mail->send("destino@example.com", "Asunto", "Cuerpo del mensaje")
```

### Métodos

#### `auth(string $user, string $pass)`
Configura autenticación.

#### `secure(bool $enabled)`
Habilita SSL/TLS.

#### `send(string $to, string $subject, string $body)`
Envía correo.

**Configuración en env.joss**:
```bash
MAIL_HOST="smtp.gmail.com"
MAIL_PORT="587"
MAIL_USER="tu_email@gmail.com"
MAIL_PASS="tu_password"
```

---

## Response

Manejo de respuestas HTTP.

### Métodos

#### `Response::json(mixed $data, int $status)`
Respuesta JSON.

```joss
return Response::json({"mensaje": "OK"}, 200)
```

#### `Response::redirect(string $url)`
Redirección.

```joss
return Response::redirect("/dashboard")
```

#### `Response::error(string $mensaje, int $code)`
Respuesta de error.

```joss
return Response::error("No autorizado", 401)
```

---

## Request

Acceso a datos de la petición HTTP.

### Métodos

#### `Request::get(string $key)`
Obtiene parámetro GET.

```joss
$id = Request::get("id")
```

#### `Request::post(string $key)`
Obtiene parámetro POST.

```joss
$email = Request::post("email")
// Alias de Request::input()
```

#### `Request::input(string $key)`
Obtiene un parámetro de la petición (GET o POST).

```joss
$nombre = Request::input("nombre")
```

#### `Request::except(array $keys)`
Obtiene todos los parámetros excepto los especificados.

```joss
$datos = Request::except(["_token", "password"])
```

#### `Request::all()`
Obtiene todos los parámetros.

```joss
$datos = Request::all()
```

---

## Cron

Tareas programadas tipo demonio.

### Uso

```joss
// config/cron.joss
Cron::schedule("backup_diario", "00:00", {
    System::backupDatabase()
})

Cron::schedule("limpieza", "03:00", {
    DB::table("logs")->where("created_at", "<", "30 days ago")->delete()
})
```

---

## Task

Tareas basadas en hits (tráfico web).

### Uso

```joss
// main.joss
Task::on_request("limpiar_tokens", "1 hour", {
    Auth::cleanExpiredTokens()
})
```

---

## Schema

Creación de esquemas de base de datos.

### Métodos

#### `create(string $tabla, map $columnas)`
Crea una tabla.

```joss
$schema = new Schema()
$schema->create("posts", {
    "id": "INT AUTO_INCREMENT PRIMARY KEY",
    "title": "VARCHAR(255)",
    "content": "TEXT",
    "user_id": "INT",
    "created_at": "TIMESTAMP DEFAULT CURRENT_TIMESTAMP"
})
```

---

## System

Utilidades del sistema.

### Métodos

> [!CAUTION]
> **Riesgo de Seguridad**: El método `System::Run` permite la ejecución de comandos del sistema operativo. Por defecto está **bloqueado**. Para habilitarlo, debe configurar `ALLOW_SYSTEM_RUN=true` en su archivo de entorno. Úselo con extrema precaución.

#### `System::env(string $key, mixed $default)`
Lee variable de entorno.

```joss
$debug = System::env("DEBUG", false)
```

---

## Redis

Cache y sesiones con Redis.

### Configuración

```bash
# env.joss
SESSION_DRIVER="redis"
REDIS_HOST="localhost:6379"
REDIS_PASSWORD=""
```

---

## Queue

Sistema de colas de trabajo.

```joss
Queue::push("enviar_email", {"to": "user@example.com"})
```

---

## WebSocket

Comunicación en tiempo real.

WebSocket::broadcast("mensaje", {"texto": "Hola a todos"})
```

---

## Math

Funciones matemáticas de utilidad.

### Métodos

#### `Math::random(int $min, int $max)`
Genera un número entero aleatorio entre min y max.

```joss
$dado = Math::random(1, 6)
```

#### `Math::floor(float $val)`
Redondea hacia abajo.

```joss
$entero = Math::floor(4.9) // 4
```

#### `Math::ceil(float $val)`
Redondea hacia arriba.

```joss
$entero = Math::ceil(4.1) // 5
```

#### `Math::abs(number $val)`
Valor absoluto.

```joss
$positivo = Math::abs(-10) // 10
```

---

## Session

Gestión directa de la sesión del usuario.

### Métodos

#### `Session::put(string $key, mixed $value)`
Guarda un valor en la sesión.

```joss
Session::put("carrito_id", 123)
```

#### `Session::get(string $key)`
Obtiene un valor de la sesión.

```joss
$id = Session::get("carrito_id")
```

#### `Session::has(string $key)`
Verifica si existe una clave.

```joss
(Session::has("user_id")) ? {
    // ...
} : {
    // ...
}
```

#### `Session::forget(string $key)`
Elimina una clave de la sesión.

```joss
Session::forget("temp_data")
```

#### `Session::all()`
Obtiene todos los datos de la sesión.

```joss
$datos = Session::all()
```

