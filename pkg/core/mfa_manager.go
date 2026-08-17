package core

// EnsureMFATables creates MFA related tables if they don't exist
func (r *Runtime) EnsureMFATables() {
	if r.GetDB() == nil {
		return
	}

	// 1. Create MFA Methods Table
	r.executeSchemaMethod(nil, "create", []interface{}{
		"user_mfa_methods",
		map[string]interface{}{
			"id":          "increments",
			"user_id":     "bigInteger",
			"method_type": "string(50)",
			"secret":      "text|nullable",
			"is_active":   "boolean|default(0)",
			"created_at":  "timestamp|nullable",
			"updated_at":  "timestamp|nullable",
		},
	})

	// 2. Create Recovery Codes Table
	r.executeSchemaMethod(nil, "create", []interface{}{
		"user_recovery_codes",
		map[string]interface{}{
			"id":         "increments",
			"user_id":    "bigInteger",
			"code_hash":  "string(255)",
			"used":       "boolean|default(0)",
			"used_at":    "timestamp|nullable",
			"created_at": "timestamp|nullable",
			"updated_at": "timestamp|nullable",
		},
	})

	// 3. Create Security Logs Table
	r.executeSchemaMethod(nil, "create", []interface{}{
		"security_logs",
		map[string]interface{}{
			"id":         "increments",
			"user_id":    "bigInteger|nullable",
			"event_type": "string(100)",
			"ip_address": "string(45)|nullable",
			"user_agent": "text|nullable",
			"created_at": "timestamp|nullable",
		},
	})
}
