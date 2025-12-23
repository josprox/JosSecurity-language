# Extensión de JosSecurity para VS Code

La extensión oficial de **JosSecurity** para Visual Studio Code proporciona un entorno de desarrollo rico y potente, diseñado para maximizar tu productividad con el lenguaje JOSS.

## Características Principales

### 1. Resaltado de Sintaxis (Syntax Highlighting)
Colores y formatos optimizados para todos los archivos del ecosistema:
- **.joss**: Código fuente (clases, funciones, variables).
- **.joss.html**: Plantillas de vista con soporte embebido de HTML.
- **env.joss** y archivos de configuración.

### 2. IntelliSense y Autocompletado
El servidor de lenguaje (LSP) analiza tu proyecto en tiempo real para ofrecer:
- **Completado de código**: Sugerencias inteligentes para clases (`Router`, `Auth`, `DB`), métodos y funciones.
- **Snippets**: Fragmentos de código rápidos para estructuras comunes como `class`, `function` (y `func`), `try-catch`, `foreach`, etc.
- **Variables de Vista**: Al editar archivos `.joss.html`, la extensión sugiere las variables pasadas desde el controlador mediante `View::render`.

### 3. Navegación (Go to Definition)
- **Ctrl + Click** en nombres de clases o funciones para saltar a su definición.
- Navegación fluida entre controladores y vistas.

### 4. Diagnóstico de Errores (LSP)
Detecta problemas antes de ejecutar el código:
- **Rutas Rotas**: Verifica que los controladores y métodos definidos en `routes.joss` realmente existan.
- **Controladores Faltantes**: Alerta si referencias un controlador que no está en el proyecto.
- **Seguridad**: Análisis estático básico para detectar patrones inseguros.

### 5. Soporte para Rutas Dinámicas
La extensión entiende la estructura de tu proyecto (`Router::get` y `Router.get`) sin importar dónde esté alojado, gracias a un sistema de detección de workspace dinámico.

## Instalación

### Desde el Marketplace (Próximamente)
Busca "JosSecurity Language" en el panel de extensiones de VS Code.

### Instalación Manual (.vsix)
Si tienes el archivo `.vsix` empaquetado:

1. Abre la paleta de comandos (`Ctrl+Shift+P` o `F1`).
2. Escribe y selecciona: **Extensions: Install from VSIX...**
3. Selecciona el archivo `joss-language-x.x.x.vsix`.
4. Recarga la ventana de VS Code.

O mediante terminal:
```bash
code --install-extension joss-language-x.x.x.vsix
```

## Solución de Problemas Common

### "Controller not found" (Falso Positivo)
Asegúrate de estar usando la versión **v3.2.8+** de la extensión, que corrige problemas con rutas de Windows y sintaxis de punto (`Router.get`).

### "Method not found" en funciones `func`
La versión **v3.2.9+** soporta nativamente la palabra clave `func`. Actualiza tu extensión si ves este error en código válido.

### El autocompletado no funciona
1. Verifica que tu proyecto tenga estructura válida (archivo `env.joss` en raíz).
2. Intenta recargar la ventana (`Developer: Reload Window`).
3. Revisa la salida del canal "JosSecurity Language Server" en la pestaña "Output" de VS Code.
