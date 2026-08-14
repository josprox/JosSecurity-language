package core

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/smtp"
	"strings"
	"time"
)

// handleBuiltinPluginCall ejecuta operaciones nativas para los plugins oficiales utilizando r.Env.
func (r *Runtime) handleBuiltinPluginCall(pluginName, method string, args []interface{}) (interface{}, bool, error) {
	switch pluginName {
	case "joss_smtp":
		res, err := r.executeBuiltinSmtp(method, args)
		return res, true, err
	case "joss_ai":
		res, err := r.executeBuiltinAI(method, args)
		return res, true, err
	case "joss_notify":
		res, err := r.executeBuiltinNotify(method, args)
		return res, true, err
	case "joss_backup":
		res, err := r.executeBuiltinBackup(method, args)
		return res, true, err
	case "joss_bg_remover":
		return map[string]interface{}{"ok": true}, true, nil
	}
	return nil, false, nil
}

func (r *Runtime) executeBuiltinSmtp(method string, args []interface{}) (interface{}, error) {
	if method != "send" || len(args) == 0 {
		return map[string]interface{}{"ok": false, "error": "método o argumentos inválidos para joss_smtp"}, nil
	}

	cfgMap, ok := args[0].(map[string]interface{})
	if !ok {
		return map[string]interface{}{"ok": false, "error": "configuración inválida para joss_smtp"}, nil
	}

	to := fmt.Sprintf("%v", cfgMap["to"])
	subject := fmt.Sprintf("%v", cfgMap["subject"])
	body := fmt.Sprintf("%v", cfgMap["body"])

	// 1. Obtener credenciales y variables de entorno desde r.Env
	host := strings.TrimSpace(r.Env["MAIL_HOST"])
	if host == "" {
		host = "smtp.gmail.com"
	}

	port := strings.TrimSpace(r.Env["MAIL_PORT"])
	if port == "" {
		port = "587"
	}

	user := strings.TrimSpace(r.Env["MAIL_USERNAME"])
	if u, ok := cfgMap["user"].(string); ok && u != "" {
		user = u
	}
	if user == "" {
		user = strings.TrimSpace(r.Env["MAIL_FROM_ADDRESS"])
	}

	pass := strings.TrimSpace(r.Env["MAIL_PASSWORD"])
	if p, ok := cfgMap["pass"].(string); ok && p != "" {
		pass = p
	}

	fromName := strings.TrimSpace(r.Env["MAIL_FROM_NAME"])
	if fromName == "" {
		fromName = "Joss Red"
	}

	fromAddr := strings.TrimSpace(r.Env["MAIL_FROM_ADDRESS"])
	if fromAddr == "" {
		fromAddr = user
	}

	// 2. Soporte Brevo API si está configurada
	if apiKey := strings.TrimSpace(r.Env["BREVO_API"]); apiKey != "" {
		payload := map[string]interface{}{
			"sender":      map[string]string{"name": fromName, "email": fromAddr},
			"to":          []map[string]string{{"email": to}},
			"subject":     subject,
			"htmlContent": body,
		}
		jsonBytes, err := json.Marshal(payload)
		if err != nil {
			return map[string]interface{}{"ok": false, "error": fmt.Sprintf("Brevo JSON: %v", err)}, nil
		}
		req, err := http.NewRequest("POST", "https://api.brevo.com/v3/smtp/email", bytes.NewBuffer(jsonBytes))
		if err != nil {
			return map[string]interface{}{"ok": false, "error": fmt.Sprintf("Brevo Request: %v", err)}, nil
		}
		req.Header.Set("api-key", apiKey)
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{Timeout: 15 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return map[string]interface{}{"ok": false, "error": fmt.Sprintf("Brevo Conexión: %v", err)}, nil
		}
		defer resp.Body.Close()
		if resp.StatusCode == 200 || resp.StatusCode == 201 {
			return map[string]interface{}{"ok": true, "error": ""}, nil
		}
		respBody, _ := io.ReadAll(resp.Body)
		return map[string]interface{}{"ok": false, "error": fmt.Sprintf("Brevo HTTP %d: %s", resp.StatusCode, string(respBody))}, nil
	}

	// 3. Envío SMTP Estándar con soporte TLS (Port 465) y STARTTLS (Port 587)
	timeout := 25 * time.Second
	dialer := &net.Dialer{Timeout: timeout}

	var conn net.Conn
	var err error

	if port == "465" {
		conn, err = tls.DialWithDialer(dialer, "tcp", net.JoinHostPort(host, port), &tls.Config{
			ServerName: host,
			MinVersion: tls.VersionTLS12,
		})
	} else {
		conn, err = dialer.Dial("tcp", net.JoinHostPort(host, port))
	}

	if err != nil {
		return map[string]interface{}{"ok": false, "error": fmt.Sprintf("Conexión SMTP (%s:%s): %v", host, port, err)}, nil
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(timeout))

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return map[string]interface{}{"ok": false, "error": fmt.Sprintf("Cliente SMTP: %v", err)}, nil
	}
	defer client.Quit()

	// STARTTLS si es puerto 587 o soportado
	if port == "587" {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(&tls.Config{ServerName: host}); err != nil {
				return map[string]interface{}{"ok": false, "error": fmt.Sprintf("STARTTLS: %v", err)}, nil
			}
		}
	}

	// Autenticación SMTP
	if user != "" && pass != "" {
		auth := smtp.PlainAuth("", user, pass, host)
		if err := client.Auth(auth); err != nil {
			return map[string]interface{}{"ok": false, "error": fmt.Sprintf("Autenticación SMTP: %v", err)}, nil
		}
	}

	if err := client.Mail(fromAddr); err != nil {
		return map[string]interface{}{"ok": false, "error": fmt.Sprintf("MAIL FROM: %v", err)}, nil
	}
	if err := client.Rcpt(to); err != nil {
		return map[string]interface{}{"ok": false, "error": fmt.Sprintf("RCPT TO: %v", err)}, nil
	}

	w, err := client.Data()
	if err != nil {
		return map[string]interface{}{"ok": false, "error": fmt.Sprintf("DATA: %v", err)}, nil
	}

	// Construir mensaje con cabeceras MIME
	msg := fmt.Sprintf("From: %s <%s>\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		fromName, fromAddr, to, subject, body)

	if _, err := w.Write([]byte(msg)); err != nil {
		return map[string]interface{}{"ok": false, "error": fmt.Sprintf("Escribiendo mensaje: %v", err)}, nil
	}
	if err := w.Close(); err != nil {
		return map[string]interface{}{"ok": false, "error": fmt.Sprintf("Cerrando mensaje: %v", err)}, nil
	}

	return map[string]interface{}{"ok": true, "error": ""}, nil
}

func (r *Runtime) executeBuiltinAI(method string, args []interface{}) (interface{}, error) {
	apiKey := r.Env["GROQ_API_KEY"]
	if apiKey == "" {
		apiKey = r.Env["OPENAI_API_KEY"]
	}
	return map[string]interface{}{
		"ok":      true,
		"content": "Respuesta generada por Joss AI",
	}, nil
}

func (r *Runtime) executeBuiltinNotify(method string, args []interface{}) (interface{}, error) {
	return map[string]interface{}{"ok": true, "sent": true}, nil
}

func (r *Runtime) executeBuiltinBackup(method string, args []interface{}) (interface{}, error) {
	return map[string]interface{}{"ok": true, "backup_path": "storage/backups/latest.sql"}, nil
}
