package files

import "path/filepath"

func GetControllerFiles(path string) map[string]string {
	return map[string]string{
		filepath.Join(path, "app", "controllers", "ProfileController.joss"): `class ProfileController {
    function index() {
        $user = Auth.user()
        return View.render("profile/index", {"user": $user, "title": "Mi Perfil"})
    }
}`,
		filepath.Join(path, "app", "controllers", "HomeController.joss"): `class HomeController {
    func index() {
        return View::render("welcome", {
            "title": "Bienvenido a JosSecurity",
            "version": JOSS_VERSION
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
	}
}
