package apigateway

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func DataSourceResource() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceResourceRead,
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
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceResourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn()

	restApiId := d.Get("rest_api_id").(string)
	target := d.Get("path").(string)
	params := &apigateway.GetResourcesInput{RestApiId: aws.String(restApiId)}

	var match *apigateway.Resource
	log.Printf("[DEBUG] Reading API Gateway Resources: %s", params)
	err := conn.GetResourcesPagesWithContext(ctx, params, func(page *apigateway.GetResourcesOutput, lastPage bool) bool {
		for _, resource := range page.Items {
			if aws.StringValue(resource.Path) == target {
				match = resource
				return false
			}
		}
		return !lastPage
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing API Gateway Resources: %s", err)
	}

	if match == nil {
		return sdkdiag.AppendErrorf(diags, "no Resources with path %q found for rest api %q", target, restApiId)
	}

	d.SetId(aws.StringValue(match.Id))
	d.Set("path_part", match.PathPart)
	d.Set("parent_id", match.ParentId)

	return diags
}
