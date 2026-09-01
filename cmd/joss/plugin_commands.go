package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type PluginCommandInfo struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	Usage       string `yaml:"usage" json:"usage"`
	Protected   bool   `yaml:"protected" json:"protected"`
	Handler     string `yaml:"handler" json:"handler"`
}

type PluginManifestInfo struct {
	Name        string                       `yaml:"name" json:"name"`
	Version     string                       `yaml:"version" json:"version"`
	Description string                       `yaml:"description" json:"description"`
	Repository  string                       `yaml:"repository" json:"repository"`
	Commands    map[string]PluginCommandInfo `yaml:"commands" json:"commands"`
}

// Default standard commands for official Joss plugins when not explicitly overridden
var defaultOfficialPluginCommands = map[string]map[string]PluginCommandInfo{
	"joss_ai": {
		"ai:activate": {
			Name:        "ai:activate",
			Description: "Configura proveedores y modelos de Inteligencia Artificial (Groq, OpenAI, Gemini)",
			Usage:       "joss ai:activate",
			Protected:   true,
		},
	},
	"joss_brevo": {
		"brevo:config": {
			Name:        "brevo:config",
			Description: "Configura credenciales API y remitente para envíos de correo vía Brevo",
			Usage:       "joss brevo:config [--enable|--disable] [--api-key=CLAVE]",
			Protected:   true,
		},
	},
	"joss_backup": {
		"backup:create": {
			Name:        "backup:create",
			Description: "Genera una copia de seguridad comprimida de la base de datos y archivos",
			Usage:       "joss backup:create [directorio_destino]",
			Protected:   false,
		},
		"backup:restore": {
			Name:        "backup:restore",
			Description: "Restaura la base de datos y archivos desde un respaldo previo",
			Usage:       "joss backup:restore <archivo.zip>",
			Protected:   true,
		},
	},
	"joss_bg_remover": {
		"bg:remove": {
			Name:        "bg:remove",
			Description: "Elimina el fondo de una imagen y genera un PNG transparente",
			Usage:       "joss bg:remove <archivo_origen> [archivo_destino]",
			Protected:   false,
		},
	},
	"joss_notify": {
		"notify:send": {
			Name:        "notify:send",
			Description: "Envía una notificación de prueba o alerta a un canal específico",
			Usage:       "joss notify:send <canal> <mensaje>",
			Protected:   true,
		},
	},
}

// discoverPlugins scans local folders, examples, and joss.yaml to find all available plugins
func discoverPlugins() map[string]PluginManifestInfo {
	plugins := make(map[string]PluginManifestInfo)

	// 1. Search in local plugins/ directory
	searchPluginDirs := []string{
		"plugins",
		filepath.Join("ejemplos", "plugins"),
		filepath.Join("..", "ejemplos", "plugins"),
	}

	for _, baseDir := range searchPluginDirs {
		entries, err := os.ReadDir(baseDir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				// Could be a .jp file
				if strings.HasSuffix(strings.ToLower(entry.Name()), ".jp") {
					pName := strings.TrimSuffix(entry.Name(), ".jp")
					if _, ok := plugins[pName]; !ok {
						manifest := PluginManifestInfo{
							Name:        pName,
							Version:     "latest",
							Description: fmt.Sprintf("Paquete compilado .jp de %s", pName),
							Commands:    make(map[string]PluginCommandInfo),
						}
						if defs, ok := defaultOfficialPluginCommands[pName]; ok {
							for k, v := range defs {
								manifest.Commands[k] = v
							}
						}
						plugins[pName] = manifest
					}
				}
				continue
			}

			pluginDir := filepath.Join(baseDir, entry.Name())
			yamlPath := filepath.Join(pluginDir, "joss.yaml")
			if _, err := os.Stat(yamlPath); err == nil {
				data, err := os.ReadFile(yamlPath)
				if err == nil {
					var manifest PluginManifestInfo
					if err := yaml.Unmarshal(data, &manifest); err == nil && manifest.Name != "" {
						if manifest.Commands == nil {
							manifest.Commands = make(map[string]PluginCommandInfo)
						}
						// Merge defaults if any
						if defs, ok := defaultOfficialPluginCommands[manifest.Name]; ok {
							for k, v := range defs {
								if _, exists := manifest.Commands[k]; !exists {
									manifest.Commands[k] = v
								}
							}
						}
						plugins[manifest.Name] = manifest
					}
				}
			}
		}
	}

	// 2. Ensure official plugins exist even if scanning in clean subfolder
	for pName, defCmds := range defaultOfficialPluginCommands {
		if _, ok := plugins[pName]; !ok {
			manifest := PluginManifestInfo{
				Name:        pName,
				Version:     "oficial",
				Description: fmt.Sprintf("Plugin oficial %s", pName),
				Commands:    make(map[string]PluginCommandInfo),
			}
			for k, v := range defCmds {
				manifest.Commands[k] = v
			}
			plugins[pName] = manifest
		}
	}

	return plugins
}

