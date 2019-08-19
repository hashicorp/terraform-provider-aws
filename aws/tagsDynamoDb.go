package aws

import (
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

// setTags is a helper to set the tags for a resource. It expects the
// tags field to be named "tags" and the ARN field to be named "arn".
func setTagsDynamoDb(conn *dynamodb.DynamoDB, d *schema.ResourceData) error {
	arn := d.Get("arn").(string)
	oraw, nraw := d.GetChange("tags")
	o := oraw.(map[string]interface{})
	n := nraw.(map[string]interface{})
	create, remove := diffTagsDynamoDb(tagsFromMapDynamoDb(o), tagsFromMapDynamoDb(n))

	// Set tags
	if len(remove) > 0 {
		input := &dynamodb.UntagResourceInput{
			ResourceArn: aws.String(arn),
			TagKeys:     remove,
		}
		err := resource.Retry(2*time.Minute, func() *resource.RetryError {
			log.Printf("[DEBUG] Removing tags: %#v from %s", remove, d.Id())
			_, err := conn.UntagResource(input)
			if err != nil {
				if isAWSErr(err, dynamodb.ErrCodeResourceNotFoundException, "") {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})
		if isResourceTimeoutError(err) {
			_, err = conn.UntagResource(input)
		}
		if err != nil {
			return err
		}
	}
	if len(create) > 0 {
		input := &dynamodb.TagResourceInput{
			ResourceArn: aws.String(arn),
			Tags:        create,
		}
		err := resource.Retry(2*time.Minute, func() *resource.RetryError {
			log.Printf("[DEBUG] Creating tags: %s for %s", create, d.Id())
			_, err := conn.TagResource(input)
			if err != nil {
				if isAWSErr(err, dynamodb.ErrCodeResourceNotFoundException, "") {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})
		if isResourceTimeoutError(err) {
			_, err = conn.TagResource(input)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// diffTags takes our tags locally and the ones remotely and returns
// the set of tags that must be created, and the set of tags that must
// be destroyed.
func diffTagsDynamoDb(oldTags, newTags []*dynamodb.Tag) ([]*dynamodb.Tag, []*string) {
	// First, we're creating everything we have
	create := make(map[string]interface{})
	for _, t := range newTags {
		create[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
	}

	// Build the list of what to remove
	var remove []*string
	for _, t := range oldTags {
		old, ok := create[aws.StringValue(t.Key)]
		if !ok || old != aws.StringValue(t.Value) {
			remove = append(remove, t.Key)
		} else if ok {
			// already present so remove from new
			delete(create, aws.StringValue(t.Key))
		}
	}

	return tagsFromMapDynamoDb(create), remove
}

// tagsFromMapDynamoDb returns the tags for the given map of data.
func tagsFromMapDynamoDb(m map[string]interface{}) []*dynamodb.Tag {
	result := make([]*dynamodb.Tag, 0, len(m))
	for k, v := range m {
		t := &dynamodb.Tag{
			Key:   aws.String(k),
			Value: aws.String(v.(string)),
		}
		if !tagIgnoredDynamoDb(t) {
			result = append(result, t)
		}
	}

	return result
}

// tagsToMap turns the list of tags into a map.
func tagsToMapDynamoDb(ts []*dynamodb.Tag) map[string]string {
	result := make(map[string]string)
	for _, t := range ts {
		if !tagIgnoredDynamoDb(t) {
			result[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
		}
	}
	return result
}

// compare a tag against a list of strings and checks if it should
// be ignored or not
func tagIgnoredDynamoDb(t *dynamodb.Tag) bool {
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
