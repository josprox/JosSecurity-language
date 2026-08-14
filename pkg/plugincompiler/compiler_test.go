package plugincompiler_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jossecurity/joss/pkg/plugincompiler"
	"github.com/jossecurity/joss/pkg/plugincompiler/ir"
	"github.com/jossecurity/joss/pkg/plugincompiler/optimizer"
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

	outPath, result, err := plugincompiler.CompileProject(opts)
	if err != nil {
		t.Fatalf("error compilando plugin Java: %v", err)
	}

	if result.OptimizedFuncs == 0 {
		t.Errorf("esperado al menos 1 funcion optimizada, obtenido %d", result.OptimizedFuncs)
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

	outPath, result, err := plugincompiler.CompileProject(opts)
	if err != nil {
		t.Fatalf("error compilando plugin Python: %v", err)
	}

	if result.OptimizedFuncs != 1 {
		t.Errorf("Tree shaking debio conservar 1 funcion exportada, conservo %d", result.OptimizedFuncs)
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

func TestTreeShakingDeepCallGraph(t *testing.T) {
	module := ir.NewModule("graph_test", "1.0.0", "custom")
	module.Exports = []string{"mainExport"}

	// mainExport -> calls helperA
	module.Functions["mainExport"] = &ir.IRFunction{
		Name:       "mainExport",
		IsExported: true,
		Blocks: []*ir.IRBlock{
			{
				Label: "entry",
				Instructions: []ir.IRInstruction{
					{Op: ir.OpCallStatic, Args: []string{"helperA"}},
					{Op: ir.OpReturn},
				},
			},
		},
	}

	// helperA -> calls helperB and recursiveCall
	module.Functions["helperA"] = &ir.IRFunction{
		Name: "helperA",
		Blocks: []*ir.IRBlock{
			{
				Label: "entry",
				Instructions: []ir.IRInstruction{
					{Op: ir.OpCallStatic, Args: []string{"helperB"}},
					{Op: ir.OpCallStatic, Args: []string{"recursiveA"}},
					{Op: ir.OpReturn},
				},
			},
		},
	}

	// helperB -> terminal
	module.Functions["helperB"] = &ir.IRFunction{
		Name: "helperB",
		Blocks: []*ir.IRBlock{
			{Label: "entry", Instructions: []ir.IRInstruction{{Op: ir.OpReturn}}},
		},
	}

	// Mutual recursion: recursiveA <-> recursiveB
	module.Functions["recursiveA"] = &ir.IRFunction{
		Name: "recursiveA",
		Blocks: []*ir.IRBlock{
			{Label: "entry", Instructions: []ir.IRInstruction{{Op: ir.OpCallStatic, Args: []string{"recursiveB"}}, {Op: ir.OpReturn}}},
		},
	}
	module.Functions["recursiveB"] = &ir.IRFunction{
		Name: "recursiveB",
		Blocks: []*ir.IRBlock{
			{Label: "entry", Instructions: []ir.IRInstruction{{Op: ir.OpCallStatic, Args: []string{"recursiveA"}}, {Op: ir.OpReturn}}},
		},
	}

	// Unreachable functions: dead1 -> dead2
	module.Functions["dead1"] = &ir.IRFunction{
		Name: "dead1",
		Blocks: []*ir.IRBlock{
			{Label: "entry", Instructions: []ir.IRInstruction{{Op: ir.OpCallStatic, Args: []string{"dead2"}}, {Op: ir.OpReturn}}},
		},
	}
	module.Functions["dead2"] = &ir.IRFunction{
		Name: "dead2",
		Blocks: []*ir.IRBlock{
			{Label: "entry", Instructions: []ir.IRInstruction{{Op: ir.OpReturn}}},
		},
	}

	result, err := optimizer.TreeShake(module)
	if err != nil {
		t.Fatalf("tree shake fallo: %v", err)
	}

	if result.OptimizedFuncs != 5 {
		t.Errorf("esperadas 5 funciones alcanzables, obtenidas %d", result.OptimizedFuncs)
	}
	if result.RemovedFuncs != 2 {
		t.Errorf("esperadas 2 funciones eliminadas, obtenidas %d", result.RemovedFuncs)
	}

	if _, ok := result.Module.Functions["dead1"]; ok {
		t.Error("dead1 debio ser eliminada")
	}
	if _, ok := result.Module.Functions["dead2"]; ok {
		t.Error("dead2 debio ser eliminada")
	}
	if _, ok := result.Module.Functions["helperB"]; !ok {
		t.Error("helperB debio ser conservada por llamada transitiva")
	}
}

func TestIRValidationErrors(t *testing.T) {
	// 1. Modulo sin nombre
	m1 := ir.NewModule("", "1.0.0", "custom")
	m1.Exports = []string{"fn"}
	m1.Functions["fn"] = &ir.IRFunction{Name: "fn", Blocks: []*ir.IRBlock{{Instructions: []ir.IRInstruction{{Op: ir.OpReturn}}}}}
	if err := m1.Validate(); err == nil {
		t.Error("esperaba error por nombre vacio")
	}

	// 2. Exportacion inexistente
	m2 := ir.NewModule("test", "1.0.0", "custom")
	m2.Exports = []string{"missingFn"}
	if err := m2.Validate(); err == nil {
		t.Error("esperaba error por exportacion inexistente")
	}

	// 3. Exportacion duplicada
	m3 := ir.NewModule("test", "1.0.0", "custom")
	m3.Exports = []string{"fn", "fn"}
	m3.Functions["fn"] = &ir.IRFunction{Name: "fn", Blocks: []*ir.IRBlock{{Instructions: []ir.IRInstruction{{Op: ir.OpReturn}}}}}
	if err := m3.Validate(); err == nil {
		t.Error("esperaba error por exportacion duplicada")
	}

	// 4. Constante fuera de limites
	m4 := ir.NewModule("test", "1.0.0", "custom")
	m4.Exports = []string{"fn"}
	m4.Functions["fn"] = &ir.IRFunction{
		Name: "fn",
		Blocks: []*ir.IRBlock{
			{Instructions: []ir.IRInstruction{{Op: ir.OpConst, ConstIdx: 99}}},
		},
	}
	if err := m4.Validate(); err == nil {
		t.Error("esperaba error por ConstIdx fuera de límites")
	}
}

func TestCompilerDeterminism(t *testing.T) {
	tempDir := t.TempDir()
	pyFile := filepath.Join(tempDir, "deterministic.py")
	_ = os.WriteFile(pyFile, []byte("def calculate_total(a, b):\n    return a + b\n"), 0644)

	opts := plugincompiler.Options{
		SourceDir: tempDir,
		Language:  "python",
		EntryFile: pyFile,
		Name:      "det_plugin",
		Version:   "1.0.0",
		Exports:   []string{"calculate_total"},
	}

	out1, _, err := plugincompiler.CompileProject(opts)
	if err != nil {
		t.Fatalf("compilacion 1 fallo: %v", err)
	}
	data1, _ := os.ReadFile(out1)

	out2, _, err := plugincompiler.CompileProject(opts)
	if err != nil {
		t.Fatalf("compilacion 2 fallo: %v", err)
	}
	data2, _ := os.ReadFile(out2)

	if !bytes.Equal(data1, data2) {
		h1 := sha256.Sum256(data1)
		h2 := sha256.Sum256(data2)
		t.Fatalf("compilaciones consecutivas produjeron hashes diferentes (%x != %x)", h1, h2)
	}
}

func TestInvalidOptionsValidation(t *testing.T) {
	tempDir := t.TempDir()

	// Missing Name
	_, _, err := plugincompiler.CompileProject(plugincompiler.Options{
		SourceDir: tempDir,
		Language:  "python",
		Exports:   []string{"fn"},
	})
	if err == nil {
		t.Error("esperaba error por nombre de plugin vacio")
	}

	// Missing Exports
	_, _, err = plugincompiler.CompileProject(plugincompiler.Options{
		SourceDir: tempDir,
		Name:      "test-plugin",
		Language:  "python",
		Exports:   []string{},
	})
	if err == nil {
		t.Error("esperaba error por exports vacio")
	}

	// Unsupported Language
	_, _, err = plugincompiler.CompileProject(plugincompiler.Options{
		SourceDir: tempDir,
		Name:      "test-plugin",
		Language:  "cobol",
		Exports:   []string{"fn"},
	})
	if err == nil {
		t.Error("esperaba error por lenguaje no soportado")
	}
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

	if archive.Metadata.Signature == "" {
		t.Errorf("el archivo .jp no contiene firma Ed25519")
	}

	// Verificar indice de simbolos Schema 1
	symData, ok := archive.Files[pluginpkg.SymbolsPath]
	if !ok {
		t.Fatalf("el paquete no contiene %s", pluginpkg.SymbolsPath)
	}

	var symIndex pluginpkg.SymbolIndex
	if err := json.Unmarshal(symData, &symIndex); err != nil {
		t.Fatalf("error al deserializar SymbolIndex: %v", err)
	}

	if symIndex.Schema != pluginpkg.SymbolSchemaVersion {
		t.Errorf("schema de simbolos esperado %d, obtenido %d", pluginpkg.SymbolSchemaVersion, symIndex.Schema)
	}
	if symIndex.Package != expectedName {
		t.Errorf("package de simbolos esperado %s, obtenido %s", expectedName, symIndex.Package)
	}
}
