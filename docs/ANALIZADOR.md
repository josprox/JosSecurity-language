# Analizador Estático AST (`joss analyze`)

Joss incluye un analizador estático integrado en el motor (`pkg/core/analyzer.go`) que valida la corrección del código fuente mediante el recorrido del Árbol de Sintaxis Abstracta (AST) antes de su ejecución.

---

## 🎯 Propósito del Analizador Estático

El analizador estático previene errores comunes de desarrollo sin sobrecargar la experiencia del programador con restricciones rígidas innecesarias:

1. **Variables No Declaradas**: Identifica referencias a variables que no han sido inicializadas en el ámbito actual.
2. **Variables Sin Uso**: Notifica sobre variables declaradas pero nunca accedidas, ayudando a mantener un código limpio.
3. **Llamadas a Funciones Inexistentes**: Verifica invocaciones a funciones no definidas ni nativas (builtins).
4. **Instanciaciones Inválidas**: Inspecciona nombres de clases e instanciaciones con `new`.
5. **Validación de Builtins Centralizada**: Consulta el registro unificado `core.IsBuiltin(name)` para reconocer de forma segura todas las funciones integradas en el runtime nativo de Joss.

---

## 🚀 Modos de Ejecución

### 1. Inspección Standalone con `joss analyze`

Ejecuta el analizador en cualquier proyecto Joss para obtener un reporte completo de la calidad de código:

```bash
joss analyze
```

**Ejemplo de Salida:**

```text
🔍 Analizando proyecto Joss...
--------------------------------------------------
[AVISO] Archivo: main.joss, Línea 14, Columna 5: Variable '$temporal' declarada pero nunca usada.
[ERROR] Archivo: app/controllers/UserController.joss, Línea 28: Función 'calcular_hash()' no existe ni es una función builtin centralizada.

--------------------------------------------------
⚠️ Análisis completado con 1 aviso(s) y 1 error(es).
```

### 2. Inspección Automática en `joss run` y `joss server start`

Al ejecutar:

```bash
joss run main.joss
# o
joss server start
```

El analizador AST se ejecuta automáticamente en milisegundos previa evaluación del intérprete. Si existen advertencias o sugerencias no bloqueantes, las muestra de forma discreta antes del inicio del servidor o programa.

---

## ⚙️ Integración con la Arquitectura ALIM

Conforme a la filosofía **ALIM (Arquitectura de Lenguaje Integral Modular)**, el analizador estático de Joss es parte nativa del compilador y motor en Go, garantizando que el análisis estático no requiera plugins de terceros ni herramientas externas en el entorno del desarrollador.
