# Examples Directory

Este directorio contiene scripts de prueba para verificar las funcionalidades de JosSecurity.

## Scripts Disponibles

### 1. test_complete.joss
**Prueba completa de Auth y GranMySQL**
- Verifica auto-creación de tabla `js_users`
- Prueba registro de usuarios
- Prueba login exitoso y fallido
- Verifica fluent builder API
- Prueba insert con builder

```bash
go run ../cmd/joss/main.go run test_complete.joss
```

### 2. oop_test.joss
**Prueba de OOP (Programación Orientada a Objetos)**
- Clases y métodos
- Instanciación con `new`
- Llamadas a métodos

```bash
go run ../cmd/joss/main.go run oop_test.joss
```

### 3. helpers_test.joss
**Prueba de funciones helper**
- `isset()` - Verifica si variable está definida
- `empty()` - Verifica si variable está vacía
- `len()` - Longitud de arrays
- `count()` - Cuenta elementos

```bash
go run ../cmd/joss/main.go run helpers_test.joss
```

### 4. polyglot_test.joss
**Prueba de I/O Polyglot**
- `echo` / `print` (PHP/Python style)
- `cout <<` (C++ style)
- `cin >>` (C++ style)
- `printf` (C style)

```bash
go run ../cmd/joss/main.go run polyglot_test.joss
```

### 5. prueba_insercion/
**Directorio con tests de Auth y DB**
- `test_auth_db.joss` - Test completo de Auth con ternarios
- `migration.joss` - Script de migración de tablas

## Características Verificadas

✅ **GranMySQL con Prefijos**: Todas las tablas usan `js_` prefix  
✅ **Auth Auto-Migración**: Tabla `js_users` se crea automáticamente  
✅ **Fluent Builder API**: Sintaxis encadenada funcional  
✅ **Prepared Statements**: Protección contra SQL injection  
✅ **OOP**: Clases, métodos, instanciación  
✅ **Helpers**: isset, empty, len, count  
✅ **Polyglot I/O**: Múltiples estilos de entrada/salida  
