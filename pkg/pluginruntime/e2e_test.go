package pluginruntime_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/jossecurity/joss/pkg/bytecode"
	"github.com/jossecurity/joss/pkg/parser"
	"github.com/jossecurity/joss/pkg/plugincompiler"
	"github.com/jossecurity/joss/pkg/plugincompiler/ir"
	"github.com/jossecurity/joss/pkg/pluginpkg"
	"github.com/jossecurity/joss/pkg/pluginruntime"
)

// TestE2ENativeJossFullFeature verifies comprehensive Joss plugin execution
func TestE2ENativeJossFullFeature(t *testing.T) {
	jossCode := `
public class Calculator {
    public func compute(mixed $mode, mixed $val) {
        return $mode == "double" ? $val * 2 : $val + 10
    }
}

public func helper_square(mixed $n) {
    return $n * $n
}

public func process_calculation(mixed $x) {
    $sq = helper_square($x)
    return $sq + 5
}
`
	l := parser.NewLexer(jossCode)
	p := parser.NewParser(l)
	prog := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	encodedAST, err := bytecode.Encode(prog)
	if err != nil {
		t.Fatalf("error encoding AST: %v", err)
	}

	metadata := pluginpkg.Metadata{
		Name:     "advanced_calc",
		Version:  "1.0.0",
		Language: "joss",
		Bytecode: "bytecode/main.jbc",
		Symbols:  pluginpkg.SymbolsPath,
		Exports:  []string{"process_calculation", "Calculator"},
	}

	files := map[string][]byte{
		"bytecode/main.jbc":   encodedAST,
		pluginpkg.SymbolsPath: []byte(`{"schema":1,"package":"advanced_calc"}`),
		"joss.yaml":           []byte("name: advanced_calc\nversion: 1.0.0\n"),
	}

	key, _, err := pluginpkg.LoadOrCreateSigningKey("advanced_calc")
	if err != nil {
		t.Fatalf("error creating signing key: %v", err)
	}

	jpBytes, err := pluginpkg.BuildSigned(metadata, files, key)
	if err != nil {
		t.Fatalf("error packaging signed .jp: %v", err)
	}

	// 1. Cargar y verificar
	engine := &MockASTEngine{}
	plugin, err := pluginruntime.LoadPluginWithEngine(jpBytes, engine)
	if err != nil {
		t.Fatalf("error loading plugin: %v", err)
	}

	reg := pluginruntime.NewPluginRegistry(nil)
	if err := reg.Register(plugin); err != nil {
		t.Fatalf("error registering plugin: %v", err)
	}

	// 2. Ejecutar función que invoca función auxiliar (process_calculation -> helper_square)
	// (4 * 4) + 5 = 21
	res, err := reg.CallFunction("advanced_calc", "process_calculation", []interface{}{4})
	if err != nil {
		t.Fatalf("error calling process_calculation: %v", err)
	}
	if res != int64(21) && res != 21 && res != 21.0 {
		t.Errorf("esperado 21, obtenido %v", res)
	}

	// 3. Instanciar clase y ejecutar método con lógica condicional
	calcInst, err := reg.Instantiate("advanced_calc", "Calculator", nil)
	if err != nil {
		t.Fatalf("error instantiating Calculator: %v", err)
	}

	resDouble, err := reg.CallMethod("advanced_calc", "Calculator", "compute", calcInst, []interface{}{"double", 50})
	if err != nil {
		t.Fatalf("error calling compute: %v", err)
	}
	if resDouble != int64(100) && resDouble != 100 && resDouble != 100.0 {
		t.Errorf("esperado 100, obtenido %v", resDouble)
	}

	resAdd, err := reg.CallMethod("advanced_calc", "Calculator", "compute", calcInst, []interface{}{"other", 50})
	if err != nil {
		t.Fatalf("error calling compute: %v", err)
	}
	if resAdd != int64(60) && resAdd != 60 && resAdd != 60.0 {
		t.Errorf("esperado 60, obtenido %v", resAdd)
	}
}

