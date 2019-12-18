package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	events "github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/hashicorp/terraform/helper/schema"
)

// setTags is a helper to set the tags for a resource. It expects the
// tags field to be named "tags"
func setTagsCloudWatchEvents(conn *events.CloudWatchEvents, d *schema.ResourceData, arn string) error {
	if d.HasChange("tags") {
		oraw, nraw := d.GetChange("tags")
		o := oraw.(map[string]interface{})
		n := nraw.(map[string]interface{})
		create, remove := diffTagsCloudWatchEvents(tagsFromMapCloudWatchEvents(o), tagsFromMapCloudWatchEvents(n))

		// Set tags
		if len(remove) > 0 {
			log.Printf("[DEBUG] Removing tags: %#v", remove)
			k := make([]*string, 0, len(remove))
			for _, t := range remove {
				k = append(k, t.Key)
			}
			_, err := conn.UntagResource(&events.UntagResourceInput{
				ResourceARN: aws.String(arn),
				TagKeys:     k,
			})
			if err != nil {
				return err
			}
		}
		if len(create) > 0 {
			log.Printf("[DEBUG] Creating tags: %#v", create)
			_, err := conn.TagResource(&events.TagResourceInput{
				ResourceARN: aws.String(arn),
				Tags:        create,
			})
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
func diffTagsCloudWatchEvents(oldTags, newTags []*events.Tag) ([]*events.Tag, []*events.Tag) {
	// First, we're creating everything we have
	create := make(map[string]interface{})
	for _, t := range newTags {
		create[*t.Key] = *t.Value
	}

	// Build the list of what to remove
	var remove []*events.Tag
	for _, t := range oldTags {
		old, ok := create[*t.Key]
		if !ok || old != *t.Value {
			// Delete it!
			remove = append(remove, t)
		}
	}

	return tagsFromMapCloudWatchEvents(create), remove
}

// tagsFromMap returns the tags for the given map of data.
func tagsFromMapCloudWatchEvents(m map[string]interface{}) []*events.Tag {
	var result []*events.Tag
	for k, v := range m {
		t := &events.Tag{
			Key:   aws.String(k),
			Value: aws.String(v.(string)),
		}
		if !tagIgnoredCloudWatchEvents(t) {
			result = append(result, t)
		}
	}

	return result
}

// tagsToMap turns the list of tags into a map.
func tagsToMapCloudWatchEvents(ts []*events.Tag) map[string]string {
	result := make(map[string]string)
	for _, t := range ts {
		if !tagIgnoredCloudWatchEvents(t) {
			result[*t.Key] = *t.Value
		}
	}

	return result
}

func saveTagsCloudWatchEvents(conn *events.CloudWatchEvents, d *schema.ResourceData, arn string) error {
	resp, err := conn.ListTagsForResource(&events.ListTagsForResourceInput{
		ResourceARN: aws.String(arn),
	})

	if err != nil {
		return fmt.Errorf("Error retreiving tags for %s: %s", arn, err)
	}

	var tagList []*events.Tag
	if len(resp.Tags) > 0 {
		tagList = resp.Tags
	}

	return d.Set("tags", tagsToMapCloudWatchEvents(tagList))
}

// compare a tag against a list of strings and checks if it should
// be ignored or not
func tagIgnoredCloudWatchEvents(t *events.Tag) bool {
	filter := []string{"^aws:"}
	for _, v := range filter {
		log.Printf("[DEBUG] Matching %v with %v\n", v, *t.Key)
		r, _ := regexp.MatchString(v, *t.Key)
		if r {
			log.Printf("[DEBUG] Found AWS specific tag %s (val: %s), ignoring.\n", *t.Key, *t.Value)
			return true
		}
	}
	return false
}
