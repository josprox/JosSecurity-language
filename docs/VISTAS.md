# Vistas

`view("dashboard.index", $data)` (o `View::render`) busca `app/views/dashboard/index.joss.html` y después `.html`.

```html
<h1>{{ $title }}</h1>
<div>{{! $trusted_html }}</div>
```

`{{ expr }}` escapa HTML. `{{! expr }}` inserta salida sin escapar y solo debe recibir contenido confiable. `{{ csrf_field() }}` se transforma en salida raw para generar el input CSRF.

## Layouts e includes

```html
@extends('layouts.master')
@section('content')
    @include('partials.alert')
@endsection
```

El layout usa `@yield('content')`. `@extends` solo se reconoce al inicio lógico de la vista. Los includes se resuelven antes de compilar la plantilla.

## Foreach y condicionales

```html
@foreach($users as $user)
    <p>{{ $user.name }}</p>
    {{ ($user.active) ? { <span>Activo</span> } : { <span>Inactivo</span> } }}
@endforeach

<!-- Soporte para expresiones anidadas y colecciones indexadas -->
@foreach($order["items"] as $it)
    <div>{{ $it["product"]["title"] }} - ${{ $it["unit_price"] }}</div>
@endforeach
```

El compilador procesa recursivamente el cuerpo de cada `@foreach` (soportando expresiones arbitrarias, variables simples, arrays anidados o accesos por propiedad) y los ternarios de bloque pueden interactuar libremente con las variables iteradas. Recordatorio: En Joss `@if`, `@else` y `@endif` no existen; se utilizan ternarios funcionales `($cond) ? { ... } : { ... }`.

La notación `$map.key` dentro de expresiones de vista se traduce a `$map->key`. El evaluador permite leer mapas e instancias con esa forma.

## Directivas y Comentarios en Plantillas

1. **Directiva `@json($data)`**:
   Permite volcar objetos, mapas o arrays en atributos JavaScript de forma segura:
   ```html
   <script>
       const config = @json($appConfig);
   </script>
   ```

2. **Comentarios de Plantilla `{{-- Comentario --}}`**:
   Los bloques de comentarios Blade se eliminan antes de renderizar y no llegan al cliente HTML:
   ```html
   {{-- Este comentario no se muestra en el navegador ni en el código fuente --}}
   ```

## Métodos Nativos de `View`

- **`View::exists("vista.nombre")`**: Retorna `true` o `false` según la existencia del archivo de vista tanto en disco como en el VFS.
- **`View::share("key", $value)`** o **`View::share($map)`**: Comparte variables globales disponibles para todas las vistas renderizadas durante la aplicación (ej. configuraciones del sitio, datos de usuario, branding).
- **`View::render("vista.nombre", $data)`**: Renderiza la vista con los datos proporcionados.

## Datos globales automáticos

El renderizador inyecta automáticamente `auth_check`, `auth_guest`, `auth_user`, `auth_role`, `auth_email`, `csrf_token` y mensajes flash cuando existen.
