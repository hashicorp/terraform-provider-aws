package aws

import (
	"reflect"
	"testing"
)

// go test -v -run="TestDiffCodeCommitTags"
func TestDiffCodeCommitTags(t *testing.T) {
	cases := []struct {
		Old, New       map[string]interface{}
		Create, Remove map[string]string
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
			Remove: map[string]string{
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
			Create: map[string]string{
				"foo": "baz",
			},
			Remove: map[string]string{
				"foo": "bar",
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
			Remove: map[string]string{
				"foo": "bar",
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
			Remove: map[string]string{
				"bar": "baz",
			},
		},
	}

	for i, tc := range cases {
		c, r := diffTagsCodeCommit(tagsFromMapCodeCommit(tc.Old), tagsFromMapCodeCommit(tc.New))
		cm := tagsToMapCodeCommit(c)
		rm := tagsToMapCodeCommit(r)
		if !reflect.DeepEqual(cm, tc.Create) {
			t.Fatalf("%d: bad create: %#v", i, cm)
		}
		if !reflect.DeepEqual(rm, tc.Remove) {
			t.Fatalf("%d: bad remove: %#v", i, rm)
		}
	}
}

// go test -v -run="TestIgnoringTagsCodeCommit"
func TestIgnoringTagsCodeCommit(t *testing.T) {
	ignoredTags := map[string]string{
		"aws:cloudformation:logical-id": "foo",
		"aws:foo:bar":                   "baz",
	}
	for k, v := range ignoredTags {
		if !tagIgnoredCodeCommit(k, v) {
			t.Fatalf("Tag %v with value %v not ignored, but should be!", k, v)
		}
	}
}
