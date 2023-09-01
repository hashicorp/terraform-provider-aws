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
)

// @SDKDataSource("aws_ec2_local_gateway_virtual_interface_group")
func DataSourceLocalGatewayVirtualInterfaceGroup() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceLocalGatewayVirtualInterfaceGroupRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"filter": CustomFiltersSchema(),
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"local_gateway_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"local_gateway_virtual_interface_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceLocalGatewayVirtualInterfaceGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeLocalGatewayVirtualInterfaceGroupsInput{}

	if v, ok := d.GetOk("id"); ok {
		input.LocalGatewayVirtualInterfaceGroupIds = []*string{aws.String(v.(string))}
	}

	input.Filters = BuildAttributeFilterList(
		map[string]string{
			"local-gateway-id": d.Get("local_gateway_id").(string),
		},
	)

	input.Filters = append(input.Filters, BuildTagFilterList(
		Tags(tftags.New(ctx, d.Get("tags").(map[string]interface{}))),
	)...)

	input.Filters = append(input.Filters, BuildCustomFilterList(
		d.Get("filter").(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filters = nil
	}

	output, err := conn.DescribeLocalGatewayVirtualInterfaceGroupsWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing EC2 Local Gateway Virtual Interface Groups: %s", err)
	}

	if output == nil || len(output.LocalGatewayVirtualInterfaceGroups) == 0 {
		return sdkdiag.AppendErrorf(diags, "no matching EC2 Local Gateway Virtual Interface Group found")
	}

	if len(output.LocalGatewayVirtualInterfaceGroups) > 1 {
		return sdkdiag.AppendErrorf(diags, "multiple EC2 Local Gateway Virtual Interface Groups matched; use additional constraints to reduce matches to a single EC2 Local Gateway Virtual Interface Group")
	}

	localGatewayVirtualInterfaceGroup := output.LocalGatewayVirtualInterfaceGroups[0]

	d.SetId(aws.StringValue(localGatewayVirtualInterfaceGroup.LocalGatewayVirtualInterfaceGroupId))
	d.Set("local_gateway_id", localGatewayVirtualInterfaceGroup.LocalGatewayId)
	d.Set("local_gateway_virtual_interface_group_id", localGatewayVirtualInterfaceGroup.LocalGatewayVirtualInterfaceGroupId)

	if err := d.Set("local_gateway_virtual_interface_ids", aws.StringValueSlice(localGatewayVirtualInterfaceGroup.LocalGatewayVirtualInterfaceIds)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting local_gateway_virtual_interface_ids: %s", err)
	}

	if err := d.Set("tags", KeyValueTags(ctx, localGatewayVirtualInterfaceGroup.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
