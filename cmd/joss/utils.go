package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jossecurity/joss/pkg/core"
)

func GetEnvFile() string {
	return core.GetEnvFile()
}

func readEnvFile(path string) map[string]string {
	return core.ReadEnvFile(path)
}

func updateEnvFile(path, key, value string) {
	_ = core.UpdateEnvFile(path, key, value)
}

func printHelp(topics ...string) {
	if len(topics) > 0 {
		if strings.ToLower(topics[0]) == "plugins" || strings.ToLower(topics[0]) == "plugin" {
			targetPlugin := ""
			if len(topics) > 1 {
				targetPlugin = topics[1]
			}
			handlePluginHelp(targetPlugin)
			return
		}
		printTopicHelp(topics[0])
		return
	}

	fmt.Println("Joss — Lenguaje y Plataforma Moderna de Desarrollo")
	fmt.Println("Uso: joss <comando> [argumentos] [opciones]")
	fmt.Println()
	fmt.Println("EJECUCIÓN Y SERVIDOR:")
	fmt.Println("  server start                   Inicia el servidor web (requiere main.joss)")
	fmt.Println("  program start                  Inicia la aplicación en modo escritorio")
	fmt.Println("  run <archivo.joss>             Ejecuta un script Joss directamente")
	fmt.Println("  build [web|program|native]     Compila el proyecto para distribución")
	fmt.Println("    build native [os] [arch] [--gui]")
	fmt.Println()
	fmt.Println("CALIDAD DE CÓDIGO Y TOOLING:")
	fmt.Println("  test [ruta] [--filter=nombre]  Ejecuta la suite de pruebas unitarias (*_test.joss)")
	fmt.Println("  check [ruta]                   Verificación integral (formato, tipos, sintaxis y lint)")
	fmt.Println("  format [ruta] [--write|--check] Formatea archivos .joss según el estándar canónico")
	fmt.Println("  lint [ruta] [--json]           Análisis estático de estilo, tipos y seguridad")
	fmt.Println("  fix [ruta] [--dry-run]         Aplica correcciones automáticas seguras y formato")
	fmt.Println("  analyze [archivo]              Análisis semántico del AST (por defecto main.joss)")
	fmt.Println()
	fmt.Println("GENERADORES Y ESTRUCTURA (SCAFFOLDING):")
	fmt.Println("  new [web|console|package|plugin] <ruta>  Crea un nuevo proyecto o módulo")
	fmt.Println("  make:controller <Nombre>       Genera un controlador web")
	fmt.Println("  make:model <Nombre>            Genera un modelo de datos")
	fmt.Println("  make:view <Nombre>             Genera una plantilla de vista")
	fmt.Println("  make:middleware <Nombre>       Genera un middleware HTTP")
	fmt.Println("  make:mvc <Nombre>              Genera Modelo, Vista y Controlador en un paso")
	fmt.Println("  make:crud <Tabla>              Genera CRUD completo con rutas y vistas")
	fmt.Println("  remove:crud <Tabla>            Elimina un módulo CRUD generado")
	fmt.Println("  make:migration <Nombre>        Genera un archivo de migración")
	fmt.Println()
	fmt.Println("BASE DE DATOS Y STORAGE:")
	fmt.Println("  migrate                        Aplica migraciones pendientes")
	fmt.Println("  migrate:fresh                  Restablece y re-ejecuta todas las migraciones")
	fmt.Println("  db:seed                        Ejecuta seeders de app/database/seeders")
	fmt.Println("  change db [motor]              Cambia el motor de BD (sqlite, mysql, etc.)")
	fmt.Println("  change db migrate              Migra la conexión actual a un nuevo MySQL")
	fmt.Println("  change db prefix <prefijo>     Configura prefijo de tablas")
	fmt.Println("  userstorage [local|oci]        Configura almacenamiento de archivos")
	fmt.Println("  userstorage sync-oci|sync-local Sincroniza archivos con Oracle Cloud")
	fmt.Println()
	fmt.Println("PAQUETES Y PLUGINS:")
	fmt.Println("  help plugins [nombre]          Lista comandos y opciones provistos por plugins")
	fmt.Println("  pub <add|remove|install|publish|search|update>  Gestor de paquetes Joss")
	fmt.Println("  plugin compile <fuente/dir>    Compila código a paquete binario .jp")
	fmt.Println("  plugin inspect <archivo.jp>    Inspecciona bytecode y símbolos de un plugin")
	fmt.Println("  plugin verify <archivo.jp>     Verifica firmas Ed25519 e integridad de un plugin")
	fmt.Println("  package inspect <archivo.jp>   Inspecciona metadatos y firmas de un paquete")
	fmt.Println()
	fmt.Println("SISTEMA Y UTILIDADES:")
	fmt.Println("  version                        Muestra la versión de Joss instalada")
	fmt.Println("  update [-f|--canary|--stable]  Actualiza la versión de Joss, SDK y plugins")
	fmt.Println("  help [comando]                 Muestra ayuda detallada de un comando específico")
	fmt.Println()
	fmt.Println("Usa 'joss help <comando>' o 'joss help plugins' para ver opciones y ejemplos.")
}

