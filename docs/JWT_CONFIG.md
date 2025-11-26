# JWT Configuration - JosSecurity

## Variables de Entorno

Configura la expiración de los tokens JWT en tu archivo `env.joss`:

```joss
// JWT Secret (IMPORTANTE: Cambiar en producción)
JWT_SECRET=tu_secreto_super_seguro_aqui

// Expiración del token inicial (en meses)
// Default: 3 meses
JWT_INITIAL_EXPIRY_MONTHS=3

// Expiración del token refrescado (en meses)
// Default: 6 meses
JWT_REFRESH_EXPIRY_MONTHS=6
```

## Uso en Joss

### Login Inicial (3 meses)

```joss
$auth = new Auth()
$token = $auth.attempt("user@example.com", "password123")
// Retorna JWT válido por 3 meses
```

### Refrescar Token (6 meses)

```joss
$auth = new Auth()
$newToken = $auth.refresh($oldToken)
// Retorna nuevo JWT válido por 6 meses
```

## Estructura del JWT

El token incluye los siguientes claims:

```json
{
  "user_id": 1,
  "email": "user@example.com",
  "name": "User Name",
  "exp": 1740326878,  // Timestamp de expiración
  "iat": 1732590878   // Timestamp de emisión
}
```

## Seguridad

- **Algoritmo**: HS256 (HMAC with SHA-256)
- **Secret Key**: Configurable via `JWT_SECRET` env var
- **Validación**: El método `refresh()` valida el token antes de generar uno nuevo
- **Expiración**: Tokens expiran automáticamente según configuración

## Ejemplo Completo

```joss
class Main {
    Init main() {
        $auth = new Auth()
        
        // 1. Login
        $token = $auth.attempt("user@example.com", "pass123")
        print("Token inicial (3 meses):")
        print($token)
        
        // 2. Usar el token...
        // (Aquí iría la lógica de tu aplicación)
        
        // 3. Refrescar antes de que expire
        $newToken = $auth.refresh($token)
        print("Token refrescado (6 meses):")
        print($newToken)
    }
}
```

## Notas Importantes

1. **Cambiar JWT_SECRET en producción**: El secret por defecto es solo para desarrollo
2. **Almacenar tokens de forma segura**: No expongas los tokens en logs o URLs
3. **Validar tokens en cada request**: Usa el método `refresh()` para validar y renovar
4. **Monitorear expiración**: Implementa lógica para refrescar tokens antes de que expiren
