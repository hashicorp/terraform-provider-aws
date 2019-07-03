package aws

import (
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codecommit"
	"github.com/hashicorp/terraform/helper/schema"
)

// setTags is a helper to set the tags for a resource. It expects the
// tags field to be named "tags"
func setTagsCodeCommit(conn *codecommit.CodeCommit, d *schema.ResourceData) error {
	if d.HasChange("tags") {
		oraw, nraw := d.GetChange("tags")
		o := oraw.(map[string]interface{})
		n := nraw.(map[string]interface{})
		create, remove := diffTagsCodeCommit(tagsFromMapCodeCommit(o), tagsFromMapCodeCommit(n))

		// Set tags
		if len(remove) > 0 {
			log.Printf("[DEBUG] Removing tags: %#v", remove)
			keys := make([]*string, 0, len(remove))
			for k := range remove {
				keys = append(keys, aws.String(k))
			}

			_, err := conn.UntagResource(&codecommit.UntagResourceInput{
				ResourceArn: aws.String(d.Get("arn").(string)),
				TagKeys:     keys,
			})
			if err != nil {
				return err
			}
		}
		if len(create) > 0 {
			log.Printf("[DEBUG] Creating tags: %#v", create)
			_, err := conn.TagResource(&codecommit.TagResourceInput{
				ResourceArn: aws.String(d.Get("arn").(string)),
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
func diffTagsCodeCommit(oldTags, newTags map[string]*string) (map[string]*string, map[string]*string) {
	// Build the list of what to remove
	remove := make(map[string]*string)
	for k, v := range oldTags {
		newVal, existsInNew := newTags[k]
		if !existsInNew || *newVal != *v {
			// Delete it!
			remove[k] = v
		} else if existsInNew {
			delete(newTags, k)
		}
	}
	return newTags, remove
}

// tagsFromMap returns the tags for the given map of data.
func tagsFromMapCodeCommit(m map[string]interface{}) map[string]*string {
	result := make(map[string]*string)
	for k, v := range m {
		if !tagIgnoredCodeCommit(k, v.(string)) {
			result[k] = aws.String(v.(string))
		}
	}

	return result
}

// tagsToMap turns the list of tags into a map.
func tagsToMapCodeCommit(ts map[string]*string) map[string]string {
	result := make(map[string]string)
	for k, v := range ts {
		if !tagIgnoredCodeCommit(k, aws.StringValue(v)) {
			result[k] = aws.StringValue(v)
		}
	}

	return result
}

// compare a tag against a list of strings and checks if it should
// be ignored or not
func tagIgnoredCodeCommit(key, value string) bool {
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
