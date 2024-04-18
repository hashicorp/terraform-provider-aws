// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tftypes

import "fmt"

// walkAttributePath will return the Value that `path` is pointing to within the
// Value. If an error is returned, the AttributePath returned will indicate
// will indicate the steps that remained to be applied when the error was
// encountered.
//
// This implementation, along with one for Type, could be exported to deprecate
// the overly generic WalkAttributePath function.
func (v Value) walkAttributePath(path *AttributePath) (Value, *AttributePath, error) {
	if path == nil || len(path.steps) == 0 {
		return v, path, nil
	}

	nextValueI, err := v.ApplyTerraform5AttributePathStep(path.NextStep())

	if err != nil {
		return Value{}, path, err
	}

	nextValue, ok := nextValueI.(Value)

	if !ok {
		return Value{}, path, fmt.Errorf("unknown type %T returned from tftypes.ApplyTerraform5AttributePathStep", nextValueI)
	}

	return nextValue.walkAttributePath(NewAttributePathWithSteps(path.steps[1:]))
}
