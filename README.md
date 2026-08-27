<p align="center">
  <img src="./assets/JosSecurity%20logo%20color/default.png" alt="Joss Programming Language" width="280">
</p>

<h1 align="center">Joss Programming Language</h1>

<p align="center">
  <strong>Un lenguaje de programación y framework backend moderno, seguro y extensible de alto rendimiento construido sobre Go.</strong><br>
  <em>A modern, high-performance, secure, and extensible backend programming language and framework built on Go.</em>
</p>

<p align="center">
  <a href="https://joss.red/docs"><img alt="Documentación" src="https://img.shields.io/badge/docs-joss.red-ff5f6d?style=flat-square"></a>
  <a href="https://joss.red/pub"><img alt="Joss Pub" src="https://img.shields.io/badge/pub-librerías-ff8a65?style=flat-square"></a>
  <a href="https://github.com/josprox/Joss-language/releases"><img alt="Latest Release" src="https://img.shields.io/github/v/release/josprox/Joss-language?color=3f51b5&label=release&style=flat-square"></a>
  <img alt="Plataformas" src="https://img.shields.io/badge/platform-Windows%20%7C%20Linux%20%7C%20macOS-263238?style=flat-square">
  <a href="./LICENSE"><img alt="Licencia MIT" src="https://img.shields.io/badge/license-MIT-2e7d32?style=flat-square"></a>
</p>

<p align="center">
  <a href="https://joss.red/docs">Documentación</a> ·
  <a href="https://joss.red/pub">Joss Pub</a> ·
  <a href="./docs/PLUGINS.md">Crear Plugins</a> ·
  <a href="https://github.com/josprox/Joss-language/releases">Descargas / Releases</a>
</p>

---

## 🌐 Summary & Highlights / Resumen Multilingüe

### 🇪🇸 Español
**Joss Programming Language** combina la velocidad de desarrollo expresiva con un motor de ejecución eficiente escrito en Go. Incluye tipado dinámico y validado, concurrencia con `async`/`await`, servidor HTTP/HTTPS nativo, recarga en vivo al estilo Flutter, motor de base de datos GranDB (SQLite, MySQL, PostgreSQL), WebSockets en tiempo real y arquitectura extensible mediante paquetes JP v2 firmados.

### 🇺🇸 English
**Joss Programming Language** combines expressive development speed with a high-performance execution runtime written in Go. It features dynamic and validated typing, `async`/`await` concurrency, a native HTTP/HTTPS server, Flutter-style live hot reload, the GranDB database engine (SQLite, MySQL, PostgreSQL), real-time WebSockets, and an extensible plugin system using signed JP v2 packages.

### 🇵🇹 Português
O **Joss Programming Language** combina a velocidade de desenvolvimento expressiva com um mecanismo de execução de alto desempenho escrito em Go. Possui tipagem dinâmica e validada, concorrência com `async`/`await`, servidor HTTP/HTTPS nativo, recarga em tempo real estilo Flutter, motor de banco de dados GranDB (SQLite, MySQL, PostgreSQL), WebSockets em tempo real e uma arquitetura extensível por meio de pacotes JP v2 assinados.

---

## 🚀 Capacidades y Características Principales

| Área | Descripción e Integraciones |
| --- | --- |
| **Lenguaje Core** | Tipos dinámicos/validados, clases OOP, funciones, closures, ternarios de bloque, arrays, maps y evaluación de expresiones. |
| **Backend & HTTP** | Router expresivo, Request/Response, middleware, vistas Blade, sesiones en archivo/memory/redis, JWT y WebSockets. |
| **Hot Reload Vivo** | Recarga en vivo estilo Flutter: recargas instantáneas sin refrescar la página, actualización de CSS/views y avisos flotantes. |
| **GranDB ORM** | Motor de base de datos para SQLite, MySQL/MariaDB y PostgreSQL, migraciones, seeders y Schema Builder. |
| **Seguridad Avanzada** | Cifrado de entorno, CSRF, cookies HTTP-only, rate limiting por IP y utilidades criptográficas. |
| **Concurrencia & Cron** | Fork de runtime con `async`/`await`, canales thread-safe y motor Cron programado dinámicamente en memoria. |
| **Extensibilidad JP v2** | Plugins autocontenidos firmados con Ed25519, carga automática sin `import`, RPC aislado y ABI C en memoria. |
| **Ecosistema & SDKs** | CLI completo, extensión oficial para VS Code y SDKs nativos para C, C++, Python, Rust, Java, Kotlin, PHP y Dart. |

---

## 📦 Instalación Rápida

El instalador automático configura el runtime de **Joss Programming Language**, el SDK para plugins y la extensión oficial de VS Code.

### 💻 Windows
Abre PowerShell como Administrador y ejecuta:

```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope Process; iwr -useb https://raw.githubusercontent.com/josprox/Joss-language/main/install/remote-install.ps1 | iex
```

### 🐧 Linux & 🍎 macOS
Abre una terminal y ejecuta:

```bash
curl -fsSL https://raw.githubusercontent.com/josprox/Joss-language/main/install/remote-install.sh | bash
```

Comprueba la instalación con:

```bash
joss version
```

---

## 🛠️ Creación de Proyectos

### Aplicación Web (MVC por Dominios)
```bash
joss new mi_aplicacion
cd mi_aplicacion
joss server start
```

### Aplicación de Consola / Backend Puro
```bash
joss new console mi_herramienta
cd mi_herramienta
joss run main.joss
```

---

## 💻 Sintaxis de Ejemplo

### 1. Variables, Clases y Concatenación Explicita (`.`)
En Joss, la concatenación de cadenas se realiza **estrictamente con el operador punto (`.`)**, mientras que `+` es exclusivo para operaciones matemáticas numéricas.

