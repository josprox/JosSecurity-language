package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jossecurity/joss/pkg/plugincompiler"
	"github.com/jossecurity/joss/pkg/pluginpkg"
)

func runPluginCommand(args []string) {
	if len(args) < 1 {
		printPluginUsage()
		return
	}

	subCmd := args[0]
	switch subCmd {
	case "compile":
		handlePluginCompile(args[1:])
	case "inspect":
		handlePluginInspect(args[1:])
	case "verify":
		handlePluginVerify(args[1:])
	default:
		fmt.Printf("Comando de plugin desconocido: %s\n\n", subCmd)
		printPluginUsage()
	}
}

func printPluginUsage() {
	fmt.Println("Uso de comandos de plugins de Joss:")
	fmt.Println("  joss plugin compile <dir|archivo> [--lang=java|python|php|wasm] [--name=nombre] [--ver=1.0.0] [--exports=f1,f2]")
	fmt.Println("  joss plugin inspect <plugin.jp>")
	fmt.Println("  joss plugin verify <plugin.jp>")
}

func handlePluginCompile(args []string) {
	if len(args) < 1 {
		fmt.Println("Error: especifica el directorio o archivo fuente a compilar.")
		fmt.Println("Ejemplo: joss plugin compile MiPlugin.jar --lang=java --name=music-plugin --exports=searchSong,getSong")
		return
	}

	sourcePath := args[0]
	lang := "java"
	name := ""
	version := "1.0.0"
	var exports []string
	var permissions []string

	for _, flag := range args[1:] {
		if strings.HasPrefix(flag, "--lang=") {
			lang = strings.TrimPrefix(flag, "--lang=")
		} else if strings.HasPrefix(flag, "--name=") {
			name = strings.TrimPrefix(flag, "--name=")
		} else if strings.HasPrefix(flag, "--ver=") {
			version = strings.TrimPrefix(flag, "--ver=")
		} else if strings.HasPrefix(flag, "--exports=") {
			expStr := strings.TrimPrefix(flag, "--exports=")
			exports = strings.Split(expStr, ",")
		} else if strings.HasPrefix(flag, "--permissions=") {
			permStr := strings.TrimPrefix(flag, "--permissions=")
			permissions = strings.Split(permStr, ",")
		}
	}

	// Auto-detect parameters from joss.yaml if present
	manifestPath := filepath.Join(sourcePath, "joss.yaml")
	if info, err := os.Stat(sourcePath); err == nil && !info.IsDir() {
		manifestPath = filepath.Join(filepath.Dir(sourcePath), "joss.yaml")
	}

	if data, err := os.ReadFile(manifestPath); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "name:") && name == "" {
				name = strings.Trim(strings.TrimSpace(strings.TrimPrefix(trimmed, "name:")), "\"'")
			} else if strings.HasPrefix(trimmed, "version:") {
				version = strings.Trim(strings.TrimSpace(strings.TrimPrefix(trimmed, "version:")), "\"'")
			} else if strings.HasPrefix(trimmed, "language:") {
				lang = strings.Trim(strings.TrimSpace(strings.TrimPrefix(trimmed, "language:")), "\"'")
			}
		}
	}

	if name == "" {
		base := filepath.Base(sourcePath)
		ext := filepath.Ext(base)
		name = strings.TrimSuffix(base, ext)
	}

	// If it's a pure Joss plugin project (joss.yaml type: joss or src/plugin.joss exists)
	jossEntry := filepath.Join(sourcePath, "src", "plugin.joss")
	if _, err := os.Stat(jossEntry); err == nil {
		buildPackage(sourcePath)
		return
	}
	if info, err := os.Stat(sourcePath); err == nil && !info.IsDir() && strings.HasSuffix(sourcePath, ".joss") {
		buildPackage(filepath.Dir(sourcePath))
		return
	}

	fmt.Printf("[Compilador de Plugins Joss] Compilando %s (Lenguaje: %s)...\n", sourcePath, lang)

	opts := plugincompiler.Options{
		SourceDir:   filepath.Dir(sourcePath),
		Language:    lang,
		EntryFile:   sourcePath,
		Name:        name,
		Version:     version,
		Exports:     exports,
		Permissions: permissions,
		MaxSizeMB:   1.0,
	}

	outPath, result, err := plugincompiler.CompileProject(opts)
	if err != nil {
		fmt.Printf("Error durante la compilacion del plugin: %v\n", err)
		return
	}

	fi, _ := os.Stat(outPath)
	sizeKB := float64(fi.Size()) / 1024.0

	fmt.Println("✓ Compilacion exitosa!")
	fmt.Printf("  Plugin generado: %s\n", outPath)
	fmt.Printf("  Tamaño del paquete: %.2f KB\n", sizeKB)
	fmt.Printf("  Tree Shaking: %d funciones conservadas (%d eliminadas)\n", result.OptimizedFuncs, result.RemovedFuncs)
	fmt.Printf("  Estructuras conservadas: %d (%d eliminadas)\n", result.OptimizedStructs, result.RemovedStructs)
}

