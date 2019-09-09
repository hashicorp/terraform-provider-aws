package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform/helper/schema"
)

// return a slice of s3 tags associated with the given s3 bucket. Essentially
// s3.GetBucketTagging, except returns an empty slice instead of an error when
// there are no tags.
func getTagSetS3Bucket(conn *s3.S3, bucket string) ([]*s3.Tag, error) {
	resp, err := conn.GetBucketTagging(&s3.GetBucketTaggingInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		if isAWSErr(err, "NoSuchTagSet", "") {
			return nil, nil
		}
		return nil, err
	}

	return resp.TagSet, nil
}

// setTags is a helper to set the tags for a resource. It expects the
// tags field to be named "tags"
func setTagsS3Bucket(conn *s3.S3, d *schema.ResourceData) error {
	if d.HasChange("tags") {
		oraw, nraw := d.GetChange("tags")
		o := oraw.(map[string]interface{})
		n := nraw.(map[string]interface{})

		// Get any existing system tags.
		var sysTagSet []*s3.Tag
		oTagSet, err := getTagSetS3Bucket(conn, d.Get("bucket").(string))
		if err != nil {
			return fmt.Errorf("error getting S3 bucket tags: %s", err)
		}
		for _, tag := range oTagSet {
			if tagIgnoredS3(tag) {
				sysTagSet = append(sysTagSet, tag)
			}
		}

		if len(n)+len(sysTagSet) > 0 {
			// The bucket's tag set must include any system tags that Terraform ignores.
			nTagSet := append(tagsFromMapS3(n), sysTagSet...)

			req := &s3.PutBucketTaggingInput{
				Bucket: aws.String(d.Get("bucket").(string)),
				Tagging: &s3.Tagging{
					TagSet: nTagSet,
				},
			}
			if _, err := RetryOnAwsCodes([]string{"NoSuchBucket", "OperationAborted"}, func() (interface{}, error) {
				return conn.PutBucketTagging(req)
			}); err != nil {
				return fmt.Errorf("error setting S3 bucket tags: %s", err)
			}
		} else if len(o) > 0 && len(sysTagSet) == 0 {
			req := &s3.DeleteBucketTaggingInput{
				Bucket: aws.String(d.Get("bucket").(string)),
			}
			if _, err := RetryOnAwsCodes([]string{"NoSuchBucket", "OperationAborted"}, func() (interface{}, error) {
				return conn.DeleteBucketTagging(req)
			}); err != nil {
				return fmt.Errorf("error deleting S3 bucket tags: %s", err)
			}
		}
	}

	return nil
}

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
