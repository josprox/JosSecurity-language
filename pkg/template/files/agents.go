package files

import "path/filepath"

func GetAgentsFiles(path string) map[string]string {
	return map[string]string{
		filepath.Join(path, "AGENTS.md"): `# ⚡ AGENTS.md - Joss Programming Language & Framework Guide

Welcome! This application has been built using **Joss (JosSecurity)** — a high-performance, developer-first programming language and web framework created by Joss Estrada (JOSPROX).

This document provides AI assistants (Antigravity, Cursor, Claude, Copilot, ChatGPT) with essential context, directory conventions, language syntax rules, and standard library helpers.

---

## 🏛️ Project Architecture & Directory Layout

Joss applications use a domain-scoped MVC directory layout:

` + "```" + `
project-root/
├── main.joss                 # Entry point (Starts HTTP server & services)
├── env.joss                  # Environment configuration variables
├── routes.joss               # Web HTTP routes & middleware definitions
├── api.joss                  # REST API endpoints
├── joss.yaml                 # Package manifest & dependencies
├── AGENTS.md                 # AI Agent instructions & language guide
├── config/                   # Global configuration & constants
│   └── reglas.joss
├── app/
│   ├── controllers/          # Domain-scoped HTTP Controllers
│   │   ├── web/              # Web page controllers (HTML/Blade templates)
│   │   ├── auth/             # Authentication & Account controllers
│   │   └── api/              # REST JSON API controllers
│   ├── models/               # GranDB ORM Models
│   │   └── auth/             # User & Auth models
│   ├── services/             # Background services & API integrations
│   ├── middleware/           # Request & Security Middleware
│   ├── database/             # Migrations & Database Seeds
│   └── libs/                 # Custom helper libraries
├── app/views/                # HTML/Template Views
│   ├── layouts/              # Master layouts (@extends, @section)
│   ├── auth/                 # Login & Registration templates
│   └── dashboard/            # Admin dashboard templates
├── assets/                   # CSS, JS, and Documentation assets
│   └── docs/                 # Offline markdown documentation
├── public/                   # Static web root (CSS, JS, images)
└── storage/                  # SQLite DB, uploads & temporary logs
` + "```" + `

---

## 🔤 Joss Language Core Syntax Cheatsheet

### 1. Variables & Data Types
- Variables MUST start with ` + "`$`" + `:
  ` + "```joss" + `
  $name = "Joss Red"
  $count = 42
  $is_active = true
  $items = ["apple", "banana", "orange"]
  $user = { "id": 1, "username": "joss" }
  ` + "```" + `

### 2. Control Flow & Expression Blocks
- Ternary syntax and block evaluation:
  ` + "```joss" + `
  ($user != null) ? {
      print("User found: " . $user["username"])
  } : {
      print("User not found")
  }

  $status = ($age >= 18) ? "Adult" : "Minor"
  ` + "```" + `

### 3. OOP Classes & Methods
- Define classes with ` + "`class`" + `, methods with ` + "`func`" + `, constructors with ` + "`Init constructor()`" + `:
  ` + "```joss" + `
  class UserService {
      $prefix = "usr_"

      func getUserName($user) {
          return $user->first_name . " " . $user->last_name
      }
  }
  
  $service = new UserService()
  $name = $service->getUserName($u)
  ` + "```" + `

### 4. Database Models (GranDB ORM)
- Models extend ` + "`GranDB`" + ` and set ` + "`$tabla`" + `:
  ` + "```joss" + `
  class User extends GranDB {
      Init constructor() {
          $this->tabla = "users"
      }
  }

  // GranDB Queries
  $users = User::all()
  $user = User::find(1)
  $activeUsers = User::where("is_active", 1)->get()
  ` + "```" + `

### 5. Routing (` + "`routes.joss`" + ` & ` + "`api.joss`" + `)
  ` + "```joss" + `
  // Web routes
  Router::get("/", "HomeController@index")

  // Protected routes
  Router::middleware("auth")
      Router::get("/dashboard", "DashboardController@index")
  Router::end()

  // JSON API response
  Router::post("/api/login", "ApiController@login")
  ` + "```" + `

### 6. Standard Native Classes
- **Http Client**: ` + "`Http::get($url, $headers)`" + `, ` + "`Http::json($method, $url, $body, $headers)`" + `
- **HTTP Responses**: ` + "`Response::json($data, $code)`" + `, ` + "`Response::redirect($url)`" + `, ` + "`View::render($view, $data)`" + `
- **Authentication**: ` + "`Auth::user()`" + `, ` + "`Auth::check()`" + `, ` + "`Auth::attempt($email, $password)`" + `, ` + "`Auth::logout()`" + `
- **String Helpers**: ` + "`Str::contains($str, $sub)`" + `, ` + "`Str::upper($str)`" + `, ` + "`Str::lower($str)`" + `
- **File I/O**: ` + "`file_put_contents($path, $content)`" + `, ` + "`file_get_contents($path)`" + `
- **Cron Jobs**: ` + "`Cron::schedule($name, $cronExpr, $closure)`" + `
- **System**: ` + "`System::env($key)`" + `, ` + "`System::log($msg)`" + `, ` + "`System::now()`" + `

---

## 🛠️ CLI Development Commands

- **Start Development Server (with Live Hot Reload)**:
  ` + "```bash" + `
  joss server start
  ` + "```" + `
- **Generate Controllers**:
  ` + "```bash" + `
  joss make:controller web/CatalogController
  joss make:controller api/v1/OrderController
  ` + "```" + `
- **Generate Models**:
  ` + "```bash" + `
  joss make:model shop/Product
  ` + "```" + `
- **Generate Full CRUD**:
  ` + "```bash" + `
  joss make:crud products
  ` + "```" + `
- **Run Migrations**:
  ` + "```bash" + `
  joss migrate
  ` + "```" + `
- **Execute Script**:
  ` + "```bash" + `
  joss run main.joss
  ` + "```" + `

---

## 📚 Official Documentation & Links

- **Official Website**: [https://joss.red](https://joss.red)
- **GitHub Repository**: [https://github.com/joss-language/Joss-Programming-Language](https://github.com/joss-language/Joss-Programming-Language)
- **Offline Markdown Docs**: Check ` + "`assets/docs/`" + ` in this project.
`,
	}
}
