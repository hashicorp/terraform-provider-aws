package glue

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func mapKeyInSlice(valid []string, ignoreCase bool) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (warnings []string, errors []error) {
		v, ok := i.(map[string]interface{})
		if !ok {
			errors = append(errors, fmt.Errorf("expected type of %[1]q to be Map, got %[1]T", k))
			return warnings, errors
		}

		for key := range v {
			for _, str := range valid {
				if key == str || (ignoreCase && strings.EqualFold(key, str)) {
					return warnings, errors
				}
			}

			errors = append(errors, fmt.Errorf("expected %s to be one of %v, got %s", k, valid, key))
			return warnings, errors
		}

		return warnings, errors
	}
}
