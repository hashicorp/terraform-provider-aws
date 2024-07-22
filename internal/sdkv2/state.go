// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sdkv2

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
)

// NormalizeJsonStringSchemaStateFunc normalizes a JSON string value before storing it in state.
func NormalizeJsonStringSchemaStateFunc(v interface{}) string { // nosemgrep:ci.caps2-in-func-name
	json, _ := structure.NormalizeJsonString(v)
	return json
}

// ToUpperSchemaStateFunc converts a string value to uppercase before storing it in state.
func ToUpperSchemaStateFunc(v interface{}) string {
	return strings.ToUpper(v.(string))
}