// TestPluginIsolation verifies that two independent plugins do not contaminate each other's state
func TestPluginIsolation(t *testing.T) {
	// Plugin 1: State A
	codeA := `
public class Service {
    public func value() {
        return "Service A"
    }
}
`
	progA := parser.NewParser(parser.NewLexer(codeA)).ParseProgram()
	astA, _ := bytecode.Encode(progA)
	keyA, _, _ := pluginpkg.LoadOrCreateSigningKey("plugin_a")
	jpA, _ := pluginpkg.BuildSigned(pluginpkg.Metadata{
		Name:     "plugin_a",
		Version:  "1.0.0",
		Language: "joss",
		Bytecode: "bytecode/main.jbc",
	}, map[string][]byte{"bytecode/main.jbc": astA, "joss.yaml": []byte("name: plugin_a\n")}, keyA)

	// Plugin 2: State B (same class name 'Service', different implementation)
	codeB := `
public class Service {
    public func value() {
        return "Service B"
    }
}
`
	progB := parser.NewParser(parser.NewLexer(codeB)).ParseProgram()
	astB, _ := bytecode.Encode(progB)
	keyB, _, _ := pluginpkg.LoadOrCreateSigningKey("plugin_b")
	jpB, _ := pluginpkg.BuildSigned(pluginpkg.Metadata{
		Name:     "plugin_b",
		Version:  "1.0.0",
		Language: "joss",
		Bytecode: "bytecode/main.jbc",
	}, map[string][]byte{"bytecode/main.jbc": astB, "joss.yaml": []byte("name: plugin_b\n")}, keyB)

	reg := pluginruntime.NewPluginRegistry(nil)
	pA, errA := pluginruntime.LoadPluginWithEngine(jpA, &MockASTEngine{Tag: "Service A"})
	pB, errB := pluginruntime.LoadPluginWithEngine(jpB, &MockASTEngine{Tag: "Service B"})
	if errA != nil || errB != nil {
		t.Fatalf("error cargando plugins: %v / %v", errA, errB)
	}

	_ = reg.Register(pA)
	_ = reg.Register(pB)

	instA, _ := reg.Instantiate("plugin_a", "Service", nil)
	instB, _ := reg.Instantiate("plugin_b", "Service", nil)

	valA, _ := reg.CallMethod("plugin_a", "Service", "value", instA, nil)
	valB, _ := reg.CallMethod("plugin_b", "Service", "value", instB, nil)

	if valA != "Service A" {
		t.Errorf("Plugin A esperado 'Service A', obtenido: %v", valA)
	}
	if valB != "Service B" {
		t.Errorf("Plugin B esperado 'Service B', obtenido: %v", valB)
	}
}

// TestInfiniteLoopProtection verifies VM halts when instruction step limit is reached
func TestInfiniteLoopProtection(t *testing.T) {
	loopMod := &pluginruntime.JPBCModule{
		MajorVersion: 1,
		MinorVersion: 0,
		Functions: map[string]*pluginruntime.JPBCFunction{
			"infinite_loop": {
				Name: "infinite_loop",
				Instructions: []pluginruntime.JPBCInstruction{
					{Op: ir.OpNop},
					{Op: ir.OpBranch, ConstIdx: 0}, // Jump to inst 0 indefinitely
				},
			},
		},
	}

	vm := pluginruntime.NewJPBCVM(loopMod, nil, nil)
	_, err := vm.Execute("infinite_loop", nil)
	if err == nil {
		t.Fatal("esperaba error de limite de ejecucion por bucle infinito")
	}

	if !errors.Is(err, pluginruntime.ErrExecutionLimitExceeded) {
		t.Errorf("esperado ErrExecutionLimitExceeded, obtenido: %v", err)
	}
}

