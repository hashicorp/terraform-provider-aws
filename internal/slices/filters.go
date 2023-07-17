// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package slices

func FilterEquals[T comparable](v T) FilterFunc[T] {
	return func(x T) bool {
		return x == v
	}
}
