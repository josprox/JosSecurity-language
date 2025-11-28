# JosSecurity (Joss)

Lenguaje de programación moderno con enfoque en seguridad, inspirado en PHP, Python, Java y Go.

## Características Principales

### 🚀 Sistema de Tipos Robusto
- **Smart Numerics**: Promoción automática de int a float (división siempre retorna float)
- **Maps Nativos**: Sintaxis `{ key: value }` con soporte completo
- **Tipos Dinámicos**: Sistema flexible con optimización automática

### ⚡ Concurrencia
- **async/await**: Ejecución asíncrona aprovechando Goroutines de Go
- **Futures**: Manejo de valores asíncronos con canales de Go

### 🔐 Seguridad Integrada
- **Auth Module**: Autenticación con JWT
- **GranMySQL**: ORM seguro con protección contra SQL injection
- **Entorno Encriptado**: Variables de entorno en RAM

### 📦 Autoloading
- Carga automática de clases desde `./classes/`
- Sin necesidad de imports manuales

## Instalación

### Requisitos
- Go 1.20 o superior
- MySQL (para características de base de datos)

### Compilar
```bash
go build -o joss.exe ./cmd/joss
```

## Uso

### Ejecutar un Script
```bash
./joss.exe run examples/final_test.joss
```

### Comandos Disponibles
```bash
# Crear nuevo proyecto (Estructura Biblia)
./joss.exe new myproject

# Crear nuevo proyecto web (Estructura legacy)
./joss.exe new web mywebproject

# Ver versión
./joss.exe version

# Iniciar servidor
./joss.exe server start

# Ejecutar migraciones
./joss.exe migrate

# Crear controlador
./joss.exe make:controller UserController

# Crear modelo
./joss.exe make:model User
```

## Estructura de Proyecto

### Estructura Biblia (Por Defecto)
Siguiendo "La Gran Biblia de JosSecurity", el comando `joss new` crea:

```
myproject/
├── main.joss           # Entry Point
├── env.joss            # Variables de Entorno
├── api.joss            # Rutas API (JSON/TOON)
├── routes.joss         # Rutas Web (HTML)
├── config/
│   ├── reglas.joss     # Constantes Globales
│   └── cron.joss       # Tareas Programadas
├── app/
│   ├── controllers/    # Lógica de Negocio
│   ├── models/         # Acceso a Datos
│   ├── views/          # Plantillas HTML
│   └── libs/           # Extensiones
└── assets/             # CSS, JS, Imágenes
```

### Estructura Web (Legacy)
Para compatibilidad con proyectos anteriores, usa `joss new web`:

```
mywebproject/
├── main.joss
├── env.joss
├── routes.joss
├── api.joss
├── config/
│   └── global.joss
├── app/
│   ├── controllers/
│   ├── models/
│   ├── views/
│   ├── assets/
│   └── database/migrations/
└── public/
```

## Ejemplos

Ver el directorio `examples/` para ejemplos completos:
- `final_test.joss`: Test comprehensivo de todas las características
- `jwt_test.joss`: Autenticación con JWT
- `jwt_refresh_test.joss`: Refresh tokens

## Estructura del Proyecto

```
JosSecurity/
├── cmd/joss/          # CLI principal
├── pkg/
│   ├── core/          # Runtime y ejecución
│   ├── parser/        # Lexer, Parser y AST
│   └── server/        # Servidor HTTP
├── examples/          # Ejemplos de código
├── docs/              # Documentación
└── vscode-joss/       # Extensión de VS Code
```

## Sintaxis Básica

```joss
// Clases y Herencia
class Animal {
    string $type = "Animal"
    
    Init constructor($t) {
        $this->type = $t
    }
}

class Dog extends Animal {
    function makeSound() {
        print("Woof!")
    }
}

// Smart Numerics
$result = 10 / 3  // Retorna 3.333... (float)

// Maps Nativos
$config = {
    "host": "localhost",
    "port": 3306
}
print($config["host"])

// Async/Await
$future = async(10 + 20)
$result = await($future)  // 30

// Auth con JWT
Auth.create(["user@example.com", "password", "Name"])
$token = Auth.attempt("user@example.com", "password")
```

## Desarrollo

El proyecto está en desarrollo activo. Las tres fases principales están completadas:
- ✅ Fase 1: Smart Numerics y Maps
- ✅ Fase 2: Autoloading
- ✅ Fase 3: Concurrencia (async/await)

## Licencia

Software cerrado fuente, derechos reservados. 
