package plugincompiler

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jossecurity/joss/pkg/plugincompiler/backends/java"
	"github.com/jossecurity/joss/pkg/plugincompiler/backends/nativewasm"
	"github.com/jossecurity/joss/pkg/plugincompiler/backends/php"
	"github.com/jossecurity/joss/pkg/plugincompiler/backends/python"
	"github.com/jossecurity/joss/pkg/plugincompiler/codegen"
	"github.com/jossecurity/joss/pkg/plugincompiler/ir"
	"github.com/jossecurity/joss/pkg/plugincompiler/optimizer"
	"github.com/jossecurity/joss/pkg/pluginpkg"
)

// Options representa la configuracion para la compilacion de un plugin Joss.
type Options struct {
	SourceDir   string
	Language    string
	EntryFile   string
	Name        string
	Version     string
	Exports     []string
	Permissions []string
	MaxSizeMB   float64
}

// CompileProject analiza el proyecto fuente (Java/Python/PHP/Rust/C/Dart/etc), aplica Tree Shaking y compila a .jp.
func CompileProject(opts Options) (string, *optimizer.Result, error) {
	if opts.MaxSizeMB <= 0 {
		opts.MaxSizeMB = 1.0 // Objetivo predeterminado: 1 MB
	}

	var module *ir.IRModule
	var err error

	lang := strings.ToLower(opts.Language)

	switch lang {
	case "java", "kotlin":
		backend := java.NewJavaBackend()
		if filepath.Ext(opts.EntryFile) == ".jar" {
			module, err = backend.CompileFromJar(opts.EntryFile, opts.Name, opts.Version, opts.Exports, opts.Permissions)
		} else {
			module, err = backend.CompileFromClassFile(opts.EntryFile, opts.Name, opts.Version, opts.Exports, opts.Permissions)
		}
	case "python", "py":
		backend := python.NewPythonBackend()
		module, err = backend.Compile(opts.EntryFile, opts.Name, opts.Version, opts.Exports, opts.Permissions)
	case "php":
		backend := php.NewPHPBackend()
		module, err = backend.Compile(opts.EntryFile, opts.Name, opts.Version, opts.Exports, opts.Permissions)
	case "rust", "c", "cpp", "c++", "dart", "flutter", "wasm":
		backend := nativewasm.NewNativeWasmBackend(lang)
		module, err = backend.Compile(opts.EntryFile, opts.Name, opts.Version, opts.Exports, opts.Permissions)
	default:
		return "", nil, fmt.Errorf("compilador de plugins: lenguaje %q no soportado (soportados: java, python, php, rust, c, c++, dart, flutter, kotlin, wasm)", opts.Language)
	}

	if err != nil {
		return "", nil, fmt.Errorf("error en backend %s: %w", opts.Language, err)
	}

	// Aplicar Tree Shaking / Dead Code Elimination
	optRes, err := optimizer.TreeShake(module)
	if err != nil {
		return "", nil, fmt.Errorf("error durante optimizacion de codigo: %w", err)
	}

	// Generar Bytecode JPBC y simbolos
	cg := codegen.NewCodeGenerator()
	mainJBC, symbolsJSON, err := cg.GenerateBytecode(optRes.Module)
	if err != nil {
		return "", nil, fmt.Errorf("error generando bytecode JPBC: %w", err)
	}

	manifestYAML := fmt.Sprintf(`name: %s
version: %s
type: joss
language: %s
bytecode: jpbc
exports:
`, opts.Name, opts.Version, opts.Language)

	for _, exp := range opts.Exports {
		manifestYAML += fmt.Sprintf("  - %s\n", exp)
	}

	manifestYAML += "permissions:\n"
	for _, perm := range opts.Permissions {
		manifestYAML += fmt.Sprintf("  - %s\n", perm)
	}

	files := map[string][]byte{
		"joss.yaml":           []byte(manifestYAML),
		"bytecode/main.jbc":   mainJBC,
		pluginpkg.SymbolsPath: symbolsJSON,
	}

	metadata := pluginpkg.Metadata{
		Name:         opts.Name,
		Version:      opts.Version,
		Language:     opts.Language,
		Bytecode:     "bytecode/main.jbc",
		Exports:      opts.Exports,
		Permissions:  opts.Permissions,
		Symbols:      pluginpkg.SymbolsPath,
		Dependencies: make(map[string]string),
	}

	signingKey, _, err := loadOrCreatePluginSigningKey(opts.Name)
	if err != nil {
		return "", nil, fmt.Errorf("error al preparar firma del plugin: %w", err)
	}

	archiveBytes, err := pluginpkg.BuildSigned(metadata, files, signingKey)
	if err != nil {
		return "", nil, fmt.Errorf("error construyendo contenedor firmado .jp: %w", err)
	}

	outPath := filepath.Join(opts.SourceDir, opts.Name+".jp")
	if err := os.WriteFile(outPath, archiveBytes, 0644); err != nil {
		return "", nil, fmt.Errorf("error guardando paquete .jp: %w", err)
	}

	sizeMB := float64(len(archiveBytes)) / (1024 * 1024)
	if sizeMB > opts.MaxSizeMB {
		fmt.Printf("[Warning] El plugin %s sobrepasa el objetivo de %.2f MB (tamaño actual: %.2f MB)\n", opts.Name, opts.MaxSizeMB, sizeMB)
	}

	return outPath, optRes, nil
}

func loadOrCreatePluginSigningKey(name string) ([]byte, string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, "", err
	}
	keyDir := filepath.Join(home, ".joss", "keys")
	if err := os.MkdirAll(keyDir, 0700); err != nil {
		return nil, "", err
	}
	keyPath := filepath.Join(keyDir, name+".ed25519")
	if data, err := os.ReadFile(keyPath); err == nil && len(data) == ed25519.PrivateKeySize {
		return data, keyPath, nil
	}
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, "", fmt.Errorf("error generando llave Ed25519: %w", err)
	}
	_ = os.WriteFile(keyPath, privateKey, 0600)
	return privateKey, keyPath, nil
}
