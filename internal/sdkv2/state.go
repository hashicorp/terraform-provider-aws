// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sdkv2

import (
	"strings"
)

// ToUpperSchemaStateFunc converts a string value to uppercase before storing it in state.
func ToUpperSchemaStateFunc(v interface{}) string {
	return strings.ToUpper(v.(string))
}
