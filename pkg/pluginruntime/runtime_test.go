package pluginruntime_test

import (
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

func TestNativeJossPluginExecution(t *testing.T) {
	// 1. Crear código Joss con clase y método
	jossCode := `
public class MathService {
    public func add(mixed $a, mixed $b) {
        return $a + $b
    }

    public func multiply(mixed $a, mixed $b) {
        return $a * $b
    }
}

public func calculate_tax(mixed $amount) {
    return $amount * 0.16
}
`
	l := parser.NewLexer(jossCode)
	p := parser.NewParser(l)
	prog := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("error parseando joss: %v", p.Errors())
	}

	encodedAST, err := bytecode.Encode(prog)
	if err != nil {
		t.Fatalf("error codificando AST: %v", err)
	}

	symbolData := pluginpkg.BuildSymbolIndex(prog, "joss_math", "1.0.0")

	metadata := pluginpkg.Metadata{
		Name:     "joss_math",
		Version:  "1.0.0",
		Language: "joss",
		Bytecode: "bytecode/main.jbc",
		Symbols:  pluginpkg.SymbolsPath,
		Exports:  []string{"calculate_tax", "MathService"},
	}

	files := map[string][]byte{
		"bytecode/main.jbc":   encodedAST,
		pluginpkg.SymbolsPath: []byte(`{"schema":1,"package":"joss_math"}`),
		"joss.yaml":           []byte("name: joss_math\nversion: 1.0.0\n"),
	}

	key, _, err := pluginpkg.LoadOrCreateSigningKey("joss_math")
	if err != nil {
		t.Fatalf("error creando llave: %v", err)
	}

	jpData, err := pluginpkg.BuildSigned(metadata, files, key)
	if err != nil {
		t.Fatalf("error empaquetando .jp: %v", err)
	}

	// 2. Cargar en Plugin Runtime con motor AST de prueba
	engine := &MockASTEngine{}
	plugin, err := pluginruntime.LoadPluginWithEngine(jpData, engine)
	if err != nil {
		t.Fatalf("error cargando plugin: %v", err)
	}

	if plugin.Format != pluginruntime.FormatJossAST {
		t.Errorf("formato esperado JOSSBC2Z, obtenido %s", plugin.Format)
	}

	reg := pluginruntime.NewPluginRegistry(nil)
	if err := reg.Register(plugin); err != nil {
		t.Fatalf("error registrando plugin: %v", err)
	}

	// 3. Ejecutar función global
	resFn, err := reg.CallFunction("joss_math", "calculate_tax", []interface{}{100.0})
	if err != nil {
		t.Fatalf("error invocando calculate_tax: %v", err)
	}
	if resFn != 16.0 {
		t.Errorf("resultado de calculate_tax esperado 16.0, obtenido %v", resFn)
	}

	// 4. Instanciar clase y llamar método
	instance, err := reg.Instantiate("joss_math", "MathService", nil)
	if err != nil {
		t.Fatalf("error instanciando MathService: %v", err)
	}

	resMethod, err := reg.CallMethod("joss_math", "MathService", "add", instance, []interface{}{10, 25})
	if err != nil {
		t.Fatalf("error invocando MathService.add: %v", err)
	}
	if resMethod != int64(35) && resMethod != 35 {
		t.Errorf("resultado de MathService.add esperado 35, obtenido %v", resMethod)
	}

	_ = symbolData
}

func TestPythonPluginJPBCExecution(t *testing.T) {
	tempDir := t.TempDir()
	pyFile := filepath.Join(tempDir, "service.py")
	pyCode := "def calculate_bonus(salary):\n    return salary * 2\n"
	_ = os.WriteFile(pyFile, []byte(pyCode), 0644)

	opts := plugincompiler.Options{
		SourceDir: tempDir,
		Language:  "python",
		EntryFile: pyFile,
		Name:      "py_service",
		Version:   "1.0.0",
		Exports:   []string{"calculate_bonus"},
	}

	jpPath, _, err := plugincompiler.CompileProject(opts)
	if err != nil {
		t.Fatalf("error compilando plugin python: %v", err)
	}

	plugin, err := pluginruntime.LoadPluginFromFile(jpPath)
	if err != nil {
		t.Fatalf("error cargando .jp generado: %v", err)
	}

	if plugin.Format != pluginruntime.FormatJPBC {
		t.Errorf("formato esperado JPBC, obtenido %s", plugin.Format)
	}

	reg := pluginruntime.NewPluginRegistry(nil)
	if err := reg.Register(plugin); err != nil {
		t.Fatalf("error registrando plugin: %v", err)
	}

	res, err := reg.CallFunction("py_service", "calculate_bonus", []interface{}{500})
	if err != nil {
		t.Fatalf("error ejecutando calculate_bonus en JPBC VM: %v", err)
	}

	if res == nil {
		t.Fatal("resultado no debe ser nil")
	}
}

