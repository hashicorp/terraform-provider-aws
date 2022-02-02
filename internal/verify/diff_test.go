package verify

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func TestSuppressEquivalentTypeStringBoolean(t *testing.T) {
	testCases := []struct {
		old        string
		new        string
		equivalent bool
	}{
		{
			old:        "false",
			new:        "0",
			equivalent: true,
		},
		{
			old:        "true",
			new:        "1",
			equivalent: true,
		},
		{
			old:        "",
			new:        "0",
			equivalent: false,
		},
		{
			old:        "",
			new:        "1",
			equivalent: false,
		},
	}

	for i, tc := range testCases {
		value := SuppressEquivalentTypeStringBoolean("test_property", tc.old, tc.new, nil)

		if tc.equivalent && !value {
			t.Fatalf("expected test case %d to be equivalent", i)
		}

		if !tc.equivalent && value {
			t.Fatalf("expected test case %d to not be equivalent", i)
		}
	}
}

func TestDiffStringMaps(t *testing.T) {
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
