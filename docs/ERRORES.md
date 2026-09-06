# Entender y manejar errores

[Índice](README.md)

Antes: [clases](CLASES.md). Después: [concurrencia](CONCURRENCIA.md).
Consulta rápida: [códigos de diagnóstico](DIAGNOSTICOS.md).

Un error indica que el programa no puede cumplir una instrucción o contrato.
Puede estar en la escritura del programa o surgir por circunstancias externas,
como un archivo que desapareció. Distinguir esas situaciones ayuda a corregirlo.

| Momento | Ejemplo | Cómo resolverlo |
|---|---|---|
| Sintaxis | Falta una comilla o un paréntesis. | Corrige el texto señalado por el parser. |
| Análisis semántico | Se pasa texto a un parámetro entero. | Corrige valores o firma antes de ejecutar. |
| Ejecución | Un índice calculado está fuera de rango. | Verifica los datos y maneja el fallo esperado. |
| Operación de biblioteca | No existe un archivo. | Comprueba su retorno; no todas las APIs lanzan. |

## Leer un diagnóstico

```text
error[JOSS-TYPE-001] ejemplo.joss:3:2: ...
```

`error` es la severidad, `JOSS-TYPE-001` identifica una regla estable y `3:2`
indica línea y columna. El mensaje explica la incompatibilidad; la sugerencia
propone una corrección. Comienza por el primer error: un problema de sintaxis
puede generar varios mensajes derivados.

Un **warning** avisa de algo como una variable sin uso. `joss analyze` termina
con código cero cuando sólo hay warnings; los errores bloquean `joss run`.
`try/catch` no atrapa errores de análisis: todavía no se ha ejecutado código.

## Lanzar y recuperar

Una **excepción** interrumpe el camino normal. `throw` la lanza; `try` delimita
el trabajo y `catch` contiene una alternativa para un fallo durante ese trabajo.

<!-- joss-run: ["No se pudo continuar: faltan datos"] -->
```joss
try {
    throw "faltan datos"
} catch ($error) {
    print("No se pudo continuar: " . $error)
}
```

El cuerpo de `catch` se ejecuta una vez y el programa sigue después del bloque.
La variable de captura no lleva un tipo en esta sintaxis. No hay `finally`,
`defer` ni filtros de captura por tipo. Un `return` o el control de un ciclo no
deben ser tratados como errores del usuario.

Una función puede lanzar en lugar de retornar cuando su contrato no puede
cumplirse. Evita un `catch` vacío que oculta un fallo del que depende el resto
del programa.

## No todos los errores tienen la misma forma

El runtime conserva errores estructurados con archivo, posición y frames de
llamada. Una **traza de pila** muestra por qué llamadas se llegó al fallo.
No es una lista de instrucciones que debas ejecutar.

Dentro de `catch`, un `JossError` se expone como mapa con `message`, `type`,
`file`, `line` y `error`; otros errores se convierten a string. No presupongas
que `$error` siempre es una instancia `Exception`, ni que el mapa incluya
`code` o `stack`: el adaptador actual no expone todos los campos internos.

## Comprobar retornos de APIs

`file_get_contents` devuelve `null` al fallar; no dispara el `catch` por sí solo.
Este ejemplo trabaja con un archivo de nombre deliberadamente ausente en una
carpeta vacía:

<!-- joss-run: ["No se pudo leer el archivo"] -->
```joss
$contenido = file_get_contents("archivo-que-no-existe.txt")
($contenido == null) ? {
    print("No se pudo leer el archivo")
} : {
    print($contenido)
}
```

`file_put_contents` devuelve `true` o `false`. `Http::request` devuelve un mapa
con resultado y estado. Para cada API consulta **retorno y fallos**, además de
su nombre. Una respuesta HTTP 404 y un fallo de conexión tampoco son lo mismo.

Ejercicio: escribe una función que lea un archivo y lance `throw "lectura fallida"`
si recibe `null`; recupérala en un `try/catch`. El tutorial de
[consola](PROYECTO_CONSOLA.md) combina lectura, validación y escritura.
