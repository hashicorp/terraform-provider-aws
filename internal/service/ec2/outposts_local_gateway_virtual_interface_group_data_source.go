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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ec2_local_gateway_virtual_interface_group", name="Local Gateway Virtual Interface Group")
// @Tags
// @Testing(tagsTest=false)
func dataSourceLocalGatewayVirtualInterfaceGroup() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceLocalGatewayVirtualInterfaceGroupRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrFilter: customFiltersSchema(),
			names.AttrID: {
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
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceLocalGatewayVirtualInterfaceGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeLocalGatewayVirtualInterfaceGroupsInput{}

	if v, ok := d.GetOk(names.AttrID); ok {
		input.LocalGatewayVirtualInterfaceGroupIds = []string{v.(string)}
	}

	input.Filters = newAttributeFilterList(
		map[string]string{
			"local-gateway-id": d.Get("local_gateway_id").(string),
		},
	)

	input.Filters = append(input.Filters, newTagFilterList(
		Tags(tftags.New(ctx, d.Get(names.AttrTags).(map[string]interface{}))),
	)...)

	input.Filters = append(input.Filters, newCustomFilterList(
		d.Get(names.AttrFilter).(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filters = nil
	}

	localGatewayVirtualInterfaceGroup, err := findLocalGatewayVirtualInterfaceGroup(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("EC2 Local Gateway Virtual Interface Group", err))
	}

	d.SetId(aws.ToString(localGatewayVirtualInterfaceGroup.LocalGatewayVirtualInterfaceGroupId))
	d.Set("local_gateway_id", localGatewayVirtualInterfaceGroup.LocalGatewayId)
	d.Set("local_gateway_virtual_interface_group_id", localGatewayVirtualInterfaceGroup.LocalGatewayVirtualInterfaceGroupId)
	d.Set("local_gateway_virtual_interface_ids", localGatewayVirtualInterfaceGroup.LocalGatewayVirtualInterfaceIds)

	setTagsOut(ctx, localGatewayVirtualInterfaceGroup.Tags)

	return diags
}
