# Sistema de Vistas

JosSecurity incorpora un motor de plantillas potente y seguro, inspirado en Blade, que permite separar la lógica de presentación del código de la aplicación.

**Ubicación**: `app/views/`
**Extensión**: `.joss.html`

## Sintaxis Básica

### Mostrar Variables
Para imprimir variables, usa las llaves dobles. El contenido es escapado automáticamente para prevenir XSS.

```html
<h1>Hola, {{ $name }}</h1>
<p>Tu saldo es: {{ $balance }}</p>
```

### Comentarios
```html
{{-- Este comentario no será visible en el HTML final --}}
```

## Estructuras de Control

### Condicionales (Ternarios)
Dado que Joss no usa `if` tradicionales, las vistas aprovechan el soporte de evaluación de expresiones dentro de las etiquetas `{{ }}`.

```html
<!-- Ejemplo simple -->
<span>Estado: {{ $isActive ? "Activo" : "Inactivo" }}</span>

<!-- Clases condicionales -->
<div class="{{ $hasError ? 'bg-red-500' : 'bg-green-500' }}">
    {{ $message }}
</div>
```

### Loops (@foreach)
Itera sobre arrays o colecciones de objetos.

```html
<ul>
    @foreach($users as $user)
        <li>{{ $user.name }} - {{ $user.email }}</li>
    @endforeach
</ul>
```

## Herencia de Plantillas (Layouts)

El sistema permite definir "Layouts" maestros y extenderlos en vistas individuales.

### 1. Definir el Layout (`layouts/main.joss.html`)
Usa `@yield` para definir secciones que serán rellenadas por las vistas hijas.

```html
<!DOCTYPE html>
<html>
<head>
    <title>Mi App - @yield('title')</title>
</head>
<body>
    <nav>...</nav>

    <div class="container">
        @yield('content')
    </div>

    <footer>...</footer>
</body>
</html>
```

### 2. Extender el Layout (`home.joss.html`)
Usa `@extends` al inicio y `@section`...`@endsection` para inyectar contenido.

```html
@extends('layouts.main')

@section('title')
    Página de Inicio
@endsection

@section('content')
    <h1>Bienvenido</h1>
    <p>Este es el contenido principal.</p>
@endsection
```

### 3. Inclusión de Vistas Parciales (@include)
Puedes reutilizar fragmentos de código (como menús, pies de página, alertas) usando `@include`.

```html
<!-- Cargar app/views/partials/menu.joss.html -->
@include('partials.menu')
```

## Renderizar Vistas

Desde un controlador:

```javascript
// Carga 'app/views/home.joss.html'
return View::render("home", {"name": "Jose"})

// Carga 'app/views/auth/login.joss.html' con notación de punto
return View::render("auth.login")
```
