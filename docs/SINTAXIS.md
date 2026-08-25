# Sintaxis de Joss

## Variables y tipos

Las variables usan `$`. La primera asignación simple infiere un tipo que permanece estable; no convierte la variable en dinámica. `func` es la única palabra para funciones y closures.

```joss
$age = 20       // infiere int
$age = 21       // válido
var $count = 1  // inferencia explícita
let $dynamic = 10 // mixed explícito
int $port = 9000
string $name = "Ada"
bool $active = true
array $items = [1, 2, 3]
$config = {"port": 8000}
```

`$age = "veinte"` es un error porque `$age` ya fue inferida como `int`. Sólo `let $dynamic` permite cambiar de tipo deliberadamente. Los tipos reconocidos incluyen `int`, `float`, `string`, `bool`, `array`, `map`, `object`, `channel` y clases. Una declaración numérica explícita puede convertir una cadena completa y válida antes de fallar; nunca trunca `"20.5"` a `int`.

`const $name = valor` declara una constante inferida y `const Tipo $name = valor` una constante tipada. Las uniones se escriben `int|string`; `int?` es el atajo de `int|null`. Las funciones aceptan retorno opcional (`func name(...): Tipo`) y toda función anotada debe retornar o lanzar en cada ruta demostrable. Consulte [Sistema de tipos](SISTEMA_TIPOS.md).

## Operadores y Concatenación

- **Concatenación de Cadenas (`.`)**: En Joss, la concatenación de cadenas se realiza **estrictamente mediante el operador punto (`.`)**.
  ```joss
  $nombre = "Joss"
  $mensaje = "Hola " . $nombre . ", bienvenido!"
  ```
- **Suma Numérica (`+`)**: El operador `+` se utiliza **exclusivamente para operaciones matemáticas numéricas** (`10 + 5`). Si intentas concatenar texto usando `+`, el runtime arrojará un error explícito solicitando el uso del operador `.`.
- **Coalescencia Nula (`??`) y Elvis (`?:`)**:
  ```joss
  $port = $config["port"] ?? 8000
  $name = $user_name ?: "Anónimo"
  ```


## Funciones, Closures y Tipado de Parámetros

```joss
// Funciones globales con tipado y coerción automática inteligente
func transferir(int $userId, float $monto, string $concepto) {
    // Si $monto entra como string "150.50", Joss lo convierte limpiamente al tipo float
    return "Transferidos $" . $monto . " al usuario #" . $userId
}

$doble = func(int $valor) {
    return $valor * 2
}
```

## Clases, Visibilidad y Herencia

Joss soporta modificadores explícitos de visibilidad (`public`, `private`, `protected`), métodos estáticos (`static`) y herencia entre clases (`extends`):

```joss
public class BaseController {
    protected $db

    func constructor() {
        $this->db = new GranDB()
    }

    protected func respondJson($data, $code = 200) {
        return json($data, $code)
    }
}

// Herencia de clases
public class UserController extends BaseController {
    private string $secret = "sk_prod_123"

    public static func make() {
        return new UserController()
    }

    public func index() {
        $users = $this->db->table("users")->get()
        return $this->respondJson($users)
    }
}
```

Las APIs estáticas usan `Clase::metodo()`. Las instancias usan `$object->method()` y `$object->property`.

## Control de flujo

En Joss **no existen las sentencias `if`, `else`, `elif`, `switch` ni `for`**. El compilador cuenta con un **Syntax Guard** educacional que detecta estas palabras clave foráneas e indica de inmediato la alternativa nativa de Joss:

1. **Operador Ternario**: `$cond ? $a : $b` o bloques expresivos:
```joss
($age >= 18) ? {
    print("Adulto")
} : {
    print("Menor")
}

$label = ($active) ? "activo" : "inactivo"
```

2. **Estructura Match**: Sustituye a `switch` / `if-else` encadenados:
```joss
$message = match ($status) {
    200, 201 => "correcto",
    404 => "no encontrado",
    default => "error"
}
```

