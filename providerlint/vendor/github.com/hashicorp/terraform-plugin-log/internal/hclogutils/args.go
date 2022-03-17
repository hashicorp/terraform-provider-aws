package hclogutils

// MapsToArgs will shallow merge field maps into the slice of key1, value1,
// key2, value2, ... arguments expected by hc-log.Logger methods.
func MapsToArgs(maps ...map[string]interface{}) []interface{} {
	switch len(maps) {
	case 0:
		return nil
	case 1:
		result := make([]interface{}, 0, len(maps[0])*2)

		for k, v := range maps[0] {
			result = append(result, k)
			result = append(result, v)
		}

		return result
	default:
		mergedMap := make(map[string]interface{}, 0)

		for _, m := range maps {
			for k, v := range m {
				mergedMap[k] = v
			}
		}

		result := make([]interface{}, 0, len(mergedMap)*2)

		for k, v := range mergedMap {
			result = append(result, k)
			result = append(result, v)
		}

		return result
	}
}
