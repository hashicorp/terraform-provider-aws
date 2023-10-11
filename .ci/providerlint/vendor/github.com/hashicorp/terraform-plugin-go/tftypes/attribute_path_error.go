// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tftypes

import (
	"fmt"
)

// AttributePathError represents an error associated with part of a
// tftypes.Value, indicated by the Path property.
type AttributePathError struct {
	Path *AttributePath
	err  error
}

// Equal returns true if two AttributePathErrors are semantically equal. To be
// considered equal, they must have the same path and if errors are set, the
// strings returned by their `Error()` methods must match.
func (a AttributePathError) Equal(o AttributePathError) bool {
	if !a.Path.Equal(o.Path) {
		return false
	}

	if (a.err == nil && o.err != nil) || (a.err != nil && o.err == nil) {
		return false
	}

	if a.err == nil {
		return true
	}

	return a.err.Error() == o.err.Error()
}

func (a AttributePathError) Error() string {
	var path string
	if len(a.Path.Steps()) > 0 {
		path = a.Path.String() + ": "
	}
	return fmt.Sprintf("%s%s", path, a.err)
}

func (a AttributePathError) Unwrap() error {
	return a.err
}
