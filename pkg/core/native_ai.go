package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/jossecurity/joss/pkg/parser"
)

// AI Handler
func (r *Runtime) executeAIMethod(instance *Instance, method string, args []interface{}) interface{} {
	switch method {
	case "chat":
		// AI.chat(provider, model, messages)
		if len(args) < 3 {
			return nil
		}
		provider, _ := args[0].(string)
		model, _ := args[1].(string)
		messages := args[2]

		return r.aiChat(provider, model, messages)

	case "stream":
		// AI.stream(provider, model, messages, callback)
		if len(args) < 4 {
			return nil
		}
		provider, _ := args[0].(string)
		model, _ := args[1].(string)
		messages := args[2]
		callback := args[3]

		r.aiStream(provider, model, messages, callback)
		return nil

	case "client":
		// AI.client() -> Returns ChatClient Instance
		return r.newChatClient()
	}
	return nil
}

// ChatClient Factory
func (r *Runtime) newChatClient() *Instance {
	if _, ok := r.Classes["ChatClient"]; !ok {
		// Fallback: Register on the fly if missing
		fmt.Println("[AI DEBUG] ChatClient class not found in Runtime. Registering on-the-fly...")
		methods := []string{"user", "system", "prompt", "assistant", "call", "stream", "streamTo"}
		// We need to construct the ClassStatement manually since registerNative is private
		stmts := []parser.Statement{}
		for _, m := range methods {
			stmts = append(stmts, &parser.MethodStatement{Name: &parser.Identifier{Value: m}})
		}
		classStmt := &parser.ClassStatement{
			Name: &parser.Identifier{Value: "ChatClient"},
			Body: &parser.BlockStatement{Statements: stmts},
		}
		r.Classes["ChatClient"] = classStmt
		r.NativeHandlers["ChatClient"] = (*Runtime).executeChatClientMethod
	}
	// Defaults from Env
	provider := r.Env["AI_PROVIDER"]
	if provider == "" {
		provider = os.Getenv("AI_PROVIDER")
	}
	if provider == "" {
		provider = "groq"
	}

	model := r.Env["AI_MODEL"]
	if model == "" {
		model = os.Getenv("AI_MODEL")
	}
	if model == "" {
		model = "llama-3.1-8b-instant"
	}

	return &Instance{
		Class: r.Classes["ChatClient"],
		Fields: map[string]interface{}{
			"provider": provider,
			"model":    model,
			"messages": []interface{}{},
		},
	}
}

// ChatClient Handler
func (r *Runtime) executeChatClientMethod(instance *Instance, method string, args []interface{}) interface{} {
	// Helper to get messages
	msgs, _ := instance.Fields["messages"].([]interface{})

	switch method {
	case "user":
		if len(args) > 0 {
			msgs = append(msgs, map[string]interface{}{"role": "user", "content": args[0]})
			instance.Fields["messages"] = msgs
		}
		return instance // Fluent

	case "system":
		if len(args) > 0 {
			msgs = append(msgs, map[string]interface{}{"role": "system", "content": args[0]})
			instance.Fields["messages"] = msgs
		}
		return instance

	case "prompt":
		// Alias for user
		if len(args) > 0 {
			msgs = append(msgs, map[string]interface{}{"role": "user", "content": args[0]})
			instance.Fields["messages"] = msgs
		}
		return instance

	case "assistant":
		if len(args) > 0 {
			msgs = append(msgs, map[string]interface{}{"role": "assistant", "content": args[0]})
			instance.Fields["messages"] = msgs
		}
		return instance

	case "call":
		// Synchronous call
		provider := instance.Fields["provider"].(string)
		model := instance.Fields["model"].(string)

		res := r.aiChat(provider, model, msgs)
		// Return just the content string for simplicity?
		// Or the full object? Spring AI returns an object usually but .content() gets string.
		// Let's return the content string if success, or map if error.
		if m, ok := res.(map[string]interface{}); ok {
			// Extract content from choices[0].message.content
			if choices, ok := m["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					if msg, ok := choice["message"].(map[string]interface{}); ok {
						if content, ok := msg["content"].(string); ok {
							return content
						}
					}
				}
			}
			return res // Return full if structure unexpected
		}
		return res

	case "stream":
		// stream(callback)
		if len(args) > 0 {
			callback := args[0]
			provider := instance.Fields["provider"].(string)
			model := instance.Fields["model"].(string)
			r.aiStream(provider, model, msgs, callback)
		}
		return nil

	case "streamTo":
		// streamTo(wsInstance)
		if len(args) > 0 {
			if wsInst, ok := args[0].(*Instance); ok {
				// We create a callback that writes to WS
				provider := instance.Fields["provider"].(string)
				model := instance.Fields["model"].(string)

				// WS Writer Callback
				// We need to invoke Stream.send or WebSocket.send on wsInst?
				// wsInst is likely a "WebSocket" instance from Router::ws
				// It should have a 'send' native method logic.
				// We can call r.CallMethod implementation?
				// Or just call executeNativeMethod directly if we know the class?
				// But we are in Go. We can just execute the "send" handler for that instance.
				// It's cleaner to reuse executeWebSocketMethod (renamed/new).

				// Let's assume wsInst is the standard WebSocket instance wrapping the Conn.
				// We call its 'send' method.
				// For now, let's just implement the streaming logic HERE specifically for streamTo to avoid recursion complexity.
				return r.aiStreamToWS(provider, model, msgs, wsInst)
			}
		}
		return nil
	}
	return nil
}

