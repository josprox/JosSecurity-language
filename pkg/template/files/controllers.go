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
        (!Auth::guest()) ? { return Response::redirect("/dashboard") } : {}
        return View::render("auth.login", {"title": "Iniciar Sesión"})
    }
    
    func showRegister() {
        (!Auth::guest()) ? { return Response::redirect("/dashboard") } : {}
        return View::render("auth.register", {"title": "Crear Cuenta"})
    }
    
    func doLogin() {
        $email = Request::input("email")
        $password = Request::input("password")
        
        // Auth::attempt now checks for verification
        var $acceso = Auth::attempt($email, $password)
        
        ($acceso) ? {
            return Response::redirect("/dashboard")
        } : {
            return Response::back()->with("error", "Credenciales inválidas o cuenta no verificada.")
        }
    }

    func doRegister() {
        $data = {
            "first_name": Request::input("first_name"),
            "last_name": Request::input("last_name"),
            "username": Request::input("username"),
            "email": Request::input("email"),
            "password": Request::input("password"),
            "phone": Request::input("phone")
        }
        
        // Create user - returns token on success, false on failure
        var $token = Auth::create($data)
        
        ($token) ? {
            // Send Verification Email
            // $link = env("APP_URL") + "/verify/" + $token
            $link = Request::root() + "/verify/" + $token
            $body = "<h1>Bienvenido a JosSecurity</h1><p>Por favor verifica tu cuenta haciendo click en el siguiente enlace:</p><a href='" + $link + "'>Verificar Cuenta</a>"
            
            SmtpClient::send($data["email"], "Verifica tu cuenta", $body)
            
            return Response::redirect("/login")->with("success", "Cuenta creada. Por favor verifica tu correo (revisa spam).")
        } : {
            return Response::back()->with("error", "Error al crear la cuenta.")
        }
    }

    func verify($token) {
        $verified = Auth::verify($token)
        ($verified) ? {
            return Response::redirect("/login")->with("success", "Cuenta verificada exitosamente. Ya puedes iniciar sesión.")
        } : {
            return Response::redirect("/login")->with("error", "Token de verificación inválido o expirado.")
        }
    }

    func logout() {
        Auth::logout()
        return Response::redirect("/login")->withCookie("joss_token", "")
    }
    
    // API JWT Login
    func apiLogin() {
        $email = Request::input("email")
        $password = Request::input("password")
        
        var $token = Auth::attempt($email, $password)
        
        ($token) ? {
            return Response::json({
                "status": "success",
                "token": $token,
                "user": Auth::user()
            })
        } : {
            return Response::json({
                "status": "error",
                "message": "Invalid credentials"
            }, 401)
        }
    }
}`,
		filepath.Join(path, "app", "controllers", "ApiController.joss"): `class ApiController {
    func register() {
        $data = {
            "first_name": Request::input("first_name"),
            "last_name": Request::input("last_name"),
            "username": Request::input("username"),
            "email": Request::input("email"),
            "password": Request::input("password"),
            "phone": Request::input("phone")
        }
        
        var $token = Auth::create($data)
        
        ($token) ? {
            // Send verification email logic could go here too
            return Response::json({
                "status": "success",
                "message": "User created successfully",
                "token": $token
            }, 201)
        } : {
            return Response::json({
                "status": "error",
                "message": "Registration failed"
            }, 400)
        }
    }

    func login() {
        $email = Request::input("email")
        $password = Request::input("password")
        
        var $token = Auth::attempt($email, $password)
        
        ($token) ? {
            return Response::json({
                "status": "success",
                "token": $token,
                "user": Auth::user()
            })
        } : {
            return Response::json({
                "status": "error",
                "message": "Invalid credentials or not verified"
            }, 401)
        }
    }

    func refresh() {
        
        $user = Auth::user()
        ($user) ? {
            $newToken = Auth::refresh($user.id)
            return Response::json({
                "status": "success",
                "token": $newToken
            })
        } : {
            return Response::json({"error": "Unauthorized"}, 401)
        }
    }

    func delete() {
        $user = Auth::user()
        ($user) ? {
            $deleted = Auth::delete($user.id)
            ($deleted) ? {
                 return Response::json({"status": "success", "message": "User deleted"})
            } : {
                 return Response::json({"error": "Failed to delete"}, 500)
            }
        } : {
            return Response::json({"error": "Unauthorized"}, 401)
        }
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