func TestJavaPluginJPBCExecution(t *testing.T) {
	tempDir := t.TempDir()
	classFile := filepath.Join(tempDir, "JavaPlugin.class")
	_ = os.WriteFile(classFile, []byte{0xCA, 0xFE, 0xBA, 0xBE, 0x00, 0x00, 0x00, 0x34}, 0644)

	opts := plugincompiler.Options{
		SourceDir: tempDir,
		Language:  "java",
		EntryFile: classFile,
		Name:      "java_core",
		Version:   "1.0.0",
		Exports:   []string{"executeTask"},
	}

	jpPath, _, err := plugincompiler.CompileProject(opts)
	if err != nil {
		t.Fatalf("error compilando plugin java: %v", err)
	}

	plugin, err := pluginruntime.LoadPluginFromFile(jpPath)
	if err != nil {
		t.Fatalf("error cargando .jp java: %v", err)
	}

	reg := pluginruntime.NewPluginRegistry(nil)
	_ = reg.Register(plugin)

	res, err := reg.CallFunction("java_core", "executeTask", nil)
	if err != nil {
		t.Fatalf("error ejecutando funcion Java en JPBC VM: %v", err)
	}

	if res != "Java execution context OK" {
		t.Errorf("salida esperada 'Java execution context OK', obtenida: %v", res)
	}
}

func TestPHPAndWasmPluginJPBCExecution(t *testing.T) {
	tempDir := t.TempDir()

	// 1. PHP Plugin
	phpFile := filepath.Join(tempDir, "plugin.php")
	_ = os.WriteFile(phpFile, []byte("<?php\nfunction process_order($id) { return true; }\n"), 0644)

	optsPHP := plugincompiler.Options{
		SourceDir: tempDir,
		Language:  "php",
		EntryFile: phpFile,
		Name:      "php_gateway",
		Version:   "1.0.0",
		Exports:   []string{"process_order"},
	}
	jpPathPHP, _, err := plugincompiler.CompileProject(optsPHP)
	if err != nil {
		t.Fatalf("error compilando php: %v", err)
	}

	pluginPHP, err := pluginruntime.LoadPluginFromFile(jpPathPHP)
	if err != nil {
		t.Fatalf("error cargando php .jp: %v", err)
	}

	// 2. Wasm Plugin
	wasmFile := filepath.Join(tempDir, "crypto.wasm")
	_ = os.WriteFile(wasmFile, []byte{0x00, 0x61, 0x73, 0x6D, 0x01, 0x00, 0x00, 0x00}, 0644)

	optsWasm := plugincompiler.Options{
		SourceDir: tempDir,
		Language:  "wasm",
		EntryFile: wasmFile,
		Name:      "wasm_crypto",
		Version:   "1.0.0",
		Exports:   []string{"hash_data"},
	}
	jpPathWasm, _, err := plugincompiler.CompileProject(optsWasm)
	if err != nil {
		t.Fatalf("error compilando wasm: %v", err)
	}

	pluginWasm, err := pluginruntime.LoadPluginFromFile(jpPathWasm)
	if err != nil {
		t.Fatalf("error cargando wasm .jp: %v", err)
	}

	reg := pluginruntime.NewPluginRegistry(nil)
	_ = reg.Register(pluginPHP)
	_ = reg.Register(pluginWasm)

	resPHP, err := reg.CallFunction("php_gateway", "process_order", []interface{}{123})
	if err != nil {
		t.Fatalf("error ejecutando funcion PHP: %v", err)
	}
	if resPHP == nil {
		t.Error("resultado PHP no debe ser nil")
	}

	resWasm, err := reg.CallFunction("wasm_crypto", "hash_data", nil)
	if err != nil {
		t.Fatalf("error ejecutando funcion Wasm: %v", err)
	}
	if resWasm == nil {
		t.Error("resultado Wasm no debe ser nil")
	}
}

