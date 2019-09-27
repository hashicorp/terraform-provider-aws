package aws

import (
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/hashicorp/terraform/helper/schema"
)

// setTags is a helper to set the tags for a resource. It expects the
// tags field to be named "tags"
func setTagsAppmesh(conn *appmesh.AppMesh, d *schema.ResourceData, arn string) error {
	if d.HasChange("tags") {
		oraw, nraw := d.GetChange("tags")
		o := oraw.(map[string]interface{})
		n := nraw.(map[string]interface{})
		create, remove := diffTagsAppmesh(tagsFromMapAppmesh(o), tagsFromMapAppmesh(n))

		// Set tags
		if len(remove) > 0 {
			input := appmesh.UntagResourceInput{
				ResourceArn: aws.String(arn),
				TagKeys:     remove,
			}
			log.Printf("[DEBUG] Removing Appmesh tags: %s", input)
			_, err := conn.UntagResource(&input)
			if err != nil {
				return err
			}
		}
		if len(create) > 0 {
			input := appmesh.TagResourceInput{
				ResourceArn: aws.String(arn),
				Tags:        create,
			}
			log.Printf("[DEBUG] Adding Appmesh tags: %s", input)
			_, err := conn.TagResource(&input)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// diffTags takes our tags locally and the ones remotely and returns
// the set of tags that must be created, and the set of tags that must
// be destroyed.
func diffTagsAppmesh(oldTags, newTags []*appmesh.TagRef) ([]*appmesh.TagRef, []*string) {
	// First, we're creating everything we have
	create := make(map[string]interface{})
	for _, t := range newTags {
		create[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
	}

	// Build the list of what to remove
	var remove []*string
	for _, t := range oldTags {
		old, ok := create[aws.StringValue(t.Key)]
		if !ok || old != aws.StringValue(t.Value) {
			remove = append(remove, t.Key)
		} else if ok {
			// already present so remove from new
			delete(create, aws.StringValue(t.Key))
		}
	}

	return tagsFromMapAppmesh(create), remove
}

// tagsFromMap returns the tags for the given map of data.
func tagsFromMapAppmesh(m map[string]interface{}) []*appmesh.TagRef {
	var result []*appmesh.TagRef
	for k, v := range m {
		t := &appmesh.TagRef{
			Key:   aws.String(k),
			Value: aws.String(v.(string)),
		}
		if !tagIgnoredAppmesh(t) {
			result = append(result, t)
		}
	}

	return result
}

// tagsToMap turns the list of tags into a map.
func tagsToMapAppmesh(ts []*appmesh.TagRef) map[string]string {
	result := make(map[string]string)
	for _, t := range ts {
		if !tagIgnoredAppmesh(t) {
			result[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
		}
	}

	return result
}

func saveTagsAppmesh(conn *appmesh.AppMesh, d *schema.ResourceData, arn string) error {
	resp, err := conn.ListTagsForResource(&appmesh.ListTagsForResourceInput{
		ResourceArn: aws.String(arn),
	})
	if err != nil {
		return err
	}

	var dt []*appmesh.TagRef
	if len(resp.Tags) > 0 {
		dt = resp.Tags
	}

	return d.Set("tags", tagsToMapAppmesh(dt))
}

// compare a tag against a list of strings and checks if it should
// be ignored or not
func tagIgnoredAppmesh(t *appmesh.TagRef) bool {
	filter := []string{"^aws:"}
	for _, v := range filter {
		log.Printf("[DEBUG] Matching %v with %v\n", v, aws.StringValue(t.Key))
		r, _ := regexp.MatchString(v, aws.StringValue(t.Key))
		if r {
			log.Printf("[DEBUG] Found AWS specific tag %s (val: %s), ignoring.\n", aws.StringValue(t.Key), aws.StringValue(t.Value))
			return true
		}
	}
	return false
}
