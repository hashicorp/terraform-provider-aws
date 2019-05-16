package aws

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
)

func TestDiffMskClusterTags(t *testing.T) {
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
	}

	for i, tc := range cases {
		c, r := diffTagsMskCluster(tagsFromMapMskCluster(tc.Old), tagsFromMapMskCluster(tc.New))
		cm := tagsToMapMskCluster(c)
		rm := tagsToMapMskCluster(r)
		if !reflect.DeepEqual(cm, tc.Create) {
			t.Fatalf("%d: bad create: %#v", i, cm)
		}
		if !reflect.DeepEqual(rm, tc.Remove) {
			t.Fatalf("%d: bad remove: %#v", i, rm)
		}
	}
}

func TestTagsToMapMskCluster(t *testing.T) {
	source := map[string]*string{
		"foo": aws.String("bar"),
		"bar": aws.String("baz"),
	}

	inter := make(map[string]interface{})
	for k, v := range tagsToMapMskCluster(source) {
		inter[k] = v
	}

	final := tagsFromMapMskCluster(inter)

	if !reflect.DeepEqual(source, final) {
		t.Fatalf("bad tag transformation: %v -> %v -> %v", source, inter, final)
	}
}

func TestIgnoringTagsMskCluster(t *testing.T) {
	ignoredTags := map[string]string{
		"aws:cloudformation:logical-id": "foo",
		"aws:foo:bar":                   "baz",
	}
	for k, v := range ignoredTags {
		if !tagIgnoredMskCluster(k, v) {
			t.Fatalf("Tag %v with value %v not ignored, but should be!", k, v)
		}
	}
}

func TestCheckMskClusterTags(t *testing.T) {
	tags := make(map[string]*string)
	tags["foo"] = aws.String("bar")
	td := &kafka.ListTagsForResourceOutput{
		Tags: tags,
	}

	testFunc := testAccCheckMskClusterTags(td, "foo", "bar")
	err := testFunc(nil)
	if err != nil {
		t.Fatalf("Failed when expected to succeed: %s", err)
	}
}
