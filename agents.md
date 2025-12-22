# Contexto para Agentes de IA

Este documento sirve como memoria persistente para futuros agentes que trabajen en este proyecto.

## Lecciones Aprendidas (Sesión 21/12/2025)

### 1. Intérprete JosSecurity - Comportamientos Clave
- **Retornos Estrictos**: El intérprete detiene la ejecución de un bloque inmediatamente al encontrar un `ReturnStatement`.
- **JSON Parsing**: `JSON::parse()` requiere estrictamente un `string`. Si pasas un objeto (como una lista de BD), retornará `nil` o fallará.
- **Base de Datos**: `GranMySQL::get()` retorna un `[]map[string]interface{}` (Lista Nativa), NO un string JSON. No es necesario parsearlo.
- **Foreach**: Soporta iteración sobre `[]interface{}` y `[]map[string]interface{}`.
- **Prefix Expressions**: Operadores como `!` y `-` funcionan correctamente en el evaluador (`evaluatePrefix`).

### 2. Manejo de Archivos y Descargas
- **Uploads**: Los archivos subidos se encuentran en `$file["content"]`, no en `tmp_name`. El servidor lee el contenido en memoria.
- **Downloads**: Para descargar archivos binarios sin corrupción:
  1. Usar `Response::raw($content, $status, $mime, $headers)`.
  2. Forzar headers: `Content-Disposition: attachment; filename="..."`.
  3. **IMPORTANTE**: No retornar strings simples para binarios, ya que el servidor podría inyectar scripts (Hot Reload) y corromper el archivo.

### 3. Estructura de Controladores
- **Sintaxis**: JosSecurity no usa `if/else`, usa ternarios `cond ? { ... } : { ... }`.
- **Métodos**: Asegurarse de que cada método esté correctamente cerrado con `}`. Un error de anidamiento puede hacer que el Dispatcher no encuentre el método.

### 4. Estilo de Código
- Usar bloques `{ ... }` explícitos dentro de los ternarios para flujos complejos.
- Para concatenar strings usar `.`.
