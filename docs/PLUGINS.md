# Plugins y JP (Joss Plugin Bytecode)

Joss carga automáticamente los paquetes y dependencias declaradas en `joss.yaml`; el código consumidor utiliza las clases y métodos exportados por el plugin sin necesidad de sentencias `use` o `import`.

```bash
joss pub add mi_plugin ^1.0.0
joss pub install
```

```yaml
dependencies:
  mi_plugin: "^1.0.0"
```

---

## 🛠️ Crear y Compilar Plugins

Para crear un nuevo proyecto de plugin oficial:

```bash
joss new plugin mi_plugin
cd mi_plugin
```

Esto generará la estructura limpia del proyecto:
- `joss.yaml`: Manifiesto del paquete.
- `src/plugin.joss`: Implementación orientada a objetos en Joss.
- `.github/workflows/release.yml`: Workflow para compilación automática a `.jp` y publicación en GitHub Releases.

Para compilar el plugin a un paquete binario de Bytecode `.jp`:

```bash
joss plugin compile .
```

---

## ⚡ Compilación Multilenguaje a Bytecode Joss (.jp / JPBC)

Joss cuenta con un motor compilador (`joss plugin compile`) que permite desarrollar componentes en **Java, Python, PHP, C/C++, Rust, Kotlin, Dart o Flutter** y traducirlos automáticamente a **Bytecode nativo de Joss (JPBC)**.

### Características del Sistema:
1. **Tree Shaking Avanzado**: Analiza el Grafo de Llamadas (Call Graph) a partir de las funciones indicadas en `--exports` y elimina automáticamente todo el código muerto, clases e instrucciones no utilizadas.
2. **Cero Dependencias en el Usuario Final**: El paquete `.jp` resultante (< 10 KB a 1 MB) funciona directamente en la máquina virtual de Joss. **El usuario final no requiere tener instalado Java (JVM/JDK), Python, PHP, Node ni compilar nada**.
3. **Firma Criptográfica Ed25519**: Cada paquete `.jp` es firmado criptográficamente de manera transparente durante la compilación para garantizar su integridad.

### Ejemplos de Compilación por Lenguaje:

```bash
# 1. Compilar plugin desarrollado en Java (.class o .jar)
joss plugin compile MiPlugin.jar --lang=java --name=music-plugin --exports=searchSong,getSong

# 2. Compilar plugin desarrollado en Python (.py)
joss plugin compile script.py --lang=python --name=tax-plugin --exports=calculate_tax

# 3. Compilar plugin desarrollado en Rust / C / C++ (WebAssembly .wasm)
joss plugin compile module.wasm --lang=rust --name=crypto-plugin --exports=encrypt_payload --permissions=filesystem.read

# 4. Inspeccionar la firma, funciones exportadas y tamaño de un paquete .jp
joss plugin inspect music-plugin.jp

# 5. Instalar localmente un archivo .jp
joss plugin install music-plugin.jp
```

---

## 🔒 Validación, Seguridad y Confianza

- **Formato `.jp`**: Archivo comprimido con encabezado de firma Ed25519, Bytecode binario optimizado (`bytecode/main.jbc`) e Índice de Símbolos (`META-INF/joss-symbols.json`) para soporte de autocompletado nativo en la extensión de VS Code.
- **Firma digital**: Garantiza que el paquete no ha sido alterado ni manipulado en tránsito.

---

## 📚 Plugins Oficiales

- [joss_ai](https://github.com/joss-language/joss_ai)
- [joss_smtp](https://github.com/joss-language/joss_smtp)
- [joss_notify](https://github.com/joss-language/joss_notify)
- [joss_backup](https://github.com/joss-language/joss_backup)
- [joss_bg_remover](https://github.com/joss-language/joss_bg_remover)

