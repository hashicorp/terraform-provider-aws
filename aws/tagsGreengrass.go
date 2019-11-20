package aws

import (
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/greengrass"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// getTags is a helper to get the tags for a resource. It expects the
// tags field to be named "tags" and the ARN field to be named "arn".
func getTagsGreengrass(conn *greengrass.Greengrass, d *schema.ResourceData) error {
	req := &greengrass.ListTagsForResourceInput{
		ResourceArn: aws.String(d.Get("arn").(string)),
	}
	resp, err := conn.ListTagsForResource(req)
	if err != nil {
		return err
	}

	if err := d.Set("tags", tagsToMapGreengrass(resp.Tags)); err != nil {
		return err
	}

	return nil
}

// setTags is a helper to set the tags for a resource. It expects the
// tags field to be named "tags" and the ARN field to be named "arn".
func setTagsGreengrass(conn *greengrass.Greengrass, d *schema.ResourceData) error {
	if d.HasChange("tags") {
		oraw, nraw := d.GetChange("tags")
		o := oraw.(map[string]interface{})
		n := nraw.(map[string]interface{})
		create, remove := diffTagsGreengrass(o, n)

		// Set tags
		if len(remove) > 0 {
			log.Printf("[DEBUG] Removing tags: %#v", remove)
			_, err := conn.UntagResource(&greengrass.UntagResourceInput{
				ResourceArn: aws.String(d.Get("arn").(string)),
				TagKeys:     remove,
			})
			if err != nil {
				return err
			}
		}
		if len(create) > 0 {
			log.Printf("[DEBUG] Creating tags: %#v", create)
			_, err := conn.TagResource(&greengrass.TagResourceInput{
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
func diffTagsGreengrass(oldTags, newTags map[string]interface{}) (map[string]*string, []*string) {
	// First, we're creating everything we have
	create := make(map[string]*string)
	for key, value := range newTags {
		create[key] = aws.String(value.(string))
	}

	// Build the list of what to remove
	var remove []*string
	for key, oldValue := range oldTags {
		new, ok := create[key]
		if !ok || *new != oldValue.(string) {
			// old key is not preset in create collection or values that saved under this key are not
			// equal each other in both collection, so append tag key to `remove` array
			remove = append(remove, aws.String(key))
		} else {
			// already present so remove from `create` map
			delete(create, key)
		}
	}

	return create, remove
}

// tagsToMap turns the list of tags into a map.
func tagsToMapGreengrass(ts map[string]*string) map[string]string {
	result := make(map[string]string)
	for key, value := range ts {
		if !tagIgnoredGreengrass(key, *value) {
			result[key] = aws.StringValue(value)
		}
	}

	return result
}

// compare a tag against a list of strings and checks if it should
// be ignored or not
func tagIgnoredGreengrass(tagKey, tagValue string) bool {
	filter := []string{"^aws:"}
	for _, v := range filter {
		log.Printf("[DEBUG] Matching %v with %v\n", v, tagKey)
		if r, _ := regexp.MatchString(v, tagKey); r {
			log.Printf("[DEBUG] Found AWS specific tag %s (val: %s), ignoring.\n", tagKey, tagValue)
			return true
		}
	}
	return false
}
