// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"
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

// @SDKDataSource("aws_ec2_local_gateway")
func DataSourceLocalGateway() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceLocalGatewayRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"outpost_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"filter": customFiltersSchema(),

			names.AttrState: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			names.AttrTags: tftags.TagsSchemaComputed(),

			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceLocalGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	req := &ec2.DescribeLocalGatewaysInput{}

	if v, ok := d.GetOk(names.AttrID); ok {
		req.LocalGatewayIds = []*string{aws.String(v.(string))}
	}

	req.Filters = newAttributeFilterList(
		map[string]string{
			names.AttrState: d.Get(names.AttrState).(string),
		},
	)

	if tags, tagsOk := d.GetOk(names.AttrTags); tagsOk {
		req.Filters = append(req.Filters, newTagFilterList(
			Tags(tftags.New(ctx, tags.(map[string]interface{}))),
		)...)
	}

	req.Filters = append(req.Filters, newCustomFilterList(
		d.Get("filter").(*schema.Set),
	)...)
	if len(req.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		req.Filters = nil
	}

	log.Printf("[DEBUG] Reading AWS LOCAL GATEWAY: %s", req)
	resp, err := conn.DescribeLocalGatewaysWithContext(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing EC2 Local Gateways: %s", err)
	}
	if resp == nil || len(resp.LocalGateways) == 0 {
		return sdkdiag.AppendErrorf(diags, "no matching Local Gateway found")
	}
	if len(resp.LocalGateways) > 1 {
		return sdkdiag.AppendErrorf(diags, "multiple Local Gateways matched; use additional constraints to reduce matches to a single Local Gateway")
	}

	localGateway := resp.LocalGateways[0]

	d.SetId(aws.StringValue(localGateway.LocalGatewayId))
	d.Set("outpost_arn", localGateway.OutpostArn)
	d.Set(names.AttrOwnerID, localGateway.OwnerId)
	d.Set(names.AttrState, localGateway.State)

	if err := d.Set(names.AttrTags, KeyValueTags(ctx, localGateway.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
