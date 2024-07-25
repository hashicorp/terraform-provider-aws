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

// @SDKDataSource("aws_ec2_coip_pool", name="COIP Pool")
// @Tags
// @Testing(tagsTest=false)
func dataSourceCoIPPool() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceCoIPPoolRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrFilter: customFiltersSchema(),
			"local_gateway_route_table_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"pool_cidrs": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"pool_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceCoIPPoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeCoipPoolsInput{}

	if v, ok := d.GetOk("pool_id"); ok {
		input.PoolIds = []string{v.(string)}
	}

	if v, ok := d.GetOk("local_gateway_route_table_id"); ok {
		input.Filters = append(input.Filters, newAttributeFilterList(map[string]string{
			"coip-pool.local-gateway-route-table-id": v.(string),
		})...)
	}

	if tags, tagsOk := d.GetOk(names.AttrTags); tagsOk {
		input.Filters = append(input.Filters, newTagFilterList(
			Tags(tftags.New(ctx, tags.(map[string]interface{}))),
		)...)
	}

	input.Filters = append(input.Filters, newCustomFilterList(
		d.Get(names.AttrFilter).(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filters = nil
	}

	coip, err := findCOIPPool(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("EC2 COIP Pool", err))
	}

	d.SetId(aws.ToString(coip.PoolId))
	d.Set(names.AttrARN, coip.PoolArn)
	d.Set("local_gateway_route_table_id", coip.LocalGatewayRouteTableId)
	d.Set("pool_cidrs", coip.PoolCidrs)
	d.Set("pool_id", coip.PoolId)

	setTagsOut(ctx, coip.Tags)

	return diags
}
