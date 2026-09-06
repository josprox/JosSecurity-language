# Joss

Joss es un lenguaje implementado en Go, con herramientas para crear programas
de consola y aplicaciones web. Incluye análisis de tipos antes de ejecutar,
funciones, clases, tareas asíncronas, servidor HTTP, vistas y bases de datos.

No necesitas haber programado para empezar: sigue [Aprender Joss](docs/README.md).

## Instalar

Descarga el runtime para tu plataforma desde las
[distribuciones](https://github.com/josprox/Joss-language/releases), o sigue la
[instalación paso a paso](docs/PRIMEROS_PASOS.md).

```bash
joss version
```

Para construir esta revisión desde sus fuentes, usa la versión de Go de
[go.mod](go.mod) y ejecuta `go build -o joss ./cmd/joss`. En Windows usa
`go build -o joss.exe ./cmd/joss`.

## Tu primer programa

Guarda esto como `hola.joss` en una carpeta nueva:

<!-- joss-run: ["Hola, Joss!"] -->
```joss
print("Hola, Joss!")
```

```bash
joss analyze hola.joss
joss run hola.joss
```

La salida del programa es `Hola, Joss!`. `print` muestra texto; `analyze`
revisa el código y `run` lo revisa y ejecuta. Los warnings no bloquean.

## Una pequeña decisión

Una variable guarda un valor con nombre. Una función reúne instrucciones que
podemos reutilizar. Aquí `saludar` recibe un nombre y devuelve un mensaje:

<!-- joss-run: ["Hola, Ada", "Puedes participar"] -->
```joss
public func saludar(string $nombre): string {
    return "Hola, " . $nombre
}
$edad = 20
print(saludar("Ada"))
print(($edad >= 18) ? "Puedes participar" : "Aún debes esperar")
```

Joss usa `condición ? resultado : alternativa` para elegir; no tiene
`if/else`. La primera asignación fija el tipo inferido; `mixed` y `let $x`
permiten dinamismo explícito. Los archivos y plugins se cargan automáticamente:
no hay imports fuente.

## Encontrar lo que necesitas

| Quiero… | Leer |
|---|---|
| Aprender desde cero | [Recorrido de aprendizaje](docs/README.md) |
| Consultar una regla | [Sintaxis](docs/SINTAXIS.md), [tipos](docs/SISTEMA_TIPOS.md), [gramática](docs/GRAMATICA.md) |
| Encontrar una API | [Funciones globales](docs/FUNCIONES_GLOBALES.md), [clases nativas](docs/MODULOS_NATIVOS.md) |
| Hacer un programa completo | [Consola](docs/PROYECTO_CONSOLA.md), [web](docs/PROYECTO_WEB.md) |
| Usar herramientas y plugins | [CLI](docs/CLI.md), [plugins](docs/PLUGINS.md) |
| Modificar el lenguaje | [Arquitectura](docs/ARQUITECTURA.md), [contribuir](docs/CONTRIBUIR.md) |
| Conocer sus límites | [Estado](docs/ESTADO_IMPLEMENTACION.md), [auditoría](docs/DOCUMENTATION_AUDIT.md) |

El build de aplicaciones empaqueta un AST serializado y comprimido junto al
intérprete Go. `pkg/vm` es un prototipo separado, no el motor de `joss run`.
No hay ownership ni backend LLVM/Cranelift. La documentación distingue estas
limitaciones de las capacidades implementadas.

Licencia [MIT](LICENSE). Vulnerabilidades: [SECURITY.md](SECURITY.md).
