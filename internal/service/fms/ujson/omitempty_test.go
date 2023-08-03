package ujson_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-provider-aws/internal/service/fms/ujson"
)

func TestOmitEmpty(t *testing.T) {
	input := "{\"type\":\"SECURITY_GROUPS_CONTENT_AUDIT\",\"preManagedOptions1\":null,\"securityGroups\":[{\"id\":\"sg-041cb51d2ebd60360\"}],\"securityGroupAction\":{\"type\":\"ALLOW\",\"excludeRules\":[]},\"preManagedOptions1\":null}"
	want := "{\"type\":\"SECURITY_GROUPS_CONTENT_AUDIT\",\"securityGroups\":[{\"id\":\"sg-041cb51d2ebd60360\"}],\"securityGroupAction\":{\"type\":\"ALLOW\"}}"
	b := make([]byte, 0, len(want))
	lenBeforeArray := 0

	err := ujson.Walk([]byte(input), func(l int, key, value []byte) bool {
		n := len(b)

		// For valid JSON, values will never be empty.
		skip := false
		switch value[0] {
		case 'n': // Null (null)
			skip = true
		case '[': // Start of array
			lenBeforeArray = n
		case ']': // End of array
			if b[n-1] == '[' {
				b = b[:lenBeforeArray]
				lenBeforeArray = 0
				skip = true
			}
		}

		if skip {
			return false
		}

		if n != 0 && ujson.ShouldAddComma(value, b[n-1]) {
			b = append(b, ',')
		}
		if len(key) > 0 {
			b = append(b, key...)
			b = append(b, ':')
		}
		b = append(b, value...)

		return true
	})

	if gotErr := err != nil; gotErr {
		t.Errorf("err = %q", err)
	} else if diff := cmp.Diff(string(b), want); diff != "" {
		t.Errorf("unexpected diff (+wanted, -got): %s", diff)
	}
}
