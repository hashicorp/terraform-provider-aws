package nullable

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	TypeNullableBool = schema.TypeString
)

type Bool string

func (b Bool) IsNull() bool {
	return b == ""
}

func (b Bool) Value() (bool, bool, error) {
	if b.IsNull() {
		return false, true, nil
	}

	value, err := strconv.ParseBool(string(b))
	if err != nil {
		return false, false, err
	}
	return value, false, nil
}

func NewBool(v bool) Bool {
	return Bool(strconv.FormatBool(v))
}

// ValidateTypeStringNullableInt provides custom error messaging for TypeString ints
// Some arguments require an int value or unspecified, empty field.
func ValidateTypeStringNullableBool(v interface{}, k string) (ws []string, es []error) {
	value, ok := v.(string)
	if !ok {
		es = append(es, fmt.Errorf("expected type of %s to be string", k))
		return
	}

	if value == "" {
		return
	}

	if _, err := strconv.ParseBool(value); err != nil {
		es = append(es, fmt.Errorf("%s: cannot parse '%s' as boolean: %w", k, value, err))
	}

	return
}
