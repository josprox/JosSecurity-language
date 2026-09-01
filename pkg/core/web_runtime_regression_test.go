package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRouterGroupExecutesCapturedClosure(t *testing.T) {
	source := `
Router::group("auth", func() {
    Router::get("/inside", "Handler@index")
})
`
	runtime := NewRuntime()
	t.Cleanup(runtime.Free)
	runtime.Execute(benchmarkParse(t, source))

	route, ok := runtime.Routes["GET"]["/inside"].(map[string]interface{})
	if !ok {
		t.Fatalf("group callback did not register its route: %#v", runtime.Routes)
	}
	middleware, ok := route["middleware"].([]string)
	if !ok || len(middleware) != 1 || middleware[0] != "auth" {
		t.Fatalf("route middleware = %#v, want [auth]", route["middleware"])
	}
	if len(runtime.CurrentMiddleware) != 0 {
		t.Fatalf("middleware stack leaked after group: %#v", runtime.CurrentMiddleware)
	}
}

func TestViewRenderInsideCallableUsesTemplateScope(t *testing.T) {
	project := t.TempDir()
	views := filepath.Join(project, "app", "views")
	if err := os.MkdirAll(views, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(views, "hello.joss.html"), []byte("Hello {{ $name }}"), 0o644); err != nil {
		t.Fatal(err)
	}

	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(project); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(previous) })

	source := `
public func renderHello(): string {
    return view("hello", {"name": "Ada"})
}
`
	runtime := benchmarkPreparedRuntime(t, source)
	result := runtime.CallMethodEvaluated(runtime.Functions["renderHello"], nil, nil)
	if result != "Hello Ada" {
		t.Fatalf("renderHello() = %q, want %q", result, "Hello Ada")
	}
}
