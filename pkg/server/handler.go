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
	r.ParseMultipartForm(10 << 20) // 10MB
	reqData := make(map[string]interface{})
	for k, v := range r.URL.Query() {
		if len(v) > 0 {
			reqData[k] = v[0]
		}
	}
	for k, v := range r.PostForm {
		if len(v) > 0 {
			reqData[k] = v[0]
		}
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
		if sessionStore[sessionID] == nil {
			sessionStore[sessionID] = make(map[string]interface{})
		}
		for k, v := range sessData {
			sessionStore[sessionID][k] = v
		}
		sessionMu.Unlock()
		// fmt.Printf("[HANDLER] %s: Session lock released (Write-back).\n", requestID)
	}

	if err == nil {
		// Handle Redirect (Map)
		if resMap, ok := result.(map[string]interface{}); ok {
			if val, ok := resMap["_type"]; ok && val == "REDIRECT" {
				http.Redirect(w, r, resMap["url"].(string), http.StatusFound)
				return
			}
		}
		// Handle Redirect (Instance)
		if resInst, ok := result.(*core.Instance); ok {
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
			// Hot Reload Script (Disabled for debugging hangs)
			/*
				fmt.Fprintf(w, `<script>
					const evtSource = new EventSource("/__hot_reload");
					evtSource.onmessage = function(event) {
						if (event.data === "reload") {
							console.log("Reloading...");
							location.reload();
						}
					};
				</script>`)
			*/
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
		// Hot Reload Script (Disabled)
		/*
			fmt.Fprintf(w, `<script>
				const evtSource = new EventSource("/__hot_reload");
				evtSource.onmessage = function(event) {
					if (event.data === "reload") {
						console.log("Reloading...");
						location.reload();
					}
				};
			</script>`)
		*/
		return
	}

	http.NotFound(w, r)
}

func generateSessionID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
