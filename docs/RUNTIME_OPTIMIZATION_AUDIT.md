# Auditoría del runtime antes de optimizar

Estado auditado: commit `9d27239`, Go 1.26.2, Windows/amd64. Este documento
describe el runtime anterior a las optimizaciones. Las mediciones reproducibles
están en `benchmarks/BASELINE_9D27239.md`.

## Flujo real

```text
.joss
  -> lexer (Token por valor)
  -> parser Pratt (nodos AST enlazados por interfaces)
  -> analyzer (scopes y símbolos por map[string]*symbol)
  -> diagnostics
  -> Runtime.Execute
  -> executeStatement (type switch)
  -> evaluateExpression (type switch)
  -> helper especializado
  -> valor Go en interface{}
```

`pkg/bytecode` no altera este flujo: restaura mediante gob+flate el mismo AST
que vuelve a recorrer el evaluator. No hay IR tipada ni resolución semántica
persistida entre analyzer y runtime.

## Capas atravesadas por una expresión

Una expresión top-level normal atraviesa al menos cuatro decisiones:

1. `Execute` registra declaraciones y selecciona el modo script/Main.
2. `executeStatement` hace un type switch de statements.
3. `evaluateExpression` hace otro type switch de expressions.
4. El helper (`evaluateInfix`, `evaluateMember`, `executeCall`, etc.) vuelve a
   discriminar tipos Go, buscar nombres o recorrer AST.

Una llamada añade evaluación y materialización de argumentos, resolución por
nombre, clasificación del callable, creación del frame, binding y comprobación
de parámetros, ejecución del cuerpo, `panic/recover` de retorno y validación
del tipo de retorno. Un acceso a método añade búsqueda de clase, recorrido de
herencia, recorrido lineal de los statements de cada clase y construcción de
un `BoundMethod`.

## Representaciones y estructuras creadas

| Concepto | Representación anterior | Coste relevante |
|---|---|---|
| Valores | `interface{}` | boxing de primitivas cuando escapan; type switches repetidos |
| Variables | `Runtime.Variables map[string]interface{}` | hash por lectura/escritura local y global |
| Tipos runtime | `VarTypes map[string]string` | parseo repetido del nombre de tipo |
| Constantes | `Constants map[string]bool` | lookup separado |
| Frame | sustitución temporal de tres mapas del `Runtime` | tres mapas nuevos por llamada y copia de globals |
| Funciones | `map[string]*parser.MethodStatement` | resolución por nombre y ejecución del AST original |
| Closures | AST + copia de tres mapas | copia completa al capturar; mutex por entorno capturado |
| `ref` | tres mapas + nombre en `VariableReference` | lookup por nombre y posible referencia anidada |
| Objetos | `Instance{Class, Fields map, Constants map}` | propiedad por hash; metadato/método por recorrido del AST |
| Arrays | `[]interface{}` | elementos boxed; literales crecen con `append` sin capacidad |
| Maps | `map[string]interface{}` | claves evaluadas dinámicamente; literales sin capacidad inicial |
| Strings | `string` UTF-8 de Go | indexación anterior por byte, no por carácter |
| Futures | goroutine + runtime clonado + `chan bool` | fork copia mapas incluso para un valor trivial |
| Channels | `chan interface{}` | frontera dinámica sin contrato de elemento |

El paquete `pkg/core` contenía 104 archivos Go y 17,623 líneas. La búsqueda
estática encontró 651 menciones de `interface{}`, 243 de
`map[string]interface{}`, 17 usos de reflexión, 100 llamadas a `panic` y 29 a
`recover`. Estas cifras incluyen stdlib/framework y tests; no todas están en el
hot path, pero cuantifican la concentración de responsabilidades.

## Mapas de ejecución

### Hot path

```text
while
  -> evaluate condition
     -> identifier -> Variables[name]
     -> infix -> conversión numérica + operador
  -> execute block
     -> postfix
        -> identifier -> Variables[name]
        -> box int64
        -> Variables[name] = value
```

El perfil CPU del loop de 10,000 iteraciones sitúa `evaluateExpression` con
82.13% acumulado, `evaluateInfix` con 30.88%, `evaluatePostfix` con 46.71% y
las primitivas de mapas/hash entre los mayores costes planos.

### Allocation path

```text
call
  -> []interface{} de argumentos
  -> frameVariables map
  -> frameVarTypes map
  -> frameConstants map
  -> parameterNames map
  -> ReturnPanic heap object
```

En llamadas anidadas, `callMethodEvaluated` representa 93.12% plano del espacio
asignado. En el loop, `evaluatePostfix` representa 86.18% del espacio asignado:
el `int64` nuevo escapa al guardarse como `interface{}`.

### Dispatch path

```text
CallExpression
  -> evaluar argumentos
  -> IsBuiltin map lookup
  -> hasta cinco switches de familias builtin
  -> Functions[name]
  -> Variables[name]
  -> applyFunction
     -> PluginCallable / CapturedFunction / BoundMethod /
        MethodStatement / FunctionLiteral / NativeHandler /
        func([]interface{}) / reflect.Func
```

