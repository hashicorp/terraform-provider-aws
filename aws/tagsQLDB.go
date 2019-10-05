package aws

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/qldb"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// setTags is a helper to set the tags for a resource. It expects the
// tags field to be named "tags"
func setTagsQLDB(conn *qldb.QLDB, d *schema.ResourceData) error {
	if d.HasChange("tags") {
		oraw, nraw := d.GetChange("tags")
		o := oraw.(map[string]interface{})
		n := nraw.(map[string]interface{})

		create := tagsFromMapQLDBCreate(o)
		remove := tagsFromMapQLDBRemove(n)

		// Set tags
		if len(remove) > 0 {
			err := resource.Retry(5*time.Minute, func() *resource.RetryError {
				log.Printf("[DEBUG] Removing tags: %#v from %s", remove, d.Id())
				_, err := conn.UntagResource(&qldb.UntagResourceInput{
					ResourceArn: aws.String(d.Get("arn").(string)),
					TagKeys:     remove,
				})
				if err != nil {
					qldberr, ok := err.(awserr.Error)
					if ok && strings.Contains(qldberr.Code(), ".NotFound") {
						return resource.RetryableError(err) // retry
					}
					return resource.NonRetryableError(err)
				}
				return nil
			})
			if err != nil {
				// Retry without time bounds for QLDB throttling
				if isResourceTimeoutError(err) {
					log.Printf("[DEBUG] Removing tags: %#v from %s", remove, d.Id())
					_, err := conn.UntagResource(&qldb.UntagResourceInput{
						ResourceArn: aws.String(d.Get("arn").(string)),
						TagKeys:     remove,
					})
					if err != nil {
						return err
					}
				} else {
					return err
				}
			}
		}
		if len(create) > 0 {
			err := resource.Retry(5*time.Minute, func() *resource.RetryError {
				log.Printf("[DEBUG] Creating tags: %v for %s", create, d.Id())
				_, err := conn.TagResource(&qldb.TagResourceInput{
					ResourceArn: aws.String(d.Get("arn").(string)),
					Tags:        create,
				})
				if err != nil {
					ec2err, ok := err.(awserr.Error)
					if ok && strings.Contains(ec2err.Code(), ".NotFound") {
						return resource.RetryableError(err) // retry
					}
					return resource.NonRetryableError(err)
				}
				return nil
			})
			if err != nil {
				// Retry without time bounds for EC2 throttling
				if isResourceTimeoutError(err) {
					log.Printf("[DEBUG] Creating tags: %v for %s", create, d.Id())
					_, err := conn.TagResource(&qldb.TagResourceInput{
						ResourceArn: aws.String(d.Get("arn").(string)),
						Tags:        create,
					})
					if err != nil {
						return err
					}
				} else {
					return err
				}
			}
		}
	}

	return nil
}

// tagsFromMapQLDBCreate returns the tags for the given map of data.
func tagsFromMapQLDBCreate(m map[string]interface{}) map[string]*string {
	result := make(map[string]*string)
	for k, v := range m {
		result[k] = aws.String(v.(string))
	}

	return result
}

// tagsFromMapQLDBRemove returns the tags for the given map of data.
func tagsFromMapQLDBRemove(m map[string]interface{}) []*string {
	result := make([]*string, 0, len(m))
	for k := range m {
		result = append(result, aws.String(k))
	}

	return result
}
