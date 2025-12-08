// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package reflect

import (
	"iter"
	"reflect"
)

// StructFields returns an iterator that lists all fields in a struct, including unexported fields.
// If the struct contains an embedded struct, the fields of the embedded struct will have
// index components from each struct as used by `reflect.Value.FieldByIndex`
func StructFields(typ reflect.Type) iter.Seq[reflect.StructField] {
	return func(yield func(reflect.StructField) bool) {
		for i := range typ.NumField() {
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

func exportedFields(fields iter.Seq[reflect.StructField]) iter.Seq[reflect.StructField] {
	return func(yield func(reflect.StructField) bool) {
		for field := range fields {
			if !field.IsExported() && !field.Anonymous {
				continue
			}

			if !yield(field) {
				return
			}
		}
	}
}

// ExportedStructFields returns an iterator that lists all exported fields in a struct. If an unexported embedded field
// includes exported fields, the exported embedded fields will be included.
// If the struct contains an embedded struct, the fields of the embedded struct will have
// index components from each struct as used by `reflect.Value.FieldByIndex`
func ExportedStructFields(typ reflect.Type) iter.Seq[reflect.StructField] {
	return exportedFields(StructFields(typ))
}
