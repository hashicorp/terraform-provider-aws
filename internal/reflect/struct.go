// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package reflect

import (
	"iter"
	"reflect"
	"strings"

	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
)

// StructFields returns an iterator that lists all fields in a struct, including unexported fields.
// If the struct contains an embedded struct, the fields of the embedded struct will have
// index components from each struct as used by `reflect.Value.FieldByIndex`
func StructFields(typ reflect.Type) iter.Seq[reflect.StructField] {
	return func(yield func(reflect.StructField) bool) {
		structFields_(typ, yield)
	}
}

func structFields_(typ reflect.Type, yield func(reflect.StructField) bool) bool {
	for i := range typ.NumField() {
		field := typ.Field(i)

		if field.Anonymous {
			fieldIndexSequence := []int{i}
			if !structFieldsInner_(field.Type, fieldIndexSequence, yield) {
				return false
			}
			continue
		}

		if !yield(field) {
			return false
		}
	}
	return true
}

func structFieldsInner_(typ reflect.Type, parentIndex []int, yield func(reflect.StructField) bool) bool {
	for i := range typ.NumField() {
		field := typ.Field(i)
		field.Index = append(parentIndex, field.Index...)

		if field.Anonymous {
			fieldIndexSequence := append(parentIndex, i) //nolint:gocritic // append re-assign is intentional
			if !structFieldsInner_(field.Type, fieldIndexSequence, yield) {
				return false
			}
			continue
		}

		if !yield(field) {
			return false
		}
	}
	return true
}

func exportedFields(fields iter.Seq[reflect.StructField]) iter.Seq[reflect.StructField] {
	return tfiter.Filtered(fields, func(field reflect.StructField) bool {
		return field.IsExported() || field.Anonymous
	})
}

// ExportedStructFields returns an iterator that lists all exported fields in a struct. If an unexported embedded field
// includes exported fields, the exported embedded fields will be included.
// If the struct contains an embedded struct, the fields of the embedded struct will have
// index components from each struct as used by `reflect.Value.FieldByIndex`
func ExportedStructFields(typ reflect.Type) iter.Seq[reflect.StructField] {
	return exportedFields(StructFields(typ))
}

// FieldByTag returns the struct field whose tag under tagKey matches tagValue.
// Tag options (e.g. ",omitempty") are stripped before comparison.
func FieldByTag(v any, tagKey, tagValue string) (reflect.StructField, bool) {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return reflect.StructField{}, false
	}
	for f := range t.Fields() {
		if val, ok := f.Tag.Lookup(tagKey); ok {
			if name, _, _ := strings.Cut(val, ","); name == tagValue {
				return f, true
			}
		}
	}
	return reflect.StructField{}, false
}
