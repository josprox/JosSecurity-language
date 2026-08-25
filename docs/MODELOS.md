# GranDB y Modelos (Compatible con Laravel Eloquent)

Un modelo hereda de `GranDB` o se consulta directamente mediante la fachada fluida. La sintaxis de llamada a métodos en instancias es `->`.

```joss
$query = GranDB::table("products")
$products = $query->where("active", true)->orderByDesc("created_at")->get()
```

El prefijo configurado en `env.joss` (`DB_PREFIX`) se aplica automáticamente a las tablas y columnas.

---

## 🚀 Métodos del Builder de GranDB (Suite Completa estilo Laravel)

### 1. Control de Flujo Condicional (`when` y `unless`)
```joss
$query = GranDB::table("users")
    ->when($roleFilter, func($q, $val) {
        $q->where("role", $val)
    })
    ->unless($showDeprecated, func($q) {
        $q->where("is_deprecated", 0)
    })
    ->get()
```

### 2. Cláusulas `where`, `orWhere` y Variaciones Avanzadas
- **Comparación entre Columnas**: `whereColumn("updated_at", ">", "created_at")`, `orWhereColumn`
- **Condiciones Negadas**: `whereNot("status", "banned")`, `orWhereNot`
- **Búsqueda LIKE Automática**: `whereLike("title", "joss")`, `orWhereLike("title", "joss")`
- **Nulos e In**: `whereNull("deleted_at")`, `orWhereNull`, `whereNotNull`, `orWhereNotNull`, `whereIn("status", ["active", "pending"])`, `orWhereIn`, `whereNotIn`, `orWhereNotIn`
- **Rangos**: `whereBetween("price", [10, 100])`, `orWhereBetween`, `whereNotBetween`, `orWhereNotBetween`
- **Componentes de Fecha**: `whereDate("created_at", "2026-08-07")`, `orWhereDate`, `whereYear`, `orWhereYear`, `whereMonth`, `orWhereMonth`, `whereDay`, `whereTime`
- **Campos JSON**: `whereJsonContains("tags", "php")`, `orWhereJsonContains`

```joss
// Agrupamiento con Closures (Paréntesis SQL)
$paquetes = GranDB::table("pub_packages")
    ->where("is_deprecated", 0)
    ->where(func($subQuery) {
        $subQuery->whereLike("name", $q)
                 ->orWhereLike("display_name", $q)
    })
    ->get()
```

### 3. Métodos de Búsqueda y Lectura
- `firstWhere($col, $op, $val)`: Atajo para `where($col, $op, $val)->first()`.
- `firstOrFail()`: Devuelve el primer elemento o lanza una excepción clara si viene vacío.
- `find($id)`: Busca por clave primaria.
- `findMany([$id1, $id2])`: Busca múltiples registros por ID.
- `findOrFail($id)`: Busca por ID o lanza una excepción explicativa.
- `sole()`: Devuelve exactamente 1 registro o falla si el conteo != 1.
- `paginate($perPage, $page)`: Paginación automática en 1 línea con metadatos (`total`, `last_page`, `current_page`).
- `chunk($size, func($items) { ... })`: Procesamiento por lotes en memoria.
- `pluck($col, $keyCol)`: Extrae una lista o mapa de valores.
- `value($col)`: Obtiene el valor de una sola columna del primer registro.
- `exists()` / `doesntExist()`: Verifica la existencia de registros.

### 4. Ordenación y Límites
- `orderBy($col, $direction)` / `orderByDesc($col)` / `orderByAsc($col)`
- `latest($col)` / `oldest($col)`
- `inRandomOrder()`
- `reorder()`: Limpia las cláusulas de ordenación.
- `limit($n)` / `take($n)` (Atajo)
- `offset($n)` / `skip($n)` (Atajo)
- `forPage($page, $perPage)`: Atajo de límite y salto por número de página.

### 5. Debugging e Inspección SQL
- `toSql()`: Devuelve la cadena SQL formateada.
- `getBindings()`: Devuelve el arreglo de parámetros asignados.
- `dump()`: Imprime el SQL y bindings en consola.
- `dd()`: Imprime el SQL y bindings y detiene la ejecución inmediatamente.

### 6. Escritura, Upserts, Incrementos y Transacciones
```joss
$db = GranDB::table("products")
$id = $db->insertGetId({"name": "Teclado RGB", "price": 899.99})

// Upsert automático (Actualiza si existe, inserta si no)
$db->updateOrInsert({"email": "user@example.com"}, {"name": "Usuario Nuevo"})

// Operaciones atómicas de contadores
$db->where("id", 1)->increment("downloads", 1)
$db->where("id", 1)->decrement("credits", 5)

// Actualizar marca de tiempo updated_at
$db->where("id", 1)->touch()

// Transacciones Atómicas (Commit automático y Rollback ante excepciones)
GranDB::transaction(func($db) {
    GranDB::table("wallets")->where("user_id", 1)->decrement("balance", 50)
    GranDB::table("wallets")->where("user_id", 2)->increment("balance", 50)
})

// Eliminaciones seguras
$db->where("id", $id)->delete()
```

### 7. Motores Soportados y Cambio Dinámico
GranDB es 100% compatible y optimizado para 4 motores de base de datos:
* **SQLite**: Archivos embebidos locales con modo WAL y timeout concurrente.
* **MySQL / MariaDB**: Sintaxis nativa, `LAST_INSERT_ID` y prefijos.
* **PostgreSQL**: Rebinding transparente `$1, $2`, secuencias `SERIAL` y `RETURNING id`.
* **Microsoft SQL Server (`sqlserver` / `mssql`)**: Delimitadores `[columna]`, binding `@p1`, `OFFSET...FETCH NEXT`, `TOP` y `OUTPUT INSERTED.id`.

```joss
// Cambio dinámico de motor en tiempo de ejecución:
System::change_db("sqlite", {"DB_PATH": "local.sqlite", "DB_PREFIX": "app_"})
GranDB::connection("postgres", {"DB_HOST": "127.0.0.1", "DB_NAME": "empresa"})
```

`delete()` sin cláusula `where` se aborta automáticamente por seguridad. `deleteAll()` y `truncate()` son métodos explícitos destructivos.
