package core

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

func (r *Runtime) executeMathMethod(instance *Instance, method string, args []interface{}) interface{} {
	switch method {
	case "random":
		// random(min, max)
		if len(args) != 2 {
			fmt.Println("Error: Math.random requiere 2 argumentos (min, max)")
			return nil
		}
		min, ok1 := args[0].(int64)
		max, ok2 := args[1].(int64)
		if !ok1 || !ok2 {
			fmt.Println("Error: Argumentos de Math.random deben ser enteros")
			return nil
		}
		rand.Seed(time.Now().UnixNano())
		return min + rand.Int63n(max-min+1)

	case "floor":
		if len(args) != 1 {
			return nil
		}
		if f, ok := args[0].(float64); ok {
			return math.Floor(f)
		}
		return args[0] // If int, return as is

	case "ceil":
		if len(args) != 1 {
			return nil
		}
		if f, ok := args[0].(float64); ok {
			return math.Ceil(f)
		}
		return args[0]

	case "abs":
		if len(args) != 1 {
			return nil
		}
		if i, ok := args[0].(int64); ok {
			if i < 0 {
				return -i
			}
			return i
		}
		if f, ok := args[0].(float64); ok {
			return math.Abs(f)
		}
		return nil
	}
	return nil
}

func (r *Runtime) executeSessionMethod(instance *Instance, method string, args []interface{}) interface{} {
	// We need access to the session map.
	// In `auth.go`, we inject `$__session` into `r.Variables`.
	// We can access it from there.

	sessionVal, ok := r.Variables["$__session"]
	if !ok {
		// fmt.Println("Error: Sesión no disponible en este contexto")
		return nil
	}

	var sessionMap map[string]interface{}

	if sessMap, ok := sessionVal.(map[string]interface{}); ok {
		sessionMap = sessMap
	} else if sessInst, ok := sessionVal.(*Instance); ok {
		sessionMap = sessInst.Fields
	} else {
		return nil
	}

	switch method {
	case "get":
		if len(args) != 1 {
			return nil
		}
		key, ok := args[0].(string)
		if !ok {
			return nil
		}
		return sessionMap[key]

	case "put":
		if len(args) != 2 {
			return nil
		}
		key, ok := args[0].(string)
		if !ok {
			return nil
		}
		sessionMap[key] = args[1]

	case "has":
		if len(args) != 1 {
			return false
		}
		key, ok := args[0].(string)
		if !ok {
			return false
		}
		_, exists := sessionMap[key]
		return exists

	case "forget":
		if len(args) != 1 {
			return nil
		}
		key, ok := args[0].(string)
		if !ok {
			return nil
		}
		delete(sessionMap, key)

	case "all":
		return sessionMap
	}

	return nil
}
