package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/pinpoint"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"regexp"
)

// getTags is a helper to get the tags for a resource. It expects the
// tags field to be named "tags"
func getTagsPinPointApp(conn *pinpoint.Pinpoint, d *schema.ResourceData) error {
	resp, err := conn.ListTagsForResource(&pinpoint.ListTagsForResourceInput{
		ResourceArn: aws.String(d.Get("arn").(string)),
	})
	if err != nil {
		return err
	}

	if err := d.Set("tags", tagsToMapPinPointApp(resp.TagsModel)); err != nil {
		return err
	}

	return nil
}

// tagsToMap turns the list of tags into a map.
func tagsToMapPinPointApp(tm *pinpoint.TagsModel) map[string]string {
	result := make(map[string]string)
	for key, value := range tm.Tags {
		if !tagIgnoredPinPointApp(key, *value) {
			result[key] = aws.StringValue(value)
		}
	}

	return result
}

// compare a tag against a list of strings and checks if it should
// be ignored or not
func tagIgnoredPinPointApp(tagKey string, tagValue string) bool {
	filter := []string{"^aws:"}
	for _, v := range filter {
		log.Printf("[DEBUG] Matching %v with %v\n", v, tagKey)
		r, _ := regexp.MatchString(v, tagKey)
		if r {
			log.Printf("[DEBUG] Found AWS specific tag %s (val: %s), ignoring.\n", tagKey, tagValue)
			return true
		}
	}
	return false
}

// tagsFromMap returns the tags for the given map of data.
func tagsFromMapPinPointApp(m map[string]interface{}) map[string]*string {
	result := make(map[string]*string)
	for k, v := range m {
		if !tagIgnoredPinPointApp(k, v.(string)) {
			result[k] = aws.String(v.(string))
		}
	}

	return result
}

// setTags is a helper to set the tags for a resource. It expects the
// tags field to be named "tags"
func setTagsPinPointApp(conn *pinpoint.Pinpoint, d *schema.ResourceData) error {
	if d.HasChange("tags") {
		oraw, nraw := d.GetChange("tags")
		o := oraw.(map[string]interface{})
		n := nraw.(map[string]interface{})
		create, remove := diffTagsPinPointApp(tagsFromMapPinPointApp(o), tagsFromMapPinPointApp(n))

		// Set tags
		if len(remove) > 0 {
			log.Printf("[DEBUG] Removing tags: %#v", remove)
			k := make([]*string, 0, len(remove))
			for i := range remove {
				k = append(k, &i)
			}

			log.Printf("[DEBUG] Removing old tags: %#v", k)
			log.Printf("[DEBUG] Removing for arn: %#v", aws.String(d.Get("arn").(string)))

			_, err := conn.UntagResource(&pinpoint.UntagResourceInput{
				ResourceArn: aws.String(d.Get("arn").(string)),
				TagKeys:     k,
			})
			if err != nil {
				return err
			}
		}
		if len(create) > 0 {
			log.Printf("[DEBUG] Creating tags: %#v", create)
			_, err := conn.TagResource(&pinpoint.TagResourceInput{
				ResourceArn: aws.String(d.Get("arn").(string)),
				TagsModel:   &pinpoint.TagsModel{Tags: create},
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
func diffTagsPinPointApp(oldTags, newTags map[string]*string) (map[string]*string, map[string]*string) {
	// First, we're creating everything we have
	create := make(map[string]interface{})
	for k, v := range newTags {
		create[k] = aws.StringValue(v)
	}

	// Build the list of what to remove
	var remove = make(map[string]*string)
	for k, v := range oldTags {
		old, ok := create[k]
		if !ok || old != aws.StringValue(v) {
			// Delete it!
			remove[k] = v
		} else if ok {
			delete(create, k)
		}
	}

	return tagsFromMapPinPointApp(create), remove
}