3. **Ciclos Nativos**: Sustituyen a `for` mediante `foreach` y `while`:
```joss
foreach ($items as $item) {
    print($item)
}

while ($pending > 0) {
    $pending = $pending - 1
}

do {
    $attempts++
} while ($attempts < 3)
```

## Manejo de Errores y Panic

```joss
// Excepciones estructuradas
try {
    throw "Fallo en la operación"
} catch ($error) {
    print("Error capturado: " . $error)
}

// Aborto irrecuperable
panic("Error crítico de consistencia")
```

`panic()` interrumpe la ejecución del programa con un mensaje de error claro. Para manejo defensivo de errores esperados se utiliza el bloque `try / catch`.

## Helpers Globales y Entorno

Joss provee funciones de primer nivel integradas para operaciones comunes, optimizadas en memoria:

```joss
$port = env("PORT", "9000")
$app = config("APP_ENV", "production")
$home = view("home", {"user": $user})
$responseJson = json({"status": "ok"}, 200)
$red = redirect("/dashboard")
$b = back()->with("error", "Fallo")
$email = request("email")
$userId = session("user_id")
```

### Funciones Estándar de Utilidad:
* **Fechas**: `time()`, `date("Y-m-d H:i:s")`, `strtotime("-2 days")`, `now()`, `microtime(true)`.
* **Cadenas**: `str_contains($str, "foo")`, `str_starts_with`, `str_ends_with`, `str_replace`, `strtolower`, `strtoupper`, `trim`, `substr`, `strpos`, `implode`, `explode`, `md5`, `sha256`, `base64_encode`, `base64_decode`.
* **Arreglos**: `in_array("admin", $roles)`, `array_key_exists("key", $map)`, `array_merge($arr1, $arr2)`, `array_push($arr, $val)`, `array_pop($arr)`, `array_slice($arr, 0, 10)`, `count($arr)`.
* **Archivos**: `file_exists($path)`, `file_get_contents($path)`, `file_put_contents($path, $data)`, `unlink($path)`, `mkdir($dir)`, `is_dir($path)`.
* **Asincronía**: `async { ... }`, `await $future`, `make_chan()`, `send($chan, $val)`, `recv($chan)`.

## GranDB ORM y Bases de Datos Multimotor

GranDB es el ORM fluido nativo de Joss con soporte completo para **SQLite, MySQL, PostgreSQL y Microsoft SQL Server (`sqlserver`/`mssql`)**, con gestión transparente de prefijos de tablas (`DB_PREFIX`):

```joss
// Consultas fluidas
$users = GranDB::table("users")
    ->where("is_active", 1)
    ->whereIn("role_id", [1, 2])
    ->orderBy("created_at", "DESC")
    ->take(10)
    ->get()

// Mutaciones e inserciones
$newId = GranDB::table("products")->insertGetId({
    "name": "Teclado Mecánico",
    "price": 89.90,
    "stock": 25,
    "created_at": now()
})

GranDB::table("products")->where("id", $newId)->increment("stock", 5)

// Transacciones atómicas
GranDB::transaction(func($db) {
    GranDB::table("accounts")->where("id", 1)->decrement("balance", 100)
    GranDB::table("accounts")->where("id", 2)->increment("balance", 100)
})

// Cambio dinámico de motor en caliente
System::change_db("sqlite", {"DB_PATH": "local.sqlite", "DB_PREFIX": "app_"})
```

## Carga Automática (Zero Imports)

No existen sentencias `import` ni `use`, y no se añadirán en el futuro. Todas las clases del proyecto (`app/controllers/`, `app/models/`, `app/libs/`), así como los plugins instalados (`plugins/` / `joss.yaml`), son indexados y cargados automáticamente. `import`, `@import`, `use`, `Namespace` y la grafía histórica `function` son sintaxis eliminada y el parser las rechaza.

Las funciones pueden anotar su retorno con `: Tipo` y son recursivas. Consulte [Funciones recursivas](RECURSION.md) para frames, límites y casos base.
