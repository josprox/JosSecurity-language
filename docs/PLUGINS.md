# Plugins y JP v2

Joss carga automáticamente las dependencias de `joss.yaml`; el código consumidor usa las clases del plugin sin `use`.

```bash
joss pub add mi_plugin ^1.2.0
joss pub install
```

```yaml
dependencies:
  mi_plugin: "^1.2.0"
```

`use mi_plugin;` solo se conserva para compatibilidad.

## Crear y compilar

```bash
joss new package mi_plugin
cd mi_plugin
joss build package .
joss package inspect mi_plugin.jp
```

## Compilación de Plugins Multilenguaje a Bytecode Joss (.jp / JPBC)

Joss incluye un compilador de plugins (`joss plugin compile`) que permite desarrollar en **Java, Python, PHP, C/C++, Rust, Kotlin, Dart o Flutter** y compilar el código a **Bytecode nativo de Joss (JPBC / Joss Plugin IR)**.

El paquete resultante `.jp` es minificado (< 1 MB mediante Tree Shaking), firmado criptográficamente con Ed25519 y **no requiere que el usuario final tenga instalado Java, CPython, PHP, Maven o Gradle**.

```bash
# Compilar plugin desarrollado en Java (.class / .jar)
joss plugin compile MiPlugin.jar --lang=java --name=music-plugin --exports=searchSong,getSong

# Compilar plugin desarrollado en Python (.py)
joss plugin compile script.py --lang=python --name=tax-plugin --exports=calculate_tax

# Compilar plugin desarrollado en Rust / C++ (WebAssembly .wasm)
joss plugin compile module.wasm --lang=rust --name=crypto-plugin --exports=encrypt_payload --permissions=filesystem.read

# Inspeccionar el bytecode y permisos del plugin
joss plugin inspect music-plugin.jp

# Instalar localmente el plugin compilado
joss plugin install music-plugin.jp
```

## Payload RPC autocontenido

`native` declara un ejecutable por target. El protocolo `joss-rpc-v1` intercambia JSON por stdin/stdout y admite `Plugin::call()` y `Plugin::stream()`.

```yaml
native:
  protocol: joss-rpc-v1
  windows-amd64: native/windows-amd64/bridge.exe
  linux-amd64: native/linux-amd64/bridge
  darwin-arm64: native/darwin-arm64/bridge
```

El SDK contiene adaptadores para C/C++, Python, PHP, Java, Kotlin, Dart/Flutter y Rust. Si el autor incluye legalmente el runtime y todas sus bibliotecas redistribuibles, el consumidor solo necesita Joss y el JP.

Los sidecars no heredan automáticamente `DB_PASS`, tokens ni otras claves de `env.joss`. Reciben variables básicas del sistema y:

```env
PLUGIN_TIMEOUT_SECONDS="30"
PLUGIN_ENV_ALLOW="PUBLIC_API_KEY,LOCALE"
```

## ABI C v1 en memoria

`abi` declara una DLL, SO o dylib por target. Joss la carga dentro del proceso, sin serializar a un proceso externo. No combines `native` y `abi` en el mismo plugin.

```yaml
abi:
  windows-amd64: native/windows-amd64/math.dll
  linux-amd64: native/linux-amd64/libmath.so
  darwin-arm64: native/darwin-arm64/libmath.dylib
```

El encabezado listo para usar está en `sdk/c/joss_driver.h`. Su contrato es:

```c
#ifdef __cplusplus
extern "C" {
#endif

// args_json es un array JSON. El resultado debe ser JSON terminado en NUL.
const char *joss_driver_call(const char *method, const char *args_json);

// Opcional. Joss lo invoca después de copiar el resultado.
void joss_driver_free(const char *result);

#ifdef __cplusplus
}
#endif
```

Desde Joss, `Plugin::call("mi_plugin", "sum", [10, 20])` usa automáticamente ABI o RPC según el paquete. `System::load_driver($path, $name)` y `System::driver_call($name, $method, $args)` permiten usar el mismo contrato fuera de un JP. Streaming corresponde a RPC; ABI v1 usa llamadas completas.

## Validación y confianza

El empaquetador comprueba targets y archivos. Para ejecutables PE, ELF y Mach-O inspecciona imports y falla cuando una biblioteca no perteneciente al sistema no está incluida. Esto evita publicar accidentalmente un puente que todavía depende de una instalación local.

La firma prueba que el contenido corresponde a la llave indicada. El código nativo sigue siendo código de confianza con los permisos de la cuenta que ejecuta Joss. Instala únicamente autores y registros confiables; la firma no sustituye una auditoría ni concede derechos de redistribución.

Plugins oficiales: [joss_ai](https://github.com/josprox/joss_ai), [joss_smtp](https://github.com/josprox/joss_smtp), [joss_notify](https://github.com/josprox/joss_notify) y [joss_backup](https://github.com/josprox/joss_backup).
