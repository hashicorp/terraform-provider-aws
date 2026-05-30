// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cty

import (
	"github.com/hashicorp/go-cty/cty"
)

// HasValue returns true if the given value is known and non-null.
// If the value has a collection type, the value must be non-empty.
func HasValue(value cty.Value) bool {
	if !value.IsKnown() || value.IsNull() {
		return false
	}

	if !value.Type().IsCollectionType() {
		return true
	}

	return value.LengthInt() > 0
}
