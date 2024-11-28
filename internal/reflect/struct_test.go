// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package reflect

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type ExampleStruct struct {
	Field1          string
	Field2          int
	unexportedField string //nolint:unused // Used for testing unexported fields
}

type unexportedStruct struct {
	Field1          string
	Field2          int
	unexportedField string //nolint:unused // Used for testing unexported fields
}

type NestedEmbedStruct struct {
	Field3 bool
	ExampleStruct
	Field4 string
}

func TestStructFields(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		in       any
		expected []reflect.StructField
	}{
		"empty struct": {
			in:       struct{}{},
			expected: nil,
		},
		"basic struct": {
			in: ExampleStruct{},
			expected: []reflect.StructField{
				{
					Name:      "Field1",
					Index:     []int{0},
					Anonymous: false,
				},
				{
					Name:      "Field2",
					Index:     []int{1},
					Anonymous: false,
				},
				{
					Name:      "unexportedField",
					Index:     []int{2},
					Anonymous: false,
				},
			},
		},
		"embedded struct": {
			in: struct {
				ExampleStruct
				Field3 bool
			}{},
			expected: []reflect.StructField{
				{
					Name:      "Field1",
					Index:     []int{0, 0},
					Anonymous: false,
				},
				{
					Name:      "Field2",
					Index:     []int{0, 1},
					Anonymous: false,
				},
				{
					Name:      "unexportedField",
					Index:     []int{0, 2},
					Anonymous: false,
				},
				{
					Name:      "Field3",
					Index:     []int{1},
					Anonymous: false,
				},
			},
		},
		"unexported embedded struct": {
			in: struct {
				unexportedStruct
				Field3 bool
			}{},
			expected: []reflect.StructField{
				{
					Name:      "Field1",
					Index:     []int{0, 0},
					Anonymous: false,
				},
				{
					Name:      "Field2",
					Index:     []int{0, 1},
					Anonymous: false,
				},
				{
					Name:      "unexportedField",
					Index:     []int{0, 2},
					Anonymous: false,
				},
				{
					Name:      "Field3",
					Index:     []int{1},
					Anonymous: false,
				},
			},
		},
		"nested embedded struct": {
			in: struct {
				NestedEmbedStruct
				Field5 bool
			}{},
			expected: []reflect.StructField{
				{
					Name:      "Field3",
					Index:     []int{0, 0},
					Anonymous: false,
				},
				{
					Name:      "Field1",
					Index:     []int{0, 1, 0},
					Anonymous: false,
				},
				{
					Name:      "Field2",
					Index:     []int{0, 1, 1},
					Anonymous: false,
				},
				{
					Name:      "unexportedField",
					Index:     []int{0, 1, 2},
					Anonymous: false,
				},
				{
					Name:      "Field4",
					Index:     []int{0, 2},
					Anonymous: false,
				},
				{
					Name:      "Field5",
					Index:     []int{1},
					Anonymous: false,
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var out []reflect.StructField
			for v := range StructFields(reflect.TypeOf(testCase.in)) {
				out = append(out, v)
			}

			if diff := cmp.Diff(out, testCase.expected, cmpopts.IgnoreFields(reflect.StructField{}, "PkgPath", "Type", "Tag", "Offset")); diff != "" {
				t.Errorf("Mismatched results: %s", diff)
			}
		})
	}
}

func TestExportedStructFields(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		in       any
		expected []reflect.StructField
	}{
		"empty struct": {
			in:       struct{}{},
			expected: nil,
		},
		"basic struct": {
			in: ExampleStruct{},
			expected: []reflect.StructField{
				{
					Name:      "Field1",
					Index:     []int{0},
					Anonymous: false,
				},
				{
					Name:      "Field2",
					Index:     []int{1},
					Anonymous: false,
				},
			},
		},
		"embedded struct": {
			in: struct {
				ExampleStruct
				Field3 bool
			}{},
			expected: []reflect.StructField{
				{
					Name:      "Field1",
					Index:     []int{0, 0},
					Anonymous: false,
				},
				{
					Name:      "Field2",
					Index:     []int{0, 1},
					Anonymous: false,
				},
				{
					Name:      "Field3",
					Index:     []int{1},
					Anonymous: false,
				},
			},
		},
		"unexported embedded struct": {
			in: struct {
				unexportedStruct
				Field3 bool
			}{},
			expected: []reflect.StructField{
				{
					Name:      "Field1",
					Index:     []int{0, 0},
					Anonymous: false,
				},
				{
					Name:      "Field2",
					Index:     []int{0, 1},
					Anonymous: false,
				},
				{
					Name:      "Field3",
					Index:     []int{1},
					Anonymous: false,
				},
			},
		},
		"nested embedded struct": {
			in: struct {
				NestedEmbedStruct
				Field5 bool
			}{},
			expected: []reflect.StructField{
				{
					Name:      "Field3",
					Index:     []int{0, 0},
					Anonymous: false,
				},
				{
					Name:      "Field1",
					Index:     []int{0, 1, 0},
					Anonymous: false,
				},
				{
					Name:      "Field2",
					Index:     []int{0, 1, 1},
					Anonymous: false,
				},
				{
					Name:      "Field4",
					Index:     []int{0, 2},
					Anonymous: false,
				},
				{
					Name:      "Field5",
					Index:     []int{1},
					Anonymous: false,
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var out []reflect.StructField
			for v := range ExportedStructFields(reflect.TypeOf(testCase.in)) {
				out = append(out, v)
			}

			if diff := cmp.Diff(out, testCase.expected, cmpopts.IgnoreFields(reflect.StructField{}, "PkgPath", "Type", "Tag", "Offset")); diff != "" {
				t.Errorf("Mismatched results: %s", diff)
			}
		})
	}
}
