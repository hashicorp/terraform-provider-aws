package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// setTags is a helper to set the tags for a resource. It expects the
// tags field to be named "tags"
func setTagsCloudWatch(conn *cloudwatch.CloudWatch, d *schema.ResourceData, arn string) error {
	if d.HasChange("tags") {
		oraw, nraw := d.GetChange("tags")
		o := oraw.(map[string]interface{})
		n := nraw.(map[string]interface{})
		create, remove := diffTagsCloudWatch(tagsFromMapCloudWatch(o), tagsFromMapCloudWatch(n))

		// Set tags
		if len(remove) > 0 {
			log.Printf("[DEBUG] Removing tags: %#v", remove)
			k := make([]*string, 0, len(remove))
			for _, t := range remove {
				k = append(k, t.Key)
			}
			_, err := conn.UntagResource(&cloudwatch.UntagResourceInput{
				ResourceARN: aws.String(arn),
				TagKeys:     k,
			})
			if err != nil {
				return err
			}
		}
		if len(create) > 0 {
			log.Printf("[DEBUG] Creating tags: %#v", create)
			_, err := conn.TagResource(&cloudwatch.TagResourceInput{
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
func diffTagsCloudWatch(oldTags, newTags []*cloudwatch.Tag) ([]*cloudwatch.Tag, []*cloudwatch.Tag) {
	// First, we're creating everything we have
	create := make(map[string]interface{})
	for _, t := range newTags {
		create[*t.Key] = *t.Value
	}

	// Build the list of what to remove
	var remove []*cloudwatch.Tag
	for _, t := range oldTags {
		old, ok := create[*t.Key]
		if !ok || old != *t.Value {
			// Delete it!
			remove = append(remove, t)
		}
	}

	return tagsFromMapCloudWatch(create), remove
}

// tagsFromMap returns the tags for the given map of data.
func tagsFromMapCloudWatch(m map[string]interface{}) []*cloudwatch.Tag {
	var result []*cloudwatch.Tag
	for k, v := range m {
		t := &cloudwatch.Tag{
			Key:   aws.String(k),
			Value: aws.String(v.(string)),
		}
		if !tagIgnoredCloudWatch(t) {
			result = append(result, t)
		}
	}

	return result
}

// tagsToMap turns the list of tags into a map.
func tagsToMapCloudWatch(ts []*cloudwatch.Tag) map[string]string {
	result := make(map[string]string)
	for _, t := range ts {
		if !tagIgnoredCloudWatch(t) {
			result[*t.Key] = *t.Value
		}
	}

	return result
}

func saveTagsCloudWatch(conn *cloudwatch.CloudWatch, d *schema.ResourceData, arn string) error {
	resp, err := conn.ListTagsForResource(&cloudwatch.ListTagsForResourceInput{
		ResourceARN: aws.String(arn),
	})

	if err != nil {
		return fmt.Errorf("Error retreiving tags for ARN(%s): %s", arn, err)
	}

	var tagList []*cloudwatch.Tag
	if len(resp.Tags) > 0 {
		tagList = resp.Tags
	}

	return d.Set("tags", tagsToMapCloudWatch(tagList))
}

// compare a tag against a list of strings and checks if it should
// be ignored or not
func tagIgnoredCloudWatch(t *cloudwatch.Tag) bool {
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
