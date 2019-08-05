package aws

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/athena"
)

// go test -v -run="TestDiffAthenaTags"
func TestDiffAthenaTags(t *testing.T) {
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
		c, r := diffTagsAthena(tagsFromMapAthena(tc.Old), tagsFromMapAthena(tc.New))
		cm := tagsToMapAthena(c)
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

// go test -v -run="TestIgnoringTagsAthena"
func TestIgnoringTagsAthena(t *testing.T) {
	var ignoredTags []*athena.Tag
	ignoredTags = append(ignoredTags, &athena.Tag{
		Key:   aws.String("aws:cloudformation:logical-id"),
		Value: aws.String("foo"),
	})
	ignoredTags = append(ignoredTags, &athena.Tag{
		Key:   aws.String("aws:foo:bar"),
		Value: aws.String("baz"),
	})
	for _, tag := range ignoredTags {
		if !tagIgnoredAthena(tag) {
			t.Fatalf("Tag %v with value %v not ignored, but should be!", *tag.Key, *tag.Value)
		}
	}
}
