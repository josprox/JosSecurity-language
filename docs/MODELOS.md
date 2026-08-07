# GranDB y modelos

Un modelo puede heredar de `GranDB`, pero el constructor de consultas también se usa directamente. La sintaxis válida de instancia es `->`.

```joss
$query = GranDB::table("products")
$products = $query->where("active", true)->orderBy("created_at", "DESC")->get()
```

El prefijo de `env.joss` se aplica automáticamente a nombres sin prefijo.

## Subdirectorios y Pre-carga Automática

Los modelos pueden organizarse en subcarpetas temáticas dentro de `app/models/` (por ejemplo, `app/models/auth/User.joss`, `app/models/shop/Product.joss`). El motor de Joss escanea y precarga recursivamente todo el árbol de modelos al iniciar la aplicación.

Para generar un modelo en un subdirectorio:
```bash
joss make:model auth/User
```
El CLI generará `app/models/auth/User.joss` declarando de forma limpia `class User extends GranDB`.

## Lectura

```joss
$all = GranDB::table("products")->get()
$one = GranDB::table("products")->find(5)
$first = GranDB::table("products")->where("price", "<", 100)->first()
$names = GranDB::table("products")->pluck("name")
$exists = GranDB::table("products")->where("sku", "A-1")->exists()
$total = GranDB::table("products")->count()
```

`get()` devuelve una lista nativa de mapas; `first()` y `find()` devuelven un mapa o `nil`.

## Escritura

```joss
$db = GranDB::table("products")
$db->insert({"name": "Teclado", "price": 899.99})
$db->where("id", 5)->update({"price": 799.99})
$db->where("id", 5)->delete()
```

`delete()` sin `where` se aborta por seguridad. `deleteAll()` y `truncate()` son operaciones explícitas destructivas.

## Joins y filtros

GranDB implementa `where`, `orWhere`, `whereIn`, `orWhereIn`, `whereNotIn`, `whereNull`, `whereNotNull`, `whereBetween`, `whereNotBetween`, `join`, `innerJoin`, `leftJoin` y `rightJoin`. También expone `select`, `sum`, `avg`, `min`, `max`, `latest`, `oldest`, `inRandomOrder`, `limit` y `offset`.

No existe una capa automática de relaciones de objetos. Los joins producen mapas con las columnas seleccionadas.
