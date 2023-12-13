// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package enum

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

func Slice[T Valueser[T]](l ...T) []string {
	result := make([]string, len(l))
	for i, v := range l {
		result[i] = string(v)
	}

	return result
}
