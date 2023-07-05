// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_api_gateway_vpc_link")
func DataSourceVPCLink() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceVPCLinkRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status_message": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"target_arns": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceVPCLinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	params := &apigateway.GetVpcLinksInput{}

	target := d.Get("name")
	var matchedVpcLinks []*apigateway.UpdateVpcLinkOutput
	log.Printf("[DEBUG] Reading API Gateway VPC links: %s", params)
	err := conn.GetVpcLinksPagesWithContext(ctx, params, func(page *apigateway.GetVpcLinksOutput, lastPage bool) bool {
		for _, api := range page.Items {
			if aws.StringValue(api.Name) == target {
				matchedVpcLinks = append(matchedVpcLinks, api)
			}
		}
		return !lastPage
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing API Gateway VPC links: %s", err)
	}

	if len(matchedVpcLinks) == 0 {
		return sdkdiag.AppendErrorf(diags, "no API Gateway VPC link with name %q found in this region", target)
	}
	if len(matchedVpcLinks) > 1 {
		return sdkdiag.AppendErrorf(diags, "multiple API Gateway VPC links with name %q found in this region", target)
	}

	match := matchedVpcLinks[0]

	d.SetId(aws.StringValue(match.Id))
	d.Set("name", match.Name)
	d.Set("status", match.Status)
	d.Set("status_message", match.StatusMessage)
	d.Set("description", match.Description)
	d.Set("target_arns", flex.FlattenStringList(match.TargetArns))

	if err := d.Set("tags", KeyValueTags(ctx, match.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
