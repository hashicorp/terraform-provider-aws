package aws

import (
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iotevents"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// getTags is a helper to get the tags for a resource. It expects the
// tags field to be named "tags" and the ARN field to be named "arn".
func getTagsIotEvents(conn *iotevents.IoTEvents, d *schema.ResourceData) error {
	req := &iotevents.ListTagsForResourceInput{
		ResourceArn: aws.String(d.Get("arn").(string)),
	}
	resp, err := conn.ListTagsForResource(req)
	if err != nil {
		return err
	}

	if err := d.Set("tags", tagsToMapIotEvents(resp.Tags)); err != nil {
		return err
	}

	return nil
}

// setTags is a helper to set the tags for a resource. It expects the
// tags field to be named "tags" and the ARN field to be named "arn".
func setTagsIotEvents(conn *iotevents.IoTEvents, d *schema.ResourceData) error {
	if d.HasChange("tags") {
		oraw, nraw := d.GetChange("tags")
		o := oraw.(map[string]interface{})
		n := nraw.(map[string]interface{})
		create, remove := diffTagsIotEvents(tagsFromMapIotEvents(o), tagsFromMapIotEvents(n))

		// Set tags
		if len(remove) > 0 {
			log.Printf("[DEBUG] Removing tags: %#v", remove)
			k := make([]*string, len(remove))
			for i, t := range remove {
				k[i] = t.Key
			}

			_, err := conn.UntagResource(&iotevents.UntagResourceInput{
				ResourceArn: aws.String(d.Get("arn").(string)),
				TagKeys:     k,
			})
			if err != nil {
				return err
			}
		}
		if len(create) > 0 {
			log.Printf("[DEBUG] Creating tags: %#v", create)
			_, err := conn.TagResource(&iotevents.TagResourceInput{
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
func diffTagsIotEvents(oldTags, newTags []*iotevents.Tag) ([]*iotevents.Tag, []*iotevents.Tag) {
	// First, we're creating everything we have
	create := make(map[string]interface{})
	for _, t := range newTags {
		create[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
	}

	// Build the list of what to remove
	var remove []*iotevents.Tag
	for _, old := range oldTags {
		new, ok := create[aws.StringValue(old.Key)]
		if !ok || new != aws.StringValue(old.Value) {
			remove = append(remove, old)
		} else {
			// already present so remove from new
			delete(create, aws.StringValue(old.Key))
		}
	}

	return tagsFromMapIotEvents(create), remove
}

// tagsFromMap returns the tags for the given map of data.
func tagsFromMapIotEvents(m map[string]interface{}) []*iotevents.Tag {
	result := make([]*iotevents.Tag, 0, len(m))
	for k, v := range m {
		t := &iotevents.Tag{
			Key:   aws.String(k),
			Value: aws.String(v.(string)),
		}
		if !tagIgnoredIotEvents(t) {
			result = append(result, t)
		}
	}

	return result
}

// tagsToMap turns the list of tags into a map.
func tagsToMapIotEvents(ts []*iotevents.Tag) map[string]string {
	result := make(map[string]string)
	for _, t := range ts {
		if !tagIgnoredIotEvents(t) {
			result[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
		}
	}

	return result
}

// compare a tag against a list of strings and checks if it should
// be ignored or not
func tagIgnoredIotEvents(t *iotevents.Tag) bool {
	filter := []string{"^aws:"}
	for _, v := range filter {
		log.Printf("[DEBUG] Matching %v with %v\n", v, *t.Key)
		if r, _ := regexp.MatchString(v, *t.Key); r {
			log.Printf("[DEBUG] Found AWS specific tag %s (val: %s), ignoring.\n", *t.Key, *t.Value)
			return true
		}
	}
	return false
}
