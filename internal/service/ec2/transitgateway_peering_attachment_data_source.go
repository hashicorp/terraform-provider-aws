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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKDataSource("aws_ec2_transit_gateway_peering_attachment")
func DataSourceTransitGatewayPeeringAttachment() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceTransitGatewayPeeringAttachmentRead,

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
			"peer_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"peer_region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"peer_transit_gateway_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"transit_gateway_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceTransitGatewayPeeringAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeTransitGatewayPeeringAttachmentsInput{}

	input.Filters = append(input.Filters, BuildCustomFilterList(
		d.Get("filter").(*schema.Set),
	)...)

	if v, ok := d.GetOk("id"); ok {
		input.TransitGatewayAttachmentIds = aws.StringSlice([]string{v.(string)})
	}

	if v, ok := d.GetOk("tags"); ok {
		input.Filters = append(input.Filters, BuildTagFilterList(
			Tags(tftags.New(ctx, v.(map[string]interface{}))),
		)...)
	}

	if len(input.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filters = nil
	}

	transitGatewayPeeringAttachment, err := FindTransitGatewayPeeringAttachment(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("EC2 Transit Gateway Peering Attachment", err))
	}

	d.SetId(aws.StringValue(transitGatewayPeeringAttachment.TransitGatewayAttachmentId))

	local := transitGatewayPeeringAttachment.RequesterTgwInfo
	peer := transitGatewayPeeringAttachment.AccepterTgwInfo

	if aws.StringValue(transitGatewayPeeringAttachment.AccepterTgwInfo.OwnerId) == meta.(*conns.AWSClient).AccountID && aws.StringValue(transitGatewayPeeringAttachment.AccepterTgwInfo.Region) == meta.(*conns.AWSClient).Region {
		local = transitGatewayPeeringAttachment.AccepterTgwInfo
		peer = transitGatewayPeeringAttachment.RequesterTgwInfo
	}

	d.Set("peer_account_id", peer.OwnerId)
	d.Set("peer_region", peer.Region)
	d.Set("peer_transit_gateway_id", peer.TransitGatewayId)
	d.Set("transit_gateway_id", local.TransitGatewayId)

	if err := d.Set("tags", KeyValueTags(ctx, transitGatewayPeeringAttachment.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
