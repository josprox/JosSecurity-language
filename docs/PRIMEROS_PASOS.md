# Primeros pasos: de un archivo a un programa

Antes: [índice](README.md). Después: [valores y variables](FUNDAMENTOS.md).
Consulta rápida: [CLI](CLI.md).

## Qué vas a aprender

Un programa es una secuencia de instrucciones para una computadora. El texto
que escribes es el **código fuente**. En Joss se guarda en archivos terminados
en `.joss`. Puedes escribirlos con un editor de texto; no son documentos de Word.

Joss es el lenguaje. El **runtime**, escrito en Go, ejecuta sus instrucciones.
El comando `joss` reúne el runtime y herramientas para revisar archivos, crear
proyectos y arrancar un servidor. Un servidor espera peticiones de otros
programas, por ejemplo de un navegador. Tu primer programa no necesita uno.

## Instalar

Una **terminal** es una ventana donde escribes comandos. En Windows puedes usar
PowerShell; en Linux o macOS, Terminal. Copia solamente los comandos del interior
de los bloques de esta página, sin sus delimitadores.

Puedes descargar el archivo correspondiente a tu sistema en
[Releases](https://github.com/josprox/Joss-language/releases), descomprimirlo y
poner `joss` en una carpeta del `PATH`. `PATH` es la lista de carpetas donde la
terminal busca programas. Si lo modificas, abre una terminal nueva.

También hay instaladores con menú en el repositorio. En Windows, desde una
terminal PowerShell con los permisos que requiere instalar en Program Files:

```powershell
iwr -useb https://raw.githubusercontent.com/josprox/Joss-language/main/install/remote-install.ps1 | iex
```

En Linux/macOS:

```bash
curl -fsSL https://raw.githubusercontent.com/josprox/Joss-language/main/install/remote-install.sh | bash
```

Descargan runtime y extensión; requieren conexión. Requisitos, destinos y
desinstalación están en [instalación](../install/README.md). La extensión de
VS Code ayuda a editar, pero no es necesaria para ejecutar un archivo.

Comprueba la instalación:

```bash
joss version
```

El número depende de la distribución instalada. Para construir esta revisión
usa la versión de Go de `go.mod`: `go build -o joss ./cmd/joss` (`joss.exe` en
Windows). Desde esa carpeta se invoca como `./joss` o `.\joss.exe` si no está
en `PATH`.

## Escribir y ejecutar

Crea una carpeta vacía llamada `mi-primer-joss`, ábrela en la terminal y guarda
allí `hola.joss` con el siguiente contenido:

<!-- joss-run: ["Hola, Joss!"] -->
```joss
print("Hola, Joss!")
```

`"Hola, Joss!"` es un valor de texto. Las comillas indican dónde empieza y
termina; no se imprimen. `print` recibe ese valor entre paréntesis y lo muestra
seguido de un salto de línea.

```bash
joss analyze hola.joss
joss run hola.joss
```

Salida del programa:

```text
Hola, Joss!
```

La CLI puede mostrar mensajes propios además de esa salida. Primero divide el
texto en piezas, comprueba su estructura y revisa nombres y tipos. Si hay
errores, no ejecuta. Un warning es una observación que permite continuar.

## Hacer un cambio

<!-- joss-run: ["Hola, Ada", "Bienvenida a Joss"] -->
```joss
$nombre = "Ada"
print("Hola, " . $nombre)
print("Bienvenida a Joss")
```

`$nombre` guarda texto para usarlo después. El punto une dos textos. Cambia
`Ada` por tu nombre, guarda y ejecuta de nuevo. Verás dos líneas.

## Si algo falla

| Síntoma | Qué revisar |
|---|---|
| La terminal no reconoce `joss` | Instalación y `PATH`; prueba la ruta al ejecutable. |
| No encuentra `hola.joss` | Carpeta actual y nombre; evita `hola.joss.txt`. |
| Error de sintaxis | Comillas y paréntesis; empieza por la primera línea señalada. |
| Variable no definida | Escribe exactamente el mismo nombre, incluido `$`. |
| Muchos errores ajenos al ejemplo | Usa una carpeta vacía: el analizador también descubre fuentes del proyecto. |

No ejecutes `env.joss`: es configuración. No necesitas `Main`, clases, un
servidor ni una base de datos para este ejemplo. Más adelante aprenderás
la variante de entrada `Main` con `Init main()` de las plantillas.

Ejercicio: imprime tres líneas que formen una pequeña presentación. Después
guarda una palabra en una variable y úsala en dos de esas líneas.
