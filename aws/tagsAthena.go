package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/hashicorp/terraform/helper/schema"
)

// setTags is a helper to set the tags for a resource. It expects the
// tags field to be named "tags"
func setTagsAthena(conn *athena.Athena, d *schema.ResourceData, arn string) error {
	if d.HasChange("tags") {
		oraw, nraw := d.GetChange("tags")
		o := oraw.(map[string]interface{})
		n := nraw.(map[string]interface{})
		create, remove := diffTagsAthena(tagsFromMapAthena(o), tagsFromMapAthena(n))

		// Set tags
		if len(remove) > 0 {
			input := athena.UntagResourceInput{
				ResourceARN: aws.String(arn),
				TagKeys:     remove,
			}
			log.Printf("[DEBUG] Removing Athena tags: %s", input)
			_, err := conn.UntagResource(&input)
			if err != nil {
				return err
			}
		}
		if len(create) > 0 {
			input := athena.TagResourceInput{
				ResourceARN: aws.String(arn),
				Tags:        create,
			}
			log.Printf("[DEBUG] Adding Athena tags: %s", input)
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
func diffTagsAthena(oldTags, newTags []*athena.Tag) ([]*athena.Tag, []*string) {
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

	return tagsFromMapAthena(create), remove
}

// tagsFromMap returns the tags for the given map of data.
func tagsFromMapAthena(m map[string]interface{}) []*athena.Tag {
	result := make([]*athena.Tag, 0, len(m))
	for k, v := range m {
		t := &athena.Tag{
			Key:   aws.String(k),
			Value: aws.String(v.(string)),
		}
		if !tagIgnoredAthena(t) {
			result = append(result, t)
		}
	}

	return result
}

// tagsToMap turns the list of tags into a map.
func tagsToMapAthena(ts []*athena.Tag) map[string]string {
	result := make(map[string]string)
	for _, t := range ts {
		if !tagIgnoredAthena(t) {
			result[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
		}
	}

	return result
}

func saveTagsAthena(conn *athena.Athena, d *schema.ResourceData, arn string) error {
	resp, err := conn.ListTagsForResource(&athena.ListTagsForResourceInput{
		ResourceARN: aws.String(arn),
	})

	if err != nil {
		return fmt.Errorf("Error retreiving tags for ARN: %s", arn)
	}

	var tagList []*athena.Tag
	if len(resp.Tags) > 0 {
		tagList = resp.Tags
	}

	return d.Set("tags", tagsToMapAthena(tagList))
}

// compare a tag against a list of strings and checks if it should
// be ignored or not
func tagIgnoredAthena(t *athena.Tag) bool {
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
