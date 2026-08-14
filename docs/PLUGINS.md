# Plugins y JP (Joss Plugin Bytecode)

El sistema de plugins de Joss genera, compila, carga y ejecuta paquetes binarios `.jp` (JP v2) de alto rendimiento, portables y firmados criptográficamente.

```bash
joss pub add mi_plugin ^1.0.0
joss pub install
```

```yaml
dependencies:
  mi_plugin: "^1.0.0"
```

---

## 🛠️ Crear y Compilar Plugins en Joss

Para crear un nuevo proyecto de plugin oficial:

```bash
joss new plugin mi_plugin
cd mi_plugin
```

Esto generará la estructura limpia del proyecto:
- `joss.yaml`: Manifiesto del paquete.
- `src/plugin.joss`: Implementación orientada a objetos en Joss.
- `.github/workflows/release.yml`: Workflow para compilación automática a `.jp` y publicación en GitHub Releases.

Para compilar el plugin a un paquete binario `.jp`:

```bash
joss plugin compile .
```

---

## ⚡ Compilación Multilenguaje a Bytecode Joss (.jp / JPBC)

Joss cuenta con un motor compilador (`joss plugin compile`) que permite desarrollar componentes en **Java, Python, PHP, C/C++, Rust, Kotlin, Dart o Flutter** y traducirlos automáticamente a **Bytecode nativo de Joss (JPBC)**.

### Características del Sistema:
1. **Tree Shaking Avanzado**: Analiza el Grafo de Llamadas (Call Graph) a partir de las funciones indicadas en `--exports` y elimina automáticamente todo el código muerto, clases e instrucciones no utilizadas.
2. **Cero Dependencias en el Usuario Final**: El paquete `.jp` resultante (< 10 KB a 1 MB) contiene bytecode determinista y optimizado. **El usuario final no requiere tener instalado Java (JVM/JDK), Python, PHP, Node ni compilar nada**.
3. **Firma Criptográfica Ed25519**: Cada paquete `.jp` es firmado criptográficamente de manera transparente durante la compilación (`~/.joss/keys/<plugin>.ed25519`) para garantizar su integridad y autoría.
4. **Índice de Símbolos Estándar (`SymbolIndex`)**: Genera metadatos tipados en `META-INF/joss-symbols.json` para proveer autocompletado nativo en editores e IDEs.

### Ejemplos de Compilación por Lenguaje:

```bash
# 1. Compilar plugin desarrollado en Java (.class o .jar)
joss plugin compile MiPlugin.jar --lang=java --name=music-plugin --exports=searchSong,getSong

# 2. Compilar plugin desarrollado en Python (.py)
joss plugin compile script.py --lang=python --name=tax-plugin --exports=calculate_tax

# 3. Compilar plugin desarrollado en Rust / C / C++ (WebAssembly .wasm)
joss plugin compile module.wasm --lang=rust --name=crypto-plugin --exports=encrypt_payload --permissions=filesystem.read

# 4. Inspeccionar la firma, funciones exportadas, clases y tamaño de un paquete .jp
joss plugin inspect music-plugin.jp
```

---

## 🔒 Estructura y Validación del Formato `.jp` (JP v2)

Un contenedor `.jp` es un archivo ZIP determinista con compresión DEFLATE que incluye:
- `META-INF/joss-plugin.json`: Metadatos del plugin, exportaciones, permisos, versión y firma digital Ed25519.
- `META-INF/joss-symbols.json`: Índice tipado de símbolos (`SymbolIndex` Schema v1) con clases, propiedades, métodos y funciones.
- `bytecode/main.jbc`: Bytecode compilado binario optimizado (JPBC o AST Codificado).
- `joss.yaml`: Manifiesto declarativo del paquete.

> **Aclaración de Terminología**:
> - `.jp`: Contenedor final distribuible.
> - `main.jbc`: Archivo interno de bytecode dentro de `.jp`.
> - `JPBC`: Formato binario / magic header (`0x4A 0x50 0x42 0x43`) del bytecode multilenguaje.
> - *No existe ningún archivo con extensión `.jpbc`*.

---

## 🚀 Integración y Consumo en Programas Joss

Joss descubre y carga automáticamente los paquetes `.jp` declarados en `joss.yaml` o instalados en el directorio `plugins/`.

### 1. Invocación de Funciones Directas o Calificadas:
```joss
// Llamada calificada por namespace del plugin
$resultado = joss_ai::predict(5)

// Llamada directa a función exportada
$descuento = calculate_discount(200, 10)
```

### 2. Instanciación de Clases y Llamada a Métodos:
```joss
// Instanciación nativa de clases del plugin
$tax = new TaxService()
$total = $tax->calculate(100)
```

---

## ⚙️ Arquitectura del Runtime (`pkg/pluginruntime` & `pkg/core`)

```text
               Programa Joss (Código fuente o bytecode)
                                 │
                                 ▼
                     Evaluator & Resolver (pkg/core)
                                 │
                 ┌───────────────┴───────────────┐
                 │                               │
                 ▼                               ▼
       Invocación Calificada            Instanciación
        (Plugin::function)          (new PluginClass())
                 │                               │
                 └───────────────┬───────────────┘
                                 │
                                 ▼
                 PluginRegistry (pkg/pluginruntime)
                                 │
                     Detección Dual de Bytecode
                    /                         \
                   /                           \
          JOSSBC2Z                              JPBC
             │                                    │
             ▼                                    ▼
       JossASTExecutor                          JPBCVM
    (pkg/core AST Engine)               (17 OpCodes Dispatcher)
             │                                    │
             └─────────────────┬──────────────────┘
                               │
                               ▼
                   Retorno de Valor Tipado a Joss
```

---

## 📚 Plugins Oficiales

- [joss_ai](https://github.com/joss-language/joss_ai)
- [joss_smtp](https://github.com/joss-language/joss_smtp)
- [joss_notify](https://github.com/joss-language/joss_notify)
- [joss_backup](https://github.com/joss-language/joss_backup)
- [joss_bg_remover](https://github.com/joss-language/joss_bg_remover)
