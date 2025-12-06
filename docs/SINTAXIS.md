# Sintaxis de JosSecurity

## Tabla de Contenidos
- [Variables y Tipos](#variables-y-tipos)
- [Operadores Ternarios](#operadores-ternarios)
- [Clases y Herencia](#clases-y-herencia)
- [Estructuras de Control](#estructuras-de-control)
- [Funciones](#funciones)
- [Loops](#loops)
- [Try-Catch](#try-catch)
- [Arrays y Maps](#arrays-y-maps)
- [Operador Pipe](#operador-pipe)
- [Concurrencia](#concurrencia)

---

## Variables y Tipos

### Declaración de Variables

Todas las variables deben iniciar con `$` y tienen un tipo estático:

```joss
// Tipos primitivos
int $edad = 25
float $precio = 99.99
string $nombre = "Jose Luis"
string $alias = 'Pepe' // Comillas simples también soportadas
bool $activo = true

// Null
$valor = null
```

### Tipos Compuestos

```joss
// Arrays
array $lista = ["A", "B", "C"]
$numeros = [1, 2, 3, 4, 5]

// Maps (diccionarios)
$config = {
    "host": "localhost",
    "port": 3306,
    "debug": true
}

// Acceso a maps
print($config["host"])  // "localhost"

// Asignación en Maps
$config["port"] = 3307

// Indexación de Strings
string $texto = "Hola"
print($texto[0]) // "H"

```

### Constantes

```joss
// En config/reglas.joss
const string APP_NAME = "Mi Aplicación"
const int MAX_USERS = 100
const bool DEBUG = true
```

### Variables Especiales

```joss
__DIR__   // Directorio actual
__FILE__  // Archivo actual
```

---

## Estructuras de Control

JosSecurity soporta tanto estructuras de control tradicionales (`if/else`, `switch`) como operadores ternarios avanzados.

### If / Else

```joss
if ($edad >= 18) {
    print("Mayor de edad")
} else if ($edad >= 13) {
    print("Adolescente")
} else {
    print("Menor de edad")
}
```

### Switch

```joss
switch ($opcion) {
    case 1:
        print("Opción 1 seleccionada")
    case 2:
        print("Opción 2 seleccionada")
    default:
        print("Opción inválida")
}
```

## Operadores Ternarios

También puedes usar operadores ternarios para lógica concisa y el operador de fusión nula (Null Coalescing).

### Operador de Fusión Nula (??) - Strict Nil

El operador `??` comprueba estrictamente si el valor es `nil`. Si el valor es una cadena vacía `""` o `0`, **se conserva**.

```joss
// Sintaxis: valor ?? default
$usuario = nil
$nombre = $usuario ?? "Invitado" // "Invitado"

$titulo = ""
$texto = $titulo ?? "Sin título" // "" (Cadena vacía se mantiene)
```

### Operador Elvis (?:)

El operador Elvis `?:` es similar al ternario pero evalúa la "veracidad" (truthiness). Utilízalo si quieres defaults para valores vacíos o falsos.

```joss
// Sintaxis: valor ?: default
$nombre = "" ?: "Anónimo"  // "Anónimo"
$edad = 0 ?: 18            // 18
```

### Ternario Simple

```joss
// Sintaxis: (condición) ? valor_verdadero : valor_falso
$estado = ($edad >= 18) ? "Mayor" : "Menor"
```

### Ternario con Bloques (Ejecución)

Los bloques dentro de un ternario ahora se **ejecutan** correctamente, permitiendo `return` temprano.

```joss
($usuario->esValido()) ? {
    DB::save($usuario)
    return true
} : {
    return false
}
```

### Escalera Lógica (Múltiples Condiciones)

```joss
// Reemplazo de if/elseif/else
$nivel = ($puntos > 1000) ? "Oro"
         ($puntos > 500)  ? "Plata"
         ($puntos > 100)  ? "Bronce" :
                            "Novato"

// Ejemplo con bloques
($rol == "admin") ? {
    print("Acceso total")
    $permisos = ["read", "write", "delete"]
} : ($rol == "editor") ? {
    print("Acceso limitado")
    $permisos = ["read", "write"]
} : {
    print("Solo lectura")
    $permisos = ["read"]
}
```

### Operadores de Comparación

```joss
==  // Igual
!=  // Diferente
>   // Mayor que
<   // Menor que
>=  // Mayor o igual
<=  // Menor o igual
%   // Módulo (Resto de la división)

```

### Operadores Lógicos

```joss
&&  // AND
||  // OR
!   // NOT

// Ejemplos
($edad >= 18 && $activo) ? "Permitido" : "Denegado"
($admin || $moderador) ? "Acceso" : "Sin acceso"
(!$bloqueado) ? "Activo" : "Bloqueado"
```

---

## Clases y Herencia

### Definición de Clase

```joss
class Usuario {
    // Propiedades
    string $nombre
    string $email
    int $edad
    
    // Constructor
    Init constructor($n, $e, $ed) {
        $this->nombre = $n
        $this->email = $e
        $this->edad = $ed
    }
    
    // Métodos
    function saludar() {
        print("Hola, soy " . $this->nombre)
    }
    
    function esMayor() {
        return ($this->edad >= 18) ? true : false
    }
}
```

### Instanciación

```joss
$usuario = new Usuario("Juan", "juan@example.com", 25)
$usuario->saludar()  // "Hola, soy Juan"
```

### Herencia

```joss
class Animal {
    string $nombre
    
    Init constructor($n) {
        $this->nombre = $n
    }
    
    function hacerSonido() {
        print("...")
    }
}

class Perro extends Animal {
    // Sobrescribir método
    function hacerSonido() {
        print("Guau!")
    }
    
    // Nuevo método
    function moverCola() {
        print($this->nombre . " mueve la cola")
    }
}

$perro = new Perro("Rex")
$perro->hacerSonido()  // "Guau!"
$perro->moverCola()    // "Rex mueve la cola"
```

### Acceso a Propiedades

```joss
// Lectura
$nombre = $usuario->nombre

// Escritura
$usuario->email = "nuevo@example.com"

// Llamada a métodos
$resultado = $usuario->calcular()
```

---

## Funciones

### Declaración

```joss
// Función simple
function saludar() {
    print("Hola!")
}

// Con parámetros
function sumar($a, $b) {
    return $a + $b
}

// Con tipo de retorno
string function obtenerNombre() {
    return "Jose"
}

// Void (sin retorno)
void function mostrarMensaje($msg) {
    print($msg)
}
```

### Llamada

```joss
saludar()
$resultado = sumar(5, 3)
$nombre = obtenerNombre()
```

### Funciones en Clases

```joss
class Calculadora {
    function sumar($a, $b) {
        return $a + $b
    }
    
    function multiplicar($a, $b) {
        return $a * $b
    }
}

$calc = new Calculadora()
$suma = $calc->sumar(10, 5)
```

---

## Loops

### Foreach (Principal)

```joss
// Iterar array
$nombres = ["Juan", "María", "Pedro"]
foreach ($nombres as $nombre) {
    print($nombre)
}

// Con índice
foreach ($nombres as $i => $nombre) {
    print("$i: $nombre")
}

// Iterar map
$config = {"host": "localhost", "port": 3306}
foreach ($config as $key => $value) {
    print("$key = $value")
}

// Iterar resultados de base de datos
$usuarios = $db->table("users")->get()
foreach ($usuarios as $usuario) {
    print($usuario->nombre)
}
```

### Break y Continue

```joss
// Break: salir del loop
foreach ($items as $item) {
    ($item == "stop") ? {
        break
    } : {
        print($item)
    }
}

// Continue: siguiente iteración
foreach ($numeros as $num) {
    ($num % 2 == 0) ? {
        continue
    } : {
        print($num)  // Solo impares
    }
}
```

---

## Try-Catch

### Manejo de Errores

```joss
try {
    $resultado = operacionRiesgosa()
    print("Éxito: " . $resultado)
} catch ($error) {
    print("Error: " . $error)
}
```

### Try-Catch con Base de Datos

```joss
try {
    $db = new GranMySQL()
    $db->table("users")->insert(["nombre"], ["Juan"])
    print("Usuario creado")
} catch ($error) {
    print("Error al crear usuario: " . $error)
}
```

---

## Arrays y Maps

### Arrays

```joss
// Declaración
$lista = []
$numeros = [1, 2, 3, 4, 5]
$nombres = ["Ana", "Luis", "Carlos"]

// Acceso por índice
$primero = $lista[0]
$segundo = $lista[1]

// Modificación
$lista[0] = "Nuevo valor"

// Agregar elemento (Append)
$lista[] = "Elemento nuevo"

// Arrays Multilínea
$matriz = [
    [1, 2, 3],
    [4, 5, 6]
]

```

### Maps (Diccionarios)

```joss
// Declaración
$config = {
    "host": "localhost",
    "port": 3306,
    "debug": true
}

// Acceso
$host = $config["host"]

// Modificación
$config["port"] = 3307

// Agregar clave
$config["timeout"] = 30
```

### Funciones de Arrays

```joss
// Longitud
$cantidad = count($lista)

// Verificar si existe
$existe = isset($lista[0])

// Vacío
$vacio = empty($lista)
```

---

## Operadores Aritméticos

```joss
$suma = 10 + 5       // 15
$resta = 10 - 5      // 5
$mult = 10 * 5       // 50
$div = 10 / 3        // 3.333... (siempre float)
$mod = 10 % 3        // 1

// Incremento/Decremento
$x++
$x--
++$x
--$x

// Asignación compuesta
$x += 5
$x -= 3
$x *= 2
$x /= 4
```

---

## Operador Pipe

El operador `|>` permite encadenar llamadas de funciones de manera fluida, pasando el resultado de la izquierda como el **primer argumento** de la función de la derecha.

### Uso Básico

```joss
// Equivalente a: square(10 + 5)
$resultado = 10 + 5 |> square
print($resultado) // 225
```

### Encadenamiento

```joss
// Equivalente a: mul(add(10, 5), 2)
$res = 10 |> add(5) |> mul(2)
print($res) // 30
```

### Con Funciones Anónimas

```joss
$res = 20 |> func($x) {
    return $x / 2
}
print($res) // 10
```

### Con Builtins

```joss
"hello" |> len |> print
// Imprime: 5
```

---

## Concatenación de Strings

```joss
$nombre = "Jose"
$apellido = "Luis"

// Con operador .
$completo = $nombre . " " . $apellido

// En print
print("Hola " . $nombre)

// Interpolación en strings
$mensaje = "Bienvenido, $nombre"
```

---

## Comentarios

```joss
// Comentario de una línea

/*
 * Comentario
 * de múltiples
 * líneas
 */
```

---

## Ejemplos Completos

### Ejemplo 1: Validación de Usuario

```joss
class Validador {
    function validarEmail($email) {
        // Verificar que contenga @
        $tieneArroba = (strpos($email, "@") > 0) ? true : false
        
        return ($tieneArroba) ? {
            return true
        } : {
            return false
        }
    }
}

$validador = new Validador()
$email = "usuario@example.com"

($validador->validarEmail($email)) ? {
    print("Email válido")
} : {
    print("Email inválido")
}
```

### Ejemplo 2: Procesamiento de Lista

```joss
$numeros = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
$pares = []
$impares = []

foreach ($numeros as $num) {
    ($num % 2 == 0) ? {
        $pares[] = $num
    } : {
        $impares[] = $num
    }
}

print("Pares: " . count($pares))
print("Impares: " . count($impares))
```

### Ejemplo 3: Sistema de Permisos

```joss
class Usuario {
    string $rol
    
    Init constructor($r) {
        $this->rol = $r
    }
    
    function puedeEditar() {
        return ($this->rol == "admin" || $this->rol == "editor") ? true : false
    }
    
    function puedeEliminar() {
        return ($this->rol == "admin") ? true : false
    }
}

$usuario = new Usuario("editor")

($usuario->puedeEditar()) ? {
    print("Puede editar contenido")
} : {
    print("Sin permisos de edición")
}

($usuario->puedeEliminar()) ? {
    print("Puede eliminar")
} : {
    print("No puede eliminar")
}
```

---

## Concurrencia

JosSecurity soporta concurrencia moderna estilo Go.

- **Async/Await**: Ejecución asíncrona.
- **Canales**: Comunicación entre procesos.

Ver [CONCURRENCIA.md](./CONCURRENCIA.md) para la guía completa.
