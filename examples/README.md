# JosSecurity Examples

Este directorio contiene ejemplos de uso del lenguaje JosSecurity.

## Ejemplos Principales

### final_test.joss
Test comprehensivo que demuestra todas las características del lenguaje:
- Herencia y POO
- Smart Numerics (división automática a float)
- Maps nativos con sintaxis `{ key: value }`
- Modos de impresión (print, echo, cout, printf)
- Autenticación con JWT
- CRUD con GranMySQL
- Try-Catch y manejo de errores
- Operadores ternarios
- **Concurrencia con async/await**

### jwt_test.joss
Demostración de generación de tokens JWT con el módulo Auth.

### jwt_refresh_test.joss
Ejemplo de tokens JWT con expiración configurable y refresh tokens.

## Cómo Ejecutar

```bash
# Ejecutar el test comprehensivo
go run cmd/joss/main.go run examples/final_test.joss

# Ejecutar tests de JWT
go run cmd/joss/main.go run examples/jwt_test.joss
go run cmd/joss/main.go run examples/jwt_refresh_test.joss
```

## Características Demostradas

- **Fase 1**: Smart Numerics y Maps nativos
- **Fase 2**: Autoloading de clases desde `./classes/`
- **Fase 3**: Concurrencia con `async` y `await`
