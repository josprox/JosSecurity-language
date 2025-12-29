package core

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Stream Handler
func (r *Runtime) executeStreamMethod(instance *Instance, method string, args []interface{}) interface{} {
	// Retrieve ResponseWriter from hidden field
	wVal, ok := instance.Fields["_writer"]
	if !ok {
		return nil
	}
	w, ok := wVal.(http.ResponseWriter)
	if !ok {
		return nil
	}

	switch method {
	case "send":
		// Stream.send(data) or (type, data)
		// We format as SSE: "data: <json>\n\n"

		var payload interface{}
		var eventType string

		if len(args) == 1 {
			payload = args[0]
		} else if len(args) >= 2 {
			if t, ok := args[0].(string); ok {
				eventType = t // Optional event name? OSS Standard is usually just data
			}
			payload = args[1]
		}

		// If payload is map, json encode it
		var dataStr string
		if payloadMap, ok := payload.(map[string]interface{}); ok {
			bytes, _ := json.Marshal(payloadMap)
			dataStr = string(bytes)
		} else if payloadSlice, ok := payload.([]interface{}); ok {
			bytes, _ := json.Marshal(payloadSlice)
			dataStr = string(bytes)
		} else {
			dataStr = fmt.Sprintf("%v", payload)
		}

		// Check handling of "done" logic?
		// For now simple pass through.

		// SSE Format
		if eventType != "" {
			// event: type\ndata: ...
			// But standard just uses data usually. simple_sse:
			// "type" in payload?
			// Let's stick to simple: data: ...
		}

		fmt.Fprintf(w, "data: %s\n\n", dataStr)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		return true

	case "close":
		// SSE doesn't really have a close frame from server other than ending connection.
		// We can send [DONE] if convention requires.
		fmt.Fprintf(w, "data: [DONE]\n\n")
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		return true
	}
	return nil
}

// Helper to create Stream Instance
func NewStreamInstance(r *Runtime, w http.ResponseWriter) *Instance {
	if _, ok := r.Classes["Stream"]; !ok {
		return nil
	}
	return &Instance{
		Class: r.Classes["Stream"],
		Fields: map[string]interface{}{
			"_writer": w,
		},
	}
}
