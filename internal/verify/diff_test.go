// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package verify

import (
	"reflect"
	"testing"
	"time"

	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func TestSuppressEquivalentRoundedTime(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		old        string
		new        string
		layout     string
		d          time.Duration
		equivalent bool
	}{
		{
			old:        "2024-04-19T23:00:00.000Z",
			new:        "2024-04-19T23:00:13.000Z",
			layout:     time.RFC3339,
			d:          time.Minute,
			equivalent: true,
		},
		{
			old:        "2024-04-19T23:01:00.000Z",
			new:        "2024-04-19T23:00:45.000Z",
			layout:     time.RFC3339,
			d:          time.Minute,
			equivalent: true,
		},
		{
			old:        "2024-04-19T23:00:00.000Z",
			new:        "2024-04-19T23:00:45.000Z",
			layout:     time.RFC3339,
			d:          time.Minute,
			equivalent: false,
		},
		{
			old:        "2024-04-19T23:00:00.000Z",
			new:        "2024-04-19T23:00:45.000Z",
			layout:     time.RFC3339,
			d:          time.Hour,
			equivalent: true,
		},
	}

	for i, tc := range testCases {
		value := SuppressEquivalentRoundedTime(tc.layout, tc.d)("test_property", tc.old, tc.new, nil)

		if tc.equivalent && !value {
			t.Fatalf("expected test case %d to be equivalent", i)
		}

		if !tc.equivalent && value {
			t.Fatalf("expected test case %d to not be equivalent", i)
		}
	}
}

func TestDiffStringMaps(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Old, New                  map[string]interface{}
		Create, Remove, Unchanged map[string]interface{}
	}{
		// Add
		{
			Old: map[string]interface{}{
				"foo": "bar",
			},
			New: map[string]interface{}{
				"foo": "bar",
				"bar": "baz",
			},
			Create: map[string]interface{}{
				"bar": "baz",
			},
			Remove: map[string]interface{}{},
			Unchanged: map[string]interface{}{
				"foo": "bar",
			},
		},

		// Modify
		{
			Old: map[string]interface{}{
				"foo": "bar",
			},
			New: map[string]interface{}{
				"foo": "baz",
			},
			Create: map[string]interface{}{
				"foo": "baz",
			},
			Remove: map[string]interface{}{
				"foo": "bar",
			},
			Unchanged: map[string]interface{}{},
		},

		// Overlap
		{
			Old: map[string]interface{}{
				"foo":   "bar",
				"hello": "world",
			},
			New: map[string]interface{}{
				"foo":   "baz",
				"hello": "world",
			},
			Create: map[string]interface{}{
				"foo": "baz",
			},
			Remove: map[string]interface{}{
				"foo": "bar",
			},
			Unchanged: map[string]interface{}{
				"hello": "world",
			},
		},

		// Remove
		{
			Old: map[string]interface{}{
				"foo": "bar",
				"bar": "baz",
			},
			New: map[string]interface{}{
				"foo": "bar",
			},
			Create: map[string]interface{}{},
			Remove: map[string]interface{}{
				"bar": "baz",
			},
			Unchanged: map[string]interface{}{
				"foo": "bar",
			},
		},
	}

	for i, tc := range cases {
		c, r, u := DiffStringMaps(tc.Old, tc.New)
		cm := flex.PointersMapToStringList(c)
		rm := flex.PointersMapToStringList(r)
		um := flex.PointersMapToStringList(u)
		if !reflect.DeepEqual(cm, tc.Create) {
			t.Fatalf("%d: bad create: %#v", i, cm)
		}
		if !reflect.DeepEqual(rm, tc.Remove) {
			t.Fatalf("%d: bad remove: %#v", i, rm)
		}
		if !reflect.DeepEqual(um, tc.Unchanged) {
			t.Fatalf("%d: bad unchanged: %#v", i, rm)
		}
	}
}
