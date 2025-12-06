# Desarrollo de APIs REST con JosSecurity

JosSecurity v3.0 introduce soporte nativo robusto para la creación de APIs RESTful, incluyendo autenticación JWT y manejo de respuestas JSON.

## Configuración Básica

Para comenzar una API, utiliza el helper `Router::api()` en tu archivo `api.joss` (o `routes.joss`). Esto configura automáticamente los headers necesarios (Content-Type: application/json).

```joss
// routes.go o api.joss
Router::api()

Router::get("/api/status", "ApiController@status")
```

## Respuestas JSON

La clase `Response` ofrece el método `json` para devolver datos estructurados:

```joss
class ApiController {
    func status() {
        return Response::json({
            "status": "online",
            "uptime": System::uptime()
        }, 200) // Código HTTP opcional (default 200)
    }
}
```

## Autenticación (JWT)

JosSecurity utiliza JSON Web Tokens (JWT) para asegurar tus endpoints.

### 1. Login y Obtención de Token

Utiliza `Auth::attempt` para validar credenciales y obtener un token.

```joss
func login() {
    $email = Request::input("email")
    $password = Request::input("password")
    
    // Auth::attempt retorna el token string si es exitoso, o false
    var $token = Auth::attempt($email, $password)
    
    ($token) ? {
        return Response::json({"token": $token})
    } : {
        return Response::json({"error": "Credenciales inválidas"}, 401)
    }
}
```

### 2. Protección de Rutas

Utiliza el middleware `auth_api` para proteger tus rutas. Este middleware verifica el header `Authorization: Bearer <token>`.

```joss
Router::middleware("auth_api")
    Router::get("/api/user", "ApiController@user")
    Router::post("/api/refresh", "ApiController@refresh")
Router::end()
```

### 3. Acceso al Usuario Autenticado

Dentro de una ruta protegida, puedes acceder a los datos del usuario:

```joss
func user() {
    $user = Auth::user() // Retorna map con datos del usuario
    return Response::json($user)
}
```

## Operaciones Comunes

### Refrescar Token
Extiende la sesión del usuario obteniendo un nuevo token.

```joss
func refresh() {
    $newToken = Auth::refresh(Auth::id())
    return Response::json({"token": $newToken})
}
```

### Eliminar Cuenta
Elimina el usuario actual.

```joss
func delete() {
    Auth::delete(Auth::id())
    return Response::json({"message": "Cuenta eliminada"})
}
```

## Middleware Personalizado

Puedes definir lógica intermedia personalizada para tus APIs:

```joss
// En tu archivo de rutas o bootstrap
Router::registerMiddleware("check_balance", function() {
    $user = Auth::user()
    if ($user["balance"] < 0) {
        return Response::json({"error": "Saldo insuficiente"}, 402)
    }
})

// Uso
Router::middleware("auth_api")
    Router::middleware("check_balance")
        Router::post("/api/purchase", "ShopController@buy")
    Router::end()
Router::end()
```
