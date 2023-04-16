package pipes

func expandString(key string, param map[string]interface{}) *string {
	if val, ok := param[key]; ok {
		if value, ok := val.(string); ok {
			if value != "" {
				return &value
			}
		}
	}
	return nil
}

func expandInt32(key string, param map[string]interface{}) *int32 {
	if val, ok := param[key]; ok {
		if value, ok := val.(int); ok {
			i := int32(value)
			return &i
		}
	}
	return nil
}

func expandBool(key string, param map[string]interface{}) bool {
	if val, ok := param[key]; ok {
		if value, ok := val.(bool); ok {
			return value
		}
	}
	return false
}

func expandStringValue(key string, param map[string]interface{}) string {
	if val, ok := param[key]; ok {
		if value, ok := val.(string); ok {
			return value
		}
	}
	return ""
}
