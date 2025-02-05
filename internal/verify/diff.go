// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package verify

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Find JSON diff functions in the json.go file.

// SuppressEquivalentRoundedTime returns a difference suppression function that compares
// two time value with the specified layout rounded to the specified duration.
func SuppressEquivalentRoundedTime(layout string, d time.Duration) schema.SchemaDiffSuppressFunc {
	return func(k, old, new string, _ *schema.ResourceData) bool {
		if old, err := time.Parse(layout, old); err == nil {
			if new, err := time.Parse(layout, new); err == nil {
				return old.Round(d).Equal(new.Round(d))
			}
		}

		return false
	}
}

// SuppressMissingOptionalConfigurationBlock handles configuration block attributes in the following scenario:
//   - The resource schema includes an optional configuration block with defaults
//   - The API response includes those defaults to refresh into the Terraform state
//   - The operator's configuration omits the optional configuration block
func SuppressMissingOptionalConfigurationBlock(k, old, new string, d *schema.ResourceData) bool {
	return old == "1" && new == "0"
}
