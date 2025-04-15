// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sdkv2

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
)

// NormalizeJsonStringSchemaStateFunc normalizes a JSON string value before storing it in state.
func NormalizeJsonStringSchemaStateFunc(v any) string { // nosemgrep:ci.caps2-in-func-name
	json, _ := structure.NormalizeJsonString(v)
	return json
}

// ToLowerSchemaStateFunc converts a string value to lowercase before storing it in state.
func ToLowerSchemaStateFunc(v any) string {
	return strings.ToLower(v.(string))
}

// ToUpperSchemaStateFunc converts a string value to uppercase before storing it in state.
func ToUpperSchemaStateFunc(v any) string {
	return strings.ToUpper(v.(string))
}
