# Actualizaciones de Seguridad y Sistema

Este documento detalla las nuevas funcionalidades implementadas en JosSecurity.

## 1. Mejoras en Autenticación (Auth)

Se han añadido métodos para manejo de contraseñas y verificación obligatoria.

### Recuperación de Contraseña

#### `Auth::forgotPassword(email)`
Genera un token de recuperación y lo almacena en la base de datos.
- **Retorno**: `string` (token) si es exitoso, `false` si falla.
- **Uso**: El controlador utiliza `SmtpClient` para enviar el token por email.

> [!IMPORTANT]
> **Configuración SMTP Requerida**: Para que el envío de correos funcione, debe configurar las variables `MAIL_HOST`, `MAIL_PORT`, `MAIL_USERNAME` y `MAIL_PASSWORD` en su archivo `env.joss`.

> [!TIP]
> **Compatibilidad de Base de Datos**: El sistema de autenticación ahora utiliza `GranDB` internamente y formato de fechas estándar (`YYYY-MM-DD HH:MM:SS`), asegurando compatibilidad total con SQLite y MySQL.

```joss
$token = Auth::forgotPassword("user@example.com")
if ($token) {
    // El controlador se encarga de enviar el email
}
```

#### `Auth::resetPassword(token, newPassword)`
Restablece la contraseña si el token es válido y no ha expirado (1 hora).
- **Retorno**: `true` (éxito), `"invalid_token"`, `"expired_token"`, `"used_token"` o `false`.

```joss
$status = Auth::resetPassword($token, "NuevaPassword123")
```

### Verificación de Cuenta

Ahora `Auth::attempt` fallará si el usuario no está verificado (`verificado = 0`).

#### `Auth::resendVerification(email)`
Genera un nuevo token de verificación si el anterior expiró o se perdió.
- **Retorno**: `string` (nuevo token) o `false`.

```joss
$newToken = Auth::resendVerification("user@example.com")
```

---

## 2. Ejecución de Scripts Externos

Se ha habilitado la palabra clave global `run` para ejecutar scripts de Python y PHP.

> [!WARNING]
> Requiere `ALLOW_SYSTEM_RUN="true"` en `env.joss`.

#### Sintaxis
```joss
// Ejecutar Python
$salida = run "scripts/calculo.py" param1 param2

// Ejecutar PHP
$html = run "legacy/render.php"
```

---

## 3. Servicios Internos (Microservicios)

JosSecurity puede levantar servicios aislados en puertos específicos.

#### `Server::spawn(name, command, port)`
Ejecuta un comando en segundo plano, inyectando la variable de entorno `PORT`.

```joss
// Levantar una API en Python (Flask/FastAPI) en puerto 5000
Server::spawn("py-api", "python api.py", 5000)

// El script python debe leer os.environ.get("PORT")
```