// TestMultilanguageAllPipelinesE2E tests real compilation through compiler backends to runtime
func TestMultilanguageAllPipelinesE2E(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Python Pipeline
	pyPath := filepath.Join(tempDir, "analytics.py")
	_ = os.WriteFile(pyPath, []byte("def get_score(val):\n    return val + 100\n"), 0644)
	optsPy := plugincompiler.Options{
		SourceDir: tempDir,
		Language:  "python",
		EntryFile: pyPath,
		Name:      "py_analytics",
		Version:   "1.0.0",
		Exports:   []string{"get_score"},
	}
	jpPy, _, err := plugincompiler.CompileProject(optsPy)
	if err != nil {
		t.Fatalf("error compilando Python: %v", err)
	}
	plugPy, err := pluginruntime.LoadPluginFromFile(jpPy)
	if err != nil {
		t.Fatalf("error cargando .jp Python: %v", err)
	}

	// 2. Java Pipeline
	javaClassPath := filepath.Join(tempDir, "SecurityTask.class")
	_ = os.WriteFile(javaClassPath, []byte{0xCA, 0xFE, 0xBA, 0xBE, 0x00, 0x00, 0x00, 0x34}, 0644)
	optsJava := plugincompiler.Options{
		SourceDir: tempDir,
		Language:  "java",
		EntryFile: javaClassPath,
		Name:      "java_sec",
		Version:   "1.0.0",
		Exports:   []string{"verifyToken"},
	}
	jpJava, _, err := plugincompiler.CompileProject(optsJava)
	if err != nil {
		t.Fatalf("error compilando Java: %v", err)
	}
	plugJava, err := pluginruntime.LoadPluginFromFile(jpJava)
	if err != nil {
		t.Fatalf("error cargando .jp Java: %v", err)
	}

	// 3. PHP Pipeline
	phpPath := filepath.Join(tempDir, "mailer.php")
	_ = os.WriteFile(phpPath, []byte("<?php\nfunction send_mail($to, $msg) { return true; }\n"), 0644)
	optsPHP := plugincompiler.Options{
		SourceDir: tempDir,
		Language:  "php",
		EntryFile: phpPath,
		Name:      "php_mailer",
		Version:   "1.0.0",
		Exports:   []string{"send_mail"},
	}
	jpPHP, _, err := plugincompiler.CompileProject(optsPHP)
	if err != nil {
		t.Fatalf("error compilando PHP: %v", err)
	}
	plugPHP, err := pluginruntime.LoadPluginFromFile(jpPHP)
	if err != nil {
		t.Fatalf("error cargando .jp PHP: %v", err)
	}

	// 4. Wasm Pipeline
	wasmPath := filepath.Join(tempDir, "hasher.wasm")
	_ = os.WriteFile(wasmPath, []byte{0x00, 0x61, 0x73, 0x6D, 0x01, 0x00, 0x00, 0x00}, 0644)
	optsWasm := plugincompiler.Options{
		SourceDir: tempDir,
		Language:  "wasm",
		EntryFile: wasmPath,
		Name:      "wasm_hasher",
		Version:   "1.0.0",
		Exports:   []string{"compute_hash"},
	}
	jpWasm, _, err := plugincompiler.CompileProject(optsWasm)
	if err != nil {
		t.Fatalf("error compilando Wasm: %v", err)
	}
	plugWasm, err := pluginruntime.LoadPluginFromFile(jpWasm)
	if err != nil {
		t.Fatalf("error cargando .jp Wasm: %v", err)
	}

	// Registrar todos en el Registry y ejecutar
	reg := pluginruntime.NewPluginRegistry(nil)
	_ = reg.Register(plugPy)
	_ = reg.Register(plugJava)
	_ = reg.Register(plugPHP)
	_ = reg.Register(plugWasm)

	// Invocaciones reales sobre la VM JPBC
	resPy, err := reg.CallFunction("py_analytics", "get_score", []interface{}{50})
	if err != nil || resPy == nil {
		t.Errorf("fallo invocacion Python: %v", err)
	}

	resJava, err := reg.CallFunction("java_sec", "verifyToken", nil)
	if err != nil || resJava == nil {
		t.Errorf("fallo invocacion Java: %v", err)
	}

	resPHP, err := reg.CallFunction("php_mailer", "send_mail", []interface{}{"user@example.com", "hello"})
	if err != nil || resPHP == nil {
		t.Errorf("fallo invocacion PHP: %v", err)
	}

	resWasm, err := reg.CallFunction("wasm_hasher", "compute_hash", nil)
	if err != nil || resWasm == nil {
		t.Errorf("fallo invocacion Wasm: %v", err)
	}
}

