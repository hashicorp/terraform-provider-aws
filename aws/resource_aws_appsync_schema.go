package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/resource"
	"fmt"
)

func resourceAwsAppsyncSchema() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAppsyncSchemaPut,
		Read:   schema.Noop,
		Update: resourceAwsAppsyncSchemaPut,
		Delete: schema.Noop,

		Schema: map[string]*schema.Schema{
			"api_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"definition": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceAwsAppsyncSchemaPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appsyncconn

	apiId := d.Get("api_id").(string)
	d.SetId(apiId)

	if d.HasChange("definition") {
		input := &appsync.StartSchemaCreationInput{
			ApiId:      aws.String(apiId),
			Definition: ([]byte)(d.Get("definition").(string)),
		}
		if _, err := conn.StartSchemaCreation(input); err != nil {
			return err
		}

		activeSchemaConfig := &resource.StateChangeConf{
			Pending: []string{ "PROCESSING" },
			Target: []string{ "ACTIVE" },
			Refresh: func() (interface{}, string, error) {
				conn := meta.(*AWSClient).appsyncconn
				input := &appsync.GetSchemaCreationStatusInput{
					ApiId: aws.String(apiId),
				}
				result, err := conn.GetSchemaCreationStatus(input)

				if err != nil {
					return 0, "", err
				}
				return result, *result.Status, nil
			},
		}

		if _, err := activeSchemaConfig.WaitForState(); err != nil {
			return fmt.Errorf("Error waiting for schema creation status on AppSync API %s: %s", apiId, err)
		}
	}

	return nil
}
