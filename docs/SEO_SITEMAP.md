# SEO y Sitemap

```joss
SEO::title("Productos")
SEO::description("Catálogo")
SEO::keywords(["joss", "productos"])
SEO::canonical("https://example.com/products")
SEO::og("image", "https://example.com/cover.png")
$tags = SEO::render()
```

También existe `SEO::meta($name, $content)`. La salida escapa atributos HTML y agrega una Twitter card predeterminada.

## Generación Dinámica de Sitemap (`/sitemap.xml`)

`/sitemap.xml` se genera en vivo en cada petición con una hermosa interfaz interactiva XSL (`/sitemap.xsl`). No escribe archivos físicos en disco.

### 1. Entradas Estáticas Manuales
```joss
Sitemap::add("/docs", "2026-07-15", "weekly", 0.8)

// O con mapas/arrays asociativos:
Sitemap::add({
    "url": "/contacto",
    "changefreq": "monthly",
    "priority": 0.6
})
```

### 2. Proveedores Dinámicos (Closures)
Para incluir automáticamente URLs desde la base de datos (por ejemplo, entradas de blog, productos de tienda, paquetes, etc.):
```joss
Sitemap::provider(func() {
    $posts = GranDB::table("cms_posts")->where("status", "published")->get()
    $items = []
    foreach ($posts as $p) {
        $items[] = {
            "url": "/blog/" . $p["slug"],
            "lastmod": $p["updated_at"],
            "changefreq": "daily",
            "priority": 0.9
        }
    }
    return $items
})
```

### 3. Exclusiones de Rutas
```joss
Sitemap::exclude([
    "/api/*",
    "/admin/*",
    "/checkout/*"
])
```

La URL base usa el request actual, después `APP_URL` y finalmente `http://localhost`. Detrás de un proxy configura correctamente `Host` y `X-Forwarded-Proto`.
