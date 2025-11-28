package core

// Queue Implementation
func (r *Runtime) executeQueueMethod(instance *Instance, method string, args []interface{}) interface{} {
	if _, ok := instance.Fields["_data"]; !ok {
		instance.Fields["_data"] = []interface{}{}
	}
	data := instance.Fields["_data"].([]interface{})

	switch method {
	case "enqueue":
		if len(args) > 0 {
			instance.Fields["_data"] = append(data, args[0])
		}
		return nil
	case "dequeue":
		if len(data) == 0 {
			return nil
		}
		val := data[0]
		instance.Fields["_data"] = data[1:]
		return val
	case "peek":
		if len(data) == 0 {
			return nil
		}
		return data[0]
	}
	return nil
}
