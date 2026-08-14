package pluginruntime

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jossecurity/joss/pkg/pluginpkg"
)

// LoadPluginFromFile lee y verifica un archivo .jp desde disco.
func LoadPluginFromFile(filePath string) (*Plugin, error) {
	return LoadPluginFromFileWithEngine(filePath, nil)
}

// LoadPluginFromFileWithEngine lee, verifica y asocia un motor AST si corresponde.
func LoadPluginFromFileWithEngine(filePath string, engine ASTEngine) (*Plugin, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("pluginruntime: error leyendo archivo %s: %w", filePath, err)
	}
	return LoadPluginWithEngine(data, engine)
}

// LoadPlugin carga y verifica criptograficamente un paquete .jp.
func LoadPlugin(data []byte) (*Plugin, error) {
	return LoadPluginWithEngine(data, nil)
}

// LoadPluginWithEngine carga, verifica y asocia un motor AST opcional.
func LoadPluginWithEngine(data []byte, engine ASTEngine) (*Plugin, error) {
	// 1. Verificacion de firma Ed25519 y estructura ZIP
	archive, err := pluginpkg.ReadVerified(data)
	if err != nil {
		return nil, fmt.Errorf("pluginruntime: validacion criptografica del plugin fallida: %w", err)
	}

	metadata := archive.Metadata

	// 2. Extraer y deserializar indice de simbolos
	var symbolIndex pluginpkg.SymbolIndex
	if metadata.Symbols != "" {
		symData, ok := archive.Files[metadata.Symbols]
		if !ok {
			return nil, fmt.Errorf("pluginruntime: falta archivo de simbolos %q", metadata.Symbols)
		}
		if err := json.Unmarshal(symData, &symbolIndex); err != nil {
			return nil, fmt.Errorf("pluginruntime: error decodificando simbolos: %w", err)
		}
	}

	// 3. Extraer bytecode
	bytecodeData, ok := archive.Files[metadata.Bytecode]
	if !ok {
		return nil, fmt.Errorf("pluginruntime: falta bytecode %q", metadata.Bytecode)
	}

	// 4. Detectar formato del bytecode
	format, err := DetectBytecodeFormat(bytecodeData)
	if err != nil {
		return nil, err
	}

	manifestStr := string(archive.Files["joss.yaml"])

	plugin := &Plugin{
		Name:        metadata.Name,
		Version:     metadata.Version,
		Language:    metadata.Language,
		Format:      format,
		Metadata:    metadata,
		Symbols:     symbolIndex,
		Manifest:    manifestStr,
		RawBytecode: bytecodeData,
	}

	// 5. Instanciar el ejecutor correspondiente
	switch format {
	case FormatJossAST:
		jossExec, err := NewJossASTExecutor(bytecodeData, engine)
		if err != nil {
			return nil, err
		}
		plugin.jossProgram = jossExec.Program
		plugin.jossExecutor = jossExec

	case FormatJPBC:
		jpbcMod, err := DecodeJPBC(bytecodeData)
		if err != nil {
			return nil, fmt.Errorf("pluginruntime: error decodificando JPBC: %w", err)
		}
		plugin.jpbcModule = jpbcMod
	}

	return plugin, nil
}
