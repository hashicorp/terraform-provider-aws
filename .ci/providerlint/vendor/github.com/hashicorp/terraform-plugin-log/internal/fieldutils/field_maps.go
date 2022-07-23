package fieldutils

// MergeFieldMaps takes a slice of field maps,
// and merges all the key/value pairs into a new single field map.
//
// Input order matters: in case two or more maps use the same key,
// the last one to set that key will have the corresponding value
// persisted.
func MergeFieldMaps(maps ...map[string]interface{}) map[string]interface{} {
	// Pre-allocate a map to merge all the maps into,
	// that has at least the capacity equivalent to the number
	// of maps to merge
	result := make(map[string]interface{}, len(maps))

	// Merge all the maps into one;
	// in case of clash, only the last key is preserved
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}

	return result
}

// FieldMapsToKeys will extract all the field maps keys, avoiding repetitions
// in case two or more maps contained the same key.
func FieldMapsToKeys(maps ...map[string]interface{}) []string {
	switch len(maps) {
	case 0:
		return nil
	case 1:
		result := make([]string, 0, len(maps[0]))

		for k := range maps[0] {
			result = append(result, k)
		}

		return result
	default:
		// As we merge all maps into one, we can use this
		// same function recursively, falling back on the `switch case 1`.
		return FieldMapsToKeys(MergeFieldMaps(maps...))
	}
}
