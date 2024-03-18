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
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	params := &apigateway.GetVpcLinksInput{}

	target := d.Get("name")
	var matchedVpcLinks []awstypes.VpcLink
	log.Printf("[DEBUG] Reading API Gateway VPC links: %+v", params)

	pages := apigateway.NewGetVpcLinksPaginator(conn, params)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "describing API Gateway VPC links: %s", err)
		}

		for _, api := range page.Items {
			if aws.ToString(api.Name) == target {
				matchedVpcLinks = append(matchedVpcLinks, api)
			}
		}
	}

	match, err := tfresource.AssertSingleValueResult(matchedVpcLinks)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Exactly one API Gateway VPC link with name %q not found in this region", target)
	}

	d.SetId(aws.ToString(match.Id))
	d.Set("name", match.Name)
	d.Set("status", match.Status)
	d.Set("status_message", match.StatusMessage)
	d.Set("description", match.Description)
	d.Set("target_arns", flex.FlattenStringValueList(match.TargetArns))

	if err := d.Set("tags", KeyValueTags(ctx, match.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
