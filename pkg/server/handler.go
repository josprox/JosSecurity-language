package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/jossecurity/joss/pkg/core"

	_ "embed"
)

//go:embed default_logo.png
var DefaultLogo []byte

var (
	sessionStore = make(map[string]map[string]interface{})
	sessionMu    sync.Mutex
	// Rate Limiter
	rateLimitStore = make(map[string]*rateLimitEntry)
	rateLimitMu    sync.Mutex
)

type rateLimitEntry struct {
	count    int
	lastTime time.Time
}

func MainHandler(w http.ResponseWriter, r *http.Request) {
	// Request Watchdog: Log progress every second to find hangs
	requestID := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
	requestStartTime := time.Now()
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				fmt.Printf("[WATCHDOG] Request %s still processing (%.0fs)...\n", requestID, time.Since(requestStartTime).Seconds())
			case <-done:
				return
			}
		}
	}()
	defer close(done)

	// 1. Runtime Fork (Isolation)
	// fmt.Printf("[HANDLER] %s: Forking runtime...\n", requestID)
	mutex.RLock()
	if currentRuntime == nil {
		mutex.RUnlock()
		http.Error(w, "Server starting up...", http.StatusServiceUnavailable)
		return
	}
	rt := currentRuntime.Fork()
	mutex.RUnlock()

	// rt.LoadEnv(core.GlobalFileSystem) // Fork already has Env copied

	// 2. Rate Limiting (60 req/min)
	ip := strings.Split(r.RemoteAddr, ":")[0]
	rateLimitMu.Lock()
	entry, exists := rateLimitStore[ip]
	if !exists {
		entry = &rateLimitEntry{count: 0, lastTime: time.Now()}
		rateLimitStore[ip] = entry
	}
	if time.Since(entry.lastTime) > time.Minute {
		entry.count = 0
		entry.lastTime = time.Now()
	}
	entry.count++
	if entry.count > 60 {
		rateLimitMu.Unlock()
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprintf(w, "<h1>429 Too Many Requests</h1>")
		return
	}
	rateLimitMu.Unlock()

	// Panic Recovery
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("[SERVER PANIC] Recovered from: %v\n", r)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "<h1>500 Internal Server Error</h1><p>Something went wrong.</p><pre>%v</pre>", r)
		}
		rt.Free() // Return to pool
	}()

	// Handle Favicon
	if r.URL.Path == "/favicon.ico" {
		if _, err := os.Stat("assets/logo.png"); err == nil {
			http.ServeFile(w, r, "assets/logo.png")
			return
		}
		if _, err := os.Stat("assets/logo.ico"); err == nil {
			http.ServeFile(w, r, "assets/logo.ico")
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.Write(DefaultLogo)
		return
	}

	// 3. Parse Request Data
	reqData := make(map[string]interface{})
	for k, v := range r.URL.Query() {
		if len(v) > 0 {
			reqData[k] = v[0]
		}
	}

	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		var jsonMap map[string]interface{}
		// Use a temporary decoder to avoid EOF errors if body is empty
		if r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&jsonMap); err == nil {
				for k, v := range jsonMap {
					reqData[k] = v
				}
			}
		}
	} else {
		r.ParseMultipartForm(10 << 20) // 10MB
		for k, v := range r.PostForm {
			if len(v) > 0 {
				reqData[k] = v[0]
			}
		}

		// Handle Files
		if r.MultipartForm != nil && r.MultipartForm.File != nil {
			files := make(map[string]interface{})
			for k, fheaders := range r.MultipartForm.File {
				if len(fheaders) > 0 {
					fh := fheaders[0]
					file, err := fh.Open()
					if err == nil {
						// Read content
						content := make([]byte, fh.Size)
						file.Read(content)
						file.Close()

						// Create file object
						fileObj := map[string]interface{}{
							"name":    fh.Filename,
							"type":    fh.Header.Get("Content-Type"),
							"size":    fh.Size,
							"content": string(content), // Store as string for JOSS compatibility
						}
						files[k] = fileObj
					}
				}
			}
			reqData["_files"] = files
		}
	}

	// Inject Headers
	headers := make(map[string]interface{})
	for k, v := range r.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}
	reqData["_headers"] = headers
	// Explicitly map Authorization for cleaner access
	if val := r.Header.Get("Authorization"); val != "" {
		reqData["Authorization"] = val
	}
	reqData["_method"] = r.Method
	reqData["_referer"] = r.Referer()

	// 4. Session Management
	sessionID := ""
	cookie, err := r.Cookie("joss_session")
	if err != nil {
		sessionID = generateSessionID()
		http.SetCookie(w, &http.Cookie{Name: "joss_session", Value: sessionID, Path: "/", HttpOnly: true})
	} else {
		sessionID = cookie.Value
	}

	// fmt.Printf("[HANDLER] %s: Acquiring session lock...\n", requestID)
	sessionMu.Lock()
	// fmt.Printf("[HANDLER] %s: Session lock acquired.\n", requestID)

	var sessData map[string]interface{}
	if rt.Env["SESSION_DRIVER"] == "redis" {
		// Load from Redis
		val, err := core.GlobalRedis.Get(core.Ctx, "session:"+sessionID).Result()
		if err == nil {
			json.Unmarshal([]byte(val), &sessData)
		}
		if sessData == nil {
			sessData = make(map[string]interface{})
		}
		sessionMu.Unlock()
	} else {
		// In-Memory
		if _, ok := sessionStore[sessionID]; !ok {
			sessionStore[sessionID] = make(map[string]interface{})
		}
		// DEEP COPY Session Data
		sourceMap := sessionStore[sessionID]
		sessData = make(map[string]interface{})
		for k, v := range sourceMap {
			sessData[k] = v
		}
		sessionMu.Unlock()
	}
	// fmt.Printf("[HANDLER] %s: Session lock released (Load).\n", requestID)

	// 5. Security Headers
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	if r.TLS != nil {
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	}

	// 6. CSRF Protection
	csrfToken := ""
	if val, ok := sessData["csrf_token"]; ok {
		csrfToken = val.(string)
	} else {
		b := make([]byte, 32)
		rand.Read(b)
		csrfToken = hex.EncodeToString(b)
		sessionMu.Lock()
		if rt.Env["SESSION_DRIVER"] == "redis" {
			sessData["csrf_token"] = csrfToken
		} else {
			sessionStore[sessionID]["csrf_token"] = csrfToken
		}
		sessionMu.Unlock()
		sessData["csrf_token"] = csrfToken
	}

	// Exempt API routes from CSRF
	if (r.Method == "POST" || r.Method == "PUT" || r.Method == "DELETE" || r.Method == "PATCH") && !strings.HasPrefix(r.URL.Path, "/api/") {
		reqToken := ""
		if val, ok := reqData["_token"]; ok {
			reqToken = fmt.Sprintf("%v", val)
		} else {
			reqToken = r.Header.Get("X-CSRF-TOKEN")
		}

		fmt.Printf("[CSRF DEBUG] Session: %s | Stored: %s | Received: %s\n", sessionID, csrfToken, reqToken)

		if reqToken == "" || reqToken != csrfToken {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintf(w, "<h1>419 Page Expired</h1><p>CSRF token mismatch.</p>")
			// Print debug info to browser too (for development)
			fmt.Fprintf(w, "<!-- Debug: Stored='%s' Received='%s' -->", csrfToken, reqToken)
			return
		}
	}

	// 7. Dispatch
	// fmt.Printf("[DEBUG] Dispatching %s %s\n", r.Method, r.URL.Path)
	result, err := rt.Dispatch(r.Method, r.URL.Path, reqData, sessData)

	// 8. Save Session
	if rt.Env["SESSION_DRIVER"] == "redis" {
		data, _ := json.Marshal(sessData)
		core.GlobalRedis.Set(core.Ctx, "session:"+sessionID, data, 24*time.Hour)
	} else {
		// Save In-Memory (Write-Back)
		// fmt.Printf("[HANDLER] %s: Acquiring session lock for write-back...\n", requestID)
		sessionMu.Lock()
		// Overwrite the session data completely to ensure deletions (like logout) are persisted
		sessionStore[sessionID] = sessData
		sessionMu.Unlock()
		// fmt.Printf("[HANDLER] %s: Session lock released (Write-back).\n", requestID)
	}

	if err == nil {
		fmt.Printf("[HANDLER DEBUG] Result Type: %T\n", result)
		// Handle Redirect or JSON (Map)
		if resMap, ok := result.(map[string]interface{}); ok {
			if val, ok := resMap["_type"]; ok {
				fmt.Printf("[HANDLER DEBUG] Map Type: %v\n", val)
				if val == "REDIRECT" {
					http.Redirect(w, r, resMap["url"].(string), http.StatusFound)
					return
				}
				if val == "JSON" {
					w.Header().Set("Content-Type", "application/json")
					statusCode := http.StatusOK
					if code, ok := resMap["status_code"]; ok {
						switch v := code.(type) {
						case int:
							statusCode = v
						case int64:
							statusCode = int(v)
						case float64:
							statusCode = int(v)
						}
					}
					w.WriteHeader(statusCode)
					json.NewEncoder(w).Encode(resMap["data"])
					return
				}
				if val == "RAW" {
					contentType := "text/plain"
					if ct, ok := resMap["content_type"].(string); ok {
						contentType = ct
					}
					w.Header().Set("Content-Type", contentType)

					statusCode := http.StatusOK
					if code, ok := resMap["status_code"]; ok {
						switch v := code.(type) {
						case int:
							statusCode = v
						case int64:
							statusCode = int(v)
						case float64:
							statusCode = int(v)
						}
					}
					w.WriteHeader(statusCode)

					data := resMap["data"]
					switch v := data.(type) {
					case string:
						w.Write([]byte(v))
					case []byte:
						w.Write(v)
					default:
						fmt.Fprintf(w, "%v", v)
					}
					return
				}
			}
		}

		// Handle Redirect or JSON (Instance)
		if resInst, ok := result.(*core.Instance); ok {
			// JSON handling from Instance
			if val, ok := resInst.Fields["_type"]; ok && val == "JSON" {
				w.Header().Set("Content-Type", "application/json")
				statusCode := http.StatusOK
				if code, ok := resInst.Fields["status"]; ok {
					switch v := code.(type) {
					case int:
						statusCode = v
					case int64:
						statusCode = int(v)
					case float64:
						statusCode = int(v)
					}
				}
				w.WriteHeader(statusCode)
				json.NewEncoder(w).Encode(resInst.Fields["data"])
				return
			}

			// RAW handling
			if val, ok := resInst.Fields["_type"]; ok && val == "RAW" {
				contentType := "text/plain"
				if ct, ok := resInst.Fields["content_type"].(string); ok {
					contentType = ct
				}
				w.Header().Set("Content-Type", contentType)

				statusCode := http.StatusOK
				if code, ok := resInst.Fields["status_code"]; ok {
					switch v := code.(type) {
					case int:
						statusCode = v
					case int64:
						statusCode = int(v)
					case float64:
						statusCode = int(v)
					}
				}
				w.WriteHeader(statusCode)

				data := resInst.Fields["data"]
				switch v := data.(type) {
				case string:
					w.Write([]byte(v))
				case []byte:
					w.Write(v)
				default:
					fmt.Fprintf(w, "%v", v)
				}
				return
			}

			// Redirect handling
			if val, ok := resInst.Fields["_type"]; ok && val == "REDIRECT" {
				if flash, ok := resInst.Fields["flash"].(map[string]interface{}); ok {
					sessionMu.Lock()
					if rt.Env["SESSION_DRIVER"] == "redis" {
						for k, v := range flash {
							sessData[k] = v
						}
						data, _ := json.Marshal(sessData)
						core.GlobalRedis.Set(core.Ctx, "session:"+sessionID, data, 24*time.Hour)
					} else {
						if _, ok := sessionStore[sessionID]; !ok {
							sessionStore[sessionID] = make(map[string]interface{})
						}
						for k, v := range flash {
							sessionStore[sessionID][k] = v
						}
					}
					sessionMu.Unlock()
				}
				http.Redirect(w, r, resInst.Fields["url"].(string), http.StatusFound)
				return
			}
		}

		// If result is string, write it
		if str, ok := result.(string); ok {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(str))
			// Hot Reload Script (WebSocket)
			fmt.Fprintf(w, `<script>
				(function() {
					var conn = new WebSocket("ws://" + location.host + "/__hot_reload");
					conn.onmessage = function(evt) {
						if (evt.data === "reload") {
							console.log("Reloading...");
							location.reload();
						}
					};
					conn.onclose = function() {
						console.log("Hot reload connection closed. Reconnecting in 2s...");
						setTimeout(function() { location.reload(); }, 2000);
					};
				})();
			</script>`)
			return
		}
	} else {
		fmt.Printf("[DEBUG] Dispatch error: %v\n", err)
	}

	// Fallback / 404
	if r.URL.Path == "/" {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, "<h1>JosSecurity Server Running</h1>")
		fmt.Fprintf(w, "<p>Environment: %s</p>", rt.Env["APP_ENV"])
		fmt.Fprintf(w, `<link rel="stylesheet" href="/public/css/app.css">`)
		// Hot Reload Script (WebSocket)
		fmt.Fprintf(w, `<script>
			(function() {
				var conn = new WebSocket("ws://" + location.host + "/__hot_reload");
				conn.onmessage = function(evt) {
					if (evt.data === "reload") {
						console.log("Reloading...");
						location.reload();
					}
				};
				conn.onclose = function() {
					console.log("Hot reload connection closed. Reconnecting in 2s...");
					setTimeout(function() { location.reload(); }, 2000);
				};
			})();
		</script>`)
		return
	}

	http.NotFound(w, r)
}

func generateSessionID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
