# Novedades de Joss v3.6.7

Joss v3.6.7 representa una evolución mayor en la plataforma Joss y el ecosistema **Joss Red / JosSecurity**, consolidando la **Arquitectura de Lenguaje Integral Modular (ALIM)** propuesta en la investigación doctoral de lenguajes de programación integrales.

---

## 🌟 Principales Características de la Versión 3.6.7

### 1. 🔬 Analizador Estático AST (`joss analyze`)
- **Ejecución Automática**: Al ejecutar `joss run main.joss` o `joss server start`, el motor realiza automáticamente una inspección estática del Árbol de Sintaxis Abstracta (AST) antes del runtime.
- **Detección Precoz de Errores**: Detecta variables no declaradas o sin uso, llamadas a funciones inexistentes y métodos no definidos sin interrumpir la ejecución si no hay faltas graves.
- **Comando Standalone**: Ejecuta `joss analyze` en la raíz de cualquier proyecto para validar todo el código fuente de forma aislada.

### 2. 🛡️ Guardias de Sintaxis para Palabras Clave Foráneas
- **Aviso Educativo Amigable**: Intercepta intentos de usar sintaxis de otros lenguajes (`if`, `else`, `elif`, `switch`, `for`).
- **Sugerencias de Sintaxis Idiomática**: El compilador sugiere la sintaxis nativa de Joss:
  - Operador Ternario: `$cond ? $a : $b`
  - Control de Flujo Match: `match ($var) { "opcion" => { ... } }`
  - Bucles Nativos: `while ($cond) { ... }` o `foreach ($array as $item) { ... }`

### 3. 🔒 Sandbox WASI y Permisos en Plugins (.jp)
- **Control de Acceso Granular (`PermissionGuard`)**: Los plugins binarios compilados (`.jp`) ejecutan en la máquina virtual `JPBCVM` bajo un modelo de aislamiento WASI estricto.
- **Permisos Declarativos**: Validación en tiempo de ejecución antes de invocar funciones del host:
  - `http_get`: Solicitudes HTTP de red.
  - `file_read`: Acceso a lectura del sistema de archivos.
  - `env_read`: Lectura de variables de entorno.
  - `db_query`: Consultas a la base de datos SQL.
- **Verificación Criptográfica**: Comando `joss plugin verify mi_plugin.jp` para validar la firma digital Ed25519 y la integridad del ZIP binario.

### 4. 📦 Plugins Multilenguaje Ultraligeros (< 2 KB)
- Compilación transparente de código escrito en **Python, Java, PHP y WASM (C/C++, Rust)** hacia Bytecode binario Joss (`JPBC`).
- **Tree-Shaking Avanzado**: Grafo de llamadas que elimina instrucciones y clases no utilizadas.
- Paquetes de distribución extremadamente eficientes entre **1.20 KB y 1.38 KB** sin requerir runtimes externos (sin JVM, sin intérprete de Python en el servidor).

### 5. 🏛️ Centralización de Funciones Nativas (Builtins)
- Registro único y canónico en `pkg/core/builtins.go`.
- Previene inconsistencias en el parser, analizador y runtime al agregar nuevas funciones nativas al motor.

---

## 🚀 Actualización del CLI y Herramientas

```bash
# Validar análisis estático sin ejecutar el proyecto
joss analyze

# Ejecución de aplicación web con análisis automático
joss server start

# Compilar plugin multilenguaje a bytecode firmado (.jp)
joss plugin compile main.py --lang=python --name=mi_plugin --exports=procesar

# Verificar firma Ed25519 e integridad de un plugin
joss plugin verify mi_plugin.jp
```

---

## 🛠️ Resumen de Cambios Técnicos

| Componente | Novedad / Mejora |
| :--- | :--- |
| **AST Analyzer** | Análisis estático de variables/funciones en `joss run`, `joss server start` y `joss analyze` |
| **Syntax Guard** | Detección amigable de `if`/`else`/`switch`/`for` sugiriendo `? :`, `match`, `while`, `foreach` |
| **Plugin VM** | Guardias de permisos WASI en `JPBCVM` para I/O y red |
| **Builtins Registry** | Registro centralizado en `pkg/core/builtins.go` |
| **Release Pipeline** | Empaquetado multiplataforma optimizado con `build_all.ps1` |
