// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package enum

import (
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

type Valueser[T ~string] interface {
	~string
	Values() []T
}

func EnumValues[T Valueser[T]]() []T {
	return inttypes.Zero[T]().Values()
}

func Values[T Valueser[T]]() []string {
	return tfslices.Strings(EnumValues[T]())
}

func EnumSlice[T ~string](l ...T) []T {
	return l
}

func Slice[T ~string](l ...T) []string {
	return tfslices.Strings(l)
}
