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
	return T("").Values()
}

func Values[T Valueser[T]]() []string {
	return Slice(EnumValues[T]()...)
}

func Slice[T ~string](l ...T) []string {
	return tfslices.ApplyToAll(l, func(v T) string {
		return string(v)
	})
}
