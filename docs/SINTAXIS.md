# Sintaxis de Joss

## Variables y tipos

Las variables usan `$`. La asignación simple es dinámica; una declaración tipada valida el valor. `function` se acepta por compatibilidad, pero la palabra canónica es `func`.

```joss
$dynamic = 10
int $age = 25
string $name = "Ada"
bool $active = true
array $items = [1, 2, 3]
$config = {"port": 8000}
```

Los tipos reconocidos por la validación del runtime incluyen `int`, `float`, `string`, `bool`, `array` y `map`. Una variable tipada como número intenta convertir una cadena numérica antes de fallar.

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


## Funciones, closures y clases

```joss
func sum(int $a, int $b) {
    return $a + $b
}

$double = func(int $value) {
    return $value * 2
}

class User {
    string $name

    Init constructor(string $name) {
        $this->name = $name
    }

    func greet() {
        return "Hola " . $this->name
    }
}
```

Las APIs estáticas usan `Clase::metodo()`. Las instancias usan `$object->method()` y `$object->property`.

## Control de flujo

No existe una sentencia `if/else`. Usa ternarios; los bloques también son expresiones. `return` se propaga fuera de bloques y ciclos anidados.

```joss
($age >= 18) ? {
    print("adulto")
} : {
    print("menor")
}

$label = ($active) ? "activo" : "inactivo"
$port = $config["port"] ?? 8000
$fallback = $name ?: "Anónimo"
```

`match` compara por tipo y valor, admite varias claves y `default`:

```joss
$message = match ($status) {
    200, 201 => "correcto",
    404 => "no encontrado",
    default => "error"
}
```

## Ciclos y errores

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

try {
    throw "fallo"
} catch ($error) {
    print($error)
}
```

`break` y `continue` funcionan en ciclos. El postfix implementado es `++`; `--` no existe todavía, por lo que un decremento se escribe como asignación. `isset()` y `empty()` son expresiones del lenguaje.

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
GranDB::transaction(function($db) {
    GranDB::table("accounts")->where("id", 1)->decrement("balance", 100)
    GranDB::table("accounts")->where("id", 2)->increment("balance", 100)
})

// Cambio dinámico de motor en caliente
System::change_db("sqlite", {"DB_PATH": "local.sqlite", "DB_PREFIX": "app_"})
```

## Carga Automática (Zero Imports)

No se requieren sentencias `import` ni `use`. Todas las clases del proyecto (`app/controllers/`, `app/models/`, `app/libs/`), así como los plugins instalados (`plugins/` / `joss.yaml`), son indexados y cargados automáticamente en memoria por el runtime de Joss. Las palabras clave `import` y `use` no existen en la sintaxis moderna.
