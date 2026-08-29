// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package enum

import (
	"reflect"

	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/sync"
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
	typ := reflect.TypeFor[T]()

	s, ok := valuesCache.Load(typ)
	if ok {
		return s
	}

	// Separates the slow path so that the fast path can be inlined
	return valuesSlow[T]()
}

var valuesCache tfsync.Map[reflect.Type, []string]

func valuesSlow[T Valueser[T]]() []string {
	typ := reflect.TypeFor[T]()
	s, _ := valuesCache.LoadOrStore(
		typ,
		tfslices.Strings(EnumValues[T]()),
	)
	return s
}

func EnumSlice[T ~string](l ...T) []T {
	return l
}

func Slice[T ~string](l ...T) []string {
	return tfslices.Strings(l)
}
