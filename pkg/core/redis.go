package core

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	GlobalRedis *redis.Client
	Ctx         = context.Background()
)

// InitRedis initializes the global Redis client
func InitRedis(addr, password string, db int) {
	GlobalRedis = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
}

// EnsureRedisFromEnv auto-initializes GlobalRedis from runtime environment or OS environment
func (r *Runtime) EnsureRedisFromEnv() bool {
	if GlobalRedis != nil {
		return true
	}

	// 1. Try REDIS_URL (e.g. redis://default:pass@host:6379/0)
	redisURL := strings.TrimSpace(r.Env["REDIS_URL"])
	if redisURL == "" {
		redisURL = strings.TrimSpace(os.Getenv("REDIS_URL"))
	}
	if redisURL != "" {
		opts, err := redis.ParseURL(redisURL)
		if err == nil {
			GlobalRedis = redis.NewClient(opts)
			return true
		}
	}

	// 2. Try REDIS_HOST + REDIS_PORT + REDIS_PASSWORD + REDIS_DB
	host := strings.TrimSpace(r.Env["REDIS_HOST"])
	if host == "" {
		host = strings.TrimSpace(os.Getenv("REDIS_HOST"))
	}
	if host != "" {
		port := strings.TrimSpace(r.Env["REDIS_PORT"])
		if port == "" {
			port = strings.TrimSpace(os.Getenv("REDIS_PORT"))
		}
		if port == "" {
			port = "6379"
		}
		if !strings.Contains(host, ":") {
			host = host + ":" + port
		}

		password := r.Env["REDIS_PASSWORD"]
		if password == "" {
			password = os.Getenv("REDIS_PASSWORD")
		}

		user := r.Env["REDIS_USER"]
		if user == "" {
			user = os.Getenv("REDIS_USER")
		}

		dbStr := strings.TrimSpace(r.Env["REDIS_DB"])
		if dbStr == "" {
			dbStr = strings.TrimSpace(os.Getenv("REDIS_DB"))
		}
		db := 0
		if d, err := strconv.Atoi(dbStr); err == nil {
			db = d
		}

		GlobalRedis = redis.NewClient(&redis.Options{
			Addr:     host,
			Username: user,
			Password: password,
			DB:       db,
		})
		return true
	}

	return false
}

// Redis Native Class
func (r *Runtime) executeRedisMethod(instance *Instance, method string, args []interface{}) interface{} {
	if method == "connect" {
		// Manual connection: Redis.connect("localhost:6379", "", 0)
		if len(args) >= 1 {
			addr := args[0].(string)
			password := ""
			db := 0
			if len(args) > 1 {
				password = args[1].(string)
			}
			if len(args) > 2 {
				if d, ok := args[2].(int); ok {
					db = d
				} else if d, ok := args[2].(float64); ok {
					db = int(d)
				}
			}
			InitRedis(addr, password, db)
			return true
		}
		return false
	}

	if GlobalRedis == nil {
		if !r.EnsureRedisFromEnv() {
			fmt.Println("[Redis] Error: Not connected (Call Redis.connect first or configure REDIS_HOST/REDIS_URL in env)")
			return nil
		}
	}

	switch method {
	case "set":
		// Redis.set("key", "value", seconds_ttl)
		if len(args) >= 2 {
			key := fmt.Sprintf("%v", args[0])
			val := args[1]
			ttl := time.Duration(0)
			if len(args) > 2 {
				if t, ok := args[2].(int); ok {
					ttl = time.Duration(t) * time.Second
				} else if t, ok := args[2].(int64); ok {
					ttl = time.Duration(t) * time.Second
				} else if t, ok := args[2].(float64); ok {
					ttl = time.Duration(int(t)) * time.Second
				}
			}
			err := GlobalRedis.Set(Ctx, key, val, ttl).Err()
			if err != nil {
				fmt.Printf("[Redis] Set error: %v\n", err)
				return false
			}
			return true
		}

	case "get":
		// Redis.get("key")
		if len(args) >= 1 {
			key := fmt.Sprintf("%v", args[0])
			val, err := GlobalRedis.Get(Ctx, key).Result()
			if err == redis.Nil {
				return nil
			} else if err != nil {
				fmt.Printf("[Redis] Get error: %v\n", err)
				return nil
			}
			return val
		}

	case "del", "forget":
		// Redis.del("key") or Redis.forget("key")
		if len(args) >= 1 {
			key := fmt.Sprintf("%v", args[0])
			err := GlobalRedis.Del(Ctx, key).Err()
			if err != nil {
				return false
			}
			return true
		}

	case "has":
		// Redis.has("key")
		if len(args) >= 1 {
			key := fmt.Sprintf("%v", args[0])
			count, err := GlobalRedis.Exists(Ctx, key).Result()
			if err != nil || count == 0 {
				return false
			}
			return true
		}

	case "ttl":
		// Redis.ttl("key") -> seconds remaining
		if len(args) >= 1 {
			key := fmt.Sprintf("%v", args[0])
			dur, err := GlobalRedis.TTL(Ctx, key).Result()
			if err != nil {
				return -1
			}
			return int(dur.Seconds())
		}

	case "flush":
		// Redis.flush() -> FlushDB
		err := GlobalRedis.FlushDB(Ctx).Err()
		return err == nil
	}
	return nil
}
