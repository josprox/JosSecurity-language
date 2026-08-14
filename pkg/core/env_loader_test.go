package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnvLoaderFileOperations(t *testing.T) {
	tempDir := t.TempDir()
	envPath := filepath.Join(tempDir, "env.joss")

	// 1. Update and create
	if err := UpdateEnvFile(envPath, "APP_NAME", "JossApp"); err != nil {
		t.Fatalf("UpdateEnvFile failed: %v", err)
	}
	if err := UpdateEnvFile(envPath, "APP_PORT", "9000"); err != nil {
		t.Fatalf("UpdateEnvFile failed: %v", err)
	}

	// 2. Read map
	m := ReadEnvFile(envPath)
	if m["APP_NAME"] != "JossApp" {
		t.Errorf("Expected APP_NAME='JossApp', got %q", m["APP_NAME"])
	}
	if m["APP_PORT"] != "9000" {
		t.Errorf("Expected APP_PORT='9000', got %q", m["APP_PORT"])
	}

	// 3. Update existing
	if err := UpdateEnvFile(envPath, "APP_PORT", "9500"); err != nil {
		t.Fatalf("UpdateEnvFile failed: %v", err)
	}
	m2 := ReadEnvFile(envPath)
	if m2["APP_PORT"] != "9500" {
		t.Errorf("Expected APP_PORT='9500', got %q", m2["APP_PORT"])
	}

	// 4. Remove key
	if err := RemoveEnvKey(envPath, "APP_PORT"); err != nil {
		t.Fatalf("RemoveEnvKey failed: %v", err)
	}
	m3 := ReadEnvFile(envPath)
	if _, exists := m3["APP_PORT"]; exists {
		t.Errorf("Expected APP_PORT to be removed, but still present")
	}
	if m3["APP_NAME"] != "JossApp" {
		t.Errorf("Expected APP_NAME to still be present, got %q", m3["APP_NAME"])
	}
}

func TestResolvePort(t *testing.T) {
	// 1. Default
	if p := ResolvePort(nil); p != "8000" {
		t.Errorf("Expected default port '8000', got %q", p)
	}

	// 2. From map
	envMap := map[string]string{"PORT": "4000"}
	if p := ResolvePort(envMap); p != "4000" {
		t.Errorf("Expected port '4000', got %q", p)
	}

	// 3. From OS Env
	os.Setenv("PORT", "7777")
	defer os.Unsetenv("PORT")
	if p := ResolvePort(nil); p != "7777" {
		t.Errorf("Expected port '7777', got %q", p)
	}
}
