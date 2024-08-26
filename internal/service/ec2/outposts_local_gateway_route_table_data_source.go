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

// @SDKDataSource("aws_ec2_local_gateway_route_table", name="Local Gateway Route Table")
// @Tags
// @Testing(tagsTest=false)
func dataSourceLocalGatewayRouteTable() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceLocalGatewayRouteTableRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrFilter: customFiltersSchema(),
			"local_gateway_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"local_gateway_route_table_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"outpost_arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceLocalGatewayRouteTableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeLocalGatewayRouteTablesInput{}

	if v, ok := d.GetOk("local_gateway_route_table_id"); ok {
		input.LocalGatewayRouteTableIds = []string{v.(string)}
	}

	input.Filters = newAttributeFilterList(
		map[string]string{
			"local-gateway-id": d.Get("local_gateway_id").(string),
			"outpost-arn":      d.Get("outpost_arn").(string),
			names.AttrState:    d.Get(names.AttrState).(string),
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

	localGatewayRouteTable, err := findLocalGatewayRouteTable(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("EC2 Local Gateway Route Table", err))
	}

	d.SetId(aws.ToString(localGatewayRouteTable.LocalGatewayRouteTableId))
	d.Set("local_gateway_id", localGatewayRouteTable.LocalGatewayId)
	d.Set("local_gateway_route_table_id", localGatewayRouteTable.LocalGatewayRouteTableId)
	d.Set("outpost_arn", localGatewayRouteTable.OutpostArn)
	d.Set(names.AttrState, localGatewayRouteTable.State)

	setTagsOut(ctx, localGatewayRouteTable.Tags)

	return diags
}