func printTopicHelp(cmd string) {
	switch strings.ToLower(cmd) {
	case "format":
		fmt.Println("Uso: joss format [ruta] [opciones]")
		fmt.Println("Formatea archivos .joss según el estándar canónico del lenguaje.")
		fmt.Println("\nOpciones:")
		fmt.Println("  --write, -w    Escribe los cambios en el archivo (por defecto al pasar un archivo)")
		fmt.Println("  --check, -c    Verifica si los archivos están formateados sin modificarlos (útil en CI)")
		fmt.Println("\nEjemplos:")
		fmt.Println("  joss format app/controllers/UserController.joss")
		fmt.Println("  joss format --write .")
		fmt.Println("  joss format --check .")

	case "lint":
		fmt.Println("Uso: joss lint [ruta] [opciones]")
		fmt.Println("Ejecuta análisis estático y detecta problemas de estilo, tipos y seguridad.")
		fmt.Println("\nOpciones:")
		fmt.Println("  --json         Emite el reporte de diagnósticos en formato JSON estructurado")
		fmt.Println("\nEjemplos:")
		fmt.Println("  joss lint .")
		fmt.Println("  joss lint main.joss")
		fmt.Println("  joss lint --json .")

	case "fix":
		fmt.Println("Uso: joss fix [ruta] [opciones]")
		fmt.Println("Aplica correcciones automáticas seguras y formateo canónico.")
		fmt.Println("\nOpciones:")
		fmt.Println("  --dry-run, -d  Muestra qué cambios se aplicarían sin modificar los archivos")
		fmt.Println("\nEjemplos:")
		fmt.Println("  joss fix .")
		fmt.Println("  joss fix --dry-run .")

	case "check":
		fmt.Println("Uso: joss check [ruta]")
		fmt.Println("Ejecuta una verificación integral de calidad en 1 paso:")
		fmt.Println("  1. Comprobación de formato canónico")
		fmt.Println("  2. Chequeo de sintaxis y parseo")
		fmt.Println("  3. Análisis semántico estricto y tipos")
		fmt.Println("  4. Reglas de linter y seguridad")
		fmt.Println("\nEjemplos:")
		fmt.Println("  joss check .")

	case "test":
		fmt.Println("Uso: joss test [ruta] [opciones]")
		fmt.Println("Ejecuta la suite oficial de pruebas unitarias y aserciones (*_test.joss).")
		fmt.Println("\nOpciones:")
		fmt.Println("  --filter, -f   Filtra y ejecuta solo las pruebas cuyo nombre contenga el texto")
		fmt.Println("\nEjemplos:")
		fmt.Println("  joss test")
		fmt.Println("  joss test tests/")
		fmt.Println("  joss test --filter=login")

	case "server":
		fmt.Println("Uso: joss server start")
		fmt.Println("Inicia el servidor web ejecutando el punto de entrada 'main.joss'.")

	case "new":
		fmt.Println("Uso: joss new [web|console|package|plugin] <ruta/nombre>")
		fmt.Println("Genera una nueva estructura de proyecto.")
		fmt.Println("\nTipos:")
		fmt.Println("  web       Proyecto web MVC completo (por defecto)")
		fmt.Println("  console   Proyecto de consola CLI")
		fmt.Println("  package   Paquete distribuible para el ecosistema Joss")
		fmt.Println("  plugin    Plugin nativo con compilación a bytecode .jp")

	case "make:crud":
		fmt.Println("Uso: joss make:crud <Tabla>")
		fmt.Println("Genera automáticamente Modelo, Vistas, Controlador y Rutas para una tabla de base de datos.")

	case "migrate":
		fmt.Println("Uso: joss migrate")
		fmt.Println("Aplica las migraciones pendientes en app/database/migrations.")
		fmt.Println("Usa 'joss migrate:fresh' para reiniciar la base de datos y migrar desde cero.")

	case "pub":
		fmt.Println("Uso: joss pub <subcomando> [paquete]")
		fmt.Println("Gestor de dependencias y paquetes de Joss.")
		fmt.Println("\nSubcomandos:")
		fmt.Println("  add <paquete>      Añade una dependencia a joss.yaml e instálala")
		fmt.Println("  remove <paquete>   Elimina una dependencia de joss.yaml")
		fmt.Println("  install            Descarga e instala las dependencias de joss.yaml")
		fmt.Println("  update             Actualiza las dependencias instaladas")
		fmt.Println("  publish            Publica tu paquete en el registro oficial")

	default:
		fmt.Printf("No hay ayuda específica para el comando '%s'.\n\n", cmd)
		printHelp()
	}
}

// readLine reads a line of input from stdin in a platform-independent way, handling \r, \r\n and \n.
func readLine() string {
	reader := bufio.NewReader(os.Stdin)
	var line []byte
	for {
		b, err := reader.ReadByte()
		if err != nil {
			break
		}
		if b == '\n' {
			break
		}
		if b == '\r' {
			next, err := reader.Peek(1)
			if err == nil && next[0] == '\n' {
				_, _ = reader.ReadByte()
			}
			break
		}
		line = append(line, b)
	}
	return strings.TrimSpace(string(line))
}

func getCLIOption(flag string) string {
	flagPrefix := "--" + flag + "="
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, flagPrefix) {
			return strings.TrimPrefix(arg, flagPrefix)
		}
	}
	for i, arg := range os.Args[1:] {
		if arg == "--"+flag && i+1 < len(os.Args[1:]) {
			return os.Args[1:][i+1]
		}
	}
	return ""
}

func removeEnvKey(path, key string) {
	_ = core.RemoveEnvKey(path, key)
}
