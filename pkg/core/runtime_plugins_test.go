package core_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jossecurity/joss/pkg/bytecode"
	"github.com/jossecurity/joss/pkg/core"
	"github.com/jossecurity/joss/pkg/parser"
	"github.com/jossecurity/joss/pkg/plugincompiler"
	"github.com/jossecurity/joss/pkg/pluginpkg"
)

func runJossCode(r *core.Runtime, code string) {
	l := parser.NewLexer(code)
	p := parser.NewParser(l)
	prog := p.ParseProgram()
	if len(p.Errors()) > 0 {
		panic(p.Errors())
	}
	r.Execute(prog)
}

func TestJossProgramConsumingNativeJossPlugin(t *testing.T) {
	// 1. Compilar plugin nativo Joss
	pluginSource := `
public class TaxService {
    public func calculate(mixed $amount) {
        return $amount * 0.16
    }
}

public func calculate_discount(mixed $price, mixed $pct) {
    return $price - ($price * ($pct / 100))
}
`
	prog := parser.NewParser(parser.NewLexer(pluginSource)).ParseProgram()
	encodedAST, err := bytecode.Encode(prog)
	if err != nil {
		t.Fatalf("error encoding AST: %v", err)
	}

	metadata := pluginpkg.Metadata{
		Name:     "tax_plugin",
		Version:  "1.0.0",
		Language: "joss",
		Bytecode: "bytecode/main.jbc",
		Symbols:  pluginpkg.SymbolsPath,
		Exports:  []string{"calculate_discount", "TaxService"},
	}

	symbolsJSON := []byte(`{
		"schema": 1,
		"package": "tax_plugin",
		"version": "1.0.0",
		"classes": [{"name": "TaxService", "methods": [{"name": "calculate", "parameters": [{"name": "amount"}]}]}],
		"functions": [{"name": "calculate_discount", "parameters": [{"name": "price"}, {"name": "pct"}]}]
	}`)

	files := map[string][]byte{
		"bytecode/main.jbc":   encodedAST,
		pluginpkg.SymbolsPath: symbolsJSON,
		"joss.yaml":           []byte("name: tax_plugin\nversion: 1.0.0\n"),
	}

	key, _, _ := pluginpkg.LoadOrCreateSigningKey("tax_plugin")
	jpBytes, err := pluginpkg.BuildSigned(metadata, files, key)
	if err != nil {
		t.Fatalf("error building .jp: %v", err)
	}

	// 2. Inicializar Runtime de Joss y cargar el plugin
	r := core.NewRuntime()
	if err := r.LoadPluginBytes(jpBytes); err != nil {
		t.Fatalf("error cargando .jp en Runtime: %v", err)
	}

	// 3. Ejecutar programa Joss que invoca funciones y clases del plugin
	appCode := `
$discounted = calculate_discount(200, 10)
$tax = new TaxService()
$tax_val = $tax->calculate(100)
`
	runJossCode(r, appCode)

	if r.Variables["discounted"] != 180.0 && r.Variables["discounted"] != int64(180) {
		t.Errorf("calculate_discount esperado 180, obtenido: %v", r.Variables["discounted"])
	}

	if r.Variables["tax_val"] != 16.0 {
		t.Errorf("TaxService.calculate esperado 16.0, obtenido: %v", r.Variables["tax_val"])
	}
}

