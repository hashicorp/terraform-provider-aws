package aws

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
)

func TestDiffOrganizationsTags(t *testing.T) {
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
		c, r := diffTagsOrganizations(tagsFromMapOrganizations(tc.Old), tagsFromMapOrganizations(tc.New))
		cm := tagsToMapOrganizations(c)
		rl := []string{}
		for _, tagName := range r {
			rl = append(rl, aws.StringValue(tagName))
		}
		if !reflect.DeepEqual(cm, tc.Create) {
			t.Fatalf("%d: bad create: %#v", i, cm)
		}
		if !reflect.DeepEqual(rl, tc.Remove) {
			t.Fatalf("%d: bad remove: %#v", i, rl)
		}
	}
}

func TestIgnoringTagsOrganizations(t *testing.T) {
	var ignoredTags []*organizations.Tag
	ignoredTags = append(ignoredTags, &organizations.Tag{
		Key:   aws.String("aws:cloudformation:logical-id"),
		Value: aws.String("foo"),
	})
	ignoredTags = append(ignoredTags, &organizations.Tag{
		Key:   aws.String("aws:foo:bar"),
		Value: aws.String("baz"),
	})
	for _, tag := range ignoredTags {
		if !tagIgnoredOrganizations(tag) {
			t.Fatalf("Tag %v with value %v not ignored, but should be!", aws.StringValue(tag.Key), aws.StringValue(tag.Value))
		}
	}
}
