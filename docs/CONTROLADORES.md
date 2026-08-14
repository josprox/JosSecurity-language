# Controladores y HTTP

Un controlador es una clase Joss. El dispatcher resuelve `Controller@method` y también closures de ruta.

```joss
class ProductController {
    func index() {
        $products = GranDB::table("products")->get()
        return view("products.index", {"products": $products})
    }

    func store() {
        $name = request("name")
        (empty($name)) ? {
            return json({"error": "El nombre es obligatorio"}, 422)
        } : {}

        GranDB::table("products")->insert({"name": $name})
        return redirect("/products")
    }
}
```
## Subdirectorios de Dominio y Carga Recursiva

Joss escanea y precarga automáticamente todos los controladores organizados dentro del árbol de subcarpetas de `app/controllers/` (por ejemplo, `app/controllers/web/`, `app/controllers/auth/`, `app/controllers/api/`, `app/controllers/admin/`).

Al utilizar el generador CLI:
```bash
joss make:controller admin/DashboardController
```
El motor genera el archivo en `app/controllers/admin/DashboardController.joss` sanitizando el nombre de la clase como `class DashboardController`, manteniendo el código limpio y libre de prefijos de ruta inválidos.

## Rutas

```joss
// Verbos individuales
Router::get("/products", "ProductController@index")
Router::post("/products", "ProductController@store")
Router::put("/products/{id}", "ProductController@update")
Router::patch("/products/{id}", "ProductController@patch")
Router::delete("/products/{id}", "ProductController@destroy")
Router::query("/search", "SearchController@query")

// Captura de todos los verbos HTTP (estilo Laravel)
Router::any("/login", "AuthController@showLogin@doLogin")

// Match explícito para múltiples verbos
Router::match("GET|POST", "/contact", "ContactController@show@submit")

// Closures
Router::get("/sound/{id}", func($id) {
    return Redirect::to("https://example.com/" . $id, 302)
})
```

Los parámetros `{name}` se inyectan en handlers HTTP. Las rutas WebSocket también los soportan (`Router::ws`); allí `$ws` es el primer argumento y los parámetros siguen en orden.

## Cliente HTTP Nativo (`Http`)

El lenguaje incluye la clase nativa `Http` de propósito general para realizar peticiones externas:

```joss
// 1. Peticiones directas (GET, POST, PUT, PATCH, DELETE, QUERY, HEAD, OPTIONS)
$body = Http::get("https://api.github.com/zen")
$res = Http::post("https://api.ejemplo.com/item", JSON::stringify({"name": "nuevo"}), {"Authorization": "Bearer TOKEN"})
$queryResult = Http::query("https://api.ejemplo.com/search", JSON::stringify({"filter": "active"}))

// 2. Cliente JSON inteligente (serializa y deserializa datos automáticamente)
$data = Http::json("GET", "https://api.github.com/users/octocat")
$nombre = $data["name"]

// 3. Petición universal hiper-configurable (Http::request)
$response = Http::request("POST", "https://api.ejemplo.com/v1/resource", {
    "query": { "page": "1" },
    "headers": { "Accept": "application/json" },
    "json": { "status": "active" },
    "timeout": 10,
    "follow_redirects": true
})

if ($response["success"]) {
    $code = $response["status"]
    $json = $response["json"]
}
```

## Request

- `input()` y `post()` leen el mapa combinado de la petición.
- `all()` retorna campos públicos; `except([...])` excluye claves.
- `header()`, `cookie()` y `root()` consultan metadatos HTTP.
- `file()` retorna un mapa; el contenido subido está en `content`.

## Response

- `json($data, $status=200)`.
- `error($message, $status=400)` retorna `{"error": ...}`.
- `redirect($url)` y `back()`.
- `raw($content, $status=200, $mime="text/plain", $headers={})`.
- `stream($callback)` para SSE.

Una respuesta admite `->with()`, `->withCookie()`, `->withHeader()` y `->status()`. Para binarios usa `raw`; un string HTML normal puede recibir el script de hot reload durante desarrollo.
