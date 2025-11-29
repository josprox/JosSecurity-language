package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/jossecurity/joss/pkg/core"
)

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
	mutex.RLock()
	rt := currentRuntime
	mutex.RUnlock()

	// 0. Rate Limiting (60 req/min)
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
	}()

	// 1. Parse Request Data
	r.ParseMultipartForm(10 << 20) // 10MB
	reqData := make(map[string]interface{})

	// Query Params
	for k, v := range r.URL.Query() {
		if len(v) > 0 {
			reqData[k] = v[0]
		}
	}
	// Form Data
	for k, v := range r.PostForm {
		if len(v) > 0 {
			reqData[k] = v[0]
		}
	}
	// Method
	reqData["_method"] = r.Method
	reqData["_referer"] = r.Referer()

	// 2. Session Management
	sessionID := ""
	cookie, err := r.Cookie("joss_session")
	if err != nil {
		sessionID = generateSessionID()
		http.SetCookie(w, &http.Cookie{Name: "joss_session", Value: sessionID, Path: "/", HttpOnly: true})
	} else {
		sessionID = cookie.Value
	}

	sessionMu.Lock()
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
		sessData = sessionStore[sessionID]
		sessionMu.Unlock()
	}

	// 3. Security Headers
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	if r.TLS != nil {
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	}

	// 4. CSRF Protection
	csrfToken := ""
	if val, ok := sessData["csrf_token"]; ok {
		csrfToken = val.(string)
	} else {
		// Generate new token
		b := make([]byte, 32)
		rand.Read(b)
		csrfToken = hex.EncodeToString(b)
		sessionMu.Lock()
		if rt.Env["SESSION_DRIVER"] == "redis" {
			// Update local map for this request, save later
			sessData["csrf_token"] = csrfToken
		} else {
			sessionStore[sessionID]["csrf_token"] = csrfToken
		}
		sessionMu.Unlock()
		sessData["csrf_token"] = csrfToken
	}

	// Validate CSRF on state-changing methods
	if r.Method == "POST" || r.Method == "PUT" || r.Method == "DELETE" || r.Method == "PATCH" {
		reqToken := ""
		if val, ok := reqData["_token"]; ok {
			reqToken = fmt.Sprintf("%v", val)
		} else {
			reqToken = r.Header.Get("X-CSRF-TOKEN")
		}

		if reqToken == "" || reqToken != csrfToken {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintf(w, "<h1>419 Page Expired</h1><p>CSRF token mismatch.</p>")
			return
		}
	}

	// Dispatch
	fmt.Printf("[DEBUG] Dispatching %s %s\n", r.Method, r.URL.Path)
	result, err := rt.Dispatch(r.Method, r.URL.Path, reqData, sessData)

	// Save Session (if Redis)
	if rt.Env["SESSION_DRIVER"] == "redis" {
		data, _ := json.Marshal(sessData)
		core.GlobalRedis.Set(core.Ctx, "session:"+sessionID, data, 24*time.Hour)
	}
	if err == nil {
		fmt.Printf("[DEBUG] Dispatch success. Result type: %T\n", result)
		// Handle Redirect (Map)
		if resMap, ok := result.(map[string]interface{}); ok {
			if val, ok := resMap["_type"]; ok && val == "REDIRECT" {
				http.Redirect(w, r, resMap["url"].(string), http.StatusFound)
				return
			}
		}
		// Handle Redirect (Instance - e.g. RedirectResponse)
		if resInst, ok := result.(*core.Instance); ok {
			if val, ok := resInst.Fields["_type"]; ok && val == "REDIRECT" {
				// Handle Flash Data
				if flash, ok := resInst.Fields["flash"].(map[string]interface{}); ok {
					sessionMu.Lock()
					if rt.Env["SESSION_DRIVER"] == "redis" {
						// Merge flash into sessData and save
						for k, v := range flash {
							sessData[k] = v
						}
						data, _ := json.Marshal(sessData)
						core.GlobalRedis.Set(core.Ctx, "session:"+sessionID, data, 24*time.Hour)
					} else {
						if _, ok := sessionStore[sessionID]; !ok {
							sessionStore[sessionID] = make(map[string]interface{})
						}
						// Merge flash into session
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
			// ... hot reload script ...
			fmt.Fprintf(w, `<script>
				const evtSource = new EventSource("/__hot_reload");
				evtSource.onmessage = function(event) {
					if (event.data === "reload") {
						console.log("Reloading...");
						location.reload();
					}
				};
			</script>`)
			return
		} else {
			fmt.Println("[DEBUG] Result is not a string or redirect map")
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

		// Hot Reload Script
		fmt.Fprintf(w, `<script>
			const evtSource = new EventSource("/__hot_reload");
			evtSource.onmessage = function(event) {
				if (event.data === "reload") {
					console.log("Reloading...");
					location.reload();
				}
			};
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
