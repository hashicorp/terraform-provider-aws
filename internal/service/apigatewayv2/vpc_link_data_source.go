// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_apigatewayv2_vpc_link", name="VPC Link")
// @Tags
func dataSourceVPCLink() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceVPCLinkRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrSecurityGroupIDs: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrSubnetIDs: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"vpc_link_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceVPCLinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	vpcLinkID := d.Get("vpc_link_id").(string)
	output, err := findVPCLinkByID(ctx, conn, vpcLinkID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway v2 VPC Link (%s): %s", vpcLinkID, err)
	}

	d.SetId(vpcLinkID)
	d.Set(names.AttrARN, vpcLinkARN(meta.(*conns.AWSClient), d.Id()))
	d.Set(names.AttrName, output.Name)
	d.Set(names.AttrSecurityGroupIDs, output.SecurityGroupIds)
	d.Set(names.AttrSubnetIDs, output.SubnetIds)

	setTagsOut(ctx, output.Tags)

	return diags
}