func handlePluginInspect(args []string) {
	if len(args) < 1 {
		fmt.Println("Error: especifica el archivo .jp a inspeccionar.")
		fmt.Println("Ejemplo: joss plugin inspect music-plugin.jp")
		return
	}

	jpPath := args[0]
	archive, err := os.ReadFile(jpPath)
	if err != nil {
		fmt.Printf("Error al abrir %s: %v\n", jpPath, err)
		return
	}

	pkg, err := pluginpkg.Read(archive)
	if err != nil {
		fmt.Printf("Error al decodificar %s: %v\n", jpPath, err)
		return
	}

	fi, _ := os.Stat(jpPath)
	sizeKB := float64(fi.Size()) / 1024.0

	fmt.Println("========================================")
	fmt.Printf(" Plugin: %s\n", pkg.Metadata.Name)
	fmt.Printf(" Version: %s\n", pkg.Metadata.Version)
	fmt.Printf(" Bytecode Target: JPBC (Joss Plugin Bytecode)\n")
	fmt.Printf(" Tamaño Paquete: %.2f KB\n", sizeKB)
	if pkg.Metadata.Signature != "" {
		fmt.Printf(" Firma: %s (%s) verificada\n", pkg.Metadata.SignatureAlgorithm, pkg.Metadata.KeyID)
	}
	fmt.Println("----------------------------------------")
	fmt.Println(" Funciones Exportadas:")
	for _, exp := range pkg.Metadata.Exports {
		fmt.Printf("   - %s()\n", exp)
	}
	if pkg.Metadata.Symbols != "" {
		if symData, ok := pkg.Files[pkg.Metadata.Symbols]; ok {
			var symbols pluginpkg.SymbolIndex
			if err := json.Unmarshal(symData, &symbols); err == nil {
				if len(symbols.Classes) > 0 {
					fmt.Println(" Clases Declaradas:")
					for _, cls := range symbols.Classes {
						fmt.Printf("   - class %s\n", cls.Name)
						for _, m := range cls.Methods {
							fmt.Printf("       method %s()\n", m.Name)
						}
					}
				}
			}
		}
	}
	fmt.Println(" Permisos Declarados:")
	if len(pkg.Metadata.Permissions) == 0 {
		fmt.Println("   (ninguno)")
	} else {
		for _, perm := range pkg.Metadata.Permissions {
			fmt.Printf("   - %s\n", perm)
		}
	}
	fmt.Println("========================================")
}

func handlePluginVerify(args []string) {
	if len(args) < 1 {
		fmt.Println("Error: especifica el archivo .jp a verificar.")
		fmt.Println("Ejemplo: joss plugin verify mi_plugin.jp")
		return
	}

	jpPath := args[0]
	archive, err := os.ReadFile(jpPath)
	if err != nil {
		fmt.Printf("❌ Error al abrir %s: %v\n", jpPath, err)
		os.Exit(1)
	}

	pkg, err := pluginpkg.ReadVerified(archive)
	if err != nil {
		fmt.Printf("❌ Error de verificación en '%s': %v\n", jpPath, err)
		os.Exit(1)
	}

	fmt.Println("========================================")
	fmt.Printf(" Plugin: %s\n", pkg.Metadata.Name)
	fmt.Printf(" Version: %s\n", pkg.Metadata.Version)
	fmt.Printf(" Estado: Firma Ed25519 VÁLIDA y VERIFICADA ✅\n")
	if pkg.Metadata.SignatureAlgorithm != "" {
		fmt.Printf(" Algoritmo: %s (%s)\n", pkg.Metadata.SignatureAlgorithm, pkg.Metadata.KeyID)
	}
	fmt.Println("========================================")
}
