package aws

import (
	"errors"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/terraform/helper/schema"
	"time"
)

func resourceAwsAppsyncSchema() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAppsyncSchemaPut,
		Read:   resourceAwsAppsyncSchemaNil,
		Update: resourceAwsAppsyncSchemaPut,
		Delete: resourceAwsAppsyncSchemaNil,

		Schema: map[string]*schema.Schema{
			"api_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
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

		if err := waitForSchemaToBeActive(apiId, meta); err != nil {
			log.Printf("[DEBUG] Error waiting for schema to be active: %s", err)
			return err
		}
	}

	return nil
}

func resourceAwsAppsyncSchemaNil(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func waitForSchemaToBeActive(apiId string, meta interface{}) error {
	conn := meta.(*AWSClient).appsyncconn

	input := &appsync.GetSchemaCreationStatusInput{
		ApiId: aws.String(apiId),
	}

	iterations := 0
	for iterations < 24 {
		result, err := conn.GetSchemaCreationStatus(input)

		if err != nil {
			return err
		} else if *result.Status == "SUCCESS" {
			return nil
		} else if *result.Status == "FAILED" {
			return errors.New(*result.Details)
		}

		// Wait for a few seconds
		log.Printf("[DEBUG] Sleeping for 5 seconds for schema to become active")
		time.Sleep(5 * time.Second)
		iterations += 1
	}

	return nil
}
