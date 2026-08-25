package core

import (
	"testing"

	"github.com/jossecurity/joss/pkg/pluginruntime"
)

func TestFreeClearsPluginRegistryBeforeRuntimeReuse(t *testing.T) {
	r := NewRuntime()
	r.PluginRegistry = pluginruntime.NewPluginRegistry(r)
	r.Free()

	reused := NewRuntime()
	defer reused.Free()

	if reused.PluginRegistry != nil {
		t.Fatal("a pooled runtime retained the previous plugin registry")
	}
}
