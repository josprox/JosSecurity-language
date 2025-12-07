# Guía de Integración: JosSecurity en HestiaCP

Este documento detalla los pasos para preparar un servidor VPS con **HestiaCP** para alojar aplicaciones desarrolladas en **JosSecurity**.

El sistema utiliza una arquitectura **"Zero Config"**: detecta automáticamente usuarios, rutas y puertos mediante un script de despliegue automatizado.

-----

## 📋 1. Requisitos Previos

Antes de ejecutar el instalador, asegúrate de cumplir con lo siguiente:

1.  **VPS con HestiaCP instalado** y funcionando.
2.  **Acceso Root** al servidor.
3.  **Lenguaje Joss Instalado**: El binario `joss` debe estar instalado en el servidor y accesible globalmente (ej. en `/usr/local/bin/joss`).
      * *Verificación:* Ejecuta `joss --version` en la terminal. Si da error, instálalo primero.

-----

## 🚀 2. Instalación del Stack (Solo una vez)

Este paso configura las plantillas de Nginx, los servicios de Systemd y las herramientas de despliegue.

1.  Sube el script `install_joss_stack.sh` a tu servidor (por ejemplo, a `/root/`).
2.  Dale permisos de ejecución y ejecútalo:

<!-- end list -->

```bash
chmod +x install_joss_stack.sh
./install_joss_stack.sh
```

**¿Qué acaba de pasar?**

  * ✅ Se instalaron las plantillas `joss` en HestiaCP.
  * ✅ Se creó el servicio universal `joss@.service`.
  * ✅ Se instaló el comando global `deploy-joss`.

-----

## 🌐 3. Crear un Nuevo Sitio en HestiaCP

Cada vez que quieras alojar un nuevo proyecto Joss:

1.  Entra a tu panel de HestiaCP.
2.  Ve a **WEB** -\> **Añadir dominio web**.
3.  Ingresa el nombre del dominio.
4.  Haz clic en **Opciones Avanzadas**.
5.  En **Plantilla Web NGINX**, selecciona la opción: `joss`.
6.  Guarda los cambios.

-----

## 📦 4. Despliegue del Proyecto

### En tu computadora local (Desarrollo):

1.  Compila tu proyecto para web. Esto generará la carpeta `build/` con el archivo `nginx_port.conf` automático.
    ```bash
    joss build web
    ```

### En el servidor (Producción):

1.  Sube **todo el contenido** de la carpeta local `build/` a la carpeta `public_html` del servidor:
      * Ruta típica: `/home/TU_USUARIO/web/TU_DOMINIO.com/public_html/`
2.  Conéctate por SSH (puedes ser root o admin).
3.  Ejecuta el comando de despliegue:

<!-- end list -->

```bash
deploy-joss midominio.com
```

**El comando `deploy-joss` se encargará de:**

  * Ajustar los permisos de archivos al usuario correcto.
  * Reiniciar el servicio de tu aplicación.
  * Recargar Nginx para leer el puerto asignado en `nginx_port.conf`.

-----

## 🛠 Comandos Útiles y Troubleshooting

### Ver estado de un servicio

Si tu web no carga, verifica si la aplicación está corriendo:

```bash
systemctl status joss@midominio.com
```

### Logs de errores

  * **Logs de la aplicación Joss:**
    ```bash
    journalctl -u joss@midominio.com -f
    ```
  * **Logs de Nginx:**
    ```bash
    tail -f /var/log/hestia/nginx-error.log
    ```

### Cambiar el puerto manualmente

Si necesitas cambiar el puerto de una app ya desplegada:

1.  Edita el archivo `nginx_port.conf` en el `public_html` del dominio.
2.  Cambia el puerto en tu configuración de Joss (si aplica).
3.  Ejecuta de nuevo: `deploy-joss midominio.com`.

-----

## 📂 Estructura de Archivos Esperada

El sistema espera que la carpeta `public_html` tenga esta estructura mínima para funcionar:

```text
/home/usuario/web/dominio.com/public_html/
├── main.joss          # Tu archivo de entrada
├── nginx_port.conf    # Generado por 'joss build web' (ej. set $joss_port 9005;)
├── env.enc            # Entorno encriptado
└── assets/            # Carpetas estáticas
```