// TestMultilanguagePluginSizeAndExecution verifies that plugins generated from external languages
// produce lightweight .jp packages (in KB, not MB) and execute standalone in Joss without external runtimes.
func TestMultilanguagePluginSizeAndExecution(t *testing.T) {
	tempDir := t.TempDir()

	languages := []struct {
		lang       string
		filename   string
		content    []byte
		funcName   string
		pluginName string
	}{
		{
			lang:       "python",
			filename:   "math_utils.py",
			content:    []byte("def double_val(x):\n    return x * 2\n"),
			funcName:   "double_val",
			pluginName: "py_math",
		},
		{
			lang:       "java",
			filename:   "CryptoHelper.class",
			content:    []byte{0xCA, 0xFE, 0xBA, 0xBE, 0x00, 0x00, 0x00, 0x34},
			funcName:   "encrypt_data",
			pluginName: "java_crypto",
		},
		{
			lang:       "php",
			filename:   "string_helper.php",
			content:    []byte("<?php\nfunction sanitize_str($s) { return $s; }\n"),
			funcName:   "sanitize_str",
			pluginName: "php_str",
		},
		{
			lang:       "wasm",
			filename:   "fast_hash.wasm",
			content:    []byte{0x00, 0x61, 0x73, 0x6D, 0x01, 0x00, 0x00, 0x00},
			funcName:   "compute_fast_hash",
			pluginName: "wasm_hash",
		},
	}

	reg := pluginruntime.NewPluginRegistry(nil)

	for _, tt := range languages {
		srcPath := filepath.Join(tempDir, tt.filename)
		if err := os.WriteFile(srcPath, tt.content, 0644); err != nil {
			t.Fatalf("error escribiendo %s: %v", tt.filename, err)
		}

		opts := plugincompiler.Options{
			SourceDir: tempDir,
			Language:  tt.lang,
			EntryFile: srcPath,
			Name:      tt.pluginName,
			Version:   "1.0.0",
			Exports:   []string{tt.funcName},
		}

		jpPath, result, err := plugincompiler.CompileProject(opts)
		if err != nil {
			t.Fatalf("error compilando plugin %s (%s): %v", tt.pluginName, tt.lang, err)
		}

		// Verify size is in KB (must be under 50 KB, not MB)
		fi, err := os.Stat(jpPath)
		if err != nil {
			t.Fatalf("error obteniendo stat de %s: %v", jpPath, err)
		}
		sizeKB := float64(fi.Size()) / 1024.0
		if sizeKB > 50.0 {
			t.Errorf("Plugin %s (%s) excede el tamaño máximo deseado (%.2f KB > 50.0 KB)", tt.pluginName, tt.lang, sizeKB)
		}
		t.Logf("✓ Plugin %s (%s) compilado exitosamente: %.2f KB (Tree shaking: %d funcs)", tt.pluginName, tt.lang, sizeKB, result.OptimizedFuncs)

		// Load and execute standalone in Joss JPBC VM without external language runtime
		plug, err := pluginruntime.LoadPluginFromFile(jpPath)
		if err != nil {
			t.Fatalf("error cargando .jp %s (%s): %v", tt.pluginName, tt.lang, err)
		}

		if err := reg.Register(plug); err != nil {
			t.Fatalf("error registrando plugin %s: %v", tt.pluginName, err)
		}

		res, err := reg.CallFunction(tt.pluginName, tt.funcName, []interface{}{int64(42)})
		if err != nil {
			t.Errorf("Error invocando %s.%s: %v", tt.pluginName, tt.funcName, err)
		} else if res == nil {
			t.Errorf("Invocacion de %s.%s devolvio nil", tt.pluginName, tt.funcName)
		}
	}
}
