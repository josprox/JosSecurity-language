package plugincompiler_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jossecurity/joss/pkg/plugincompiler"
	"github.com/jossecurity/joss/pkg/pluginpkg"
)

func TestJavaPluginCompilationAndTreeShaking(t *testing.T) {
	tempDir := t.TempDir()
	classFile := filepath.Join(tempDir, "MiPlugin.class")

	// Header CAFEBABE de Java class file
	classBytes := []byte{0xCA, 0xFE, 0xBA, 0xBE, 0x00, 0x00, 0x00, 0x34}
	if err := os.WriteFile(classFile, classBytes, 0644); err != nil {
		t.Fatalf("error preparando archivo .class de prueba: %v", err)
	}

	opts := plugincompiler.Options{
		SourceDir:   tempDir,
		Language:    "java",
		EntryFile:   classFile,
		Name:        "test-music-plugin",
		Version:     "1.0.0",
		Exports:     []string{"searchSong", "getSong"},
		Permissions: []string{"network.http"},
		MaxSizeMB:   1.0,
	}

	outPath, result, err := plugincompiler.CompileProject(opts)
	if err != nil {
		t.Fatalf("error compilando plugin Java: %v", err)
	}

	fi, err := os.Stat(outPath)
	if err != nil {
		t.Fatalf("no se generó el paquete .jp: %v", err)
	}

	// Verificar objetivo de tamaño < 1 MB
	sizeMB := float64(fi.Size()) / (1024 * 1024)
	if sizeMB > 1.0 {
		t.Errorf("el plugin excede el limite de 1 MB (tamaño real: %.2f MB)", sizeMB)
	}

	if result.OptimizedFuncs == 0 {
		t.Errorf("tree shaker reporta 0 funciones conservadas")
	}

	// Probar decodificacion e inspeccion del .jp resultante
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("error leyendo .jp generado: %v", err)
	}

	archive, err := pluginpkg.Read(data)
	if err != nil {
		t.Fatalf("error al verificar firma y leer .jp: %v", err)
	}

	if archive.Metadata.Name != "test-music-plugin" {
		t.Errorf("nombre de plugin esperado 'test-music-plugin', obtenido '%s'", archive.Metadata.Name)
	}

	if len(archive.Metadata.Exports) != 2 {
		t.Errorf("se esperaban 2 exportaciones, obtenidas %d", len(archive.Metadata.Exports))
	}
}
