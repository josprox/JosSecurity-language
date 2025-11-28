- `classext` - Clase con herencia (extends)
- `init` - Método Init
- `tern` - Operador ternario con bloques
- `terni` - Ternario inline
- `foreach` - Bucle foreach
- `print` - Print statement
- `var` - Declarar variable
- `async` - Async task con await
- `asyncfn` - Función async
- `map` - Map literal `{ key: value }`
- `try` - Try-catch block
- `authcreate` - Crear usuario con Auth
- `authlogin` - Login con Auth
- `dbquery` - Query con GranMySQL
- `dbinsert` - Insert con GranMySQL
- `main` - Clase Main completa

### ⚙️ Language Features
- Auto-cierre de brackets `{}`, `[]`, `()`
- Auto-cierre de strings `""`, `''`
- Comentarios con `//` y `/* */`
- Folding de bloques de código
- Indentación automática

### 🎨 Tema Incluido
- **JosSecurity Dark**: Tema oscuro optimizado para JosSecurity

## Instalación

### Desde VSIX (Recomendado)
1. Descarga `joss-language-1.0.0.vsix`
2. En VS Code: `Ctrl+Shift+P` → "Extensions: Install from VSIX"
3. Selecciona el archivo `.vsix`

### Desde Código Fuente
1. Copia la carpeta `vscode-joss` a:
   - Windows: `%USERPROFILE%\.vscode\extensions\`
   - macOS/Linux: `~/.vscode/extensions/`
2. Reinicia VS Code

## Uso

1. Abre cualquier archivo `.joss`
2. El syntax highlighting se aplicará automáticamente
3. Usa snippets escribiendo el prefijo y presionando `Tab`

### Ejemplos de Snippets

**Crear clase:**
```joss
class // Tab
```

**Ternario:**
```joss
tern // Tab
```

**Query DB:**
```joss
dbquery // Tab
```

## Características del Lenguaje

### ✅ Soportado
- ✅ **Concurrencia**: async/await con Goroutines
- ✅ **Smart Numerics**: División automática a float
- ✅ **Maps Nativos**: Sintaxis `{ key: value }`
- ✅ **Autoloading**: Carga automática de clases
- ✅ **Herencia**: extends para clases
- ✅ **Try-Catch**: Manejo de errores
- ✅ Ternarios (NO if/else/switch)
- ✅ Foreach
- ✅ OOP (Clases, Init, Métodos)
- ✅ Variables con prefijo `$`
- ✅ Tipos estáticos (int, float, string, bool, array)
- ✅ Auth con JWT
- ✅ GranMySQL con prefijos
- ✅ Helpers (isset, empty, len, count)

### ❌ NO Soportado (Por Diseño)
- ❌ if/else/switch (Usar ternarios)
- ❌ while/do-while (Usar foreach)

## Configuración Recomendada

Añade a tu `settings.json`:

```json
{
  "[joss]": {
    "editor.tabSize": 4,
    "editor.insertSpaces": true,
    "editor.formatOnSave": false,
    "editor.wordWrap": "on"
  }
}
```

## Ejemplos

### Clase Main
```joss
class Main {
    Init main() {
        print("Hello JosSecurity")
    }
}
```

### Ternario
```joss
($edad >= 18) ? {
    print("Mayor de edad")
} : {
    print("Menor de edad")
}
```

### Auth
```joss
$auth = new Auth()
$token = $auth.attempt("user@example.com", "password")
```

### GranMySQL
```joss
$db = new GranMySQL()
$db.table("users")
$db.where("email", "user@example.com")
$result = $db.get()
```

## Soporte

- GitHub: https://github.com/jossecurity/joss
- Documentación: Ver `docs/` en el repositorio

## Licencia

MIT

---

**Desarrollado para JosSecurity** 🔒
