# Proyecto web: saludo dinámico

[Índice](README.md) · Antes: [proyecto de consola](PROYECTO_CONSOLA.md) · Después: [HTTP](CONTROLADORES.md)

Una aplicación web recibe peticiones y devuelve respuestas. En este tutorial una
ruta captura un nombre de la URL, llama a un controlador y renderiza una vista.

## Crear y ejecutar

```sh
joss new web saludo_web
cd saludo_web
joss server start
```

La plantilla genera `main.joss` con una clase `Main` y `Init main()`. Esa entrada
llama a `Server::start()`. El puerto del servidor es `PORT`, con 8000 como valor
runtime predeterminado; la plantilla actual escribe 80 en `env.joss`. Para este
tutorial cambia esa línea a `PORT="8080"` y abre
`http://127.0.0.1:8080/saludo/Ana`.

## Ruta

Añade a `routes.joss`:

<!-- joss-check: fragmento de routes.joss -->
```joss
Router::get("/saludo/{nombre}", "SaludoController@show")
```

`{nombre}` es un segmento dinámico. El dispatcher lo entrega al método en el
mismo orden en que aparece. No lo inserta automáticamente en SQL ni HTML.

## Controlador

Crea `app/controllers/SaludoController.joss`:

<!-- joss-check: fragmento de controlador web -->
```joss
public class SaludoController {
    public func show(string $nombre) {
        $nombreLimpio = trim($nombre)
        return view("saludo", {"nombre": $nombreLimpio})
    }
}
```

El controlador transforma datos y elige la respuesta. Su parámetro tiene tipo
porque todo parámetro fuente debe declararlo. La plantilla escapará el nombre;
`trim` no es una medida de seguridad por sí solo.

## Vista

Crea `app/views/saludo.joss.html`:

```html
<!doctype html>
<html lang="es">
<head><meta charset="utf-8"><title>Saludo</title></head>
<body>
  <h1>Hola, {{ $nombre }}</h1>
</body>
</html>
```

`{{ ... }}` escapa HTML. La forma raw `{{! ... }}` debe reservarse para contenido
que la aplicación ya considera confiable. Reinicia o deja que el watcher recargue,
y visita la URL. Una ruta desconocida debe responder 404.

## Comprobar el proyecto

```sh
joss analyze main.joss
joss check .
```

El análisis incluye la entrada y `app/**/*.joss`. `routes.joss` se carga por la
infraestructura del servidor, pero no forma parte del conjunto `LoadProject`
usado por `joss analyze`; sus errores se detectan al cargar el servidor. Esta
asimetría es una limitación actual.

Para aceptar formularios añade una ruta POST, valida `Request::input`, conserva
CSRF y devuelve errores 422. Sigue con [controladores](CONTROLADORES.md),
[middleware](MIDDLEWARE.md), [vistas](VISTAS.md) y [datos](MODELOS.md).
