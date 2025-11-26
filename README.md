# JosSecurity (Joss)

Implementación de referencia del lenguaje y framework JosSecurity.

## Requisitos
- Go 1.20 o superior

## Ejecución

### 1. Compilar el CLI
```bash
go build -o joss.exe ./cmd/joss
```

### 2. Comandos Disponibles

**Verificar versión:**
```bash
./joss.exe version
```

**Iniciar Servidor:**
```bash
./joss.exe server start
```
Esto iniciará el servidor en http://localhost:8000.

## Estructura
- `cmd/joss`: Punto de entrada del CLI.
- `pkg/core`: Lógica del runtime y seguridad.
- `pkg/server`: Servidor HTTP.