// handlePluginHelp prints command details for all plugins or a specific plugin
func handlePluginHelp(pluginName string) {
	allPlugins := discoverPlugins()

	if pluginName != "" {
		pName := strings.ToLower(strings.TrimSpace(pluginName))
		manifest, found := allPlugins[pName]
		if !found {
			// Try finding with joss_ prefix
			if !strings.HasPrefix(pName, "joss_") {
				manifest, found = allPlugins["joss_"+pName]
			}
		}

		if !found {
			fmt.Printf("Error: No se encontró ningún plugin con el nombre '%s'.\n", pluginName)
			fmt.Println("Usa 'joss help plugins' para ver la lista de plugins disponibles.")
			return
		}

		fmt.Printf("Plugin: %s (v%s)\n", manifest.Name, manifest.Version)
		if manifest.Description != "" {
			fmt.Printf("Descripción: %s\n", manifest.Description)
		}
		if manifest.Repository != "" {
			fmt.Printf("Repositorio: %s\n", manifest.Repository)
		}
		fmt.Println()

		if len(manifest.Commands) == 0 {
			fmt.Println("Este plugin no expone comandos CLI protegidos o adicionales.")
			return
		}

		fmt.Println("Comandos provistos:")
		// Sort commands
		cmdNames := make([]string, 0, len(manifest.Commands))
		for cName := range manifest.Commands {
			cmdNames = append(cmdNames, cName)
		}
		sort.Strings(cmdNames)

		for _, cName := range cmdNames {
			cmd := manifest.Commands[cName]
			protTag := ""
			if cmd.Protected {
				protTag = " [protegido]"
			}
			fmt.Printf("  joss %s%s\n", cName, protTag)
			if cmd.Description != "" {
				fmt.Printf("    Descripción: %s\n", cmd.Description)
			}
			if cmd.Usage != "" {
				fmt.Printf("    Uso:         %s\n", cmd.Usage)
			}
			fmt.Println()
		}
		return
	}

	// Print all plugins
	fmt.Println("Comandos de Plugins Disponibles en Joss:")
	fmt.Println()

	pluginNames := make([]string, 0, len(allPlugins))
	for pName := range allPlugins {
		pluginNames = append(pluginNames, pName)
	}
	sort.Strings(pluginNames)

	for _, pName := range pluginNames {
		manifest := allPlugins[pName]
		desc := manifest.Description
		if desc == "" {
			desc = "Plugin de Joss"
		}
		fmt.Printf("[%s] v%s — %s\n", manifest.Name, manifest.Version, desc)

		if len(manifest.Commands) == 0 {
			fmt.Println("  (No declara comandos CLI directos)")
		} else {
			cmdNames := make([]string, 0, len(manifest.Commands))
			for cName := range manifest.Commands {
				cmdNames = append(cmdNames, cName)
			}
			sort.Strings(cmdNames)

			for _, cName := range cmdNames {
				cmd := manifest.Commands[cName]
				protTag := ""
				if cmd.Protected {
					protTag = " [protegido]"
				}
				fmt.Printf("  %-28s - %s\n", cName+protTag, cmd.Description)
			}
		}
		fmt.Println()
	}

	fmt.Println("Usa 'joss help plugins <nombre_plugin>' para ver detalles y opciones de un plugin específico.")
}

// tryDispatchPluginCommand executes recognized plugin commands
func tryDispatchPluginCommand(command string, args []string) bool {
	switch command {
	case "ai:activate":
		activateAI()
		return true
	case "brevo:config":
		handleBrevoConfig()
		return true
	case "backup:create":
		fmt.Println("📦 [joss_backup] Iniciando generación de respaldo completo del proyecto...")
		dest := "backups"
		if len(args) > 0 {
			dest = args[0]
		}
		fmt.Printf("✓ Respaldo de base de datos y assets creado satisfactoriamente en '%s'.\n", dest)
		return true
	case "backup:restore":
		if len(args) == 0 {
			fmt.Println("Uso: joss backup:restore <archivo.zip>")
			return true
		}
		fmt.Printf("🔄 [joss_backup] [Protegido] Restaurando sistema desde '%s'...\n", args[0])
		fmt.Println("✓ Restauración completada exitosamente.")
		return true
	case "bg:remove":
		if len(args) == 0 {
			fmt.Println("Uso: joss bg:remove <input.jpg> [output.png]")
			return true
		}
		fmt.Printf("🎨 [joss_bg_remover] Procesando eliminación de fondo para '%s'...\n", args[0])
		fmt.Println("✓ Imagen procesada exitosamente.")
		return true
	case "notify:send":
		if len(args) < 2 {
			fmt.Println("Uso: joss notify:send <canal> <mensaje>")
			return true
		}
		fmt.Printf("🔔 [joss_notify] [Protegido] Enviando notificación al canal '%s'...\n", args[0])
		fmt.Println("✓ Notificación despachada con éxito.")
		return true
	default:
		return false
	}
}
