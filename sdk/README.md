# SDK de Desarrollo de Plugins Joss (JP v2)

Joss soporta dos métodos para el desarrollo y compilación de plugins:

---

## 1. Compilación Multilenguaje Directa a Bytecode Joss (Recomendado)

Con el compilador integrado de Joss, puedes escribir plugins en **Java, Python, PHP, C/C++ o Rust/Wasm** y compilarlos a **Bytecode nativo de Joss (`.jp` / JPBC)**.

```bash
# Compilar código Python a Bytecode Joss
joss plugin compile script.py --lang=python --name=mi_plugin --exports=calcular

# Compilar clases Java a Bytecode Joss
joss plugin compile App.class --lang=java --name=mi_plugin --exports=procesar

# Compilar módulo Rust / C / C++ (Wasm) a Bytecode Joss
joss plugin compile modulo.wasm --lang=rust --name=mi_plugin --exports=hashear
```

### Ventajas:
- **Cero subprocesos y cero dependencias**: El usuario final no necesita tener instalado Java, Python, PHP ni compiladores.
- **Tree-Shaking Automático**: Solo empaqueta el código y funciones exportadas.
- **Firma Ed25519 Automática**: Paquete ZIP JP v2 criptográficamente verificado.
- **Acceso a `r.Env`**: Acceso transparente a variables de entorno del proyecto.

---

## 2. Desarrollo Nativo en Joss

Los plugins nativos oficiales se escriben directamente en Joss (`src/plugin.joss`) y se empaquetan con:

```bash
joss plugin compile .
```

Genera un artefacto `.jp` con el AST compilado (`JOSSBC2Z`) y la tabla de símbolos en `META-INF/joss-symbols.json`.

---

Consulte [PLUGINS.md](../docs/PLUGINS.md) para más detalles y ejemplos prácticos.

