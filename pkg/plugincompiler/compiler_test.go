package plugincompiler_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jossecurity/joss/pkg/plugincompiler"
	"github.com/jossecurity/joss/pkg/pluginpkg"
)

func TestJavaPluginCompilation(t *testing.T) {
	tempDir := t.TempDir()
	classFile := filepath.Join(tempDir, "MiPlugin.class")
	classBytes := []byte{0xCA, 0xFE, 0xBA, 0xBE, 0x00, 0x00, 0x00, 0x34}
	_ = os.WriteFile(classFile, classBytes, 0644)

	opts := plugincompiler.Options{
		SourceDir:   tempDir,
		Language:    "java",
		EntryFile:   classFile,
		Name:        "java-plugin",
		Version:     "1.0.0",
		Exports:     []string{"searchSong"},
		Permissions: []string{"network.http"},
	}

	outPath, _, err := plugincompiler.CompileProject(opts)
	if err != nil {
		t.Fatalf("error compilando plugin Java: %v", err)
	}

	verifyPlugin(t, outPath, "java-plugin", "java")
}

func TestPythonPluginCompilation(t *testing.T) {
	tempDir := t.TempDir()
	pyFile := filepath.Join(tempDir, "plugin.py")
	pyCode := "def search_song(query):\n    return 'song'\n\ndef internal_helper():\n    pass\n"
	_ = os.WriteFile(pyFile, []byte(pyCode), 0644)

	opts := plugincompiler.Options{
		SourceDir:   tempDir,
		Language:    "python",
		EntryFile:   pyFile,
		Name:        "py-plugin",
		Version:     "1.0.0",
		Exports:     []string{"search_song"},
		Permissions: []string{"filesystem.read"},
	}

	outPath, _, err := plugincompiler.CompileProject(opts)
	if err != nil {
		t.Fatalf("error compilando plugin Python: %v", err)
	}

	verifyPlugin(t, outPath, "py-plugin", "python")
}

func TestPHPPluginCompilation(t *testing.T) {
	tempDir := t.TempDir()
	phpFile := filepath.Join(tempDir, "plugin.php")
	phpCode := "<?php\nfunction process_payment($amount) {\n    return true;\n}\n"
	_ = os.WriteFile(phpFile, []byte(phpCode), 0644)

	opts := plugincompiler.Options{
		SourceDir:   tempDir,
		Language:    "php",
		EntryFile:   phpFile,
		Name:        "php-plugin",
		Version:     "1.0.0",
		Exports:     []string{"process_payment"},
		Permissions: []string{"network.http"},
	}

	outPath, _, err := plugincompiler.CompileProject(opts)
	if err != nil {
		t.Fatalf("error compilando plugin PHP: %v", err)
	}

	verifyPlugin(t, outPath, "php-plugin", "php")
}

func TestRustWasmPluginCompilation(t *testing.T) {
	tempDir := t.TempDir()
	wasmFile := filepath.Join(tempDir, "plugin.wasm")
	wasmBytes := []byte{0x00, 0x61, 0x73, 0x6D, 0x01, 0x00, 0x00, 0x00}
	_ = os.WriteFile(wasmFile, wasmBytes, 0644)

	opts := plugincompiler.Options{
		SourceDir:   tempDir,
		Language:    "rust",
		EntryFile:   wasmFile,
		Name:        "rust-plugin",
		Version:     "1.0.0",
		Exports:     []string{"fast_compute"},
		Permissions: []string{},
	}

	outPath, _, err := plugincompiler.CompileProject(opts)
	if err != nil {
		t.Fatalf("error compilando plugin Rust/Wasm: %v", err)
	}

	verifyPlugin(t, outPath, "rust-plugin", "rust")
}

func verifyPlugin(t *testing.T, jpPath, expectedName, expectedLang string) {
	data, err := os.ReadFile(jpPath)
	if err != nil {
		t.Fatalf("error leyendo .jp generado: %v", err)
	}

	archive, err := pluginpkg.Read(data)
	if err != nil {
		t.Fatalf("error al verificar firma y decodificar .jp: %v", err)
	}

	if archive.Metadata.Name != expectedName {
		t.Errorf("nombre esperado '%s', obtenido '%s'", expectedName, archive.Metadata.Name)
	}

	if archive.Metadata.Language != expectedLang {
		t.Errorf("lenguaje esperado '%s', obtenido '%s'", expectedLang, archive.Metadata.Language)
	}
}
