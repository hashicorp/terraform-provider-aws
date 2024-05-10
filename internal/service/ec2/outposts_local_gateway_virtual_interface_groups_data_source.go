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
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ec2_local_gateway_virtual_interface_groups")
func DataSourceLocalGatewayVirtualInterfaceGroups() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceLocalGatewayVirtualInterfaceGroupsRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"filter": customFiltersSchema(),
			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"local_gateway_virtual_interface_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceLocalGatewayVirtualInterfaceGroupsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	input := &ec2.DescribeLocalGatewayVirtualInterfaceGroupsInput{}

	input.Filters = append(input.Filters, newTagFilterList(
		Tags(tftags.New(ctx, d.Get(names.AttrTags).(map[string]interface{}))),
	)...)

	input.Filters = append(input.Filters, newCustomFilterList(
		d.Get("filter").(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	output, err := FindLocalGatewayVirtualInterfaceGroups(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Local Gateway Virtual Interface Groups: %s", err)
	}

	var groupIDs, interfaceIDs []string

	for _, v := range output {
		groupIDs = append(groupIDs, aws.StringValue(v.LocalGatewayVirtualInterfaceGroupId))
		interfaceIDs = append(interfaceIDs, aws.StringValueSlice(v.LocalGatewayVirtualInterfaceIds)...)
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("ids", groupIDs)
	d.Set("local_gateway_virtual_interface_ids", interfaceIDs)

	return diags
}
