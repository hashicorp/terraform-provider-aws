package aws

import (
	"encoding/base64"
	"encoding/json"
	"reflect"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Base64Encode encodes data if the input isn't already encoded using base64.StdEncoding.EncodeToString.
// If the input is already base64 encoded, return the original input unchanged.
func base64Encode(data []byte) string {
	// Check whether the data is already Base64 encoded; don't double-encode
	if isBase64Encoded(data) {
		return string(data)
	}
	// data has not been encoded encode and return
	return base64.StdEncoding.EncodeToString(data)
}

func isBase64Encoded(data []byte) bool {
	_, err := base64.StdEncoding.DecodeString(string(data))
	return err == nil
}

func looksLikeJsonString(s interface{}) bool {
	return regexp.MustCompile(`^\s*{`).MatchString(s.(string))
}

func jsonBytesEqual(b1, b2 []byte) bool {
	var o1 interface{}
	if err := json.Unmarshal(b1, &o1); err != nil {
		return false
	}

	var o2 interface{}
	if err := json.Unmarshal(b2, &o2); err != nil {
		return false
	}

	return reflect.DeepEqual(o1, o2)
}

func isResourceNotFoundError(err error) bool {
	_, ok := err.(*resource.NotFoundError)
	return ok
}

func isResourceTimeoutError(err error) bool {
	timeoutErr, ok := err.(*resource.TimeoutError)
	return ok && timeoutErr.LastError == nil
}

func appendUniqueString(slice []string, elem string) []string {
	for _, e := range slice {
		if e == elem {
			return slice
		}
	}
	return append(slice, elem)
}

// attributesEqual compares two attributes or blocks for equality. This includes
// blocks containing *schema.Set, which reflect.DeepEqual can't handle.
func attributesEqual(a, b interface{}) bool {
	equalMaps := func(a, b map[string]interface{}) bool {
		// equal lengths?
		if len(a) != len(b) {
			return false
		}

		// equal keys?
		for k := range a {
			if _, ok := b[k]; !ok {
				return false
			}
		}

		// equal values?
		for k := range a {
			if !attributesEqual(a[k], b[k]) {
				return false
			}
		}

		return true
	}

	equalSlices := func(a, b []interface{}) bool {
		// equal lengths?
		if len(a) != len(b) {
			return false
		}

		// equal values?
		for i := range a {
			if !attributesEqual(a[i], b[i]) {
				return false
			}
		}

		return true
	}

	if a, aIsMap := a.(map[string]interface{}); aIsMap {
		if b, bIsMap := b.(map[string]interface{}); bIsMap {
			return equalMaps(a, b)
		} else {
			return false
		}
	}

	if a, aIsSlice := a.([]interface{}); aIsSlice {
		if b, bIsSlice := b.([]interface{}); bIsSlice {
			return equalSlices(a, b)
		} else {
			return false
		}
	}

	if a, aIsSet := a.(*schema.Set); aIsSet {
		if b, bIsSet := b.(*schema.Set); bIsSet {
			return a.Equal(b)
		} else {
			return false
		}
	}

	return reflect.DeepEqual(a, b)
}
