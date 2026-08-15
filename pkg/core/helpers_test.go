package core

import (
	"testing"
)

func TestGlobalHelpers(t *testing.T) {
	r := NewRuntime()
	r.Env["APP_ENV"] = "production"
	r.Env["PORT"] = "9000"

	// 1. env() and config()
	val, ok := r.callBuiltin("env", []interface{}{"APP_ENV"})
	if !ok || val != "production" {
		t.Fatalf("env('APP_ENV') esperado 'production', obtenido: %v", val)
	}

	valDef, ok := r.callBuiltin("env", []interface{}{"NON_EXISTENT", "default_val"})
	if !ok || valDef != "default_val" {
		t.Fatalf("env('NON_EXISTENT', 'default_val') esperado 'default_val', obtenido: %v", valDef)
	}

	valCfg, ok := r.callBuiltin("config", []interface{}{"PORT"})
	if !ok || valCfg != "9000" {
		t.Fatalf("config('PORT') esperado '9000', obtenido: %v", valCfg)
	}

	// 2. json()
	jsonRes, ok := r.callBuiltin("json", []interface{}{map[string]interface{}{"status": "ok"}, int64(201)})
	if !ok {
		t.Fatalf("json() falló")
	}
	if inst, isInst := jsonRes.(*Instance); !isInst || inst.Fields["status_code"] != 201 {
		t.Fatalf("json() status_code esperado 201, obtenido: %v", jsonRes)
	}

	// 3. redirect()
	redRes, ok := r.callBuiltin("redirect", []interface{}{"/dashboard", int64(301)})
	if !ok {
		t.Fatalf("redirect() falló")
	}
	if inst, isInst := redRes.(*Instance); !isInst || inst.Fields["url"] != "/dashboard" || inst.Fields["status_code"] != 301 {
		t.Fatalf("redirect() datos inválidos: %v", redRes)
	}

	// 4. back()
	backRes, ok := r.callBuiltin("back", nil)
	if !ok {
		t.Fatalf("back() falló")
	}
	if inst, isInst := backRes.(*Instance); !isInst || inst.Fields["_type"] != "REDIRECT" {
		t.Fatalf("back() esperado REDIRECT, obtenido: %v", backRes)
	}
}

func TestSmtpClientNative(t *testing.T) {
	r := NewRuntime()
	if _, ok := r.Classes["SmtpClient"]; !ok {
		t.Fatalf("Clase nativa SmtpClient no registrada en Runtime")
	}

	inst := &Instance{Class: r.Classes["SmtpClient"], Fields: make(map[string]interface{})}
	r.executeSmtpClientMethod(inst, "auth", []interface{}{"user@test.com", "secret"})
	if inst.Fields["user"] != "user@test.com" || inst.Fields["pass"] != "secret" {
		t.Fatalf("SmtpClient.auth no guardó credenciales")
	}
}
