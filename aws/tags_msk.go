package aws

import (
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	kafka "github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform/helper/schema"
)

// setTags is a helper to set the tags for a resource.  It expects the
// tags field to be named "tags"
func setTagsMskCluster(conn *kafka.Kafka, d *schema.ResourceData, arn string) error {
	if d.HasChange("tags") {
		oraw, nraw := d.GetChange("tags")
		o := oraw.(map[string]interface{})
		n := nraw.(map[string]interface{})
		create, remove := diffTagsMskCluster(tagsFromMapMskCluster(o), tagsFromMapMskCluster(n))

		// Set tags
		if len(remove) > 0 {
			log.Printf("[DEBUG] Removing tags: %#v", remove)
			keys := make([]*string, 0, len(remove))
			for k := range remove {
				keys = append(keys, aws.String(k))
			}
			_, err := conn.UntagResource(&kafka.UntagResourceInput{
				ResourceArn: aws.String(arn),
				TagKeys:     keys,
			})
			if err != nil {
				return err
			}
		}
		if len(create) > 0 {
			log.Printf("[DEBUG] Creating tags: %#v", create)
			_, err := conn.TagResource(&kafka.TagResourceInput{
				ResourceArn: aws.String(arn),
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
func diffTagsMskCluster(oldTags, newTags map[string]*string) (map[string]*string, map[string]*string) {

	// Build the list of what to remove
	remove := make(map[string]*string)
	for k, v := range oldTags {
		newVal, existsInNew := newTags[k]
		if !existsInNew || *newVal != *v {
			// Delete it!
			remove[k] = v
		}
	}
	return newTags, remove
}

// tagsFromMap returns the tags for the given map of data.
func tagsFromMapMskCluster(m map[string]interface{}) map[string]*string {
	result := make(map[string]*string)
	for k, v := range m {
		if !tagIgnoredMskCluster(k, v.(string)) {
			result[k] = aws.String(v.(string))
		}
	}
	return result
}

// tagsToMap turns the list of tags into a map.
func tagsToMapMskCluster(ts map[string]*string) map[string]string {
	result := make(map[string]string)
	for k, v := range ts {
		if !tagIgnoredMskCluster(k, aws.StringValue(v)) {
			result[k] = aws.StringValue(v)
		}
	}
	return result
}

// compare a tag against a list of strings and checks if it should
// be ignored or not
func tagIgnoredMskCluster(key, value string) bool {
	filter := []string{"^aws:"}
	for _, ignore := range filter {
		log.Printf("[DEBUG] Matching %v with %v\n", ignore, key)
		r, _ := regexp.MatchString(ignore, key)
		if r {
			log.Printf("[DEBUG] Found AWS specific tag %s (val %s), ignoring.\n", key, value)
			return true
		}
	}
	return false
}
