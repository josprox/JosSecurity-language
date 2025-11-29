# Migraciones de Base de Datos

Sistema de migraciones de JosSecurity para gestionar esquemas de base de datos.

## Crear Migración

### Ubicación
`app/database/migrations/`

### Nomenclatura
`001_descripcion.joss`, `002_descripcion.joss`, etc.

### Ejemplo Básico

```joss
// app/database/migrations/001_create_posts.joss
$schema = new Schema()
$schema->create("posts", {
    "id": "INT AUTO_INCREMENT PRIMARY KEY",
    "title": "VARCHAR(255) NOT NULL",
    "content": "TEXT",
    "user_id": "INT",
    "created_at": "TIMESTAMP DEFAULT CURRENT_TIMESTAMP",
    "updated_at": "TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"
})
```

## Ejecutar Migraciones

```bash
joss migrate
```

**Proceso**:
1. Crea tabla `js_migrations` si no existe
2. Crea tablas de sistema (`js_users`, `js_roles`, `js_cron`)
3. Ejecuta migraciones pendientes
4. Registra cada migración con batch number

## Tipos de Datos

### MySQL/SQLite
- `INT`, `BIGINT`, `SMALLINT`
- `VARCHAR(n)`, `TEXT`, `LONGTEXT`
- `DECIMAL(p,s)`, `FLOAT`, `DOUBLE`
- `DATE`, `DATETIME`, `TIMESTAMP`
- `BOOLEAN` (TINYINT(1))

### Modificadores
- `NOT NULL`
- `DEFAULT valor`
- `AUTO_INCREMENT`
- `PRIMARY KEY`
- `UNIQUE`

## Ejemplos

### Tabla con Relaciones

```joss
// 002_create_comments.joss
$schema = new Schema()
$schema->create("comments", {
    "id": "INT AUTO_INCREMENT PRIMARY KEY",
    "post_id": "INT NOT NULL",
    "user_id": "INT NOT NULL",
    "content": "TEXT NOT NULL",
    "created_at": "TIMESTAMP DEFAULT CURRENT_TIMESTAMP",
    "FOREIGN KEY (post_id)": "REFERENCES js_posts(id) ON DELETE CASCADE",
    "FOREIGN KEY (user_id)": "REFERENCES js_users(id) ON DELETE CASCADE"
})
```

### Índices

```joss
$schema->create("products", {
    "id": "INT AUTO_INCREMENT PRIMARY KEY",
    "name": "VARCHAR(255) NOT NULL",
    "price": "DECIMAL(10,2)",
    "INDEX idx_name": "(name)",
    "INDEX idx_price": "(price)"
})
```

## Sistema de Batches

Cada ejecución de `joss migrate` incrementa el batch number, permitiendo rollback por lotes.

**Tabla js_migrations**:
```
id | migration           | batch
1  | 001_create_posts    | 1
2  | 002_create_comments | 1
3  | 003_add_categories  | 2
```
