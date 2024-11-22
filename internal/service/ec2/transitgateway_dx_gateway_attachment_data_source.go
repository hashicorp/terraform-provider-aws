// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ec2_transit_gateway_dx_gateway_attachment", name="Transit Gateway Direct Connect Gateway Attachment")
// @Tags
// @Testing(tagsTest=false)
func dataSourceTransitGatewayDxGatewayAttachment() *schema.Resource {
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
			names.AttrFilter: customFiltersSchema(),
			names.AttrTags:   tftags.TagsSchemaComputed(),
			names.AttrTransitGatewayID: {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceTransitGatewayDxGatewayAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeTransitGatewayAttachmentsInput{
		Filters: newAttributeFilterList(map[string]string{
			"resource-type": string(awstypes.TransitGatewayAttachmentResourceTypeDirectConnectGateway),
		}),
	}

	input.Filters = append(input.Filters, newCustomFilterList(
		d.Get(names.AttrFilter).(*schema.Set),
	)...)

	if v, ok := d.GetOk(names.AttrTags); ok {
		input.Filters = append(input.Filters, newTagFilterList(
			Tags(tftags.New(ctx, v.(map[string]interface{}))),
		)...)
	}

	// to preserve original functionality
	if v, ok := d.GetOk("dx_gateway_id"); ok {
		input.Filters = append(input.Filters, newAttributeFilterList(map[string]string{
			"resource-id": v.(string),
		})...)
	}

	if v, ok := d.GetOk(names.AttrTransitGatewayID); ok {
		input.Filters = append(input.Filters, newAttributeFilterList(map[string]string{
			"transit-gateway-id": v.(string),
		})...)
	}

	transitGatewayAttachment, err := findTransitGatewayAttachment(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("EC2 Transit Gateway Direct Connect Gateway Attachment", err))
	}

	d.SetId(aws.ToString(transitGatewayAttachment.TransitGatewayAttachmentId))
	d.Set("dx_gateway_id", transitGatewayAttachment.ResourceId)
	d.Set(names.AttrTransitGatewayID, transitGatewayAttachment.TransitGatewayId)

	setTagsOut(ctx, transitGatewayAttachment.Tags)

	return diags
}