```joss
public class SaludoService {
    public func saludar(string $nombre, int $edad) {
        ($edad < 18) ? {
            return "Acceso restringido para menores"
        } : {
            return "Bienvenido a Joss, " . $nombre . "!"
        }
    }
}

$service = new SaludoService()
print($service->saludar("Joss", 25))
```

### 2. Concurrencia con `async` / `await`
```joss
$future = async {
    // Proceso asíncrono en background
    return 40 + 2
}

$resultado = await($future)
print("Resultado: " . $resultado)
```

### 3. Enrutamiento HTTP & GranDB ORM
```joss
Router::get("/api/usuarios/{id}", func($id) {
    $usuario = GranDB::table("users")->find($id)
    
    ($usuario == null) ? {
        return Response::error("Usuario no encontrado", 404)
    } : {
        return Response::json({
            "ok": true,
            "data": $usuario
        })
    }
})
```

---

## 📚 Ecosistema & Plugins Oficiales

Las dependencias y paquetes declarados en `joss.yaml` se instalan con el CLI y se cargan e indexan **automáticamente sin necesidad de sentencias `use` o `import`**.

```bash
joss pub search ai
joss pub add joss_ai 2.0.1
joss pub install
```

### Librerías Oficiales Destacadas:

| Paquete | Propósito | Repositorio Oficial |
| --- | --- | --- |
| **`joss_ai`** | Clientes de IA (OpenAI, Groq, Gemini) y streaming en tiempo real. | [josprox/joss_ai](https://github.com/josprox/joss_ai) |
| **`joss_smtp`** | Envío seguro de correos electrónicos vía SMTP, STARTTLS y TLS. | [josprox/joss_smtp](https://github.com/josprox/joss_smtp) |
| **`joss_notify`** | Notificaciones push, webhooks y alertas in-app. | [josprox/joss_notify](https://github.com/josprox/joss_notify) |
| **`joss_backup`** | Creación, verificación y restauración automatizada de respaldos. | [josprox/joss_backup](https://github.com/josprox/joss_backup) |

```joss
// Ejemplo de uso de joss_ai cargado automáticamente
$respuesta = AI::client()
    ->provider("groq")
    ->model("llama-3.3-70b-versatile")
    ->system("Responde de forma breve")
    ->user("¿Qué es Joss Programming Language?")
    ->call()
```

---

## 🛠️ SDK Multilenguaje para Plugins Nativo (JP v2)

Joss permite extender sus capacidades mediante paquetes comprimidos `.jp` firmados con Ed25519 que incluyen ejecutables nativos o bibliotecas C ABI en memoria.

| Tecnología | Recurso SDK Incluido |
| --- | --- |
| **C** | Encabezados RPC `sdk/c/joss_plugin.h` y ABI C `sdk/c/joss_driver.h`. |
| **C++ (C++17)** | Framework orientado a objetos `sdk/c/joss_plugin.hpp` con registro de métodos y exception safety. |
| **Python** | Runner `sdk/python/joss_plugin.py` con decoradores, `asyncio` y generadores streaming. |
| **Rust** | Crate `sdk/rust` con builder OOP y capturador de pánicos `catch_unwind`. |
| **Java** | Contrato `sdk/java/JossPlugin.java` para GraalVM Native Image o JVM. |
| **Kotlin** | DSL `sdk/kotlin/JossPlugin.kt` para Kotlin/Native y Kotlin/JVM. |
| **PHP** | Class runner `sdk/php/joss_plugin.php` con streaming vía `Generator`. |
| **Dart & Flutter** | Runner `sdk/dart/joss_plugin.dart` con `Stream` chunking para Flutter Desktop. |

---

## ⚡ Comandos CLI Esenciales

```bash
# Servidor y Ejecución
joss server start               # Inicia el servidor con Live Hot Reload
joss run main.joss              # Ejecuta un script de Joss directamente
joss update                     # Actualiza el motor de Joss a la última versión
joss build native               # Compila un binario autoejecutable nativo

# Generadores CLI (Soporta subcarpetas por dominio)
joss make:controller web/CatalogController
joss make:model auth/User
joss make:view productos/index
joss make:crud productos
joss remove:crud productos

# Base de Datos y Migraciones
joss migrate                    # Ejecuta migraciones pendientes
joss migrate:fresh              # Reconstruye esquema desde cero
joss db:seed                    # Corre seeders de datos
joss change db mysql            # Cambia motor de BD a MySQL, SQLite o Postgres

# Gestión de Paquetes (Joss Pub)
joss pub add joss_ai 2.0.1
joss pub install
joss build package .            # Empaqueta un plugin JP v2 firmado
```

---

## 📖 Documentación y Recursos

- **Sitio Web Oficial**: [https://joss.red](https://joss.red)
- **Documentación Completa**: [joss.red/docs](https://joss.red/docs)
- **Catálogo Joss Pub**: [joss.red/pub](https://joss.red/pub)
- **Guía de Sintaxis**: [Sintaxis del Lenguaje](./docs/SINTAXIS.md)
- **Estructura de Proyectos**: [Estructura por Dominios](./docs/ESTRUCTURA_PROYECTO.md)
- **Desarrollo de Plugins**: [Guía de Plugins JP v2](./docs/PLUGINS.md)
- **Guía de CLI**: [Referencia del CLI](./docs/CLI.md)

---

## 📄 Licencia

**Joss Programming Language** se distribuye bajo la [Licencia MIT](./LICENSE). Las aplicaciones creadas con el lenguaje pertenecen a sus respectivos autores.
