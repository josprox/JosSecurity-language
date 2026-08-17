# Módulos Nativos, Clases y Funciones Globales en Joss

La tabla enumera la superficie registrada por el runtime. Una llamada estática usa `::`; una instancia usa `->`.

| Clase | Métodos implementados |
|---|---|
| `GranDB` | `table`, `select`, `distinct`, filtros `where*` (`whereIn`, `whereNotIn`, `whereBetween`, `whereNull`, `whereLike`, `whereDate`, etc.), joins (`join`, `leftJoin`, `rightJoin`, `crossJoin`), agregados (`count`, `sum`, `avg`, `min`, `max`), orden (`orderBy`, `latest`, `oldest`, `inRandomOrder`), paginación (`limit`, `offset`, `take`, `skip`, `paginate`), terminales (`get`, `first`, `find`, `value`, `pluck`, `exists`), mutaciones (`insert`, `insertGetId`, `update`, `updateOrInsert`, `increment`, `decrement`, `delete`, `deleteAll`, `truncate`), transacciones (`transaction`), cambio dinámico de motor (`changeDB`, `connection`, `use`). Compatible con **SQLite, MySQL, PostgreSQL y Microsoft SQL Server (`sqlserver`/`mssql`)**. |
| `Auth` | `hash`, `create`, `attempt`, `login`, `complete2FA`, `check`, `guest`, `user`, `id`, `hasRole`, `verify`, `forgotPassword`, `resetPassword`, `resendVerification`, `refresh`, `update`, `delete`, `logout`, `validateToken` (100% agnóstico mediante GranDB y Schema) |
| `MFA` / `TwoFactor` | generación y verificación TOTP, códigos de recuperación, consulta de requisito y verificación del segundo factor |
| `Router` | `get`, `post`, `put`, `patch`, `delete`, `head`, `options`, `query`, `any`, `match`, `api`, `ws`, `group`, `middleware`, `registerMiddleware`, `end` |
| `Http` | `get`, `post`, `put`, `patch`, `delete`, `head`, `options`, `query`, `json`, `request` *(Cliente HTTP universal de alto rendimiento)* |
| `Request` | `input`, `post`, `all`, `except`, `file`, `cookie`, `header`, `root` |
| `Response` | `json`, `error`, `redirect`, `back`, `raw`, `stream` |
| `WebResponse` | `with`, `withCookie`, `withHeader`, `status` |
| `Schema` / `Blueprint` | creación y consulta de tablas (`create`, `table`, `drop`, `dropIfExists`, `hasTable`, `hasColumn`, `rename`), tipos y modificadores descritos en [Schema Builder](SCHEMA_BUILDER.md). Compatible con SQLite, MySQL, PostgreSQL y SQL Server. |
| `Session` | `get`, `put`, `has`, `forget`, `all` |
| `System` | `change_db` *(cambio de motor SQL en caliente)*, `env`, `Run`, `load_driver`, `driver_call`, `log`, `sleep`, `now` |
| `Plugin` | `call`, `stream`, `path`, `platform` |
| `SEO` | `title`, `description`, `keywords`, `canonical`, `og`, `twitter`, `meta` |
| Utilidades | `Math`, `Str` (`length`, `random`, `startsWith`, `substring`, `indexOf`, `contains`, `trim`, `replace`), `UUID`, `JSON`, `Markdown`, `Cache`, `Zip`, `Stack`, `Queue` |
| Procesos | `Process`, `Server`, `Stream` |
| Aplicación | `View`, `Cron`, `Task`, `Lang`, `SEO`, `Sitemap`, `UserStorage`, `SQLite`, `Redis`, `WebSocket` |

---

## Funciones Globales (Built-ins)

Joss proporciona una amplia biblioteca de funciones globales nativas listas para usar sin necesidad de imports:

### 1. Fechas y Tiempo
* `time()`: Retorna el timestamp Unix actual en segundos (`int64`).
* `date(format, [timestamp])`: Formatea una fecha según los tokens estándar (`Y-m-d H:i:s`, `d/m/Y`, `H:i`, `c`, `r`, etc.). Si no se proporciona timestamp, usa la hora actual.
* `strtotime(string, [baseTimestamp])`: Convierte expresiones humanas o fechas textuales a timestamp Unix (`"-2 days"`, `"+1 week"`, `"tomorrow"`, `"yesterday"`, `"2026-08-16 12:00:00"`).
* `now([format])`: Retorna la fecha y hora actual formateada (por defecto `Y-m-d H:i:s`).
* `microtime([asFloat])`: Retorna el timestamp con microsegundos (como float o como cadena `"msec sec"`).
* `sleep(seconds)`: Pausa la ejecución durante el número de segundos especificado.
* `usleep(microseconds)`: Pausa la ejecución en microsegundos.

