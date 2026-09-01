package main

import (
	"testing"
)

func TestPluginCommandsDiscoveryAndHelp(t *testing.T) {
	plugins := discoverPlugins()
	if len(plugins) == 0 {
		t.Fatalf("expected at least standard official plugins to be discovered")
	}

	// Verify joss_ai plugin
	ai, ok := plugins["joss_ai"]
	if !ok {
		t.Fatalf("expected joss_ai plugin to be discovered")
	}
	if _, ok := ai.Commands["ai:activate"]; !ok {
		t.Fatalf("expected ai:activate command in joss_ai")
	}

	// Verify joss_backup plugin
	backup, ok := plugins["joss_backup"]
	if !ok {
		t.Fatalf("expected joss_backup plugin to be discovered")
	}
	if _, ok := backup.Commands["backup:create"]; !ok {
		t.Fatalf("expected backup:create command in joss_backup")
	}
	if restoreCmd, ok := backup.Commands["backup:restore"]; !ok || !restoreCmd.Protected {
		t.Fatalf("expected backup:restore to be protected command in joss_backup")
	}

	// Test tryDispatchPluginCommand
	if !tryDispatchPluginCommand("backup:create", []string{"test_backups"}) {
		t.Fatalf("expected backup:create to be dispatched")
	}
}
