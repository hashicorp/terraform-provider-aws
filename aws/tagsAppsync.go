package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/terraform/helper/schema"
)

func setTagsAppsync(conn *appsync.AppSync, d *schema.ResourceData, arn string) error {
	if d.HasChange("tags") {
		oraw, nraw := d.GetChange("tags")
		o := oraw.(map[string]interface{})
		n := nraw.(map[string]interface{})
		create, remove := diffTagsGeneric(o, n)

		// remove old tags
		if len(remove) > 0 {
			log.Printf("[DEBUG] Removing tags: %#v", remove)
			keys := make([]*string, 0, len(remove))
			for k := range remove {
				keys = append(keys, aws.String(k))
			}

			_, err := conn.UntagResource(&appsync.UntagResourceInput{
				ResourceArn: aws.String(arn),
				TagKeys:     keys,
			})
			if err != nil {
				return err
			}
		}

		// create new tags
		if len(create) > 0 {
			log.Printf("[DEBUG] Creating tags: %#v", create)

			_, err := conn.TagResource(&appsync.TagResourceInput{
				ResourceArn: aws.String(arn),
				Tags:        create,
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}
