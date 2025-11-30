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
Router::end()

// Rutas Protegidas (Solo autenticados)
Router::middleware("auth")
    Router::get("/dashboard", "DashboardController@index")
    Router::get("/profile", "ProfileController@index")
    Router::get("/logout", "AuthController@logout")
Router::end()
`,
	}
}
