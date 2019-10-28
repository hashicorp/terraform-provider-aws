package aws

import (
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesisanalytics"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// getTags is a helper to get the tags for a resource. It expects the
// tags field to be named "tags" and the ARN field to be named "arn".
func getTagsKinesisAnalytics(conn *kinesisanalytics.KinesisAnalytics, d *schema.ResourceData) error {
	resp, err := conn.ListTagsForResource(&kinesisanalytics.ListTagsForResourceInput{
		ResourceARN: aws.String(d.Get("arn").(string)),
	})
	if err != nil {
		return err
	}

	if err := d.Set("tags", tagsToMapKinesisAnalytics(resp.Tags)); err != nil {
		return err
	}

	return nil
}

// setTags is a helper to set the tags for a resource. It expects the
// tags field to be named "tags" and the ARN field to be named "arn".
func setTagsKinesisAnalytics(conn *kinesisanalytics.KinesisAnalytics, d *schema.ResourceData) error {
	if d.HasChange("tags") {
		oraw, nraw := d.GetChange("tags")
		o := oraw.(map[string]interface{})
		n := nraw.(map[string]interface{})
		create, remove := diffTagsKinesisAnalytics(tagsFromMapKinesisAnalytics(o), tagsFromMapKinesisAnalytics(n))

		// Set tags
		if len(remove) > 0 {
			log.Printf("[DEBUG] Removing tags: %#v", remove)
			k := make([]*string, len(remove))
			for i, t := range remove {
				k[i] = t.Key
			}

			_, err := conn.UntagResource(&kinesisanalytics.UntagResourceInput{
				ResourceARN: aws.String(d.Get("arn").(string)),
				TagKeys:     k,
			})
			if err != nil {
				return err
			}
		}
		if len(create) > 0 {
			log.Printf("[DEBUG] Creating tags: %#v", create)
			_, err := conn.TagResource(&kinesisanalytics.TagResourceInput{
				ResourceARN: aws.String(d.Get("arn").(string)),
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
func diffTagsKinesisAnalytics(oldTags, newTags []*kinesisanalytics.Tag) ([]*kinesisanalytics.Tag, []*kinesisanalytics.Tag) {
	// First, we're creating everything we have
	create := make(map[string]interface{})
	for _, t := range newTags {
		create[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
	}

	// Build the list of what to remove
	var remove []*kinesisanalytics.Tag
	for _, t := range oldTags {
		old, ok := create[aws.StringValue(t.Key)]
		if !ok || old != aws.StringValue(t.Value) {
			remove = append(remove, t)
		} else if ok {
			// already present so remove from new
			delete(create, aws.StringValue(t.Key))
		}
	}

	return tagsFromMapKinesisAnalytics(create), remove
}

// tagsFromMap returns the tags for the given map of data.
func tagsFromMapKinesisAnalytics(m map[string]interface{}) []*kinesisanalytics.Tag {
	result := make([]*kinesisanalytics.Tag, 0, len(m))
	for k, v := range m {
		t := &kinesisanalytics.Tag{
			Key:   aws.String(k),
			Value: aws.String(v.(string)),
		}
		if !tagIgnoredKinesisAnalytics(t) {
			result = append(result, t)
		}
	}

	return result
}

// tagsToMap turns the list of tags into a map.
func tagsToMapKinesisAnalytics(ts []*kinesisanalytics.Tag) map[string]string {
	result := make(map[string]string)
	for _, t := range ts {
		if !tagIgnoredKinesisAnalytics(t) {
			result[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
		}
	}

	return result
}

// compare a tag against a list of strings and checks if it should
// be ignored or not
func tagIgnoredKinesisAnalytics(t *kinesisanalytics.Tag) bool {
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
