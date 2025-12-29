package core

import (
	"sync"
	"time"
)

var (
	globalCache = sync.Map{}
)

type cacheItem struct {
	Value      interface{}
	Expiration int64 // Unix timestamp
}

// Cache Handler
func (r *Runtime) executeCacheMethod(instance *Instance, method string, args []interface{}) interface{} {
	switch method {
	case "put":
		// Cache.put(key, value, seconds)
		if len(args) < 2 {
			return nil
		}
		key, ok := args[0].(string)
		val := args[1]
		seconds := int64(60) // Default 1 min
		if len(args) > 2 {
			if s, ok := args[2].(int64); ok {
				seconds = s
			} else if s, ok := args[2].(int); ok {
				seconds = int64(s)
			}
		}

		if !ok {
			return nil
		}

		item := cacheItem{
			Value:      val,
			Expiration: time.Now().Unix() + seconds,
		}
		globalCache.Store(key, item)
		return true

	case "get":
		// Cache.get(key, default?)
		if len(args) < 1 {
			return nil
		}
		key, ok := args[0].(string)
		if !ok {
			return nil
		}

		if val, ok := globalCache.Load(key); ok {
			item := val.(cacheItem)
			if time.Now().Unix() > item.Expiration {
				globalCache.Delete(key)
				return nil
			}
			return item.Value
		}
		if len(args) > 1 {
			return args[1] // Default
		}
		return nil

	case "has":
		if len(args) < 1 {
			return false
		}
		key, ok := args[0].(string)
		if !ok {
			return false
		}
		if val, ok := globalCache.Load(key); ok {
			item := val.(cacheItem)
			if time.Now().Unix() > item.Expiration {
				globalCache.Delete(key)
				return false
			}
			return true
		}
		return false

	case "forget":
		if len(args) < 1 {
			return nil
		}
		key, ok := args[0].(string)
		if !ok {
			return nil
		}
		globalCache.Delete(key)
		return true
	}
	return nil
}
