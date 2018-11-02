package aws

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestDiffTagsCodeBuild(t *testing.T) {
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
		c, r := diffTagsCodeBuild(tagsFromMapCodeBuild(tc.Old), tagsFromMapCodeBuild(tc.New))
		cm := tagsToMapCodeBuild(c)
		rm := tagsToMapCodeBuild(r)
		if !reflect.DeepEqual(cm, tc.Create) {
			t.Fatalf("%d: bad create: %#v", i, cm)
		}
		if !reflect.DeepEqual(rm, tc.Remove) {
			t.Fatalf("%d: bad remove: %#v", i, rm)
		}
	}
}

func TestIgnoringTagsCodeBuild(t *testing.T) {
	var ignoredTags []*codebuild.Tag
	ignoredTags = append(ignoredTags, &codebuild.Tag{
		Key:   aws.String("aws:cloudformation:logical-id"),
		Value: aws.String("foo"),
	})
	ignoredTags = append(ignoredTags, &codebuild.Tag{
		Key:   aws.String("aws:foo:bar"),
		Value: aws.String("baz"),
	})
	for _, tag := range ignoredTags {
		if !tagIgnoredCodeBuild(tag) {
			t.Fatalf("Tag %v with value %v not ignored, but should be!", *tag.Key, *tag.Value)
		}
	}
}

// testAccCheckTags can be used to check the tags on a resource.
func testAccCheckTagsCodeBuild(
	ts *[]*codebuild.Tag, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		m := tagsToMapCodeBuild(*ts)
		v, ok := m[key]
		if value != "" && !ok {
			return fmt.Errorf("Missing tag: %s", key)
		} else if value == "" && ok {
			return fmt.Errorf("Extra tag: %s", key)
		}
		if value == "" {
			return nil
		}

		if v != value {
			return fmt.Errorf("%s: bad value: %s", key, v)
		}

		return nil
	}
}

// diffTags takes our tags locally and the ones remotely and returns
// the set of tags that must be created, and the set of tags that must
// be destroyed.
func diffTagsCodeBuild(oldTags, newTags []*codebuild.Tag) ([]*codebuild.Tag, []*codebuild.Tag) {
	// First, we're creating everything we have
	create := make(map[string]interface{})
	for _, t := range newTags {
		create[*t.Key] = *t.Value
	}

	// Build the list of what to remove
	var remove []*codebuild.Tag
	for _, t := range oldTags {
		old, ok := create[*t.Key]
		if !ok || old != *t.Value {
			// Delete it!
			remove = append(remove, t)
		}
	}

	return tagsFromMapCodeBuild(create), remove
}

// compare a tag against a list of strings and checks if it should
// be ignored or not
func tagIgnoredCodeBuild(t *codebuild.Tag) bool {
	filter := []string{"^aws:"}
	for _, v := range filter {
		log.Printf("[DEBUG] Matching %v with %v\n", v, *t.Key)
		if r, _ := regexp.MatchString(v, *t.Key); r == true {
			log.Printf("[DEBUG] Found AWS specific tag %s (val: %s), ignoring.\n", *t.Key, *t.Value)
			return true
		}
	}
	return false
}
