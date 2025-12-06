package files

import "path/filepath"

func GetRoutesFiles(path string) map[string]string {
	return map[string]string{
		filepath.Join(path, "routes.joss"): `// Web Routes
// Rutas Públicas
Router::get("/", "HomeController@index")

// Rutas de Autenticación (Solo invitados)
Router::middleware("guest")
Router::match("GET|POST", "/login", "AuthController@showLogin@doLogin")
Router::match("GET|POST", "/register", "AuthController@showRegister@doRegister")
Router::get("/verify/{token}", "AuthController@verify")
Router::end()

// Rutas Protegidas (Solo autenticados)
Router::middleware("auth")
    Router::get("/dashboard", "DashboardController@index")
    Router::get("/profile", "ProfileController@index")
    Router::get("/logout", "AuthController@logout")
Router::end()
`,
		filepath.Join(path, "api.joss"): `// API Routes
Router::api() // Enable API headers

// Public Routes
Router::post("/api/register", "ApiController@register")
Router::post("/api/login", "ApiController@login")

// Protected Routes (JWT)
Router::middleware("auth_api")
    Router::post("/api/refresh", "ApiController@refresh")
    Router::delete("/api/delete", "ApiController@delete")
Router::end()
`,
	}
}
