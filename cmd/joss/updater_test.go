package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReplaceExecutable(t *testing.T) {
	tempDir := t.TempDir()

	currentExe := filepath.Join(tempDir, "joss_current_binary")
	newExe := filepath.Join(tempDir, "joss_new_binary")

	if err := os.WriteFile(currentExe, []byte("v1.0.0-binary-content"), 0755); err != nil {
		t.Fatalf("failed to create currentExe: %v", err)
	}

	if err := os.WriteFile(newExe, []byte("v2.0.0-binary-content"), 0755); err != nil {
		t.Fatalf("failed to create newExe: %v", err)
	}

	if err := replaceExecutable(currentExe, newExe); err != nil {
		t.Fatalf("replaceExecutable failed: %v", err)
	}

	updatedContent, err := os.ReadFile(currentExe)
	if err != nil {
		t.Fatalf("failed to read updated currentExe: %v", err)
	}

	if string(updatedContent) != "v2.0.0-binary-content" {
		t.Fatalf("expected 'v2.0.0-binary-content', got '%s'", string(updatedContent))
	}
}

func TestVersionParsingAndComparison(t *testing.T) {
	if !isVersionNewer("3.6.8", "3.6.7") {
		t.Fatalf("expected 3.6.8 to be newer than 3.6.7")
	}
	if !isVersionNewer("3.6.7.4", "3.6.7.3") {
		t.Fatalf("expected 3.6.7.4 to be newer than 3.6.7.3")
	}
	if isVersionNewer("3.6.7", "3.6.7") {
		t.Fatalf("expected 3.6.7 not to be newer than 3.6.7")
	}
	if isVersionNewer("3.6.6", "3.6.7") {
		t.Fatalf("expected 3.6.6 not to be newer than 3.6.7")
	}
}
