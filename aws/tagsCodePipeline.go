package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codepipeline"
	"github.com/hashicorp/terraform/helper/schema"
)

// setTags is a helper to set the tags for a resource. It expects the
// tags field to be named "tags"
func setTagsCodePipeline(conn *codepipeline.CodePipeline, d *schema.ResourceData) error {
	if d.HasChange("tags") {
		oraw, nraw := d.GetChange("tags")
		o := oraw.(map[string]interface{})
		n := nraw.(map[string]interface{})
		create, remove := diffTagsCodePipeline(tagsFromMapCodePipeline(o), tagsFromMapCodePipeline(n))

		// Set tags
		if len(remove) > 0 {
			log.Printf("[DEBUG] Removing tags: %#v", remove)
			k := make([]*string, len(remove))
			for i, t := range remove {
				k[i] = t.Key
			}

			_, err := conn.UntagResource(&codepipeline.UntagResourceInput{
				ResourceArn: aws.String(d.Get("arn").(string)),
				TagKeys:     k,
			})
			if err != nil {
				return err
			}
		}
		if len(create) > 0 {
			log.Printf("[DEBUG] Creating tags: %#v", create)
			_, err := conn.TagResource(&codepipeline.TagResourceInput{
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
func diffTagsCodePipeline(oldTags, newTags []*codepipeline.Tag) ([]*codepipeline.Tag, []*codepipeline.Tag) {
	// First, we're creating everything we have
	create := make(map[string]interface{})
	for _, t := range newTags {
		create[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
	}

	// Build the list of what to remove
	var remove []*codepipeline.Tag
	for _, t := range oldTags {
		old, ok := create[aws.StringValue(t.Key)]
		if !ok || old != aws.StringValue(t.Value) {
			// Delete it!
			remove = append(remove, t)
		} else if ok {
			// already present so remove from new
			delete(create, aws.StringValue(t.Key))
		}
	}

	return tagsFromMapCodePipeline(create), remove
}

func saveTagsCodePipeline(conn *codepipeline.CodePipeline, d *schema.ResourceData) error {
	resp, err := conn.ListTagsForResource(&codepipeline.ListTagsForResourceInput{
		ResourceArn: aws.String(d.Get("arn").(string)),
	})

	if err != nil {
		return fmt.Errorf("Error retreiving tags for ARN: %s", d.Get("arn").(string))
	}

	var dt []*codepipeline.Tag
	if len(resp.Tags) > 0 {
		dt = resp.Tags
	}

	return d.Set("tags", tagsToMapCodePipeline(dt))
}

// tagsFromMap returns the tags for the given map of data.
func tagsFromMapCodePipeline(m map[string]interface{}) []*codepipeline.Tag {
	result := make([]*codepipeline.Tag, 0, len(m))
	for k, v := range m {
		t := &codepipeline.Tag{
			Key:   aws.String(k),
			Value: aws.String(v.(string)),
		}
		if !tagIgnoredCodePipeline(t) {
			result = append(result, t)
		}
	}

	return result
}

// tagsToMap turns the list of tags into a map.
func tagsToMapCodePipeline(ts []*codepipeline.Tag) map[string]string {
	result := make(map[string]string)
	for _, t := range ts {
		if !tagIgnoredCodePipeline(t) {
			result[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
		}
	}

	return result
}

// compare a tag against a list of strings and checks if it should
// be ignored or not
func tagIgnoredCodePipeline(t *codepipeline.Tag) bool {
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
