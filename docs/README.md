# Índice de Documentación - JosSecurity

Guía completa de JosSecurity v3.0 (Gold Master)

## 📚 Documentos Disponibles

### Fundamentos
- [SINTAXIS.md](./SINTAXIS.md) - Sintaxis completa del lenguaje
  - Variables y tipos
  - Operadores ternarios (reemplazo de if/else)
  - Clases y herencia
  - Funciones
  - Loops (foreach)
  - Try-Catch
  - Arrays y Maps
- [CONCURRENCIA.md](./CONCURRENCIA.md) - Programación concurrente
  - Async/Await
  - Canales (Channels)

### Herramientas
- [CLI.md](./CLI.md) - Comandos de línea de comandos
  - Gestión de proyectos (new, new console)
  - Desarrollo (server, run, build)
  - Base de datos (migrate, change db)
  - Generadores (make:controller, make:model)
- [VSCODE_EXTENSION.md](./VSCODE_EXTENSION.md) - Extensión para VS Code (IntelliSense, Highlighting)

### Módulos
- [MODULOS_NATIVOS.md](./MODULOS_NATIVOS.md) - Módulos nativos del lenguaje
  - Auth - Autenticación y JWT
  - GranMySQL - ORM de base de datos
  - Router - Sistema de rutas
  - View - Motor de plantillas
  - SmtpClient - Correo electrónico
  - Response/Request - HTTP
  - Cron/Task - Tareas programadas
  - Schema - Esquemas de BD
  - System - Utilidades
  - Redis - Cache
  - Queue - Colas
  - WebSocket - Tiempo real

### Proyecto
- [ESTRUCTURA_PROYECTO.md](./ESTRUCTURA_PROYECTO.md) - Organización de archivos
  - Proyecto web (completo)
  - Proyecto de consola (backend-only)
  - Convenciones de nombres
  - Organización recomendada

- [CONFIGURACION.md](./CONFIGURACION.md) - Configuración del proyecto
  - env.joss - Variables de entorno
  - config/reglas.joss - Constantes globales
  - config/cron.joss - Tareas programadas
  - Base de datos (SQLite/MySQL)
  - Correo (SMTP)
  - Redis
  - Seguridad

### Avanzado
- [MIGRACIONES.md](./MIGRACIONES.md) - Sistema de migraciones
- [SERVIDOR.md](./SERVIDOR.md) - Servidor HTTP
- [EJEMPLOS.md](./EJEMPLOS.md) - Ejemplos prácticos

---

## 🚀 Inicio Rápido

### 1. Instalación

### 1. Instalación

**Windows (PowerShell)**:
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope Process; iwr -useb https://raw.githubusercontent.com/josprox/JosSecurity-language/main/install/remote-install.ps1 | iex
```

**Linux/macOS**:
```bash
curl -fsSL https://raw.githubusercontent.com/josprox/JosSecurity-language/main/install/remote-install.sh | bash
```

**Manual (Desarrollo)**:
```bash
git clone https://github.com/josprox/JosSecurity-language.git
cd JosSecurity-language
go build -o joss.exe ./cmd/joss
```

### 2. Crear Proyecto

```bash
# Proyecto web
joss new mi_proyecto

# Proyecto de consola
joss new console mi_app
```

### 3. Configurar

```bash
cd mi_proyecto
# Editar env.joss con tu configuración
```

### 4. Ejecutar

```bash
# Web
joss server start

# Consola
joss run main.joss
```

---

## 📖 Guías por Tema

### Para Principiantes
1. [SINTAXIS.md](./SINTAXIS.md) - Aprender la sintaxis
2. [CLI.md](./CLI.md) - Comandos básicos
3. [EJEMPLOS.md](./EJEMPLOS.md) - Ejemplos prácticos

### Para Desarrollo Web
1. [ESTRUCTURA_PROYECTO.md](./ESTRUCTURA_PROYECTO.md) - Organizar proyecto
2. [MODULOS_NATIVOS.md](./MODULOS_NATIVOS.md#router) - Sistema de rutas
3. [MODULOS_NATIVOS.md](./MODULOS_NATIVOS.md#view) - Plantillas HTML
4. [MODULOS_NATIVOS.md](./MODULOS_NATIVOS.md#auth) - Autenticación

### Para Backend/Consola
1. [CLI.md](./CLI.md#joss-new-console-ruta) - Crear proyecto de consola
2. [MODULOS_NATIVOS.md](./MODULOS_NATIVOS.md#granmysql) - Base de datos
3. [MODULOS_NATIVOS.md](./MODULOS_NATIVOS.md#cron) - Tareas programadas

### Para Base de Datos
1. [CONFIGURACION.md](./CONFIGURACION.md#configuración-de-base-de-datos) - Configurar BD
2. [MIGRACIONES.md](./MIGRACIONES.md) - Sistema de migraciones
3. [MODULOS_NATIVOS.md](./MODULOS_NATIVOS.md#granmysql) - ORM GranMySQL

---

## 🔍 Búsqueda Rápida

### ¿Cómo hacer...?

**Autenticación**
- Registrar usuario → [MODULOS_NATIVOS.md#auth](./MODULOS_NATIVOS.md#auth)
- Login → [MODULOS_NATIVOS.md#authattemptstringemail-stringpassword](./MODULOS_NATIVOS.md#authattemptstringemail-stringpassword)
- Proteger rutas → [MODULOS_NATIVOS.md#routermiddlewarestringnombre](./MODULOS_NATIVOS.md#routermiddlewarestringnombre)

**Base de Datos**
- Consultar datos → [MODULOS_NATIVOS.md#granmysql](./MODULOS_NATIVOS.md#granmysql)
- Crear migración → [MIGRACIONES.md](./MIGRACIONES.md)
- Cambiar motor → [CLI.md#joss-change-db-motor](./CLI.md#joss-change-db-motor)

**Vistas**
- Renderizar HTML → [MODULOS_NATIVOS.md#view](./MODULOS_NATIVOS.md#view)
- Herencia de plantillas → [MODULOS_NATIVOS.md#herencia](./MODULOS_NATIVOS.md#herencia)
- Inclusión de parciales → [VISTAS.md#3-inclusión-de-vistas-parciales-include](./VISTAS.md#3-inclusión-de-vistas-parciales-include)
- Pasar datos → [MODULOS_NATIVOS.md#viewrenderstringnombre-mapdatos](./MODULOS_NATIVOS.md#viewrenderstringnombre-mapdatos)

**Rutas**
- Definir ruta → [MODULOS_NATIVOS.md#router](./MODULOS_NATIVOS.md#router)
- Middleware → [MODULOS_NATIVOS.md#routermiddlewarestringnombre](./MODULOS_NATIVOS.md#routermiddlewarestringnombre)
- API REST → [EJEMPLOS.md](./EJEMPLOS.md)

---

## 💡 Recursos Adicionales

- **Código Fuente**: `pkg/` y `cmd/`
- **Ejemplos**: `examples/`
- **Extensión VS Code**: `vscode-joss/`

---

## 🆘 Soporte

### Problemas Comunes
Ver sección "Solución de Problemas" en [CLI.md](./CLI.md#solución-de-problemas)

### Reportar Bugs
Crear issue en el repositorio con:
- Versión de JosSecurity (`joss version`)
- Sistema operativo
- Pasos para reproducir
- Código de ejemplo

---

**Versión**: JosSecurity v3.3.0
**Última actualización**: 2026-02-22
