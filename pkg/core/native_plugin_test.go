package core

import (
	"strings"
	"testing"
)

func TestParsePluginManifestWithNative(t *testing.T) {
	yamlData := []byte(`name: joss_smtp
version: 2.0.1
type: joss
entry:
  main: src/plugin.joss
native:
  protocol: joss-rpc-v1
  windows-amd64: native/windows-amd64/joss_smtp.exe
  linux-amd64: native/linux-amd64/joss_smtp
`)
	manifest := parsePluginManifest(yamlData)
	if manifest.Name != "joss_smtp" {
		t.Fatalf("expected name joss_smtp, got %q", manifest.Name)
	}
	if manifest.Protocol != "joss-rpc-v1" {
		t.Fatalf("expected protocol joss-rpc-v1, got %q", manifest.Protocol)
	}
	if manifest.Native["windows-amd64"] != "native/windows-amd64/joss_smtp.exe" {
		t.Fatalf("expected windows-amd64 path, got %q", manifest.Native["windows-amd64"])
	}
}

func TestPluginEnvironmentDefaultPrefixes(t *testing.T) {
	r := NewRuntime()
	r.Env = map[string]string{
		"SECRET_KEY":       "hidden",
		"MAIL_USERNAME":    "test@joss.red",
		"AI_PROVIDER":      "groq",
		"FCM_SERVER_KEY":   "fcm-123",
		"PLUGIN_ENV_ALLOW": "CUSTOM_VAR",
		"CUSTOM_VAR":       "custom_val",
	}
	env := pluginCommandEnvironment(r, "C:/plugins/demo/driver.exe")
	joined := strings.Join(env, "\n")

	if strings.Contains(joined, "SECRET_KEY=hidden") {
		t.Fatal("plugin inherited SECRET_KEY without permission")
	}
	if !strings.Contains(joined, "MAIL_USERNAME=test@joss.red") {
		t.Fatal("MAIL_USERNAME was not inherited by default prefix rule")
	}
	if !strings.Contains(joined, "AI_PROVIDER=groq") {
		t.Fatal("AI_PROVIDER was not inherited by default prefix rule")
	}
	if !strings.Contains(joined, "FCM_SERVER_KEY=fcm-123") {
		t.Fatal("FCM_SERVER_KEY was not inherited by default prefix rule")
	}
	if !strings.Contains(joined, "CUSTOM_VAR=custom_val") {
		t.Fatal("CUSTOM_VAR was not inherited by explicit PLUGIN_ENV_ALLOW rule")
	}
}