// Dedicated Stream to WebSocket
func (r *Runtime) aiStreamToWS(provider, model string, messages interface{}, wsInst *Instance) string {
	// ... Copy of aiStream logic but calling wsInst.send ...
	// To avoid code duplication, we really should refactor aiStream to take a generic "chunkEmitter" func.

	emitter := func(chunk string) {
		// Wrap in standard JSON protocol for Joss Chat
		msg := map[string]string{
			"type":    "chunk",
			"content": chunk,
		}
		jsonBytes, _ := json.Marshal(msg)
		r.executeWebSocketMethod(wsInst, "send", []interface{}{string(jsonBytes)})
	}
	return r.aiStreamGeneric(provider, model, messages, emitter)
}

// Refactored Generic Streamer
func (r *Runtime) aiStreamGeneric(provider, model string, messages interface{}, emitter func(string)) string {
	var url, apiKeyEnv string
	switch provider {
	case "groq":
		url = "https://api.groq.com/openai/v1/chat/completions"
		apiKeyEnv = "GROQ_API_KEY"
	case "openai":
		url = "https://api.openai.com/v1/chat/completions"
		apiKeyEnv = "OPENAI_API_KEY"
	case "gemini":
		url = "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions"
		apiKeyEnv = "GEMINI_API_KEY"
	default:
		return ""
	}

	apiKey := r.Env[apiKeyEnv]
	if apiKey == "" {
		apiKey = os.Getenv(apiKeyEnv)
	}
	if apiKey == "" {
		emitter("Error: Missing API Key")
		return ""
	}

	payload := map[string]interface{}{
		"model":    model,
		"messages": messages,
		"stream":   true,
	}

	jsonBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		emitter("Error: " + err.Error())
		return ""
	}
	defer resp.Body.Close()

	reader := ioutil.NopCloser(resp.Body)
	buf := make([]byte, 1024)
	bufferStr := ""
	fullContent := ""

	for {
		n, err := reader.Read(buf)
		if n > 0 {
			chunk := string(buf[:n])
			bufferStr += chunk
			for {
				idx := -1
				for i, c := range bufferStr {
					if c == '\n' {
						idx = i
						break
					}
				}
				if idx == -1 {
					break
				}
				line := bufferStr[:idx]
				bufferStr = bufferStr[idx+1:]
				line = TrimSpace(line)
				if len(line) > 6 && line[:5] == "data:" {
					dataContent := TrimSpace(line[5:])
					if dataContent == "[DONE]" {
						break
					}
					var jsonChunk map[string]interface{}
					if err := json.Unmarshal([]byte(dataContent), &jsonChunk); err == nil {
						if choices, ok := jsonChunk["choices"].([]interface{}); ok && len(choices) > 0 {
							if choice, ok := choices[0].(map[string]interface{}); ok {
								if delta, ok := choice["delta"].(map[string]interface{}); ok {
									if content, ok := delta["content"].(string); ok {
										emitter(content)
										fullContent += content
									}
								}
							}
						}
					}
				}
			}
		}
		if err != nil {
			break
		}
	}
	return fullContent
}

// Original aiStream using Generic
func (r *Runtime) aiStream(provider, model string, messages interface{}, callback interface{}) {
	emitter := func(chunk string) {
		r.CallFunction(callback, []interface{}{chunk})
	}
	r.aiStreamGeneric(provider, model, messages, emitter)
}

func TrimSpace(s string) string {
	// Simple trim implementation avoiding strings import if unnecessary, but we use strings anyway?
	// Note: native_extensions used strings. native_ai currently imports bytes, encoding/json, fmt, ioutil, net/http, os.
	// We need to add "strings" to imports if not present.
	// But let's check imports.
	// original file has: bytes, encoding/json, fmt, ioutil, net/http, os
	// I need to add "strings" import or manual trim.
	// Manual trim:
	start := 0
	for start < len(s) && (s[start] == ' ' || s[start] == '\t' || s[start] == '\r') {
		start++
	}
	end := len(s)
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

func (r *Runtime) aiChat(provider, model string, messages interface{}) interface{} {
	var url, apiKeyEnv string

	switch provider {
	case "groq":
		url = "https://api.groq.com/openai/v1/chat/completions"
		apiKeyEnv = "GROQ_API_KEY"
	case "openai":
		url = "https://api.openai.com/v1/chat/completions"
		apiKeyEnv = "OPENAI_API_KEY"
	case "gemini":
		// Gemini mapping usually different (Google Generative AI), but often proxied via OpenAI compat
		// For now simple OpenAI compatible logic
		url = "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions"
		apiKeyEnv = "GEMINI_API_KEY"
	default:
		return map[string]interface{}{"error": "Provider not supported"}
	}

	apiKey := r.Env[apiKeyEnv]
	if apiKey == "" {
		apiKey = os.Getenv(apiKeyEnv)
	}
	if apiKey == "" {
		return map[string]interface{}{"error": "Missing API Key: " + apiKeyEnv}
	}

	// Payload
	payload := map[string]interface{}{
		"model":    model,
		"messages": messages,
	}

	jsonBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	return result
}
