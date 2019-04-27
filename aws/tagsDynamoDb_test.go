package aws

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func TestDiffDynamoDbTags(t *testing.T) {
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
		c, r := diffTagsDynamoDb(tagsFromMapDynamoDb(tc.Old), tagsFromMapDynamoDb(tc.New))
		cm := tagsToMapDynamoDb(c)
		if !reflect.DeepEqual(cm, tc.Create) {
			t.Fatalf("%d: bad create: %#v", i, cm)
		}
		for _, key := range r {
			if _, ok := tc.Remove[aws.StringValue(key)]; !ok {
				t.Fatalf("%d: bad remove: %#v", i, key)
			}
		}
	}
}

func TestIgnoringTagsDynamoDb(t *testing.T) {
	ignoredTags := []*dynamodb.Tag{
		{
			Key:   aws.String("aws:cloudformation:logical-id"),
			Value: aws.String("foo"),
		},
		{
			Key:   aws.String("aws:foo:bar"),
			Value: aws.String("baz"),
		},
	}
	for _, tag := range ignoredTags {
		if !tagIgnoredDynamoDb(tag) {
			t.Fatalf("Tag %v with value %v not ignored, but should be!", *tag.Key, *tag.Value)
		}
	}
}
