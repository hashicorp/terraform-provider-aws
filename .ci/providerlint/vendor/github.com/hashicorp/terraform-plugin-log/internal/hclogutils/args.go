package hclogutils

import (
	"github.com/hashicorp/terraform-plugin-log/internal/fieldutils"
)

// FieldMapsToArgs will shallow merge field maps into a slice of key/value pairs
// arguments (i.e. `[k1, v1, k2, v2, ...]`) expected by hc-log.Logger methods.
func FieldMapsToArgs(maps ...map[string]interface{}) []interface{} {
	switch len(maps) {
	case 0:
		return nil
	case 1:
		result := make([]interface{}, 0, len(maps[0])*2)

		for k, v := range maps[0] {
			result = append(result, k, v)
		}

		return result
	default:
		// As we merge all maps into one, we can use this
		// same function recursively, falling back on the `switch case 1`.
		return FieldMapsToArgs(fieldutils.MergeFieldMaps(maps...))
	}
}
