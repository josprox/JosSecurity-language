# Clases, objetos y métodos

[Índice](README.md)

Antes: [funciones](FUNCIONES.md) y [colecciones](COLECCIONES.md).
Después: [errores](ERRORES.md). Referencia: [tipos](SISTEMA_TIPOS.md).

## Datos y comportamiento juntos

Una **clase** describe un tipo de entidad y sus operaciones. Un **objeto** o
**instancia** es una entidad concreta creada con esa clase. Una **propiedad**
guarda un dato del objeto; un **método** es una función asociada a él.

Piensa en un contador: tiene una cantidad y una operación que la incrementa.

<!-- joss-run: ["1", "2"] -->
```joss
public class Contador {
    public int $valor = 0

    public func incrementar(): int {
        $this->valor = $this->valor + 1
        return $this->valor
    }
}
$contador = new Contador()
print($contador->incrementar())
print($contador->incrementar())
```

`new Contador()` crea una instancia. `$this` significa «esta instancia» dentro
de sus métodos. `->` accede a una propiedad o método de una instancia.
Dos instancias tienen campos propios; asignar la misma instancia a dos nombres
no crea un objeto nuevo.

## Construcción

Un **constructor** prepara el objeto al crearlo. En clases fuente puedes usar
un método llamado `constructor`:

<!-- joss-run: ["Hola, Ada"] -->
```joss
public class Persona {
    private string $nombre = ""

    public func constructor(string $nombre) {
        $this->nombre = $nombre
    }

    public func saludar(): string {
        return "Hola, " . $this->nombre
    }
}
$persona = new Persona("Ada")
print($persona->saludar())
```

La salida es `Hola, Ada`. La propiedad privada obliga a trabajar mediante los
métodos de la clase. El runtime también reconoce `Init constructor()` y
`Init main()` como inicializadores; `Init` no lleva visibilidad. Evita definir
varios inicializadores candidatos: no hay un sistema de sobrecargas por firma.
No presupongas encadenamiento automático del constructor de la superclase.

## Visibilidad

La **encapsulación** consiste en exponer operaciones útiles y limitar el acceso
a detalles que podrían romper el estado del objeto.

| Declaración | Modificadores de acceso |
|---|---|
| Clase o función global | `public` o `private`; privada a su archivo. |
| Método o propiedad | `public`, `protected` o `private`. |
| `Init` o closure | Sin modificador de acceso. |

`public` permite el acceso exterior, `private` lo restringe a la clase propietaria
del miembro y `protected` permite también acceso desde clases derivadas.
`static` no agrega visibilidad implícita. Declara explícitamente el acceso.

## Herencia

Una clase puede **heredar** de otra para reutilizar miembros. `extends` nombra
una sola superclase:

<!-- joss-run: ["hola"] -->
```joss
public class Mensaje {
    public func texto(): string { return "hola" }
}
public class Aviso extends Mensaje {}
$aviso = new Aviso()
print($aviso->texto())
```

Usa herencia cuando el nuevo objeto sea una especialización del anterior.
Si sólo necesitas reutilizar un cálculo, una función suele ser suficiente.
Joss no tiene interfaces, traits, protocolos ni clases genéricas definibles
por el usuario. `array<T>` y `map<K,V>` son anotaciones de colecciones, no un
sistema general para declarar clases parametrizadas.

## Clases integradas

Una **clase nativa** expone operaciones implementadas en Go, como
`JSON::stringify` o `GranDB::table`. La convención de acceso estático es `::`;
el acceso al objeto devuelto por un constructor o builder usa `->`.

No todas las clases nativas representan objetos que debas crear con `new`.
Algunas son fachadas de servicios y otras requieren estado inicializado por
una API. Consulta el contrato de cada una en la [biblioteca](MODULOS_NATIVOS.md).
La presencia de un nombre en el catálogo no prueba que cualquier combinación
de argumentos vaya a funcionar.

## Nulabilidad y acceso seguro

Una variable puede admitir una instancia o ausencia de valor con `Persona|null`
o `Persona?`. `?->` devuelve `null` si el receptor es nulo. No corrige un método
inexistente ni convierte una instancia en un mapa. Para una instancia devuelta
por `Auth::user()` usa `->`; para un registro de GranDB usa claves `[...]`.

Ejercicio: agrega a `Contador` un método `reiniciar` que asigne cero al campo.
Comprueba que reiniciar una instancia no cambie otra creada con `new`.
