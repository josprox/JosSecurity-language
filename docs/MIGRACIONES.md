# Migraciones

[Índice](README.md) · Antes: [Schema](SCHEMA_BUILDER.md) · Después: [CLI](CLI.md)

Una migración guarda un cambio de estructura para que distintos entornos puedan aplicarlo en el mismo orden. `up` aplica el cambio; `down` describe cómo revertirlo, aunque no implica que exista un comando de rollback público.

```bash
joss make:migration create_products
joss migrate
joss migrate:fresh
```

El nombre puede escribirse como `create_products`, `create_products_table` o
`product`; las tres formas normalizan la tabla lógica a `products` y generan
`CreateProductsTable`. La grafía del comando es `make:migration`.

El generador crea una clase que extiende `Migration`, con `up()` y `down()`. El runner ejecuta migraciones pendientes en orden de nombre y registra el batch en la tabla de migraciones con el prefijo configurado.

```joss
public class CreateProductsTable extends Migration {
    public func up() {
        Schema::create("products", func(Blueprint $table) {
            $table->id()
            $table->string("name")
            $table->timestamps()
        })
    }

    public func down() {
        Schema::drop("products")
    }
}
```

No añadas manualmente el prefijo a menos que quieras fijarlo en código; `Schema` lo aplica desde `PREFIX`/`DB_PREFIX`.

`migrate:fresh` elimina todas las tablas visibles del esquema, vuelve a crear las tablas internas y ejecuta las migraciones. Es destructivo y está pensado para desarrollo o entornos desechables.

Una migración sólo se considera completada después de registrar su nombre y
batch. Un fallo de lectura, parseo, estructura o registro detiene el comando con
código de salida distinto de cero.

Hay adaptadores para SQLite, MySQL, PostgreSQL y SQL Server; valida las operaciones concretas en el motor elegido. Los detalles de columnas, índices y claves foráneas están en [Schema Builder](SCHEMA_BUILDER.md).
