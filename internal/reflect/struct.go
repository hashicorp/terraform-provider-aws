// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package reflect

import (
	"iter"
	"reflect"
)

// StructFields returns an iterator that lists all fields in a struct, including unexported fields.
// If the struct contains an embedded struct, the fields of the embedded struct have the index in both structs.
func StructFields(typ reflect.Type) iter.Seq[reflect.StructField] {
	return func(yield func(reflect.StructField) bool) {
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)

			if field.Anonymous {
				fieldIndexSequence := []int{i}
				for v := range StructFields(field.Type) {
					v.Index = append(fieldIndexSequence, v.Index...)
					if !yield(v) {
						return
					}
				}
				continue
			}

			if !yield(field) {
				return
			}
		}
	}
}
