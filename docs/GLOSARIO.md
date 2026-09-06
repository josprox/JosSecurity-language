# Glosario

[Índice](README.md) · [Primeros pasos](PRIMEROS_PASOS.md) · [Referencia](SINTAXIS.md)

Estas palabras se explican también donde se introducen. Puedes volver aquí
sin interrumpir el recorrido de aprendizaje.

| Término | Significado en esta documentación |
|---|---|
| Programa | Instrucciones guardadas en archivos para que una computadora realice una tarea. |
| Valor | Un dato concreto: `12`, `"Ana"`, `true`. |
| Variable | Nombre que permite guardar y consultar un valor; por ejemplo `$edad`. |
| Binding | Asociación entre un nombre y su valor. Una constante protege esa asociación, no necesariamente el contenido de una colección. |
| Tipo | Clase de datos que una operación o variable admite. `int` representa enteros. |
| Inferencia | Deducción de un tipo a partir del valor, sin escribirlo explícitamente. |
| Conversión / casting | Obtención de un valor en otro tipo mediante reglas concretas; puede perder información. |
| `mixed` | Decisión explícita de aceptar valores de distintos tipos. |
| `unknown` | Falta de información del analizador; no es un tipo fuente para declarar variables. |
| `null` / `nil` | Ausencia de valor; ambas formas producen el mismo valor nulo. |
| Nullable | Tipo que permite además `null`, como `string|null` o `string?`. |
| Expresión | Código que produce un valor: `2 + 3`. |
| Sentencia | Instrucción que realiza un paso: declarar, retornar o repetir. |
| Bloque | Grupo de instrucciones entre llaves en un contexto que espera un cuerpo. |
| Función | Operación reutilizable que recibe datos y puede devolver un resultado. |
| Parámetro | Nombre y tipo de un dato recibido en la declaración de una función. |
| Argumento | Valor entregado cuando se llama a la función. |
| Retorno | Resultado que una función entrega con `return`. |
| Callable | Valor que el runtime puede invocar: función, método o closure, entre otros. No es una keyword fuente. |
| Ámbito / scope | Región donde un nombre se puede resolver. |
| Closure | Función anónima que conserva un entorno de variables de su creación. |
| Recursión | Función que se llama a sí misma, directa o indirectamente. |
| Clase | Declaración que agrupa propiedades y métodos. |
| Instancia | Objeto creado a partir de una clase mediante `new`. |
| Propiedad | Dato guardado dentro de una instancia. |
| Método | Función asociada a una clase; se llama con `->` en una instancia o `::` en contexto estático. |
| Constructor | Inicialización al crear una instancia; Joss admite `Init` en su contrato de clases. |
| Herencia | Reutilización de una clase base mediante `extends`. |
| Encapsulación | Control de acceso con `public`, `protected` o `private`. |
| Array | Colección ordenada de elementos, indexada desde cero; “lista” es una explicación, no el alias de tipo `list`. |
| Map | Colección de claves y valores; también se llama diccionario. |
| Referencia | Acceso al mismo almacenamiento. `ref` es además una capacidad temporal y restringida de modificar una variable del caller. |
| Mutabilidad | Posibilidad de cambiar un valor o el contenido de una estructura. |
| Copia superficial | Copia del contenedor que puede seguir compartiendo elementos interiores. |
| Error / diagnóstico | Problema detectado; un diagnóstico incluye ubicación, código y explicación. |
| Excepción | Fallo que interrumpe el flujo y puede recuperarse con `try/catch`. |
| Pila de llamadas / stack | Secuencia de funciones que están esperando que termine otra llamada. No confundir con la clase contenedora `Stack`. |
| Heap | Memoria para objetos cuya vida no se limita a una llamada; la administra Go, no el programa Joss manualmente. |
| Sincronía | Una operación termina antes de continuar con la siguiente. |
| Concurrencia | Varias tareas avanzan durante periodos superpuestos; no garantiza ejecución simultánea. |
| Asincronía | Una operación permite obtener su resultado posteriormente. |
| Future | Objeto runtime que representa un resultado pendiente de `async`; no es una keyword ni tipo fuente canónico. |
| `await` | Espera bloqueante del resultado de un Future en la ejecución actual. |
| Channel | Canal para enviar datos entre tareas y coordinarlas. |
| Runtime | Motor que ejecuta el programa y ofrece servicios integrados. |
| Lexer | Componente que convierte caracteres en tokens. |
| Token | Unidad reconocida: un nombre, un número, un operador, etc. |
| Parser | Componente que organiza tokens según la sintaxis. |
| AST | Árbol que representa la estructura del programa. |
| Analizador / analyzer | Componente que comprueba nombres, tipos y flujo antes de ejecutar. |
| Interpretación | Ejecución leyendo una representación del programa, como su AST. |
| Compilación | Transformación a otra representación. En Joss no implica necesariamente código máquina. |
| Bytecode | Formato intermedio. El principal de Joss contiene AST comprimido; JPBC tiene instrucciones para plugins. |
| Módulo | Capacidad integrada u organización física; Joss no tiene módulos fuente con imports. |
| Paquete | Unidad distribuible con metadatos y archivos. |
| Plugin | Extensión cargada por el runtime desde un paquete. |
| CLI | Programa que se controla escribiendo comandos en una terminal. |
| Terminal | Ventana para ejecutar comandos y ver su salida. |
| Ruta HTTP | Asociación entre una URL, un método HTTP y código que responde. |
| Controlador | Clase que prepara la respuesta a una petición web. |
| Middleware | Comprobación o transformación alrededor de una petición. |
| Vista | Plantilla que convierte datos en HTML. |
| Migración | Cambio versionado de la estructura de una base de datos. |
| Query builder | Objeto que construye una consulta antes de ejecutarla. |
| API | Conjunto de operaciones que otro programa o componente puede utilizar. |
