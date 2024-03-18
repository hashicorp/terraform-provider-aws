// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sdkv2

import (
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
)

// SimpleSchemaSetFunc returns a schema.SchemaSetFunc that hashes the given keys values.
func SimpleSchemaSetFunc(keys ...string) schema.SchemaSetFunc {
	return func(v interface{}) int {
		var str strings.Builder

		m := v.(map[string]interface{})
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

		return create.StringHashcode(str.String())
	}
}
