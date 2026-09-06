# El lenguaje de programación Joss

[![Go Report Card](https://goreportcard.com/badge/github.com/jossecurity/joss)](https://goreportcard.com/report/github.com/jossecurity/joss)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Docs](https://img.shields.io/badge/docs-joss.red-teal.svg)](https://joss.red/docs)

**Joss** es un lenguaje de programación moderno, tipado y de alto rendimiento implementado en Go. Está especialmente diseñado para el desarrollo backend, servicios web concurrentes, utilidades de línea de comandos y sistemas empresariales con seguridad integrada.

---

## ¿Por qué Joss?

- **Cero imports en el código (Zero-Imports)**: Olvídate de gestionar árboles interminables de `import` o `require`. Joss descubre, analiza y organiza automáticamente todas las clases y funciones públicas de tu proyecto y sus plugins.
- **Análisis estático exhaustivo**: Antes de ejecutar una sola línea, el analizador semántico de Joss (`joss analyze`) verifica tipos de datos, rutas de retorno de funciones, accesibilidad y tablas de símbolos, atrapando errores antes de que lleguen a tus usuarios.
- **Pila web y backend integrada (Batteries-Included)**: Incluye de forma nativa servidor HTTP multinivel, sistema de enrutamiento con parámetros dinámicos, motor de vistas HTML con protección anti-XSS, ORM y generador de esquemas (GranDB / Schema), autenticación, hashing criptográfico y WebSockets.
- **Concurrencia limpia con Canales y Async**: Ejecuta tareas pesadas en segundo plano con `async { ... }`, espera resultados con `await` y comunica procesos de forma segura mediante canales (`channel`).
- **Aritmética financiera exacta**: Incorpora el tipo primitivo `decimal` para cálculos monetarios en base diez sin las imprecisiones del estándar binario IEEE-754.

---

## Instalación rápida

### Instalador automático oficial

En **Windows** (PowerShell):
```powershell
iwr -useb https://raw.githubusercontent.com/josprox/Joss-language/main/install/remote-install.ps1 | iex
```

En **Linux o macOS** (Terminal):
```bash
curl -fsSL https://raw.githubusercontent.com/josprox/Joss-language/main/install/remote-install.sh | bash
```

### Compilar desde las fuentes

Si tienes [Go](https://go.dev) (versión 1.22 o superior) instalado:
```bash
git clone https://github.com/jossecurity/joss.git
cd joss
go build -o joss ./cmd/joss
```

Verifica la instalación:
```bash
joss version
```

---

## Tu primer programa en 30 segundos

Crea un archivo llamado `hola.joss`:

<!-- joss-run: ["Hola, Joss!"] -->
```joss
print("Hola, Joss!")
```

Analízalo y ejecútalo directamente desde tu terminal:

```bash
joss analyze hola.joss
joss run hola.joss
```

Salida:
```text
Hola, Joss!
```

---

## Un vistazo al lenguaje

Observa cómo se combinan funciones tipadas, inferencia, concatenación y decisiones con el operador ternario unificado:

<!-- joss-run: ["Hola, Ada", "Puedes participar"] -->
```joss
public func saludar(string $nombre): string {
    return "Hola, " . $nombre
}
$edad = 20
print(saludar("Ada"))
print(($edad >= 18) ? "Puedes participar" : "Aún debes esperar")
```

En Joss:
- Todas las variables comienzan con `$`.
- La primera asignación infiere y fija el tipo concreto (o puedes usar `mixed` para dinamismo explícito).
- La concatenación de texto se realiza con el operador punto (`.`), reservando `+` exclusivamente para la suma matemática.
- Las decisiones se expresan de forma elegante con ternarios `(condición) ? { ... } : { ... }` o con la expresión `match`.

---

## Mapa de la documentación

La documentación completa y oficial de Joss está organizada en cuatro áreas para acompañarte desde tu primera línea de código hasta la arquitectura del compilador:

### 1. Aprender Joss desde cero (Tutorial guiado)
- 0. [Primeros pasos: De un archivo a un programa](docs/PRIMEROS_PASOS.md)
- 1. [Valores, variables y operaciones fundamentales](docs/FUNDAMENTOS.md)
- 2. [Control de flujo, decisiones y bucles](docs/CONTROL_FLUJO.md)
- 3. [Funciones, ámbito (scope), closures y referencias](docs/FUNCIONES.md)
- 4. [Colecciones: Arrays, Maps y texto Unicode](docs/COLECCIONES.md)
- 5. [Sistema de tipos, inferencia y conversiones](docs/SISTEMA_TIPOS.md)
- 6. [Clases, objetos, métodos y herencia](docs/CLASES.md)
- 7. [Manejo de errores, excepciones y try/catch](docs/ERRORES.md)
- 8. [Concurrencia, asincronía, Future y canales](docs/CONCURRENCIA.md)
- 9. [Proyecto práctico: Aplicación de consola con persistencia JSON](docs/PROYECTO_CONSOLA.md)
- 10. [Proyecto práctico: Aplicación web MVC con el stack nativo](docs/PROYECTO_WEB.md)
- [Glosario completo de términos de programación](docs/GLOSARIO.md)

### 2. Referencia técnica exhaustiva
- [Sintaxis, tokens y precedencia de operadores](docs/SINTAXIS.md)
- [Gramática EBNF formal y correspondencia con el AST](docs/GRAMATICA.md)
- [Catálogo y referencia de diagnósticos (JOSS-*)](docs/DIAGNOSTICOS.md)
- [Guía de las 117 funciones globales integradas](docs/FUNCIONES_GLOBALES.md)
- [Clases nativas y servicios del runtime](docs/MODULOS_NATIVOS.md)
- [Catálogo nativo generado por docgen](docs/CATALOGO_NATIVO.md)
- [Referencia de comandos de la CLI (`joss`)](docs/CLI.md)
- [Extensión oficial para Visual Studio Code](docs/VSCODE_EXTENSION.md)
- [Estado real de implementación y límites del sistema](docs/ESTADO_IMPLEMENTACION.md)

### 3. Desarrollo de aplicaciones reales
- [Estructura de proyectos y convenciones](docs/ESTRUCTURA_PROYECTO.md)
- [Configuración y variables de entorno](docs/CONFIGURACION.md)
- [Carga automática y Zero Imports](docs/MODULOS_IMPORTS.md)
- [Arquitectura de plugins y paquetes binarios JP](docs/PLUGINS.md)
- [Servidor HTTP nativo de alto rendimiento](docs/SERVIDOR.md)
- [Controladores web y peticiones HTTP](docs/CONTROLADORES.md)
- [Middlewares y capas de seguridad](docs/MIDDLEWARE.md)
- [Motor de vistas y plantillas HTML dinámicas](docs/VISTAS.md)
- [Gestión y compilación de assets](docs/ASSETS.md)
- [WebSockets y comunicación en tiempo real](docs/WEBSOCKETS.md)
- [Modelos relacionales y consultas GranDB](docs/MODELOS.md)
- [Schema Builder y migraciones versionadas](docs/SCHEMA_BUILDER.md)
- [Sistema de autenticación, sesiones y MFA](docs/AUTENTICACION.md)

### 4. Internals y cómo contribuir al lenguaje
- [Arquitectura del compilador, analizador y runtime](docs/ARQUITECTURA.md)
- [Guía para contribuidores del núcleo](docs/CONTRIBUIR.md)
- [Informe de auditoría integral y reconstrucción documental](docs/DOCUMENTATION_AUDIT.md)
- [Auditoría técnica y optimización del runtime](docs/AUDITORIA_TECNICA_2026.md)

---

## Cómo contribuir

Las contribuciones son muy bienvenidas. Antes de enviar cambios a la sintaxis, tipos o runtime, por favor lee nuestra [Guía de arquitectura y reglas operativas (AGENTS.md)](AGENTS.md) y la [Guía de contribución](docs/CONTRIBUIR.md).

Para ejecutar la suite completa de pruebas:
```bash
go test ./...
go test -race ./pkg/parser ./pkg/typesystem ./pkg/analyzer ./pkg/core
```

---

## Licencia y Seguridad

Joss es software libre publicado bajo la [Licencia MIT](LICENSE). Para reportar vulnerabilidades de seguridad, consulta [SECURITY.md](SECURITY.md).
