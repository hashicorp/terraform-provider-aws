// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_nat_gateways")
func DataSourceNATGateways() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceNATGatewaysRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"filter": DataSourceFiltersSchema(),
			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": tftags.TagsSchemaComputed(),
			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceNATGatewaysRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	input := &ec2.DescribeNatGatewaysInput{}

	if v, ok := d.GetOk("vpc_id"); ok {
		input.Filter = append(input.Filter, BuildAttributeFilterList(
			map[string]string{
				"vpc-id": v.(string),
			},
		)...)
	}

	if tags, ok := d.GetOk("tags"); ok {
		input.Filter = append(input.Filter, BuildTagFilterList(
			Tags(tftags.New(ctx, tags.(map[string]interface{}))),
		)...)
	}

	input.Filter = append(input.Filter, BuildFiltersDataSource(
		d.Get("filter").(*schema.Set),
	)...)

	if len(input.Filter) == 0 {
		input.Filter = nil
	}

	output, err := FindNATGateways(ctx, conn, input)

	if err != nil {
		return diag.Errorf("reading EC2 NAT Gateways: %s", err)
	}

	var natGatewayIDs []string

	for _, v := range output {
		natGatewayIDs = append(natGatewayIDs, aws.StringValue(v.NatGatewayId))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("ids", natGatewayIDs)

	return nil
}
