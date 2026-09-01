# Benchmarks del runtime Joss

La suite vive junto al runtime para poder medir helpers internos sin exportar
APIs sólo para testing:

```powershell
go test ./pkg/core -run '^$' -bench '^BenchmarkJoss' -benchmem -benchtime=100ms -count=5
```

Incluye startup, operaciones básicas, control de flujo, funciones, objetos,
colecciones, exceptions, `ref`, async/await, channels y escenarios JSON, CRUD,
HTTP, transformación de arrays, templates y DB mapping.

Reglas de comparación:

1. Usar el mismo host, versión Go y `-benchtime`.
2. Ejecutar al menos cinco muestras.
3. Comparar mediana de `ns/op`, `B/op` y `allocs/op`.
4. No mezclar perfiles con `memprofilerate=1` en resultados de velocidad: esa
   opción distorsiona deliberadamente el tiempo para registrar cada allocation.
5. Conservar un benchmark vecino que detecte regresiones de seguridad cuando
   una optimización agregue un fast path.

Los perfiles de auditoría se generaron en el directorio temporal
`C:\Users\Asus\AppData\Local\Temp\joss-runtime-baseline-9d27239` y sus tops se
resumen en `docs/RUNTIME_OPTIMIZATION_AUDIT.md`; los binarios pprof no forman
parte del repositorio.
