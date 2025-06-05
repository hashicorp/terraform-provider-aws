// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package enum

import (
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
)

type Valueser[T ~string] interface {
	~string
	Values() []T
}

func EnumValues[T Valueser[T]]() []T {
	var zero T
	return zero.Values()
}

func Values[T Valueser[T]]() []string {
	return tfslices.Strings(EnumValues[T]())
}

func Slice[T ~string](l ...T) []string {
	return tfslices.Strings(l)
}
