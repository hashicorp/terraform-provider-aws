package aws

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
)

func TestDiffAppmeshTags(t *testing.T) {
	cases := []struct {
		Old, New map[string]interface{}
		Create   map[string]string
		Remove   []string
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
			Remove: []string{},
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
		c, r := diffTagsAppmesh(tagsFromMapAppmesh(tc.Old), tagsFromMapAppmesh(tc.New))
		cm := tagsToMapAppmesh(c)
		rl := []string{}
		for _, tagName := range r {
			rl = append(rl, *tagName)
		}
		if !reflect.DeepEqual(cm, tc.Create) {
			t.Fatalf("%d: bad create: %#v", i, cm)
		}
		if !reflect.DeepEqual(rl, tc.Remove) {
			t.Fatalf("%d: bad remove: %#v", i, rl)
		}
	}
}

func TestIgnoringTagsAppmesh(t *testing.T) {
	var ignoredTags []*appmesh.TagRef
	ignoredTags = append(ignoredTags, &appmesh.TagRef{
		Key:   aws.String("aws:cloudformation:logical-id"),
		Value: aws.String("foo"),
	})
	ignoredTags = append(ignoredTags, &appmesh.TagRef{
		Key:   aws.String("aws:foo:bar"),
		Value: aws.String("baz"),
	})
	for _, tag := range ignoredTags {
		if !tagIgnoredAppmesh(tag) {
			t.Fatalf("Tag %v with value %v not ignored, but should be!", *tag.Key, *tag.Value)
		}
	}
}
