# Funciones recursivas

[Índice](README.md)

Joss admite recursión directa y mutua. Para obtener comprobación estática útil, declare el tipo de retorno:

```joss
public func factorial(int $n): int {
    ($n <= 1) ? {
        return 1
    } : {
        return $n * factorial($n - 1)
    }
}
```

Las declaraciones de funciones y sus firmas se registran antes de analizar los cuerpos. Por eso una función puede referirse a sí misma o a otra función declarada después en el proyecto.

Cada llamada crea un frame independiente para parámetros, locales, tipos inferidos y constantes. Una función con nombre no puede leer por accidente los locales de su caller ni variables fuente top-level; los datos deben viajar por parámetros. Los bindings nativos y de plugins sí son visibles. Las instancias, mapas y arrays pasados como valores mantienen la semántica de referencia actual; aislar el binding local no convierte esos objetos en copias profundas. Las closures conservan un entorno capturado separado.

El runtime usa `Runtime.MaxCallDepth` y aplica 1024 frames por defecto. Excederlo produce `RecursionLimit` en lugar de dejar que la pila de Go crezca sin control. Este límite protege la ejecución, pero no sustituye un caso base correcto.

El analyzer valida los tipos de argumentos, cada `return` explícito y que toda ruta demostrable de una función anotada termine en `return` o `throw`. Reconoce bloques, ambos brazos de ternarios, `match` con `default` y ambos brazos de `try/catch`; no asume que un loop arbitrario termina.
