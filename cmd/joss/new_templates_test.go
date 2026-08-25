package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jossecurity/joss/pkg/bytecode"
	"github.com/jossecurity/joss/pkg/core"
	"github.com/jossecurity/joss/pkg/parser"
	"github.com/jossecurity/joss/pkg/pluginpkg"
)

func TestNewPackageAndPluginTemplatesCompileEndToEnd(t *testing.T) {
	root := t.TempDir()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(previous) }()

	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	keyPath := filepath.Join(root, "test-signing-key.ed25519")
	if err := os.WriteFile(keyPath, privateKey, 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("JOSS_PLUGIN_SIGNING_KEY", keyPath)

	tests := []struct {
		name   string
		create func(string)
	}{
		{name: "sample_package", create: createNewPackage},
		{name: "sample_plugin", create: createNewPluginProject},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.create(test.name)
			sourcePath := filepath.Join(test.name, "src", "plugin.joss")
			data, readErr := os.ReadFile(sourcePath)
			if readErr != nil {
				t.Fatal(readErr)
			}
			p := parser.NewParser(parser.NewLexer(string(data)))
			program := p.ParseProgram()
			if errors := p.Errors(); len(errors) > 0 {
				t.Fatalf("generated source parse errors: %v", errors)
			}
			if report := core.AnalyzeProgram(program); report.HasErrors() {
				t.Fatalf("generated source semantic errors: %#v", report.Diagnostics)
			}

			buildPackage(test.name)
			archivePath := filepath.Join(test.name, test.name+".jp")
			archiveData, readErr := os.ReadFile(archivePath)
			if readErr != nil {
				t.Fatalf("generated package was not compiled: %v", readErr)
			}
			archive, verifyErr := pluginpkg.ReadVerified(archiveData)
			if verifyErr != nil {
				t.Fatalf("generated package verification failed: %v", verifyErr)
			}
			encoded := archive.Files[archive.Metadata.Bytecode]
			if len(encoded) == 0 {
				t.Fatal("generated package has no bytecode payload")
			}
			if _, decodeErr := bytecode.Decode(encoded); decodeErr != nil {
				t.Fatalf("generated package bytecode is invalid: %v", decodeErr)
			}
			var symbols pluginpkg.SymbolIndex
			if decodeErr := json.Unmarshal(archive.Files[pluginpkg.SymbolsPath], &symbols); decodeErr != nil {
				t.Fatalf("generated package symbol index is invalid: %v", decodeErr)
			}
			if len(symbols.Classes) == 0 {
				t.Fatal("generated package has no published class symbols")
			}
		})
	}
}
