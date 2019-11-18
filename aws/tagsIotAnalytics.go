package aws

import (
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iotanalytics"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

// getTags is a helper to get the tags for a resource. It expects the
// tags field to be named "tags" and the ARN field to be named "arn".
func getTagsIotAnalytics(conn *iotanalytics.IoTAnalytics, d *schema.ResourceData) error {
	req := &iotanalytics.ListTagsForResourceInput{
		ResourceArn: aws.String(d.Get("arn").(string)),
	}
	resp, err := conn.ListTagsForResource(req)
	if err != nil {
		return err
	}

	if err := d.Set("tags", keyvaluetags.IotanalyticsKeyValueTags(resp.Tags).IgnoreAws().Map()); err != nil {
		return err
	}

	return nil
}

// setTags is a helper to set the tags for a resource. It expects the
// tags field to be named "tags" and the ARN field to be named "arn".
func setTagsIotAnalytics(conn *iotanalytics.IoTAnalytics, d *schema.ResourceData) error {
	if d.HasChange("tags") {
		oraw, nraw := d.GetChange("tags")
		o := oraw.(map[string]interface{})
		n := nraw.(map[string]interface{})
		create, remove := diffTagsIotAnalytics(tagsFromMapIotAnalytics(o), tagsFromMapIotAnalytics(n))

		// Set tags
		if len(remove) > 0 {
			log.Printf("[DEBUG] Removing tags: %#v", remove)
			k := make([]*string, len(remove))
			for i, t := range remove {
				k[i] = t.Key
			}

			_, err := conn.UntagResource(&iotanalytics.UntagResourceInput{
				ResourceArn: aws.String(d.Get("arn").(string)),
				TagKeys:     k,
			})
			if err != nil {
				return err
			}
		}
		if len(create) > 0 {
			log.Printf("[DEBUG] Creating tags: %#v", create)
			_, err := conn.TagResource(&iotanalytics.TagResourceInput{
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
func diffTagsIotAnalytics(oldTags, newTags []*iotanalytics.Tag) ([]*iotanalytics.Tag, []*iotanalytics.Tag) {
	// First, we're creating everything we have
	create := make(map[string]interface{})
	for _, t := range newTags {
		create[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
	}

	// Build the list of what to remove
	var remove []*iotanalytics.Tag
	for _, old := range oldTags {
		new, ok := create[aws.StringValue(old.Key)]
		if !ok || new != aws.StringValue(old.Value) {
			remove = append(remove, old)
		} else {
			// already present so remove from new
			delete(create, aws.StringValue(old.Key))
		}
	}

	return tagsFromMapIotAnalytics(create), remove
}

// tagsFromMap returns the tags for the given map of data.
func tagsFromMapIotAnalytics(m map[string]interface{}) []*iotanalytics.Tag {
	result := make([]*iotanalytics.Tag, 0, len(m))
	for k, v := range m {
		t := &iotanalytics.Tag{
			Key:   aws.String(k),
			Value: aws.String(v.(string)),
		}
		if !tagIgnoredIotAnalytics(t) {
			result = append(result, t)
		}
	}

	return result
}

// compare a tag against a list of strings and checks if it should
// be ignored or not
func tagIgnoredIotAnalytics(t *iotanalytics.Tag) bool {
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