La resolución de métodos de usuario recorre statements de clase en cada acceso.
La resolución de clases/métodos de plugins puede recorrer todos los plugins y
sus símbolos. Los métodos estáticos crean además una instancia dummy.

### Error path

Conviven cuatro políticas:

- `panic(*JossError)` para algunos errores del lenguaje;
- `panic(string)` o `panic(error)` para otros;
- `fmt.Print*` seguido de `nil` para fallos recuperables;
- panics tipados (`ReturnPanic`, `BreakPanic`, `ContinuePanic`) para flujo normal.

`try/catch`, llamadas, loops, async y varios componentes framework recuperan
panics con criterios diferentes. El error estructurado previo no tenía código
estable, stack Joss, contexto, sugerencia ni causa.

### Type-check path

```text
binding/assignment/call/return/property write
  -> typesystem.Parse(typeName string)
  -> runtimeTypeOf(interface{})
  -> typesystem.Assignable
  -> si es instancia: recorrido de herencia
```

El analyzer ya conoce tipos de símbolos, parámetros, retornos y miembros, pero
esa información no queda anotada en el AST. El runtime vuelve a parsear strings
de tipo, resolver bindings, localizar miembros y validar contratos. `mixed`,
plugins y bytecode externo obligan a conservar un slow path dinámico; no
obligan a penalizar los call sites verificados.

## Semántica auditada

- `mixed` se representa como string de tipo y valor `interface{}`; desactiva la
  incompatibilidad estática/dinámica intencionalmente.
- `ref` no expone punteros Go; conserva los mapas del binding y el nombre. No
  puede cruzar plugins, async ni handlers nativos.
- Las exceptions y el control `return/break/continue` usan panic/recover.
- `async` clona el runtime antes de lanzar la goroutine. DB y registros
  inmutables se comparten; variables, tipos, constantes, maps, slices e
  instancias se copian parcialmente.
- Las closures retenidas copian el entorno y serializan su uso con un mutex.
- Arrays y maps son heterogéneos. El analyzer sólo conocía `array`/`map`, por lo
  que un índice retornaba `mixed`/`unknown`.
- La reflexión de Go aparece principalmente en builtins de arrays y en el
  fallback de callables host.
- La aritmética anterior convertía enteros a `float64` y luego a `int64`. Esto
  pierde precisión por encima de 2^53 y el overflow de `int64` era silencioso.
- `string[index]` usaba `str[idx]`: retornaba un byte UTF-8 convertido a string.

## Hotspots clasificados

| Prioridad | Hotspot | Evidencia |
|---|---|---|
| P0 | Variables locales en maps + boxing de postfix | loop: ~2.10 ms, 9,752 allocs; hash/map domina CPU; postfix 86.18% de alloc_space |
| P0 | Construcción de frames por llamada | simple: 760 B/7 allocs; nested: 2,312 B/23; `callMethodEvaluated` 93.12% de alloc_space |
| P1 | Resolución repetida de métodos/campos | perfil de objetos: `evaluateMember` 16.18% acumulado y `lookupInstanceFieldOwner` 6.08% |
| P1 | Retorno por panic/recover | cada llamada que retorna crea `ReturnPanic`; aparece en 65.54% acumulado del perfil de allocations de nested calls |
| P1 | Carga de AST serializado | ~255 us, 80,208 B y 857 allocs para el programa de startup |
| P1 | Aritmética entera vía float64 | pérdida de precisión demostrable y trabajo duplicado por operación |
| P2 | Dispatch builtin por cascada de switches | lookup inicial seguido por hasta cinco familias |
| P2 | Literales sin capacidad y metadatos por string | arrays/maps pequeños asignan más de lo necesario; tipos se vuelven a parsear |
| P2 | Fork completo por async | ~4.17 us, 1,680 B y 18 allocs para async+await trivial |
| P3 | Reflexión de compatibilidad host | no apareció como hotspot en programas Joss puros; conservar como slow path |
| P3 | Framework DB/HTTP/IO | importante arquitectónicamente, pero no dominó los perfiles del evaluator puro |

## Concurrencia y GC

La suite preexistente pasa:

```text
go test -race ./pkg/parser ./pkg/typesystem ./pkg/analyzer ./pkg/core
```

El perfil de bloqueo de async atribuye el bloqueo esperado a `Future.Wait` y
`runtime.chanrecv1`; no mostró contención de mutex de aplicación. El perfil de
mutex fue casi enteramente runtime/GC. Esto no prueba ausencia de riesgos en
plugins/framework: sólo establece la línea base cubierta por tests.

## Conclusión anterior a cambios

La primera optimización debe eliminar allocations y hashing dentro del loop y
en llamadas antes de introducir una representación `Value` nueva. La evidencia
no justifica todavía reemplazar globalmente `interface{}` ni construir una VM
completa. Sí justifica conservar una ruta AST y añadir metadatos preanalizados,
frames compactos/cachés sencillas y fast paths enteros seguros.
