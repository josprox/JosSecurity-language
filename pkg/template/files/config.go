package files

import "path/filepath"

func GetConfigFiles(path string) map[string]string {
	return map[string]string{
		filepath.Join(path, "main.joss"): `class Main {
    Init main() {
        print("Iniciando Sistema JosSecurity...")
        System.Run("joss", ["server", "start"])
    }
}`,
		filepath.Join(path, "env.joss"): `APP_ENV="development"
PORT="8000"

# Database Configuration (sqlite or mysql)
DB="sqlite"
DB_PATH="database.sqlite"

# MySQL Configuration (Only if DB="mysql")
DB_HOST="localhost"
DB_NAME="joss_db"
DB_USER="root"
DB_PASS=""

JWT_SECRET="change_me_in_production"`,
		filepath.Join(path, "config", "reglas.joss"): `// Constantes Globales
const string APP_NAME = "JosSecurity Enterprise"
const string APP_VERSION = "3.0.0"`,
	}
}
