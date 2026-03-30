// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cty

import (
	"fmt"

	"github.com/hashicorp/go-cty/cty"
)

// PathSafeApply is equivalent to cty.Path.Apply but does not return an error when one of the steps has a null or unknown value.
// Instead, it returns the null value and a boolean indicating whether the path was fully applied or not.
// Other conditions, such as invalid types or non-existent indexes, will still return an error.
func PathSafeApply(p cty.Path, val cty.Value) (cty.Value, bool, error) {
	var err error

	if val.IsNull() {
		return cty.NilVal, false, nil
	}

	l := len(p)
	for i, step := range p {
		val, err = step.Apply(val)
		if err != nil {
			return cty.NilVal, false, fmt.Errorf("at step %d: %w", i, err)
		}
		if !HasValue(val) && i < l-1 {
			return cty.NilVal, false, nil
		}
	}
	return val, true, nil
}
