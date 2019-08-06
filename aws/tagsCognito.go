package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentity"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
)

// setTags is a helper to set the tags for a resource. It expects the
// tags field to be named "tags"
func setTagsCognito(conn *cognitoidentity.CognitoIdentity, d *schema.ResourceData) error {
	if d.HasChange("tags") {
		oraw, nraw := d.GetChange("tags")
		o := oraw.(map[string]interface{})
		n := nraw.(map[string]interface{})
		create, remove := diffTagsCognito(tagsFromMapCognito(o), tagsFromMapCognito(n))

		// Set tags
		if len(remove) > 0 {
			log.Printf("[DEBUG] Removing tags: %s", remove)
			tagsToRemove := &cognitoidentity.UntagResourceInput{
				ResourceArn: aws.String(d.Get("arn").(string)),
				TagKeys:     aws.StringSlice(remove),
			}

			_, err := conn.UntagResource(tagsToRemove)
			if err != nil {
				return err
			}
		}
		if len(create) > 0 {
			log.Printf("[DEBUG] Creating tags: %s", create)
			tagsToAdd := &cognitoidentity.TagResourceInput{
				ResourceArn: aws.String(d.Get("arn").(string)),
				Tags:        aws.StringMap(create),
			}
			_, err := conn.TagResource(tagsToAdd)
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
func diffTagsCognito(oldTags, newTags map[string]string) (map[string]string, []string) {
	// First, we're creating everything we have
	create := make(map[string]string)
	for k, v := range newTags {
		create[k] = v
	}

	// Build the list of what to remove
	var remove []string
	for k, v := range oldTags {
		old, ok := create[k]
		if !ok || old != v {
			// Delete it!
			remove = append(remove, k)
		} else if ok {
			// already present so remove from new
			delete(create, k)
		}
	}

	return create, remove
}

// tagsFromMap returns the tags for the given map of data.
func tagsFromMapCognito(m map[string]interface{}) map[string]string {
	results := make(map[string]string)
	for k, v := range m {
		results[k] = v.(string)
	}

	return results
}