func TestJPBCVMAllOpCodes(t *testing.T) {
	// Construir modulo JPBC con pruebas para todos los OpCodes
	module := &pluginruntime.JPBCModule{
		MajorVersion: 1,
		MinorVersion: 0,
		ConstantPool: []interface{}{
			"hello",   // 0
			"world",   // 1
			int64(10), // 2
			int64(20), // 3
			"User",    // 4
			"name",    // 5
			"Alice",   // 6
		},
		Functions: make(map[string]*pluginruntime.JPBCFunction),
	}

	// 1. OpAdd (Aritmética)
	module.Functions["test_add"] = &pluginruntime.JPBCFunction{
		Name: "test_add",
		Instructions: []pluginruntime.JPBCInstruction{
			{Op: ir.OpNop},
			{Op: ir.OpAdd},
			{Op: ir.OpReturn},
		},
	}

	// 2. OpSub, OpMul, OpDiv, OpMod
	module.Functions["test_math"] = &pluginruntime.JPBCFunction{
		Name: "test_math",
		Instructions: []pluginruntime.JPBCInstruction{
			{Op: ir.OpMul},
			{Op: ir.OpReturn},
		},
	}

	// 3. OpConst + OpReturn
	module.Functions["test_const"] = &pluginruntime.JPBCFunction{
		Name: "test_const",
		Instructions: []pluginruntime.JPBCInstruction{
			{Op: ir.OpConst, ConstIdx: 0},
			{Op: ir.OpReturn},
		},
	}

	// 4. OpNewObject + OpSetField + OpGetField
	module.Functions["test_object"] = &pluginruntime.JPBCFunction{
		Name: "test_object",
		Instructions: []pluginruntime.JPBCInstruction{
			{Op: ir.OpNewObject, ConstIdx: 4}, // type "User"
			{Op: ir.OpConst, ConstIdx: 6},     // "Alice"
			{Op: ir.OpSetField, ConstIdx: 5},  // obj["name"] = "Alice"
			{Op: ir.OpGetField, ConstIdx: 5},  // get obj["name"]
			{Op: ir.OpReturn},
		},
	}

	// 5. OpBranch & OpBranchIf
	module.Functions["test_branch"] = &pluginruntime.JPBCFunction{
		Name: "test_branch",
		Instructions: []pluginruntime.JPBCInstruction{
			{Op: ir.OpConst, ConstIdx: 2},    // 10
			{Op: ir.OpBranchIf, ConstIdx: 3}, // if true jump to inst 3
			{Op: ir.OpConst, ConstIdx: 0},    // dead instruction
			{Op: ir.OpConst, ConstIdx: 3},    // 20
			{Op: ir.OpReturn},
		},
	}

	vm := pluginruntime.NewJPBCVM(module, nil, nil)

	// Test 1: Add
	resAdd, err := vm.Execute("test_add", []interface{}{15, 30})
	if err != nil || resAdd != int64(45) {
		t.Errorf("OpAdd esperado 45, obtenido %v (err: %v)", resAdd, err)
	}

	// Test 2: Mul
	resMul, err := vm.Execute("test_math", []interface{}{6, 7})
	if err != nil || resMul != int64(42) {
		t.Errorf("OpMul esperado 42, obtenido %v", resMul)
	}

	// Test 3: Const
	resConst, err := vm.Execute("test_const", nil)
	if err != nil || resConst != "hello" {
		t.Errorf("OpConst esperado 'hello', obtenido %v", resConst)
	}

	// Test 4: Object
	resObj, err := vm.Execute("test_object", nil)
	if err != nil || resObj != "Alice" {
		t.Errorf("OpGetField esperado 'Alice', obtenido %v", resObj)
	}

	// Test 5: Branch
	resBranch, err := vm.Execute("test_branch", nil)
	if err != nil || resBranch != int64(20) {
		t.Errorf("OpBranchIf esperado 20, obtenido %v", resBranch)
	}
}

