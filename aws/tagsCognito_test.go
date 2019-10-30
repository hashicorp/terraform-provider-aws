package aws

import (
	"reflect"
	"testing"
)

// go test -v -run="TestDiffCognitoTags"
func TestDiffCognitoTags(t *testing.T) {
	cases := []struct {
		Old, New map[string]interface{}
		Create   map[string]string
		Remove   []string
	}{
		// Basic add/remove
		{
			Old: map[string]interface{}{
				"foo": "bar",
			},
			New: map[string]interface{}{
				"bar": "baz",
			},
			Create: map[string]string{
				"bar": "baz",
			},
			Remove: []string{
				"foo",
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
			Create: map[string]string{
				"foo": "baz",
			},
			Remove: []string{
				"foo",
			},
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
			Create: map[string]string{
				"foo": "baz",
			},
			Remove: []string{
				"foo",
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
			Create: map[string]string{},
			Remove: []string{
				"bar",
			},
		},
	}

	for i, tc := range cases {
		c, r := diffTagsCognito(tagsFromMapCognito(tc.Old), tagsFromMapCognito(tc.New))

		if !reflect.DeepEqual(c, tc.Create) {
			t.Fatalf("%d: bad create: %#v", i, c)

		}
		if !reflect.DeepEqual(r, tc.Remove) {
			t.Fatalf("%d: bad remove: %#v", i, r)
		}
	}
}
