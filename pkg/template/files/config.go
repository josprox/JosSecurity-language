package files

import (
	"fmt"
	"path/filepath"

	"github.com/jossecurity/joss/pkg/version"
)

func GetConfigFiles(path string) map[string]string {
	return map[string]string{
		filepath.Join(path, "main.joss"): `public class Main {
    Init main() {
        print("Iniciando Sistema Joss...")
        Server::start()
    }
}`,
		filepath.Join(path, "env.joss"): `APP_ENV="development"
PORT="80"

# Database Configuration (sqlite or mysql)
DB="sqlite"
DB_PATH="database.sqlite"

# MySQL Configuration (Only if DB="mysql")
DB_HOST="localhost"
DB_NAME="joss_db"
DB_USER="root"
DB_PASS=""

# Redis Configuration (Optional)
# SESSION_DRIVER="redis"
# REDIS_HOST="localhost:6379"
# REDIS_PASSWORD=""

# Database Table Prefix
PREFIX="js_"

JWT_SECRET="change_me_in_production"

# Email Configuration (SMTP)
MAIL_HOST="smtp.gmail.com"
MAIL_PORT="587"
MAIL_USERNAME="your_email@gmail.com"
MAIL_PASSWORD="your_app_password"
MAIL_FROM_ADDRESS="no-reply@jossecurity.com"
MAIL_FROM_NAME="${APP_NAME}"

# storage
STORAGE="local"

# Configuración de Oracle cloud storage
OCI_NAMESPACE=""
OCI_BUCKET_NAME=""
OCI_TENANCY_ID=""
OCI_USER_ID=""
OCI_REGION=""
OCI_FINGERPRINT=""
OCI_PRIVATE_KEY_PATH=""
OCI_PASSPHRASE=""
`,
		filepath.Join(path, "config", "reglas.joss"): fmt.Sprintf(`// Constantes Globales
const string $APP_NAME = "Joss Enterprise"
const string $APP_VERSION = "%s"`, version.Version),
		filepath.Join(path, "joss.yaml"): fmt.Sprintf(`name: mi_proyecto
version: 1.0.0
environment:
  joss: ">=%s <%s"

dependencies:
`, version.Version, "4.0.0"),
		filepath.Join(path, "tests", "app_test.joss"): `test("Verificación de arranque del servidor", func() {
    bool $ready = true
    assertTrue($ready)
    assertEqual(1 + 1, 2)
})`,
		filepath.Join(path, ".gitignore"): `plugins/
env.joss
env.enc
database.sqlite
log.txt
`,
		filepath.Join(path, ".dockerignore"): `plugins/
env.joss
env.enc
database.sqlite
log.txt
.git/
.github/
`,
		filepath.Join(path, "entrypoint.sh"): `#!/bin/sh
set -e

ENV_FILE="/app/env.joss"
mkdir -p /app/storage

echo "[entrypoint] Generando env.joss..."
rm -f "$ENV_FILE"
touch "$ENV_FILE"

for var in $(env | cut -d= -f1); do
    case "$var" in
        PATH|HOME|HOSTNAME|TERM|SHLVL|PWD|_|OLDPWD|DEBIAN_FRONTEND|LANG)
            continue
            ;;
        *)
            val=$(eval echo "\$$var" | sed 's/"/\\"/g')
            echo "${var}=\"${val}\"" >> "$ENV_FILE"
            ;;
    esac
done

echo "[entrypoint] ✓ env.joss generado."
echo "[entrypoint] Iniciando servidor web Joss..."
exec joss server start
`,
		filepath.Join(path, "Dockerfile"): `# ============================================================
# Joss Web — Dockerfile (Debian Minimal + Joss Release)
# ============================================================

FROM debian:bookworm-slim

# Dependencias mínimas del sistema
RUN apt-get update && apt-get install -y --no-install-recommends \
        curl \
        ca-certificates \
        unzip \
        jq \
    && rm -rf /var/lib/apt/lists/*

# Descargar e instalar la versión release oficial de Joss CLI
RUN set -eux; \
    arch="$(dpkg --print-architecture)"; \
    releases_url="https://api.github.com/repos/joss-language/Joss-Programming-Language/releases"; \
    rel_json="$(curl -fsSL "${releases_url}" | jq -c '.[] | select(.draft == false)' | head -n 1)"; \
    asset_url="$(printf '%s' "${rel_json}" | jq -r ".assets[] | select(.name | contains(\"${arch}\")) | .browser_download_url" | head -n 1)"; \
    if [ -z "${asset_url}" ] || [ "${asset_url}" = "null" ]; then \
        asset_url="$(printf '%s' "${rel_json}" | jq -r '.assets[] | select(.name | contains("linux")) | .browser_download_url' | head -n 1)"; \
    fi; \
    echo "Descargando Joss CLI para ${arch} desde: ${asset_url}"; \
    curl -fsSL "${asset_url}" -o /tmp/joss_pkg.zip; \
    unzip -q /tmp/joss_pkg.zip -d /tmp/joss_out; \
    mv /tmp/joss_out/joss* /usr/local/bin/joss || mv /tmp/joss_out/*/joss* /usr/local/bin/joss; \
    rm -rf /tmp/joss_pkg.zip /tmp/joss_out; \
    chmod +x /usr/local/bin/joss; \
    joss version

WORKDIR /app

# Copiar el código del proyecto
COPY . .

# Instalar dependencias/plugins si joss.yaml las declara
RUN joss pub install || true

RUN chmod +x /app/entrypoint.sh

EXPOSE 80

ENTRYPOINT ["/app/entrypoint.sh"]
`,
	}
}