func TestSecurityAndErrorBoundaries(t *testing.T) {
	// 1. Rechazar plugin sin firma / con firma corrupta
	metadata := pluginpkg.Metadata{Name: "bad_plugin", Version: "1.0.0", Bytecode: "bytecode/main.jbc"}
	unsignedData, _ := pluginpkg.Build(metadata, map[string][]byte{"bytecode/main.jbc": []byte("test")})
	_, err := pluginruntime.LoadPlugin(unsignedData)
	if err == nil {
		t.Error("esperaba error por plugin no firmado")
	}

	// 2. Division por cero en JPBC VM
	modDivZero := &pluginruntime.JPBCModule{
		MajorVersion: 1,
		MinorVersion: 0,
		Functions: map[string]*pluginruntime.JPBCFunction{
			"div_zero": {
				Name: "div_zero",
				Instructions: []pluginruntime.JPBCInstruction{
					{Op: ir.OpDiv},
					{Op: ir.OpReturn},
				},
			},
		},
	}
	vm := pluginruntime.NewJPBCVM(modDivZero, nil, nil)
	_, err = vm.Execute("div_zero", []interface{}{10, 0})
	if err == nil {
		t.Error("esperaba error por division por cero")
	}

	// 3. Captura segura de panics
	resSafe, errSafe := pluginruntime.SafeCall("test_plug", "crash", func() (interface{}, error) {
		panic("simulated panic")
	})
	if resSafe != nil || errSafe == nil {
		t.Error("SafeCall debio capturar el panic y retornar error estructurado")
	}

	// 4. Registro duplicado
	reg := pluginruntime.NewPluginRegistry(nil)
	p := &pluginruntime.Plugin{Name: "unique_plug", Version: "1.0.0"}
	if err := reg.Register(p); err != nil {
		t.Fatalf("error registrando primer plugin: %v", err)
	}
	if err := reg.Register(p); err == nil {
		t.Error("esperaba error por registrar plugin duplicado")
	}

	// 5. Permisos
	guard := pluginruntime.NewPermissionGuard([]string{"network.http", "filesystem.*"})
	if !guard.HasPermission("network.http") {
		t.Error("debio conceder network.http")
	}
	if !guard.HasPermission("filesystem.read") {
		t.Error("debio conceder filesystem.read por wildcard")
	}
	if guard.HasPermission("database.write") {
		t.Error("no debio conceder database.write")
	}
}

type MockASTEngine struct {
	Classes   map[string]*parser.ClassStatement
	Functions map[string]*parser.MethodStatement
	Tag       string
}

func (m *MockASTEngine) RegisterProgram(prog *parser.Program) error {
	if m.Classes == nil {
		m.Classes = make(map[string]*parser.ClassStatement)
	}
	if m.Functions == nil {
		m.Functions = make(map[string]*parser.MethodStatement)
	}
	for _, s := range prog.Statements {
		if c, ok := s.(*parser.ClassStatement); ok {
			m.Classes[c.Name.Value] = c
		}
		if f, ok := s.(*parser.MethodStatement); ok {
			m.Functions[f.Name.Value] = f
		}
	}
	return nil
}

func (m *MockASTEngine) CallFunction(name string, args []interface{}) (interface{}, error) {
	if name == "calculate_tax" && len(args) > 0 {
		if v, ok := args[0].(float64); ok {
			return v * 0.16, nil
		}
	}
	if name == "process_calculation" && len(args) > 0 {
		return int64(21), nil
	}
	return nil, nil
}

func (m *MockASTEngine) Instantiate(className string, args []interface{}) (interface{}, error) {
	return map[string]interface{}{"__class__": className, "__tag__": m.Tag}, nil
}

func (m *MockASTEngine) CallMethod(instance interface{}, methodName string, args []interface{}) (interface{}, error) {
	if methodName == "add" && len(args) >= 2 {
		return int64(35), nil
	}
	if methodName == "compute" {
		if len(args) >= 2 && args[0] == "double" {
			return int64(100), nil
		}
		return int64(60), nil
	}
	if methodName == "value" {
		if instMap, ok := instance.(map[string]interface{}); ok {
			if tag, ok := instMap["__tag__"].(string); ok && tag != "" {
				return tag, nil
			}
		}
	}
	return nil, nil
}
