# JosSecurity (Joss)

Lenguaje de programación moderno con enfoque en seguridad, inspirado en PHP, Python, Java y Go.

## Características Principales

### 🚀 Sistema de Tipos Robusto
- **Smart Numerics**: Promoción automática de int a float (división siempre retorna float)
- **Maps Nativos**: Sintaxis `{ key: value }` con soporte completo
- **Tipos Dinámicos**: Sistema flexible con optimización automática
- **Operadores Ternarios**: Reemplazo de if/else con sintaxis concisa

### ⚡ Concurrencia
- **async/await**: Ejecución asíncrona aprovechando Goroutines de Go
- **Futures**: Manejo de valores asíncronos con canales de Go

### 🔐 Seguridad Integrada
- **Auth Module**: Autenticación con JWT y Bcrypt (12 rondas)
- **GranMySQL**: ORM seguro con protección contra SQL injection
- **Entorno Encriptado**: Variables de entorno con AES-256
- **CSRF Protection**: Protección nativa contra ataques CSRF
- **Rate Limiting**: Limitación de peticiones por IP

### 📦 Módulos Nativos
- **Router**: Sistema de rutas con middleware
- **View**: Motor de plantillas HTML
- **SMTP**: Cliente de correo con SSL/TLS
- **Cron**: Tareas programadas
- **Redis**: Cache y sesiones
- **WebSocket**: Comunicación en tiempo real

### 🎨 Desarrollo Web
- **Hot Reload**: Recarga automática en desarrollo
- **SCSS Compilation**: Compilación automática de estilos
- **Static Files**: Servidor de archivos estáticos
- **Security Headers**: Cabeceras de seguridad automáticas

## Instalación

### 🚀 Instalación Rápida (One-Liner)

**Windows (PowerShell como Administrador)**:
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope Process; iwr -useb https://raw.githubusercontent.com/josprox/JosSecurity-language/main/install/remote-install.ps1 | iex
```

**Linux/macOS**:
```bash
curl -fsSL https://raw.githubusercontent.com/josprox/JosSecurity-language/main/install/remote-install.sh | bash
```

### 📦 Instalación Manual

```bash
# Clonar repositorio
git clone https://github.com/josprox/JosSecurity-language.git
cd JosSecurity-language/install

# Ejecutar instalador
./install.sh  # Linux/macOS
# o
.\install.ps1  # Windows
```

### 🔄 Actualizar

**Windows**:
```powershell
iwr -useb https://raw.githubusercontent.com/josprox/JosSecurity-language/main/install/update.ps1 | iex
```

**Linux/macOS**:
```bash
curl -fsSL https://raw.githubusercontent.com/josprox/JosSecurity-language/main/install/update.sh | bash
```

### 🗑️ Desinstalar

**Windows**:
```powershell
iwr -useb https://raw.githubusercontent.com/josprox/JosSecurity-language/main/install/uninstall.ps1 | iex
```

**Linux/macOS**:
```bash
curl -fsSL https://raw.githubusercontent.com/josprox/JosSecurity-language/main/install/uninstall.sh | bash
```

**Ver más comandos**: [ONE_LINER_COMMANDS.md](ONE_LINER_COMMANDS.md)

## Inicio Rápido

### Crear Proyecto Web
```bash
# Proyecto web completo (default)
joss new mi_proyecto
cd mi_proyecto
joss server start
```

### Crear Proyecto de Consola
```bash
# Proyecto backend-only (sin UI)
joss new console mi_app_consola
cd mi_app_consola
joss run main.joss
```

## Comandos CLI

```bash
# Gestión de Proyectos
joss new [ruta]               # Crea proyecto web
joss new console [ruta]       # Crea proyecto de consola
joss new web [ruta]           # Crea proyecto web (explícito)

# Desarrollo
joss server start             # Inicia servidor HTTP (puerto 8000)
joss run [archivo]            # Ejecuta un script .joss
joss build                    # Compila para producción

# Base de Datos
joss migrate                  # Ejecuta migraciones pendientes
joss change db [mysql|sqlite] # Cambia motor de base de datos

# Generadores
joss make:controller [Nombre] # Crea controlador
joss make:model [Nombre]      # Crea modelo

# Utilidades
joss version                  # Muestra versión
joss help                     # Muestra ayuda
```

## Estructura de Proyecto

### Proyecto Web
```
mi_proyecto/
├── main.joss              # Entry Point
├── env.joss               # Variables de Entorno
├── api.joss               # Rutas API (JSON)
├── routes.joss            # Rutas Web (HTML)
├── config/
│   ├── reglas.joss        # Constantes Globales
│   └── cron.joss          # Tareas Programadas
├── app/
│   ├── controllers/       # Lógica de Negocio
│   ├── models/            # Acceso a Datos
│   ├── views/             # Plantillas HTML
│   ├── libs/              # Extensiones
│   └── database/
│       └── migrations/    # Migraciones
├── assets/
│   ├── css/               # Estilos (SCSS)
│   ├── js/                # JavaScript
│   └── images/            # Imágenes
└── public/                # Archivos públicos compilados
```

### Proyecto de Consola
```
mi_app_consola/
├── main.joss              # Entry Point
├── env.joss               # Variables de Entorno
├── config/
│   └── reglas.joss        # Constantes Globales
└── app/
    ├── controllers/       # Lógica de Negocio
    ├── models/            # Acceso a Datos
    ├── libs/              # Extensiones
    └── database/
        └── migrations/    # Migraciones
```

## Sintaxis Básica

### Variables y Tipos
```joss
// Tipos primitivos
int $edad = 25
float $precio = 99.99
string $nombre = "Jose"
bool $activo = true

