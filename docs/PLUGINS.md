# Plugins y JP (Joss Plugin Bytecode)

El sistema de plugins de Joss genera, compila, carga y ejecuta paquetes binarios `.jp` (JP v2) de alto rendimiento, portables, aislados y firmados criptográficamente.

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
2. **Cero Dependencias en el Usuario Final**: El paquete `.jp` resultante (< 1.5 KB a 1 MB) contiene bytecode determinista y optimizado. **El usuario final no requiere tener instalado Java (JVM/JDK), Python, PHP, Node ni compilar nada**.
3. **Firma Criptográfica Ed25519**: Cada paquete `.jp` es firmado criptográficamente de manera transparente durante la compilación (`~/.joss/keys/<plugin>.ed25519`) para garantizar su integridad y autoría.
4. **Verificación de Paquetes**: Comando `joss plugin verify mi_plugin.jp` para validar la firma y estructura interna del paquete.
5. **Índice de Símbolos Estándar (`SymbolIndex`)**: Genera nombres, parámetros y tipos de retorno en `META-INF/joss-symbols.json` para el analyzer y el autocompletado. Paquetes antiguos sin `return_type` conservan retorno `unknown`, no un tipo inventado.

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

# 5. Verificar firma e integridad digital de un paquete .jp
joss plugin verify music-plugin.jp
```

---

## 🔒 Sandbox WASI y Permisos en el Runtime (`PermissionGuard`)

Los plugins ejecutan en la máquina virtual `JPBCVM` bajo un modelo estricto de **Aislamiento WASI**:

- **Permisos Declarativos**: Los permisos requeridos por el plugin se declaran en `joss.yaml` y en el manifiesto interno `META-INF/joss-plugin.json`.
- **Enforcement en Tiempo de Ejecución**: La máquina virtual valida explícitamente mediante `PermissionGuard` antes de realizar llamadas al sistema host:
  - `http_get`: Realizar peticiones HTTP salientes.
  - `file_read` / `file_write`: Acceso a lectura/escritura de archivos locales.
  - `env_read`: Acceso a variables de entorno del servidor host.
  - `db_query`: Ejecutar consultas SQL en la base de datos de la aplicación.

Si un plugin intenta realizar I/O o acceso a red sin contar con el permiso explícito concedido, la máquina virtual bloquea la ejecución de inmediato lanzando una excepción de seguridad.

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
       JossASTExecutor                          JPBCVM (WASI PermissionGuard)
    (pkg/core AST Engine)               (17 OpCodes Dispatcher + Sandbox)
             │                                    │
             └─────────────────┬──────────────────┘
                               │
                               ▼
                   Retorno de Valor Tipado a Joss
```

---

## 📚 Plugins Oficiales

- [joss_ai](https://github.com/joss-language/joss_ai)
- [joss_notify](https://github.com/joss-language/joss_notify)
- [joss_backup](https://github.com/joss-language/joss_backup)
- [joss_bg_remover](https://github.com/joss-language/joss_bg_remover)
- [joss_brevo](https://github.com/joss-language/joss_brevo)
