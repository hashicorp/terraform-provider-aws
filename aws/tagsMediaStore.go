package aws

import (
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediastore"
	"github.com/hashicorp/terraform/helper/schema"
)

// setTags is a helper to set the tags for a resource. It expects the
// tags field to be named "tags"
func setTagsMediaStore(conn *mediastore.MediaStore, d *schema.ResourceData, arn string) error {
	if d.HasChange("tags") {
		oraw, nraw := d.GetChange("tags")
		o := oraw.(map[string]interface{})
		n := nraw.(map[string]interface{})
		create, remove := diffTagsMediaStore(tagsFromMapMediaStore(o), tagsFromMapMediaStore(n))

		// Set tags
		if len(remove) > 0 {
			log.Printf("[DEBUG] Removing tags: %s", remove)
			k := make([]*string, len(remove))
			for i, t := range remove {
				k[i] = t.Key
			}

			_, err := conn.UntagResource(&mediastore.UntagResourceInput{
				Resource: aws.String(arn),
				TagKeys:  k,
			})
			if err != nil {
				return err
			}
		}
		if len(create) > 0 {
			log.Printf("[DEBUG] Creating tags: %s", create)
			_, err := conn.TagResource(&mediastore.TagResourceInput{
				Resource: aws.String(arn),
				Tags:     create,
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
func diffTagsMediaStore(oldTags, newTags []*mediastore.Tag) ([]*mediastore.Tag, []*mediastore.Tag) {
	// First, we're creating everything we have
	create := make(map[string]interface{})
	for _, t := range newTags {
		create[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
	}

	// Build the list of what to remove
	var remove []*mediastore.Tag
	for _, t := range oldTags {
		old, ok := create[aws.StringValue(t.Key)]
		if !ok || old != aws.StringValue(t.Value) {
			// Delete it!
			remove = append(remove, t)
		} else if ok {
			delete(create, aws.StringValue(t.Key))
		}
	}

	return tagsFromMapMediaStore(create), remove
}

// tagsFromMap returns the tags for the given map of data.
func tagsFromMapMediaStore(m map[string]interface{}) []*mediastore.Tag {
	result := make([]*mediastore.Tag, 0, len(m))
	for k, v := range m {
		t := &mediastore.Tag{
			Key:   aws.String(k),
			Value: aws.String(v.(string)),
		}
		if !tagIgnoredMediaStore(t) {
			result = append(result, t)
		}
	}

	return result
}

// tagsToMap turns the list of tags into a map.
func tagsToMapMediaStore(ts []*mediastore.Tag) map[string]string {
	result := make(map[string]string)
	for _, t := range ts {
		if !tagIgnoredMediaStore(t) {
			result[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
		}
	}

	return result
}

// compare a tag against a list of strings and checks if it should
// be ignored or not
func tagIgnoredMediaStore(t *mediastore.Tag) bool {
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

func saveTagsMediaStore(conn *mediastore.MediaStore, d *schema.ResourceData, arn string) error {
	resp, err := conn.ListTagsForResource(&mediastore.ListTagsForResourceInput{
		Resource: aws.String(arn),
	})

	if err != nil {
		return err
	}

	var dt []*mediastore.Tag
	if len(resp.Tags) > 0 {
		dt = resp.Tags
	}

	return d.Set("tags", tagsToMapMediaStore(dt))
}