// Arrays
array $lista = ["A", "B", "C"]
$mapa = {"key": "value"}
```

### Operadores Ternarios (No hay if/else)
```joss
// Ternario simple
$estado = ($edad >= 18) ? "Mayor" : "Menor"

// Ternario con bloques
($usuario->esValido()) ? {
    DB::save($usuario)
    print("Usuario guardado")
} : {
    print("Usuario inválido")
}

// Escalera lógica
$nivel = ($puntos > 1000) ? "Oro"
         ($puntos > 500)  ? "Plata" :
                            "Bronce"
```

### Clases y Herencia
```joss
class Animal {
    string $nombre
    
    Init constructor($n) {
        $this->nombre = $n
    }
    
    function hablar() {
        print("...")
    }
}

class Perro extends Animal {
    function hablar() {
        print("Guau!")
    }
}

$perro = new Perro("Rex")
$perro->hablar()  // "Guau!"
```

### Loops
```joss
// Foreach (principal)
foreach ($usuarios as $u) {
    print($u->nombre)
}

// Foreach con índice
foreach ($items as $i => $item) {
    print("$i: $item")
}
```

### Try-Catch
```joss
try {
    $resultado = operacionRiesgosa()
    print($resultado)
} catch ($error) {
    print("Error: " . $error)
}
```

## Módulos Nativos

### Auth (Autenticación)
```joss
// Registro con hash automático
Auth::create(["user@example.com", "password123", "Juan Pérez"])

// Login
$success = Auth::attempt("user@example.com", "password123")

// Verificar sesión
($Auth::check()) ? {
    $nombre = Auth::user()
    print("Hola " . $nombre)
} : {
    print("No autenticado")
}

// Logout
Auth::logout()
```

### GranMySQL (Base de Datos)
```joss
// API Fluida
$db = new GranMySQL()
$usuarios = $db->table("users")
               ->where("edad", ">", 18)
               ->get()

// Inserción
$db->table("users")
   ->insert(["nombre", "email"], ["Juan", "juan@example.com"])

// Primera coincidencia
$usuario = $db->table("users")
              ->where("email", "juan@example.com")
              ->first()
```

### Router (Rutas)
```joss
// routes.joss
Router::get("/", "HomeController@index")
Router::post("/login", "AuthController@login")

// Middleware
Router::middleware("auth")
Router::get("/dashboard", "DashboardController@index")
Router::end()

// Múltiples métodos
Router::match("GET|POST", "/contact", "ContactController@show@submit")
```

### View (Plantillas)
```joss
// Renderizar vista
return View::render("welcome", {"nombre": "Juan"})

// En la vista (app/views/welcome.joss.html)
<h1>Hola {{nombre}}</h1>
```

### SmtpClient (Correo)
```joss
$mail = new SmtpClient()
$mail->auth(env("MAIL_USER"), env("MAIL_PASS"))
$mail->send("destino@example.com", "Asunto", "Cuerpo del mensaje")
```

### Cron (Tareas Programadas)
```joss
// config/cron.joss
Cron::schedule("backup", "00:00", {
    System::backupDatabase()
})
```

## Async/Await
```joss
// Crear Future
$future = async(operacionLenta())

// Esperar resultado
$resultado = await($future)

// Múltiples operaciones
$f1 = async(consulta1())
$f2 = async(consulta2())
$r1 = await($f1)
$r2 = await($f2)
```

## Configuración

### env.joss
```bash
APP_ENV="development"
PORT="8000"

# Base de datos
DB="sqlite"
DB_PATH="database.sqlite"

# MySQL (alternativa)
# DB="mysql"
# DB_HOST="localhost"
# DB_NAME="mi_db"
# DB_USER="root"
# DB_PASS=""

# Prefijo de tablas
DB_PREFIX="js_"

# JWT
JWT_SECRET="tu_secreto_aqui"
```

## Migraciones

### Crear Migración
```joss
// app/database/migrations/001_create_users.joss
$schema = new Schema()
$schema->create("users", {
    "id": "INT AUTO_INCREMENT PRIMARY KEY",
    "name": "VARCHAR(255)",
    "email": "VARCHAR(255) UNIQUE",
    "password": "VARCHAR(255)",
    "created_at": "TIMESTAMP DEFAULT CURRENT_TIMESTAMP"
})
```

### Ejecutar Migraciones
```bash
joss migrate
```

## Ejemplos

Ver el directorio `examples/` para ejemplos completos:
- `final_test.joss`: Test comprehensivo de todas las características
- `jwt_test.joss`: Autenticación con JWT
- `jwt_refresh_test.joss`: Refresh tokens

## Documentación Completa

📚 Ver la carpeta [`docs/`](./docs/) para documentación detallada:

- [Sintaxis](./docs/SINTAXIS.md) - Operadores, variables, clases, loops
- [CLI](./docs/CLI.md) - Todos los comandos disponibles
- [Módulos Nativos](./docs/MODULOS_NATIVOS.md) - Auth, GranMySQL, Router, etc.
- [Estructura de Proyecto](./docs/ESTRUCTURA_PROYECTO.md) - Web y consola
- [Configuración](./docs/CONFIGURACION.md) - env.joss, reglas.joss
- [Migraciones](./docs/MIGRACIONES.md) - Sistema de base de datos
- [Servidor](./docs/SERVIDOR.md) - Hot reload, SCSS, WebSocket
- [Ejemplos](./docs/EJEMPLOS.md) - CRUD, Auth, API REST

## Desarrollo

El proyecto está en desarrollo activo:
- ✅ Fase 1: Smart Numerics y Maps
- ✅ Fase 2: Autoloading
- ✅ Fase 3: Concurrencia (async/await)
- ✅ Fase 4: Proyectos de Consola

## Licencia

Software cerrado fuente, derechos reservados.
