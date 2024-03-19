// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKDataSource("aws_api_gateway_resource")
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
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	restApiId := d.Get("rest_api_id").(string)
	target := d.Get("path").(string)
	params := &apigateway.GetResourcesInput{RestApiId: aws.String(restApiId)}

	var resources []awstypes.Resource
	log.Printf("[DEBUG] Reading API Gateway Resources: %+v", params)

	pages := apigateway.NewGetResourcesPaginator(conn, params)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "describing API Gateway Resources: %s", err)
		}

		for _, resource := range page.Items {
			if aws.ToString(resource.Path) == target {
				resources = append(resources, resource)
			}
		}
	}

	match, err := tfresource.AssertSingleValueResult(resources)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "no Resources with path %q found for rest api %q", target, restApiId)
	}

	d.SetId(aws.ToString(match.Id))
	d.Set("path_part", match.PathPart)
	d.Set("parent_id", match.ParentId)

	return diags
}
