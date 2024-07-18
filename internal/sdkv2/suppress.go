// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sdkv2

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// SuppressEquivalentStringCaseInsensitive provides custom difference suppression
// for strings that are equal under case-insensitivity.
func SuppressEquivalentStringCaseInsensitive(k, old, new string, d *schema.ResourceData) bool {
	return strings.EqualFold(old, new)
}
