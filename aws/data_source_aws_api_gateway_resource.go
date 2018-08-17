package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsApiGatewayResource() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsApiGatewayResourceRead,
		Schema: map[string]*schema.Schema{
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"path": {
				Type:     schema.TypeString,
				Required: true,
			},
			"path_part": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"parent_id": {
				Type: schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsApiGatewayResourceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway

	restApiId := d.Get("rest_api_id").(string)
	target := d.Get("path").(string)
	params := &apigateway.GetResourcesInput{RestApiId: aws.String(restApiId)}

	var matches []*apigateway.Resource
	log.Printf("[DEBUG] Reading API Gateway Resources: %s", params)
	err := conn.GetResourcesPages(params,func(page *apigateway.GetResourcesOutput, lastPage bool) bool {
		for _, resource := range page.Items {
			if aws.StringValue(resource.Path) == target {
				matches = append(matches, resource)
			}
		}
		return !lastPage
	})
	if err != nil {
		return fmt.Errorf("error describing API Gateway Resources: %s", err)
	}

	if len(matches) == 0 {
		return fmt.Errorf("no Resources with path %q found for rest api %q", target, restApiId)
	}
	if len(matches) > 1 {
		return fmt.Errorf("multiple Resources with path %q found for rest api %q", target, restApiId)
	}

	match := matches[0]
	d.SetId(*match.Id)

	return nil
}
