package plugincompiler_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jossecurity/joss/pkg/bytecode"
	"github.com/jossecurity/joss/pkg/core"
	"github.com/jossecurity/joss/pkg/parser"
	"github.com/jossecurity/joss/pkg/pluginpkg"
)

func TestCompileAndLoadAllExamplePlugins(t *testing.T) {
	examplesDir := filepath.Join("..", "..", "ejemplos", "plugins")

	entries, err := os.ReadDir(examplesDir)
	if err != nil {
		t.Skipf("Omitiendo prueba: el directorio '%s' no existe en este entorno: %v", examplesDir, err)
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pluginDir := filepath.Join(examplesDir, entry.Name())
		entryFile := filepath.Join(pluginDir, "src", "plugin.joss")
		if _, err := os.Stat(entryFile); os.IsNotExist(err) {
			continue
		}

		t.Run(entry.Name(), func(t *testing.T) {
			data, err := os.ReadFile(entryFile)
			if err != nil {
				t.Fatalf("error leyendo %s: %v", entryFile, err)
			}

			l := parser.NewLexer(string(data))
			p := parser.NewParser(l)
			program := p.ParseProgram()
			if len(p.Errors()) > 0 {
				t.Fatalf("Errores de parseo en %s: %v", entryFile, p.Errors())
			}

			encodedAST, err := bytecode.Encode(program)
			if err != nil {
				t.Fatalf("error codificando AST para %s: %v", entry.Name(), err)
			}

			symbolIndex := pluginpkg.BuildSymbolIndex(program, entry.Name(), "1.0.0")
			symbolDataBytes, err := json.MarshalIndent(symbolIndex, "", "  ")
			if err != nil {
				t.Fatalf("error serializando indice de simbolos: %v", err)
			}

			files := map[string][]byte{
				"bytecode/main.jbc":   encodedAST,
				pluginpkg.SymbolsPath: symbolDataBytes,
				"joss.yaml":           []byte("name: " + entry.Name() + "\nversion: 1.0.0\n"),
			}

			metadata := pluginpkg.Metadata{
				Name:     entry.Name(),
				Version:  "1.0.0",
				Language: "joss",
				Bytecode: "bytecode/main.jbc",
				Symbols:  pluginpkg.SymbolsPath,
			}

			key, _, err := pluginpkg.LoadOrCreateSigningKey(entry.Name())
			if err != nil {
				t.Fatalf("error cargando llave de firma para %s: %v", entry.Name(), err)
			}

			jpBytes, err := pluginpkg.BuildSigned(metadata, files, key)
			if err != nil {
				t.Fatalf("error empaquetando plugin %s: %v", entry.Name(), err)
			}

			// Write updated .jp file back to plugin directory
			jpOutPath := filepath.Join(pluginDir, entry.Name()+".jp")
			if err := os.WriteFile(jpOutPath, jpBytes, 0644); err != nil {
				t.Fatalf("error guardando %s: %v", jpOutPath, err)
			}

			fi, _ := os.Stat(jpOutPath)
			sizeKB := float64(fi.Size()) / 1024.0
			t.Logf("✓ Plugin de ejemplo %s actualizado y verificado: %.2f KB", entry.Name(), sizeKB)

			// Verify plugin loads into Joss Runtime and registers symbols
			r := core.NewRuntime()
			if err := r.LoadPluginBytes(jpBytes); err != nil {
				t.Fatalf("Error cargando .jp de ejemplo %s en Joss Runtime: %v", entry.Name(), err)
			}
		})
	}
}
