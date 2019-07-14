package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/terraform/helper/schema"
)

// diffTags takes our tags locally and the ones remotely and returns
// the set of tags that must be created, and the set of tags that must
// be destroyed.
func diffTagsLightsail(oldTags, newTags []*lightsail.Tag) ([]*lightsail.Tag, []*lightsail.Tag) {
	// First, we're creating everything we have
	create := make(map[string]interface{})
	for _, t := range newTags {
		create[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
	}

	// Build the list of what to remove
	var remove []*lightsail.Tag
	for _, t := range oldTags {
		old, ok := create[aws.StringValue(t.Key)]
		if !ok || old != aws.StringValue(t.Value) {
			// Delete it!
			remove = append(remove, t)
		} else if ok {
			delete(create, aws.StringValue(t.Key))
		}
	}

	return tagsFromMapLightsail(create), remove
}

// setTags is a helper to set the tags for a resource. It expects the
// tags field to be named "tags"
func setTagsLightsail(conn *lightsail.Lightsail, d *schema.ResourceData) error {
	if d.HasChange("tags") {
		oraw, nraw := d.GetChange("tags")
		o := oraw.(map[string]interface{})
		n := nraw.(map[string]interface{})
		create, remove := diffTagsLightsail(tagsFromMapLightsail(o), tagsFromMapLightsail(n))

		// Set tags
		if len(remove) > 0 {
			log.Printf("[DEBUG] Removing tags: %#v", remove)
			k := make([]*string, len(remove))
			for i, t := range remove {
				k[i] = t.Key
			}

			_, err := conn.UntagResource(&lightsail.UntagResourceInput{
				ResourceName: aws.String(d.Get("name").(string)),
				TagKeys:      k,
			})
			if err != nil {
				return err
			}
		}
		if len(create) > 0 {
			log.Printf("[DEBUG] Creating tags: %#v", create)
			_, err := conn.TagResource(&lightsail.TagResourceInput{
				ResourceName: aws.String(d.Get("name").(string)),
				Tags:         create,
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// tagsFromMap returns the tags for the given map of data.
func tagsFromMapLightsail(m map[string]interface{}) []*lightsail.Tag {
	result := make([]*lightsail.Tag, 0, len(m))
	for k, v := range m {
		result = append(result, &lightsail.Tag{
			Key:   aws.String(k),
			Value: aws.String(v.(string)),
		})
	}

	return result
}

// tagsToMap turns the list of tags into a map.
func tagsToMapLightsail(ts []*lightsail.Tag) map[string]string {
	result := make(map[string]string)
	for _, t := range ts {
		result[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
	}

	return result
}
