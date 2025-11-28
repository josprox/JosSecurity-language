package files

import "path/filepath"

func GetViewFiles(path string) map[string]string {
	return map[string]string{
		filepath.Join(path, "app", "views", "welcome.joss.html"): `@extends('layouts.master')

@section('content')
    <div class="text-center">
        <h1 class="display-4 mb-4">{{ $title }}</h1>
        <p class="lead mb-5">La plataforma de seguridad empresarial más avanzada.</p>
        <div class="d-flex justify-content-center gap-3">
            <a href="/login" class="btn btn-primary btn-lg">Iniciar Sesión</a>
            <a href="/register" class="btn btn-outline-light btn-lg">Registrarse</a>
        </div>
        <p class="mt-5 text-muted">Versión {{ $version }}</p>
    </div>
@endsection`,
		filepath.Join(path, "app", "views", "auth", "login.joss.html"): `@extends('layouts.master')

@section('content')
    <div class="card">
        <div class="card-header">
            <h2>{{ $title }}</h2>
        </div>
        <div class="card-body">
            <div class="alert alert-danger" style="display: {{ $error ? 'block' : 'none' }}">
                {{ $error }}
            </div>
            <div class="alert alert-success" style="display: {{ $success ? 'block' : 'none' }}">
                {{ $success }}
            </div>

            <form method="POST" action="/login">
                <div class="form-group">
                    <label>Correo Electrónico</label>
                    <input type="email" name="email" class="form-control" required placeholder="ejemplo@correo.com">
                </div>
                <div class="form-group">
                    <label>Contraseña</label>
                    <input type="password" name="password" class="form-control" required placeholder="******">
                </div>
                <button type="submit" class="btn btn-primary btn-block">Entrar</button>
            </form>
        </div>
        <div class="card-footer text-center">
            <p>¿No tienes cuenta? <a href="/register">Regístrate aquí</a></p>
        </div>
    </div>
@endsection`,
		filepath.Join(path, "app", "views", "auth", "register.joss.html"): `@extends('layouts.master')

@section('content')
    <div class="card">
        <div class="card-header">
            <h2>{{ $title }}</h2>
        </div>
        <div class="card-body">
            <div class="alert alert-danger" style="display: {{ $error ? 'block' : 'none' }}">
                {{ $error }}
            </div>

            <form method="POST" action="/register">
                <div class="form-group">
                    <label>Nombre Completo</label>
                    <input type="text" name="name" class="form-control" required placeholder="Juan Pérez">
                </div>
                <div class="form-group">
                    <label>Correo Electrónico</label>
                    <input type="email" name="email" class="form-control" required placeholder="ejemplo@correo.com">
                </div>
                <div class="form-group">
                    <label>Contraseña</label>
                    <input type="password" name="password" class="form-control" required placeholder="******">
                </div>
                <button type="submit" class="btn btn-primary btn-block">Crear Cuenta</button>
            </form>
        </div>
        <div class="card-footer text-center">
            <p>¿Ya tienes cuenta? <a href="/login">Inicia sesión</a></p>
        </div>
    </div>
@endsection`,
		filepath.Join(path, "app", "views", "dashboard", "index.joss.html"): `@extends('layouts.master')

@section('content')
    <div class="card">
        <div class="card-header d-flex justify-content-between align-items-center">
            <h2>Dashboard</h2>
            <div class="d-flex align-items-center gap-3">
                <span class="badge badge-info">{{ $role }}</span>
                <a href="/logout" class="btn btn-outline-danger btn-sm">Cerrar Sesión</a>
            </div>
        </div>
        <div class="card-body">
            <h3>Bienvenido, {{ $user }}</h3>
            <p>Has iniciado sesión correctamente en el sistema JosSecurity Enterprise.</p>
            
            <div class="alert alert-info" style="display: {{ $isAdmin ? 'block' : 'none' }}">
                <strong>Panel de Administrador:</strong> Tienes acceso total al sistema.
            </div>

            <div class="row mt-4">
                <div class="col-md-4">
                    <div class="stat-card">
                        <h4>Proyectos</h4>
                        <p class="stat-number">12</p>
                    </div>
                </div>
                <div class="col-md-4">
                    <div class="stat-card">
                        <h4>Alertas</h4>
                        <p class="stat-number">3</p>
                    </div>
                </div>
                <div class="col-md-4">
                    <div class="stat-card">
                        <h4>Usuarios</h4>
                        <p class="stat-number">150</p>
                    </div>
                </div>
            </div>
        </div>
    </div>
@endsection`,
		filepath.Join(path, "app", "views", "layouts", "master.joss.html"): `<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{ $title ?? "JosSecurity" }}</title>
    <link rel="stylesheet" href="/public/css/app.css">
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;600;700&display=swap" rel="stylesheet">
</head>
<body>
    <nav class="navbar">
        <div class="container-nav">
            <a href="/" class="brand">JosSecurity</a>
            <div class="nav-links">
                {{ $auth_check ? '<a href="/dashboard">Dashboard</a>' : '' }}
                {{ $auth_guest ? '<a href="/login">Login</a><a href="/register">Registro</a>' : '' }}
            </div>
        </div>
    </nav>

    <div class="container">
        @yield('content')
    </div>

    <footer class="footer">
        <p>&copy; 2024 JosSecurity Enterprise. Todos los derechos reservados.</p>
    </footer>
</body>
</html>`,
	}
}
