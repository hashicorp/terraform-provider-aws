package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsApiGatewayStage() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsApiGatewayStageRead,
		Schema: map[string]*schema.Schema{
			"stage_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"deployment_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsApiGatewayStageRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway
	restApiID := d.Get("rest_api_id").(string)
	stageName := d.Get("stage_name").(string)
	params := &apigateway.GetStageInput{
		RestApiId: aws.String(restApiID),
		StageName: aws.String(stageName),
	}
	log.Printf("[DEBUG] Reading API Gateway Stage: %s", params)
	stage, err := conn.GetStage(params)
	if err != nil {
		return fmt.Errorf("error describing API Gateway Stage: %s", err)
	}

	d.SetId(fmt.Sprintf("ags-%s-%s", restApiID, stageName))
	d.Set("deployment_id", stage.DeploymentId)
	return nil
}
