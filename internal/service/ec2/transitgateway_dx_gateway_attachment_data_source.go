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

// @SDKDataSource("aws_ec2_transit_gateway_dx_gateway_attachment")
func DataSourceTransitGatewayDxGatewayAttachment() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceTransitGatewayDxGatewayAttachmentRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"dx_gateway_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"filter": CustomFiltersSchema(),
			"tags":   tftags.TagsSchemaComputed(),
			"transit_gateway_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceTransitGatewayDxGatewayAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeTransitGatewayAttachmentsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"resource-type": ec2.TransitGatewayAttachmentResourceTypeDirectConnectGateway,
		}),
	}

	input.Filters = append(input.Filters, BuildCustomFilterList(
		d.Get("filter").(*schema.Set),
	)...)

	if v, ok := d.GetOk("tags"); ok {
		input.Filters = append(input.Filters, BuildTagFilterList(
			Tags(tftags.New(ctx, v.(map[string]interface{}))),
		)...)
	}

	// to preserve original functionality
	if v, ok := d.GetOk("dx_gateway_id"); ok {
		input.Filters = append(input.Filters, BuildAttributeFilterList(map[string]string{
			"resource-id": v.(string),
		})...)
	}

	if v, ok := d.GetOk("transit_gateway_id"); ok {
		input.Filters = append(input.Filters, BuildAttributeFilterList(map[string]string{
			"transit-gateway-id": v.(string),
		})...)
	}

	transitGatewayAttachment, err := FindTransitGatewayAttachment(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("EC2 Transit Gateway Direct Connect Gateway Attachment", err))
	}

	d.SetId(aws.StringValue(transitGatewayAttachment.TransitGatewayAttachmentId))
	d.Set("dx_gateway_id", transitGatewayAttachment.ResourceId)
	d.Set("transit_gateway_id", transitGatewayAttachment.TransitGatewayId)

	if err := d.Set("tags", KeyValueTags(ctx, transitGatewayAttachment.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