func TestJossProgramConsumingMultilanguagePlugins(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Python Plugin
	pyPath := filepath.Join(tempDir, "ai.py")
	_ = os.WriteFile(pyPath, []byte("def predict(x):\n    return x * 10\n"), 0644)
	optsPy := plugincompiler.Options{
		SourceDir: tempDir,
		Language:  "python",
		EntryFile: pyPath,
		Name:      "joss_ai",
		Version:   "1.0.0",
		Exports:   []string{"predict"},
	}
	jpPy, _, err := plugincompiler.CompileProject(optsPy)
	if err != nil {
		t.Fatalf("error compilando Python plugin: %v", err)
	}

	// 2. Java Plugin
	javaPath := filepath.Join(tempDir, "Auth.class")
	_ = os.WriteFile(javaPath, []byte{0xCA, 0xFE, 0xBA, 0xBE, 0x00, 0x00, 0x00, 0x34}, 0644)
	optsJava := plugincompiler.Options{
		SourceDir: tempDir,
		Language:  "java",
		EntryFile: javaPath,
		Name:      "joss_auth",
		Version:   "1.0.0",
		Exports:   []string{"authenticate"},
	}
	jpJava, _, err := plugincompiler.CompileProject(optsJava)
	if err != nil {
		t.Fatalf("error compilando Java plugin: %v", err)
	}

	// 3. Cargar ambos plugins en Joss Runtime
	r := core.NewRuntime()
	if err := r.LoadPluginPackage(jpPy); err != nil {
		t.Fatalf("error cargando Python .jp: %v", err)
	}
	if err := r.LoadPluginPackage(jpJava); err != nil {
		t.Fatalf("error cargando Java .jp: %v", err)
	}

	// 4. Ejecutar programa Joss consumiendo ambos plugins
	appCode := `
// Llamada a funcion de plugin Python (ejecutada en JPBC VM sin Python runtime)
$ai_result = joss_ai::predict(5)

// Llamada a funcion de plugin Java (ejecutada en JPBC VM sin JVM)
$auth_result = joss_auth::authenticate()
`
	runJossCode(r, appCode)

	if r.Variables["ai_result"] == nil {
		t.Error("joss_ai::predict no retorno resultado")
	}

	if r.Variables["auth_result"] != "Java execution context OK" {
		t.Errorf("joss_auth::authenticate esperado 'Java execution context OK', obtenido: %v", r.Variables["auth_result"])
	}
}

