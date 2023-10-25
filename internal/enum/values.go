// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package enum

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
)

type Valueser[T ~string] interface {
	~string
	Values() []T
}

func Values[T Valueser[T]]() []string {
	l := T("").Values()

	return Slice(l...)
}

func Slice[T Valueser[T]](l ...T) []string {
	result := make([]string, len(l))
	for i, v := range l {
		result[i] = string(v)
	}

	return result
}

func FrameworkDefault[T ~string](t T) defaults.String {
	return stringdefault.StaticString(string(t))
}
