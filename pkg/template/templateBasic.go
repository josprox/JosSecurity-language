package template

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// CreateBibleProject crea un nuevo proyecto con la estructura de La Biblia de JosSecurity
func CreateBibleProject(path string) {
	fmt.Printf("Creando proyecto JosSecurity (Estructura Biblia) en: %s\n", path)

	// Create directory structure
	dirs := []string{
		filepath.Join(path, "config"),
		filepath.Join(path, "app", "models"),
		filepath.Join(path, "app", "controllers"),
		filepath.Join(path, "app", "views", "layouts"),
		filepath.Join(path, "app", "views", "auth"),
		filepath.Join(path, "app", "views", "dashboard"),
		filepath.Join(path, "app", "libs"),
		filepath.Join(path, "assets", "css"),
		filepath.Join(path, "assets", "js"),
		filepath.Join(path, "assets", "images"),
	}

	for _, dir := range dirs {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			fmt.Printf("Error creando directorio %s: %v\n", dir, err)
			return
		}
	}

	// Create files
	files := map[string]string{
		filepath.Join(path, "main.joss"): `class Main {
    Init main() {
        print("Iniciando Sistema JosSecurity...")
        System.Run("joss", ["server", "start"])
    }
}`,
		filepath.Join(path, "env.joss"): `APP_ENV="development"
PORT="8000"
DB_HOST="localhost"
DB_NAME="joss_db"
DB_USER="root"
DB_PASS=""
JWT_SECRET="change_me_in_production"`,
		filepath.Join(path, "routes.joss"): `// Web Routes
// Rutas Públicas
Router::get("/", "HomeController@index")

// Rutas de Autenticación (Solo invitados)
Router::middleware("guest")
Router::match("GET|POST", "/login", "AuthController@showLogin@doLogin")
Router::match("GET|POST", "/register", "AuthController@showRegister@doRegister")
Router::end()

// Rutas Protegidas (Solo autenticados)
Router::middleware("auth")
Router::get("/dashboard", "DashboardController@index")
Router::get("/logout", "AuthController@logout")
Router::end()
`,
		filepath.Join(path, "config", "reglas.joss"): `// Constantes Globales
const string APP_NAME = "JosSecurity Enterprise"
const string APP_VERSION = "3.0.0"`,
		filepath.Join(path, "app", "controllers", "HomeController.joss"): `class HomeController {
    func index() {
        return View::render("welcome", {
            "title": "Bienvenido a JosSecurity",
            "version": "3.0"
        })
    }
}`,
		filepath.Join(path, "app", "controllers", "AuthController.joss"): `class AuthController {
    func showLogin() {
        // Redirect if already logged in
        (!Auth::guest()) ? {
            return Response::redirect("/dashboard")
        } : {}
        return View::render("auth.login", {"title": "Iniciar Sesión"})
    }
    
    func showRegister() {
        // Redirect if already logged in
        (!Auth::guest()) ? {
            return Response::redirect("/dashboard")
        } : {}
        return View::render("auth.register", {"title": "Crear Cuenta"})
    }
    
    func doLogin() {
        $email = Request::input("email")
        $password = Request::input("password")
        var $acceso = Auth::attempt($email, $password)
        
        ($acceso) ? {
            return Response::redirect("/dashboard")
        } : {
            return Response::back()->with("error", "Credenciales inválidas")
        }
    }

    func doRegister() {
        $name = Request::input("name")
        $email = Request::input("email")
        $password = Request::input("password")
        
        // Auth::create([email, password, name, role_id])
        // Default role is 2 (Client), pass 1 for Admin (manually)
        var $creado = Auth::create([$email, $password, $name])
        
        ($creado) ? {
            // Auto login after register
            Auth::attempt($email, $password)
            return Response::redirect("/dashboard")->with("success", "Cuenta creada exitosamente.")
        } : {
            return Response::back()->with("error", "Error al crear la cuenta. El correo podría estar en uso.")
        }
    }

    func logout() {
        Auth::logout()
        return Response::redirect("/login")
    }
}`,
		filepath.Join(path, "app", "controllers", "DashboardController.joss"): `class DashboardController {
    func index() {
        // Protect Route
        var $check = Auth::check()
        (!$check) ? {
            return Response::redirect("/login")->with("error", "Debes iniciar sesión para ver esta página.")
        } : {}

        $isAdmin = Auth::hasRole("admin")
        $roleName = ($isAdmin) ? "Administrador" : "Cliente"

        return View::render("dashboard.index", {
            "title": "Dashboard",
            "user": Auth::user(),
            "role": $roleName,
            "isAdmin": $isAdmin
        })
    }
}`,
		filepath.Join(path, "app", "models", "User.joss"): `class User extends GranMySQL {
    Init constructor() {
        $this->tabla = "users"
    }
}`,
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
    <link rel="stylesheet" href="/assets/css/app.css">
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
		filepath.Join(path, "assets", "css", "app.css"): `:root {
    --primary: #2563eb;
    --primary-dark: #1e40af;
    --secondary: #64748b;
    --background: #f8fafc;
    --surface: #ffffff;
    --text: #1e293b;
    --text-light: #64748b;
    --border: #e2e8f0;
    --danger: #ef4444;
    --success: #22c55e;
    --info: #3b82f6;
}

body {
    font-family: 'Inter', sans-serif;
    background-color: var(--background);
    color: var(--text);
    margin: 0;
    line-height: 1.6;
}

.navbar {
    background-color: var(--surface);
    border-bottom: 1px solid var(--border);
    padding: 1rem 0;
    margin-bottom: 2rem;
}

.container-nav {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 1rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.brand {
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--primary);
    text-decoration: none;
}

.nav-links a {
    margin-left: 1.5rem;
    text-decoration: none;
    color: var(--text);
    font-weight: 500;
    transition: color 0.2s;
}

.nav-links a:hover {
    color: var(--primary);
}

.container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 1rem;
    min-height: 80vh;
}

.card {
    background: var(--surface);
    border-radius: 0.75rem;
    box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.1);
    margin-bottom: 2rem;
    overflow: hidden;
}

.card-header {
    padding: 1.5rem;
    border-bottom: 1px solid var(--border);
    background: #f1f5f9;
}

.card-header h2 {
    margin: 0;
    font-size: 1.25rem;
}

.card-body {
    padding: 2rem;
}

.card-footer {
    padding: 1rem 2rem;
    background: #f8fafc;
    border-top: 1px solid var(--border);
}

.btn {
    display: inline-block;
    padding: 0.75rem 1.5rem;
    border-radius: 0.5rem;
    font-weight: 600;
    text-decoration: none;
    cursor: pointer;
    transition: all 0.2s;
    border: none;
}

.btn-primary {
    background-color: var(--primary);
    color: white;
}

.btn-primary:hover {
    background-color: var(--primary-dark);
}

.btn-outline-light {
    border: 2px solid var(--primary);
    color: var(--primary);
    background: transparent;
}

.btn-outline-light:hover {
    background: var(--primary);
    color: white;
}

.btn-outline-danger {
    border: 1px solid var(--danger);
    color: var(--danger);
    background: transparent;
    padding: 0.5rem 1rem;
}

.btn-outline-danger:hover {
    background: var(--danger);
    color: white;
}

.btn-block {
    display: block;
    width: 100%;
    text-align: center;
}

.form-group {
    margin-bottom: 1.5rem;
}

.form-group label {
    display: block;
    margin-bottom: 0.5rem;
    font-weight: 500;
}

.form-control {
    width: 100%;
    padding: 0.75rem;
    border: 1px solid var(--border);
    border-radius: 0.5rem;
    font-family: inherit;
    font-size: 1rem;
    box-sizing: border-box;
}

.form-control:focus {
    outline: none;
    border-color: var(--primary);
    box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
}

.alert {
    padding: 1rem;
    border-radius: 0.5rem;
    margin-bottom: 1.5rem;
}

.alert-danger {
    background-color: #fef2f2;
    color: #991b1b;
    border: 1px solid #fecaca;
}

.alert-success {
    background-color: #f0fdf4;
    color: #166534;
    border: 1px solid #bbf7d0;
}

.alert-info {
    background-color: #eff6ff;
    color: #1e40af;
    border: 1px solid #dbeafe;
}

.badge {
    padding: 0.25rem 0.75rem;
    border-radius: 9999px;
    font-size: 0.875rem;
    font-weight: 600;
}

.badge-info {
    background-color: #e0f2fe;
    color: #0369a1;
}

.text-center { text-align: center; }
.d-flex { display: flex; }
.justify-content-center { justify-content: center; }
.justify-content-between { justify-content: space-between; }
.align-items-center { align-items: center; }
.gap-3 { gap: 1rem; }
.mt-4 { margin-top: 1.5rem; }
.mb-4 { margin-bottom: 1.5rem; }
.mb-5 { margin-bottom: 3rem; }

.stat-card {
    background: #f8fafc;
    padding: 1.5rem;
    border-radius: 0.5rem;
    text-align: center;
    border: 1px solid var(--border);
}

.stat-number {
    font-size: 2.5rem;
    font-weight: 700;
    color: var(--primary);
    margin: 0.5rem 0 0;
}

.footer {
    text-align: center;
    padding: 2rem 0;
    color: var(--text-light);
    border-top: 1px solid var(--border);
    margin-top: auto;
}
`,
		filepath.Join(path, "assets", "js", "app.js"): `console.log('JosSecurity Enterprise v3.0 - Inicializado');`,
	}

	for file, content := range files {
		err := ioutil.WriteFile(file, []byte(content), 0644)
		if err != nil {
			fmt.Printf("Error creando archivo %s: %v\n", file, err)
		}
	}

	fmt.Println("\n✓ Proyecto creado exitosamente")
	fmt.Printf("  cd %s\n", path)
	fmt.Println("  joss server start")
}
