package core

// Stack Implementation
func (r *Runtime) executeStackMethod(instance *Instance, method string, args []interface{}) interface{} {
	// Initialize internal state if needed
	if _, ok := instance.Fields["_data"]; !ok {
		instance.Fields["_data"] = []interface{}{}
	}
	data := instance.Fields["_data"].([]interface{})

	switch method {
	case "push":
		if len(args) > 0 {
			instance.Fields["_data"] = append(data, args[0])
		}
		return nil
	case "pop":
		if len(data) == 0 {
			return nil
		}
		val := data[len(data)-1]
		instance.Fields["_data"] = data[:len(data)-1]
		return val
	case "peek":
		if len(data) == 0 {
			return nil
		}
		return data[len(data)-1]
	}
	return nil
}
