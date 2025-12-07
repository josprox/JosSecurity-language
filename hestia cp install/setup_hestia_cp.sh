#!/bin/bash

# ==============================================================================
# INSTALL JOSSECURITY STACK FOR HESTIACP
# Autor: JossSecurity (Generado por Asistente)
# Descripción: Prepara un servidor HestiaCP para alojar múltiples apps Joss
#              de forma dinámica y sin configuración manual por dominio.
# ==============================================================================

# Colores
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

# Verificar que se ejecuta como root
if [ "$EUID" -ne 0 ]; then
  echo -e "${RED}Por favor, ejecuta este script como root.${NC}"
  exit 1
fi

echo -e "${BLUE}>>> INICIANDO INSTALACIÓN DEL STACK JOSSECURITY...${NC}"

# ------------------------------------------------------------------------------
# 1. VERIFICACIÓN DE DEPENDENCIAS
# ------------------------------------------------------------------------------
echo -e "${BLUE}[1/5] Verificando instalación de Joss...${NC}"

JOSS_BIN=$(which joss)
if [ -z "$JOSS_BIN" ]; then
    # Intento buscar en rutas comunes si no está en el PATH del root
    if [ -f "/usr/local/bin/joss" ]; then
        JOSS_BIN="/usr/local/bin/joss"
    elif [ -f "/usr/bin/joss" ]; then
        JOSS_BIN="/usr/bin/joss"
    else
        echo -e "${RED}ERROR CRÍTICO: No se encontró el ejecutable 'joss'.${NC}"
        echo "Primero instala tu lenguaje en el VPS y asegúrate de que sea accesible."
        exit 1
    fi
fi
echo -e "${GREEN} -> Joss detectado en: $JOSS_BIN${NC}"

# ------------------------------------------------------------------------------
# 2. INSTALACIÓN DE PLANTILLAS NGINX (HESTIACP)
# ------------------------------------------------------------------------------
echo -e "${BLUE}[2/5] Instalando plantillas Nginx dinámicas...${NC}"

HESTIA_TPL_DIR="/usr/local/hestia/data/templates/web/nginx"

if [ ! -d "$HESTIA_TPL_DIR" ]; then
    echo -e "${RED}Error: No se encontró el directorio de plantillas de HestiaCP.${NC}"
    echo "¿Está HestiaCP instalado?"
    exit 1
fi

# Plantilla HTTP (.tpl)
cat > "$HESTIA_TPL_DIR/joss.tpl" <<EOF
server {
    listen      %ip%:%proxy_port%;
    server_name %domain_idn% %alias_idn%;
    error_log   /var/log/%web_system%/domains/%domain%.error.log error;

    include %home%/%user%/conf/web/%domain%/nginx.forcessl.conf*;

    location / {
        # --- Configuración Dinámica JosSecurity ---
        set \$joss_port 9000; 
        include %docroot%/nginx_port.conf*;
        # ------------------------------------------

        proxy_pass http://127.0.0.1:\$joss_port;
        
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;

        # Archivos estáticos
        location ~* ^.+\.(%proxy_extensions%)\$ {
            root       %docroot%;
            try_files  \$uri @joss_app;
            access_log /var/log/%web_system%/domains/%domain%.log combined;
            expires    max;
        }
    }

    location @joss_app {
        set \$joss_port 9000;
        include %docroot%/nginx_port.conf*;
        
        proxy_pass http://127.0.0.1:\$joss_port;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
    }

    location /error/ {
        alias %home%/%user%/web/%domain%/document_errors/;
    }

    include %home%/%user%/conf/web/%domain%/nginx.conf_*;
}
EOF

# Plantilla HTTPS (.stpl)
cat > "$HESTIA_TPL_DIR/joss.stpl" <<EOF
server {
    listen      %ip%:%proxy_ssl_port% ssl;
    server_name %domain_idn% %alias_idn%;
    error_log   /var/log/%web_system%/domains/%domain%.error.log error;

    ssl_certificate     %ssl_pem%;
    ssl_certificate_key %ssl_key%;
    ssl_stapling        on;
    ssl_stapling_verify on;

    include %home%/%user%/conf/web/%domain%/nginx.hsts.conf*;

    location / {
        # --- Configuración Dinámica JosSecurity ---
        set \$joss_port 9000;
        include %docroot%/nginx_port.conf*;
        # ------------------------------------------

        proxy_pass http://127.0.0.1:\$joss_port;
        
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;

        location ~* ^.+\.(%proxy_extensions%)\$ {
            root       %sdocroot%;
            try_files  \$uri @joss_app;
            access_log /var/log/%web_system%/domains/%domain%.log combined;
            expires    max;
        }
    }

    location @joss_app {
        set \$joss_port 9000;
        include %sdocroot%/nginx_port.conf*;
        
        proxy_pass http://127.0.0.1:\$joss_port;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
    }

    location /error/ {
        alias %home%/%user%/web/%domain%/document_errors/;
    }

    proxy_hide_header Upgrade;
    include %home%/%user%/conf/web/%domain%/nginx.ssl.conf_*;
}
EOF

