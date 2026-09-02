package core

import (
	"testing"
)

func TestCustomMiddlewareReceivesRequestAndHeaders(t *testing.T) {
	r := NewRuntime()
	r.RegisterNativeClasses()

	capturedHeader := ""
	capturedMethod := ""

	r.CustomMiddlewares["TestAuth"] = func(args []interface{}) interface{} {
		// Test that Request native methods see the dispatched request
		hdr := r.executeRequestMethod(nil, "header", []interface{}{"Authorization"})
		if hdr != nil {
			capturedHeader = hdr.(string)
		}
		m := r.executeRequestMethod(nil, "method", nil)
		if m != nil {
			capturedMethod = m.(string)
		}
		return nil
	}

	r.Routes["POST"] = map[string]interface{}{
		"/api/test": map[string]interface{}{
			"handler": func(args []interface{}) interface{} {
				return "ok"
			},
			"middleware": []string{"TestAuth"},
		},
	}

	reqData := map[string]interface{}{
		"_method":       "POST",
		"Authorization": "Bearer test-secret-token",
		"_headers": map[string]interface{}{
			"Authorization": "Bearer test-secret-token",
		},
	}
	sessData := map[string]interface{}{}

	_, err := r.Dispatch("POST", "/api/test", reqData, sessData)
	if err != nil {
		t.Fatalf("Dispatch error: %v", err)
	}

	if capturedMethod != "POST" {
		t.Errorf("Expected method POST, got %q", capturedMethod)
	}
	if capturedHeader != "Bearer test-secret-token" {
		t.Errorf("Expected header Bearer test-secret-token, got %q", capturedHeader)
	}
}
