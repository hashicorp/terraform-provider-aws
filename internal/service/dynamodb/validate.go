package dynamodb

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// http://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_CreateGlobalTable.html
func validGlobalTableName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if (len(value) > 255) || (len(value) < 3) {
		errors = append(errors, fmt.Errorf("%s length must be between 3 and 255 characters: %q", k, value))
	}
	pattern := `^[a-zA-Z0-9_.-]+$`
	if !regexp.MustCompile(pattern).MatchString(value) {
		errors = append(errors, fmt.Errorf("%s must only include alphanumeric, underscore, period, or hyphen characters: %q", k, value))
	}
	return
}

func validStreamSpec(d *schema.ResourceDiff) error {
	enabled := d.Get("stream_enabled").(bool)
	if enabled {
		if v, ok := d.GetOk("stream_view_type"); ok {
			value := v.(string)
			if len(value) == 0 {
				return errors.New("stream_view_type must be non-empty when stream_enabled = true")
			}
			return nil
		}
		return errors.New("stream_view_type is required when stream_enabled = true")
	}
	return nil
}

// checkIfNonKeyAttributesChanged returns true if non_key_attributes between old map and new map are different
func checkIfNonKeyAttributesChanged(oldMap, newMap map[string]interface{}) bool {
	oldNonKeyAttributes, oldNkaExists := oldMap["non_key_attributes"].(*schema.Set)
	newNonKeyAttributes, newNkaExists := newMap["non_key_attributes"].(*schema.Set)

	if oldNkaExists && newNkaExists {
		return !oldNonKeyAttributes.Equal(newNonKeyAttributes)
	}

	return oldNkaExists != newNkaExists
}