echo -e "${GREEN} -> Plantillas creadas en $HESTIA_TPL_DIR${NC}"

# ------------------------------------------------------------------------------
# 3. CREAR JOSS LAUNCHER (Detector de Usuarios)
# ------------------------------------------------------------------------------
echo -e "${BLUE}[3/5] Creando Launcher Inteligente...${NC}"

cat > /usr/local/bin/joss-launcher <<EOF
#!/bin/bash
DOMAIN=\$1
JOSS_EXEC="$JOSS_BIN"

# Busca la ruta exacta del dominio en la estructura de Hestia
HOME_PATH=\$(ls -d /home/*/web/\$DOMAIN/public_html 2>/dev/null | head -n 1)

if [ -z "\$HOME_PATH" ]; then
    echo "ERROR: No se encuentra directorio para \$DOMAIN en /home/*/web/"
    exit 1
fi

# Extrae el usuario dueño de la carpeta
HESTIA_USER=\$(echo "\$HOME_PATH" | awk -F/ '{print \$3}')

echo ">>> Launcher: Iniciando \$DOMAIN (Usuario: \$HESTIA_USER)"
cd "\$HOME_PATH"

# Ejecuta Joss como el usuario correcto
exec runuser -u "\$HESTIA_USER" -- "\$JOSS_EXEC" run main.joss
EOF

chmod +x /usr/local/bin/joss-launcher

# ------------------------------------------------------------------------------
# 4. CONFIGURAR SYSTEMD
# ------------------------------------------------------------------------------
echo -e "${BLUE}[4/5] Configurando servicio Systemd...${NC}"

cat > /etc/systemd/system/joss@.service <<EOF
[Unit]
Description=JosSecurity Service for %i
After=network.target

[Service]
User=root
Group=root
ExecStart=/usr/local/bin/joss-launcher %i
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
echo -e "${GREEN} -> Servicio joss@.service registrado.${NC}"

# ------------------------------------------------------------------------------
# 5. CREAR HERRAMIENTA DE DESPLIEGUE (deploy-joss)
# ------------------------------------------------------------------------------
echo -e "${BLUE}[5/5] Creando comando deploy-joss...${NC}"

cat > /usr/local/bin/deploy-joss <<EOF
#!/bin/bash
DOMAIN=\$1

if [ -z "\$DOMAIN" ]; then
    echo "Uso: deploy-joss <dominio.com>"
    exit 1
fi

# Busca la ruta del proyecto
HOME_PATH=\$(ls -d /home/*/web/\$DOMAIN/public_html 2>/dev/null | head -n 1)

if [ -z "\$HOME_PATH" ]; then
    echo -e "${RED}Error: El dominio \$DOMAIN no existe en este servidor.${NC}"
    exit 1
fi

HESTIA_USER=\$(echo "\$HOME_PATH" | awk -F/ '{print \$3}')

echo -e "${BLUE}>>> Desplegando \$DOMAIN...${NC}"

# 1. Ajustar permisos (Vital para que el usuario pueda leer/escribir)
chown -R \$HESTIA_USER:\$HESTIA_USER "\$HOME_PATH"
chmod +x "\$HOME_PATH/main.joss" 2>/dev/null

# 2. Reiniciar el servicio específico del dominio
systemctl restart joss@\$DOMAIN
systemctl enable joss@\$DOMAIN

# 3. Recargar Nginx (por si cambió el puerto en nginx_port.conf)
systemctl reload nginx

echo -e "${GREEN}>>> ¡ÉXITO! Tu aplicación Joss está corriendo.${NC}"
systemctl status joss@\$DOMAIN --no-pager
EOF

chmod +x /usr/local/bin/deploy-joss

# ------------------------------------------------------------------------------
# FINALIZACIÓN
# ------------------------------------------------------------------------------
echo -e "\n${GREEN}==============================================${NC}"
echo -e "${GREEN}    INSTALACIÓN DE JOSSECURITY COMPLETADA     ${NC}"
echo -e "${GREEN}==============================================${NC}"
echo "Pasos para usar en un sitio nuevo:"
echo "1. Crea el dominio en HestiaCP."
echo "2. En la config del dominio, elige Plantilla Nginx -> 'joss'."
echo "3. Sube tu carpeta 'build' al public_html."
echo "4. Ejecuta: deploy-joss tudominio.com"
echo ""