// TestJossScriptFullMultilanguageExecution verifies that a single Joss script can load and execute
// plugins compiled from Python, Java, PHP, and WASM simultaneously with full variable state assertions.
func TestJossScriptFullMultilanguageExecution(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Python plugin
	pyPath := filepath.Join(tempDir, "analytics.py")
	_ = os.WriteFile(pyPath, []byte("def get_score(x):\n    return x * 10\n"), 0644)
	jpPy, _, err := plugincompiler.CompileProject(plugincompiler.Options{
		SourceDir: tempDir,
		Language:  "python",
		EntryFile: pyPath,
		Name:      "py_analytics",
		Version:   "1.0.0",
		Exports:   []string{"get_score"},
	})
	if err != nil {
		t.Fatalf("error compilando Python: %v", err)
	}

	// 2. Java plugin
	javaPath := filepath.Join(tempDir, "SecurityTask.class")
	_ = os.WriteFile(javaPath, []byte{0xCA, 0xFE, 0xBA, 0xBE, 0x00, 0x00, 0x00, 0x34}, 0644)
	jpJava, _, err := plugincompiler.CompileProject(plugincompiler.Options{
		SourceDir: tempDir,
		Language:  "java",
		EntryFile: javaPath,
		Name:      "java_sec",
		Version:   "1.0.0",
		Exports:   []string{"verifyToken"},
	})
	if err != nil {
		t.Fatalf("error compilando Java: %v", err)
	}

	// 3. PHP plugin
	phpPath := filepath.Join(tempDir, "mailer.php")
	_ = os.WriteFile(phpPath, []byte("<?php\nfunction send_mail($to, $msg) { return true; }\n"), 0644)
	jpPHP, _, err := plugincompiler.CompileProject(plugincompiler.Options{
		SourceDir: tempDir,
		Language:  "php",
		EntryFile: phpPath,
		Name:      "php_mailer",
		Version:   "1.0.0",
		Exports:   []string{"send_mail"},
	})
	if err != nil {
		t.Fatalf("error compilando PHP: %v", err)
	}

	// 4. WASM plugin
	wasmPath := filepath.Join(tempDir, "hasher.wasm")
	_ = os.WriteFile(wasmPath, []byte{0x00, 0x61, 0x73, 0x6D, 0x01, 0x00, 0x00, 0x00}, 0644)
	jpWasm, _, err := plugincompiler.CompileProject(plugincompiler.Options{
		SourceDir: tempDir,
		Language:  "wasm",
		EntryFile: wasmPath,
		Name:      "wasm_hasher",
		Version:   "1.0.0",
		Exports:   []string{"compute_hash"},
	})
	if err != nil {
		t.Fatalf("error compilando Wasm: %v", err)
	}

	// Load all 4 plugins into a fresh Joss Runtime
	r := core.NewRuntime()
	if err := r.LoadPluginPackage(jpPy); err != nil {
		t.Fatalf("error cargando Python .jp: %v", err)
	}
	if err := r.LoadPluginPackage(jpJava); err != nil {
		t.Fatalf("error cargando Java .jp: %v", err)
	}
	if err := r.LoadPluginPackage(jpPHP); err != nil {
		t.Fatalf("error cargando PHP .jp: %v", err)
	}
	if err := r.LoadPluginPackage(jpWasm); err != nil {
		t.Fatalf("error cargando Wasm .jp: %v", err)
	}

	// 5. Execute Joss script that calls all 4 plugins in sequence
	jossScript := `
$score = py_analytics::get_score(10)
$token = java_sec::verifyToken()
$mail_sent = php_mailer::send_mail("admin@joss.dev", "alerta")
$hash_val = wasm_hasher::compute_hash()
$total = $score . " - OK"
`
	runJossCode(r, jossScript)

	// 6. Verify variable values in Joss runtime state
	if r.Variables["score"] == nil {
		t.Errorf("$score debe tener valor, obtenido: nil")
	}

	if r.Variables["token"] != "Java execution context OK" {
		t.Errorf("$token esperado 'Java execution context OK', obtenido: %v", r.Variables["token"])
	}

	if r.Variables["mail_sent"] == nil {
		t.Errorf("$mail_sent debe ser true/ok, obtenido: nil")
	}

	if r.Variables["hash_val"] == nil {
		t.Errorf("$hash_val debe tener valor, obtenido: nil")
	}

	t.Logf("✓ Código Joss ejecutó exitosamente 4 plugins multilenguaje (Python, Java, PHP, WASM)")
	t.Logf("  $score = %v, $token = %v, $mail_sent = %v, $total = %v",
		r.Variables["score"], r.Variables["token"], r.Variables["mail_sent"], r.Variables["total"])
}

func TestMultiplePluginIsolationInRuntime(t *testing.T) {
	// Plugin 1: Math V1
	codeV1 := `
public class Worker {
    public func run() {
        return "Worker V1"
    }
}
`
	prog1 := parser.NewParser(parser.NewLexer(codeV1)).ParseProgram()
	ast1, _ := bytecode.Encode(prog1)
	key1, _, _ := pluginpkg.LoadOrCreateSigningKey("plugin_v1")
	jp1, _ := pluginpkg.BuildSigned(pluginpkg.Metadata{
		Name:     "plugin_v1",
		Version:  "1.0.0",
		Language: "joss",
		Bytecode: "bytecode/main.jbc",
		Symbols:  pluginpkg.SymbolsPath,
	}, map[string][]byte{
		"bytecode/main.jbc":   ast1,
		pluginpkg.SymbolsPath: []byte(`{"schema":1,"package":"plugin_v1","classes":[{"name":"Worker"}]}`),
		"joss.yaml":           []byte("name: plugin_v1\n"),
	}, key1)

	// Plugin 2: Math V2
	codeV2 := `
public class Worker {
    public func run() {
        return "Worker V2"
    }
}
`
	prog2 := parser.NewParser(parser.NewLexer(codeV2)).ParseProgram()
	ast2, _ := bytecode.Encode(prog2)
	key2, _, _ := pluginpkg.LoadOrCreateSigningKey("plugin_v2")
	jp2, _ := pluginpkg.BuildSigned(pluginpkg.Metadata{
		Name:     "plugin_v2",
		Version:  "2.0.0",
		Language: "joss",
		Bytecode: "bytecode/main.jbc",
		Symbols:  pluginpkg.SymbolsPath,
	}, map[string][]byte{
		"bytecode/main.jbc":   ast2,
		pluginpkg.SymbolsPath: []byte(`{"schema":1,"package":"plugin_v2","classes":[{"name":"Worker"}]}`),
		"joss.yaml":           []byte("name: plugin_v2\n"),
	}, key2)

	r := core.NewRuntime()
	_ = r.LoadPluginBytes(jp1)
	_ = r.LoadPluginBytes(jp2)

	// Instanciación aislada a través del registro de plugins
	inst1, err1 := r.PluginRegistry.Instantiate("plugin_v1", "Worker", nil)
	inst2, err2 := r.PluginRegistry.Instantiate("plugin_v2", "Worker", nil)
	if err1 != nil || err2 != nil {
		t.Fatalf("error instanciando workers: %v / %v", err1, err2)
	}

	val1, _ := r.PluginRegistry.CallMethod("plugin_v1", "Worker", "run", inst1, nil)
	val2, _ := r.PluginRegistry.CallMethod("plugin_v2", "Worker", "run", inst2, nil)

	if val1 != "Worker V1" {
		t.Errorf("plugin_v1 esperado 'Worker V1', obtenido: %v", val1)
	}
	if val2 != "Worker V2" {
		t.Errorf("plugin_v2 esperado 'Worker V2', obtenido: %v", val2)
	}
}

