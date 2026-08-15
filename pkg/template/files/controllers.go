package files

import "path/filepath"

func GetControllerFiles(path string) map[string]string {
	return map[string]string{
		filepath.Join(path, "app", "controllers", "auth", "ProfileController.joss"): `class ProfileController {
    func index() {
        $u = Auth::user()
        $userId = Auth::id()

        $prefix = env("PREFIX", "js_")
        $mfa = new MfaManager()
        $mfa->setPrefix($prefix)
        
        $hasTOTP = $mfa->hasTOTP($userId)
        $qrCode = ""
        
        (!$hasTOTP) ? {
            $qrCode = $mfa->generateTOTP($userId, $u->email)
        } : {}

        return view("profile.index", {
            "title":       "Mi Perfil",
            "first_name":  $u->first_name,
            "last_name":   $u->last_name,
            "email":       $u->email,
            "phone":       $u->phone,
            "role_id":     $u->role_id,
            "username":    $u->username,
            "mfa_enabled": $hasTOTP,
            "qr_code":     $qrCode,
            "success":     session("success"),
            "error":       session("error")
        })
    }

    func update() {
        $id = Auth::user()->id
        
        $data = {
            "first_name": request("first_name"),
            "last_name":  request("last_name"),
            "phone":      request("phone"),
            "password":   request("password")
        }

        // Auth::update returns true/false
        $success = Auth::update($id, $data)

        return ($success) ? redirect("/profile")->with("success", "Perfil actualizado correctamente.") : back()->with("error", "Error al actualizar el perfil.")
    }

    func activate2FA() {
        $code = request("code")
        $prefix = env("PREFIX", "js_")
        $mfa = new MfaManager()
        $mfa->setPrefix($prefix)
        
        $success = $mfa->verifyAndActivateTOTP(Auth::id(), $code)
        return ($success) ? redirect("/profile")->with("success", "Autenticación de dos factores (2FA) activada con éxito.") : redirect("/profile")->with("error", "Código de verificación inválido.")
    }

    func deactivate2FA() {
        $prefix = env("PREFIX", "js_")
        $mfa = new MfaManager()
        $mfa->setPrefix($prefix)
        
        $success = $mfa->deactivateTOTP(Auth::id())
        return ($success) ? redirect("/profile")->with("success", "Autenticación de dos factores (2FA) desactivada.") : redirect("/profile")->with("error", "Error al desactivar la autenticación de dos factores.")
    }

    func delete() {
        $id = Auth::user()->id
        
        // Remove account
        $success = Auth::delete($id)

        return ($success) ? {
            Auth::logout()
            return redirect("/login")->with("success", "Tu cuenta ha sido eliminada permanentemente.")
        } : {
            return back()->with("error", "Error al eliminar la cuenta.")
        }
    }
}`,

		filepath.Join(path, "app", "controllers", "web", "HomeController.joss"): `class HomeController {
    func index() {
        return view("welcome", {
            "title": "Bienvenido a Joss",
            "version": JOSS_VERSION
        })
    }
}`,

		filepath.Join(path, "app", "controllers", "auth", "AuthController.joss"): `class AuthController {
    func showLogin() {
        (!Auth::guest()) ? { return redirect("/dashboard") } : {}
        return view("auth.login", {"title": "Iniciar Sesión"})
    }
    
    func showRegister() {
        (!Auth::guest()) ? { return redirect("/dashboard") } : {}
        return view("auth.register", {"title": "Crear Cuenta"})
    }
    
    func doLogin() {
        $email = request("email")
        $password = request("password")
        
        // Auth::attempt checks credentials and verification
        $acceso = Auth::attempt($email, $password)
        
        return ($acceso) ? {
            return redirect("/dashboard")->withCookie("joss_token", $acceso)
        } : {
            // Check if unverified and resend
            $newToken = Auth::resendVerification($email)
            
            return ($newToken && $newToken != "already_verified") ? {
                 $link = Request::root() . "/verify/" . $newToken
                 $body = "<h1>Verifica tu cuenta</h1><p>Hemos detectado un intento de inicio de sesión, pero tu cuenta no está verificada. Haz click aquí:</p><a href='" . $link . "'>Verificar Cuenta</a>"
                 
                 SmtpClient::send($email, "Verifica tu cuenta", $body)
                 
                 return back()->with("error", "Cuenta no verificada. Se ha enviado un nuevo correo de verificación.")
            } : {
                 return back()->with("error", "Credenciales inválidas o cuenta no verificada.")
            }
        }
    }

    func doRegister() {
        $data = {
            "first_name": request("first_name"),
            "last_name":  request("last_name"),
            "username":   request("username"),
            "email":      request("email"),
            "password":   request("password"),
            "phone":      request("phone")
        }
        
        // Create user - returns token on success, false on failure
        $token = Auth::create($data)
        
        return ($token) ? {
            // Send Verification Email
            $link = Request::root() . "/verify/" . $token
            $body = "<h1>Bienvenido a Joss</h1><p>Por favor verifica tu cuenta haciendo click en el siguiente enlace:</p><a href='" . $link . "'>Verificar Cuenta</a>"
            
            SmtpClient::send($data["email"], "Verifica tu cuenta", $body)
            
            return redirect("/login")->with("success", "Cuenta creada. Por favor verifica tu correo (revisa spam).")
        } : {
            return back()->with("error", "Error al crear la cuenta.")
        }
    }

    func verify($token) {
        $verified = Auth::verify($token)
        return ($verified) ? {
            return redirect("/login")->with("success", "Cuenta verificada exitosamente. Ya puedes iniciar sesión.")
        } : {
            return redirect("/login")->with("error", "Token de verificación inválido o expirado.")
        }
    }

    func logout() {
        Auth::logout()
        return redirect("/login")->withCookie("joss_token", "")
    }
    
    // API JWT Login
    func apiLogin() {
        $email = request("email")
        $password = request("password")
        
        $token = Auth::attempt($email, $password)
        
        return ($token) ? {
            return json({
                "status": "success",
                "token": $token,
                "user": Auth::user()
            })
        } : {
            return json({
                "status": "error",
                "message": "Invalid credentials"
            }, 401)
        }
    }
}`,

		filepath.Join(path, "app", "controllers", "api", "ApiController.joss"): `class ApiController {
    func register() {
        $data = {
            "first_name": request("first_name"),
            "last_name":  request("last_name"),
            "username":   request("username"),
            "email":      request("email"),
            "password":   request("password"),
            "phone":      request("phone")
        }
        
        $token = Auth::create($data)
        
        return ($token) ? {
            return json({
                "status": "success",
                "message": "User created successfully",
                "token": $token
            }, 201)
        } : {
            return json({
                "status": "error",
                "message": "Registration failed"
            }, 400)
        }
    }

    func login() {
        $email = request("email")
        $password = request("password")
        
        $token = Auth::attempt($email, $password)
        
        return ($token) ? {
            return json({
                "status": "success",
                "token": $token,
                "user": Auth::user()
            })
        } : {
            return json({
                "status": "error",
                "message": "Invalid credentials or not verified"
            }, 401)
        }
    }

    func refresh() {
        $user = Auth::user()
        return ($user) ? {
            $newToken = Auth::refresh($user->id)
            return json({
                "status": "success",
                "token": $newToken
            })
        } : {
            return json({"error": "Unauthorized"}, 401)
        }
    }

    func delete() {
        $user = Auth::user()
        return ($user) ? {
            $deleted = Auth::delete($user->id)
            return ($deleted) ? {
                 return json({"status": "success", "message": "User deleted"})
            } : {
                 return json({"error": "Failed to delete"}, 500)
            }
        } : {
            return json({"error": "Unauthorized"}, 401)
        }
    }

    func forgotPassword() {
        $email = request("email")
        $token = Auth::forgotPassword($email)
        
        return ($token) ? {
            $link = Request::root() . "/password/reset?token=" . $token
            $body = "<h1>Recuperar Contraseña</h1><p>Has solicitado restablecer tu contraseña. Haz click aquí:</p><a href='" . $link . "'>Restablecer Contraseña</a>"
            SmtpClient::send($email, "Recuperar Contraseña", $body)

            return json({
                "status": "success",
                "message": "Si el correo existe, recibirás un enlace de recuperación."
            })
        } : {
             return json({
                "status": "success",
                "message": "Si el correo existe, recibirás un enlace de recuperación."
            })
        }
    }

    func resetPassword() {
        $token = request("token")
        $password = request("password")

        $result = Auth::resetPassword($token, $password)

        return ($result == true) ? {
            return json({
                "status": "success",
                "message": "Contraseña restablecida correctamente"
            })
        } : {
            return json({
                "status": "error",
                "message": "Error al restablecer: " . $result
            }, 400)
        }
    }
}`,

		filepath.Join(path, "app", "controllers", "web", "DashboardController.joss"): `class DashboardController {
    func index() {
        $check = Auth::check()
        (!$check) ? {
            return redirect("/login")->with("error", "Debes iniciar sesión para ver esta página.")
        } : {}

        $isAdmin = Auth::hasRole("admin")
        $roleName = ($isAdmin) ? "Administrador" : "Cliente"
        $u = Auth::user()

        return view("dashboard.index", {
            "title":      "Dashboard",
            "user_name":  $u->name,
            "user_email": $u->email,
            "role":       $roleName,
            "isAdmin":    $isAdmin
        })
    }
}`,

		filepath.Join(path, "app", "controllers", "auth", "PasswordController.joss"): `class PasswordController {
    func showForgot() {
        return view("auth.forgot", { "title": "Recuperar Contraseña" })
    }

    func sendResetLink() {
        $email = request("email")
        $token = Auth::forgotPassword($email)
        
        return ($token) ? {
            $link = Request::root() . "/password/reset?token=" . $token
            $body = "<h1>Recuperar Contraseña</h1><p>Has solicitado restablecer tu contraseña. Haz click aquí:</p><a href='" . $link . "'>Restablecer Contraseña</a>"
            
            SmtpClient::send($email, "Recuperación de Contraseña", $body)

            return view("auth.forgot", { 
                "success": "Se ha enviado un enlace de recuperación a tu correo."
            })
        } : {
            return view("auth.forgot", { "error": "No se pudo generar el token. Verifica el email." })
        }
    }

    func showReset() {
        $token = request("token")
        return view("auth.reset", { "token": $token, "title": "Nueva Contraseña" })
    }

    func resetPassword() {
        $token = request("token")
        $password = request("password")
        
        $result = Auth::resetPassword($token, $password)
        
        return ($result == true) ? {
            return redirect("/login")->withCookie("flash", "Contraseña restablecida correctamente")
        } : {
            return view("auth.reset", { 
                "token": $token, 
                "error": "Error al restablecer: " . $result 
            })
        }
    }
}`,
	}
}
