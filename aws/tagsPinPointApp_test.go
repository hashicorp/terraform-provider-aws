package aws

import (
	"github.com/aws/aws-sdk-go/service/pinpoint"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
)

// go test -v -run="TestDiffTagsPinPointApp"
func TestDiffTagsPinPointApp(t *testing.T) {
	cases := []struct {
		Old, New       map[string]interface{}
		Create, Remove map[string]string
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
			Create: map[string]string{
				"bar": "baz",
			},
			Remove: map[string]string{},
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
		c, r := diffTagsPinPointApp(tagsFromMapPinPointApp(tc.Old), tagsFromMapPinPointApp(tc.New))

		cm := tagsToMapPinPointApp(&pinpoint.TagsModel{Tags: c})
		rm := tagsToMapPinPointApp(&pinpoint.TagsModel{Tags: r})
		if !reflect.DeepEqual(cm, tc.Create) {
			t.Fatalf("%d: bad create: %#v", i, cm)
		}
		if !reflect.DeepEqual(rm, tc.Remove) {
			t.Fatalf("%d: bad remove: %#v", i, rm)
		}
	}
}

// go test -v -run="TestIgnoringTagsPinPointApp"
func TestIgnoringTagsPinPointApp(t *testing.T) {
	var ignoredTags = make(map[string]*string)
	ignoredTags["aws:cloudformation:logical-id"] = aws.String("foo")
	ignoredTags["aws:foo:bar"] = aws.String("baz")
	for key, value := range ignoredTags {
		if !tagIgnoredPinPointApp(key, *value) {
			t.Fatalf("Tag %v with value %v not ignored, but should be!", &key, &value)
		}
	}
}
