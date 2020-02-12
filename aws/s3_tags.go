package aws

import (
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func getTagsS3Object(conn *s3.S3, d *schema.ResourceData) error {
	resp, err := retryOnAwsCode(s3.ErrCodeNoSuchKey, func() (interface{}, error) {
		return conn.GetObjectTagging(&s3.GetObjectTaggingInput{
			Bucket: aws.String(d.Get("bucket").(string)),
			Key:    aws.String(d.Get("key").(string)),
		})
	})
	if err != nil {
		return err
	}

	if err := d.Set("tags", tagsToMapS3(resp.(*s3.GetObjectTaggingOutput).TagSet)); err != nil {
		return err
	}

	return nil
}

func setTagsS3Object(conn *s3.S3, d *schema.ResourceData) error {
	if d.HasChange("tags") {
		oraw, nraw := d.GetChange("tags")
		o := oraw.(map[string]interface{})
		n := nraw.(map[string]interface{})

		// Set tags
		if len(o) > 0 {
			if _, err := conn.DeleteObjectTagging(&s3.DeleteObjectTaggingInput{
				Bucket: aws.String(d.Get("bucket").(string)),
				Key:    aws.String(d.Get("key").(string)),
			}); err != nil {
				return err
			}
		}
		if len(n) > 0 {
			if _, err := conn.PutObjectTagging(&s3.PutObjectTaggingInput{
				Bucket: aws.String(d.Get("bucket").(string)),
				Key:    aws.String(d.Get("key").(string)),
				Tagging: &s3.Tagging{
					TagSet: tagsFromMapS3(n),
				},
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

// tagsFromMap returns the tags for the given map of data.
func tagsFromMapS3(m map[string]interface{}) []*s3.Tag {
	result := make([]*s3.Tag, 0, len(m))
	for k, v := range m {
		t := &s3.Tag{
			Key:   aws.String(k),
			Value: aws.String(v.(string)),
		}
		if !tagIgnoredS3(t) {
			result = append(result, t)
		}
	}

	return result
}

// tagsToMap turns the list of tags into a map.
func tagsToMapS3(ts []*s3.Tag) map[string]string {
	result := make(map[string]string)
	for _, t := range ts {
		if !tagIgnoredS3(t) {
			result[*t.Key] = *t.Value
		}
	}

	return result
}

// compare a tag against a list of strings and checks if it should
// be ignored or not
func tagIgnoredS3(t *s3.Tag) bool {
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
