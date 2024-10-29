// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_nat_gateways", name="NAT Gateways")
func dataSourceNATGateways() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceNATGatewaysRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrFilter: customFiltersSchema(),
			names.AttrIDs: {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceNATGatewaysRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeNatGatewaysInput{}

	if v, ok := d.GetOk(names.AttrVPCID); ok {
		input.Filter = append(input.Filter, newAttributeFilterList(
			map[string]string{
				"vpc-id": v.(string),
			},
		)...)
	}

	if tags, ok := d.GetOk(names.AttrTags); ok {
		input.Filter = append(input.Filter, newTagFilterList(
			Tags(tftags.New(ctx, tags.(map[string]interface{}))),
		)...)
	}

	input.Filter = append(input.Filter, newCustomFilterList(
		d.Get(names.AttrFilter).(*schema.Set),
	)...)

	if len(input.Filter) == 0 {
		input.Filter = nil
	}

	output, err := findNATGateways(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 NAT Gateways: %s", err)
	}

	var natGatewayIDs []string

	for _, v := range output {
		natGatewayIDs = append(natGatewayIDs, aws.ToString(v.NatGatewayId))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set(names.AttrIDs, natGatewayIDs)

	return diags
}
