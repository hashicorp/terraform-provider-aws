package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform/helper/schema"
)

func updateTagsAPIGatewayStage(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway

	stageArn := arnString(
		meta.(*AWSClient).partition,
		meta.(*AWSClient).region,
		"apigateway",
		"",
		fmt.Sprintf("/restapis/%s/stages/%s", d.Get("rest_api_id").(string), d.Get("stage_name").(string)),
	)

	return updateTagsAPIGateway(conn, d, stageArn)
}

func updateTagsAPIGateway(conn *apigateway.APIGateway, d *schema.ResourceData, arn string) error {
	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		oldT := o.(map[string]interface{})
		newT := n.(map[string]interface{})

		create, remove := diffTagsGeneric(oldT, newT)

		if len(remove) > 0 {
			log.Printf("[DEBUG] Removing APIGateway tags: %#v", remove)

			removeKeys := make([]*string, 0, len(remove))
			for k := range remove {
				removeKeys = append(removeKeys, aws.String(k))
			}

			_, err := conn.UntagResource(&apigateway.UntagResourceInput{
				ResourceArn: aws.String(arn),
				TagKeys:     removeKeys,
			})

			if err != nil {
				return err
			}
		}

		if len(create) > 0 {
			log.Printf("[DEBUG] Creating APIGateway tags: %#v", create)

			_, err := conn.TagResource(&apigateway.TagResourceInput{
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
