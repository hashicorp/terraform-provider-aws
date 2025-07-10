// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sdkv2

import (
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
)

// StringCaseInsensitiveSetFunc hashes strings in a case insensitive manner.
// If you want a Set of strings and are case insensitive, this is the SchemaSetFunc you want.
func StringCaseInsensitiveSetFunc(v any) int {
	return create.StringHashcode(strings.ToLower(v.(string)))
}

// SimpleSchemaSetFunc returns a schema.SchemaSetFunc that hashes the given keys values.
func SimpleSchemaSetFunc(keys ...string) schema.SchemaSetFunc {
	return func(v any) int {
		var str strings.Builder

		if m, ok := v.(map[string]any); ok {
			for _, key := range keys {
				if v, ok := m[key]; ok {
					switch v := v.(type) {
					case bool:
						str.WriteRune('-')
						str.WriteString(strconv.FormatBool(v))
					case int:
						str.WriteRune('-')
						str.WriteString(strconv.Itoa(v))
					case string:
						str.WriteRune('-')
						str.WriteString(v)
					}
				}
			}
		}

		return create.StringHashcode(str.String())
	}
}

// HashStringValueMap returns a non-negative hash value for a map[string]string.
func HashStringValueMap(m map[string]string) int {
	if len(m) == 0 {
		return 0
	}

	var str strings.Builder

	for _, k := range slices.Sorted(maps.Keys(m)) {
		str.WriteString(k)
		str.WriteRune('=')
		str.WriteString(m[k])
		str.WriteRune(',')
	}

	return create.StringHashcode(str.String())
}
