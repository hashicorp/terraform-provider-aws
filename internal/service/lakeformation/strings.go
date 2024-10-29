// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation

import (
	"reflect"
	"slices"
)

func StringSlicesEqualIgnoreOrder(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}

	v1 := s1
	v2 := s2

	slices.Sort(v1)
	slices.Sort(v2)

	return reflect.DeepEqual(v1, v2)
}