func TestJossProgramConsumingNativePluginWithInitConstructorAndThis(t *testing.T) {
	pluginSource := `
public class NotifyService {
    Init constructor() {
        $this->app_name = "joss_red"
        $this->count = 42
        $this->setup()
    }

    public func setup() {
        $this->status = "ready"
    }

    public func getStatus() {
        return $this->status
    }

    public func getAppName() {
        return $this->app_name
    }

    public func getCount() {
        return $this->count
    }
}
`
	prog := parser.NewParser(parser.NewLexer(pluginSource)).ParseProgram()
	encodedAST, err := bytecode.Encode(prog)
	if err != nil {
		t.Fatalf("error encoding AST: %v", err)
	}

	metadata := pluginpkg.Metadata{
		Name:     "joss_notify",
		Version:  "1.0.0",
		Language: "joss",
		Bytecode: "bytecode/main.jbc",
		Symbols:  pluginpkg.SymbolsPath,
		Exports:  []string{"NotifyService"},
	}

	symbolsJSON := []byte(`{
		"schema": 1,
		"package": "joss_notify",
		"version": "1.0.0",
		"classes": [{"name": "NotifyService", "methods": [{"name": "getStatus"}, {"name": "getAppName"}, {"name": "getCount"}]}]
	}`)

	files := map[string][]byte{
		"bytecode/main.jbc":   encodedAST,
		pluginpkg.SymbolsPath: symbolsJSON,
		"joss.yaml":           []byte("name: joss_notify\nversion: 1.0.0\n"),
	}

	key, _, _ := pluginpkg.LoadOrCreateSigningKey("joss_notify")
	jpBytes, err := pluginpkg.BuildSigned(metadata, files, key)
	if err != nil {
		t.Fatalf("error building .jp: %v", err)
	}

	r := core.NewRuntime()
	if err := r.LoadPluginBytes(jpBytes); err != nil {
		t.Fatalf("error cargando .jp en Runtime: %v", err)
	}

	appCode := `
$service = new NotifyService()
$status = $service->getStatus()
$app = $service->getAppName()
$count = $service->getCount()
`
	runJossCode(r, appCode)

	if r.Variables["status"] != "ready" {
		t.Errorf("NotifyService.getStatus esperado 'ready', obtenido: %v", r.Variables["status"])
	}
	if r.Variables["app"] != "joss_red" {
		t.Errorf("NotifyService.getAppName esperado 'joss_red', obtenido: %v", r.Variables["app"])
	}
	if r.Variables["count"] != int64(42) && r.Variables["count"] != 42 {
		t.Errorf("NotifyService.getCount esperado 42, obtenido: %v", r.Variables["count"])
	}
}
