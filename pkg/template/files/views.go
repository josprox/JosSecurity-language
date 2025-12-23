package files

import "path/filepath"

func GetViewFiles(path string) map[string]string {
	return map[string]string{
		filepath.Join(path, "app", "views", "profile", "index.joss.html"): `@extends('layouts.master')

@section('content')
    <div class="card">
        <div class="card-header">
            <h2>Mi Perfil</h2>
        </div>
        <div class="card-body">
            <div class="row">
                <div class="col-md-4 text-center">
                    <img src="https://ui-avatars.com/api/?name={{ $user.name }}&background=2563eb&color=fff&size=128" alt="Avatar" class="rounded-circle mb-3">
                    <h4>{{ $user.name }}</h4>
                    <p class="text-muted">{{ $user.email }}</p>
                    <span class="badge badge-primary">{{ $user.role_id == 1 ? 'Administrador' : 'Usuario' }}</span>
                </div>
                <div class="col-md-8">
                    <h3>Información de la Cuenta</h3>
                    <hr>
                    <form action="/profile/update" method="POST">
                        {{ csrf_field() }}
                        
                        <div class="row">
                            <div class="col-md-6 form-group">
                                <label>Nombre</label>
                                <input type="text" name="first_name" class="form-control" value="{{ $user.first_name }}" required>
                            </div>
                            <div class="col-md-6 form-group">
                                <label>Apellido</label>
                                <input type="text" name="last_name" class="form-control" value="{{ $user.last_name }}" required>
                            </div>
                        </div>

                        <div class="form-group">
                            <label>Teléfono</label>
                            <input type="text" name="phone" class="form-control" value="{{ $user.phone }}">
                        </div>

                        <div class="form-group">
                            <label>Correo Electrónico</label>
                            <input type="email" class="form-control" value="{{ $user.email }}" readonly disabled>
                            <small class="text-muted">El correo no se puede cambiar.</small>
                        </div>

                        <hr>
                        <h5>Cambiar Contraseña</h5>
                        <p class="text-muted small">Deja en blanco para mantener la actual.</p>
                        
                        <div class="form-group">
                            <label>Nueva Contraseña</label>
                            <input type="password" name="password" class="form-control" placeholder="********">
                        </div>

                        <button type="submit" class="btn btn-primary d-block w-100 mt-3">Guardar Cambios</button>
                    </form>

                    <hr class="my-5">

                    <div class="alert alert-danger">
                        <h4>Zona de Peligro</h4>
                        <p>Una vez que elimines tu cuenta, no hay vuelta atrás. Por favor, asegúrate.</p>
                        
                        <form action="/profile/delete" method="POST" onsubmit="return confirm('¿Estás SEGURO de que deseas eliminar tu cuenta permanentemente?');">
                             {{ csrf_field() }}
                             <button type="submit" class="btn btn-danger">Eliminar Cuenta</button>
                        </form>
                    </div>
                </div>
            </div>
        </div>
    </div>
@endsection`,
		filepath.Join(path, "app", "views", "welcome.joss.html"): `@extends('layouts.master')

@section('content')
    <div class="hero-section text-center">
        <h1 class="display-4 mb-4">{{ $title }}</h1>
        <p class="lead mb-5">La plataforma de seguridad empresarial más avanzada.</p>
        
        <div class="d-flex justify-content-center gap-3">
            <a href="/login" class="btn btn-primary btn-lg">Iniciar Sesión</a>
            <a href="/register" class="btn btn-outline-light btn-lg">Registrarse</a>
        </div>

        <div class="mt-5">
            <span class="badge badge-info">Compilador v{{ $version }}</span>
        </div>
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
                {{ csrf_field() }}
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
                {{ csrf_field() }}
                <div class="form-group">
                    <label>Nombre</label>
                    <input type="text" name="first_name" class="form-control" required placeholder="Juan">
                </div>
                <div class="form-group">
                    <label>Apellidos</label>
                    <input type="text" name="last_name" class="form-control" required placeholder="Pérez">
                </div>
                <div class="form-group">
                    <label>Usuario</label>
                    <input type="text" name="username" class="form-control" required placeholder="juanperez">
                </div>
                <div class="form-group">
                    <label>Teléfono</label>
                    <input type="tel" name="phone" class="form-control" placeholder="+1234567890">
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
            <h3>Bienvenido, {{ $user.name }}</h3>
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
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
    <!-- Icons -->
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
</head>
<body>
    <div class="app-container">
        <!-- Sidebar -->
        <aside class="sidebar" id="sidebar">
            <div class="sidebar-header">
                <a href="/" class="brand">
                    <i class="fas fa-shield-alt"></i> JosSecurity
                </a>
                <button class="close-sidebar d-md-none" id="closeSidebar">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            
            <nav class="sidebar-nav">
                <ul>
                    <li><a href="/" class="{{ $current_path == '/' ? 'active' : '' }}"><i class="fas fa-home"></i> Inicio</a></li>
                    
                    {{ ($auth_check) ? {
                        <li class="nav-header">Aplicación</li>
                        <li><a href="/dashboard" class="{{ $current_path == '/dashboard' ? 'active' : '' }}"><i class="fas fa-tachometer-alt"></i> Dashboard</a></li>
                        <!-- Injected Links Here -->
                    } : {} }}

                    <li class="nav-header">Cuenta</li>
                    {{ ($auth_check) ? {
                        <li><a href="/profile"><i class="fas fa-user"></i> Perfil</a></li>
                        <li><a href="/logout" class="text-danger"><i class="fas fa-sign-out-alt"></i> Cerrar Sesión</a></li>
                    } : {
                        <li><a href="/login"><i class="fas fa-sign-in-alt"></i> Iniciar Sesión</a></li>
                        <li><a href="/register"><i class="fas fa-user-plus"></i> Registrarse</a></li>
                    } }}
                </ul>
            </nav>
        </aside>

        <!-- Main Content -->
        <main class="main-content">
            <!-- Top Navbar -->
            <header class="top-navbar">
                <button class="toggle-sidebar" id="toggleSidebar">
                    <i class="fas fa-bars"></i>
                </button>
                
                <div class="navbar-right">
                    {{ ($auth_check) ? {
                        <div class="user-menu">
                            <span class="user-name">{{ $auth_user ?? 'Usuario' }}</span>
                            <img src="https://ui-avatars.com/api/?name={{ $auth_user ?? 'U' }}&background=2563eb&color=fff" alt="Avatar" class="user-avatar">
                        </div>
                    } : {} }}
                </div>
            </header>

            <!-- Page Content -->
            <div class="content-wrapper">
                @yield('content')
            </div>

            <footer class="footer">
                <p>&copy; 2024 JosSecurity Enterprise. <span class="d-none d-sm-inline">Todos los derechos reservados.</span></p>
            </footer>
        </main>
        
        <!-- Overlay for mobile -->
        <div class="sidebar-overlay" id="sidebarOverlay"></div>
    </div>

    <script>
        document.addEventListener('DOMContentLoaded', function() {
            const sidebar = document.getElementById('sidebar');
            const toggleBtn = document.getElementById('toggleSidebar');
            const closeBtn = document.getElementById('closeSidebar');
            const overlay = document.getElementById('sidebarOverlay');

            function toggleMenu() {
                sidebar.classList.toggle('active');
                overlay.classList.toggle('active');
            }

            if(toggleBtn) toggleBtn.addEventListener('click', toggleMenu);
            if(closeBtn) closeBtn.addEventListener('click', toggleMenu);
            if(overlay) overlay.addEventListener('click', toggleMenu);
        });
    </script>
</body>
</html>`,
	}
}
