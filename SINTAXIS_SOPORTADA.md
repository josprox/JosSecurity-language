# Resumen de Sintaxis NO Soportada por el Parser Actual

## ❌ NO Soportado

1. **Anotaciones de tipo en parámetros de función:**
   ```joss
   func findByEmail(string $email) { }  // ❌ NO
   ```
   
2. **Declaraciones de tipo en variables dentro de funciones:**
   ```joss
   string $email = Request::input("email")  // ❌ NO
   ```

3. **Propiedades públicas en clases:**
   ```joss
   public string $tabla = "users"  // ❌ NO
   ```

4. **Import/Namespace:**
   ```joss
   Import System.Security  // ❌ NO
   Namespace App  // ❌ NO
   @import "global"  // ❌ NO
   ```

5. **Sintaxis de array con =>:**
   ```joss
   ["key" => "value"]  // ❌ NO
   ```

## ✅ Sintaxis Correcta

1. **Parámetros sin tipo:**
   ```joss
   func findByEmail($email) { }  // ✅ SÍ
   ```

2. **Variables sin declaración de tipo:**
   ```joss
   $email = Request::input("email")  // ✅ SÍ
   var $email = Request::input("email")  // ✅ SÍ también
   ```

3. **Sin propiedades públicas:**
   ```joss
   class User extends GranMySQL {
       func findByEmail($email) { }  // ✅ SÍ
   }
   ```

4. **Sin imports:**
   ```joss
   class AuthController {  // ✅ SÍ - directo
       func showLogin() { }
   }
   ```

5. **Maps con dos puntos:**
   ```joss
   {"key": "value"}  // ✅ SÍ
   ```

## Actualizar Plantillas

Recuerda actualizar `pkg/template/templateBasic.go` para que genere código compatible.