### 2. Cadenas de Texto y Criptografía
* `str_contains(haystack, needle)`: Verifica si `needle` está contenida en `haystack` (booleano).
* `str_starts_with(haystack, prefix)`: Verifica si la cadena inicia con el prefijo dado.
* `str_ends_with(haystack, suffix)`: Verifica si la cadena termina con el sufijo dado.
* `str_replace(search, replace, subject)`: Reemplaza todas las apariciones de `search` por `replace` en `subject`.
* `strtolower(string)` / `to_lower`: Convierte la cadena a minúsculas.
* `strtoupper(string)` / `to_upper`: Convierte la cadena a mayúsculas.
* `trim(string, [cutset])`: Elimina espacios o caracteres dados al inicio y final.
* `ltrim(string, [cutset])` / `rtrim(string, [cutset])`: Recorta espacios a la izquierda o derecha.
* `substr(string, start, [length])`: Extrae una subcadena soportando índices negativos.
* `strpos(haystack, needle)`: Encuentra la posición numérica de la primera coincidencia o `false`.
* `implode(glue, array)` / `join`: Une los elementos de un arreglo en una cadena con el delimitador dado.
* `explode(delimiter, string)`: Divide una cadena en un arreglo según el delimitador.
* `md5(string)`: Genera el hash MD5 hexadecimal.
* `sha1(string)`: Genera el hash SHA-1 hexadecimal.
* `sha256(string)`: Genera el hash SHA-256 hexadecimal.
* `base64_encode(string)`: Codifica una cadena en Base64.
* `base64_decode(string)`: Decodifica una cadena Base64.
* `html_escape(string)`: Escapa caracteres especiales HTML.
* `__(key)`: Función de internacionalización / traducción (`i18n`).
* `csrf_field()`: Retorna el campo `<input type="hidden" name="_token" ...>` para formularios web.
* `print(args...)`, `echo(args...)`, `printf(format, args...)`.

### 3. Arreglos y Mapas
* `in_array(needle, haystack)`: Comprueba si un valor existe dentro de un arreglo o lista.
* `array_key_exists(key, map)`: Comprueba si una clave existe dentro de un mapa.
* `array_merge(arr1, arr2, ...)`: Fusiona dos o más arreglos o mapas.
* `array_push(array, ...items)`: Añade uno o más elementos al final del arreglo.
* `array_pop(array)`: Extrae y retorna el último elemento del arreglo.
* `array_shift(array)`: Extrae y retorna el primer elemento del arreglo.
* `array_slice(array, offset, [length])`: Extrae una porción de un arreglo.
* `keys(map)` / `array_keys`: Retorna una lista con todas las claves de un mapa.
* `values(map)` / `array_values`: Retorna una lista con todos los valores de un mapa.
* `end(array)`: Obtiene el último elemento de un arreglo sin modificarlo.
* `count(item)` / `len(item)`: Retorna la longitud de arreglos, mapas o cadenas.
* `isset(var)`: Evalúa si una variable o índice existe y no es nulo.
* `empty(var)`: Evalúa si una variable o valor está vacío (`null`, `""`, `0`, `[]`, `{}`).
* `is_string(val)`, `is_array(val)`, `is_null(val)`.

### 4. Sistema de Archivos
* `file_exists(path)`: Comprueba si un archivo o directorio existe.
* `file_get_contents(path)`: Lee el contenido completo de un archivo como string.
* `file_put_contents(path, content)`: Escribe datos en un archivo (creándolo si no existe).
* `unlink(path)` / `file_delete(path)`: Elimina un archivo del sistema de archivos.
* `mkdir(path)`: Crea un directorio (incluyendo carpetas padre intermedias).
* `is_dir(path)`: Comprueba si la ruta es un directorio.
* `is_file(path)`: Comprueba si la ruta es un archivo regular.

### 5. Asincronía, Concurrencia y Canales
* `async { ... }`: Ejecuta un bloque o función en segundo plano en una goroutine aislada retornando un objeto `Future`.
* `await $future`: Bloquea hasta que la tarea en segundo plano finalice y retorna su resultado.
* `make_chan([bufferSize])`: Crea un canal de comunicación concurrente seguro.
* `send($chan, $val)` o `$chan << $val`: Envía un mensaje por el canal.
* `recv($chan)`: Recibe un mensaje del canal.
* `close($chan)`: Cierra el canal.

### 6. Serialización y Formatos
* `json_encode(data)`: Convierte un objeto/arreglo/mapa en string JSON formateado.
* `json_decode(string)`: Parsea una cadena JSON a estructuras de datos nativas.
* `json_verify(string)`: Valida si una cadena es un JSON sintácticamente correcto.
* `toon_encode(data)`: Serializa estructuras en formato binario TOON de alta velocidad.
* `toon_decode(string)`: Decodifica un paquete TOON.
* `toon_verify(string)`: Verifica la integridad de un paquete TOON.
* `hive_read_box(path)`: Lee y decodifica cajas de almacenamiento local Hive.

### 7. Helpers de Aplicación Web y Servidor
* `env(key, [default])`: Lee variables de entorno cargadas desde `.env`.
* `config(key, [default])`: Alias para configuración y entorno.
* `view(viewName, [data])`: Renderiza una plantilla de vista con motor Joss Blade.
* `json(data, [status])`: Emite una respuesta HTTP JSON.
* `redirect(url)`: Emite una redirección HTTP.
* `back()`: Redirige a la página anterior con soporte de `.with()`.
* `response(content, [status])`: Emite una respuesta HTTP raw.
* `request([key], [default])`: Accede a los parámetros y cuerpo de la petición HTTP.
* `session([key], [value])`: Accede a la sesión de usuario activa.
* `run(scriptPath, ...args)`: Ejecuta scripts externos (`.py` o `.php`) si `ALLOW_SYSTEM_RUN=true`.

---

## Contratos relevantes

- `GranDB::table("users")->get()` retorna una lista nativa de mapas.
- `first()` retorna un mapa o `nil`.
- `Auth::user()` retorna una instancia; accede con `$user->email` o `$user->full_name`.
- `Request::file()` retorna un mapa cuyo contenido está en `content`.
- `Response::raw($data, $status, $mime, $headers)` evita la transformación HTML y sirve binarios.
- `Response::error($message, $status)` crea JSON con la clave `error`; el status predeterminado es 400.
- `System::load_driver($path, $name=nil)` carga una DLL, SO o dylib con la ABI C v1; `driver_call($name, $method, $args=[])` la invoca y decodifica su JSON.
