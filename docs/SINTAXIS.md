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

## Carga Automática (Zero Imports)

No se requieren sentencias `import` ni `use`. Todas las clases del proyecto (`app/controllers/`, `app/models/`, `app/libs/`), así como los plugins instalados (`plugins/` / `joss.yaml`), son indexados y cargados automáticamente en memoria por el runtime de Joss. Las palabras clave `import` y `use` no existen en la sintaxis moderna.
